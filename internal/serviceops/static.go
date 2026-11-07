package serviceops

import (
	"os"

	"github.com/Joeydekruis/dockyrd/internal/staticsite"
)

// TeardownStaticSite removes on-disk release directories for a static site service.
func TeardownStaticSite(sitesRoot, serviceSlug string) error {
	if sitesRoot == "" || serviceSlug == "" {
		return nil
	}
	paths := staticsite.PathsForService(sitesRoot, serviceSlug)
	return os.RemoveAll(paths.ServiceRoot)
}
