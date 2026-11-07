package mail

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/instance"
)

const (
	dialTimeout  = 10 * time.Second
	sessionLimit = 30 * time.Second
)

// SendPlain sends a UTF-8 plain-text message.
func SendPlain(cfg instance.SMTPConfig, to, subject, body string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient required")
	}
	toAddr, err := mailParseAddress(to)
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	from := cfg.From
	if from == "" {
		from = cfg.Username
	}
	from = sanitizeHeaderValue(from)
	subject = sanitizeHeaderValue(subject)
	encodedSubject := mime.QEncoding.Encode("utf-8", subject)
	msgID := messageID(cfg.Host)
	msg := buildPlainMessage(from, to, encodedSubject, msgID, body)
	return sendMessage(cfg, from, []string{toAddr}, msg)
}

func buildPlainMessage(from, to, encodedSubject, messageID, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + sanitizeHeaderValue(to) + "\r\n")
	b.WriteString("Subject: " + encodedSubject + "\r\n")
	b.WriteString("Date: " + time.Now().UTC().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("Message-ID: " + messageID + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

func messageID(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		host = "dockyrd.local"
	}
	var rnd [12]byte
	_, _ = rand.Read(rnd[:])
	return "<" + hex.EncodeToString(rnd[:]) + "@" + host + ">"
}

func sanitizeHeaderValue(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.TrimSpace(s)
}

func mailParseAddress(s string) (string, error) {
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return "", err
	}
	return addr.Address, nil
}

func sendMessage(cfg instance.SMTPConfig, from string, to []string, msg []byte) error {
	addr := smtpAddr(cfg)
	if cfg.UseTLS && cfg.Port == 465 {
		return sendTLS(addr, cfg.Host, cfg, from, to, msg)
	}
	dialer := net.Dialer{Timeout: dialTimeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	deadline := time.Now().Add(sessionLimit)
	_ = conn.SetDeadline(deadline)

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if cfg.UseTLS && cfg.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: cfg.Host}); err != nil {
				return err
			}
		}
	}

	auth := smtpAuth(cfg)
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(extractEmail(from)); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(extractEmail(rcpt)); err != nil {
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

func smtpAuth(cfg instance.SMTPConfig) smtp.Auth {
	if strings.TrimSpace(cfg.Username) == "" {
		return nil
	}
	return smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
}

func sendTLS(addr, host string, cfg instance.SMTPConfig, from string, to []string, msg []byte) error {
	dialer := net.Dialer{Timeout: dialTimeout}
	conn, err := tls.DialWithDialer(&dialer, "tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(sessionLimit))

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	auth := smtpAuth(cfg)
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(extractEmail(from)); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(extractEmail(rcpt)); err != nil {
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
	dialer := net.Dialer{Timeout: dialTimeout}
	conn, err := dialer.Dial("tcp", smtpAddr(cfg))
	if err != nil {
		return err
	}
	return conn.Close()
}

func smtpAddr(cfg instance.SMTPConfig) string {
	return net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
}
