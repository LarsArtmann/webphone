package session

import (
	"net/http"
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

// TestRenewDuePolicy pins the sliding-renewal decision table: the idle
// window extends only once past its halfway point (the write throttle),
// the absolute cap is never crossed, and a cap-bound session stops
// extending entirely — the cap is what re-signs even a continuously
// active device in.
func TestRenewDuePolicy(t *testing.T) {
	const idle = time.Hour
	const maxAge = 30 * 24 * time.Hour
	base := time.Unix(1700000000, 0)

	cases := []struct {
		name    string
		created time.Time
		expires time.Time
		now     time.Time
		want    time.Time
		wantOK  bool
	}{
		{
			name:    "young session: more than half the window remains, no write",
			created: base,
			expires: base.Add(idle),
			now:     base.Add(10 * time.Minute),
			wantOK:  false,
		},
		{
			name:    "past half-life: extends back to the full idle window",
			created: base,
			expires: base.Add(idle),
			now:     base.Add(40 * time.Minute),
			want:    base.Add(40*time.Minute + idle),
			wantOK:  true,
		},
		{
			name:    "cap binds: the extension is clamped to created+max",
			created: base,
			expires: base.Add(29*24*time.Hour + 45*time.Minute),
			now:     base.Add(29*24*time.Hour + 30*time.Minute),
			want:    base.Add(maxAge),
			wantOK:  true,
		},
		{
			name:    "at the cap: nothing left to extend",
			created: base,
			expires: base.Add(maxAge),
			now:     base.Add(maxAge - 15*time.Minute),
			wantOK:  false,
		},
		{
			name:    "expired: dead rows never renew",
			created: base,
			expires: base.Add(idle),
			now:     base.Add(idle + time.Minute),
			wantOK:  false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sess := Session{CreatedAt: tc.created, ExpiresAt: tc.expires}
			got, ok := renewDue(sess, idle, maxAge, tc.now)
			if ok != tc.wantOK {
				t.Fatalf("renewDue ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && !got.Equal(tc.want) {
				t.Errorf("renewDue = %s, want %s", got, tc.want)
			}
		})
	}
}

// TestMemStoreRenew pins the in-memory sliding renewal end to end:
// throttle, extension, idempotent no-op while young, unknown and
// expired tokens.
func TestMemStoreRenew(t *testing.T) {
	ext := testExtension(t)
	store := NewMemStore(time.Hour)

	token, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Renew("never-existed", time.Hour, 30*24*time.Hour); ok {
		t.Fatal("unknown token renewed")
	}
	backdateRemaining(t, store, token, 20*time.Minute)

	if renewed, ok := store.Renew(token, time.Hour, 30*24*time.Hour); !ok {
		t.Fatal("past-half-life session did not renew")
	} else if until := time.Until(renewed.ExpiresAt); until < 55*time.Minute || until > time.Hour {
		t.Errorf("renewed expiry %s from now, want ~the full idle window", until)
	}

	if _, ok := store.Renew(token, time.Hour, 30*24*time.Hour); ok {
		t.Error("young session renewed again — the write throttle is broken")
	}

	// An expired row reads dead and is removed, not renewed.
	backdateRemaining(t, store, token, -time.Minute)
	if _, ok := store.Renew(token, time.Hour, 30*24*time.Hour); ok {
		t.Fatal("expired session renewed")
	}
	store.mu.Lock()
	_, stillThere := store.sessions[token]
	store.mu.Unlock()
	if stillThere {
		t.Error("expired row survived Renew")
	}
}

// backdateRemaining rewrites one row's ExpiresAt to `remaining` from
// now (white-box: the tests live beside the store).
func backdateRemaining(t *testing.T, store *MemStore, token string, remaining time.Duration) {
	t.Helper()
	store.mu.Lock()
	defer store.mu.Unlock()
	sess, ok := store.sessions[token]
	if !ok {
		t.Fatalf("no session %q", token)
	}
	sess.ExpiresAt = time.Now().Add(remaining)
	store.sessions[token] = sess
}

// TestAttachSlidesAndReIssuesCookie pins the middleware half of the
// sliding renewal: activity past the half-life extends the row AND
// re-issues the cookie with the server's remaining lifetime (the
// Max-Age parity the login established) — while a young session's
// request writes neither.
func TestAttachSlidesAndReIssuesCookie(t *testing.T) {
	ext := testExtension(t)
	store := NewMemStore(time.Hour)
	lifetime := Lifetime{Idle: time.Hour, Max: 30 * 24 * time.Hour}

	token, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}
	backdateRemaining(t, store, token, 20*time.Minute)

	attach := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: CookieName, Value: token})
		Attach(store, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := From(r.Context()); !ok {
				t.Error("live session not attached to the context")
			}
		}), lifetime).ServeHTTP(rec, req)
		return rec
	}

	rec := attach()
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("renewal did not re-issue the cookie (got %d)", len(cookies))
	}
	if maxAge := cookies[0].MaxAge; maxAge < 3500 || maxAge > int(time.Hour.Seconds()) {
		t.Errorf("renewed cookie Max-Age %ds, want ~the full idle window", maxAge)
	}
	if sess, ok := store.Get(token); !ok {
		t.Fatal("session lost")
	} else if until := time.Until(sess.ExpiresAt); until < 55*time.Minute {
		t.Errorf("row expiry %s from now — the extension was not persisted", until)
	}

	// Young again: the throttle keeps the second request write-free.
	if rec := attach(); len(rec.Result().Cookies()) != 0 {
		t.Error("young session's request re-issued the cookie — no throttle")
	}
}

// TestAttachWithoutLifetimeDisablesRenewal pins the zero-value guard: a
// Lifetime without an idle window keeps the pre-sliding behavior — no
// renewal, no cookie churn, expired sessions stay dead.
func TestAttachWithoutLifetimeDisablesRenewal(t *testing.T) {
	ext := testExtension(t)
	store := NewMemStore(time.Hour)
	token, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}
	backdateRemaining(t, store, token, 20*time.Minute)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: token})
	Attach(store, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), Lifetime{}).ServeHTTP(rec, req)

	if cookies := rec.Result().Cookies(); len(cookies) != 0 {
		t.Errorf("zero Lifetime re-issued %d cookie(s)", len(cookies))
	}
	if sess, ok := store.Get(token); !ok {
		t.Fatal("session lost")
	} else if until := time.Until(sess.ExpiresAt); until > 21*time.Minute {
		t.Errorf("row expiry moved to %s from now without a Lifetime — renewal ran anyway", until)
	}
}

// TestRequireSlides pins that gated surfaces renew too: an active SSE
// feed or phone-api call is activity.
func TestRequireSlides(t *testing.T) {
	ext := testExtension(t)
	store := NewMemStore(time.Hour)
	token, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}
	backdateRemaining(t, store, token, 20*time.Minute)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/events", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: token})
	Require(store, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), Lifetime{Idle: time.Hour, Max: 30 * 24 * time.Hour}).ServeHTTP(rec, req)

	if sess, ok := store.Get(token); !ok || time.Until(sess.ExpiresAt) < 55*time.Minute {
		t.Errorf("Require did not slide the session (ok=%v)", ok)
	}
}
