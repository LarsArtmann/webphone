package store

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// The legacy tables stay in webphone.db after the 2026-10-15 move of the
// personal-contacts, fax, snippet and transcript truth into the metaengine
// read models, so an UPGRADED deployment's data rides the transitional reads
// in this package exactly once at boot. A regression here does not red a
// screen — it silently imports nothing and stamps the "imported" flag, which
// is the one failure mode the user cannot notice until the data is gone. That
// is why these readers get their own pins even though the tables are retired.

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

func TestReadLegacyContactsIsOldestFirstAndTranslates(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	now := time.Now()

	// Inserted newest-first on purpose: the reader's ORDER BY is the contract.
	// (owner, phone) is UNIQUE on the live table: two rows need two numbers.
	for _, row := range []struct {
		id      string
		name    string
		phone   string
		created time.Time
	}{
		{domain.GenerateContactID().String(), "newer", "+441632960961", now},
		{domain.GenerateContactID().String(), "older", "+441632960962", now.Add(-time.Hour)},
	} {
		if _, err := messages.db.ExecContext(ctx,
			`INSERT INTO contacts (id, owner, name, phone, created_at) VALUES (?, ?, ?, ?, ?)`,
			row.id, "1001", row.name, row.phone, row.created.Unix()); err != nil {
			t.Fatal(err)
		}
	}

	rows, err := ReadLegacyContacts(ctx, messages.db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Name != "older" || rows[1].Name != "newer" {
		t.Fatalf("order = %+v, want oldest first", rows)
	}

	contact, ok := rows[0].ToContact()
	if !ok {
		t.Fatal("a well-formed legacy row must translate")
	}
	if contact.Name != "older" || contact.Owner.String() != "1001" || contact.Phone.String() != "+441632960962" {
		t.Errorf("contact = %+v, want the row's own values", contact)
	}
	if contact.ID != domain.MustContactID(rows[0].ID) {
		t.Errorf("id = %s, want the stored id preserved", contact.ID)
	}
}

func TestToContactSkipsCorruptRowsAndDemotesTheID(t *testing.T) {
	good := LegacyContact{ID: "not-a-contact-id", Owner: "1001", Name: "n", Phone: "+441632960961"}
	contact, ok := good.ToContact()
	if !ok {
		t.Fatal("an unparseable ID must DEMOTE, not drop the contact")
	}
	if contact.ID.IsZero() {
		t.Error("a demoted id must still be minted, never empty")
	}
	if _, ok := (LegacyContact{Owner: "---", Phone: "+441632960961"}).ToContact(); ok {
		t.Error("an owner with no dialable characters must be skipped")
	}
	if _, ok := (LegacyContact{Owner: "1001", Phone: "---"}).ToContact(); ok {
		t.Error("a phone with no dialable characters must be skipped")
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
