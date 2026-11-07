package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestApplyUserContainerHardeningApp(t *testing.T) {
	h := &container.HostConfig{}
	applyUserContainerHardening(h, 256, false)
	if len(h.SecurityOpt) != 1 || h.SecurityOpt[0] != "no-new-privileges:true" {
		t.Fatalf("SecurityOpt: %v", h.SecurityOpt)
	}
	if len(h.CapDrop) != 1 || h.CapDrop[0] != "ALL" {
		t.Fatalf("CapDrop: %v", h.CapDrop)
	}
	if len(h.CapAdd) != 0 {
		t.Fatalf("CapAdd: %v", h.CapAdd)
	}
	if h.Resources.PidsLimit == nil || *h.Resources.PidsLimit != defaultPidsLimit {
		t.Fatalf("PidsLimit: %v", h.Resources.PidsLimit)
	}
	if h.Resources.NanoCPUs != oneCPU {
		t.Fatalf("NanoCPUs: %d", h.Resources.NanoCPUs)
	}
	if h.Resources.Memory != 256*1024*1024 {
		t.Fatalf("Memory: %d", h.Resources.Memory)
	}
}

func TestApplyUserContainerHardeningDatabase(t *testing.T) {
	h := &container.HostConfig{}
	applyUserContainerHardening(h, 512, true)
	if h.Resources.NanoCPUs != 0 {
		t.Fatalf("database should not set CPU quota by default")
	}
	if len(h.CapAdd) != len(containerCapsDatabase) {
		t.Fatalf("CapAdd: %v", h.CapAdd)
	}
}
