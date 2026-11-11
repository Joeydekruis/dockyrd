package project

const (
	// UnassignedName is the internal project name for resources created without a project.
	UnassignedName = "Unassigned"
	// UnassignedSlug is the stable slug used to locate the bucket project.
	UnassignedSlug = "unassigned"
)

// IsUnassignedSlug reports whether a project slug is the instance unassigned bucket.
func IsUnassignedSlug(slug string) bool {
	return slug == UnassignedSlug
}

// DisplayName returns an empty string for the unassigned bucket (omit in UI).
func DisplayName(name, slug string) string {
	if IsUnassignedSlug(slug) {
		return ""
	}
	return name
}

// DisplaySlug returns an empty string for the unassigned bucket (omit project links).
func DisplaySlug(slug string) string {
	if IsUnassignedSlug(slug) {
		return ""
	}
	return slug
}
