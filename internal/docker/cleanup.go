package docker

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	"github.com/Joeydekruis/dockyrd/internal/staticsite"
)

var ErrUnavailable = errors.New("docker unavailable")

// CleanupSummary estimates reclaimable Docker resources (best-effort).
type CleanupSummary struct {
	BuildCacheBytes       int64 `json:"build_cache_bytes"`
	DanglingImagesBytes   int64 `json:"dangling_images_bytes"`
	StoppedUnmanagedCount int   `json:"stopped_unmanaged_containers"`
	OldReleaseBytes       int64 `json:"old_releases_bytes"`
	DockerAvailable       bool  `json:"docker_available"`
}

// ContainerRow is a managed container for the Analytics table.
type ContainerRow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	Kind      string `json:"kind,omitempty"`
	ServiceID string `json:"service_id,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
}

// CleanupSummary collects disk usage and stopped unmanaged container count.
func (c *Client) CleanupSummary(ctx context.Context) (CleanupSummary, error) {
	if c == nil || c.moby == nil {
		return CleanupSummary{}, ErrUnavailable
	}
	out := CleanupSummary{DockerAvailable: true}
	usage, err := c.moby.DiskUsage(ctx, types.DiskUsageOptions{})
	if err == nil {
		for _, rec := range usage.BuildCache {
			out.BuildCacheBytes += rec.Size
		}
		for _, img := range usage.Images {
			if len(img.RepoTags) == 0 && len(img.RepoDigests) == 0 {
				out.DanglingImagesBytes += img.Size
			}
		}
	}
	stopped, err := c.countStoppedUnmanaged(ctx)
	if err == nil {
		out.StoppedUnmanagedCount = stopped
	}
	return out, nil
}

func (c *Client) countStoppedUnmanaged(ctx context.Context) (int, error) {
	list, err := c.moby.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("status", "exited")),
	})
	if err != nil {
		return 0, err
	}
	n := 0
	for _, item := range list {
		if item.Labels[LabelManaged] == managedTrue {
			continue
		}
		n++
	}
	return n, nil
}

// ListContainerRows returns managed containers.
func (c *Client) ListContainerRows(ctx context.Context) ([]ContainerRow, error) {
	managed, err := c.ListManaged(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ContainerRow, 0, len(managed))
	for _, m := range managed {
		id := m.ID
		if len(id) > 12 {
			id = id[:12]
		}
		out = append(out, ContainerRow{
			ID:        id,
			Name:      m.Name,
			State:     m.State,
			Kind:      m.Identity.Kind,
			ServiceID: m.Identity.ServiceID,
			CreatedAt: m.Created,
		})
	}
	return out, nil
}

// RunCleanup prunes selected targets; never removes running managed workloads.
func (c *Client) RunCleanup(ctx context.Context, sitesRoot string, targets []string) (map[string]int64, error) {
	if c == nil || c.moby == nil {
		return nil, ErrUnavailable
	}
	results := map[string]int64{}
	for _, t := range targets {
		switch t {
		case "dangling_images":
			report, err := c.moby.ImagesPrune(ctx, filters.NewArgs(filters.Arg("dangling", "true")))
			if err != nil {
				return results, err
			}
			results[t] = int64(report.SpaceReclaimed)
		case "build_cache":
			report, err := c.moby.BuildCachePrune(ctx, types.BuildCachePruneOptions{All: false})
			if err != nil {
				return results, err
			}
			results[t] = int64(report.SpaceReclaimed)
		case "stopped_unmanaged":
			n, err := c.pruneStoppedUnmanaged(ctx)
			if err != nil {
				return results, err
			}
			results[t] = n
		case "old_releases":
			results[t] = staticsite.PruneAllOldReleases(sitesRoot, staticsite.DefaultReleaseKeep)
		}
	}
	return results, nil
}

func (c *Client) pruneStoppedUnmanaged(ctx context.Context) (int64, error) {
	list, err := c.moby.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return 0, err
	}
	var removed int64
	for _, item := range list {
		if item.State == "running" {
			continue
		}
		if item.Labels[LabelManaged] == managedTrue {
			continue
		}
		if err := c.moby.ContainerRemove(ctx, item.ID, container.RemoveOptions{Force: true}); err != nil {
			continue
		}
		removed++
	}
	return removed, nil
}
