package dbkind

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
)

// BucketKey is an access key row stored in bucket service config.
type BucketKey struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	AccessKeyID string `json:"access_key_id"`
	SecretEnc   string `json:"secret_enc"`
}

// BucketConfig is services.config for kind=bucket.
type BucketConfig struct {
	BucketName string      `json:"bucket_name"`
	PublicRead bool        `json:"public_read"`
	Keys       []BucketKey `json:"keys"`
}

// NewBucketConfigJSON creates a bucket service config with one generated key.
func NewBucketConfigJSON(masterKey []byte, bucketName string, publicRead bool) (json.RawMessage, error) {
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return nil, fmt.Errorf("bucket_name required")
	}
	accessKeyID, secret, err := randomAccessKeyPair()
	if err != nil {
		return nil, err
	}
	secretEnc, err := crypto.Encrypt(masterKey, []byte(secret))
	if err != nil {
		return nil, err
	}
	cfg := BucketConfig{
		BucketName: bucketName,
		PublicRead: publicRead,
		Keys: []BucketKey{{
			ID:          ulid.Make().String(),
			Name:        "default",
			AccessKeyID: accessKeyID,
			SecretEnc:   base64.StdEncoding.EncodeToString(secretEnc),
		}},
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// ParseBucket decodes bucket config.
func ParseBucket(raw json.RawMessage) (BucketConfig, error) {
	var cfg BucketConfig
	if len(raw) == 0 {
		return BucketConfig{}, fmt.Errorf("empty bucket config")
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return BucketConfig{}, err
	}
	if strings.TrimSpace(cfg.BucketName) == "" {
		return BucketConfig{}, fmt.Errorf("bucket_name required")
	}
	return cfg, nil
}

// DecryptKeySecret returns plaintext secret for a bucket key entry.
func DecryptKeySecret(masterKey []byte, k BucketKey) (string, error) {
	enc, err := base64.StdEncoding.DecodeString(strings.TrimSpace(k.SecretEnc))
	if err != nil {
		return "", err
	}
	plain, err := crypto.Decrypt(masterKey, enc)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// NewAccessKeyPair generates a Garage-compatible access key id and secret.
func NewAccessKeyPair() (accessKeyID, secret string, err error) {
	return randomAccessKeyPair()
}

func randomAccessKeyPair() (accessKeyID, secret string, err error) {
	id, err := randomPassword(20)
	if err != nil {
		return "", "", err
	}
	sec, err := randomPassword(32)
	if err != nil {
		return "", "", err
	}
	return "gk_" + id, sec, nil
}
