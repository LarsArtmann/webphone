// Package retention runs the bounded deletion pass (plan T25): with
// retention_days configured, everything stored older than the window
// is deleted daily — messages (with attachments), fax jobs (with
// documents), the threads those deletions empty, and session rows
// that expired unattended. The default (0) keeps everything forever;
// the sweeper then never starts.
package retention

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// Start runs one sweep immediately and then once a day until ctx ends.
// A non-positive window is the disabled default — Start returns
// without spawning anything. Session rows are NOT touched here: the
// session store already sweeps its own expiry lazily (on read and on
// the next mint). engine is the read-model container: fax jobs and
// call transcripts live in readmodels.db since the T11/T13 swaps, so
// their retention legs ride the engine sweeps (fax documents come
// back for unlinking; transcripts carry no blobs).
func Start(ctx context.Context, db *sql.DB, blobs *blob.Store, engine *readmodels.Engine, window time.Duration) {
	if window <= 0 {
		return
	}
	run := func() {
		// Loop containment: a panicking sweep must not kill the process,
		// and the sweeper must outlive any single tick.
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("retention: sweep PANIC recovered; tick skipped, sweeper survives", "panic", rec)
			}
		}()
		cutoff := time.Now().Add(-window)
		res, err := store.Sweep(ctx, db, cutoff)
		if err != nil {
			slog.Error("retention: sweep failed", "error", err)
			return
		}
		faxPaths, faxesRemoved, err := engine.Faxes().SweepBefore(ctx, cutoff)
		if err != nil {
			slog.Error("retention: fax sweep failed", "error", err)
		}
		transcriptsRemoved, err := engine.Transcripts().SweepBefore(ctx, cutoff)
		if err != nil {
			slog.Error("retention: transcript sweep failed", "error", err)
		}
		removedBlobs := 0
		for _, path := range append(res.BlobPaths, faxPaths...) {
			// A corrupt stored path skips its unlink (logged) instead of
			// killing the sweep: retention is daily background hygiene, and
			// one bad row must not protect every other blob from expiry.
			rel, err := domain.ParseRelPath(path)
			if err != nil {
				slog.Warn("retention: corrupt blob path skipped", "path", path, "error", err)
				continue
			}
			if err := blobs.Remove(rel); err != nil {
				slog.Warn("retention: blob unlink failed", "path", path, "error", err)
				continue
			}
			removedBlobs++
		}
		slog.Info("retention: sweep complete",
			"cutoff", cutoff.Format(time.RFC3339),
			"messages", res.Messages,
			"faxes", faxesRemoved,
			"threads", res.Threads,
			"transcripts", transcriptsRemoved,
			"blobs_removed", removedBlobs,
		)
	}
	go func() {
		run()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
