package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSessionGatesLiveOnlyInTheHelper pins the convention the 22:55 dedup
// session established: every handler gates itself through h.requireSession
// (safe by construction no matter how routes are wired). A session gate —
// a 401 rejection — may exist only in the helper (actions.go), and the
// wording has one home, session.SignInFirst. The shell's anonymous-friendly
// session.From enrichment (pages.go) is not a gate and stays allowed.
func TestSessionGatesLiveOnlyInTheHelper(t *testing.T) {
	const helperFile = "actions.go"
	const hookFile = "webhooks.go" // the secret gate: provider hooks 401 on a bad secret, not on a missing session
	sawHelper := false
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range entries {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name) //nolint:gosec // test reads its own package sources
		if err != nil {
			t.Fatal(err)
		}
		code := string(src)
		if strings.Contains(code, `"sign in first"`) {
			t.Errorf("%s hardcodes the 401 body — use session.SignInFirst (one home per wording)", name)
		}
		if n := strings.Count(code, "http.StatusUnauthorized"); n > 0 && name != helperFile && name != hookFile {
			t.Errorf("%s writes %d× 401 inline — session gates go through h.requireSession only", name, n)
		}
		if strings.Contains(code, "session.From(") && name == helperFile {
			sawHelper = true
		}
	}
	if !sawHelper {
		t.Fatalf("%s no longer gates via session.From — the helper this test pins is gone", helperFile)
	}
}

// TestHtmxConfigMetaPrecedesScript pins the ORDER, not just the presence:
// htmx reads the htmx-config meta only when it is served before the htmx
// script. The 22:08 CSP session proved a meta after the script never
// disables the CSP-hostile inline indicator styles.
func TestHtmxConfigMetaPrecedesScript(t *testing.T) {
	c := newClient(t)
	resp, body := c.do(http.MethodGet, "/", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("page status %d", resp.StatusCode)
	}
	page := string(body)
	metaIdx := strings.Index(page, `"includeIndicatorStyles":false`)
	scriptIdx := strings.Index(page, `src="/htmx.min.js"`)
	if metaIdx == -1 || scriptIdx == -1 {
		t.Fatalf("served page incomplete: htmx-config meta at %d, htmx script at %d", metaIdx, scriptIdx)
	}
	if metaIdx > scriptIdx {
		t.Errorf("htmx-config meta (at %d) must precede the htmx script (at %d) — htmx would ignore it", metaIdx, scriptIdx)
	}
}

// TestShellJSHandlesReloadButtons pins the shell asset to the error
// panel's contract: error.templ renders a data-reload button (the CSP-safe
// replacement for its old inline onclick), and shell.js owns the delegated
// click handler for it. Drop either side and the reload button dies
// silently; TestStaticAssetsServe only covers data-dial.
func TestShellJSHandlesReloadButtons(t *testing.T) {
	c := newClient(t)
	resp, body := c.do(http.MethodGet, "/assets/shell.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("shell.js status %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), `closest("[data-reload]")`) {
		t.Error("shell.js lost the data-reload delegated handler — the error panel's reload button would stop working")
	}
}
