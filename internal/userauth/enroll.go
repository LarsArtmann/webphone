package userauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	identitymodel "github.com/larsartmann/cqrs-htmx/identity-model/v4"
	"github.com/larsartmann/go-error-family"
)

// EnrollTokenTTL bounds how long a CLI-minted enrollment link lives.
// Fifteen minutes: enough to walk from the shell to the browser, short
// enough that a leaked URL (terminal scrollback) is a narrow window.
const EnrollTokenTTL = 15 * time.Minute

// EnrollToken is what a verified enrollment token resolves to.
type EnrollToken struct {
	Email  string
	UserID identitymodel.UserID
}

// MintEnrollToken creates a one-time enrollment token for an existing
// usermgmt account and returns the URL-safe secret exactly once. Only its
// SHA-256 is stored; the row is burned on first use and expired rows are
// swept opportunistically on every mint.
func (s *Service) MintEnrollToken(ctx context.Context, email string, userID identitymodel.UserID) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", errorfamily.WrapInfrastructure(err, "userauth.enroll.entropy", "generate enrollment token entropy") //nolint:erraudit // crypto/rand takes no ctx; the wrap owns the code
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	if err := s.enroll.Mint(ctx, hex.EncodeToString(sum[:]), email, userID.Get().String(), time.Now().Add(EnrollTokenTTL)); err != nil {
		return "", err //nolint:erraudit // the enroll seam owns the classification (readmodels.enroll_*)
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
	hash := hex.EncodeToString(sum[:])
	email, rawUserID, expiresAt, found, err := s.enroll.Lookup(ctx, hash)
	if err != nil {
		return EnrollToken{}, err //nolint:erraudit // the enroll seam owns the classification (readmodels.enroll_lookup)
	}
	if !found {
		return EnrollToken{}, errorfamily.NewRejection("userauth.enroll.invalid",
			"unknown, expired, or already-used enrollment token — mint a fresh one with: webphone -enroll-passkey <email>")
	}
	userID, err := identitymodel.ParseUserID(rawUserID)
	if err != nil {
		return EnrollToken{}, errorfamily.Wrapf(err, errorfamily.Corruption, "userauth.enroll.userid",
			"stored enrollment token names an invalid user id %q", rawUserID)
	}
	if !time.Now().Before(expiresAt) {
		if err := s.enroll.Burn(ctx, hash); err != nil {
			return EnrollToken{}, err //nolint:erraudit // the enroll seam owns the classification (readmodels.enroll_burn)
		}
		return EnrollToken{}, errorfamily.NewRejection("userauth.enroll.expired",
			"enrollment token expired — mint a fresh one with: webphone -enroll-passkey <email>")
	}
	if err := s.enroll.Burn(ctx, hash); err != nil {
		return EnrollToken{}, err //nolint:erraudit // the enroll seam owns the classification (readmodels.enroll_burn)
	}
	return EnrollToken{Email: email, UserID: userID}, nil
}
