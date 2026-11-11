package diskusage

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/staticsite"
)

// Category is one row in the data-directory breakdown.
type Category struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Bytes int64  `json:"bytes"`
}

// ServiceRow attributes on-disk usage to one deployable service.
type ServiceRow struct {
	ServiceID string   `json:"service_id"`
	Name      string   `json:"name"`
	Slug      string   `json:"slug"`
	Kind      string   `json:"kind"`
	Bytes     int64    `json:"bytes"`
	Sources   []string `json:"sources,omitempty"`
}

// VolumeRow is a Docker named volume with optional service link.
type VolumeRow struct {
	Name      string `json:"name"`
	Bytes     int64  `json:"bytes"`
	ServiceID string `json:"service_id,omitempty"`
	Slug      string `json:"slug,omitempty"`
}

// LogFile is a large log under the data directory.
type LogFile struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

// DockerUsage summarizes Docker engine disk use (may live off the data volume).
type DockerUsage struct {
	Available        bool  `json:"docker_available"`
	ImagesBytes      int64 `json:"images_bytes"`
	ContainersBytes  int64 `json:"containers_bytes"`
	VolumesBytes     int64 `json:"volumes_bytes"`
	BuildCacheBytes  int64 `json:"build_cache_bytes"`
	DanglingImages   int64 `json:"dangling_images_bytes"`
	Volumes          []VolumeRow `json:"volumes,omitempty"`
}

// Report is the full disk inventory for the analytics Disk tab.
type Report struct {
	DataDir    string     `json:"data_dir"`
	ScannedAt  int64      `json:"scanned_at"`
	Categories []Category `json:"categories"`
	Services   []ServiceRow `json:"services"`
	Docker     DockerUsage `json:"docker"`
	LogFiles   []LogFile  `json:"log_files"`
	OtherBytes int64      `json:"other_bytes"`
}

var categoryLabels = map[string]string{
	"control.db": "Control plane database",
	"caddy":      "Proxy, TLS & access logs",
	"sites":      "Static site releases",
	"compose":    "Compose stack data",
	"builds":     "Build workspaces",
	"backups":    "Database backups",
	"garage":     "Object storage",
	"builders":   "Remote builders",
	"host":       "Host agent data",
}

// BuildReport scans cfg.DataDir and optional Docker usage.
func BuildReport(ctx context.Context, cfg config.Config, q *sqlc.Queries, cli *docker.Client) Report {
	root := cfg.DataDir
	report := Report{
		DataDir:   root,
		ScannedAt: 0,
	}
	report.Categories = dataDirCategories(root)
	report.LogFiles = findLargeLogs(cfg)
	report.Docker = dockerBreakdown(ctx, cli, q)

	if q != nil {
		report.Services = serviceUsage(ctx, cfg, q, report.Docker.Volumes)
	}
	report.OtherBytes = uncategorizedBytes(root, report.Categories)
	return report
}

func dataDirCategories(root string) []Category {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var out []Category
	for _, e := range entries {
		name := e.Name()
		path := filepath.Join(root, name)
		var bytes int64
		if e.IsDir() {
			bytes = DirSize(path)
		} else {
			bytes = FileSize(path)
		}
		if bytes <= 0 {
			continue
		}
		label := categoryLabels[name]
		id := name
		if label == "" {
			label = "Other: " + name
			id = "other:" + name
		}
		out = append(out, Category{ID: id, Label: label, Bytes: bytes})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Bytes > out[j].Bytes })
	return out
}

func uncategorizedBytes(root string, cats []Category) int64 {
	var known int64
	for _, c := range cats {
		if strings.HasPrefix(c.ID, "other:") {
			continue
		}
		known += c.Bytes
	}
	total := DirSize(root)
	if total < known {
		return 0
	}
	return total - known
}

func findLargeLogs(cfg config.Config) []LogFile {
	var files []LogFile
	caddyDir := cfg.CaddyDir()
	_ = filepath.WalkDir(caddyDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".log") {
			return nil
		}
		size := FileSize(path)
		if size <= 0 {
			return nil
		}
		rel, _ := filepath.Rel(cfg.DataDir, path)
		if rel == "" || strings.HasPrefix(rel, "..") {
			rel = path
		}
		files = append(files, LogFile{Path: rel, Bytes: size})
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].Bytes > files[j].Bytes })
	if len(files) > 8 {
		files = files[:8]
	}
	return files
}

func dockerBreakdown(ctx context.Context, cli *docker.Client, q *sqlc.Queries) DockerUsage {
	out := DockerUsage{}
	if cli == nil {
		return out
	}
	usage, err := cli.EngineDiskUsage(ctx)
	if err != nil {
		return out
	}
	out.Available = true
	slugToID := map[string]string{}
	if q != nil {
		rows, err := q.ListServicesDiskMeta(ctx)
		if err == nil {
			for _, row := range rows {
				slugToID[row.Slug] = row.ID
			}
		}
	}
	for _, img := range usage.Images {
		out.ImagesBytes += img.Size
		if len(img.RepoTags) == 0 && len(img.RepoDigests) == 0 {
			out.DanglingImages += img.Size
		}
	}
	for _, c := range usage.Containers {
		out.ContainersBytes += c.SizeRw
		if c.SizeRootFs > 0 {
			out.ContainersBytes += c.SizeRootFs
		}
	}
	for _, rec := range usage.BuildCache {
		out.BuildCacheBytes += rec.Size
	}
	for _, vol := range usage.Volumes {
		out.VolumesBytes += vol.UsageData.Size
		row := VolumeRow{Name: vol.Name, Bytes: vol.UsageData.Size}
		for slug, id := range slugToID {
			if vol.Name == docker.VolumeName(slug) {
				row.ServiceID = id
				row.Slug = slug
				break
			}
		}
		out.Volumes = append(out.Volumes, row)
	}
	sort.Slice(out.Volumes, func(i, j int) bool { return out.Volumes[i].Bytes > out.Volumes[j].Bytes })
	if len(out.Volumes) > 12 {
		out.Volumes = out.Volumes[:12]
	}
	return out
}

func serviceUsage(ctx context.Context, cfg config.Config, q *sqlc.Queries, vols []VolumeRow) []ServiceRow {
	rows, err := q.ListServicesDiskMeta(ctx)
	if err != nil {
		return nil
	}
	volBySlug := map[string]int64{}
	for _, v := range vols {
		if v.Slug != "" {
			volBySlug[v.Slug] += v.Bytes
		}
	}
	var out []ServiceRow
	for _, svc := range rows {
		var total int64
		var sources []string
		composeDir := filepath.Join(cfg.DataDir, "compose", svc.ID)
		if n := DirSize(composeDir); n > 0 {
			total += n
			sources = append(sources, "compose data")
		}
		if svc.Kind == "static_site" || svc.Kind == "app" {
			siteDir := staticsite.PathsForService(cfg.SitesDir(), svc.Slug).ServiceRoot
			if n := DirSize(siteDir); n > 0 {
				total += n
				sources = append(sources, "static releases")
			}
		}
		if n := volBySlug[svc.Slug]; n > 0 {
			total += n
			sources = append(sources, "database volume")
		}
		if total <= 0 {
			continue
		}
		out = append(out, ServiceRow{
			ServiceID: svc.ID,
			Name:      svc.Name,
			Slug:      svc.Slug,
			Kind:      svc.Kind,
			Bytes:     total,
			Sources:   sources,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Bytes > out[j].Bytes })
	if len(out) > 20 {
		out = out[:20]
	}
	return out
}
