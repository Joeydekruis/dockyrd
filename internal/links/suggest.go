package links

import "fmt"

// SuggestedVariable is a default env row when creating a link.
type SuggestedVariable struct {
	Key        string `json:"key"`
	Ref        string `json:"ref"`
	AtBuild    bool   `json:"at_build"`
	AtRuntime  bool   `json:"at_runtime"`
	DefaultOn  bool   `json:"default_on"`
	IsSecret   bool   `json:"is_secret"`
}

// SuggestedVariables returns default reference env rows for a provider kind.
func SuggestedVariables(providerKind, providerSlug string) []SuggestedVariable {
	ref := func(key string) string {
		return fmt.Sprintf("${{%s.%s}}", providerSlug, key)
	}
	switch providerKind {
	case "postgres":
		return []SuggestedVariable{
			{Key: "DATABASE_URL", Ref: ref("DATABASE_URL"), AtBuild: true, AtRuntime: true, DefaultOn: true, IsSecret: true},
			{Key: "PGHOST", Ref: ref("PGHOST"), AtRuntime: true, DefaultOn: false},
			{Key: "PGPORT", Ref: ref("PGPORT"), AtRuntime: true, DefaultOn: false},
			{Key: "PGUSER", Ref: ref("PGUSER"), AtRuntime: true, DefaultOn: false},
			{Key: "PGPASSWORD", Ref: ref("PGPASSWORD"), AtRuntime: true, DefaultOn: false, IsSecret: true},
			{Key: "PGDATABASE", Ref: ref("PGDATABASE"), AtRuntime: true, DefaultOn: false},
		}
	case "redis":
		return []SuggestedVariable{
			{Key: "REDIS_URL", Ref: ref("REDIS_URL"), AtBuild: true, AtRuntime: true, DefaultOn: true, IsSecret: true},
			{Key: "REDIS_HOST", Ref: ref("REDIS_HOST"), AtRuntime: true, DefaultOn: false},
			{Key: "REDIS_PORT", Ref: ref("REDIS_PORT"), AtRuntime: true, DefaultOn: false},
			{Key: "REDIS_PASSWORD", Ref: ref("REDIS_PASSWORD"), AtRuntime: true, DefaultOn: false, IsSecret: true},
		}
	case "bucket":
		return []SuggestedVariable{
			{Key: "S3_ENDPOINT", Ref: ref("S3_ENDPOINT"), AtRuntime: true, DefaultOn: true},
			{Key: "S3_REGION", Ref: ref("S3_REGION"), AtRuntime: true, DefaultOn: true},
			{Key: "S3_BUCKET", Ref: ref("S3_BUCKET"), AtRuntime: true, DefaultOn: true},
			{Key: "S3_ACCESS_KEY_ID", Ref: ref("S3_ACCESS_KEY_ID"), AtRuntime: true, DefaultOn: true, IsSecret: true},
			{Key: "S3_SECRET_ACCESS_KEY", Ref: ref("S3_SECRET_ACCESS_KEY"), AtRuntime: true, DefaultOn: true, IsSecret: true},
		}
	case "app":
		return []SuggestedVariable{
			{Key: fmt.Sprintf("%s_URL", upperSlug(providerSlug)), Ref: ref("PRIVATE_HOST"), AtRuntime: true, DefaultOn: true},
		}
	default:
		return nil
	}
}

func upperSlug(slug string) string {
	out := make([]byte, len(slug))
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		} else if c == '-' {
			c = '_'
		}
		out[i] = c
	}
	return string(out)
}
