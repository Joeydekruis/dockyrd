package staticsite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var defaultOutputCandidates = []string{"dist", "out", "build", "public", "_site"}

// DetectOutputDir finds build output under workDir or uses explicit relative dir.
func DetectOutputDir(workDir, explicit string) (string, error) {
	explicit = strings.TrimSpace(explicit)
	if explicit == "." {
		if plainSiteRoot(workDir) {
			return workDir, nil
		}
		return "", fmt.Errorf("output dir \".\" requires index.html in the project root")
	}
	if explicit != "" {
		dir := filepath.Join(workDir, explicit)
		if statOK(dir) {
			return dir, nil
		}
		return "", fmt.Errorf("output dir %q missing after build", explicit)
	}
	for _, name := range defaultOutputCandidates {
		dir := filepath.Join(workDir, name)
		if statOK(dir) {
			return dir, nil
		}
	}
	if plainSiteRoot(workDir) {
		return workDir, nil
	}
	return "", fmt.Errorf("no output directory found (tried %v)", defaultOutputCandidates)
}

func plainSiteRoot(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "index.html"))
	return err == nil && !info.IsDir()
}

func statOK(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}
