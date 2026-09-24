package server

import (
	"net/http"
	"strings"
	"testing"
)

// requireSessionMultipart caps the raw request body at uploadBodyLimit:
// ParseMultipartForm's argument is only a memory threshold, so without
// MaxBytesReader an authenticated client could stream an unbounded body
// to server disk (audit finding #1). The envelope (fields + boundaries)
// rides inside the 1 MB headroom above uploadLimit.
func TestUploadBodyIsBounded(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	t.Run("oversize body answers 400 too large", func(t *testing.T) {
		form, contentType := multipartBody(t,
			map[string]string{"to": "+15551234567", "body": "hi"},
			map[string]struct {
				Name    string
				Content []byte
			}{"attachment": {Name: "huge.bin", Content: make([]byte, uploadBodyLimit+1)}},
		)
		resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("oversize upload: %d (want 400)", resp.StatusCode)
		}
		if !strings.Contains(string(body), "too large") {
			t.Fatalf("oversize 400 lost its reason: %.200s", body)
		}
	})

	t.Run("body at the boundary still parses", func(t *testing.T) {
		// A file sized so the TOTAL envelope (fields + boundaries + file)
		// stays just under uploadBodyLimit: form parsing must succeed and
		// reach the per-tab validation (422 for the invalid "to"), proving
		// the cap does not clip legitimate uploads at the threshold.
		form, contentType := multipartBody(t,
			map[string]string{"to": "&&&", "body": "hi"},
			map[string]struct {
				Name    string
				Content []byte
			}{"attachment": {Name: "fits.bin", Content: make([]byte, uploadBodyLimit-4096)}},
		)
		resp, body := c.do(http.MethodPost, "/messages/send", form, contentType)
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("boundary upload: %d %s (want 422 = parsed past the cap)", resp.StatusCode, body)
		}
		if strings.Contains(string(body), "too large") {
			t.Fatalf("boundary upload must not trip the cap: %.200s", body)
		}
	})
}
