package server

import (
	"bytes"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// The golden posture is OWNER-RATIFIED via the portfolio plan's execution
// order (2026-11-08): the contract is generated, so a handler DTO change
// that reshapes the wire surface must be a REVIEWED contract event — the
// fixture fails until the spec is consciously regenerated.
var updateOpenapiGolden = flag.Bool("update", false,
	"rewrite internal/server/testdata/openapi.golden.json from the live render")

// TestOpenapiGoldenFixture pins GET /openapi.json byte-for-byte against
// internal/server/testdata/openapi.golden.json. The render is
// deterministic (json.Deterministic marshaling behind sync.OnceValues),
// so the bytes are stable across runs and a spec drift can only come
// from a real DTO/contract change. Regenerate consciously:
//
//	go test ./internal/server -run TestOpenapiGoldenFixture -update
//
// then read the diff — it IS the wire-contract change.
func TestOpenapiGoldenFixture(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("openapi status = %d", resp.StatusCode)
	}
	body := readAll(t, resp)

	goldenPath := filepath.Join("testdata", "openapi.golden.json")
	if *updateOpenapiGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, body, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("golden rewritten (%d bytes) — read the diff before committing", len(body))
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden fixture missing (run with -update to seed it): %v", err)
	}
	if !bytes.Equal(body, want) {
		t.Fatalf(`the served OpenAPI contract drifted from the committed golden (%d vs %d bytes); regenerate with: go test ./internal/server -run TestOpenapiGoldenFixture -update — and READ the diff, it is the wire-contract change`,
			len(body), len(want))
	}
}
