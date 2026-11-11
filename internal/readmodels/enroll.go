package readmodels

import (
	"context"
	"time"
)

// EnrollTokenView is the enrollment-token projection row both backends
// fold into (the journal-backed system.db projection reads/writes this
// exact shape; the hash is the key; the raw secret never exists here).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type EnrollTokenView struct {
	TokenHash string `json:"token_hash"`
	Email     string `json:"email"`
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
}

// EnrollGate is the enrollment-token seam as the userauth door sees
// it, answered by the journal-backed backend (internal/cqrs). Token
// classification (invalid vs expired vs unknown) stays in the userauth
// service — the seam only ever reports found/not-found plus the
// expiry.
type EnrollGate interface {
	Mint(ctx context.Context, tokenHash, email, userID string, expiresAt time.Time) error
	Lookup(ctx context.Context, tokenHash string) (email, userID string, expiresAt time.Time, found bool, err error)
	Burn(ctx context.Context, tokenHash string) error
}
