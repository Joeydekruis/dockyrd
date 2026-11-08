package server

import (
	"net/http"

	"github.com/Joeydekruis/dockyrd/internal/proxy"
	"github.com/Joeydekruis/dockyrd/internal/watch"
)

func (d apiDeps) handleServerEdgeReload(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	cli, ok := d.requireDocker(w)
	if !ok {
		return
	}
	ctx := r.Context()
	d.repairStaleLoopbackPlatformDomains(ctx)
	hosts := watch.EdgeProbeHostnames(ctx, d.cfg, d.q)
	needs := proxy.EdgeNeedsRepair(hosts)
	restart := r.URL.Query().Get("restart") == "1" || r.URL.Query().Get("restart") == "true"
	var err error
	if restart {
		err = proxy.ReloadEdge(ctx, d.cfg, d.q, cli, true, nil)
	} else {
		err = proxy.SyncEdgeFromDB(ctx, d.cfg, d.q, cli, nil)
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "edge reload failed")
		return
	}
	after := proxy.EdgeNeedsRepair(watch.EdgeProbeHostnames(ctx, d.cfg, d.q))
	writeJSON(w, http.StatusOK, map[string]any{
		"restarted": restart,
		"was_unhealthy": needs,
		"healthy_after": !after,
	})
}
