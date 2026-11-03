package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const defaultRepo = "Joeydekruis/dockyrd"

// Release describes an installable GitHub release asset.
type Release struct {
	Tag         string
	AssetURL    string
	AssetName   string
	ChecksumURL string
}

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// FetchLatest loads the latest stable release and matching binary asset for this OS/arch.
func FetchLatest(ctx context.Context, repo, currentVersion string) (Release, bool, error) {
	if repo == "" {
		repo = defaultRepo
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "dockyard-updater")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return Release{}, false, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return Release{}, false, fmt.Errorf("github releases: %s: %s", res.Status, strings.TrimSpace(string(body)))
	}
	var payload ghRelease
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return Release{}, false, err
	}
	tag := normalizeTag(payload.TagName)
	wantPrefix := fmt.Sprintf("dockyard_%s_%s_%s", strings.TrimPrefix(tag, "v"), runtime.GOOS, runtime.GOARCH)
	var rel Release
	rel.Tag = tag
	for _, a := range payload.Assets {
		if strings.HasPrefix(a.Name, wantPrefix) && strings.HasSuffix(a.Name, ".tar.gz") {
			rel.AssetName = a.Name
			rel.AssetURL = a.BrowserDownloadURL
		}
		if a.Name == "checksums.txt" {
			rel.ChecksumURL = a.BrowserDownloadURL
		}
	}
	if rel.AssetURL == "" {
		return Release{}, false, fmt.Errorf("no release asset matching %s", wantPrefix)
	}
	newer := CompareVersions(currentVersion, tag) < 0
	return rel, newer, nil
}
