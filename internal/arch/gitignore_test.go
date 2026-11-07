// Package-level doc lives in arch_test.go.
package arch

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The 2026-11-06 near-miss this test exists for: `.gitignore` carried an
// UNANCHORED `reports/` rule, which matches a directory named `reports` at ANY
// depth — so the new `internal/reports` package was invisible to git while
// every local build passed. Only the sandbox build (which sees tracked files
// alone) noticed, with "no required module provides package". A name-based
// ignore rule can therefore delete a whole package from history, silently.
//
// The rule this test enforces: every directory the Go toolchain considers a
// package must be trackable. It is deliberately a WALK over `go list ./...`
// rather than a review of .gitignore, because the failure mode is a rule
// someone adds later for an unrelated reason.

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the module root")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
}

func TestGoPackageDirectoriesAreNotGitIgnored(t *testing.T) {
	root := moduleRoot(t)

	// The rules can only be evaluated inside a checkout; a sandbox build has
	// no .git and no git binary. Skipping is honest here (the CI go-tests job
	// runs in a checkout, so the gate does fire in the pipeline).
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		t.Skip("no .git under the module root — ignore rules cannot be evaluated")
	}
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not on PATH — ignore rules cannot be evaluated")
	}

	list := exec.Command("go", "list", "-f", "{{.Dir}}", "./...")
	list.Dir = root
	out, err := list.Output()
	if err != nil {
		t.Fatalf("go list ./...: %v", err)
	}
	dirs := strings.Fields(strings.TrimSpace(string(out)))
	if len(dirs) < 20 {
		t.Fatalf("go list returned %d package directories — the walk is not seeing the module", len(dirs))
	}

	// The historical victim: if this ever disappears from the walk, the test
	// has gone blind rather than the package having been deleted.
	sawReports := false
	for _, dir := range dirs {
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			t.Fatalf("rel(%q): %v", dir, err)
		}
		if rel == filepath.Join("internal", "reports") {
			sawReports = true
		}
		// git check-ignore exits 0 when the path IS ignored, 1 when it is
		// not (the good case), and 128 on a hard error. --no-index is
		// load-bearing: by default check-ignore stays silent for a path that
		// is already TRACKED, so a landed package would mask the rule that
		// would shadow the NEXT one (proved by planting `reports/`, which
		// the index silently excused).
		check := exec.Command(gitBin, "-C", root, "check-ignore", "--no-index", "--quiet", rel)
		checkErr := check.Run()
		if checkErr == nil {
			t.Errorf("package directory %q is GIT-IGNORED — a name-based .gitignore rule would drop it from history; anchor the rule to the repo root", rel)
			continue
		}
		var exitErr *exec.ExitError
		if !errors.As(checkErr, &exitErr) || exitErr.ExitCode() != 1 {
			t.Fatalf("git check-ignore %q: %v", rel, checkErr)
		}
	}
	if !sawReports {
		t.Fatal("internal/reports was not in the walk — the gate has gone blind")
	}
}
