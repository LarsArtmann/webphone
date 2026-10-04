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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/larsartmann/cqrs-htmx/usermgmt/v4"
	waprovider "github.com/larsartmann/cqrs-htmx/usermgmt/webauthn/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// dbDriver is the modernc.org/sqlite driver name the whole app registers.
const dbDriver = "sqlite"

// Service is the embedded identity layer. Construct only when the passkey
// config is enabled; nil on Deps means the mode is off and every
// passkey surface 404s.
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
	RPID                    string
	RPDisplayName           string
	RPOrigins               []string
	Users                   map[string]MappedUser
	ExtensionPasswordFiles  map[string]string
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

// BeginLogin starts the passkey login ceremony for an email.
func (s *Service) BeginLogin(ctx context.Context, email string) (*usermgmt.BeginLoginResponse, error) {
	return s.users.BeginLogin(ctx, email)
}

// FinishLogin verifies the assertion and returns the authenticated user.
// The usermgmt session in the response is intentionally ignored (see the
// package doc).
func (s *Service) FinishLogin(ctx context.Context, email string, r *requestLike) (*usermgmt.User, error) {
	mapped, ok := s.Resolve(email)
	if !ok {
		return nil, errorfamily.NewRejection("userauth.unmapped_email", "email is not in the passkey users mapping")
	}
	resp, err := s.users.FinishLogin(ctx, mustUserID(mapped.Email), r)
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}
