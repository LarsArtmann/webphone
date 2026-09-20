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

// Store is the session persistence seam: mint, look up, drop.
// Attach/Require are package-level functions over any Store.
type Store interface {
	Create(extension domain.Extension, password string) (string, error)
	Get(token string) (Session, bool)
	Delete(token string)
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

// Create mints a session for the extension and returns its token.
func (s *MemStore) Create(extension domain.Extension, password string) (string, error) {
	token, err := mintToken()
	if err != nil {
		return "", err
	}

	now := time.Now()
	s.mu.Lock()
	s.gcLocked(now)
	s.sessions[token] = Session{
		Extension: extension,
		Password:  password,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}
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

// Attach stashes a live session (when the request carries one) into the
// context and always continues — pages render for anonymous visitors too.
// Handlers that need a session call From and reject themselves.
func Attach(store Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sess, ok := store.Get(TokenFromRequest(r)); ok {
			r = r.WithContext(With(r.Context(), sess))
		}
		next.ServeHTTP(w, r)
	})
}

// Require gates a handler behind a live session: anonymous requests get a
// 401.
func Require(store Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := store.Get(TokenFromRequest(r))
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
