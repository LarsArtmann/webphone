package server

import (
	"encoding/json/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/store"
)

// TestHealthzReportsOkWhenBackingResourcesAnswer pins the honest-healthz
// contract (200 path): the endpoint reflects real probes, not a constant.
func TestHealthzReportsOkWhenBackingResourcesAnswer(t *testing.T) {
	server := newTestServer(t)
	c := clientFor(t, server)

	resp, body := c.do(http.MethodGet, "/healthz", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("healthz status %d, want 200", resp.StatusCode)
	}

	var report struct {
		Status string `json:"status"`
		Checks map[string]struct {
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(body, &report); err != nil {
		t.Fatalf("healthz body is not the readiness JSON: %v (%s)", err, body)
	}
	if report.Status != "ok" {
		t.Errorf("healthz status %q, want ok", report.Status)
	}
	for _, name := range []string{"sqlite", "blob-dir"} {
		check, ok := report.Checks[name]
		if !ok {
			t.Errorf("healthz missing %q check", name)
			continue
		}
		if check.Status != "ok" {
			t.Errorf("%s check status %q, want ok", name, check.Status)
		}
	}
}

// TestHealthzNamesTheFailingCheck pins the 503 path: a broken resource
// turns the endpoint degraded AND says which check failed — the old
// constant "ok" could never do this.
func TestHealthzNamesTheFailingCheck(t *testing.T) {
	t.Run("closed sqlite", func(t *testing.T) {
		db, err := store.Open(":memory:")
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.DB = db })
		c := clientFor(t, server)

		resp, body := c.do(http.MethodGet, "/healthz", nil, "")
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("healthz status %d, want 503 (body %s)", resp.StatusCode, body)
		}
		if !strings.Contains(string(body), `"sqlite"`) {
			t.Errorf("503 body does not name the sqlite check: %s", body)
		}
	})

	t.Run("unwritable blob dir", func(t *testing.T) {
		blocked := filepath.Join(t.TempDir(), "file", "blocked")
		if err := os.WriteFile(filepath.Dir(blocked), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.BlobRoot = blocked })
		c := clientFor(t, server)

		resp, body := c.do(http.MethodGet, "/healthz", nil, "")
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("healthz status %d, want 503 (body %s)", resp.StatusCode, body)
		}
		if !strings.Contains(string(body), `"blob-dir"`) {
			t.Errorf("503 body does not name the blob-dir check: %s", body)
		}
	})
}
