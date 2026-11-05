package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Client calls Dockyard HTTP APIs with a stored Bearer token.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		Token:   strings.TrimSpace(token),
		HTTP:    http.DefaultClient,
	}
}

// ConfigDir returns the dy CLI config directory.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "dy")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// LoadToken reads the saved API token from disk.
func LoadToken() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(filepath.Join(dir, "token"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

// SaveToken persists the API token for future dy commands.
func SaveToken(token string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "token"), []byte(strings.TrimSpace(token)+"\n"), 0o600)
}

// SaveBaseURL stores the panel base URL.
func SaveBaseURL(base string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "base_url"), []byte(strings.TrimRight(strings.TrimSpace(base), "/")+"\n"), 0o600)
}

// LoadBaseURL reads the saved panel URL or DOCKYARD_URL.
func LoadBaseURL() (string, error) {
	if v := strings.TrimSpace(os.Getenv("DOCKYARD_URL")); v != "" {
		return strings.TrimRight(v, "/"), nil
	}
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(filepath.Join(dir, "base_url"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

func (c *Client) do(method, path string, body any) (*http.Response, []byte, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, rdr)
	if err != nil {
		return nil, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	cli := c.HTTP
	if cli == nil {
		cli = http.DefaultClient
	}
	res, err := cli.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	return res, data, err
}

// Whoami returns the authenticated user from GET /api/v1/auth/me (session) or implied token owner via projects.
func (c *Client) Whoami() (map[string]any, error) {
	res, data, err := c.do(http.MethodGet, "/api/v1/auth/me", nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusUnauthorized {
		// Bearer tokens cannot call /auth/me — list projects instead.
		res, data, err = c.do(http.MethodGet, "/api/v1/projects", nil)
		if err != nil {
			return nil, err
		}
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("api error %d: %s", res.StatusCode, strings.TrimSpace(string(data)))
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListServices(projectID string) (map[string]any, error) {
	res, data, err := c.do(http.MethodGet, "/api/v1/services?project_id="+projectID, nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("api error %d: %s", res.StatusCode, strings.TrimSpace(string(data)))
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateDeployment(serviceID string) (map[string]any, error) {
	res, data, err := c.do(http.MethodPost, "/api/v1/services/"+serviceID+"/deployments", map[string]any{})
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("api error %d: %s", res.StatusCode, strings.TrimSpace(string(data)))
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ServiceLogs(serviceID, stream, tail string) (string, error) {
	path := fmt.Sprintf("/api/v1/services/%s/logs?stream=%s&tail=%s", serviceID, stream, tail)
	res, data, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("api error %d: %s", res.StatusCode, strings.TrimSpace(string(data)))
	}
	return string(data), nil
}
