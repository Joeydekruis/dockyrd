package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/instance"
)

// SendPlain sends a UTF-8 plain-text message.
func SendPlain(cfg instance.SMTPConfig, to, subject, body string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient required")
	}
	from := cfg.From
	if from == "" {
		from = cfg.Username
	}
	msg := strings.Builder{}
	msg.WriteString("From: " + from + "\r\n")
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	addr := smtpAddr(cfg)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if cfg.UseTLS && cfg.Port == 465 {
		return sendTLS(addr, cfg.Host, auth, from, []string{to}, []byte(msg.String()))
	}
	return smtp.SendMail(addr, auth, extractEmail(from), []string{to}, []byte(msg.String()))
}

func sendTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(extractEmail(from)); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func extractEmail(from string) string {
	if i := strings.LastIndex(from, "<"); i >= 0 {
		if j := strings.Index(from[i:], ">"); j > 0 {
			return strings.TrimSpace(from[i+1 : i+j])
		}
	}
	return strings.TrimSpace(from)
}

// VerifyConnection checks TCP reachability to the SMTP host.
func VerifyConnection(cfg instance.SMTPConfig) error {
	conn, err := net.Dial("tcp", smtpAddr(cfg))
	if err != nil {
		return err
	}
	return conn.Close()
}

func smtpAddr(cfg instance.SMTPConfig) string {
	return net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
}
