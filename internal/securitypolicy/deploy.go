package securitypolicy

import (
	"context"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// DeployLimits returns docker resource limits for a service deploy.
func DeployLimits(ctx context.Context, q *sqlc.Queries, serviceID string) (memMB int, nanoCPUs int64, pids *int64) {
	inst, err := LoadInstance(ctx, q)
	if err != nil {
		return 512, docker.OneCPU, nil
	}
	site, err := LoadSite(ctx, q, serviceID)
	if err != nil {
		site = DefaultSite(serviceID)
	}
	memMB = EffectiveMemMB(inst, site)
	cpu := EffectiveCPU(inst, site)
	nanoCPUs = int64(cpu * float64(docker.OneCPU))
	p := int64(EffectivePIDs(inst, site))
	pids = &p
	return memMB, nanoCPUs, pids
}
