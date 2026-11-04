package staticsite

import (
	"path"
	"strings"
)

// IndexFileNames returns Caddy index_names for the site config.
func IndexFileNames(cfg Config) []string {
	name := strings.TrimSpace(cfg.IndexFile)
	if name == "" {
		return []string{"index.html"}
	}
	name = strings.TrimPrefix(name, "/")
	return []string{path.Base(name)}
}
