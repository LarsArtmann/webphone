package server

import (
	"bytes"
	"net/http"
	"testing"
)

// TestTemplFormPostBucketsFenceTheLanes pins the templ form-POST
// buckets: every write lane outside /api/* rides a keyed bucket by cost
// class (compose = provider/paid surfaces, contacts = import grade,
// state = hook grade), and the exhausted bucket answers 429 with the
// .wp-error element htmx's responseHandling swaps into #wp-tab-error —
// a flooded lane says so inside the tab instead of dying silently.
func TestTemplFormPostBucketsFenceTheLanes(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	cases := []struct {
		name  string
		path  string
		body  string
		ctype string
		burst uint
	}{
		{"compose messages/send", "/messages/send", "remote=%2B4915112345678&body=hi", "application/x-www-form-urlencoded", composeBurst},
		{"compose fax/send", "/fax/send", "remote=%2B4915112345678", "application/x-www-form-urlencoded", composeBurst},
		{"compose ai-draft", "/messages/ai-draft", "thread_id=x", "application/x-www-form-urlencoded", composeBurst},
		{"contacts import", "/contacts/import", "contacts=%5B%5D", "application/x-www-form-urlencoded", contactsBurst},
		{"state flag", "/messages/does-not-exist/mute?on=1", "", "", hookBurst},
		{"state snippet save", "/snippets/save", "label=a&body=b", "application/x-www-form-urlencoded", hookBurst},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seen429 := false
			var lastBody []byte
			for range tc.burst + 3 {
				resp, body := c.do(http.MethodPost, tc.path, []byte(tc.body), tc.ctype)
				if resp.StatusCode == http.StatusTooManyRequests {
					seen429 = true
					lastBody = body
					break
				}
			}
			if !seen429 {
				t.Fatalf("flood of %s never got limited", tc.path)
			}
			if !bytes.Contains(lastBody, []byte("wp-error")) || !bytes.Contains(lastBody, []byte(templRateLimitedMsg)) {
				t.Fatalf("429 body lost the htmx-selected honest copy: %s", lastBody)
			}
		})
	}
}
