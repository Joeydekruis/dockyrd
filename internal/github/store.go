package github

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

const defaultGitSourceID = "default"

// DefaultGitSourceID is the sqlite row key for the GitHub App integration.
func DefaultGitSourceID() string {
	return defaultGitSourceID
}

// ConversionResult is returned by GitHub after manifest conversion.
type ConversionResult struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	ClientID    string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	WebhookSecret string `json:"webhook_secret"`
	PEM         string `json:"pem"`
}

// SaveConversion encrypts secrets and upserts git_sources.
func SaveConversion(ctx context.Context, q *sqlc.Queries, keyPath string, res ConversionResult, installationID int64) error {
	key, err := crypto.LoadMasterKey(keyPath)
	if err != nil {
		return err
	}
	pkEnc, err := crypto.Encrypt(key, []byte(res.PEM))
	if err != nil {
		return err
	}
	whEnc, err := crypto.Encrypt(key, []byte(res.WebhookSecret))
	if err != nil {
		return err
	}
	csEnc, err := crypto.Encrypt(key, []byte(res.ClientSecret))
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return q.UpsertGitSource(ctx, sqlc.UpsertGitSourceParams{
		ID:               defaultGitSourceID,
		Kind:             sql.NullString{String: "github", Valid: true},
		GithubAppID:      sql.NullInt64{Int64: res.ID, Valid: true},
		InstallationID:   sql.NullInt64{Int64: installationID, Valid: installationID > 0},
		PrivateKeyEnc:    pkEnc,
		WebhookSecretEnc: whEnc,
		ClientID:         sql.NullString{String: res.ClientID, Valid: res.ClientID != ""},
		ClientSecretEnc:  csEnc,
		Slug:             sql.NullString{String: res.Slug, Valid: res.Slug != ""},
		CreatedAt:        sql.NullInt64{Int64: now, Valid: true},
	})
}

// WebhookSecretPlain returns the decrypted webhook secret for verification.
func WebhookSecretPlain(ctx context.Context, q *sqlc.Queries, keyPath string) (string, error) {
	row, err := q.GetGitSource(ctx, defaultGitSourceID)
	if err != nil {
		return "", err
	}
	if len(row.WebhookSecretEnc) == 0 {
		return "", nil
	}
	key, err := crypto.LoadMasterKey(keyPath)
	if err != nil {
		return "", err
	}
	raw, err := crypto.Decrypt(key, row.WebhookSecretEnc)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// StatusResponse describes whether GitHub App is linked.
type StatusResponse struct {
	Linked            bool   `json:"linked"`
	Slug              string `json:"slug,omitempty"`
	InstallationID    int64  `json:"installation_id,omitempty"`
	InstalledOn       string `json:"installed_on,omitempty"`
	RepositoryCount   int    `json:"repository_count,omitempty"`
	WebhookActive     bool   `json:"webhook_active"`
	PublicURL         string `json:"public_url,omitempty"`
	InstallURL        string `json:"install_url,omitempty"`
	ManageURL         string `json:"manage_url,omitempty"`
	AppConfigured     bool   `json:"app_configured"`
	InstallationReady bool   `json:"installation_ready"`
}

// LoadStatus reads git_sources for the dashboard.
func LoadStatus(ctx context.Context, q *sqlc.Queries) (StatusResponse, error) {
	row, err := q.GetGitSource(ctx, defaultGitSourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StatusResponse{}, nil
		}
		return StatusResponse{}, err
	}
	out := StatusResponse{Linked: row.GithubAppID.Valid}
	if row.Slug.Valid {
		out.Slug = row.Slug.String
	}
	if row.InstallationID.Valid {
		out.InstallationID = row.InstallationID.Int64
	}
	return out, nil
}

// SaveInstallationID persists the GitHub App installation id after manifest or webhook.
func SaveInstallationID(ctx context.Context, q *sqlc.Queries, installID int64) error {
	if q == nil || installID <= 0 {
		return nil
	}
	return q.SetGitSourceInstallation(ctx, sqlc.SetGitSourceInstallationParams{
		InstallationID: sql.NullInt64{Int64: installID, Valid: true},
		ID:             defaultGitSourceID,
	})
}

// ListMatchingServices returns service IDs that should auto-deploy for a push.
func ListMatchingServices(ctx context.Context, q *sqlc.Queries, repoFullName, branch string) ([]string, error) {
	rows, err := q.ListDeployableServices(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0)
	for _, row := range rows {
		if matchAutoDeploy(row.Config, row.Kind, repoFullName, branch) {
			out = append(out, row.ID)
		}
	}
	return out, nil
}

type deployMatch struct {
	AutoDeploy bool `json:"autoDeploy"`
	Source     struct {
		Type   string `json:"type"`
		Repo   string `json:"repo"`
		Branch string `json:"branch"`
		URL    string `json:"url"`
	} `json:"source"`
}

func matchAutoDeploy(raw any, kind, repoFullName, branch string) bool {
	if kind != "app" && kind != "static_site" {
		return false
	}
	cfg, err := json.Marshal(raw)
	if err != nil {
		return false
	}
	var m deployMatch
	if err := json.Unmarshal(cfg, &m); err != nil {
		return false
	}
	if !m.AutoDeploy {
		return false
	}
	srcType := stringsToLower(m.Source.Type)
	if srcType != "github" {
		return false
	}
	if stringsTrim(m.Source.Repo) != stringsTrim(repoFullName) {
		return false
	}
	wantBranch := stringsTrim(m.Source.Branch)
	if wantBranch == "" {
		wantBranch = "main"
	}
	return wantBranch == stringsTrim(branch)
}

func stringsToLower(s string) string {
	return stringsTrim(strings.ToLower(s))
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}

// NewStateToken returns a random state value for OAuth-style CSRF (stored in settings).
func NewStateToken() string {
	return ulid.Make().String()
}
