package assets

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The island type gate (tsconfig.json) checks a SUBSET of the asset tree: a
// file joins it only once its whole import closure is clean, because
// TypeScript type-checks every module an included module imports. The files
// that are not in yet are not hidden — they are the list below, by name, and
// this test is what makes the gate un-gameable: it refuses a shrunken include,
// a disabled checkJs/strict, a bare @ts-ignore, an unchecked file that is
// quietly dropped out of the debt list, or a nix check pointed at the
// never-green measuring project instead of the gate.
//
// Annotating a file is a two-line change: add it to tsconfig.json's include
// and delete its name from typecheckDebt.

// repoRoot is the module root: this package lives three directories below it,
// and every path below stays repo-relative so it matches tsconfig.json's own
// include entries byte for byte.
const repoRoot = "../../.."

// typecheckDebt is the island tree's declared untyped remainder. Every file
// here must exist and must NOT be reachable from tsconfig.json's include; the
// only legal edit is DELETE a line (add the file to the include first).
var typecheckDebt = []string{
	"internal/web/assets/island/app/calls.js",
	"internal/web/assets/island/app/connection.js",
	"internal/web/assets/island/app/main.js",
	"internal/web/assets/island/app/panels.js",
	"internal/web/assets/island/app/shortcuts.js",
	"internal/web/assets/island/app/transcribe.js",
	"internal/web/assets/shell.js",
}

// typecheckConfig is the subset of tsconfig.json this test reasons about.
type typecheckConfig struct {
	CompilerOptions map[string]any `json:"compilerOptions"`
	Include         []string       `json:"include"`
}

func loadTypecheckConfig(t *testing.T, name string) typecheckConfig {
	t.Helper()
	raw, err := os.ReadFile(path.Join(repoRoot, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	// tsconfig.json is JSONC: the comments carry the gate's rationale, so
	// strip them before unmarshalling.
	stripped := stripJSONComments(raw)
	var cfg typecheckConfig
	if err := json.Unmarshal([]byte(stripped), &cfg); err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return cfg
}

// stripJSONComments removes // and /* */ comments outside string literals —
// enough for a tsconfig file we own (no escaped quotes in keys or values).
func stripJSONComments(raw []byte) string {
	var out strings.Builder
	inString, inLine, inBlock := false, false, false
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case inLine:
			if c == '\n' {
				inLine = false
				out.WriteByte(c)
			}
		case inBlock:
			if c == '*' && i+1 < len(raw) && raw[i+1] == '/' {
				inBlock = false
				i++
			}
		case inString:
			out.WriteByte(c)
			if c == '\\' && i+1 < len(raw) {
				i++
				out.WriteByte(raw[i])
			} else if c == '"' {
				inString = false
			}
		case c == '"':
			inString = true
			out.WriteByte(c)
		case c == '/' && i+1 < len(raw) && raw[i+1] == '/':
			inLine = true
			i++
		case c == '/' && i+1 < len(raw) && raw[i+1] == '*':
			inBlock = true
			i++
		default:
			out.WriteByte(c)
		}
	}
	return out.String()
}

// assetSources is every served JS file the type gate could cover, as repo
// relative paths. It is derived from the EMBEDDED tree (the same bytes the
// browser gets), never from a hand-kept list, so a new module cannot join the
// repo without showing up as new debt.
func assetSources(t *testing.T) []string {
	t.Helper()
	var files []string
	err := fs.WalkDir(embedded, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".js") {
			return nil
		}
		// The vendored sip.js bundle is third-party bytes: never typed,
		// never annotated, and out of scope by construction.
		if strings.HasPrefix(p, "vendor/") {
			return nil
		}
		files = append(files, "internal/web/assets/"+p)
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded assets: %v", err)
	}
	sort.Strings(files)
	return files
}

// includeRoots expands tsconfig.json's include entries against the embedded
// tree: an entry may be a directory glob or a single file.
func includeRoots(t *testing.T, include []string) []string {
	t.Helper()
	seen := map[string]bool{}
	for _, entry := range include {
		// types/**/*.d.ts is not an asset source: the declarations are the
		// GATE's own input, not something the gate checks.
		if strings.HasPrefix(entry, "types/") {
			continue
		}
		matches, err := fs.Glob(os.DirFS(repoRoot), entry)
		if err != nil {
			t.Fatalf("include glob %q: %v", entry, err)
		}
		if len(matches) == 0 {
			t.Errorf("include entry %q matches no file — a stale entry silently widens the debt", entry)
		}
		for _, m := range matches {
			seen[m] = true
		}
	}
	roots := make([]string, 0, len(seen))
	for root := range seen {
		roots = append(roots, root)
	}
	sort.Strings(roots)
	return roots
}

var relativeImportRe = regexp.MustCompile(`from "\./([A-Za-z0-9_.-]+)"`)

// TestTypeGateConfigIsStillStrict pins the four compiler options the whole
// campaign rests on. Turning any of them off would let every existing error
// disappear without a single file being annotated.
func TestTypeGateConfigIsStillStrict(t *testing.T) {
	cfg := loadTypecheckConfig(t, "tsconfig.json")
	for _, want := range []string{"checkJs", "strict", "noEmit", "allowJs"} {
		got, ok := cfg.CompilerOptions[want]
		if !ok {
			t.Fatalf("tsconfig.json lost %q — the gate would stop checking", want)
		}
		if got != true {
			t.Fatalf("tsconfig.json %q = %v, want true", want, got)
		}
	}
	if len(cfg.Include) == 0 {
		t.Fatal("tsconfig.json include is empty")
	}
}

// TestTypeGateIncludeIsClosed asserts the include is import-closed. A listed
// file that imports an unlisted one still drags that file into the program, so
// a non-closed include means the "debt list" would lie about which files are
// unchecked.
func TestTypeGateIncludeIsClosed(t *testing.T) {
	cfg := loadTypecheckConfig(t, "tsconfig.json")
	roots := includeRoots(t, cfg.Include)
	listed := make(map[string]bool, len(roots))
	for _, root := range roots {
		listed[root] = true
	}
	for _, root := range roots {
		src, err := os.ReadFile(path.Join(repoRoot, root))
		if err != nil {
			t.Fatalf("read %s: %v", root, err)
		}
		for _, m := range relativeImportRe.FindAllStringSubmatch(string(src), -1) {
			target := path.Join(path.Dir(root), m[1])
			if !listed[target] {
				t.Errorf("%s imports %s, which is NOT in tsconfig.json's include — the import closure is open, so the debt list misreports what is checked", root, target)
			}
		}
	}
}

// TestTypeGateDebtMatchesTheTree is the monotonicity pin: the declared debt
// list must equal exactly the asset tree minus the include, so a file can only
// leave the debt by being annotated, and no untyped file can exist off the
// books.
func TestTypeGateDebtMatchesTheTree(t *testing.T) {
	cfg := loadTypecheckConfig(t, "tsconfig.json")
	checked := map[string]bool{}
	for _, root := range includeRoots(t, cfg.Include) {
		checked[root] = true
	}
	declared := map[string]bool{}
	for _, name := range typecheckDebt {
		if declared[name] {
			t.Errorf("typecheckDebt lists %s twice", name)
		}
		declared[name] = true
	}

	for _, file := range assetSources(t) {
		switch {
		case checked[file]:
			if declared[file] {
				t.Errorf("%s is BOTH in tsconfig.json's include and in typecheckDebt — delete the debt line", file)
			}
		case !declared[file]:
			t.Errorf("%s is neither checked nor declared as debt: add it to tsconfig.json's include (and make it type-clean) or to typecheckDebt — an untyped file off the books is how the gate rots", file)
		}
	}
	for name := range declared {
		if !checked[name] {
			if _, err := os.Stat(path.Join(repoRoot, name)); err != nil {
				t.Errorf("typecheckDebt lists %s, which does not exist: %v", name, err)
			}
		}
	}
}

// TestTypeGateRefusesBlanketSuppressions keeps the escape hatch honest: a bare
// @ts-ignore or @ts-nocheck silences every error on the next line (or in the
// whole file) with no reason recorded, which is exactly the "green by
// disabling" the gate exists to prevent. A reasoned suppression is written as
// @ts-expect-error with an explanation, and this test tolerates it.
func TestTypeGateRefusesBlanketSuppressions(t *testing.T) {
	for _, file := range assetSources(t) {
		src, err := os.ReadFile(path.Join(repoRoot, file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			switch {
			case strings.Contains(line, "@ts-nocheck"):
				t.Errorf("%s:%d has @ts-nocheck — one file of untyped code with no debt entry", file, i+1)
			case strings.Contains(line, "@ts-ignore"):
				t.Errorf("%s:%d has a bare @ts-ignore — use @ts-expect-error with a reason, or annotate", file, i+1)
			}
		}
	}
}

// TestTypeGateNixCheckUsesTheGate pins the wiring: the flake check must read
// tsconfig.json. Pointing it at tsconfig.all.json (the campaign's measuring
// project) would either fail forever or, worse, get deleted — and pointing it
// nowhere is how a gate becomes advisory.
func TestTypeGateNixCheckUsesTheGate(t *testing.T) {
	raw, err := os.ReadFile(path.Join(repoRoot, "nix/checks.nix"))
	if err != nil {
		t.Fatalf("read nix/checks.nix: %v", err)
	}
	nix := string(raw)
	if !strings.Contains(nix, "tsc -p tsconfig.json") && !strings.Contains(nix, "tsc -p tsconfig.json --noEmit") {
		t.Fatal("nix/checks.nix no longer runs `tsc -p tsconfig.json` — the type gate is unwired")
	}
	if strings.Contains(nix, "tsc -p tsconfig.all.json") {
		t.Fatal("nix/checks.nix runs the measuring project tsconfig.all.json — that project is red by design and must never gate")
	}
}
