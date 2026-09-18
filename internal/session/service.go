// Package session owns the webphone login session: an in-memory store of
// extension credentials keyed by an opaque cookie token. Identity comes
// from the PBX directory — the island's successful SIP REGISTER is the
// proof of credentials; the server session only carries them to the tabs
// and the phone-api proxy. Sessions vanish on restart by design (the
// island re-logs); nothing about the user is persisted server-side except
// her own messages, faxes and contacts.
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
)

// CookieName is the session cookie.
const CookieName = "webphone_session"

// tokenBytes is the entropy of one session token (256 bit).
const tokenBytes = 32

// Session is one signed-in extension.
type Session struct {
	Extension domain.Extension
	Password  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Store keeps sessions in memory with TTL-based expiry.
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
}

// NewStore builds the session store.
func NewStore(ttl time.Duration) *Store {
	return &Store{sessions: make(map[string]*Session), ttl: ttl}
}

// Create mints a session for the extension and returns its token.
func (s *Store) Create(extension domain.Extension, password string) (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buf)

	now := time.Now()
	s.mu.Lock()
	s.gcLocked(now)
	s.sessions[token] = &Session{
		Extension: extension,
		Password:  password,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}
	s.mu.Unlock()

	return token, nil
}

// Get returns a live session by token.
func (s *Store) Get(token string) (Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok || time.Now().After(sess.ExpiresAt) {
		return Session{}, false
	}
	return *sess, true
}

// Delete drops a session (logout).
func (s *Store) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// gcLocked removes expired sessions; caller holds the write lock.
func (s *Store) gcLocked(now time.Time) {
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
func (store *Store) Attach(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sess, ok := store.Get(TokenFromRequest(r)); ok {
			r = r.WithContext(With(r.Context(), sess))
		}
		next.ServeHTTP(w, r)
	})
}

// Require gates a handler behind a live session: anonymous requests get a
// 401.
func (store *Store) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := store.Get(TokenFromRequest(r))
		if !ok {
			http.Error(w, "sign in first", http.StatusUnauthorized)
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
