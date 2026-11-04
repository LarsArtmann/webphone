package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/config"
)

// TestAppStartRunsChatReceiverLoops pins the wiring the chat-network
// trains documented but the composition never performed: the
// config-gated receiver loops START with the app (App.Start runs what
// wireChatNetworks returns). A fake homeserver answers the boot whoami
// and the first sync; the assertion is the sync actually arriving —
// before the fix the loops existed only as uncalled closures.
func TestAppStartRunsChatReceiverLoops(t *testing.T) {
	var syncs atomic.Int64

	mux := http.NewServeMux()
	mux.HandleFunc("GET /_matrix/client/v3/account/whoami", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"user_id": "@webphone:example.org"})
	})
	// The initial sync (empty since) answers immediately with a batch
	// token; every later long-poll blocks until the request context dies,
	// so the loop goes quiet after the round this test waits for.
	mux.HandleFunc("GET /_matrix/client/v3/sync", func(w http.ResponseWriter, r *http.Request) {
		syncs.Add(1)
		if syncs.Load() > 1 {
			select {
			case <-r.Context().Done():
			case <-time.After(30 * time.Second):
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"next_batch":"s1"}`))
	})
	homeserver := httptest.NewServer(mux)
	t.Cleanup(homeserver.Close)

	cfg := testConfig(t, false)
	cfg.Matrix = config.Matrix{
		Homeserver:  homeserver.URL,
		UserID:      "@webphone:example.org",
		AccessToken: "test-token",
		Owner:       "100",
	}

	application, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
		if err := application.Shutdown(); err != nil {
			t.Errorf("Shutdown: %v", err)
		}
	})
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if syncs.Load() > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("receiver never synced: the chat loop did not start with the app")
}
