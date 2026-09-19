package server

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBoundedCheckTimesOutHangingCheck(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	check := boundedCheck("sqlite", 30*time.Millisecond, func() error {
		<-release
		return nil
	})
	err := check()
	if err == nil || !strings.Contains(err.Error(), "sqlite: timed out after 30ms") {
		t.Fatalf("want timeout error naming the check, got %v", err)
	}
}

func TestBoundedCheckPassesResultThrough(t *testing.T) {
	sentinel := errors.New("disk full")
	if err := boundedCheck("blob-dir", time.Second, func() error { return sentinel })(); !errors.Is(err, sentinel) {
		t.Fatalf("want the check's error unchanged, got %v", err)
	}
	if err := boundedCheck("blob-dir", time.Second, func() error { return nil })(); err != nil {
		t.Fatalf("want nil for a healthy check, got %v", err)
	}
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
