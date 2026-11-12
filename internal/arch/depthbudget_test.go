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
	"internal/cqrs": {loc: 6000, exports: 100},
	// The domain vocabulary: exported types ARE the product here (every
	// branded id, event and machine table lives in this package).
	"internal/domain": {loc: 2500, exports: 180},
	// The html residual (the hub pages.go/panels.go/actions.go + the
	// non-carved surfaces). The 2026-11-12 carve train already split
	// auth+platform out; the TODO_LIST carve row owns the html question
	// (ruled: only a real architecture change reopens it, never a
	// mechanical move).
	"internal/server": {loc: 5200, exports: 60},
	// The carved island JSON surface: fixture-heavy (wire DTOs + fuzz
	// corpora) but the ceiling still pins today's shape.
	"internal/server/api": {loc: 1900, exports: 60},
	// The SQL ladder is FROZEN (store/db.go header): sessions stay SQL,
	// everything else is import-only history the retention sweep
	// empties — the package only shrinks from here.
	"internal/store": {loc: 1900, exports: 60},
	// The composition root: one field per optional seam (the DI
	// surface's documented shape) plus the adapters.
	"internal/app": {loc: 1650, exports: 40},
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
	roots, err := os.ReadDir("internal")
	if err != nil {
		t.Fatalf("read internal/: %v", err)
	}
	for _, root := range roots {
		if !root.IsDir() || root.Name() == "testdata" {
			continue
		}
		base := filepath.Join("internal", root.Name())
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

	ceiling, exempt := depthBudgetNames[dir]
	if !exempt {
		ceiling = depthCeiling{loc: defaultMaxLOC, exports: defaultMaxExports}
	}
	rel := strings.TrimPrefix(dir, "internal/")
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
