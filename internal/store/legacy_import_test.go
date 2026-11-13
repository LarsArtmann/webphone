package store

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// The legacy tables stay in webphone.db (the migration ladder never
// rewrites shipped shape; docs/planning/2026-11-09_storage-ladder-retirement.md
// §4 rows the whole frozen family for the operator-gated final DROP).
// The transitional BOOT importers are gone since the 2026-11-09
// unification (7122de92 deleted internal/app/legacy.go) — the remaining
// pins cover the app-flag helpers the sync cursors ride, and the sibling
// legacy readers (faxes/snippets/transcripts) until their own deletion
// pass. The contacts reader itself is deleted (no production caller
// since the unification; its table stays for the §5 drop).

func seedAppFlag(t *testing.T, messages *Messages, key, value string) {
	t.Helper()
	if err := SetFlag(context.Background(), messages.db, key, value); err != nil {
		t.Fatalf("SetFlag(%q): %v", key, err)
	}
}

func TestAppFlagsRoundTripAndAbsence(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()

	if present, err := FlagPresent(ctx, messages.db, FlagLegacyContactsImported); err != nil || present {
		t.Fatalf("absent flag: present=%v err=%v, want false/nil", present, err)
	}
	if value, ok, err := FlagValue(ctx, messages.db, FlagLegacyContactsImported); err != nil || ok || value != "" {
		t.Fatalf("absent value: %q/%v/%v, want empty/false/nil", value, ok, err)
	}

	seedAppFlag(t, messages, FlagLegacyContactsImported, "2026-11-07")
	if present, err := FlagPresent(ctx, messages.db, FlagLegacyContactsImported); err != nil || !present {
		t.Fatalf("present flag: present=%v err=%v", present, err)
	}
	if value, ok, err := FlagValue(ctx, messages.db, FlagLegacyContactsImported); err != nil || !ok || value != "2026-11-07" {
		t.Fatalf("value = %q/%v/%v, want the stamped date", value, ok, err)
	}

	// SetFlag is insert-or-replace: the matrix sync cursor is rewritten every
	// poll, so a second write must UPDATE, not fail on the primary key.
	seedAppFlag(t, messages, FlagLegacyContactsImported, "2026-11-08")
	if value, _, err := FlagValue(ctx, messages.db, FlagLegacyContactsImported); err != nil || value != "2026-11-08" {
		t.Fatalf("after rewrite: %q/%v, want the newer value", value, err)
	}
}

func TestReadLegacyFaxesAndToFaxRefusesUnknownLifecycle(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	now := time.Now()

	// provider_ref is UNIQUE on the live table: each seed needs its own ref.
	insert := func(id, status, ref string) {
		t.Helper()
		if _, err := messages.db.ExecContext(ctx,
			`INSERT INTO fax_jobs (id, owner, remote, direction, status, pages, document_path, provider_ref, error, created_at, updated_at)
			 VALUES (?, '1001', '+441632960961', 'out', ?, 2, 'fax/probe.pdf', ?, '', ?, ?)`,
			id, status, ref, now.Unix(), now.Unix()); err != nil {
			t.Fatal(err)
		}
	}
	insert(domain.GenerateFaxID().String(), string(domain.FaxQueued), "ref-1")
	insert(domain.GenerateFaxID().String(), "teleported", "ref-2")

	rows, err := ReadLegacyFaxes(ctx, messages.db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want both legacy rows read", len(rows))
	}

	fax, ok := rows[0].ToFax()
	if !ok {
		t.Fatal("a queued fax must translate")
	}
	if fax.Status != domain.FaxQueued || fax.Pages != 2 || fax.DocumentPath.String() != "fax/probe.pdf" {
		t.Errorf("fax = %+v, want the row's own values", fax)
	}
	// An unknown lifecycle value must not reach the projection: it would
	// render as a state the machine table cannot advance.
	if _, ok := rows[1].ToFax(); ok {
		t.Error("an unknown fax status must be skipped, never imported")
	}
	if _, ok := (LegacyFax{Owner: "1001", Remote: "+441632960961", Direction: "sideways", Status: "queued"}).ToFax(); ok {
		t.Error("an unknown direction must be skipped")
	}
}

func TestReadLegacySnippetsAndToSnippet(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	now := time.Now()

	if _, err := messages.db.ExecContext(ctx,
		`INSERT INTO snippets (id, owner, body, quick, created_at) VALUES (?, '1001', 'on my way', 1, ?)`,
		domain.GenerateSnippetID().String(), now.Unix()); err != nil {
		t.Fatal(err)
	}
	rows, err := ReadLegacySnippets(ctx, messages.db)
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows = %d/%v, want 1", len(rows), err)
	}
	snippet, ok := rows[0].ToSnippet()
	if !ok {
		t.Fatal("a well-formed snippet must translate")
	}
	if !snippet.Quick || snippet.Body != "on my way" || !snippet.CreatedAt.Equal(time.Unix(now.Unix(), 0)) {
		t.Errorf("snippet = %+v, want quick+body+createdAt preserved", snippet)
	}
	if _, ok := (LegacySnippet{Owner: "---", Body: "x", Quick: 1}).ToSnippet(); ok {
		t.Error("an owner with no dialable characters must be skipped")
	}
	if demoted, ok := (LegacySnippet{ID: "bogus", Owner: "1001", Body: "x"}).ToSnippet(); !ok || demoted.Quick {
		t.Error("an unparseable id must demote and quick must default false")
	}
}

func TestReadLegacyTranscriptsAndToSegment(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	now := time.Now()

	insert := func(direction string) {
		t.Helper()
		if _, err := messages.db.ExecContext(ctx,
			`INSERT INTO call_transcripts (owner, call_id, direction, remote, started_at, text, created_at)
			 VALUES ('1001', 'opaque-call-id', ?, '+441632960961', ?, 'hello', ?)`,
			direction, now.Unix(), now.Unix()); err != nil {
			t.Fatal(err)
		}
	}
	insert(string(domain.DirectionInbound))
	insert("sideways")

	rows, err := ReadLegacyTranscripts(ctx, messages.db)
	if err != nil || len(rows) != 2 {
		t.Fatalf("rows = %d/%v, want both legacy rows read", len(rows), err)
	}
	segment, ok := rows[0].ToSegment()
	if !ok {
		t.Fatal("an inbound segment must translate")
	}
	// The CallID stays the island's opaque token verbatim — never branded,
	// never rewritten (strong-id wire ruling).
	if segment.CallID != "opaque-call-id" || segment.Direction != domain.DirectionInbound || segment.Text != "hello" {
		t.Errorf("segment = %+v, want the row's own values", segment)
	}
	if _, ok := rows[1].ToSegment(); ok {
		t.Error("an unknown direction must be skipped")
	}
	if _, ok := (LegacyTranscript{Owner: "---", CallID: "c", Direction: "in"}).ToSegment(); ok {
		t.Error("an owner with no dialable characters must be skipped")
	}
}
