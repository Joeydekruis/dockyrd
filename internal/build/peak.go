package build

import (
	"context"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/resources"
)

// PeakSampler tracks max memory/swap during a build.
type PeakSampler struct {
	res      resources.Resources
	interval time.Duration
	stop     chan struct{}
	done     chan struct{}
	memPeak  int64
	swapPeak int64
}

// StartPeakSampler polls build slice usage until Stop is called.
func StartPeakSampler(ctx context.Context, res resources.Resources) *PeakSampler {
	if res == nil || !res.Available() {
		return nil
	}
	s := &PeakSampler{
		res:      res,
		interval: time.Second,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go s.loop(ctx)
	return s
}

func (s *PeakSampler) loop(ctx context.Context) {
	defer close(s.done)
	t := time.NewTicker(s.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stop:
			return
		case <-t.C:
			u, err := s.res.BuildUsage(ctx)
			if err != nil {
				continue
			}
			if u.MemCurrent > s.memPeak {
				s.memPeak = u.MemCurrent
			}
			if u.SwapCurrent > s.swapPeak {
				s.swapPeak = u.SwapCurrent
			}
		}
	}
}

func (s *PeakSampler) Stop() (memPeak, swapPeak int64) {
	if s == nil {
		return 0, 0
	}
	close(s.stop)
	<-s.done
	return s.memPeak, s.swapPeak
}

// FailureMessage returns user-facing copy with build-elsewhere hints.
func FailureMessage(kind, detail string) string {
	hints := " You can add swap, add a remote builder under Server settings → Builders, or build on GitHub Actions."
	switch kind {
	case "oom":
		return "This build ran out of memory under the server budget." + hints
	case "frozen_timeout":
		return "Build paused too long while your apps needed memory." + hints
	case "timeout":
		return "Build timed out." + hints
	case "interrupted":
		return "Build was interrupted (Dockyard restarted or the job was cancelled)."
	default:
		if detail != "" {
			return detail + hints
		}
		return "Build failed." + hints
	}
}
