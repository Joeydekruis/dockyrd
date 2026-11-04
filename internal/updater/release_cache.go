package updater

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	// AutoRefreshInterval is how often the panel polls GitHub in the background.
	AutoRefreshInterval = 5 * time.Minute
	// MaxManualRefreshPerHour limits "Check again" when using unauthenticated GitHub API.
	MaxManualRefreshPerHour = 40
)

// DefaultReleaseCache is shared by HTTP handlers and the background refresh loop.
var DefaultReleaseCache = NewReleaseCache()

// ReleaseSnapshot is a cached view of the last GitHub release check.
type ReleaseSnapshot struct {
	Result            CheckResult
	Err               error
	CheckedAt         time.Time
	HasData           bool
	FromCache         bool
	ManualRateLimited bool
}

// ReleaseCache coalesces GitHub release API calls to stay within unauthenticated rate limits.
type ReleaseCache struct {
	mu              sync.RWMutex
	result          CheckResult
	err             error
	checkedAt       time.Time
	hasData         bool
	manualRefreshAt []time.Time
}

func NewReleaseCache() *ReleaseCache {
	return &ReleaseCache{}
}

// Snapshot returns the last check without calling GitHub.
func (c *ReleaseCache) Snapshot() ReleaseSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return ReleaseSnapshot{
		Result:    c.result,
		Err:       c.err,
		CheckedAt: c.checkedAt,
		HasData:   c.hasData,
		FromCache: true,
	}
}

// RefreshScheduled runs from the background ticker (≈12 GitHub calls/hour).
func (c *ReleaseCache) RefreshScheduled(ctx context.Context, current string, opts Options) error {
	return c.refresh(ctx, current, opts)
}

// RefreshManual is used by "Check again" (rate-limited without a GitHub token).
func (c *ReleaseCache) RefreshManual(ctx context.Context, current string, opts Options) ReleaseSnapshot {
	if limited, retryAfter := c.wouldExceedManualLimit(); limited {
		snap := c.Snapshot()
		snap.ManualRateLimited = true
		snap.Err = fmt.Errorf(
			"manual update check limit reached (%d per hour without a GitHub token). Try again in %s or set DOCKYARD_GITHUB_TOKEN in dockyard.env",
			MaxManualRefreshPerHour,
			retryAfter.Round(time.Minute),
		)
		return snap
	}
	c.recordManualRefresh()
	err := c.refresh(ctx, current, opts)
	snap := c.Snapshot()
	snap.FromCache = false
	if err != nil {
		snap.Err = err
	}
	return snap
}

// RefreshForce bypasses manual limits (e.g. immediately before apply).
func (c *ReleaseCache) RefreshForce(ctx context.Context, current string, opts Options) error {
	return c.refresh(ctx, current, opts)
}

func (c *ReleaseCache) refresh(ctx context.Context, current string, opts Options) error {
	check, err := Check(ctx, current, opts)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.result = check
	c.err = err
	c.checkedAt = time.Now()
	c.hasData = true
	return err
}

func (c *ReleaseCache) wouldExceedManualLimit() (bool, time.Duration) {
	if GitHubToken() != "" {
		return false, 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pruneManualRefreshAtLocked(time.Now())
	if len(c.manualRefreshAt) >= MaxManualRefreshPerHour {
		oldest := c.manualRefreshAt[0]
		retry := time.Until(oldest.Add(time.Hour))
		if retry < 0 {
			retry = 0
		}
		return true, retry
	}
	return false, 0
}

func (c *ReleaseCache) recordManualRefresh() {
	if GitHubToken() != "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	c.pruneManualRefreshAtLocked(now)
	c.manualRefreshAt = append(c.manualRefreshAt, now)
}

func (c *ReleaseCache) pruneManualRefreshAtLocked(now time.Time) {
	cutoff := now.Add(-time.Hour)
	i := 0
	for _, t := range c.manualRefreshAt {
		if t.After(cutoff) {
			c.manualRefreshAt[i] = t
			i++
		}
	}
	c.manualRefreshAt = c.manualRefreshAt[:i]
}

// StartAutoRefresh polls GitHub on a fixed interval for the lifetime of ctx.
func StartAutoRefresh(ctx context.Context, interval time.Duration, current func() string, opts func() Options) {
	if interval <= 0 {
		interval = AutoRefreshInterval
	}
	go func() {
		run := func() {
			cur := current()
			o := opts()
			if err := DefaultReleaseCache.RefreshScheduled(ctx, cur, o); err != nil {
				// Errors are stored on the cache for API consumers.
				_ = err
			}
		}
		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
