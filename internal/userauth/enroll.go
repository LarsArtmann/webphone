package userauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/larsartmann/cqrs-htmx/usermgmt/v4"
	"github.com/larsartmann/go-error-family"
)

// EnrollTokenTTL bounds how long a CLI-minted enrollment link lives.
// Fifteen minutes: enough to walk from the shell to the browser, short
// enough that a leaked URL (terminal scrollback) is a narrow window.
const EnrollTokenTTL = 15 * time.Minute

// EnrollToken is what a verified enrollment token resolves to.
type EnrollToken struct {
	Email  string
	UserID usermgmt.UserID
}

// migrateEnrollTokens owns the ONE webphone-side table in the usermgmt
// database. The wp_ prefix keeps it out of usermgmt's namespace; the
// schema is additive and idempotent (no versioned migration needed — it
// never changes shape after introduction; a future change adds a new
// table, never an ALTER).
func (s *Service) migrateEnrollTokens(ctx context.Context) error {
	const create = `CREATE TABLE IF NOT EXISTS wp_enroll_tokens (
		token_hash TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		user_id TEXT NOT NULL,
		expires_at INTEGER NOT NULL
	)`
	if _, err := s.db.ExecContext(ctx, create); err != nil {
		return errorfamily.WrapInfrastructure(err, "userauth.enroll.migrate", "create wp_enroll_tokens table")
	}
	return nil
}

// MintEnrollToken creates a one-time enrollment token for an existing
// usermgmt account and returns the URL-safe secret exactly once. Only its
// SHA-256 is stored; the row is burned on first use and expired rows are
// swept opportunistically on every mint.
func (s *Service) MintEnrollToken(ctx context.Context, email string, userID usermgmt.UserID) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", errorfamily.WrapInfrastructure(err, "userauth.enroll.entropy", "generate enrollment token entropy") //nolint:erraudit // crypto/rand takes no ctx; the wrap owns the code
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	now := time.Now()
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM wp_enroll_tokens WHERE expires_at < ?`, now.Unix()); err != nil {
		return "", errorfamily.WrapInfrastructure(err, "userauth.enroll.sweep", "sweep expired enrollment tokens") //nolint:erraudit // ctx is bound to the exec; the wrap adds the code the journal greps
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO wp_enroll_tokens (token_hash, email, user_id, expires_at) VALUES (?, ?, ?, ?)`,
		hex.EncodeToString(sum[:]), email, userID.Get().String(), now.Add(EnrollTokenTTL).Unix()); err != nil {
		return "", errorfamily.WrapInfrastructure(err, "userauth.enroll.mint", "store enrollment token") //nolint:erraudit // ctx is bound to the exec; the wrap adds the code the journal greps
	}
	return token, nil
}

// VerifyEnrollToken resolves AND burns a token. Unknown and
// already-used tokens answer the SAME rejection; an expired one carries
// its own code — the distinction is operator-log truth only, every
// class answers the identical client shape (503 via the handler), so
// the surface leaks no usable token-state oracle.
func (s *Service) VerifyEnrollToken(ctx context.Context, token string) (EnrollToken, error) {
	if len(token) < 16 {
		return EnrollToken{}, errorfamily.NewRejection("userauth.enroll.invalid", "invalid enrollment token")
	}
	sum := sha256.Sum256([]byte(token))
	row := s.db.QueryRowContext(ctx,
		`SELECT email, user_id, expires_at FROM wp_enroll_tokens WHERE token_hash = ?`,
		hex.EncodeToString(sum[:]))
	var out EnrollToken
	var rawUserID string
	var expiresAt int64
	if err := row.Scan(&out.Email, &rawUserID, &expiresAt); err != nil {
		if err == sql.ErrNoRows {
			return EnrollToken{}, errorfamily.NewRejection("userauth.enroll.invalid",
				"unknown, expired, or already-used enrollment token — mint a fresh one with: webphone -enroll-passkey <email>")
		}
		return EnrollToken{}, errorfamily.WrapInfrastructure(err, "userauth.enroll.lookup", "look up enrollment token")
	}
	userID, err := usermgmt.ParseUserID(rawUserID)
	if err != nil {
		return EnrollToken{}, errorfamily.Wrapf(err, errorfamily.Corruption, "userauth.enroll.userid",
			"stored enrollment token names an invalid user id")
	}
	out.UserID = userID
	if time.Now().Unix() >= expiresAt {
		if _, err := s.db.ExecContext(ctx,
			`DELETE FROM wp_enroll_tokens WHERE token_hash = ?`, hex.EncodeToString(sum[:])); err != nil {
			return EnrollToken{}, errorfamily.WrapInfrastructure(err, "userauth.enroll.expire", "delete expired enrollment token")
		}
		return EnrollToken{}, errorfamily.NewRejection("userauth.enroll.expired",
			"enrollment token expired — mint a fresh one with: webphone -enroll-passkey <email>")
	}
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM wp_enroll_tokens WHERE token_hash = ?`, hex.EncodeToString(sum[:])); err != nil {
		return EnrollToken{}, errorfamily.WrapInfrastructure(err, "userauth.enroll.burn", "burn enrollment token")
	}
	return out, nil
}
