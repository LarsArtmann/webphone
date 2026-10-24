package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestAPIContractIsValid fails the build when the declared catalog is
// internally inconsistent (a dangling security reference, a malformed
// operation, …). The served /openapi.json is generated from this
// catalog, so a violation here is a broken contract before any request.
func TestAPIContractIsValid(t *testing.T) {
	if violations := apiCatalog().Validate(); len(violations) > 0 {
		t.Fatalf("api contract violations: %v", violations)
	}
}

// TestAPIContractPathsAreMounted walks the GENERATED contract and probes
// the live mux: every declared (path, method) must reach a real handler.
// The discriminator is the 404 SHAPE — a declared path the mux never
// mounted falls through to the styled catch-all page (404 + text/html),
// while every real handler answers plain-text errors (http.Error /
// http.NotFound) or a status. This is the hole Catalog.Validate cannot
// see: it knows the declarations, not the router.
func TestAPIContractPathsAreMounted(t *testing.T) {
	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	body, err := openAPIContract()
	if err != nil {
		t.Fatalf("openapi contract render: %v", err)
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("generated openapi doc is not JSON: %v", err)
	}
	if len(doc.Paths) == 0 {
		t.Fatal("generated contract declares no paths")
	}

	server := newTestServer(t)
	client := server.Client()
	for path, methods := range doc.Paths {
		for method := range methods {
			req, err := http.NewRequest(method, server.URL+path, nil)
			if err != nil {
				t.Fatalf("probe %s %s: %v", method, path, err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("probe %s %s: %v", method, path, err)
			}
			readAll(t, resp)
			ct := resp.Header.Get("Content-Type")
			if resp.StatusCode == http.StatusNotFound && strings.HasPrefix(ct, "text/html") {
				t.Errorf("%s %s is declared in the contract but fell through to the styled catch-all 404 — the route is not mounted", method, path)
			}
		}
	}
}
