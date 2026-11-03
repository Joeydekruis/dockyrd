package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// VerifyWebhookSignature checks X-Hub-Signature-256 when secret is configured.
func VerifyWebhookSignature(secret string, body []byte, header string) bool {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return true
	}
	if !strings.HasPrefix(header, "sha256=") {
		return false
	}
	sigHex := strings.TrimPrefix(header, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)
	got, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, got)
}

// PushEvent is the subset of a GitHub push webhook we care about.
type PushEvent struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"head_commit"`
}

// ParsePushEvent decodes a push delivery body.
func ParsePushEvent(body []byte) (PushEvent, error) {
	var ev PushEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		return PushEvent{}, err
	}
	if ev.Repository.FullName == "" {
		return PushEvent{}, fmt.Errorf("missing repository.full_name")
	}
	return ev, nil
}

// BranchFromRef maps refs/heads/main → main.
func BranchFromRef(ref string) string {
	ref = strings.TrimSpace(ref)
	const prefix = "refs/heads/"
	if strings.HasPrefix(ref, prefix) {
		return strings.TrimPrefix(ref, prefix)
	}
	return ref
}

// InstallationEvent is a subset of GitHub installation webhook payloads.
type InstallationEvent struct {
	Action       string `json:"action"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
}

// ParseInstallationEvent decodes an installation webhook body.
func ParseInstallationEvent(body []byte) (InstallationEvent, error) {
	var ev InstallationEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		return InstallationEvent{}, err
	}
	if ev.Installation.ID == 0 {
		return InstallationEvent{}, fmt.Errorf("missing installation.id")
	}
	return ev, nil
}
