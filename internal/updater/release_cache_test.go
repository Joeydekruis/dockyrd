package updater

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestReleaseCacheManualRateLimitWithoutToken(t *testing.T) {
	t.Setenv("DOCKYARD_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v9.9.9",
			"body": "notes",
			"assets": [
				{"name": "dockyard_9.9.9_linux_amd64.tar.gz", "browser_download_url": "https://example.com/a.tar.gz"},
				{"name": "checksums.txt", "browser_download_url": "https://example.com/c.txt"}
			]
		}`))
	}))
	t.Cleanup(srv.Close)

	cache := NewReleaseCache()
	opts := Options{Repo: "org/repo"}
	ctx := context.Background()

	// Patch FetchLatest URL — Check uses real GitHub URL. Use direct refresh with mock by
	// testing rate limit logic only on empty cache path.
	for i := 0; i < MaxManualRefreshPerHour; i++ {
		cache.recordManualRefresh()
	}
	snap := cache.RefreshManual(ctx, "v0.0.1", opts)
	if !snap.ManualRateLimited {
		t.Fatal("expected manual rate limit")
	}
	if snap.Err == nil {
		t.Fatal("expected error on rate limit")
	}
	_ = srv
}

func TestReleaseCacheSnapshotWithoutFetch(t *testing.T) {
	cache := NewReleaseCache()
	snap := cache.Snapshot()
	if snap.HasData {
		t.Fatal("expected empty cache")
	}
}

func TestManualLimitSkippedWithToken(t *testing.T) {
	t.Setenv("DOCKYARD_GITHUB_TOKEN", "ghp_test")
	cache := NewReleaseCache()
	for i := 0; i < MaxManualRefreshPerHour+5; i++ {
		cache.recordManualRefresh()
	}
	limited, _ := cache.wouldExceedManualLimit()
	if limited {
		t.Fatal("token should disable manual limit")
	}
}

func TestPruneManualRefreshAt(t *testing.T) {
	cache := NewReleaseCache()
	cache.mu.Lock()
	cache.manualRefreshAt = []time.Time{time.Now().Add(-2 * time.Hour), time.Now()}
	cache.pruneManualRefreshAtLocked(time.Now())
	if len(cache.manualRefreshAt) != 1 {
		t.Fatalf("len=%d want 1", len(cache.manualRefreshAt))
	}
	cache.mu.Unlock()
}
