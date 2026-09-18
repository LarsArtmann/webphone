package server

import (
	"bytes"
	"encoding/json/v2"
	"net/http"
	"strings"
	"testing"
)

func TestSessionGatesAndFlows(t *testing.T) {
	c := newClient(t)

	// Anonymous partials are gated.
	resp, _ := c.do(http.MethodGet, "/partials/messages", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anonymous partial: %d", resp.StatusCode)
	}

	payload, _ := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	resp, body := c.do(http.MethodPost, "/api/session", payload, "application/json")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("session create: %d %s", resp.StatusCode, body)
	}

	// Wrong CSRF token is rejected.
	req, _ := http.NewRequest(http.MethodPost, c.base+"/messages/send", bytes.NewReader(nil))
	req.Header.Set("X-CSRF-Token", "wrong")
	if badResp, _ := c.http.Do(req); badResp.StatusCode != http.StatusForbidden {
		t.Fatalf("bad CSRF token: %d (want 403)", badResp.StatusCode)
	}

	// Signed-in partial renders.
	resp, body = c.do(http.MethodGet, "/partials/messages", nil, "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "No conversations yet") {
		t.Fatalf("messages partial: %d", resp.StatusCode)
	}
}
