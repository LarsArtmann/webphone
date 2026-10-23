package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
)

// The styled 404 is the app's face for every wrong address: an unknown
// page, an absent optional surface (/health while the dashboard is off,
// with and without the trailing slash), and a missing asset the browser
// navigated to. The stdlib default ("404 page not found", 19 bytes of
// text/plain) is exactly what these pins keep off the HTML surfaces.
func TestUnknownPathsAnswerTheStyledNotFoundPage(t *testing.T) {
	server := newTestServer(t)

	for _, path := range []string{"/nope", "/health", "/health/", "/health/livez", "/health/readyz", "/health/startupz"} {
		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body := readAll(t, resp)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404", path, resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
			t.Errorf("%s: content-type %q, want text/html (bare stdlib text is the regression this pins)", path, got)
		}
		page := string(body)
		if !strings.Contains(page, "404 — page not found") {
			t.Errorf("%s: the page must name the 404, got %.200s", path, page)
		}
		if !strings.Contains(page, path) {
			t.Errorf("%s: the requested path must be echoed back, got %.200s", path, page)
		}
		if !strings.Contains(page, "Back to Messages") {
			t.Errorf("%s: the way back is missing, got %.200s", path, page)
		}
	}
}

// A 404 has no retry story: refreshing the same wrong address 404s again.
// The styled page therefore carries the way home instead of the generic
// reload button.
func TestNotFoundPageOffersNoReloadButton(t *testing.T) {
	server := newTestServer(t)

	resp, err := http.Get(server.URL + "/nope")
	if err != nil {
		t.Fatal(err)
	}
	page := string(readAll(t, resp))
	_ = resp.Body.Close()
	if strings.Contains(page, "data-reload") {
		t.Error("the 404 page must not offer Reload: it would reload the same wrong address")
	}
	if !strings.Contains(page, `href="/"`) {
		t.Error("the 404 page must link back to the app root")
	}
}

// Assets split by intent: a browser navigation (Accept: text/html) gets
// the app's own page, while a programmatic fetch gets the terse answer.
func TestMissingAssetAnswersStyledForNavigationsOnly(t *testing.T) {
	server := newTestServer(t)

	nav, err := http.NewRequest(http.MethodGet, server.URL+"/assets/nope.js", nil)
	if err != nil {
		t.Fatal(err)
	}
	nav.Header.Set("Accept", "text/html,application/xhtml+xml")
	navResp, err := http.DefaultClient.Do(nav)
	if err != nil {
		t.Fatal(err)
	}
	navBody := string(readAll(t, navResp))
	_ = navResp.Body.Close()
	if navResp.StatusCode != http.StatusNotFound {
		t.Fatalf("navigation: status %d, want 404", navResp.StatusCode)
	}
	if !strings.Contains(navBody, "404 — page not found") {
		t.Errorf("navigation to a missing asset must render the app page, got %.200s", navBody)
	}

	fetch, err := http.NewRequest(http.MethodGet, server.URL+"/assets/nope.js", nil)
	if err != nil {
		t.Fatal(err)
	}
	fetch.Header.Set("Accept", "application/javascript")
	fetchResp, err := http.DefaultClient.Do(fetch)
	if err != nil {
		t.Fatal(err)
	}
	fetchBody := string(readAll(t, fetchResp))
	_ = fetchResp.Body.Close()
	if fetchResp.StatusCode != http.StatusNotFound {
		t.Fatalf("fetch: status %d, want 404", fetchResp.StatusCode)
	}
	if strings.Contains(fetchBody, "404 — page not found") {
		t.Error("a programmatic fetch must not receive a whole HTML page")
	}
}

// The 404 echoes the requested path. It must not become a cache key: a
// shared cache would serve one visitor's crafted URL to the next.
func TestNotFoundPageIsNotCacheable(t *testing.T) {
	server := newTestServer(t)

	resp, err := http.Get(server.URL + "/nope")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store (the echoed path must not be cached)", got)
	}
}

// HEAD is the cheap prober form: the status and content type must be the
// real ones, the body must be absent.
func TestNotFoundAnswersHeadWithoutBody(t *testing.T) {
	server := newTestServer(t)

	resp, err := http.Head(server.URL + "/nope")
	if err != nil {
		t.Fatal(err)
	}
	body := readAll(t, resp)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status %d, want 404", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Errorf("content-type %q, want text/html", got)
	}
	if len(body) != 0 {
		t.Errorf("HEAD body = %d bytes, want none", len(body))
	}
}

// The error contract claims a method mismatch is unreachable without a
// valid CSRF token, so the user never meets a bare 405. This pins the
// claim: an unsafe request to a GET-only path dies at the CSRF gate.
func TestUnsafeRequestsFailAtCSRFBeforeMethodMismatch(t *testing.T) {
	server := newTestServer(t)

	resp, err := http.Post(server.URL+"/messages", "text/plain", strings.NewReader("x"))
	if err != nil {
		t.Fatal(err)
	}
	body := readAll(t, resp)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status %d, want 403 from the CSRF gate (a 405 here would break the documented contract)", resp.StatusCode)
	}
	if !strings.Contains(string(body), "csrf") {
		t.Errorf("body must name the CSRF rejection, got %.120s", body)
	}
}

// The query string is not part of the echoed path: a crafted `?redirect=`
// value must not round-trip into the page.
func TestNotFoundEchoesThePathWithoutTheQuery(t *testing.T) {
	server := newTestServer(t)

	resp, err := http.Get(server.URL + "/nope?next=https://evil.example")
	if err != nil {
		t.Fatal(err)
	}
	page := string(readAll(t, resp))
	_ = resp.Body.Close()
	if !strings.Contains(page, "/nope") {
		t.Error("the requested path must be echoed")
	}
	if strings.Contains(page, "evil.example") {
		t.Error("the query string must not be echoed into the page")
	}
}

// The 404 follows the negotiated language like every other page.
func TestNotFoundPageFollowsTheLanguage(t *testing.T) {
	server := newTestServer(t)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/nope", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	page := string(readAll(t, resp))
	_ = resp.Body.Close()
	for _, want := range []string{"404 — Seite nicht gefunden", "Zurück zu Nachrichten"} {
		if !strings.Contains(page, want) {
			t.Errorf("German 404 missing %q: %.200s", want, page)
		}
	}
}

// The dashboard serves its page at "/health"; the subtree pattern is for
// /health/sse and friends, and the bare "/health/" used to fall through to
// the stdlib text 404 even with the dashboard mounted. It must land on the
// page (a canonical redirect, not an error).
func TestDashboardSubtreeRootReachesThePage(t *testing.T) {
	dashboard := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("dashboard page"))
	})
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.Dashboard.Enable = true
	}, func(d *Deps) {
		d.Dashboard = dashboard
	})

	for _, path := range []string{"/health", "/health/"} {
		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body := readAll(t, resp)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: status %d, want 200 (the trailing slash must resolve)", path, resp.StatusCode)
		}
		if !strings.Contains(string(body), "dashboard page") {
			t.Errorf("%s: dashboard body missing, got %.120s", path, body)
		}
	}
}
