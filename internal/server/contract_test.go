package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSessionGatesLiveOnlyInTheHelper pins the convention the 22:55 dedup
// session established: every handler gates itself through h.RequireSession
// (safe by construction no matter how routes are wired). A session gate —
// a 401 rejection — may exist only in the helper (actions.go), and the
// wording has one home, session.SignInFirst. The shell's anonymous-friendly
// session.From enrichment (pages.go) is not a gate and stays allowed.
// Three more 401 sites exist by design, all rejecting BEFORE any session
// exists (different semantic from a missing-session gate): the carve's
// hooks package (webhooks.go — bad provider secret), session_api.go
// (login credentials rejected by the PBX directory), and passkey_api.go
// (the passkey login gate: unknown email or no enrolled credential — the
// uniform anti-enumeration 401, same credential class as session_api.go's).
func TestSessionGatesLiveOnlyInTheHelper(t *testing.T) {
	const helperFile = "actions.go"
	const hookFile = "../hooks/webhooks.go" // the secret gate: provider hooks 401 on a bad secret, not on a missing session
	const loginFile = "session_api.go"      // the login gate: bad credentials, not a missing session
	const passkeyFile = "passkey_api.go"    // the passkey login gate: unknown email / no credential, uniform by design
	sawHelper := false
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	// The carve moved the webhook ingest into internal/server/hooks (its
	// 401 writer travels with it) and the island JSON endpoints into
	// internal/server/api — those must ride the Server seam, ZERO inline
	// 401s there.
	hookEntries, err := filepath.Glob(filepath.Join("..", "hooks", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	entries = append(entries, hookEntries...)
	apiEntries, err := filepath.Glob(filepath.Join("..", "api", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range apiEntries {
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
		if n := strings.Count(code, "http.StatusUnauthorized"); n > 0 {
			t.Errorf("%s writes %d× 401 inline — carved api handlers gate through the Server seam only", name, n)
		}
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
		if n := strings.Count(code, "http.StatusUnauthorized"); n > 0 && name != helperFile && name != hookFile && name != loginFile && name != passkeyFile {
			t.Errorf("%s writes %d× 401 inline — session gates go through h.RequireSession only", name, n)
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
	// The delegated handler routes through the shell's typed closest helper
	// (closestHTMLElement), so the assertion names the whole call: the
	// SELECTOR alone would pass on a comment or an unused string.
	// Whitespace-free comparison: the formatter wraps these calls across
	// lines, and a pin that breaks on formatting is a pin people delete.
	if !strings.Contains(withoutSpace(string(body)), `closestHTMLElement(eventElement(event),"[data-reload]")`) {
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

// TestShellJSHandlesTranscribeButtons pins the shell asset to the
// server-rendered transcribe contract: voicemail.templ and messages.templ
// render data-transcribe-src buttons, and shell.js owns BOTH the
// delegated click handler and the auto-start pass behind them. Drop
// either side and every server-rendered transcription surface dies
// silently (the behavioral specs live island-side, shell.test.mjs).
func TestShellJSHandlesTranscribeButtons(t *testing.T) {
	c := newClient(t)
	resp, body := c.do(http.MethodGet, "/assets/shell.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("shell.js status %d", resp.StatusCode)
	}
	page := withoutSpace(string(body))
	for _, want := range []string{
		// No closing paren: the formatter decides whether this call gets a
		// trailing comma, and the pin is about the delegated lookup, not
		// about punctuation.
		`closestHTMLElement(eventElement(event),"[data-transcribe-src]"`, // the delegated click handler
		"autoTranscribe", // the auto-start pass
		"PBX_CONFIG.asr", // gated on the served seam flag
	} {
		if !strings.Contains(page, want) {
			t.Errorf("shell.js lost %q — server-rendered transcription would stop working", want)
		}
	}
}

// withoutSpace collapses every whitespace run out of a served asset. The
// served-byte pins below name whole call expressions (a bare selector would
// also match a comment); the formatter wraps those calls across lines, so
// comparing whitespace-free text keeps the pins exact without making them
// depend on prettier's line breaking.
func withoutSpace(s string) string {
	return strings.Join(strings.Fields(s), "")
}
