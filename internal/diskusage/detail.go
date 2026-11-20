package diskusage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/metrics"
	"github.com/Joeydekruis/dockyrd/internal/staticsite"
)

// DetailEntry is one row in a drill-down listing.
type DetailEntry struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Bytes     int64  `json:"bytes"`
	IsDir     bool   `json:"is_dir"`
	Drillable bool   `json:"drillable"`
	Hint      string `json:"hint,omitempty"`
}

// DetailReport is returned by the disk-usage detail API.
type DetailReport struct {
	Title    string        `json:"title"`
	Scope    string        `json:"scope"`
	Path     string        `json:"path,omitempty"`
	Bytes    int64         `json:"bytes"`
	Children []DetailEntry   `json:"children,omitempty"`
	Note     string        `json:"note,omitempty"`
}

// VolumeSummary reconciles filesystem usage vs measured data dir vs Docker.
type VolumeSummary struct {
	FilesystemUsed   int64  `json:"filesystem_used"`
	FilesystemTotal  int64  `json:"filesystem_total"`
	DataDirBytes     int64  `json:"data_dir_bytes"`
	DockerRoot       string `json:"docker_root,omitempty"`
	DockerOnVolume   bool   `json:"docker_on_data_volume"`
	DockerBytes      int64  `json:"docker_bytes"`
	OutsideDataBytes int64  `json:"outside_data_dir_bytes"`
}

func volumeSummary(ctx context.Context, cfg config.Config, cli *docker.Client, du DockerUsage) VolumeSummary {
	out := VolumeSummary{
		DataDirBytes: DirSize(cfg.DataDir),
		DockerBytes: du.ImagesBytes + du.ContainersBytes + du.VolumesBytes + du.BuildCacheBytes,
	}
	if sample, err := metrics.SampleHost(cfg.DataDir); err == nil {
		out.FilesystemUsed = sample.DiskUsed
		out.FilesystemTotal = sample.DiskTotal
	}
	if cli != nil {
		root, err := cli.DockerRootDir(ctx)
		if err == nil && root != "" {
			out.DockerRoot = root
			same, err := sameFilesystem(cfg.DataDir, root)
			if err == nil {
				out.DockerOnVolume = same
			}
		}
	}
	if out.FilesystemUsed > out.DataDirBytes {
		out.OutsideDataBytes = out.FilesystemUsed - out.DataDirBytes
	}
	return out
}

// Detail resolves a drill-down scope.
func Detail(ctx context.Context, cfg config.Config, q *sqlc.Queries, cli *docker.Client, scope, path, serviceID string) (DetailReport, error) {
	switch scope {
	case "data":
		return dataDetail(cfg, path)
	case "docker":
		return dockerDetail(ctx, cli, q, path)
	case "service":
		return serviceDetail(ctx, cfg, q, cli, serviceID)
	default:
		return DetailReport{}, fmt.Errorf("unknown scope")
	}
}

func safeRelPath(dataDir, rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	rel = strings.Trim(rel, "/")
	if rel == "" {
		return ".", nil
	}
	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("invalid path")
	}
	abs := filepath.Join(dataDir, rel)
	absClean := filepath.Clean(abs)
	rootClean := filepath.Clean(dataDir)
	if absClean != rootClean && !strings.HasPrefix(absClean, rootClean+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid path")
	}
	return rel, nil
}

func dataDetail(cfg config.Config, rel string) (DetailReport, error) {
	rel, err := safeRelPath(cfg.DataDir, rel)
	if err != nil {
		return DetailReport{}, err
	}
	abs := cfg.DataDir
	if rel != "." {
		abs = filepath.Join(cfg.DataDir, rel)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return DetailReport{}, err
	}
	report := DetailReport{
		Scope: "data",
		Path:  rel,
		Bytes: fileOrDirSize(abs, info.IsDir()),
	}
	report.Title = filepath.Base(abs)
	if rel == "." {
		report.Title = cfg.DataDir
	}
	if categoryLabels[filepath.Base(abs)] != "" {
		report.Title = categoryLabels[filepath.Base(abs)]
	}

	if !info.IsDir() {
		report.Note = "Single file — there is nothing further to expand on disk."
		return report, nil
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return DetailReport{}, err
	}
	for _, e := range entries {
		childRel := rel
		if childRel == "." {
			childRel = e.Name()
		} else {
			childRel = filepath.Join(rel, e.Name())
		}
		childAbs := filepath.Join(cfg.DataDir, childRel)
		var bytes int64
		if e.IsDir() {
			bytes = DirSize(childAbs)
		} else {
			bytes = FileSize(childAbs)
		}
		hint := ""
		if e.IsDir() && bytes == 0 {
			hint = "empty"
		}
		report.Children = append(report.Children, DetailEntry{
			Name:      e.Name(),
			Path:      childRel,
			Bytes:     bytes,
			IsDir:     e.IsDir(),
			Drillable: e.IsDir() && bytes > 0,
			Hint:      hint,
		})
	}
	sort.Slice(report.Children, func(i, j int) bool {
		return report.Children[i].Bytes > report.Children[j].Bytes
	})
	if len(report.Children) > 50 {
		report.Children = report.Children[:50]
		report.Note = "Showing the 50 largest entries in this folder."
	}
	return report, nil
}

func fileOrDirSize(path string, isDir bool) int64 {
	if isDir {
		return DirSize(path)
	}
	return FileSize(path)
}

func dockerDetail(ctx context.Context, cli *docker.Client, q *sqlc.Queries, kind string) (DetailReport, error) {
	if cli == nil {
		return DetailReport{}, fmt.Errorf("docker unavailable")
	}
	usage, err := cli.EngineDiskUsage(ctx)
	if err != nil {
		return DetailReport{}, err
	}
	kind = strings.TrimSpace(strings.ToLower(kind))
	switch kind {
	case "images":
		return dockerImagesDetail(usage), nil
	case "containers":
		return dockerContainersDetail(usage), nil
	case "volumes":
		return dockerVolumesDetail(ctx, usage, q), nil
	case "build_cache":
		return dockerBuildCacheDetail(usage), nil
	default:
		return DetailReport{}, fmt.Errorf("unknown docker kind")
	}
}

func dockerImagesDetail(usage types.DiskUsage) DetailReport {
	report := DetailReport{
		Scope: "docker",
		Path:  "images",
		Title: "Docker images",
		Bytes: 0,
		Note:  "Image layers shared between tags are counted once in Docker's total.",
	}
	imgs := append([]*image.Summary(nil), usage.Images...)
	sort.Slice(imgs, func(i, j int) bool { return imgs[i].Size > imgs[j].Size })
	for _, img := range imgs {
		report.Bytes += img.Size
		if len(report.Children) >= 40 {
			continue
		}
		name := strings.Join(img.RepoTags, ", ")
		if name == "" {
			name = strings.Join(img.RepoDigests, ", ")
		}
		if name == "" {
			name = img.ID[:min(12, len(img.ID))]
		}
		hint := ""
		if len(img.RepoTags) == 0 && len(img.RepoDigests) == 0 {
			hint = "dangling"
		}
		report.Children = append(report.Children, DetailEntry{
			Name:  name,
			Bytes: img.Size,
			Hint:  hint,
		})
	}
	if len(imgs) > 40 {
		report.Note += " Showing top 40 images by size."
	}
	return report
}

func dockerContainersDetail(usage types.DiskUsage) DetailReport {
	report := DetailReport{
		Scope: "docker",
		Path:  "containers",
		Title: "Docker containers",
		Note:  "Writable layer size per container (excludes shared image layers).",
	}
	type row struct {
		name string
		id   string
		size int64
	}
	var rows []row
	for _, c := range usage.Containers {
		sz := c.SizeRw
		if c.SizeRootFs > 0 {
			sz += c.SizeRootFs
		}
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		rows = append(rows, row{name: name, id: c.ID, size: sz})
		report.Bytes += sz
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].size > rows[j].size })
	for i, r := range rows {
		if i >= 40 {
			break
		}
		idShort := r.id
		if len(idShort) > 12 {
			idShort = idShort[:12]
		}
		report.Children = append(report.Children, DetailEntry{
			Name:  r.name,
			Path:  idShort,
			Bytes: r.size,
			Hint:  r.id,
		})
	}
	return report
}

func dockerVolumesDetail(ctx context.Context, usage types.DiskUsage, q *sqlc.Queries) DetailReport {
	report := DetailReport{
		Scope: "docker",
		Path:  "volumes",
		Title: "Docker volumes",
		Note:  "Named volumes live under Docker's data root, not inside the dockyrd data folder.",
	}
	slugToID := map[string]string{}
	if q != nil {
		rows, _ := q.ListServicesDiskMeta(ctx)
		for _, row := range rows {
			slugToID[row.Slug] = row.ID
		}
	}
	vols := append([]*volume.Volume(nil), usage.Volumes...)
	sort.Slice(vols, func(i, j int) bool {
		return vols[i].UsageData.Size > vols[j].UsageData.Size
	})
	for _, vol := range vols {
		sz := vol.UsageData.Size
		report.Bytes += sz
		hint := ""
		for slug := range slugToID {
			if vol.Name == docker.VolumeName(slug) {
				hint = "service: " + slug
				break
			}
		}
		if strings.Contains(vol.Name, "buildkit") {
			hint = "BuildKit builder state"
		}
		report.Children = append(report.Children, DetailEntry{
			Name: vol.Name,
			Bytes: sz,
			Hint: hint,
		})
	}
	return report
}

func dockerBuildCacheDetail(usage types.DiskUsage) DetailReport {
	report := DetailReport{
		Scope: "docker",
		Path:  "build_cache",
		Title: "Docker build cache",
	}
	for _, rec := range usage.BuildCache {
		report.Bytes += rec.Size
		report.Children = append(report.Children, DetailEntry{
			Name:  rec.ID[:min(12, len(rec.ID))],
			Bytes: rec.Size,
			Hint:  rec.Description,
		})
	}
	sort.Slice(report.Children, func(i, j int) bool {
		return report.Children[i].Bytes > report.Children[j].Bytes
	})
	if len(report.Children) > 30 {
		report.Children = report.Children[:30]
	}
	return report
}

func serviceDetail(ctx context.Context, cfg config.Config, q *sqlc.Queries, cli *docker.Client, serviceID string) (DetailReport, error) {
	serviceID = strings.TrimSpace(serviceID)
	if serviceID == "" || q == nil {
		return DetailReport{}, fmt.Errorf("service_id required")
	}
	svc, err := q.GetServiceByID(ctx, serviceID)
	if err != nil {
		return DetailReport{}, err
	}
	report := DetailReport{
		Scope: "service",
		Path:  serviceID,
		Title: svc.Name,
		Note:  "Measured under " + cfg.DataDir + " and linked Docker volumes.",
	}
	composeDir := filepath.Join(cfg.DataDir, "compose", svc.ID)
	if n := DirSize(composeDir); n > 0 {
		report.Bytes += n
		report.Children = append(report.Children, DetailEntry{
			Name:      "Compose stack data",
			Path:      filepath.Join("compose", svc.ID),
			Bytes:     n,
			IsDir:     true,
			Drillable: true,
		})
		appendComposeChildren(cfg, composeDir, &report)
	}
	siteRoot := staticsite.PathsForService(cfg.SitesDir(), svc.Slug).ServiceRoot
	if n := DirSize(siteRoot); n > 0 {
		report.Bytes += n
		releaseDir := filepath.Join(siteRoot, "releases")
		if releases := DirSize(releaseDir); releases > 0 {
			report.Children = append(report.Children, DetailEntry{
				Name:      "Static releases",
				Path:      filepath.Join("sites", svc.Slug, "releases"),
				Bytes:     releases,
				IsDir:     true,
				Drillable: true,
			})
		}
	}
	if cli != nil {
		usage, err := cli.EngineDiskUsage(ctx)
		if err == nil {
			volName := docker.VolumeName(svc.Slug)
			for _, vol := range usage.Volumes {
				if vol.Name == volName {
					sz := vol.UsageData.Size
					report.Bytes += sz
					report.Children = append(report.Children, DetailEntry{
						Name:  "Database volume (" + volName + ")",
						Bytes: sz,
						Hint:  "Docker named volume",
					})
				}
			}
		}
	}
	sort.Slice(report.Children, func(i, j int) bool {
		return report.Children[i].Bytes > report.Children[j].Bytes
	})
	return report, nil
}

func appendComposeChildren(cfg config.Config, composeDir string, report *DetailReport) {
	entries, err := os.ReadDir(composeDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		child := filepath.Join(composeDir, e.Name())
		n := DirSize(child)
		if n <= 0 {
			continue
		}
		rel, _ := filepath.Rel(cfg.DataDir, child)
		report.Children = append(report.Children, DetailEntry{
			Name:      e.Name(),
			Path:      rel,
			Bytes:     n,
			IsDir:     true,
			Drillable: true,
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
