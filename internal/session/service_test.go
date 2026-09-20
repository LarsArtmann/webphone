package session

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestSessionTTLExpiryAndSweep pins the TTL/sweeper interaction: a session
// is live until its ExpiresAt and dead after; expiry alone hides the
// session lazily (Get), while the next Create's sweep removes the dead
// entry from the map entirely — expired sessions cannot accumulate.
func TestSessionTTLExpiryAndSweep(t *testing.T) {
	ext, err := domain.ParseExtension("1001")
	if err != nil {
		t.Fatal(err)
	}
	store := NewMemStore(50 * time.Millisecond)

	token, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Get(token); !ok {
		t.Fatal("fresh session not live")
	}

	time.Sleep(60 * time.Millisecond)

	if _, ok := store.Get(token); ok {
		t.Fatal("expired session still live — TTL not honored")
	}

	// The map still holds the dead entry (expiry is lazy until here);
	// the next Create sweeps it.
	if _, err := store.Create(ext, "pw"); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	_, stillThere := store.sessions[token]
	store.mu.Unlock()
	if stillThere {
		t.Fatal("expired session survived the Create-time sweep")
	}

	// A young session survives a sweep that ran in the same window.
	token2, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Get(token2); !ok {
		t.Fatal("young session not live after another session was swept")
	}
}

// TestCookieMaxAgeMirrorsTTL pins the cookie/store TTL agreement: the
// browser cookie must expire with the server-side session, or the
// island would send a dead token until the browser clears it.
func TestCookieMaxAgeMirrorsTTL(t *testing.T) {
	ttl := 7 * time.Minute
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/session", nil)
	SetCookie(rec, req, "token-abc", ttl)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	got := cookies[0]
	if got.MaxAge != int(ttl.Seconds()) {
		t.Errorf("cookie Max-Age %ds, want %ds (the SessionTTL)", got.MaxAge, int(ttl.Seconds()))
	}
	if got.Value != "token-abc" {
		t.Errorf("unexpected cookie value %q", got.Value)
	}
	if !got.HttpOnly {
		t.Error("session cookie must stay HttpOnly")
	}
}
