package server

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
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

// TestOpenAPISeamSurfacesMatchHandlers extends the middleware_test pins
// to the config-gated surfaces: every one documents the honest 404 the
// disabled seam answers, and the limiter-carrying POSTs document their
// Retry-After contract.
func TestOpenAPISeamSurfacesMatchHandlers(t *testing.T) {
	var doc struct {
		Paths map[string]map[string]struct {
			Responses map[string]struct {
				Description string `json:"description"`
			} `json:"responses"`
		} `json:"paths"`
	}
	body, err := openAPIContract()
	if err != nil {
		t.Fatalf("openapi contract render: %v", err)
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("generated openapi doc is not JSON: %v", err)
	}

	type want struct {
		path, method   string
		codes          []string
		retryAfterCode string
	}
	for _, w := range []want{
		{"/api/transcribe", "post", []string{"200", "400", "401", "403", "404", "429", "502", "503"}, "429"},
		{"/api/transcripts", "post", []string{"204", "400", "401", "403", "404", "413", "429", "500", "503"}, "429"},
		{"/api/transcripts", "delete", []string{"204", "400", "401", "403", "404", "429", "500", "503"}, "429"},
		{"/api/push-devices", "get", []string{"200", "401", "404", "500"}, ""},
		{"/api/push-devices", "post", []string{"204", "400", "401", "403", "404", "422", "429", "500"}, "429"},
		{"/api/push-devices", "delete", []string{"204", "401", "403", "404", "500"}, ""},
		{"/api/calls/{id}/note", "get", []string{"200", "400", "401", "403", "429", "500", "503"}, "429"},
		{"/api/calls/{id}/note", "put", []string{"204", "400", "401", "403", "409", "413", "429", "500", "503"}, "429"},
		{"/api/calls/{id}/disposition", "post", []string{"204", "400", "401", "403", "409", "429", "500", "502", "503"}, "429"},
	} {
		op, ok := doc.Paths[w.path][w.method]
		if !ok {
			t.Fatalf("%s %s missing from the contract", w.method, w.path)
		}
		for _, code := range w.codes {
			if _, ok := op.Responses[code]; !ok {
				t.Errorf("%s %s spec missing documented response %s", w.method, w.path, code)
			}
		}
		// The disabled-seam 404 wording is pinned ONLY on surfaces that
		// document one (the notes surface is always mounted — its
		// conditional-less rows carry no 404 at all).
		if _, has404 := op.Responses["404"]; has404 {
			if !strings.Contains(op.Responses["404"].Description, "configured") {
				t.Errorf("%s %s must document the disabled-seam 404 contract", w.method, w.path)
			}
		}
		if w.retryAfterCode != "" && !strings.Contains(op.Responses[w.retryAfterCode].Description, "Retry-After") {
			t.Errorf("%s %s %s must document the Retry-After header (limiter contract)", w.method, w.path, w.retryAfterCode)
		}
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
		Paths map[string]map[string]jsontext.Value `json:"paths"`
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
			// OpenAPI method keys are lowercase; the CSRF gate treats
			// non-canonical methods as unsafe and would mask the
			// catch-all 404, so probe with the canonical verb.
			req, err := http.NewRequest(strings.ToUpper(method), server.URL+path, nil)
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
