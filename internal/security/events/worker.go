package events

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// AuditLogPath is the host path for Coraza JSON audit lines (mounted into dy-proxy).
func AuditLogPath(cfg config.Config) string {
	return filepath.Join(cfg.CaddyDir(), "coraza-audit.log")
}

// Worker tails (or reads new lines from) the Coraza audit log and purges old rows.
type Worker struct {
	Cfg    config.Config
	Q      *sqlc.Queries
	Log    *slog.Logger
	Now    func() time.Time
	offset int64
}

func (w *Worker) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now().UTC()
}

// TickOnce ingests new audit lines when the log exists; skips cleanly if missing.
// Also runs the 7-day purge.
func (w *Worker) TickOnce(ctx context.Context) error {
	if w.Q == nil {
		return nil
	}
	_, _ = PurgeDefaultRetention(ctx, w.Q, w.now())
	path := AuditLogPath(w.Cfg)
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // soft-dep #79 / no traffic yet
		}
		return err
	}
	if st.Size() < w.offset {
		w.offset = 0 // rotated
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Seek(w.offset, io.SeekStart); err != nil {
		return err
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	ins := Inserter{Q: w.Q}
	for sc.Scan() {
		line := sc.Text()
		hit, err := ParseCorazaAuditLine(line)
		if err != nil {
			continue
		}
		if hit.Hostname == "" {
			continue
		}
		if _, err := ins.InsertWAFHit(ctx, hit); err != nil && w.Log != nil {
			w.Log.Debug("security event insert skipped", "err", err, "host", hit.Hostname)
		}
	}
	pos, _ := f.Seek(0, io.SeekCurrent)
	w.offset = pos
	return sc.Err()
}

// Run starts a periodic ingest/purge loop until ctx is done.
func (w *Worker) Run(ctx context.Context, every time.Duration) {
	if every <= 0 {
		every = 30 * time.Second
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		if err := w.TickOnce(ctx); err != nil && w.Log != nil {
			w.Log.Warn("security events tick", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
