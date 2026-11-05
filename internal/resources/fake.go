package resources

import "context"

// Fake is a no-op Resources for macOS dev and unit tests.
type Fake struct{}

func (f *Fake) Available() bool { return false }

func (f *Fake) Snapshot(context.Context) (Snapshot, error) {
	return Snapshot{
		MemTotal:      8 * 1024 * 1024 * 1024,
		AppsWS:        512 * 1024 * 1024,
		SystemWS:      256 * 1024 * 1024,
		SwapFree:      2 * 1024 * 1024 * 1024,
		KernelReserve: 256 * 1024 * 1024,
	}, nil
}

func (f *Fake) SetBuildLimits(context.Context, BuildLimits) error { return nil }

func (f *Fake) FreezeBuild(context.Context, bool) error { return nil }

func (f *Fake) BuildUsage(context.Context) (Usage, error) { return Usage{}, nil }
