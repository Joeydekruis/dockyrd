package docker

import (
	"regexp"
	"strings"
)

const (
	EdgeNetworkName = "dy-edge"
	projectNetPrefix = "dy-p-"
	linkNetPrefix    = "dy-l-"
)

var netSanitize = regexp.MustCompile(`[^a-z0-9-]+`)

// ProjectNetworkName returns the isolated Docker network for a project slug.
func ProjectNetworkName(projectSlug string) string {
	s := strings.ToLower(strings.TrimSpace(projectSlug))
	s = netSanitize.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "default"
	}
	if len(s) > 32 {
		s = s[:32]
	}
	return projectNetPrefix + s
}

// LinkNetworkName returns the per-link bridge network name.
func LinkNetworkName(linkID string) string {
	short := DeployShort(linkID)
	if short == "" {
		short = "link"
	}
	return linkNetPrefix + short
}

// AppContainerName is dy-<service_slug>-<deploy_short>.
func AppContainerName(serviceSlug, deployShort string) string {
	return "dy-" + sanitizeName(serviceSlug) + "-" + sanitizeName(deployShort)
}

// ServiceContainerName is the fixed database-style name dy-<service_slug>.
func ServiceContainerName(serviceSlug string) string {
	return "dy-" + sanitizeName(serviceSlug)
}

// VolumeName is dy-<service_short>-data.
func VolumeName(serviceShort string) string {
	return "dy-" + sanitizeName(serviceShort) + "-data"
}

// DeployShort returns the first 8 characters of a deployment ULID for naming.
func DeployShort(deploymentID string) string {
	if len(deploymentID) <= 8 {
		return deploymentID
	}
	return deploymentID[:8]
}

var nameSanitize = regexp.MustCompile(`[^a-z0-9-]+`)

func sanitizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nameSanitize.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "svc"
	}
	if len(s) > 48 {
		return s[:48]
	}
	return s
}
