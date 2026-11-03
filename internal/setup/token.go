package setup

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

const (
	settingKeySetupToken = "setup_token"
	tokenTTL             = 60 * time.Minute
	tokenAlphabet        = "abcdefghijklmnopqrstuvwxyz0123456789"
)

type tokenSetting struct {
	TokenHash string `json:"token_hash"`
	ExpiresAt int64  `json:"expires_at"`
	UsedAt    *int64 `json:"used_at,omitempty"`
}

// GenerateToken returns a human-readable token and its sha256 hex hash for storage.
func GenerateToken(now time.Time) (plain string, setting tokenSetting, err error) {
	plain, err = randomTokenString(3, 4)
	if err != nil {
		return "", tokenSetting{}, err
	}
	sum := sha256.Sum256([]byte(plain))
	setting = tokenSetting{
		TokenHash: hex.EncodeToString(sum[:]),
		ExpiresAt: now.Add(tokenTTL).Unix(),
	}
	return plain, setting, nil
}

func randomTokenString(groups, groupLen int) (string, error) {
	max := big.NewInt(int64(len(tokenAlphabet)))
	parts := make([]string, groups)
	for g := range parts {
		b := make([]byte, groupLen)
		for i := range b {
			n, err := rand.Int(rand.Reader, max)
			if err != nil {
				return "", err
			}
			b[i] = tokenAlphabet[n.Int64()]
		}
		parts[g] = string(b)
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += "-" + p
	}
	return out, nil
}

func marshalTokenSetting(s tokenSetting) ([]byte, error) {
	return json.Marshal(s)
}

func parseTokenSetting(raw interface{}) (tokenSetting, error) {
	data, err := settingValueBytes(raw)
	if err != nil {
		return tokenSetting{}, err
	}
	var s tokenSetting
	if err := json.Unmarshal(data, &s); err != nil {
		return tokenSetting{}, err
	}
	if s.TokenHash == "" || s.ExpiresAt == 0 {
		return tokenSetting{}, fmt.Errorf("invalid setup_token setting")
	}
	return s, nil
}

func settingValueBytes(v interface{}) ([]byte, error) {
	switch t := v.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	default:
		return json.Marshal(v)
	}
}
