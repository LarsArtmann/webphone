package assets

import (
	"encoding/json/v2"
	"io/fs"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The island type gate (tsconfig.json) covers the WHOLE served asset tree: the
// island's ES modules, the shell glue, the theme preload and the enrollment
// page. This test is what makes the gate un-gameable — it refuses a shrunken
// include, a widened exclude, a disabled checkJs/strict, a bare @ts-ignore, an
// asset file that no include entry reaches, or a nix check pointed somewhere
// other than the gate.
//
// The single carve-out is the vendored sip.js bundle: third-party bytes, served
// verbatim, never typed. Adding a second exclusion is a test failure by design
// — excluding a red file is how a gate rots.

// repoRoot is the module root: this package lives three directories below it,
// and every path below stays repo-relative so it matches tsconfig.json's own
// include entries byte for byte.
const repoRoot = "../../.."

// vendoredBundle is the one legal exclusion: the pinned sip.js build.
const vendoredBundle = "internal/web/assets/vendor/**"

// vendoredDir / bundlePath describe what that carve-out must and must not
// reach: everything under the vendored directory, and the served bundle itself.
const (
	vendoredDir = "internal/web/assets/vendor/"
	bundlePath  = "internal/web/assets/vendor/sip.min.js"
)

// typecheckConfig is the subset of tsconfig.json this test reasons about.
type typecheckConfig struct {
	CompilerOptions map[string]any `json:"compilerOptions"`
	Include         []string       `json:"include"`
	Exclude         []string       `json:"exclude"`
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

// includeRoots expands tsconfig.json's include entries against the tree: an
// entry may be a directory glob or a single file. Entries matched by `exclude`
// are removed, because a file the compiler skips is not checked.
func includeRoots(t *testing.T, cfg typecheckConfig) []string {
	t.Helper()
	excluded := map[string]bool{}
	for _, p := range matchedPaths(t, cfg.Exclude) {
		excluded[p] = true
	}
	seen := map[string]bool{}
	for _, entry := range cfg.Include {
		// types/**/*.d.ts is not an asset source: the declarations are the
		// GATE's own input, not something the gate checks.
		if strings.HasPrefix(entry, "types/") {
			continue
		}
		matches := matchedPaths(t, []string{entry})
		if len(matches) == 0 {
			t.Errorf("include entry %q matches no file — a stale entry silently narrows the gate", entry)
		}
		for _, m := range matches {
			if excluded[m] {
				continue
			}
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

// matchedPaths expands tsconfig-style globs over the repo's asset tree.
// os.DirFS+fs.Glob cannot do this: fs.Glob speaks path.Match, where "**" is a
// single-segment "*" that never crosses a separator — the exact reason the
// include glob has to be interpreted here instead.
func matchedPaths(t *testing.T, patterns []string) []string {
	t.Helper()
	if len(patterns) == 0 {
		return nil
	}
	matchers := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		matchers[i] = globMatcher(t, pattern)
	}
	var out []string
	err := fs.WalkDir(os.DirFS(repoRoot), "internal/web/assets", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		for _, re := range matchers {
			if re.MatchString(p) {
				out = append(out, p)
				break
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk asset tree: %v", err)
	}
	sort.Strings(out)
	return out
}

// globMatcher compiles a tsconfig glob: "**/" matches zero or more directories
// (the TypeScript meaning), "*" stays inside one segment.
func globMatcher(t *testing.T, pattern string) *regexp.Regexp {
	t.Helper()
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch {
		case c == '*' && i+1 < len(pattern) && pattern[i+1] == '*':
			i++
			if i+1 < len(pattern) && pattern[i+1] == '/' {
				i++
				b.WriteString("(?:[^/]+/)*")
			} else {
				b.WriteString(".*")
			}
		case c == '*':
			b.WriteString("[^/]*")
		case c == '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		t.Fatalf("glob %q: %v", pattern, err)
	}
	return re
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
// an open closure means the include does not say what is checked.
func TestTypeGateIncludeIsClosed(t *testing.T) {
	cfg := loadTypecheckConfig(t, "tsconfig.json")
	roots := includeRoots(t, cfg)
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
				t.Errorf("%s imports %s, which is NOT in tsconfig.json's include — the import closure is open, so the include misreports what is checked", root, target)
			}
		}
	}
}

// TestTypeGateCoversEveryAssetSource is the monotonicity pin: the checked set
// must equal exactly the asset tree minus the vendored bundle. No untyped file
// can exist off the books, and no file can be dropped out of the gate without
// deleting it from the served tree.
func TestTypeGateCoversEveryAssetSource(t *testing.T) {
	cfg := loadTypecheckConfig(t, "tsconfig.json")
	checked := map[string]bool{}
	for _, root := range includeRoots(t, cfg) {
		checked[root] = true
	}

	for _, file := range assetSources(t) {
		if !checked[file] {
			t.Errorf("%s is served but NOT type-checked: widen tsconfig.json's include — an untyped file off the books is how the gate rots", file)
		}
	}
}

// TestTypeGateOnlyExcludesTheVendoredBundle pins the single carve-out. A second
// exclusion entry is how a red file gets hidden, so it is a test failure — and
// the entry itself must cover the vendored directory EXACTLY: a glob that also
// reaches a hand-written asset would silently un-type it.
func TestTypeGateOnlyExcludesTheVendoredBundle(t *testing.T) {
	cfg := loadTypecheckConfig(t, "tsconfig.json")
	if len(cfg.Exclude) != 1 || cfg.Exclude[0] != vendoredBundle {
		t.Fatalf("tsconfig.json exclude = %v, want exactly [%q] — excluding anything else hides untyped served code", cfg.Exclude, vendoredBundle)
	}

	matched := matchedPaths(t, cfg.Exclude)
	if len(matched) == 0 {
		t.Fatalf("exclude %q matches nothing — a vacuous carve-out that would let the bundle be typed by accident", vendoredBundle)
	}
	var hasBundle bool
	for _, p := range matched {
		if p == bundlePath {
			hasBundle = true
		}
		if !strings.HasPrefix(p, vendoredDir) {
			t.Errorf("exclude %q also reaches %s, which is a hand-written asset — that file would silently leave the gate", vendoredBundle, p)
		}
	}
	if !hasBundle {
		t.Errorf("exclude %q does not reach %s — the vendored bundle would enter the type program", vendoredBundle, bundlePath)
	}
}

// TestTypeGateIncludeStaysAGlob refuses a narrowed include. Listing individual
// files works right up until the next asset is added and joins the program as
// an unannotated import; a tree glob cannot miss. The only legal single-file
// entries point at the declarations (types/), which are the gate's input rather
// than something it checks.
func TestTypeGateIncludeStaysAGlob(t *testing.T) {
	cfg := loadTypecheckConfig(t, "tsconfig.json")
	var sawAssetGlob bool
	for _, entry := range cfg.Include {
		if strings.HasPrefix(entry, "types/") {
			if !strings.Contains(entry, "**") {
				t.Errorf("include entry %q names the declarations without a glob — a new .d.ts would silently stop being loaded", entry)
			}
			continue
		}
		if !strings.HasPrefix(entry, "internal/web/assets/") {
			t.Errorf("include entry %q is outside the served asset tree — the gate is scoped to what the browser gets", entry)
			continue
		}
		if !strings.Contains(entry, "**") {
			t.Errorf("include entry %q is a file or single-level pattern; a new asset module would slip in untyped — use a ** glob", entry)
		}
		sawAssetGlob = true
	}
	if !sawAssetGlob {
		t.Fatal("tsconfig.json has no ** glob over internal/web/assets — the served tree is not covered")
	}
}

// islandDomIDs parses the ids the gate declares for the island's typed lookups
// (types/island.d.ts, interface IslandDomIds). Both quoted and bare property
// names are legal in the declaration, so both are read.
func islandDomIDs(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(path.Join(repoRoot, "types", "island.d.ts"))
	if err != nil {
		t.Fatalf("read types/island.d.ts: %v", err)
	}
	const open = "interface IslandDomIds {"
	body := string(raw)
	begin := strings.Index(body, open)
	if begin < 0 {
		t.Fatalf("types/island.d.ts lost %q", open)
	}
	end := strings.Index(body[begin:], "\n}")
	if end < 0 {
		t.Fatal("types/island.d.ts has an unterminated IslandDomIds interface")
	}
	var ids []string
	for _, line := range strings.Split(body[begin+len(open):begin+end], "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		name, _, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("cannot parse IslandDomIds member %q", line)
		}
		ids = append(ids, strings.Trim(strings.TrimSpace(name), `"`))
	}
	if len(ids) < 20 {
		t.Fatalf("types/island.d.ts declares only %d island ids — truncated?", len(ids))
	}
	return ids
}

// documentedDomIDs reads the fenced id block of docs/dom-contract.md, the list's
// one home (the same block internal/server parses).
func documentedDomIDs(t *testing.T) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(path.Join(repoRoot, "docs", "dom-contract.md"))
	if err != nil {
		t.Fatalf("read docs/dom-contract.md: %v", err)
	}
	const beginMark, endMark = "<!-- dom-contract:begin -->", "<!-- dom-contract:end -->"
	body := string(raw)
	begin := strings.Index(body, beginMark)
	end := strings.Index(body, endMark)
	if begin < 0 || end < begin {
		t.Fatal("docs/dom-contract.md lost its dom-contract markers")
	}
	ids := map[string]bool{}
	for _, line := range strings.Split(body[begin+len(beginMark):end], "\n") {
		if id := strings.TrimSpace(line); id != "" {
			ids[id] = true
		}
	}
	if len(ids) < 20 {
		t.Fatalf("docs/dom-contract.md carries only %d ids — truncated?", len(ids))
	}
	return ids
}

// TestIslandDomIdsMirrorTheDomContract pins the third mirror of the island's id
// list. The document is the source, the served markup is asserted against it by
// internal/server, and the type declarations are what the COMPILER checks — so a
// renamed element has to move in all three, or typing catches the stale one here
// instead of a month later as a silent null.
//
// The relation is one-directional on purpose: the document also lists
// shell-owned ids (wp-live, wp-tab-skeleton) and the markup-only call-btn, which
// the island never looks up by id and therefore has no typed handle for.
func TestIslandDomIdsMirrorTheDomContract(t *testing.T) {
	documented := documentedDomIDs(t)
	for _, id := range islandDomIDs(t) {
		if !documented[id] {
			t.Errorf("types/island.d.ts declares id %q, which docs/dom-contract.md does not carry — a renamed element would keep type-checking against a phantom id", id)
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
// tsconfig.json. Pointing it at a second, redder project would either fail
// forever or, worse, get deleted — and pointing it nowhere is how a gate
// becomes advisory.
func TestTypeGateNixCheckUsesTheGate(t *testing.T) {
	raw, err := os.ReadFile(path.Join(repoRoot, "nix/checks.nix"))
	if err != nil {
		t.Fatalf("read nix/checks.nix: %v", err)
	}
	nix := string(raw)
	if !strings.Contains(nix, "tsc -p tsconfig.json") && !strings.Contains(nix, "tsc -p tsconfig.json --noEmit") {
		t.Fatal("nix/checks.nix no longer runs `tsc -p tsconfig.json` — the type gate is unwired")
	}
	for _, forbidden := range []string{"tsconfig.all.json", "tsconfig.debt.json"} {
		if strings.Contains(nix, forbidden) {
			t.Fatalf("nix/checks.nix runs %s — a second project is red by design and must never gate", forbidden)
		}
	}
}

// TestGlobMatcher pins the tsconfig glob semantics the coverage pins rely on:
// "**/" spans zero or more directories (TypeScript's meaning), "*" stays within
// one segment. This matcher once returned nothing for every pattern — Go's
// fs.Glob speaks path.Match, where "**" cannot cross a separator — and the
// coverage pin then "passed" over an empty set.
func TestGlobMatcher(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"internal/web/assets/**/*.js", "internal/web/assets/shell.js", true},
		{"internal/web/assets/**/*.js", "internal/web/assets/island/app/main.js", true},
		{"internal/web/assets/**/*.js", "internal/web/assets/island/app/deep/x.js", true},
		{"internal/web/assets/**/*.js", "internal/web/assets/style.css", false},
		{"internal/web/assets/**/*.js", "internal/web/assets/island/app/main.ts", false},
		{"internal/web/assets/vendor/**", "internal/web/assets/vendor/sip.min.js", true},
		{"internal/web/assets/vendor/**", "internal/web/assets/vendor/nested/x.txt", true},
		{"internal/web/assets/vendor/**", "internal/web/assets/shell.js", false},
		{"types/**/*.d.ts", "types/island.d.ts", true},
		{"types/**/*.d.ts", "types/nested/island.d.ts", true},
		{"types/*.d.ts", "types/nested/island.d.ts", false},
		{"*.js", "shell.js", true},
		{"*.js", "app/shell.js", false},
	}
	for _, tc := range cases {
		re := globMatcher(t, tc.pattern)
		if got := re.MatchString(tc.path); got != tc.want {
			t.Errorf("globMatcher(%q).MatchString(%q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
		}
	}
}
