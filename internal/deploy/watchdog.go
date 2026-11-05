package deploy

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
)

func (rt *Runtime) startBuildWatchdog(ctx context.Context, serviceID, deploymentID string, cancelBuild context.CancelFunc) (stop context.CancelFunc, frozenMs *int64) {
	if rt.Resources == nil || !rt.Resources.Available() {
		return func() {}, new(int64)
	}
	var acc int64
	var freezeAt int64
	wctx, stopWatch := context.WithCancel(ctx)
	go build.RunBuildWatchdog(wctx, rt.Resources, build.WatchdogCallbacks{
		OnFrozen: func() {
			atomic.StoreInt64(&freezeAt, time.Now().UnixMilli())
			if rt.Events != nil {
				_ = rt.Events.Record(wctx, eventlog.RecordInput{
					Level: "info", Category: "build", Title: "build.frozen",
					Detail: "Build paused while apps recover memory",
					ServiceID: serviceID, DeploymentID: deploymentID,
				})
			}
		},
		OnResumed: func() {
			start := atomic.LoadInt64(&freezeAt)
			if start > 0 {
				delta := time.Now().UnixMilli() - start
				if delta > 0 {
					atomic.AddInt64(&acc, delta)
				}
				atomic.StoreInt64(&freezeAt, 0)
			}
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
	return stopWatch, &acc
}
