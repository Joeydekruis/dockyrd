package hostops

import "context"

type stubAction struct {
	kind string
}

func newStub(kind string) stubAction {
	return stubAction{kind: kind}
}

func (s stubAction) Kind() string { return s.kind }

func (s stubAction) Plan(_ context.Context) (Plan, error) {
	return Plan{"preview": "placeholder", "kind": s.kind}, nil
}

func (s stubAction) Apply(_ context.Context) (RevertPayload, error) {
	return nil, ErrNotImplemented
}

func (s stubAction) Verify(_ context.Context) error {
	return ErrNotImplemented
}

func (s stubAction) Revert(_ context.Context, _ RevertPayload) error {
	return ErrNotImplemented
}

func (s stubAction) RequiresLockoutConfirm() bool { return true }

func NewFirewallEnsureStub() Action        { return newStub(KindFirewallEnsure) }
func NewSSHPasswordAuthDisableStub() Action { return newStub(KindSSHPasswordAuthDisable) }
func NewUnattendedEnableStub() Action      { return newStub(KindUnattendedEnable) }
func NewCrowdsecBouncerInstallStub() Action { return newStub(KindCrowdsecBouncerInstall) }
func NewProxySwapStub() Action             { return newStub(KindProxySwap) }
