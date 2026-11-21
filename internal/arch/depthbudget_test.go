package arch

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The depth budget (SUPERB deepening plan T07): package size is a
// NUMBER, not a narrative. Every internal/ package must fit the default
// ceilings — hand-written non-test LOC and top-level exported
// declarations — or appear in the exemption table below with its
// measured pin and a one-line reason. The table is a ratchet: raising a
// pinned ceiling is a deliberate registry-protocol act (edit the number
// AND the reason in the same commit), never a green-buying move; the
// direction of travel is DOWN (the 2026-11-12 carve train split server
// into auth/platform and the html residual is the named next carve).
//
// Generated code is out of scope by construction: *_templ.go is the
// compiler's output (the formatter excludes it the same way), and
// _test.go files are their packages' own safety nets, not depth.

// depthBudgetNames are the measured exemptions. `loc` and `exports` are
// CEILINGS pinned at (or just above) the 2026-11-12 measurements — a
// package that grows past its pin breaks this test, and the fix is a
// split (the carve precedent), not a bigger number.
var depthBudgetNames = map[string]depthCeiling{
	// The journal world: deciders + projections + machine tables. ONE
	// home by ratified ruling (the single go.mod + ONE journal world);
	// its size is the product of that ruling, not neglect.
	"cqrs": {loc: 6000, exports: 100},
	// The domain vocabulary: exported types ARE the product here (every
	// branded id, event and machine table lives in this package).
	"domain": {loc: 2500, exports: 180},
	// The html residual (the hub pages.go/panels.go/actions.go + the
	// non-carved surfaces). The 2026-11-12 carve train already split
	// auth+platform out; the TODO_LIST carve row owns the html question
	// (ruled: only a real architecture change reopens it, never a
	// mechanical move). 5200 -> 5250 on 2026-11-18: the line-ownership
	// unit (Display-rule consumption, active-first ordering, admin
	// chips) grew the identity renderers; 5250 -> 5300 on 2026-11-21:
	// the note-only query join (NotesGate.SearchBodies merge + the
	// shared transcriptReviewFields helper) rides in panels.go with its
	// feature; the number rides WITH the feature, not as silent drift —
	// direction stays DOWN (the T06 runway note owns the next split).
	"server": {loc: 5300, exports: 60},
	// The carved island JSON surface: fixture-heavy (wire DTOs + fuzz
	// corpora) but the ceiling still pins today's shape.
	"server/api": {loc: 1900, exports: 60},
	// The SQL ladder is FROZEN (store/db.go header): sessions stay SQL,
	// everything else is import-only history the retention sweep
	// empties — the package only shrinks from here.
	"store": {loc: 1900, exports: 60},
	// The composition root: one field per optional seam (the DI
	// surface's documented shape) plus the adapters. Raised 1650 → 1700
	// 2026-11-20 with the projection-DLQ observability decorator
	// (dlq.go, TC): the dead-letter seam is composition-root-owned by
	// doctrine; the store package's messages.go deletion the same day
	// did NOT credit app. Direction stays DOWN.
	"app": {loc: 1700, exports: 40},
	// One struct, one loader, per-section validation: every optional
	// seam adds its section + its validator (asr, push, chat, dial,
	// lens…). Named pin 2026-11-21 at the lens seam (1532 measured),
	// per the depth-pin runway note's ruling: the split candidate is
	// section-seams into internal/config/validate when the pin bites
	// again — NOT a package split (the one-struct loader is the point).
	"config": {loc: 1600, exports: 60},
}

const (
	defaultMaxLOC     = 1500
	defaultMaxExports = 60
)

type depthCeiling struct {
	loc     int
	exports int
}

// TestPackageDepthBudget walks every internal/ package and asserts the
// ceilings. Generated and test files are excluded; testdata is not a
// package.
func TestPackageDepthBudget(t *testing.T) {
	roots, err := os.ReadDir("..")
	if err != nil {
		t.Fatalf("read internal/: %v", err)
	}
	for _, root := range roots {
		if !root.IsDir() || root.Name() == "arch" || root.Name() == "testdata" {
			continue
		}
		base := filepath.Join("..", root.Name())
		pkgs, err := os.ReadDir(base)
		if err != nil {
			t.Fatalf("read %s: %v", base, err)
		}
		for _, pkg := range pkgs {
			if !pkg.IsDir() || pkg.Name() == "testdata" {
				continue
			}
			// Nested packages (internal/server/{api,auth,platform,...})
			// are measured at their own depth.
			dir := filepath.Join(base, pkg.Name())
			measureDepth(t, dir)
		}
		measureDepth(t, base)
	}
}

// measureDepth sums one directory's hand-written non-test sources and
// checks the applicable ceiling.
func measureDepth(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	var (
		loc     int
		exports int
		files   int
	)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_templ.go") {
			continue
		}
		files++
		raw, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("open %s/%s: %v", dir, name, err)
		}
		scanner := bufio.NewScanner(raw)
		for scanner.Scan() {
			loc++
			line := strings.TrimSpace(scanner.Text())
			for _, kw := range []string{"func ", "type ", "var ", "const "} {
				if strings.HasPrefix(line, kw) && len(line) > len(kw) {
					if c := line[len(kw)]; c >= 'A' && c <= 'Z' {
						exports++
					}
					break
				}
			}
		}
		if err := scanner.Err(); err != nil {
			t.Fatalf("scan %s/%s: %v", dir, name, err)
		}
		_ = raw.Close()
	}
	if files == 0 {
		return // a grouping directory (internal/ itself), not a package
	}

	key := strings.TrimPrefix(dir, "../")
	ceiling, exempt := depthBudgetNames[key]
	if !exempt {
		ceiling = depthCeiling{loc: defaultMaxLOC, exports: defaultMaxExports}
	}
	rel := key
	if loc > ceiling.loc {
		t.Errorf("%s: %d hand-written LOC exceeds the %d ceiling%s — split the package (the carve precedent: api/hooks/auth/platform) or lower its depth; raising the pin needs a reason in the exemption table's same commit",
			rel, loc, ceiling.loc, exemptionNote(exempt))
	}
	if exports > ceiling.exports {
		t.Errorf("%s: %d top-level exported declarations exceed the %d ceiling%s — the seam surface is growing; narrow it or pin a bigger budget with a reason",
			rel, exports, ceiling.exports, exemptionNote(exempt))
	}
}

func exemptionNote(exempt bool) string {
	if exempt {
		return " (a named exemption — the table comment owns its story)"
	}
	return ""
}

// templBudgetPins are the .templ-tree ceilings, in non-blank
// non-comment templ lines. The views tree is the repo's largest
// hand-written presentation artifact (3136 lines measured 2026-11-19,
// 3150 at the 2026-11-21 note-join train) while its Go side measures
// ~1200 — the Go-only budget above saw none of it. One pin for the one
// tree; a second templ home (a new surface package) gets its own row
// with a reason. Ratchet rules identical to the package budget.
var templBudgetPins = map[string]int{
	"web/views": 3300,
}

// TestTemplDepthBudget walks every internal/ .templ tree and asserts
// its pinned ceiling — the depth ratchet's templ column (registry act
// 2026-11-21: the row asked for a column or a views-only pin; the pin
// is the honest shape, because .templ files have no package boundary
// to hang a per-package number on).
func TestTemplDepthBudget(t *testing.T) {
	counts := map[string]int{}
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "testdata" || name == "arch" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".templ") {
			return nil
		}
		rel, err := filepath.Rel("..", path)
		if err != nil {
			return err
		}
		// Group by the internal/<tree> root (views has no subpackages
		// today; a nested surface joins its tree's count).
		parts := strings.Split(rel, string(filepath.Separator))
		key := strings.Join(parts[:2], "/")
		raw, err := os.Open(path)
		if err != nil {
			return err
		}
		defer raw.Close()
		scanner := bufio.NewScanner(raw)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "//") {
				counts[key]++
			}
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatalf("walk .templ trees: %v", err)
	}
	for key, n := range counts {
		ceiling, pinned := templBudgetPins[key]
		if !pinned {
			ceiling = 1000
		}
		if n > ceiling {
			t.Errorf("%s: %d templ lines exceed the %d ceiling%s — split the surface (a new views subpackage gets its own pin) or lower its depth; raising the pin needs a reason in this table's same commit",
				key, n, ceiling, exemptionNote(pinned))
		}
	}
}
