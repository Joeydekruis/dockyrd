package compose_test

import (
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/compose"
)

func TestPolicyRejectsBuild(t *testing.T) {
	doc := map[string]any{
		"services": map[string]any{
			"web": map[string]any{"build": "."},
		},
	}
	issues := compose.CheckPolicy(doc, compose.PolicyOptions{})
	if !compose.HasBlockingErrors(issues) {
		t.Fatal("expected build to be blocked")
	}
}

func TestPolicyRejectsPrivileged(t *testing.T) {
	doc := map[string]any{
		"services": map[string]any{
			"web": map[string]any{"image": "nginx:alpine", "privileged": true},
		},
	}
	issues := compose.CheckPolicy(doc, compose.PolicyOptions{})
	if !compose.HasBlockingErrors(issues) {
		t.Fatal("expected privileged to be blocked")
	}
}

func TestPolicyAllowsImageOnly(t *testing.T) {
	doc := map[string]any{
		"services": map[string]any{
			"web": map[string]any{"image": "nginx:alpine"},
		},
	}
	issues := compose.CheckPolicy(doc, compose.PolicyOptions{})
	if compose.HasBlockingErrors(issues) {
		t.Fatalf("unexpected blocking issues: %+v", issues)
	}
}

func TestMutateStripsHostPorts(t *testing.T) {
	doc := map[string]any{
		"services": map[string]any{
			"web": map[string]any{
				"image": "nginx:alpine",
				"ports": []any{"8080:80"},
			},
		},
	}
	in := compose.MutateInput{
		ServiceID:    "0123456789abcdef01234567",
		DeploymentID: "dep1",
		ServerID:     "srv",
		ProjectID:    "proj",
	}
	out, warnings := compose.Mutate(doc, in)
	svc := out["services"].(map[string]any)["web"].(map[string]any)
	if _, ok := svc["ports"]; ok {
		t.Fatal("expected host ports stripped")
	}
	if len(warnings) == 0 {
		t.Fatal("expected port strip warning")
	}
	if compose.ProjectName(in.ServiceID) != "dyc-01234567" {
		t.Fatalf("unexpected project name %q", compose.ProjectName(in.ServiceID))
	}
}

func TestPolicyBindMountOutsideStack(t *testing.T) {
	doc := map[string]any{
		"services": map[string]any{
			"web": map[string]any{
				"image":   "nginx:alpine",
				"volumes": []any{"/etc/passwd:/etc/passwd:ro"},
			},
		},
	}
	issues := compose.CheckPolicy(doc, compose.PolicyOptions{StackDataDir: "/var/lib/dockyrd/compose/svc1"})
	if !compose.HasBlockingErrors(issues) {
		t.Fatal("expected bind mount outside stack to be blocked")
	}
}
