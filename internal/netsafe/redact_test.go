package netsafe

import (
	"strings"
	"testing"
)

func TestRedactError(t *testing.T) {
	t.Parallel()
	secret := "https://hooks.slack.com/services/SECRET"
	err := &urlErr{msg: "Post " + secret + ": timeout"}
	got := RedactError(err, secret)
	if strings.Contains(got, "SECRET") {
		t.Fatalf("secret leaked: %s", got)
	}
}

type urlErr struct{ msg string }

func (e *urlErr) Error() string { return e.msg }
