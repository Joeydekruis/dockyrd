package appconfig

import "testing"

func TestDeploymentImageRefImage(t *testing.T) {
	cfg, err := Parse([]byte(`{"source":{"type":"image","ref":"nginx:alpine"},"port":80}`))
	if err != nil {
		t.Fatal(err)
	}
	ref, build, err := cfg.DeploymentImageRef("app", "abc")
	if err != nil || build || ref != "nginx:alpine" {
		t.Fatalf("got ref=%q build=%v err=%v", ref, build, err)
	}
}

func TestDeploymentImageRefGit(t *testing.T) {
	cfg, err := Parse([]byte(`{
		"source":{"type":"git","url":"https://github.com/example/app.git","branch":"main"},
		"builder":"dockerfile",
		"port":3000
	}`))
	if err != nil {
		t.Fatal(err)
	}
	ref, build, err := cfg.DeploymentImageRef("my-app", "01ARZ3NDEK")
	if err != nil || !build || ref != "dockyard/my-app:01ARZ3NDEK" {
		t.Fatalf("got ref=%q build=%v err=%v", ref, build, err)
	}
}

func TestEffectiveBuilderDefaultsRailpack(t *testing.T) {
	cfg, _ := Parse([]byte(`{"source":{"type":"git","url":"https://x"},"port":80}`))
	if cfg.EffectiveBuilder("/tmp/nodf") != BuilderRailpack {
		t.Fatalf("builder=%q", cfg.EffectiveBuilder("/tmp/nodf"))
	}
}

func TestGitHubSourceCloneRemote(t *testing.T) {
	cfg, _ := Parse([]byte(`{"source":{"type":"github","repo":"acme/site","branch":"dev"},"port":80}`))
	url, branch, err := cfg.Source.CloneRemote()
	if err != nil || url != "https://github.com/acme/site.git" || branch != "dev" {
		t.Fatalf("url=%q branch=%q err=%v", url, branch, err)
	}
}

func TestBuildContextDirMonorepo(t *testing.T) {
	cfg, _ := Parse([]byte(`{"source":{"type":"git","url":"https://x","rootDir":"apps/api"},"builder":"dockerfile","port":80}`))
	got := cfg.BuildContextDir("/build/src")
	want := "/build/src/apps/api"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
