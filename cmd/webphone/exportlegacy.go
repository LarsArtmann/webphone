package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	_ "modernc.org/sqlite" // registers the "sqlite" driver for the read-only legacy handle

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
)

// runExportLegacy serves the `webphone export-legacy` one-shot: the §5
// DROP-window safety tool (docs/planning/2026-11-14_t14-export-or-confirm-design.md).
// Two modes share the flag set:
//
//	webphone export-legacy -table contacts -format json|csv [-data DIR]
//	    dumps every legacy row with a self-describing header (count,
//	    sha256 over the row payload, timestamp) — archive this beside
//	    the deployment record before any DROP.
//	webphone export-legacy -check [-confirmations PATH] [-data DIR]
//	    classifies legacy rows against the journal projection by the
//	    digit-string key and exits NONZERO while legacy-only rows
//	    exist without a confirmation record covering them.
//
// The legacy side is RAW read-only sqlite (never store.Open — a dump
// binary at a different schema version must not execute migrations,
// including the §5 DROP itself); a missing webphone.db dumps zero
// rows, which is the fresh-install proof of absence.
func runExportLegacy(args []string) error {
	fs := flag.NewFlagSet("export-legacy", flag.ExitOnError)
	table := fs.String("table", "contacts", "legacy table to dump (contacts; the same-class family extends later)")
	format := fs.String("format", "json", "dump format: json | csv")
	dataDir := fs.String("data", "", "data directory (default: the loaded config's data_dir)")
	check := fs.Bool("check", false, "classify legacy rows against the journal projection instead of dumping")
	confirmations := fs.String("confirmations", "", "confirmation file covering legacy-only rows (-check only)")
	_ = fs.Parse(args)

	if *check { //nolint:branching-flow // flag.Bool never returns a nil pointer
		dir, err := resolveDataDir(*dataDir)
		if err != nil {
			return err
		}
		return runExportLegacyCheck(dir, *confirmations)
	}
	if *table != "contacts" { //nolint:branching-flow // flag.String never returns a nil pointer
		return bootErr(bootConfigInvalid, staticBootError(fmt.Sprintf("table %q is not dumpable yet (contacts is the pilot; the unswept-imported family extends here)", *table)))
	}
	if *format != "json" && *format != "csv" { //nolint:branching-flow // flag.String never returns a nil pointer
		return bootErr(bootConfigInvalid, staticBootError(fmt.Sprintf("format %q is not one of json | csv", *format)))
	}
	dir, err := resolveDataDir(*dataDir)
	if err != nil {
		return err
	}
	return runExportLegacyDump(dir, *format)
}

// resolveDataDir prefers the explicit -data flag and falls back to the
// loaded config (env nesting applies: WEBPHONE_DATA_DIR).
func resolveDataDir(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	cfg, err := config.Load()
	if err != nil {
		return "", bootErr(bootConfigInvalid, propagatef("load config: %w", err))
	}
	return cfg.DataDir, nil
}

// legacyContactRow is one raw legacy contacts row, verbatim values plus
// the ISO rendering of the unix-second stamp.
type legacyContactRow struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	CreatedAt    int64  `json:"created_at"`
	CreatedAtISO string `json:"created_at_iso"`
}

// readLegacyContacts SELECTs the v1 contacts shape read-only. A missing
// database is ZERO rows (the fresh-install proof of absence), any other
// open/query failure refuses the tool.
func readLegacyContacts(dir string) ([]legacyContactRow, error) {
	path := filepath.Join(dir, "webphone.db")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "stat "+path))
	}
	// query_only(1): the dump must be inert at ANY point in the schema
	// chain — no migrations, no WAL stamping (design memo §2).
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=query_only(1)")
	if err != nil {
		return nil, bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "open "+path+" read-only"))
	}
	defer func() { _ = db.Close() }()
	rows, err := db.Query(`SELECT id, owner, name, phone, created_at FROM contacts ORDER BY owner, created_at, id`)
	if err != nil {
		return nil, bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "select legacy contacts"))
	}
	defer func() { _ = rows.Close() }()
	out := []legacyContactRow{}
	for rows.Next() {
		var r legacyContactRow
		if err := rows.Scan(&r.ID, &r.Owner, &r.Name, &r.Phone, &r.CreatedAt); err != nil {
			return nil, bootErr(bootGeneric, errorfamily.WrapCorruption(err, "webphone.cli.export_legacy", "scan legacy contact row"))
		}
		r.CreatedAtISO = time.Unix(r.CreatedAt, 0).UTC().Format(time.RFC3339)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "iterate legacy contacts"))
	}
	return out, nil
}

// legacyDumpEnvelope is the JSON dump's self-describing shape; the
// sha256 covers the marshaled rows slice so the envelope is verifiable
// without the envelope itself.
type legacyDumpEnvelope struct {
	Table      string             `json:"table"`
	Generated  string             `json:"generated"`
	Count      int                `json:"count"`
	RowsSHA256 string             `json:"rows_sha256"`
	Rows       []legacyContactRow `json:"rows"`
}

func runExportLegacyDump(dir, format string) error {
	rows, err := readLegacyContacts(dir)
	if err != nil {
		return err
	}
	generated := time.Now().UTC().Format(time.RFC3339)
	if format == "csv" {
		return dumpLegacyCSV(os.Stdout, rows, generated)
	}
	return dumpLegacyJSON(os.Stdout, rows, generated)
}

func dumpLegacyJSON(w io.Writer, rows []legacyContactRow, generated string) error {
	payload, err := json.Marshal(rows)
	if err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "marshal rows"))
	}
	sum := sha256.Sum256(payload)
	env := legacyDumpEnvelope{
		Table:      "contacts",
		Generated:  generated,
		Count:      len(rows),
		RowsSHA256: hex.EncodeToString(sum[:]),
		Rows:       rows,
	}
	out, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "marshal dump envelope"))
	}
	_, err = fmt.Fprintln(w, string(out))
	return err
}

func dumpLegacyCSV(w io.Writer, rows []legacyContactRow, generated string) error {
	body, err := csvBody(rows)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(body)
	_, err = fmt.Fprintf(w, "# table=contacts count=%d sha256=%s generated=%s\n", len(rows), hex.EncodeToString(sum[:]), generated)
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

// csvBody renders the CSV payload (header row + one row per contact)
// into memory first: the sha256 header must precede a hash of exactly
// these bytes.
func csvBody(rows []legacyContactRow) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"id", "owner", "name", "phone", "created_at", "created_at_iso"}); err != nil {
		return nil, bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "write csv header"))
	}
	for _, r := range rows {
		if err := w.Write([]string{r.ID, r.Owner, r.Name, r.Phone, fmt.Sprintf("%d", r.CreatedAt), r.CreatedAtISO}); err != nil {
			return nil, bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "write csv row"))
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "flush csv body"))
	}
	return buf.Bytes(), nil
}

// ─── -check: classify legacy rows against the journal world ─────────

// confirmedLoss is one record in the confirmation file: a per-owner
// written loss confirmation naming the row by its phone (matched on
// the digit string, the same region-free axis the aggregate upserts
// on) plus the human reason the loss is accepted.
type confirmedLoss struct {
	Owner  string `json:"owner"`
	Phone  string `json:"phone"`
	Reason string `json:"reason,omitempty"`
}

type confirmationFile struct {
	DumpSHA256 string          `json:"dump_sha256,omitempty"`
	Confirmed  []confirmedLoss `json:"confirmed"`
}

// checkClassification is the -check verdict: every legacy row lands in
// exactly one of in-both / legacy-only (parseable) / unparseable, and
// the exit gate is the UNCONFIRMED legacy-only count.
type checkClassification struct {
	LegacyTotal    int
	InBoth         int
	LegacyOnly     int
	Unparseable    int
	NewWorldTotal  int64
	ConfirmedCount int
	Unconfirmed    []legacyContactRow
}

// loadConfirmations parses the optional confirmation file; an empty
// path is simply no confirmations.
func loadConfirmations(path string) (map[string]bool, error) {
	if path == "" {
		return map[string]bool{}, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, bootErr(bootGeneric, errorfamily.WrapInfrastructure(err, "webphone.cli.export_legacy", "read confirmations "+path))
	}
	var file confirmationFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, bootErr(bootGeneric, errorfamily.WrapCorruption(err, "webphone.cli.export_legacy", "parse confirmations "+path))
	}
	covered := map[string]bool{}
	for _, c := range file.Confirmed {
		covered[confirmKey(c.Owner, c.Phone)] = true
	}
	return covered, nil
}

// confirmKey is the confirmation match axis: owner + the phone's digit
// string — the same key the aggregate upserts on, so a confirmation
// written against any display form of the number covers the row.
func confirmKey(owner, phone string) string {
	return owner + "|" + domain.DigitsOf(phone)
}

func runExportLegacyCheck(dir, confirmationsPath string) error {
	rows, err := readLegacyContacts(dir)
	if err != nil {
		return err
	}
	covered, err := loadConfirmations(confirmationsPath)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	sys, err := cqrs.NewContactsSystem(ctx, filepath.Join(dir, "system.db"))
	if err != nil {
		return propagatef("open contacts check system: %w", err)
	}
	defer func() {
		if err := sys.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "export-legacy: contacts check system shutdown failed: %v\n", err)
		}
	}()
	svc := cqrs.NewContacts(sys)
	classification, err := classifyLegacyContacts(ctx, svc, rows, covered)
	if err != nil {
		return err
	}
	printCheckReport(os.Stdout, classification, confirmationsPath)
	if len(classification.Unconfirmed) > 0 {
		return bootErr(bootGeneric, staticBootError(fmt.Sprintf(
			"%d legacy-only contact row(s) remain unconfirmed — re-import them (a re-save in the UI re-classifies them as in-both) or record per-owner written confirmations via -confirmations before the §5 DROP window",
			len(classification.Unconfirmed))))
	}
	return nil
}

// rowClass is one legacy row's classification before confirmations:
// unparseable rows can never auto-migrate, the rest split by the digit
// key's presence in the owner's new-world list.
func rowClass(r legacyContactRow, newWorld map[string]bool) string {
	if _, err := domain.ParseExtension(r.Owner); err != nil {
		return "unparseable"
	}
	if _, err := domain.ParsePhone(r.Phone); err != nil {
		return "unparseable"
	}
	if newWorld[confirmKey(r.Owner, r.Phone)] {
		return "in-both"
	}
	return "legacy-only"
}

// classifyLegacyContacts walks the legacy rows once: rows whose owner
// or phone do not parse are unparseable (they can never auto-migrate),
// the rest split by presence of the digit key in the owner's new-world
// list; legacy-only rows are confirmed or unconfirmed by the file.
func classifyLegacyContacts(ctx context.Context, svc *cqrs.Contacts, rows []legacyContactRow, covered map[string]bool) (checkClassification, error) {
	classification := checkClassification{}
	newWorld := map[string]bool{}
	owners := map[string]bool{}
	for _, r := range rows {
		owners[r.Owner] = true
	}
	for owner := range owners {
		ext, err := domain.ParseExtension(owner)
		if err != nil {
			continue // no new-world list can exist for an unparseable owner
		}
		listed, err := svc.List(ctx, ext)
		if err != nil {
			return classification, propagatef("list new-world contacts for %s: %w", owner, err)
		}
		for _, c := range listed {
			newWorld[confirmKey(c.Owner.String(), c.Phone.String())] = true
		}
	}
	total, err := svc.Count(ctx)
	if err != nil {
		return classification, propagatef("count new-world contacts: %w", err)
	}
	classification.NewWorldTotal = total
	classification.LegacyTotal = len(rows)
	for _, r := range rows {
		switch rowClass(r, newWorld) {
		case "in-both":
			classification.InBoth++
			continue // present in the journal world — no confirmation needed
		case "legacy-only":
			classification.LegacyOnly++
		default:
			classification.Unparseable++
		}
		if covered[confirmKey(r.Owner, r.Phone)] {
			classification.ConfirmedCount++
			continue
		}
		classification.Unconfirmed = append(classification.Unconfirmed, r)
	}
	sort.Slice(classification.Unconfirmed, func(i, j int) bool {
		if classification.Unconfirmed[i].Owner != classification.Unconfirmed[j].Owner {
			return classification.Unconfirmed[i].Owner < classification.Unconfirmed[j].Owner
		}
		return classification.Unconfirmed[i].Phone < classification.Unconfirmed[j].Phone
	})
	return classification, nil
}

// printCheckReport renders the verdict a human archives beside the
// dump: totals first, then every unconfirmed row named per owner.
func printCheckReport(w io.Writer, c checkClassification, confirmationsPath string) {
	source := "(no confirmation file)"
	if confirmationsPath != "" {
		source = confirmationsPath
	}
	fmt.Fprintf(w, "legacy contacts check\n")
	fmt.Fprintf(w, "  legacy total        %d\n", c.LegacyTotal)
	fmt.Fprintf(w, "  in both worlds      %d\n", c.InBoth)
	fmt.Fprintf(w, "  legacy-only         %d\n", c.LegacyOnly)
	fmt.Fprintf(w, "  unparseable         %d\n", c.Unparseable)
	fmt.Fprintf(w, "  new-world total     %d\n", c.NewWorldTotal)
	fmt.Fprintf(w, "  confirmed losses    %d  %s\n", c.ConfirmedCount, source)
	fmt.Fprintf(w, "  unconfirmed         %d\n", len(c.Unconfirmed))
	for _, r := range c.Unconfirmed {
		fmt.Fprintf(w, "    owner=%s phone=%q name=%q created=%s\n", r.Owner, r.Phone, r.Name, r.CreatedAtISO)
	}
}
