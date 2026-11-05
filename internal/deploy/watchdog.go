package deploy

import (
	"context"

	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
)

func (rt *Runtime) startBuildWatchdog(ctx context.Context, serviceID, deploymentID string, cancelBuild context.CancelFunc) context.CancelFunc {
	if rt.Resources == nil || !rt.Resources.Available() {
		return func() {}
	}
	wctx, stop := context.WithCancel(ctx)
	go build.RunBuildWatchdog(wctx, rt.Resources, build.WatchdogCallbacks{
		OnFrozen: func() {
			if rt.Events != nil {
				_ = rt.Events.Record(wctx, eventlog.RecordInput{
					Level: "info", Category: "build", Title: "build.frozen",
					Detail: "Build paused while apps recover memory",
					ServiceID: serviceID, DeploymentID: deploymentID,
				})
			}
		},
		OnResumed: func() {
			if rt.Events != nil {
				_ = rt.Events.Record(wctx, eventlog.RecordInput{
					Level: "info", Category: "build", Title: "build.resumed",
					Detail: "Build resumed",
					ServiceID: serviceID, DeploymentID: deploymentID,
				})
			}
		},
		OnFailed: func(reason string) {
			if cancelBuild != nil {
				cancelBuild()
			}
			if rt.Events != nil {
				_ = rt.Events.Record(wctx, eventlog.RecordInput{
					Level: "error", Category: "build", Title: "build.failed",
					Detail: reason,
					ServiceID: serviceID, DeploymentID: deploymentID,
				})
			}
		},
	}, build.WatchdogThresholds{})
	return stop
}
