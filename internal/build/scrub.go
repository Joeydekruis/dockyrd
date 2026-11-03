package build

import (
	"io"
	"strings"
)

// ScrubWriter replaces secret values in build output before writing to disk.
type ScrubWriter struct {
	W      io.Writer
	Secrets []string
}

func (s ScrubWriter) Write(p []byte) (int, error) {
	out := string(p)
	for _, secret := range s.Secrets {
		secret = strings.TrimSpace(secret)
		if secret == "" {
			continue
		}
		out = strings.ReplaceAll(out, secret, "••••")
	}
	n, err := s.W.Write([]byte(out))
	if err != nil {
		return n, err
	}
	return len(p), nil
}
