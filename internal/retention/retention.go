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
	"github.com/larsartmann/webphone/internal/store"
)

// sweepSummary reports one pass: what was deleted and how many blob
// files were unlinked. It is the sweep log line's data and the
// package tests' assertion surface.
type sweepSummary struct {
	messages       int64
	faxes          int64
	journalFaxes   int64
	threads        int64
	journalThreads int64
	transcripts    int64
	blobs          int64
}

// ThreadSweeper is the threads journal leg's seam: the backend's
// SweepBefore (paths out for unlinking, rows and journal bytes gone
// before they return).
type ThreadSweeper interface {
	SweepBefore(ctx context.Context, cutoff time.Time) ([]string, int64, error)
}

// FaxSweeper is the faxes journal leg's seam (the same shape).
type FaxSweeper interface {
	SweepBefore(ctx context.Context, cutoff time.Time) ([]string, int64, error)
}

// TranscriptSweeper is the transcripts journal leg's seam (segments
// carry no blobs; only the count returns).
type TranscriptSweeper interface {
	SweepBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

// JournalSweepers bundles the journal-backed retention legs (ONE param
// for the composition root instead of a growing positional list).
type JournalSweepers struct {
	Threads     ThreadSweeper
	Faxes       FaxSweeper
	Transcripts TranscriptSweeper
}

// Start runs one sweep immediately and then once a day until ctx ends.
// A non-positive window is the disabled default — Start returns
// without spawning anything. Session rows are NOT touched here: the
// session store already sweeps its own expiry lazily (on read and on
// the next mint). engine is the read-model container: fax jobs and
// call transcripts live in readmodels.db since the T11/T13 swaps, so
// their retention legs ride the engine sweeps (fax documents come
// back for unlinking; transcripts carry no blobs). threads is the
// journal-backed core leg (A36): expired threads tombstone out of the
// views AND the journal itself, so the GDPR arm deletes bytes, not
// just rows.
func Start(ctx context.Context, db *sql.DB, blobs *blob.Store, journal JournalSweepers, window time.Duration) {
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
		sweepOnce(ctx, db, blobs, journal, time.Now().Add(-window))
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

// sweepOnce runs one deletion pass at a fixed cutoff — the store rows
// first (their blob paths leave with the rows), then the engine's fax
// jobs and transcripts, then the collected blob files. It is the
// tests' deterministic seam: Start computes the cutoff and calls this
// on its ticker, so the pins below exercise the real composition with
// zero timing dependence. A store failure skips the pass entirely
// (the next tick retries); fax and transcript failures degrade to a
// logged zero, never a dropped store sweep.
func sweepOnce(ctx context.Context, db *sql.DB, blobs *blob.Store, journal JournalSweepers, cutoff time.Time) sweepSummary {
	var sum sweepSummary
	res, err := store.Sweep(ctx, db, cutoff)
	if err != nil {
		slog.Error("retention: sweep failed", "error", err)
		return sum
	}
	sum.messages, sum.threads = res.Messages, res.Threads

	// The journal leg (A36): expired threads leave the views AND the
	// journal (bytes gone); their blob paths join the faxes' in the one
	// unlink loop. A nil seam skips the leg (pre-A36 test compositions)
	// — never silently: the composition root always wires the real one.
	threadPaths := []string{}
	if journal.Threads != nil {
		paths, journalRemoved, err := journal.Threads.SweepBefore(ctx, cutoff)
		if err != nil {
			slog.Error("retention: thread journal sweep failed", "error", err)
		} else {
			sum.journalThreads = journalRemoved
			threadPaths = paths
		}
	}

	// The faxes journal leg (Phase 6): expired jobs leave the views AND
	// the journal (the readmodels engine's frozen lane retired with A43
	// — system.db is the only fax home).
	faxPaths := []string{}
	if journal.Faxes != nil {
		paths, journalFaxes, err := journal.Faxes.SweepBefore(ctx, cutoff)
		if err != nil {
			slog.Error("retention: fax journal sweep failed", "error", err)
		} else {
			sum.journalFaxes = journalFaxes
			faxPaths = paths
		}
	}
	sum.faxes = sum.journalFaxes

	// The transcripts journal leg: segments carry no blobs, only the
	// count reports (the Phase-4 swap's retention re-point — the engine
	// leg had kept sweeping the frozen lane while the journal grew).
	if journal.Transcripts != nil {
		transcriptsRemoved, err := journal.Transcripts.SweepBefore(ctx, cutoff)
		if err != nil {
			slog.Error("retention: transcript sweep failed", "error", err)
		}
		sum.transcripts = transcriptsRemoved
	}

	for _, path := range append(append(res.BlobPaths, threadPaths...), faxPaths...) {
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
		sum.blobs++
	}
	slog.Info("retention: sweep complete",
		"cutoff", cutoff.Format(time.RFC3339),
		"messages", sum.messages,
		"faxes", sum.faxes,
		"journal_faxes", sum.journalFaxes,
		"threads", sum.threads,
		"journal_threads", sum.journalThreads,
		"transcripts", sum.transcripts,
		"blobs_removed", sum.blobs,
	)
	return sum
}
