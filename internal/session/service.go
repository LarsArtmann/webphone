// Package session owns the webphone login session: an opaque cookie
// token mapped to one signed-in extension. Identity comes from the PBX
// directory — the island's successful SIP REGISTER is the proof of
// credentials; the server session only carries them to the tabs and the
// phone-api proxy. The Store interface is the persistence seam: tests
// and loopback dev use the in-memory store (sessions vanish on restart,
// the island re-logs), deployments back it with SQLite so a service
// restart no longer signs everyone out (see
// docs/planning/2026-09-20_17-41_session-persistence-spike-verdict.md).
package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/pbx"
)

// CookieName is the session cookie.
const CookieName = "webphone_session"

// SignInFirst is the 401 body every unauthenticated request gets — shared
// by the session middleware and the handlers' own gates so the wording
// stays identical everywhere.
const SignInFirst = "sign in first"

// tokenBytes is the entropy of one session token (256 bit).
const tokenBytes = 32

// Session is one signed-in extension.
type Session struct {
	Extension domain.Extension
	Password  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// PBXCredentials derives the phone-api credentials this session carries —
// the same extension + directory password the SIP REGISTER proved.
func (s Session) PBXCredentials() pbx.Credentials {
	return pbx.Credentials{Extension: s.Extension.String(), Password: s.Password}
}

// Store is the session persistence seam: mint, look up, renew, drop.
// Attach/Require are package-level functions over any Store.
type Store interface {
	Create(extension domain.Extension, password string) (string, error)
	Get(token string) (Session, bool)
	// Renew applies one sliding-renewal decision (see renewDue) and
	// returns the refreshed session; ok=false when the token is absent,
	// expired, or its expiry is already where the policy wants it.
	Renew(token string, idle, maxAge time.Duration) (Session, bool)
	Delete(token string)
}

// Lifetime is the session-expiry policy. Idle is the sliding idle
// window: activity past its halfway point extends the session back to
// the full window, so a device in regular use never re-signs-in. Max is
// the ABSOLUTE lifetime measured from sign-in — the unconditional bound
// that expires even a continuously renewed (or stolen) session.
type Lifetime struct {
	Idle time.Duration
	Max  time.Duration
}

// normalized guards the invariants the renewal math relies on. A
// missing or mis-ordered Max degrades to Max = Idle (the pre-sliding
// behavior: absolute lifetime equals one idle window) instead of
// capping every session at sign-in time; a non-positive Idle disables
// renewal entirely.
func (l Lifetime) normalized() Lifetime {
	if l.Idle <= 0 {
		return Lifetime{}
	}
	if l.Max < l.Idle {
		l.Max = l.Idle
	}
	return l
}

// renewDue computes one sliding-renewal decision: extend the idle
// window back to full, never past the absolute cap, and only once per
// half-life — a request while more than half the window remains writes
// nothing (the throttle that keeps a busy tab from touching the store
// on every request). ok=false means keep the current expiry. A
// cap-bound session simply stops extending: the cap is what re-signs
// even a continuously active device in.
func renewDue(sess Session, idle, maxAge time.Duration, now time.Time) (time.Time, bool) {
	if !now.Before(sess.ExpiresAt) {
		return time.Time{}, false
	}
	extended := now.Add(idle)
	if hardCap := sess.CreatedAt.Add(maxAge); extended.After(hardCap) {
		extended = hardCap
	}
	if !extended.After(sess.ExpiresAt) {
		return time.Time{}, false // already at (or past) the cap
	}
	if sess.ExpiresAt.Sub(now) > idle/2 {
		return time.Time{}, false // still young: throttle the write
	}
	return extended, true
}

// MemStore keeps sessions in a map with TTL-based expiry. Sessions are
// stored by value: small, immutable after creation, and immune to
// nil-dereference by construction.
type MemStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
	ttl      time.Duration
}

// NewMemStore builds the in-memory session store (tests, loopback dev).
func NewMemStore(ttl time.Duration) *MemStore {
	return &MemStore{sessions: make(map[string]Session), ttl: ttl}
}

// mintToken returns a fresh 256-bit URL-safe token.
func mintToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// makeSession mints a token and the Session around it; the birth
// invariant (ExpiresAt = CreatedAt + ttl) has this single home, shared
// by both store backends.
func makeSession(extension domain.Extension, password string, ttl time.Duration) (string, Session, error) {
	token, err := mintToken()
	if err != nil {
		return "", Session{}, err
	}
	now := time.Now()
	return token, Session{
		Extension: extension,
		Password:  password,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}, nil
}

// Create mints a session for the extension and returns its token.
func (s *MemStore) Create(extension domain.Extension, password string) (string, error) {
	token, sess, err := makeSession(extension, password, s.ttl)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.gcLocked(sess.CreatedAt)
	s.sessions[token] = sess
	s.mu.Unlock()

	return token, nil
}

// Get returns a live session by token.
func (s *MemStore) Get(token string) (Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok || time.Now().After(sess.ExpiresAt) {
		return Session{}, false
	}
	return sess, true
}

// Delete drops a session (logout).
func (s *MemStore) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// Renew applies the sliding-renewal policy to one live session.
func (s *MemStore) Renew(token string, idle, maxAge time.Duration) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[token]
	if !ok {
		return Session{}, false
	}
	now := time.Now()
	if !now.Before(sess.ExpiresAt) {
		delete(s.sessions, token)
		return Session{}, false
	}
	extended, due := renewDue(sess, idle, maxAge, now)
	if !due {
		return Session{}, false
	}
	sess.ExpiresAt = extended
	s.sessions[token] = sess
	return sess, true
}

// gcLocked removes expired sessions; caller holds the write lock.
func (s *MemStore) gcLocked(now time.Time) {
	for token, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, token)
		}
	}
}

type contextKey struct{}

// With returns a context carrying the session.
func With(ctx context.Context, sess Session) context.Context {
	return context.WithValue(ctx, contextKey{}, sess)
}

// From extracts the session a middleware stashed in the context.
func From(ctx context.Context) (Session, bool) {
	sess, ok := ctx.Value(contextKey{}).(Session)
	return sess, ok
}

// TokenFromRequest reads the session cookie.
func TokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// lookup resolves (and maybe renews) the session for one request: the
// shared body of Attach and Require. A true renewal re-issues the
// cookie with the server's remaining lifetime, keeping the Max-Age
// parity the login SetCookie established — without it the browser
// cookie would die mid-session while the row lives on.
func lookup(store Store, w http.ResponseWriter, r *http.Request, lifetime Lifetime) (Session, bool) {
	token := TokenFromRequest(r)
	sess, ok := store.Get(token)
	if !ok || lifetime.Idle <= 0 {
		return sess, ok
	}
	if renewed, did := store.Renew(token, lifetime.Idle, lifetime.Max); did {
		SetCookie(w, r, token, time.Until(renewed.ExpiresAt))
		return renewed, true
	}
	return sess, ok
}

// Attach stashes a live session (when the request carries one) into the
// context and always continues — pages render for anonymous visitors too.
// Handlers that need a session call From and reject themselves. A positive
// Lifetime.Idle slides the session forward on activity past the window's
// halfway point.
func Attach(store Store, next http.Handler, lifetime Lifetime) http.Handler {
	lifetime = lifetime.normalized()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sess, ok := lookup(store, w, r, lifetime); ok {
			r = r.WithContext(With(r.Context(), sess))
		}
		next.ServeHTTP(w, r)
	})
}

// Require gates a handler behind a live session: anonymous requests get a
// 401. Slides the session forward like Attach (an active SSE feed or
// phone-api call is activity too).
func Require(store Store, next http.Handler, lifetime Lifetime) http.Handler {
	lifetime = lifetime.normalized()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := lookup(store, w, r, lifetime)
		if !ok {
			http.Error(w, SignInFirst, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(With(r.Context(), sess)))
	})
}

// SetCookie writes the session cookie.
func SetCookie(w http.ResponseWriter, r *http.Request, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
}

// ClearCookie removes the session cookie.
func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: CookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
	slog.Debug("session cookie cleared")
}
