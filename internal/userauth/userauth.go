// Package userauth embeds the cqrs-htmx/usermgmt identity layer: passkey
// (WebAuthn) users persisted in their own SQLite database under the data
// dir, plus the config-declared email→extension mapping that turns a
// verified passkey login into a webphone session.
//
// The webphone session stays the single session truth — usermgmt's own
// session (a FinishLogin side effect) is deliberately ignored; the island
// and the /phone-api proxy ride the webphone session as before. The SIP
// directory password a passkey session carries is read from an
// operator-managed file at login time (config ExtensionPasswordFiles),
// never cached, mirroring how the bridge sources its secrets.
package userauth

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	identitymodel "github.com/larsartmann/cqrs-htmx/identity-model/v4"
	"github.com/larsartmann/cqrs-htmx/usermgmt/v4"
	waprovider "github.com/larsartmann/cqrs-htmx/usermgmt/webauthn/v4"
	"github.com/larsartmann/go-error-family"
	_ "modernc.org/sqlite" // registers the "sqlite" driver (the same one the store package registers)

	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// dbDriver is the modernc.org/sqlite driver name the whole app registers.
const dbDriver = "sqlite"

// Service is the embedded identity layer. Construct only when the passkey
// config is enabled; nil on Deps means the mode is off and every passkey
// surface 404s.
type Service struct {
	users  *usermgmt.Service
	db     *sql.DB
	enroll *readmodels.EnrollTokens
	mapped *authmap.Map
	cfg    PasskeyRuntime
	log    *slog.Logger
}

// PasskeyRuntime is the resolved passkey configuration the service needs
// (a narrow copy of config.Passkey so this package stays decoupled from
// the config package — the composition root adapts).
type PasskeyRuntime struct {
	RPID                   string
	RPDisplayName          string
	RPOrigins              []string
	Users                  map[string]MappedUser
	ExtensionPasswordFiles map[string]string
	// WebAuthn optionally overrides the go-webauthn provider built from
	// RPID/RPOrigins. Tests inject the deterministic stub here; the
	// production composition root leaves it nil.
	WebAuthn identitymodel.WebAuthnProvider
}

// MappedUser is one email's resolved mapping. It IS the shared
// authmap.User (a type alias): passkey and OIDC logins must resolve
// identities through one implementation, never two.
type MappedUser = authmap.User

// New opens (creating if needed) <dataDir>/usermgmt.db and wires the
// usermgmt service: SQL event store + SQL read models + the WebAuthn
// provider. Enrollment tokens live in the metaengine store the caller
// hands over (engine.EnrollTokens()) — the wp_enroll_tokens SQL table
// is gone (the 15-minute TTL means an upgrade carries no continuity;
// in-flight links get re-minted). The journal is tiny; startup drains
// projections synchronously (the default) so a broken store fails the
// boot instead of the login.
func New(ctx context.Context, cfg PasskeyRuntime, dataDir string, enroll *readmodels.EnrollTokens, log *slog.Logger) (*Service, error) {
	if log == nil {
		log = slog.Default()
	}
	if cfg.RPID == "" || len(cfg.RPOrigins) == 0 || len(cfg.Users) == 0 {
		return nil, errorfamily.NewRejection("userauth.config", "passkey runtime config is incomplete (rp_id, rp_origins and users are all required)") //nolint:erraudit // pure config verdict: nothing failed, no ctx to carry
	}
	db, err := sql.Open(dbDriver, filepath.Join(dataDir, "usermgmt.db")+
		"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, errorfamily.WrapInfrastructure(err, "userauth.db.open", "open usermgmt database") //nolint:erraudit // the wrap owns classification + operator copy; sql.Open takes no ctx
	}
	db.SetMaxOpenConns(1)
	if err := usermgmt.OptimizeSQLiteDB(ctx, db); err != nil {
		db.Close()                                                                                                //nolint:errcheck // best-effort cleanup on a failed boot path
		return nil, errorfamily.WrapInfrastructure(err, "userauth.db.optimize", "tune usermgmt database pragmas") //nolint:erraudit // ctx is bound to the call; the wrap adds the code the journal greps
	}
	eventStore, err := usermgmt.NewSQLEventStore(ctx, db, "sqlite")
	if err != nil {
		db.Close()                                                                                                         //nolint:errcheck // best-effort cleanup on a failed boot path
		return nil, errorfamily.Wrapf(err, errorfamily.Classify(err), "userauth.event_store", "create sqlite event store") //nolint:erraudit // ctx is bound to the call; the wrap adds the code the journal greps
	}
	displayName := cfg.RPDisplayName
	if displayName == "" {
		displayName = "webphone"
	}
	// The provider seam: the composition root gets the real
	// go-webauthn provider; an injected stub keeps the service logic
	// testable without browser signatures.
	provider := cfg.WebAuthn
	if provider == nil {
		provider, err = waprovider.New(waprovider.Config{
			RPID:          cfg.RPID,
			RPDisplayName: displayName,
			RPOrigins:     cfg.RPOrigins,
		})
		if err != nil {
			db.Close()                                                                                                              //nolint:errcheck // best-effort cleanup on a failed boot path
			return nil, errorfamily.Wrapf(err, errorfamily.Classify(err), "userauth.webauthn_provider", "create webauthn provider") //nolint:erraudit // ctx is bound to the call; the wrap adds the code the journal greps
		}
	}
	users, err := usermgmt.NewService(usermgmt.ServiceConfig{
		EventStore:       eventStore,
		ReadModelDB:      db,
		ReadModelDialect: "sqlite",
		WebAuthn:         provider,
		Logger:           log,
		// A projection worker exhausting its restart budget is NOT a boot
		// error (usermgmt's own contract) — it surfaces via
		// ProjectionStatuses, which the /healthz "projections" check
		// reads. This hook makes the same event greppable in the journal
		// the moment it happens, not only when a prober asks.
		OnProjectionFailed: func(projectionName, lastError string) {
			log.Warn("usermgmt projection worker failed", "projection", projectionName, "last_error", lastError)
		},
	})
	if err != nil {
		db.Close()                                                                                                   //nolint:errcheck // best-effort cleanup on a failed boot path
		return nil, errorfamily.Wrapf(err, errorfamily.Classify(err), "userauth.service", "create usermgmt service") //nolint:erraudit // ctx is bound to the call; the wrap adds the code the journal greps
	}
	mapped, err := authmap.New(cfg.Users, cfg.ExtensionPasswordFiles)
	if err != nil {
		db.Close()                                                                                          //nolint:errcheck // best-effort cleanup on a failed boot path
		return nil, errorfamily.WrapRejection(err, "userauth.mapping", "build the email→extension mapping") //nolint:erraudit // pure config verdict: nothing failed, no ctx to carry
	}
	svc := &Service{users: users, db: db, enroll: enroll, mapped: mapped, cfg: cfg, log: log}
	return svc, nil
}

// HealthCheck pings the identity layer's own SQLite database (the
// do.HealthcheckerWithContext contract): the go-health probe and /healthz
// both ride it, so a broken usermgmt.db is visible on every health
// surface instead of surfacing as unexplained passkey 503s. The
// usermgmt service itself needs no separate probe — its failure mode
// IS the database.
func (s *Service) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Shutdown drains usermgmt's projections and closes the database (the
// do.ShutdownerWithError lifecycle method; part of the app's shutdown
// order). Safe to call once.
func (s *Service) Shutdown() error {
	var firstErr error
	if err := s.users.Close(); err != nil {
		firstErr = errorfamily.WrapInfrastructure(err, "userauth.close.users", "close usermgmt service")
	}
	if err := s.db.Close(); err != nil && firstErr == nil {
		firstErr = errorfamily.WrapInfrastructure(err, "userauth.close.db", "close usermgmt database")
	}
	return firstErr
}

// Usermgmt exposes the embedded usermgmt service for READ-ONLY page
// wiring: the loginpage package reads HasWebAuthn to decide which
// sections to render. Every ceremony still runs through this seam's
// methods — nothing outside may drive the identity layer directly.
func (s *Service) Usermgmt() *usermgmt.Service {
	return s.users
}

// BeginLogin starts the passkey login ceremony for an email. The
// response's SessionKey doubles as the FinishLogin userID (usermgmt's own
// wire contract — see its AuthHandler).
func (s *Service) BeginLogin(ctx context.Context, email string) (*usermgmt.BeginLoginResponse, error) {
	return s.users.BeginLogin(ctx, email)
}

// FinishLogin verifies the assertion for the userID the begin ceremony
// returned and resolves the authenticated user's extension mapping. The
// userID arrives strictly parsed (a ULID-branded identitymodel.UserID): the
// browser-provided session key is validated at the handler door, never
// hashed into a pseudo-user here. The usermgmt session in the ceremony
// result is intentionally ignored (the webphone session is the single
// session truth).
func (s *Service) FinishLogin(ctx context.Context, userID identitymodel.UserID, r *http.Request) (MappedUser, error) {
	empty := MappedUser{}
	resp, err := s.users.FinishLogin(ctx, userID, r)
	if err != nil {
		return empty, err //nolint:erraudit // usermgmt owns the classification; this seam only resolves the mapping
	}
	mapped, ok := s.Resolve(resp.User.Email)
	if !ok {
		return empty, errorfamily.NewRejection("userauth.unmapped_email",
			"the passkey account %q is not in the users mapping (auth.passkey.users) — add the mapping or remove the account").WithContext("email", resp.User.Email)
	}
	return mapped, nil
}

// Resolve maps an email to its configured extensions; ok=false when the
// email has no mapping (login fails closed — a passkey alone must never
// mint a session for an unmapped account).
func (s *Service) Resolve(email string) (MappedUser, bool) {
	return s.mapped.Resolve(email)
}

// MappedByExtension reverse-resolves: which mapped user (if any) a
// session extension belongs to. Powers the display identity on the
// extension login too — a mapped user gets their name and numbers in
// the whoami line regardless of how they signed in. Deterministic on
// shared extensions (see authmap).
func (s *Service) MappedByExtension(extension domain.Extension) (MappedUser, bool) {
	return s.mapped.MappedByExtension(extension)
}

// MappedByLogin resolves the identity a SESSION should display: the
// session's own login email when it maps to the session's extension,
// else the deterministic extension lookup (see authmap — the one home
// of the rule).
func (s *Service) MappedByLogin(extension domain.Extension, loginEmail string) (MappedUser, bool) {
	return s.mapped.MappedByLogin(extension, loginEmail)
}

// SIPPassword reads the extension's directory password from its
// operator-managed file — at login time, never cached, so rotation via
// the secrets pipeline takes effect on the next login. Empty or
// whitespace-only files are Rejections (the 2026-10-01 outage class: an
// empty secret must fail loudly, never mint a passwordless session).
func (s *Service) SIPPassword(extension domain.Extension) (string, error) {
	return s.mapped.SIPPassword(extension)
}

// Register creates the usermgmt account if it does not exist yet and
// returns its ID. An existing account is NOT an error — minting an
// enrollment token for it is the normal add-a-device flow.
func (s *Service) Register(ctx context.Context, email string) (identitymodel.UserID, error) {
	resp, err := s.users.Register(ctx, usermgmt.RegisterRequest{Email: email})
	if err == nil {
		return resp.User.ID, nil
	}
	if errors.Is(err, identitymodel.ErrEmailExists) {
		existing, ok := s.users.ReadModel().FindByEmail(email)
		if !ok {
			return identitymodel.UserID{}, errorfamily.WrapInfrastructure(err, "userauth.register.exists_unreadable",
				"account exists but the read model cannot resolve it").WithContext("email", email)
		}
		return existing.ID, nil
	}
	return identitymodel.UserID{}, err
}

// BeginRegistration / FinishRegistration drive the passkey enrollment
// ceremony for an existing usermgmt account (the userID strictly parsed
// by the caller — a CLI literal or a token-resolved account).
func (s *Service) BeginRegistration(ctx context.Context, userID identitymodel.UserID) (*usermgmt.BeginRegistrationResponse, error) {
	return s.users.BeginRegistration(ctx, userID)
}

func (s *Service) FinishRegistration(ctx context.Context, userID identitymodel.UserID, r *http.Request, credentialName string) error {
	return s.users.FinishRegistration(ctx, userID, r, credentialName)
}

// PasskeyCredential is one enrolled passkey as the settings surface sees
// it: the base64url ID (the revoke handle), the operator/device name and
// the enrollment time. No key material ever leaves the store.
// brandid-lint:ignore(BD003) base64url encoding of the authenticator's WebAuthn credential id, decoded at the remove door (strong-id ruling)
type PasskeyCredential struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// Credentials lists the mapped account's enrolled passkeys (owner-match:
// the caller resolves the email from the session's extension mapping, so
// an extension can only ever see its own account). ok=false when the
// email has no usermgmt account (nothing enrolled — the settings section
// renders its empty state).
func (s *Service) Credentials(email string) ([]PasskeyCredential, bool) {
	user, ok := s.users.ReadModel().FindByEmail(strings.ToLower(strings.TrimSpace(email)))
	if !ok {
		return nil, false
	}
	out := make([]PasskeyCredential, 0, len(user.Credentials))
	for _, cred := range user.Credentials {
		out = append(out, PasskeyCredential{
			ID:        base64.RawURLEncoding.EncodeToString(cred.ID),
			Name:      cred.Name,
			CreatedAt: cred.CreatedAt,
		})
	}
	return out, true
}

// RemoveCredential revokes one passkey of the mapped account by its
// base64url ID. The email comes from the session's extension mapping
// (owner-match by construction); an unmapped account or an unknown id is
// a Rejection.
// brandid-lint:ignore(BD003) handler-door input: the base64url decode validates it (strong-id wire ruling)
func (s *Service) RemoveCredential(ctx context.Context, email, credentialID string) error { //nolint:branching-flow // email+credentialID: handler-door inputs with distinct validation paths; branding the email would fight usermgmt.UserID
	user, ok := s.users.ReadModel().FindByEmail(strings.ToLower(strings.TrimSpace(email)))
	if !ok {
		return errorfamily.NewRejection("userauth.passkey.unmapped",
			"no identity account for %q — nothing to revoke").WithContext("email", email)
	}
	credID, err := base64.RawURLEncoding.DecodeString(credentialID)
	if err != nil {
		return errorfamily.NewRejection("userauth.passkey.bad_id", "credential id is not base64url")
	}
	if err := s.users.RemoveCredential(ctx, user.ID, credID); err != nil {
		return err //nolint:wrapcheck // usermgmt owns the classification; the handler maps families to statuses
	}
	return nil
}
