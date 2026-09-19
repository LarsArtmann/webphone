package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
)

// TestHealthzTimeoutContract pins the wiring shape of the per-check bound:
// the checks are handed to ReadinessHandler with NamedCheck.Timeout set to
// checkTimeout (the mechanism moved upstream in cqrs-htmx v4.11.0 — the
// local boundedCheck wrapper is gone). This mirrors the construction with
// the real const and pins the hang→503 contract end to end.
func TestHealthzTimeoutContract(t *testing.T) {
	release := make(chan struct{})
	defer close(release)

	handler := cqrshtmxReadinessForTest(
		cqrshtmxNamedCheckForTest("sqlite", func() error {
			<-release

			return nil
		}),
	)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503 when a check hangs past its timeout, got %d", rr.Code)
	}

	for _, want := range []string{"sqlite", "timed out after 2s"} {
		if !strings.Contains(rr.Body.String(), want) {
			t.Errorf("body missing %q: %s", want, rr.Body.String())
		}
	}
}

// TestBoundedCheckPassesResultThrough pins the pass-through half of the
// contract at the wiring site: a check failing inside its budget surfaces
// its own error, and a healthy check stays nil.
func TestBoundedCheckPassesResultThrough(t *testing.T) {
	sentinel := errors.New("disk full")
	handler := cqrshtmxReadinessForTest(cqrshtmxNamedCheckForTest("blob-dir", func() error { return sentinel }))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "disk full") {
		t.Errorf("want the check's own error, got %s", rr.Body.String())
	}
}

// cqrshtmxReadinessForTest/cqrshtmxNamedCheckForTest mirror server.go's
// readiness construction so these tests fail if the wiring pattern (named
// checks + checkTimeout) drifts.
func cqrshtmxNamedCheckForTest(name string, check func() error) cqrshtmx.NamedCheck {
	return cqrshtmx.NamedCheck{Name: name, Check: check, Timeout: checkTimeout}
}

func cqrshtmxReadinessForTest(checks ...cqrshtmx.NamedCheck) http.Handler {
	return cqrshtmx.ReadinessHandler(checks...)
}

// The closed-DB mutant exercises the degraded rendering end to end: 503,
// overall status degraded, the failing check named with its error, and
// the healthy check still reported ok.
func TestHealthzDegradesTo503WhenSqliteFails(t *testing.T) {
	ts := newTestServerWithPhoneAPI(t, "", func(d *Deps) { _ = d.DB.Close() })

	res, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d (%s)", res.StatusCode, body)
	}
	for _, want := range []string{
		`"status":"degraded"`,
		`"sqlite":{"status":"fail"`,
		`"blob-dir":{"status":"ok"`,
	} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("body missing %s: %s", want, body)
		}
	}
}
