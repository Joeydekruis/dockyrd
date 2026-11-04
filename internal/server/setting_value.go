package server

import "encoding/json"

// settingValueBytes normalizes a settings row value for JSON decode.
func settingValueBytes(v any) ([]byte, error) {
	switch t := v.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	default:
		return json.Marshal(v)
	}
}
