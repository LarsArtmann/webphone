package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
)

// seedLegacyContacts builds a v1-shaped legacy webphone.db: the exact
// DDL the §5 DROP would delete, with one row per classification class.
func seedLegacyContacts(t *testing.T, dir string) {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(dir, "webphone.db"))
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	defer func() { _ = db.Close() }()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS contacts (
			id         TEXT PRIMARY KEY,
			owner      TEXT NOT NULL,
			name       TEXT NOT NULL,
			phone      TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			UNIQUE(owner, phone)
		)`,
		`INSERT INTO contacts (id, owner, name, phone, created_at) VALUES
			('legacy-1', '1001', 'In Both',    '+49 170 111111', 1700000001),
			('legacy-2', '1001', 'Legacy Only', '+49 170 222222', 1700000002),
			('legacy-3', '1001', 'Garbage',     '!!!',            1700000003),
			('legacy-4', '1002', 'Second Owner', '+49 170 333333', 1700000004)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("seed legacy db: %v", err)
		}
	}
}

// seedNewWorld saves the in-both contact into a fresh contacts system
// over dir/system.db, then closes the system so the -check boot reads
// PERSISTED projections (the quiesced-server posture the DROP window
// mandates).
func seedNewWorld(t *testing.T, dir string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sys, err := cqrs.NewContactsSystem(ctx, filepath.Join(dir, "system.db"))
	if err != nil {
		t.Fatalf("boot new-world system: %v", err)
	}
	svc := cqrs.NewContacts(sys)
	if err := svc.Save(ctx, domain.Contact{
		ID:        domain.MustContactID("000000000000000000001"), // 21-char slot, the MustContactID contract
		Owner:     domain.MustParseExtension("1001"),
		Name:      "In Both",
		Phone:     domain.MustParsePhone("+49 170 111111"),
		CreatedAt: time.Unix(1700000099, 0),
	}); err != nil {
		t.Fatalf("save new-world contact: %v", err)
	}
	if err := sys.Close(); err != nil {
		t.Fatalf("close new-world system: %v", err)
	}
}

func TestExportLegacyDumpJSONVerifiesItsOwnSHA(t *testing.T) {
	dir := t.TempDir()
	seedLegacyContacts(t, dir)

	var buf bytes.Buffer
	if err := dumpLegacyJSON(&buf, mustReadLegacy(t, dir), "2026-11-20T00:00:00Z"); err != nil {
		t.Fatalf("dump json: %v", err)
	}
	var env legacyDumpEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("parse dump envelope: %v", err)
	}
	if env.Count != 4 || len(env.Rows) != 4 {
		t.Fatalf("count = %d rows = %d, want 4/4", env.Count, len(env.Rows))
	}
	payload, err := json.Marshal(env.Rows)
	if err != nil {
		t.Fatalf("re-marshal rows: %v", err)
	}
	sum := sha256.Sum256(payload)
	if env.RowsSHA256 != hex.EncodeToString(sum[:]) {
		t.Fatalf("rows_sha256 %s does not verify against the marshaled rows", env.RowsSHA256)
	}
	// The ISO rendering rides beside the raw stamp.
	if !strings.HasSuffix(env.Rows[0].CreatedAtISO, "Z") || env.Rows[0].CreatedAt != 1700000001 {
		t.Fatalf("row stamp shape wrong: %+v", env.Rows[0])
	}
}

func TestExportLegacyDumpCSVCarriesHeaderAndRows(t *testing.T) {
	dir := t.TempDir()
	seedLegacyContacts(t, dir)

	var buf bytes.Buffer
	if err := dumpLegacyCSV(&buf, mustReadLegacy(t, dir), "2026-11-20T00:00:00Z"); err != nil {
		t.Fatalf("dump csv: %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "# table=contacts count=4 sha256=") {
		t.Fatalf("csv header missing or wrong count: %q", strings.SplitN(out, "\n", 2)[0])
	}
	if got := strings.Count(out, "\n"); got != 6 { // header comment + csv header + 4 rows + trailing newline
		t.Fatalf("csv line count = %d, want 6:\n%s", got, out)
	}
}

func TestExportLegacyMissingDatabaseDumpsZeroRows(t *testing.T) {
	dir := t.TempDir() // no webphone.db at all
	rows, err := readLegacyContacts(dir)
	if err != nil {
		t.Fatalf("read missing db: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("missing db yielded %d rows, want 0", len(rows))
	}
	var buf bytes.Buffer
	if err := dumpLegacyJSON(&buf, rows, "2026-11-20T00:00:00Z"); err != nil {
		t.Fatalf("dump zero rows: %v", err)
	}
	if !strings.Contains(buf.String(), `"count": 0`) {
		t.Fatalf("zero-row dump missing count 0:\n%s", buf.String())
	}
}

func TestExportLegacyCheckExitsNonzeroWhileUnconfirmed(t *testing.T) {
	dir := t.TempDir()
	seedLegacyContacts(t, dir)
	seedNewWorld(t, dir)

	err := runExportLegacyCheck(dir, "")
	if err == nil {
		t.Fatal("check passed while 3 legacy-only rows are unconfirmed")
	}
	if !strings.Contains(err.Error(), "unconfirmed") {
		t.Fatalf("check error does not name the unconfirmed class: %v", err)
	}
}

func TestExportLegacyCheckPassesWhenConfirmationsCoverTheLosses(t *testing.T) {
	dir := t.TempDir()
	seedLegacyContacts(t, dir)
	seedNewWorld(t, dir)

	// Written confirmations for the three legacy-only rows — the phone
	// arrives in a DIFFERENT display form than the legacy row to prove
	// the digit-string match axis.
	confirmations := confirmationFile{Confirmed: []confirmedLoss{
		// Deliberately reformatted (same digit string, different display
		// form): the match axis is DigitsOf, and a trunk-form 0049… would
		// NOT cover a +49… row — the aggregate's honest boundary.
		{Owner: "1001", Phone: "+49/170-222222", Reason: "duplicate of the synced copy"},
		{Owner: "1001", Phone: "???", Reason: "unparseable garbage (empty digit key, like the row it covers)"},
		{Owner: "1002", Phone: "+49 170 333333", Reason: "owner left"},
	}}
	raw, err := json.Marshal(confirmations)
	if err != nil {
		t.Fatalf("marshal confirmations: %v", err)
	}
	path := filepath.Join(dir, "confirmations.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write confirmations: %v", err)
	}
	if err := runExportLegacyCheck(dir, path); err != nil {
		t.Fatalf("check with full coverage failed: %v", err)
	}
}

func TestExportLegacyCheckFreshInstallIsClean(t *testing.T) {
	dir := t.TempDir() // neither database exists
	if err := runExportLegacyCheck(dir, ""); err != nil {
		t.Fatalf("fresh-install check failed: %v", err)
	}
}

func mustReadLegacy(t *testing.T, dir string) []legacyContactRow {
	t.Helper()
	rows, err := readLegacyContacts(dir)
	if err != nil {
		t.Fatalf("read legacy contacts: %v", err)
	}
	return rows
}
