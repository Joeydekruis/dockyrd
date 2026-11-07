package instance

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

const (
	SettingKeyInstance = "instance"
	SettingKeySMTP     = "smtp"
)

// Panel holds instance-level dashboard settings.
type Panel struct {
	PublicURL            string `json:"public_url,omitempty"`
	AcmeEmail            string `json:"acme_email,omitempty"`
	Wildcard             string `json:"wildcard_domain,omitempty"`
	WildcardVerifiedAt   int64  `json:"wildcard_verified_at,omitempty"`
	WildcardProbeHost    string `json:"wildcard_probe_host,omitempty"`
	Require2FAAll        bool   `json:"require_2fa_all,omitempty"`
	TerminalAllowHost    bool   `json:"terminal_allow_host,omitempty"`
}

// SMTPPublic is returned to clients (no secrets).
type SMTPPublic struct {
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	Username   string `json:"username,omitempty"`
	From       string `json:"from,omitempty"`
	UseTLS     bool   `json:"use_tls"`
	Configured bool   `json:"configured"`
}

type smtpStored struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	From         string `json:"from"`
	UseTLS       bool   `json:"use_tls"`
	PasswordEnc  string `json:"password_enc,omitempty"`
	PasswordSet  bool   `json:"password_set"`
}

// SMTPConfig is a decrypted SMTP profile for sending mail.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

func loadJSONSetting(ctx context.Context, q *sqlc.Queries, key string, dest any) error {
	row, err := q.GetSetting(ctx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	raw, err := valueBytes(row.Value)
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dest)
}

func valueBytes(v any) ([]byte, error) {
	switch t := v.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	default:
		return json.Marshal(v)
	}
}

// LoadPanel reads instance settings.
func LoadPanel(ctx context.Context, q *sqlc.Queries) (Panel, error) {
	var p Panel
	if err := loadJSONSetting(ctx, q, SettingKeyInstance, &p); err != nil {
		return Panel{}, err
	}
	return p, nil
}

// SavePanel persists instance settings.
func SavePanel(ctx context.Context, q *sqlc.Queries, p Panel) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return q.UpsertSetting(ctx, sqlc.UpsertSettingParams{
		Key:   SettingKeyInstance,
		Value: string(raw),
	})
}

// LoadSMTPPublic returns non-secret SMTP fields.
func LoadSMTPPublic(ctx context.Context, q *sqlc.Queries) (SMTPPublic, error) {
	var st smtpStored
	if err := loadJSONSetting(ctx, q, SettingKeySMTP, &st); err != nil {
		return SMTPPublic{}, err
	}
	if st.Host == "" {
		return SMTPPublic{Configured: false}, nil
	}
	return SMTPPublic{
		Host:       st.Host,
		Port:       st.Port,
		Username:   st.Username,
		From:       st.From,
		UseTLS:     st.UseTLS,
		Configured: st.PasswordSet || st.PasswordEnc != "",
	}, nil
}

// SaveSMTP stores SMTP settings; password replaced when password non-empty.
func SaveSMTP(ctx context.Context, q *sqlc.Queries, cfg config.Config, in SMTPPublic, password string) error {
	var st smtpStored
	_ = loadJSONSetting(ctx, q, SettingKeySMTP, &st)
	st.Host = in.Host
	if in.Port > 0 {
		st.Port = in.Port
	} else if st.Port == 0 {
		st.Port = 587
	}
	st.Username = in.Username
	st.From = in.From
	st.UseTLS = in.UseTLS
	if password != "" {
		key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
		if err != nil {
			return err
		}
		enc, err := crypto.Encrypt(key, []byte(password))
		if err != nil {
			return err
		}
		st.PasswordEnc = base64.StdEncoding.EncodeToString(enc)
		st.PasswordSet = true
	}
	raw, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return q.UpsertSetting(ctx, sqlc.UpsertSettingParams{
		Key:   SettingKeySMTP,
		Value: string(raw),
	})
}

// LoadSMTPConfig decrypts SMTP credentials for sending.
func LoadSMTPConfig(ctx context.Context, q *sqlc.Queries, cfg config.Config) (SMTPConfig, bool, error) {
	var st smtpStored
	if err := loadJSONSetting(ctx, q, SettingKeySMTP, &st); err != nil {
		return SMTPConfig{}, false, err
	}
	if st.Host == "" || st.PasswordEnc == "" {
		return SMTPConfig{}, false, nil
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return SMTPConfig{}, false, err
	}
	enc, err := base64.StdEncoding.DecodeString(st.PasswordEnc)
	if err != nil {
		return SMTPConfig{}, false, err
	}
	pass, err := crypto.Decrypt(key, enc)
	if err != nil {
		return SMTPConfig{}, false, err
	}
	port := st.Port
	if port == 0 {
		port = 587
	}
	return SMTPConfig{
		Host:     st.Host,
		Port:     port,
		Username: st.Username,
		Password: string(pass),
		From:     st.From,
		UseTLS:   st.UseTLS,
	}, true, nil
}

// PublicBaseURL resolves external URL for links in email.
func PublicBaseURL(cfg config.Config, panel Panel) string {
	if u := panel.PublicURL; u != "" {
		return trimSlash(u)
	}
	return trimSlash(cfg.PublicURL())
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// PanelHostname returns the host part of Public URL for HTTPS routing (no port).
func PanelHostname(publicURL string) string {
	u := strings.TrimSpace(publicURL)
	if u == "" {
		return ""
	}
	if !strings.Contains(u, "://") {
		u = "https://" + u
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" || host == "localhost" || host == "127.0.0.1" {
		return ""
	}
	return host
}
