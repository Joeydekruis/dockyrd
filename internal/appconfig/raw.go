package appconfig

import "encoding/json"

// RawConfig normalizes sqlc JSON columns to RawMessage.
func RawConfig(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case json.RawMessage:
		return t
	case []byte:
		return json.RawMessage(t)
	case string:
		return json.RawMessage(t)
	default:
		b, _ := json.Marshal(v)
		return b
	}
}
