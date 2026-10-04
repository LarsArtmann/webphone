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
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/larsartmann/cqrs-htmx/usermgmt/v4"
	waprovider "github.com/larsartmann/cqrs-htmx/usermgmt/webauthn/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// dbDriver is the modernc.org/sqlite driver name the whole app registers.
const dbDriver = "sqlite"

// Service is the embedded identity layer. Construct only when the passkey
// config is enabled; nil on Deps means the mode is off and every passkey
// surface 404s.
type Service struct {
	users *usermgmt.Service
	db    *sql.DB
	cfg   PasskeyRuntime
	log   *slog.Logger
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
}

// MappedUser is one email's resolved mapping.
type MappedUser struct {
	Email       string
	DisplayName string
	Extensions  []domain.Extension
}

// SessionExtension is the extension a passkey login binds its session to
// (the first mapped one — all stores are extension-scoped).
func (m MappedUser) SessionExtension() domain.Extension {
	if len(m.Extensions) == 0 {
		return domain.Extension{}
	}
	return m.Extensions[0]
}

// New opens (creating if needed) <dataDir>/usermgmt.db and wires the
// usermgmt service: SQL event store + SQL read models + the WebAuthn
// provider. The journal is tiny; startup drains projections synchronously
// (the default) so a broken store fails the boot instead of the login.
func New(ctx context.Context, cfg PasskeyRuntime, dataDir string, log *slog.Logger) (*Service, error) {
	if log == nil {
		log = slog.Default()
	}
	if cfg.RPID == "" || len(cfg.RPOrigins) == 0 || len(cfg.Users) == 0 {
		return nil, errorfamily.NewRejection("userauth.config", "passkey runtime config is incomplete (rp_id, rp_origins and users are all required)")
	}
	db, err := sql.Open(dbDriver, filepath.Join(dataDir, "usermgmt.db")+
		"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, errorfamily.WrapInfrastructure(err, "userauth.db.open", "open usermgmt database")
	}
	db.SetMaxOpenConns(1)
	if err := usermgmt.OptimizeSQLiteDB(ctx, db); err != nil {
		db.Close() //nolint:errcheck // best-effort cleanup on a failed boot path
		return nil, errorfamily.WrapInfrastructure(err, "userauth.db.optimize", "tune usermgmt database pragmas")
	}
	eventStore, err := usermgmt.NewSQLEventStore(ctx, db, "sqlite")
	if err != nil {
		db.Close() //nolint:errcheck // best-effort cleanup on a failed boot path
		return nil, errorfamily.Wrapf(err, errorfamily.Classify(err), "userauth.event_store", "create sqlite event store")
	}
	displayName := cfg.RPDisplayName
	if displayName == "" {
		displayName = "webphone"
	}
	provider, err := waprovider.New(waprovider.Config{
		RPID:          cfg.RPID,
		RPDisplayName: displayName,
		RPOrigins:     cfg.RPOrigins,
	})
	if err != nil {
		db.Close() //nolint:errcheck // best-effort cleanup on a failed boot path
		return nil, errorfamily.Wrapf(err, errorfamily.Classify(err), "userauth.webauthn_provider", "create webauthn provider")
	}
	users, err := usermgmt.NewService(usermgmt.ServiceConfig{
		EventStore:       eventStore,
		ReadModelDB:      db,
		ReadModelDialect: "sqlite",
		WebAuthn:         provider,
		Logger:           log,
	})
	if err != nil {
		db.Close() //nolint:errcheck // best-effort cleanup on a failed boot path
		return nil, errorfamily.Wrapf(err, errorfamily.Classify(err), "userauth.service", "create usermgmt service")
	}
	svc := &Service{users: users, db: db, cfg: cfg, log: log}
	if err := svc.migrateEnrollTokens(ctx); err != nil {
		svc.Close() //nolint:errcheck // best-effort cleanup on a failed boot path
		return nil, err
	}
	return svc, nil
}

// Close drains usermgmt's projections and closes the database. Safe to
// call once; part of the app's shutdown order.
func (s *Service) Close() error {
	var firstErr error
	if err := s.users.Close(); err != nil && firstErr == nil {
		firstErr = errorfamily.WrapInfrastructure(err, "userauth.close.users", "close usermgmt service")
	}
	if err := s.db.Close(); err != nil && firstErr == nil {
		firstErr = errorfamily.WrapInfrastructure(err, "userauth.close.db", "close usermgmt database")
	}
	return firstErr
}

// BeginLogin starts the passkey login ceremony for an email. The
// response's SessionKey doubles as the FinishLogin userID (usermgmt's own
// wire contract — see its AuthHandler).
func (s *Service) BeginLogin(ctx context.Context, email string) (*usermgmt.BeginLoginResponse, error) {
	return s.users.BeginLogin(ctx, email)
}

// FinishLogin verifies the assertion for the userID the begin ceremony
// returned and resolves the authenticated user's extension mapping. The
// usermgmt session in the ceremony result is intentionally ignored (the
// webphone session is the single session truth).
func (s *Service) FinishLogin(ctx context.Context, userID string, r *http.Request) (MappedUser, error) {
	empty := MappedUser{}
	resp, err := s.users.FinishLogin(ctx, usermgmt.NewUserID(userID), r)
	if err != nil {
		return empty, err
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
	mapped, ok := s.cfg.Users[strings.ToLower(strings.TrimSpace(email))]
	return mapped, ok
}

// SIPPassword reads the extension's directory password from its
// operator-managed file — at login time, never cached, so rotation via
// the secrets pipeline takes effect on the next login. Empty or
// whitespace-only files are Rejections (the 2026-10-01 outage class: an
// empty secret must fail loudly, never mint a passwordless session).
func (s *Service) SIPPassword(extension domain.Extension) (string, error) {
	file, ok := s.cfg.ExtensionPasswordFiles[extension.String()]
	if !ok || file == "" {
		return "", errorfamily.NewRejection("userauth.password_file.missing",
			"no password file configured for extension %q (auth.passkey.extension_password_files)").WithContext("extension", extension.String())
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return "", errorfamily.WrapRejection(err, "userauth.password_file.read",
			"read password file for extension %q (check the secrets pipeline and the secrets-perms grants)").WithContext("extension", extension.String())
	}
	password := strings.TrimSpace(string(raw))
	if password == "" {
		return "", errorfamily.NewRejection("userauth.password_file.empty",
			"password file for extension %q is empty").WithContext("extension", extension.String())
	}
	return password, nil
}

// Register creates the usermgmt account if it does not exist yet and
// returns its ID. An existing account is NOT an error — minting an
// enrollment token for it is the normal add-a-device flow.
func (s *Service) Register(ctx context.Context, email string) (string, error) {
	resp, err := s.users.Register(ctx, usermgmt.RegisterRequest{Email: email})
	if err == nil {
		return resp.User.ID.Get().String(), nil
	}
	if errors.Is(err, usermgmt.ErrEmailExists) {
		existing, ok := s.users.ReadModel().FindByEmail(email)
		if !ok {
			return "", errorfamily.WrapInfrastructure(err, "userauth.register.exists_unreadable",
				"account exists but the read model cannot resolve it").WithContext("email", email)
		}
		return existing.ID.Get().String(), nil
	}
	return "", err
}

// BeginRegistration / FinishRegistration drive the passkey enrollment
// ceremony for an existing usermgmt account.
func (s *Service) BeginRegistration(ctx context.Context, userID string) (*usermgmt.BeginRegistrationResponse, error) {
	return s.users.BeginRegistration(ctx, usermgmt.NewUserID(userID))
}

func (s *Service) FinishRegistration(ctx context.Context, userID string, r *http.Request, credentialName string) error {
	return s.users.FinishRegistration(ctx, usermgmt.NewUserID(userID), r, credentialName)
}
