package github

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

const githubAPI = "https://api.github.com"

// AppCredentials loaded from git_sources.
type AppCredentials struct {
	AppID          int64
	InstallationID int64
	PrivateKeyPEM  []byte
	WebhookSecret  string
}

// LoadCredentials reads and decrypts the default GitHub App row.
func LoadCredentials(ctx context.Context, q *sqlc.Queries, keyPath string) (AppCredentials, error) {
	row, err := q.GetGitSource(ctx, defaultGitSourceID)
	if err != nil {
		return AppCredentials{}, err
	}
	key, err := crypto.LoadMasterKey(keyPath)
	if err != nil {
		return AppCredentials{}, err
	}
	var pemBytes []byte
	if len(row.PrivateKeyEnc) > 0 {
		pemBytes, err = crypto.Decrypt(key, row.PrivateKeyEnc)
		if err != nil {
			return AppCredentials{}, err
		}
	}
	var wh string
	if len(row.WebhookSecretEnc) > 0 {
		raw, err := crypto.Decrypt(key, row.WebhookSecretEnc)
		if err != nil {
			return AppCredentials{}, err
		}
		wh = string(raw)
	}
	creds := AppCredentials{WebhookSecret: wh}
	if row.GithubAppID.Valid {
		creds.AppID = row.GithubAppID.Int64
	}
	if row.InstallationID.Valid {
		creds.InstallationID = row.InstallationID.Int64
	}
	creds.PrivateKeyPEM = pemBytes
	return creds, nil
}

func (c AppCredentials) appJWT() (string, error) {
	if c.AppID == 0 || len(c.PrivateKeyPEM) == 0 {
		return "", fmt.Errorf("github app not configured")
	}
	block, _ := pem.Decode(c.PrivateKeyPEM)
	if block == nil {
		return "", fmt.Errorf("invalid app private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return "", err
		}
		var ok bool
		key, ok = parsed.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("unexpected private key type")
		}
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": strconv.FormatInt(c.AppID, 10),
	})
	return token.SignedString(key)
}

// InstallationAccessToken mints a short-lived installation token.
func (c AppCredentials) InstallationAccessToken(ctx context.Context) (string, error) {
	if c.InstallationID == 0 {
		return "", fmt.Errorf("github app installation missing")
	}
	jwtStr, err := c.appJWT()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/app/installations/%d/access_tokens", githubAPI, c.InstallationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+jwtStr)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("installation token %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.Token == "" {
		return "", fmt.Errorf("empty installation token")
	}
	return out.Token, nil
}

// InstallationAccount is the GitHub user or org that installed the app.
type InstallationAccount struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

// InstallationAccount loads the account for the configured installation.
func (c AppCredentials) InstallationAccount(ctx context.Context) (InstallationAccount, error) {
	if c.InstallationID == 0 {
		return InstallationAccount{}, fmt.Errorf("github app installation missing")
	}
	jwtStr, err := c.appJWT()
	if err != nil {
		return InstallationAccount{}, err
	}
	url := fmt.Sprintf("%s/app/installations/%d", githubAPI, c.InstallationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return InstallationAccount{}, err
	}
	req.Header.Set("Authorization", "Bearer "+jwtStr)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return InstallationAccount{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return InstallationAccount{}, fmt.Errorf("installation %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		Account InstallationAccount `json:"account"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return InstallationAccount{}, err
	}
	return out.Account, nil
}

// RepoSummary is a GitHub repository visible to the installation.
type RepoSummary struct {
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
}

// ListRepos returns installation repositories (paginated up to maxPages).
func ListRepos(ctx context.Context, token, query string) ([]RepoSummary, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	var all []RepoSummary
	for page := 1; page <= 5; page++ {
		url := fmt.Sprintf("%s/installation/repositories?per_page=100&page=%d", githubAPI, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("list repos %d: %s", resp.StatusCode, string(body))
		}
		var pageResp struct {
			Repositories []RepoSummary `json:"repositories"`
		}
		if err := json.Unmarshal(body, &pageResp); err != nil {
			return nil, err
		}
		if len(pageResp.Repositories) == 0 {
			break
		}
		for _, r := range pageResp.Repositories {
			if query == "" || strings.Contains(strings.ToLower(r.FullName), query) {
				all = append(all, r)
			}
		}
		if len(pageResp.Repositories) < 100 {
			break
		}
	}
	return all, nil
}

// BranchSummary is a repo branch head.
type BranchSummary struct {
	Name string `json:"name"`
}

// ListBranches lists branch names for a repository.
func ListBranches(ctx context.Context, token, owner, repo string) ([]BranchSummary, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/branches?per_page=100", githubAPI, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("list branches %d: %s", resp.StatusCode, string(body))
	}
	var branches []BranchSummary
	if err := json.Unmarshal(body, &branches); err != nil {
		return nil, err
	}
	return branches, nil
}

// AuthenticatedCloneURL returns an HTTPS clone URL with embedded token when needed.
func AuthenticatedCloneURL(token, ownerRepo string) string {
	ownerRepo = strings.Trim(strings.TrimSpace(ownerRepo), "/")
	if token == "" {
		return "https://github.com/" + ownerRepo + ".git"
	}
	return "https://x-access-token:" + token + "@github.com/" + ownerRepo + ".git"
}
