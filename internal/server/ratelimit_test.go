package server

import (
	"io"
	"net/http"
	"strconv"
	"testing"
)

func TestLoginRateLimitPerClient(t *testing.T) {
	server := newTestServer(t)
	c := clientFor(t, server)

	limited := false
	for range loginBurst + 3 {
		resp := c.loginRaw("1001", "pw")
		if resp.StatusCode == http.StatusTooManyRequests {
			limited = true
			retryAfter(t, resp)
			break
		}
		// A successful login rotates the CSRF token; a scripted flooder
		// re-arms via GET /api/csrf between attempts, so mimic that:
		// otherwise the next POST dies at the CSRF layer and never
		// reaches the limiter this test exercises.
		if resp.StatusCode == http.StatusCreated {
			c.adoptCsrfToken()
		}
	}
	if !limited {
		t.Errorf("login never hit the rate limit after %d attempts", loginBurst+3)
	}
}

func TestHookRateLimitPerClient(t *testing.T) {
	server := newTestServer(t)

	limited := false
	for range hookBurst + 5 {
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
			retryAfter(t, resp)
			break
		}
	}
	if !limited {
		t.Errorf("hooks never hit the rate limit after %d requests", hookBurst+5)
	}
}

// TestCSRFRefreshRateLimitPerClient pins the token endpoint's flood
// bound: GET /api/csrf is anonymous, so without a budget a client could
// churn masked tokens unbounded. Real traffic is one fetch per login
// rotation; the shared hook budget (60/min burst 60) is far above that
// and rejected requests carry the computed Retry-After like every
// other limited surface.
func TestCSRFRefreshRateLimitPerClient(t *testing.T) {
	server := newTestServer(t)

	limited := false
	for range hookBurst + 5 {
		resp, err := server.Client().Get(server.URL + "/api/csrf")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			limited = true
			retryAfter(t, resp)
			break
		}
	}
	if !limited {
		t.Errorf("csrf refresh never hit the rate limit after %d requests", hookBurst+5)
	}
}

// retryAfter pins the contract the hand-rolled limiter never honored: the
// header is computed from the window (1 minute here), not a hardcoded guess.
func retryAfter(t *testing.T, resp *http.Response) {
	t.Helper()
	raw := resp.Header.Get("Retry-After")
	seconds, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("Retry-After %q is not seconds: %v", raw, err)
	}
	if seconds < 1 || seconds > 60 {
		t.Errorf("Retry-After %ds outside the 1-minute window", seconds)
	}
}

// TestEventsRateLimitBounded pins the reconnect-churn bound: GET /events
// shares the flood budget (60/min burst 60) so a reconnect loop cannot
// open unlimited streams, and rejected connects carry a computed
// Retry-After like every other limited surface.
func TestEventsRateLimitBounded(t *testing.T) {
	server := newTestServer(t)

	limited := false
	for range hookBurst + 5 {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/events", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			limited = true
			retryAfter(t, resp)
			break
		}
	}
	if !limited {
		t.Errorf("events never hit the rate limit after %d connects", hookBurst+5)
	}
}

// TestHookLimiterWrapsSecretGate pins the middleware order decided when
// the hooks landed: the limiter sits OUTSIDE the secret gate, so (a) a
// flood is 429'd even when every request also fails the gate, and (b) an
// unconfigured secret answers 503 only while the peer still has budget.
// Reorder the wrap in server.New and this test fails.
func TestHookLimiterWrapsSecretGate(t *testing.T) {
	t.Run("429 precedes 401 under flood", func(t *testing.T) {
		server := newTestServer(t)

		seen429 := false
		for range hookBurst + 5 {
			req, err := http.NewRequest(http.MethodPost, server.URL+"/hooks/fax/status", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", "Bearer wrong-secret")
			resp, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusTooManyRequests {
				seen429 = true
			}
		}
		if !seen429 {
			t.Fatal("flood of bad-secret requests never got limited")
		}

		req, err := http.NewRequest(http.MethodPost, server.URL+"/hooks/fax/status", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer test-secret")
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("valid-secret request after flood got %d, want 429 (limiter must wrap the gate)", resp.StatusCode)
		}
	})

	t.Run("503 unconfigured secret only under budget", func(t *testing.T) {
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) {
			d.Config.Gateway.WebhookSecret = ""
		})

		req, err := http.NewRequest(http.MethodPost, server.URL+"/hooks/fax/status", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("unconfigured secret under budget got %d, want 503", resp.StatusCode)
		}
	})
}
