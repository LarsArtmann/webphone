package server

import (
	"encoding/json/v2"
	"net/http"
	"testing"
)

func TestLoginRateLimitPerClient(t *testing.T) {
	server := newTestServer(t)
	c := clientFor(t, server)
	payload, err := json.Marshal(map[string]string{"extension": "1001", "password": "pw"})
	if err != nil {
		t.Fatal(err)
	}

	limited := false
	for i := 0; i < loginBurst+3; i++ {
		resp, _ := c.do(http.MethodPost, "/api/session", payload, "application/json")
		if resp.StatusCode == http.StatusTooManyRequests {
			limited = true
			break
		}
	}
	if !limited {
		t.Errorf("login never hit the rate limit after %d attempts", loginBurst+3)
	}
}

func TestHookRateLimitPerClient(t *testing.T) {
	server := newTestServer(t)

	limited := false
	for i := 0; i < hookBurst+5; i++ {
		req, err := http.NewRequest(http.MethodPost, server.URL+"/hooks/message", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer test-secret")
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			limited = true
			break
		}
	}
	if !limited {
		t.Errorf("hooks never hit the rate limit after %d requests", hookBurst+5)
	}
}
