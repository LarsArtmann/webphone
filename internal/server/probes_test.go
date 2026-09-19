package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/store"
)

// TestLivezIsAlwaysOK pins the liveness contract: /livez answers 200 while
// the process serves, no matter what the backing resources are doing (it
// performs no checks — that is /healthz's and /startupz's job). A prober can
// distinguish "wedged process" from "degraded dependencies" only if liveness
// stays fetch-free.
func TestLivezIsAlwaysOK(t *testing.T) {
	server := newTestServer(t)
	c := clientFor(t, server)

	resp, body := c.do(http.MethodGet, "/livez", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("livez status %d, want 200 (%s)", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"pass"`) {
		t.Errorf("livez body does not report pass: %s", body)
	}
}

// TestStartupzIs503UntilBackingResourcesPass pins the startup latch's live
// evaluation: with a backing resource broken, /startupz is 503 (the instance
// must not be declared "started") while /livez on the same server stays 200 —
// the pair is the whole point of the split.
func TestStartupzIs503UntilBackingResourcesPass(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.DB = db })
	c := clientFor(t, server)

	resp, body := c.do(http.MethodGet, "/startupz", nil, "")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("startupz status %d, want 503 (body %s)", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"sqlite"`) {
		t.Errorf("503 body does not name the sqlite check: %s", body)
	}

	livez, _ := c.do(http.MethodGet, "/livez", nil, "")
	if livez.StatusCode != http.StatusOK {
		t.Errorf("livez status %d with broken deps, want 200 (liveness is fetch-free)", livez.StatusCode)
	}
}

// TestStartupzPassesWithHealthyBackingResources pins the 200 path: a healthy
// instance passes the startup checks on first evaluation.
func TestStartupzPassesWithHealthyBackingResources(t *testing.T) {
	server := newTestServer(t)
	c := clientFor(t, server)

	resp, body := c.do(http.MethodGet, "/startupz", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("startupz status %d, want 200 (%s)", resp.StatusCode, body)
	}
}

// TestUnwritableBlobDirKeepsLivezOK complements the 503 split: a wedged
// blob root must degrade startupz but never liveness.
func TestUnwritableBlobDirKeepsLivezOK(t *testing.T) {
	blocked := filepath.Join(t.TempDir(), "file", "blocked")
	if err := os.WriteFile(filepath.Dir(blocked), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.BlobRoot = blocked })
	c := clientFor(t, server)

	resp, _ := c.do(http.MethodGet, "/startupz", nil, "")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("startupz status %d, want 503", resp.StatusCode)
	}

	livez, _ := c.do(http.MethodGet, "/livez", nil, "")
	if livez.StatusCode != http.StatusOK {
		t.Errorf("livez status %d, want 200", livez.StatusCode)
	}
}
