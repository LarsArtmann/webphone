package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestAssetsCarryContentETag pins the perf contract (plan T10): every
// /assets/* response carries a strong content ETag, and a repeat request
// presenting it revalidates with a bodyless 304 — a redeploy still shows
// up immediately because Cache-Control stays no-cache.
func TestAssetsCarryContentETag(t *testing.T) {
	c := newClient(t)
	for _, path := range []string{
		"/assets/app.css",
		"/assets/shell.js",
		"/assets/island/app/calls.js",
		"/assets/vendor/sip.min.js",
	} {
		t.Run(path, func(t *testing.T) {
			resp, body := c.do(http.MethodGet, path, nil, "")
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want 200", resp.StatusCode)
			}
			etag := resp.Header.Get("ETag")
			if !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) {
				t.Fatalf("ETag %q is not a strong quoted validator", etag)
			}
			if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
				t.Errorf("Cache-Control = %q, want no-cache", got)
			}

			// Presenting the ETag revalidates: 304, no body.
			req, err := http.NewRequest(http.MethodGet, c.base+path, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("If-None-Match", etag)
			resp2, err := c.http.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = resp2.Body.Close() }()
			if resp2.StatusCode != http.StatusNotModified {
				t.Fatalf("revalidation status = %d, want 304", resp2.StatusCode)
			}
			if resp2.Header.Get("ETag") != etag {
				t.Errorf("304 lost the ETag: %q", resp2.Header.Get("ETag"))
			}
			revalidated, _ := io.ReadAll(resp2.Body)
			if len(revalidated) != 0 {
				t.Errorf("304 carried a body (%d bytes)", len(revalidated))
			}
			if len(body) == 0 {
				t.Error("baseline 200 carried no body")
			}
		})
	}
}

// TestAssetsGzipWhenAccepted pins the scoped gzip (plan T10): a client
// that advertises gzip gets a compressed static subtree. The Accept-
// Encoding header is set by the test, so Go's transport does NOT
// transparently decompress — the raw bytes must gunzip back to the
// identity body.
func TestAssetsGzipWhenAccepted(t *testing.T) {
	c := newClient(t)
	const path = "/assets/shell.js"
	req, err := http.NewRequest(http.MethodGet, c.base+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if enc := resp.Header.Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", enc)
	}
	if vary := resp.Header.Get("Vary"); !strings.Contains(vary, "Accept-Encoding") {
		t.Errorf("Vary = %q, want it to name Accept-Encoding", vary)
	}
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("not a gzip stream: %v", err)
	}
	defer func() { _ = gz.Close() }()
	plain, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plain), "data-dial") {
		t.Errorf("gunzipped body lost its content: %.80s", plain)
	}
}
