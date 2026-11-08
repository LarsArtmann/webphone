package app

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/id/v4"

	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// mediumImportContactSeq mints unique 21-char contact ids for the
// legacy seeds.
var mediumImportContactSeq int

// seedContact writes one legacy engine contact row exactly as the old
// product wrote it (a fresh nanoid-slot id).
func seedContact(t *testing.T, engine *readmodels.Engine, owner domain.Extension, phone, name string) domain.Contact {
	t.Helper()
	mediumImportContactSeq++
	contact := domain.Contact{
		ID:        domain.MustContactID(fmt.Sprintf("%021d", mediumImportContactSeq)),
		Owner:     owner,
		Name:      name,
		Phone:     domain.MustParsePhone(phone),
		CreatedAt: time.Now(),
	}
	if err := engine.Contacts().Save(context.Background(), contact); err != nil {
		t.Fatalf("seed legacy contact: %v", err)
	}
	return contact
}

// TestContactsImportPortsLegacyRows pins the A29 contacts continuity:
// the legacy row survives the journal swap with its id and creation
// stamp (the rename-keeps-identity semantics re-derive), the marker
// latches, and a second boot imports nothing new.
func TestContactsImportPortsLegacyRows(t *testing.T) {
	cfg := testConfig(t, false)
	ctx := context.Background()
	owner := mustExtension(t, "1001")

	legacyEngine, err := readmodels.Open(cfg.DataDir)
	if err != nil {
		t.Fatalf("open legacy engine: %v", err)
	}
	seeded := seedContact(t, legacyEngine, owner, "+15550001234", "Legacy Ada")
	if err := legacyEngine.Shutdown(); err != nil {
		t.Fatalf("close legacy engine: %v", err)
	}

	application, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	svc := mustResolveCqrs(t, application)

	deadline := time.Now().Add(10 * time.Second)
	for {
		list, err := svc.ContactsSvc().List(ctx, owner)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(list) == 1 && list[0].ID == seeded.ID && list[0].Name == "Legacy Ada" {
			if !list[0].CreatedAt.Truncate(time.Millisecond).Equal(seeded.CreatedAt.Truncate(time.Millisecond)) {
				t.Fatalf("imported contact must keep its creation stamp: got %v want %v", list[0].CreatedAt, seeded.CreatedAt)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("imported contact never reached the journal projection: %+v", list)
		}
		time.Sleep(10 * time.Millisecond)
	}

	done, err := store.FlagPresent(ctx, mustSQL(t, application), flagContactsJournalImported)
	if err != nil || !done {
		t.Fatalf("contacts import marker must latch: done=%v err=%v", done, err)
	}

	streamID, err := id.ParseStreamID(owner.String())
	if err != nil {
		t.Fatal(err)
	}
	ref := id.NewStreamRef(cqrs.ContactsStreamType, streamID)
	before, err := svc.Sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("load stream: %v", err)
	}
	if err := application.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	second, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("second New: %v", err)
	}
	svc2 := mustResolveCqrs(t, second)
	after, err := svc2.Sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("load stream after second boot: %v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("second boot must not append import facts: before=%d after=%d", len(before), len(after))
	}
	if err := second.Shutdown(); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
}

// TestTranscriptsImportPortsLegacyRows pins the A29 transcripts
// continuity: legacy segments survive the swap grouped into their
// per-(owner, call) streams in seq order, and the marker latches.
func TestTranscriptsImportPortsLegacyRows(t *testing.T) {
	cfg := testConfig(t, false)
	ctx := context.Background()
	owner := mustExtension(t, "1001")

	legacyEngine, err := readmodels.Open(cfg.DataDir)
	if err != nil {
		t.Fatalf("open legacy engine: %v", err)
	}
	segA := domain.CallTranscriptSegment{Owner: owner, CallID: "legacy-call", Direction: domain.DirectionOutbound, Remote: "+15550001", StartedAt: time.Now(), Text: "first"}
	segB := segA
	segB.Text = "second"
	for _, seg := range []domain.CallTranscriptSegment{segA, segB} {
		if err := legacyEngine.Transcripts().Append(ctx, seg); err != nil {
			t.Fatalf("seed legacy transcript: %v", err)
		}
	}
	if err := legacyEngine.Shutdown(); err != nil {
		t.Fatalf("close legacy engine: %v", err)
	}

	application, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	svc := mustResolveCqrs(t, application)

	deadline := time.Now().Add(10 * time.Second)
	for {
		calls, err := svc.TranscriptsSvc().Recent(ctx, owner, 10)
		if err != nil {
			t.Fatalf("Recent: %v", err)
		}
		if len(calls) == 1 && len(calls[0].Lines) == 2 && calls[0].Lines[0] == "first" && calls[0].Lines[1] == "second" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("imported transcript never reassembled whole in spoken order: %+v", calls)
		}
		time.Sleep(10 * time.Millisecond)
	}

	done, err := store.FlagPresent(ctx, mustSQL(t, application), flagTranscriptsJournalImported)
	if err != nil || !done {
		t.Fatalf("transcripts import marker must latch: done=%v err=%v", done, err)
	}
	if err := application.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

// TestNotesAndReviewsImportPortsLegacyRows pins the A29 notes/reviews
// continuity: both slot families survive the swap with their whole-slot
// snapshots (body+disposition together, review dims together), and the
// markers latch.
func TestNotesAndReviewsImportPortsLegacyRows(t *testing.T) {
	cfg := testConfig(t, false)
	ctx := context.Background()
	owner := mustExtension(t, "1001")

	legacyEngine, err := readmodels.Open(cfg.DataDir)
	if err != nil {
		t.Fatalf("open legacy engine: %v", err)
	}
	if err := legacyEngine.Notes().Set(ctx, owner, "legacy-call", "legacy note body"); err != nil {
		t.Fatalf("seed legacy note: %v", err)
	}
	if err := legacyEngine.Notes().SetDisposition(ctx, owner, "legacy-call", domain.DispositionAnswered); err != nil {
		t.Fatalf("seed legacy disposition: %v", err)
	}
	if err := legacyEngine.Reviews().Set(ctx, owner, "legacy-call", domain.ReviewPresent, "draft-legacy", "looks solid", "en", 91); err != nil {
		t.Fatalf("seed legacy review: %v", err)
	}
	if err := legacyEngine.Shutdown(); err != nil {
		t.Fatalf("close legacy engine: %v", err)
	}

	application, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	svc := mustResolveCqrs(t, application)

	deadline := time.Now().Add(10 * time.Second)
	for {
		note, err := svc.NotesSvc().Get(ctx, owner, "legacy-call")
		if err != nil {
			t.Fatalf("note Get: %v", err)
		}
		review, err := svc.ReviewsSvc().Get(ctx, owner, "legacy-call")
		if err != nil {
			t.Fatalf("review Get: %v", err)
		}
		if note != nil && review != nil {
			if note.Body != "legacy note body" || note.Disposition != domain.DispositionAnswered.String() {
				t.Fatalf("imported note must keep its whole-slot snapshot: %+v", note)
			}
			if review.State != domain.ReviewPresent.String() || review.DraftID != "draft-legacy" || review.Confidence != 91 {
				t.Fatalf("imported review must keep its whole-slot snapshot: %+v", review)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("imported note/review never reached the journal projection: note=%+v review=%+v", note, review)
		}
		time.Sleep(10 * time.Millisecond)
	}

	sqlDB := mustSQL(t, application)
	for _, marker := range []string{flagCallNotesJournalImported, flagCallReviewsJournalImported} {
		done, err := store.FlagPresent(ctx, sqlDB, marker)
		if err != nil || !done {
			t.Fatalf("import marker %s must latch: done=%v err=%v", marker, done, err)
		}
	}
	if err := application.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}
