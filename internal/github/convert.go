package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ConvertManifest exchanges a manifest code for app credentials.
func ConvertManifest(ctx context.Context, code string) (ConversionResult, error) {
	code = stringsTrim(code)
	if code == "" {
		return ConversionResult{}, fmt.Errorf("manifest code required")
	}
	url := "https://api.github.com/app-manifests/" + code + "/conversions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return ConversionResult{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ConversionResult{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return ConversionResult{}, fmt.Errorf("github conversion %d: %s", resp.StatusCode, string(body))
	}
	var res ConversionResult
	if err := json.Unmarshal(body, &res); err != nil {
		return ConversionResult{}, err
	}
	return res, nil
}
