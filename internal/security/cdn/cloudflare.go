package cdn

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	cloudflareIPv4URL = "https://www.cloudflare.com/ips-v4"
	cloudflareIPv6URL = "https://www.cloudflare.com/ips-v6"
)

// FetchCloudflareRanges returns published Cloudflare edge CIDRs (v4 + v6).
func FetchCloudflareRanges(ctx context.Context, client *http.Client) ([]string, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	var out []string
	for _, u := range []string{cloudflareIPv4URL, cloudflareIPv6URL} {
		ranges, err := fetchPlainLines(ctx, client, u)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", u, err)
		}
		out = append(out, ranges...)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("cloudflare returned no ranges")
	}
	return out, nil
}

func fetchPlainLines(ctx context.Context, client *http.Client, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", res.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines, nil
}
