package arch

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestErrorCodeRegistryIsFresh keeps docs/error-contract.md's error-code
// registry honest with the source: every errorfamily code literal in
// non-test Go files must be documented, every documented code must still
// exist in source, and the family sets must match. The pin compares cell
// content with padding collapsed, so markdown formatters re-aligning the
// table's columns cannot break it — the auto-commit daemon's table reflow
// broke the byte-exact pin twice on 2026-10-05. The codes render as
// `[family:code]` in error strings and are grepped by journal recipes and
// the stack runbook, so a rename is a contract break — this test is the
// guard the family-adoption follow-up (f35) asked for. Regenerate the
// documented block after a deliberate code change:
//
//	go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update
//
// The writer emits the table in the shape the auto-commit daemon's
// markdown formatter leaves it (columns at max content width), so
// regeneration is idempotent and no longer churns the doc.
var updateRegistry = flag.Bool("update", false, "rewrite the generated error-code registry block in docs/error-contract.md")

const (
	registryDocPath = "../../docs/error-contract.md"
	registryBegin   = "<!-- error-code-registry: BEGIN (generated block; do not edit by hand; go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update rewrites it) -->"
	registryEnd     = "<!-- error-code-registry: END -->"
)

var (
	// Matches every constructor shape: New/Newf/Wrap/Wrapf/WrapOnce/WrapOncef
	// with an explicit family argument, plus the family-fixed helpers
	// New{Family}(f?) / Wrap{Family}(f?). Group 2 holds the fixed-family
	// suffix when present; group 3 is the argument text before the code and
	// group 4 the code itself (always the first quoted string after the
	// open paren — explicit-family forms pass errorfamily.X, an identifier,
	// and family-fixed forms pass err or nothing).
	ctorRe = regexp.MustCompile(`errorfamily\.((?:WrapOnce|Wrap|New)((?:Rejection|Conflict|Transient|Corruption|Infrastructure|Orchestration))?f?)\(([^)"]*)\"([^"]+)\"`)
	// Family argument inside group 3, e.g. `errorfamily.Rejection, `.
	familyArgRe = regexp.MustCompile(`errorfamily\.([A-Z][A-Za-z]*)\s*,`)
	// A plain identifier family argument (`Newf(family, ...)`) — the site
	// classifies at runtime; pbx.http/crm.http do the deliberate 4xx
	// Rejection / 5xx Transient split.
	identArgRe        = regexp.MustCompile(`^([a-z][A-Za-z0-9_]*)\s*,`)
	runtimeSplitLabel = "runtime-split"
	familyRe          = regexp.MustCompile(`^(Rejection|Conflict|Transient|Corruption|Infrastructure|Orchestration)$`)
	// P5: codes are stable `<seam>.<op>` names, lowercase snake.
	codeShapeRe = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z0-9_]+)+$`)
)

type codeEntry struct {
	families map[string]bool
	site     string
}

func TestErrorCodeRegistryIsFresh(t *testing.T) {
	entries := scanErrorFamilyCodes(t)

	docBytes, err := os.ReadFile(registryDocPath)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(docBytes)
	begin := strings.Index(doc, registryBegin)
	end := strings.Index(doc, registryEnd)
	if begin < 0 || end < 0 || end < begin {
		t.Fatalf("registry markers missing or misordered in %s", registryDocPath)
	}
	existing := parseRegistryRows(doc[begin+len(registryBegin) : end])
	generated := registryBlock(entries)

	var problems []string
	for code := range entries {
		if _, ok := existing[code]; !ok {
			problems = append(problems, "source code missing from the doc registry: "+code)
		}
	}
	for code, got := range existing {
		want, ok := entries[code]
		if !ok {
			problems = append(problems, "documented code has no source constructor: "+code)
			continue
		}
		if got != registryRow(code, want) {
			problems = append(problems, "registry drift for "+code+": doc has "+got)
		}
	}

	if *updateRegistry {
		rewritten := rewriteRegistryBlock(doc, generated)
		if err := os.WriteFile(registryDocPath, []byte(rewritten), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("rewrote the registry block in %s (%d codes)", registryDocPath, len(entries))
		return
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		t.Errorf(
			"error-code registry drifted from source (%d problems); regenerate with: go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update\n%s",
			len(problems), strings.Join(problems, "\n"),
		)
	}
}

// scanErrorFamilyCodes walks every non-test Go file in the repo and
// extracts the (code → families, first site) map, failing on literals
// that break the code contract.
func scanErrorFamilyCodes(t *testing.T) map[string]*codeEntry {
	t.Helper()
	entries := map[string]*codeEntry{}
	err := filepath.WalkDir(filepath.Join("..", ".."), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", ".direnv", "result":
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(filepath.ToSlash(path), "../../")
		for _, m := range ctorRe.FindAllStringSubmatch(string(data), -1) {
			code := m[4]
			if !codeShapeRe.MatchString(code) {
				t.Errorf("%s: suspicious errorfamily code %q — want a lowercase <seam>.<op> name (a mis-parse also lands here)", rel, code)
				continue
			}
			family := m[2]
			if family == "" {
				switch fa := familyArgRe.FindStringSubmatch(m[3]); {
				case fa != nil:
					family = fa[1]
				case identArgRe.MatchString(strings.TrimSpace(m[3])):
					family = runtimeSplitLabel
				default:
					t.Errorf("%s: cannot determine the family of code %q", rel, code)
					continue
				}
			}
			if family != runtimeSplitLabel && !familyRe.MatchString(family) {
				t.Errorf("%s: unknown errorfamily %q for code %q", rel, family, code)
				continue
			}
			entry := entries[code]
			if entry == nil {
				entry = &codeEntry{families: map[string]bool{}, site: rel}
				entries[code] = entry
			}
			entry.families[family] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no errorfamily codes found — the extractor likely broke")
	}
	return entries
}

// registryBlock renders the generated block the way the auto-commit
// daemon's markdown formatter leaves it: every column as wide as its
// widest cell (header included), cells left-aligned, the separator
// filled to the same width. Emitting that shape directly makes -update
// idempotent — the daemon has nothing left to re-pad, so regeneration
// no longer churns the doc (the 2026-10-05 CI reds came from exactly
// that churn in the byte-exact-pin era; the content-based pin made the
// drift inert, and the aligned writer removes the churn itself).
func registryBlock(entries map[string]*codeEntry) []string {
	codes := make([]string, 0, len(entries))
	for code := range entries {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	cells := make([][]string, 0, len(entries)+1)
	cells = append(cells, []string{"Code", "Families", "First site"})
	for _, code := range codes {
		cells = append(cells, []string{code, familyCell(entries[code]), entries[code].site})
	}
	widths := make([]int, len(cells[0]))
	for _, row := range cells {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	rows := make([]string, 0, len(cells)+1)
	for i, row := range cells {
		padded := make([]string, len(row))
		for j, cell := range row {
			padded[j] = cell + strings.Repeat(" ", widths[j]-len(cell))
		}
		rows = append(rows, "| "+strings.Join(padded, " | ")+" |")
		if i == 0 {
			sep := make([]string, len(row))
			for j := range row {
				sep[j] = strings.Repeat("-", widths[j])
			}
			rows = append(rows, "| "+strings.Join(sep, " | ")+" |")
		}
	}
	return rows
}

// rewriteRegistryBlock re-splices the generated rows between the markers
// with the daemon-stable framing: a blank line after BEGIN and before END.
// The auto-commit daemon's markdown formatter preserves those blank lines
// and the aligned rows leave nothing to re-pad, so -update output is
// byte-stable. Panic on missing markers: the freshness test validates them
// before the update branch runs.
func rewriteRegistryBlock(doc string, generated []string) string {
	begin := strings.Index(doc, registryBegin)
	end := strings.Index(doc, registryEnd)
	if begin < 0 || end < 0 || end < begin {
		panic("registry markers missing or misordered")
	}
	return doc[:begin+len(registryBegin)] + "\n\n" +
		strings.Join(generated, "\n") + "\n\n" + doc[end:]
}

// familyCell renders an entry's family set for both the comparison row
// and the writer: sorted, comma-joined.
func familyCell(entry *codeEntry) string {
	families := make([]string, 0, len(entry.families))
	for family := range entry.families {
		families = append(families, family)
	}
	sort.Strings(families)
	return strings.Join(families, ", ")
}

func registryRow(code string, entry *codeEntry) string {
	return canonicalRow(code, familyCell(entry), entry.site)
}

// canonicalRow rebuilds a table row from raw cell texts with all padding
// collapsed, making generated and documented rows comparable regardless of
// how a markdown formatter aligns the columns.
func canonicalRow(cells ...string) string {
	normalized := make([]string, len(cells))
	for i, cell := range cells {
		normalized[i] = strings.Join(strings.Fields(cell), " ")
	}
	return "| " + strings.Join(normalized, " | ") + " |"
}

// parseRegistryRows reads the generated table back: code → full row text.
func parseRegistryRows(block string) map[string]string {
	rows := map[string]string{}
	for line := range strings.SplitSeq(block, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		parts := strings.Split(strings.TrimSuffix(line, "|"), "|")
		if len(parts) != 4 {
			continue
		}
		code := strings.TrimSpace(parts[1])
		families := strings.TrimSpace(parts[2])
		if code == "" || code == "Code" || strings.HasPrefix(families, "--") {
			continue
		}
		rows[code] = canonicalRow(parts[1], parts[2], parts[3])
	}
	return rows
}

// The 2026-10-05 daemon-reflow breaks pinned this property: padding-only
// re-alignment of the registry table must stay inert, while any real cell
// change must still read as drift.
func TestRegistryRowComparisonIgnoresPadding(t *testing.T) {
	want := registryRow("blob.escape", &codeEntry{families: map[string]bool{"Rejection": true}, site: "internal/blob/store.go"})
	blocks := map[string]string{
		"generator-plain":      "| blob.escape | Rejection | internal/blob/store.go |\n",
		"formatter-reflowed":   "| blob.escape                                  | Rejection                 | internal/blob/store.go        |\n",
		"header-and-separator": "| Code | Families | First site |\n| ---- | -------- | ---------- |\n| blob.escape | Rejection | internal/blob/store.go |\n",
	}
	for name, block := range blocks {
		rows := parseRegistryRows(block)
		if got := rows["blob.escape"]; got != want {
			t.Errorf("%s: canonical form mismatch: got %q, want %q", name, got, want)
		}
	}
	drifted := parseRegistryRows("| blob.escape | Infrastructure | internal/blob/store.go |\n")
	if got := drifted["blob.escape"]; got == want {
		t.Errorf("a real family change must still read as drift (got the canonical form %q)", got)
	}
}

// The 2026-10-05 daemon-reflow churn pinned this property: the writer
// emits the formatter-aligned shape, so -update output is byte-stable
// (the daemon has nothing to re-pad) while still parsing back to the
// canonical comparison rows.
func TestRegistryBlockEmitsDaemonAlignedRows(t *testing.T) {
	entries := map[string]*codeEntry{
		"blob.escape":           {families: map[string]bool{"Rejection": true}, site: "internal/blob/store.go"},
		"messaging.send.failed": {families: map[string]bool{"Transient": true, "Rejection": true}, site: "internal/messaging/service.go"},
	}
	rows := registryBlock(entries)
	want := []string{
		"| Code                  | Families             | First site                    |",
		"| --------------------- | -------------------- | ----------------------------- |",
		"| blob.escape           | Rejection            | internal/blob/store.go        |",
		"| messaging.send.failed | Rejection, Transient | internal/messaging/service.go |",
	}
	if len(rows) != len(want) {
		t.Fatalf("registryBlock produced %d rows, want %d", len(rows), len(want))
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Errorf("row %d aligned-shape mismatch:\n got %q\nwant %q", i, rows[i], want[i])
		}
	}
	parsed := parseRegistryRows(strings.Join(rows, "\n") + "\n")
	for code, entry := range entries {
		if got := parsed[code]; got != registryRow(code, entry) {
			t.Errorf("aligned block must parse back to the canonical row for %s: got %q", code, got)
		}
	}
	// Uniform row length: equal-width columns equalize every row's byte
	// length, separator included. If a future writer edit breaks this, the
	// daemon regains something to re-pad and -update churn returns.
	for i, row := range rows {
		if len(row) != len(rows[0]) {
			t.Errorf("row %d has length %d, want the uniform %d of row 0", i, len(row), len(rows[0]))
		}
	}
	// Framing pin: the -update writer splices the block with blank lines
	// after BEGIN and before END; the daemon formatter keeps them, so the
	// surrounding bytes are as stable as the rows.
	doc := "preamble\n\n" + registryBegin + "\n| stale | rows |\n" + registryEnd + "\ntail\n"
	wantDoc := "preamble\n\n" + registryBegin + "\n\n" + strings.Join(rows, "\n") + "\n\n" + registryEnd + "\ntail\n"
	if got := rewriteRegistryBlock(doc, rows); got != wantDoc {
		t.Errorf("framing mismatch:\n got %q\nwant %q", got, wantDoc)
	}
}
