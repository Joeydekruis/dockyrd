package mail

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/instance"
)

func TestSanitizeHeaderStripsCRLF(t *testing.T) {
	t.Parallel()
	got := buildPlainMessage("from@x.com", "to@x.com", "subj", "<id@h>", "body")
	raw := string(got)
	if strings.Contains(raw, "evil\r\nBcc:") {
		t.Fatal("CRLF injection in headers")
	}
	subject := sanitizeHeaderValue("hello\r\nBcc: evil@x.com")
	if strings.Contains(subject, "\n") || strings.Contains(subject, "\r") {
		t.Fatalf("subject not sanitized: %q", subject)
	}
}

func TestSendPlainTimeoutOnHang(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(2 * time.Minute)
	}()

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	cfg := instance.SMTPConfig{Host: host, Port: mustAtoi(port), From: "from@example.com"}
	start := time.Now()
	err = SendPlain(cfg, "to@example.com", "subj", "body")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed := time.Since(start); elapsed > 45*time.Second {
		t.Fatalf("took too long: %v", elapsed)
	}
}

func TestBuildPlainMessageHasDateAndMessageID(t *testing.T) {
	t.Parallel()
	raw := string(buildPlainMessage("a@b.c", "d@e.f", "Sub", "<m@h>", "hi"))
	for _, line := range []string{"Date:", "Message-ID:", "MIME-Version:"} {
		if !strings.Contains(raw, line) {
			t.Fatalf("missing %s in %q", line, raw)
		}
	}
}

func TestFakeSMTPSendPlain(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go fakeSMTP(ln)

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	cfg := instance.SMTPConfig{Host: host, Port: mustAtoi(port), From: "from@example.com"}
	if err := SendPlain(cfg, "to@example.com", "Test", "hello"); err != nil {
		t.Fatal(err)
	}
}

func fakeSMTP(ln net.Listener) {
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	write := func(s string) { _, _ = rw.WriteString(s); _ = rw.Flush() }
	write("220 localhost\r\n")
	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		upper := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			write("250-localhost\r\n250 STARTTLS\r\n")
		case strings.HasPrefix(upper, "MAIL FROM"):
			write("250 ok\r\n")
		case strings.HasPrefix(upper, "RCPT TO"):
			write("250 ok\r\n")
		case strings.HasPrefix(upper, "DATA"):
			write("354 go\r\n")
		case upper == ".":
			write("250 queued\r\n")
		case strings.HasPrefix(upper, "QUIT"):
			write("221 bye\r\n")
			return
		default:
			if strings.HasPrefix(upper, "STARTTLS") {
				write("502 not supported\r\n")
				continue
			}
		}
	}
}

func mustAtoi(s string) int {
	var n int
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
