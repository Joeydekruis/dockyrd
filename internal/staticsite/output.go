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
	if explicit = strings.TrimSpace(explicit); explicit != "" && explicit != "." {
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
	return "", fmt.Errorf("no output directory found (tried %v)", defaultOutputCandidates)
}

func statOK(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}
