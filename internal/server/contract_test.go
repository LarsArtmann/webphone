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
// Two more 401 sites exist by design, both rejecting BEFORE any session
// exists (different semantic from a missing-session gate): webhooks.go
// (bad provider secret) and session_api.go (login credentials rejected by
// the PBX directory).
func TestSessionGatesLiveOnlyInTheHelper(t *testing.T) {
	const helperFile = "actions.go"
	const hookFile = "webhooks.go"     // the secret gate: provider hooks 401 on a bad secret, not on a missing session
	const loginFile = "session_api.go" // the login gate: bad credentials, not a missing session
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
		if n := strings.Count(code, "http.StatusUnauthorized"); n > 0 && name != helperFile && name != hookFile && name != loginFile {
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

// TestShellJSSurfacesHtmxErrors pins the client half of the error-feedback
// story: htmx swaps NOTHING on error responses, so without the 3c handler
// a dead tab session (the session store now survives restarts, but hard
// crashes and manual cookie clears still 401) made every tab click and
// form submit fail silently. The behavioral specs live island-side
// (island-tests/shell.test.mjs); this greps the SERVED asset the way
// TestShellJSHandlesReloadButtons does, so an asset regression fails the
// Go build too.
//
// Mutation-proven 2026-09-20 (plan T05): blanking the "Tab session ended"
// marker in shell.js turned this test RED, restoring it GREEN — the gate
// can actually fail.
func TestShellJSSurfacesHtmxErrors(t *testing.T) {
	c := newClient(t)
	resp, body := c.do(http.MethodGet, "/assets/shell.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("shell.js status %d", resp.StatusCode)
	}
	page := string(body)
	for _, want := range []string{
		"htmx:responseError",
		"htmx:sendError",
		"HX-Trigger", // never double-toast server-authored feedback
		"Tab session ended",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("shell.js lost the htmx error surfacing: %q missing", want)
		}
	}
}
