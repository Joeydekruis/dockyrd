package watch

import (
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/notify"
)

func TestEvalDiskAlert(t *testing.T) {
	const warn, crit, resolve = 85, 95, 80
	cases := []struct {
		pct   float64
		emit  bool
		level notify.Level
		state notify.State
	}{
		{50, true, notify.LevelWarning, notify.StateResolved},
		{84, false, "", ""},
		{85, true, notify.LevelWarning, notify.StateFiring},
		{90, true, notify.LevelWarning, notify.StateFiring},
		{95, true, notify.LevelError, notify.StateFiring},
		{96, true, notify.LevelError, notify.StateFiring},
		{79, true, notify.LevelWarning, notify.StateResolved},
		{80, false, "", ""},
	}
	for _, tc := range cases {
		dec := evalDiskAlert(tc.pct, warn, crit, resolve)
		if dec.emit != tc.emit {
			t.Fatalf("pct %.0f emit=%v want %v", tc.pct, dec.emit, tc.emit)
		}
		if !tc.emit {
			continue
		}
		if dec.level != tc.level || dec.state != tc.state {
			t.Fatalf("pct %.0f got level=%s state=%s want %s %s", tc.pct, dec.level, dec.state, tc.level, tc.state)
		}
	}
}

func TestEvalCertLifetime(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	notBefore := now.Add(-45 * 24 * time.Hour)
	notAfter := now.Add(45 * 24 * time.Hour)

	valid := evalCertLifetime(now, notBefore, notAfter, 15)
	if valid.status != "valid" || valid.emit {
		t.Fatalf("healthy cert: %+v", valid)
	}

	expiring := evalCertLifetime(now, notBefore, now.Add(10*24*time.Hour), 15)
	if !expiring.emit || expiring.status != "expiring" || expiring.level != notify.LevelWarning {
		t.Fatalf("expiring warning: %+v", expiring)
	}

	expired := evalCertLifetime(now, notBefore, now.Add(-time.Hour), 15)
	if !expired.emit || expired.status != "expired" || expired.level != notify.LevelError {
		t.Fatalf("expired: %+v", expired)
	}

	renewed := evalCertLifetime(now, now, now.Add(90*24*time.Hour), 15)
	if !renewed.resolve || renewed.status != "valid" {
		t.Fatalf("renewed resolve: %+v", renewed)
	}
}

func TestLoadNotifySettingsDefaults(t *testing.T) {
	s := notifySettings{}
	if s.DiskWarnPct != 0 {
		t.Fatal("zero value")
	}
	out := notifySettings{DiskWarnPct: 85, DiskCritPct: 95, CertWarnPct: 15}
	if out.DiskWarnPct != 85 || out.CertWarnPct != 15 {
		t.Fatal(out)
	}
}

func TestBytesGiB(t *testing.T) {
	if bytesGiB(1024*1024*1024) != "1.0 GiB" {
		t.Fatal(bytesGiB(1024 * 1024 * 1024))
	}
}
