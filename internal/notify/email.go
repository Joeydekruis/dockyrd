package notify

import (
	"context"
	"fmt"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/mail"
)

// EmailSender delivers plain-text alerts via instance SMTP.
type EmailSender struct {
	Cfg config.Config
	Q   *sqlc.Queries
}

func (s *EmailSender) Kind() string { return "email" }

func (s *EmailSender) Validate(cfg ChannelConfig, _ ChannelSecret, _ bool) error {
	if err := ValidateEmailChannelConfig(cfg); err != nil {
		return err
	}
	if s.Q == nil {
		return fmt.Errorf("database not ready")
	}
	ok, err := SMTPConfigured(context.Background(), s.Q)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("configure SMTP first")
	}
	return nil
}

func (s *EmailSender) Render(ev EventView, _ ChannelConfig, publicBase string) (Rendered, error) {
	subject := emailSubject(ev)
	body := emailBody(ev, publicBase)
	return Rendered{
		Body:        []byte(body),
		ContentType: "text/plain; charset=UTF-8",
		Headers:     map[string]string{"Subject": subject},
	}, nil
}

func emailSubject(ev EventView) string {
	prefix := "[Dockyard]"
	if ev.State == StateResolved {
		prefix = "[Dockyard] [RESOLVED]"
	}
	level := strings.ToUpper(string(ev.Level))
	title := strings.TrimSpace(ev.Title)
	if title == "" {
		title = ev.Type
	}
	return fmt.Sprintf("%s %s %s", prefix, level, title)
}

func emailBody(ev EventView, publicBase string) string {
	var b strings.Builder
	b.WriteString(ev.Title)
	b.WriteString("\n\n")
	if ev.Detail != "" {
		b.WriteString(ev.Detail)
		b.WriteString("\n\n")
	}
	link := absoluteURL(publicBase, ev.Href)
	if link != "" {
		b.WriteString("Open in Dockyard:\n")
		b.WriteString(link)
		b.WriteString("\n")
	}
	b.WriteString("\nSent by Dockyard.\n")
	return b.String()
}

func (s *EmailSender) Send(ctx context.Context, _ ChannelConfig, _ ChannelSecret, r Rendered, meta DeliveryMeta, _ bool) Result {
	if s.Q == nil {
		return Result{Permanent: true, Err: "database not ready"}
	}
	smtpCfg, ok, err := instance.LoadSMTPConfig(ctx, s.Q, s.Cfg)
	if err != nil || !ok || smtpCfg.Host == "" {
		return Result{Permanent: true, Err: "SMTP not configured"}
	}
	to := strings.TrimSpace(meta.Target)
	if to == "" {
		return Result{Permanent: true, Err: "recipient required"}
	}
	subject := r.Headers["Subject"]
	if subject == "" {
		subject = "Dockyard notification"
	}
	from := smtpFromAddress(smtpCfg)
	sendErr := mail.SendPlain(smtpCfg, to, subject, string(r.Body))
	if sendErr != nil {
		RecordEmailLog(ctx, s.Q, EmailLogKindNotification, from, to, subject, "failed", sendErr.Error(), smtpCfg)
		permanent := strings.Contains(strings.ToLower(sendErr.Error()), "invalid recipient")
		return Result{Err: sendErr.Error(), Permanent: permanent}
	}
	RecordEmailLog(ctx, s.Q, EmailLogKindNotification, from, to, subject, "sent", "Message accepted for delivery", smtpCfg)
	return Result{OK: true}
}
