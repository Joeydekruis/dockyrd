package docker

// Standard dockyrd object labels (briefing §3.3).
const (
	LabelManaged   = "dev.dockyrd.managed"
	LabelServer    = "dev.dockyrd.server"
	LabelProject   = "dev.dockyrd.project"
	LabelService   = "dev.dockyrd.service"
	LabelKind      = "dev.dockyrd.kind"
	LabelDeployment = "dev.dockyrd.deployment"
	LabelLink       = "dev.dockyrd.link"
)

const managedTrue = "true"

// Kind values for LabelKind.
const (
	KindApp      = "app"
	KindCompose  = "compose"
	KindDatabase = "database"
	KindRedis    = "redis"
	KindBucket   = "bucket"
	KindSystem   = "system"
)

// Identity holds the label set for a managed object.
type Identity struct {
	ServerID       string
	ProjectID      string
	ServiceID      string
	Kind           string
	DeploymentID   string // apps only; empty for databases etc.
	ComposeService string // compose stack service name
}

// Labels returns Docker labels for create/update.
func (id Identity) Labels() map[string]string {
	out := map[string]string{
		LabelManaged: managedTrue,
		LabelServer:  id.ServerID,
		LabelProject: id.ProjectID,
		LabelService: id.ServiceID,
		LabelKind:    id.Kind,
	}
	if id.DeploymentID != "" {
		out[LabelDeployment] = id.DeploymentID
	}
	return out
}

// ParseIdentity reads dockyrd labels from a container label map.
func ParseIdentity(labels map[string]string) (Identity, bool) {
	if labels == nil || labels[LabelManaged] != managedTrue {
		return Identity{}, false
	}
	return Identity{
		ServerID:       labels[LabelServer],
		ProjectID:      labels[LabelProject],
		ServiceID:      labels[LabelService],
		Kind:           labels[LabelKind],
		DeploymentID:   labels[LabelDeployment],
		ComposeService: labels["com.docker.compose.service"],
	}, true
}

// ManagedFilter is the list filter for dockyrd-owned containers.
func ManagedFilter() map[string][]string {
	return map[string][]string{"label": {LabelManaged + "=" + managedTrue}}
}
