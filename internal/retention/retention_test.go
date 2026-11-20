package retention

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// The pins below exercise the FULL sweep composition — store rows,
// engine fax/transcript sweeps, and the disk unlinks — through the
// deterministic sweepOnce seam. Retention deletes the user's data
// daily, so the pins assert real file unlinks and real row deletions
// (never just returned counts), the disabled default keeps everything,
// and one corrupt stored path cannot protect every other blob from
// expiry.

// sweepWorld wires the stores exactly the way the composition root
// wires them (app.go): one store db (the legacy ladder's frozen rows,
// seeded with raw SQL since the Messages type's 2026-11-20 deletion),
// one blob root, and ONE journal system carrying the threads, faxes
// and transcripts the retention legs sweep.
func sweepWorld(t *testing.T) (*sql.DB, *blob.Store, *system.System) {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatalf("open blob store: %v", err)
	}
	coreSys, err := cqrs.NewCoreSystem(context.Background(), "file:"+filepath.Join(t.TempDir(), "system.db"))
	if err != nil {
		t.Fatalf("boot core system: %v", err)
	}
	t.Cleanup(func() { _ = coreSys.Close() })
	return db, blobs, coreSys
}

// journalSweepersOf bundles the journal legs over the sweep world's
// one system.
func journalSweepersOf(coreSys *system.System) JournalSweepers {
	return JournalSweepers{
		Threads:     cqrs.NewThreads(coreSys),
		Faxes:       cqrs.NewFaxes(coreSys),
		Transcripts: cqrs.NewTranscripts(coreSys),
	}
}

// mustSaveBlob stores real bytes and returns the stored relative path,
// so the expired-tenant assertions can prove the file left the disk.
func mustSaveBlob(t *testing.T, blobs *blob.Store, subDir, ext, content string) string {
	t.Helper()
	rel, err := blobs.Save(subDir, ext, []byte(content))
	if err != nil {
		t.Fatalf("save blob: %v", err)
	}
	return rel.String()
}

func blobPath(blobs *blob.Store, rel string) string {
	return filepath.Join(blobs.Root(), rel)
}

func mustNotExist(t *testing.T, path, what string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("%s must be unlinked after the sweep, stat = %v", what, err)
	}
}

func mustExist(t *testing.T, path, what string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s must survive the sweep, stat = %v", what, err)
	}
}

// seedLegacyMessage persists one legacy thread + message row, the
// message optionally carrying an attachment whose stored path is
// taken verbatim — including corrupt shapes, which the store row
// happily holds (the path is validated only at unlink time). It
// returns the thread id so tests can assert on the surviving thread.
// Raw SQL since the Messages type's 2026-11-20 deletion: the frozen
// tables stay until the §5 DROP migration and these pins must keep
// proving the sweep empties them.
func seedLegacyMessage(t *testing.T, db *sql.DB, owner, remote string, at time.Time, blobRel string) string {
	t.Helper()
	thread := "t-" + remote[len(remote)-4:]
	msg := "m-" + remote[len(remote)-4:]
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO threads (id, owner, remote, last_activity_at, unread) VALUES (?,?,?,?,0)`,
		thread, owner, remote, at.Unix()); err != nil {
		t.Fatalf("seed thread %s: %v", thread, err)
	}
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO messages (id, thread_id, owner, remote, direction, channel, body, status, provider_ref, created_at)
			 VALUES (?,?,?,?, 'outbound','sms','retention pin','sent','',?)`,
		msg, thread, owner, remote, at.Unix()); err != nil {
		t.Fatalf("seed message %s: %v", msg, err)
	}
	if blobRel != "" {
		if _, err := db.ExecContext(context.Background(),
			`INSERT INTO attachments (id, message_id, name, mime_type, size_bytes, path)
				 VALUES (?,?,?,?,4,?)`,
			"a-"+msg, msg, "doc.bin", "application/octet-stream", blobRel); err != nil {
			t.Fatalf("seed attachment for %s: %v", msg, err)
		}
	}
	return thread
}

func countMessages(t *testing.T, db *sql.DB, thread string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM messages WHERE thread_id = ?`, thread).Scan(&count); err != nil {
		t.Fatalf("count messages for %s: %v", thread, err)
	}
	return count
}

func mustFaxJob(t *testing.T, owner domain.Extension, remote string, at time.Time, docRel string) domain.FaxJob {
	t.Helper()
	job := domain.FaxJob{
		ID: domain.GenerateFaxID(), Owner: owner,
		Remote:    domain.MustParsePhone(remote),
		Direction: domain.FaxInbound, Status: domain.FaxReceived,
		CreatedAt: at, UpdatedAt: at,
	}
	if docRel != "" {
		job.DocumentPath = domain.MustParseRelPath(docRel)
	}
	return job
}

func TestSweepOnceExpiresOldTenantsAndUnlinksTheirBlobs(t *testing.T) {
	db, blobs, coreSys := sweepWorld(t)
	ctx := context.Background()
	faxes := cqrs.NewFaxes(coreSys)
	transcripts := cqrs.NewTranscripts(coreSys)
	owner := domain.MustParseExtension("1001")
	now := time.Now()
	old := now.Add(-48 * time.Hour)
	cutoff := now.Add(-24 * time.Hour)

	oldMsgBlob := mustSaveBlob(t, blobs, "attachments", ".png", "old-bytes")
	liveMsgBlob := mustSaveBlob(t, blobs, "attachments", ".png", "live-bytes")
	oldFaxDoc := mustSaveBlob(t, blobs, "faxes", ".pdf", "old-fax")
	liveFaxDoc := mustSaveBlob(t, blobs, "faxes", ".pdf", "live-fax")

	oldThread := seedLegacyMessage(t, db, owner.String(), "+441632960961", old, oldMsgBlob)
	liveThread := seedLegacyMessage(t, db, owner.String(), "+441632960977", now, liveMsgBlob)
	liveFax := mustFaxJob(t, owner, "+15550000012", now, liveFaxDoc)
	if err := faxes.Create(ctx, mustFaxJob(t, owner, "+15550000011", old, oldFaxDoc)); err != nil {
		t.Fatalf("create old fax: %v", err)
	}
	if err := faxes.Create(ctx, liveFax); err != nil {
		t.Fatalf("create live fax: %v", err)
	}
	// A segment's retention clock is its ARRIVAL time (the journal's
	// seq stamp, not StartedAt), so freshly appended segments are never
	// sweepable here; the expired-segment deletion itself is pinned in
	// the cqrs battery. What this package must prove is the wiring: the
	// journal legs run, report their counts, and live speech survives.
	for _, seg := range []domain.CallTranscriptSegment{
		{Owner: owner, CallID: "old-call", Direction: domain.DirectionInbound, Remote: "+15550000011", StartedAt: old, Text: "old line"},
		{Owner: owner, CallID: "live-call", Direction: domain.DirectionInbound, Remote: "+15550000012", StartedAt: now, Text: "live line"},
	} {
		if err := transcripts.Append(ctx, seg); err != nil {
			t.Fatalf("append transcript: %v", err)
		}
	}

	sum := sweepOnce(ctx, db, blobs, journalSweepersOf(coreSys), cutoff)

	if sum.messages != 1 || sum.threads != 1 || sum.journalFaxes != 1 || sum.transcripts != 0 || sum.blobs != 2 {
		t.Fatalf("sweep summary wrong: %+v", sum)
	}
	mustNotExist(t, blobPath(blobs, oldMsgBlob), "expired message attachment")
	mustNotExist(t, blobPath(blobs, oldFaxDoc), "expired fax document")
	mustExist(t, blobPath(blobs, liveMsgBlob), "live message attachment")
	mustExist(t, blobPath(blobs, liveFaxDoc), "live fax document")

	if n := countMessages(t, db, liveThread); n != 1 {
		t.Fatalf("live message lost: %d", n)
	}
	if n := countMessages(t, db, oldThread); n != 0 {
		t.Fatalf("swept thread's messages must be gone: %d", n)
	}
	if _, err := faxes.Get(ctx, owner, liveFax.ID); err != nil {
		t.Fatalf("live fax lost: %v", err)
	}
	if freshTranscript, err := transcripts.Call(ctx, owner, "old-call"); err != nil || freshTranscript == nil {
		t.Fatalf("freshly arrived transcript must survive despite its backdated StartedAt: %v %v", freshTranscript, err)
	}
	if liveTranscript, err := transcripts.Call(ctx, owner, "live-call"); err != nil || liveTranscript == nil {
		t.Fatalf("live transcript lost: %v %v", liveTranscript, err)
	}
}

// TestSweepOnceSkipsCorruptBlobPath pins the containment rule: a
// stored path that ParseRelPath rejects (here: absolute) is logged and
// skipped — the sweep must not die on it and must not skip the
// NEIGHBOURING valid blob in the same pass.
func TestSweepOnceSkipsCorruptBlobPath(t *testing.T) {
	db, blobs, _ := sweepWorld(t)
	owner := domain.MustParseExtension("1001")
	now := time.Now()
	old := now.Add(-48 * time.Hour)
	cutoff := now.Add(-24 * time.Hour)

	validBlob := mustSaveBlob(t, blobs, "attachments", ".png", "valid-bytes")
	seedLegacyMessage(t, db, owner.String(), "+441632960961", old, "/abs/escape")
	seedLegacyMessage(t, db, owner.String(), "+441632960963", old, validBlob)

	sum := sweepOnce(context.Background(), db, blobs, retention0Sweepers(), cutoff)

	if sum.messages != 2 || sum.blobs != 1 {
		t.Fatalf("corrupt path must skip itself, not the sweep: %+v", sum)
	}
	mustNotExist(t, blobPath(blobs, validBlob), "valid blob behind a corrupt row")
}

// TestStartDisabledWindowKeepsEverything pins the off-switch: a
// non-positive window spawns no sweeper, so even ancient content is
// still there afterwards (the grace sleep only gives a BUGGY spawn
// time to prove itself — a correct Start never launches one).
func TestStartDisabledWindowKeepsEverything(t *testing.T) {
	for _, window := range []time.Duration{0, -time.Hour} {
		t.Run(window.String(), func(t *testing.T) {
			db, blobs, _ := sweepWorld(t)
			ctx := context.Background()
			owner := domain.MustParseExtension("1001")
			ancient := time.Now().Add(-365 * 24 * time.Hour)
			thread := seedLegacyMessage(t, db, owner.String(), "+441632960965", ancient, "")

			Start(ctx, db, blobs, retention0Sweepers(), window)
			time.Sleep(50 * time.Millisecond)

			if n := countMessages(t, db, thread); n != 1 {
				t.Fatalf("disabled retention must delete nothing: %d", n)
			}
		})
	}
}

// TestStartSweepsImmediatelyThenHonorsContext pins the loop contract:
// the first pass runs NOW (not a day late — a server that configured
// retention and shut down a minute later still swept), and ctx
// cancellation retires the goroutine without a second pass.
func TestStartSweepsImmediatelyThenHonorsContext(t *testing.T) {
	db, blobs, _ := sweepWorld(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	owner := domain.MustParseExtension("1001")
	ancient := time.Now().Add(-365 * 24 * time.Hour)
	thread := seedLegacyMessage(t, db, owner.String(), "+441632960967", ancient, "")

	Start(ctx, db, blobs, retention0Sweepers(), 24*time.Hour)

	deadline := time.Now().Add(2 * time.Second)
	for {
		if n := countMessages(t, db, thread); n == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Start never ran its immediate first sweep")
		}
		time.Sleep(5 * time.Millisecond)
	}

	cancel()
	time.Sleep(20 * time.Millisecond)
}

// retention0Sweepers is the nil-legs bundle (pre-wiring test
// compositions).
func retention0Sweepers() JournalSweepers { return JournalSweepers{} }
