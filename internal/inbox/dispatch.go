package inbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/events"
	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/mail"
)

// Dispatcher creates per-user inbox rows and optional email for critical events.
type Dispatcher struct {
	Cfg config.Config
	Q   *sqlc.Queries
	Hub *events.Hub
	Log *slog.Logger
}

type userNotificationPrefs struct {
	Deployments    bool `json:"deployments"`
	SecurityAlerts bool `json:"security_alerts"`
	ProductUpdates bool `json:"product_updates"`
}

func defaultPrefs() userNotificationPrefs {
	return userNotificationPrefs{Deployments: true, SecurityAlerts: true, ProductUpdates: false}
}

func loadUserPrefs(ctx context.Context, q *sqlc.Queries, userID string) userNotificationPrefs {
	prefs := defaultPrefs()
	if q == nil {
		return prefs
	}
	key := fmt.Sprintf("user_notifications:%s", userID)
	row, err := q.GetSetting(ctx, key)
	if err != nil {
		return prefs
	}
	raw := []byte(fmt.Sprint(row.Value))
	if len(raw) == 0 || string(raw) == "<nil>" {
		return prefs
	}
	_ = json.Unmarshal(raw, &prefs)
	return prefs
}

// NotifyDeployFailed creates inbox notifications (and email when SMTP + prefs allow).
func (d *Dispatcher) NotifyDeployFailed(ctx context.Context, serviceID, deploymentID, detail string) {
	d.notifyDeploy(ctx, serviceID, deploymentID, "error", "Build failed", detail, true)
}

// NotifyDeployLive creates in-app-only success notifications.
func (d *Dispatcher) NotifyDeployLive(ctx context.Context, serviceID, deploymentID string) {
	d.notifyDeploy(ctx, serviceID, deploymentID, "success", "Deployed successfully", "", false)
}

func (d *Dispatcher) notifyDeploy(
	ctx context.Context,
	serviceID, deploymentID, level, title, detail string,
	sendEmail bool,
) {
	if d == nil || d.Q == nil {
		return
	}
	ctxRow, err := d.Q.GetServiceDeployContext(ctx, serviceID)
	if err != nil {
		return
	}
	dep, err := d.Q.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		return
	}
	sourceLabel := strings.ToUpper(ctxRow.ProjectSlug + " / " + ctxRow.ServiceSlug)
	body := deployBody(dep, detail)
	ctaLabel, ctaHref := deployCTA(ctxRow.ServiceKind, ctxRow.ServiceSlug, deploymentID)

	users, err := d.Q.ListUsers(ctx)
	if err != nil {
		return
	}
	smtpCfg, _, baseURL, smtpOK := d.smtpReady(ctx)
	now := time.Now().Unix()

	for _, u := range users {
		if u.DisabledAt.Valid {
			continue
		}
		prefs := loadUserPrefs(ctx, d.Q, u.ID)
		if !prefs.Deployments {
			continue
		}
		shouldEmail := sendEmail && smtpOK && prefs.Deployments
		if err := d.insertForUser(ctx, insertParams{
			userID:       u.ID,
			userEmail:    u.Email,
			category:     "deploys",
			level:        level,
			title:        title,
			body:         body,
			sourceLabel:  sourceLabel,
			ctaLabel:     ctaLabel,
			ctaHref:      ctaHref,
			sendEmail:    shouldEmail,
			smtpCfg:      smtpCfg,
			baseURL:      baseURL,
			deploymentID: deploymentID,
			serviceID:    serviceID,
			projectID:    ctxRow.ProjectID,
			createdAt:    now,
		}); err != nil && d.Log != nil {
			d.Log.Warn("inbox notify", "user", u.ID, "err", err)
		}
	}
}

type insertParams struct {
	userID       string
	userEmail    string
	category     string
	level        string
	title        string
	body         string
	sourceLabel  string
	ctaLabel     string
	ctaHref      string
	sendEmail    bool
	smtpCfg      instance.SMTPConfig
	baseURL      string
	deploymentID string
	serviceID    string
	projectID    string
	createdAt    int64
}

func (d *Dispatcher) insertForUser(ctx context.Context, p insertParams) error {
	id := ulid.Make().String()
	var tokenHash []byte
	var plainToken string
	if p.sendEmail {
		var err error
		plainToken, tokenHash, err = auth.NewSessionToken()
		if err != nil {
			p.sendEmail = false
		}
	}

	err := d.Q.InsertInboxNotification(ctx, sqlc.InsertInboxNotificationParams{
		ID:             id,
		UserID:         p.userID,
		Category:       p.category,
		Level:          p.level,
		Title:          p.title,
		Body:           p.body,
		SourceLabel:    p.sourceLabel,
		CtaLabel:       nullString(p.ctaLabel),
		CtaHref:        nullString(p.ctaHref),
		EmailTokenHash: tokenHash,
		EmailSentAt:    sql.NullInt64{},
		ReadAt:         sql.NullInt64{},
		DeploymentID:   nullString(p.deploymentID),
		ServiceID:      nullString(p.serviceID),
		ProjectID:      nullString(p.projectID),
		CreatedAt:      p.createdAt,
	})
	if err != nil {
		return err
	}

	if p.sendEmail && plainToken != "" {
		openPath := "/api/v1/inbox/open?token=" + plainToken
		link := openPath
		if p.baseURL != "" {
			link = p.baseURL + openPath
		}
		subject := "Dockyard: " + p.title
		text := fmt.Sprintf("%s\n\n%s\n\n%s\n\nOpen in Dockyard (marks as read):\n%s\n",
			p.title, p.body, p.sourceLabel, link)
		if err := mail.SendPlain(p.smtpCfg, p.userEmail, subject, text); err != nil {
			if d.Log != nil {
				d.Log.Warn("inbox email", "user", p.userID, "err", err)
			}
		} else {
			_ = d.Q.SetInboxNotificationEmailSent(ctx, sqlc.SetInboxNotificationEmailSentParams{
				EmailSentAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
				ID:          id,
			})
		}
	}

	if d.Hub != nil {
		d.Hub.Publish("inbox:"+p.userID, map[string]any{
			"id":    id,
			"level": p.level,
			"title": p.title,
		})
	}
	return nil
}

func deployBody(dep sqlc.GetDeploymentByIDRow, detail string) string {
	var parts []string
	if dep.CommitSha.Valid && dep.CommitSha.String != "" {
		sha := dep.CommitSha.String
		if len(sha) > 7 {
			sha = sha[:7]
		}
		parts = append(parts, sha)
	}
	if dep.CommitMessage.Valid && strings.TrimSpace(dep.CommitMessage.String) != "" {
		parts = append(parts, strings.TrimSpace(dep.CommitMessage.String))
	}
	if detail != "" {
		parts = append(parts, detail)
	}
	return strings.Join(parts, ". ")
}

func deployCTA(serviceKind, serviceSlug, deploymentID string) (string, string) {
	switch serviceKind {
	case "static_site":
		return "View build log", "/static-sites/" + serviceSlug + "/deployments/" + deploymentID
	case "app":
		return "View build log", "/applications/" + serviceSlug
	default:
		return "", ""
	}
}

func (d *Dispatcher) smtpReady(ctx context.Context) (instance.SMTPConfig, instance.Panel, string, bool) {
	panel, _ := instance.LoadPanel(ctx, d.Q)
	base := instance.PublicBaseURL(d.Cfg, panel)
	cfg, ok, _ := instance.LoadSMTPConfig(ctx, d.Q, d.Cfg)
	return cfg, panel, base, ok && cfg.Host != ""
}

func nullString(s string) sql.NullString {
	if strings.TrimSpace(s) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
