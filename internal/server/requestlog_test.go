package server

import (
	"encoding/json/v2"
	"net/http"
	"strings"
	"testing"
)

func doAndDrain(c *client, method, path string, body []byte, contentType string) *http.Response {
	resp, _ := c.do(method, path, body, contentType)
	_ = resp.Body.Close()
	return resp
}

// TestRequestLogCoversEverySurface pins the request-logging contract: every
// request, whatever its outcome, leaves exactly one structured log line
// behind — the server-side twin of the browser event log the runbook greps.
func TestRequestLogCoversEverySurface(t *testing.T) {
	log := captureDefaultLogger(t)
	server := newTestServer(t)
	c := clientFor(t, server)

	doAndDrain(c, http.MethodGet, "/", nil, "")
	doAndDrain(c, http.MethodGet, "/no-such-page", nil, "")
	doAndDrain(c, http.MethodGet, "/events", nil, "")

	payload, err := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	if err != nil {
		t.Fatal(err)
	}
	for range loginBurst + 3 {
		// Successful logins rotate the CSRF token; re-arm like a real
		// client so these POSTs exercise the limiter, not the CSRF gate.
		if resp := doAndDrain(c, http.MethodPost, "/api/session", payload, "application/json"); resp.StatusCode == http.StatusCreated {
			c.adoptCsrfToken()
		}
	}

	entry := log.String()
	statuses := map[string]bool{}
	for line := range strings.SplitSeq(entry, "\n") {
		if !strings.Contains(line, "http request") {
			continue // startup noise (store ready, CSRF dev warning) is out of scope
		}
		for _, field := range []string{"method=", "path=", "status=", "duration="} {
			if !strings.Contains(line, field) {
				t.Errorf("request log line missing %q:\n%s", field, line)
			}
		}
		for _, code := range []string{"status=200", "status=404", "status=401", "status=429"} {
			if strings.Contains(line, code) {
				statuses[code] = true
			}
		}
	}
	for _, code := range []string{"status=200", "status=404", "status=401", "status=429"} {
		if !statuses[code] {
			t.Errorf("no request log line carried %s; captured:\n%s", code, entry)
		}
	}
}

// TestRequestLogNeverCarriesSecrets pins the no-credentials contract: bodies
// and passwords stay out of the log — the middleware records method, path,
// status and duration, nothing else.
func TestRequestLogNeverCarriesSecrets(t *testing.T) {
	log := captureDefaultLogger(t)
	server := newTestServer(t)
	c := clientFor(t, server)

	const password = "s3cret-hunter2-password"
	payload, err := json.Marshal(map[string]string{"extension": "1001", "password": password})
	if err != nil {
		t.Fatal(err)
	}
	resp := doAndDrain(c, http.MethodPost, "/api/session", payload, "application/json")
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("login probe returned %d, want a 2xx session creation", resp.StatusCode)
	}

	entry := log.String()
	if strings.Contains(entry, password) {
		t.Errorf("password leaked into request logs:\n%s", entry)
	}
	if strings.Contains(entry, "s3cret") {
		t.Errorf("credential-looking content leaked into request logs:\n%s", entry)
	}
}
