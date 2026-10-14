// Package readmodels owns webphone's metaengine storage: the
// sqlite-backed engine whose read models and dedup windows live in
// their OWN database file (readmodels.db), separate from webphone.db.
//
// Why a separate file: the hand-written migration ladder in store/db.go
// is a ratified invariant (plan N5/N6) — engine-owned DDL never gets
// schema_version entries, and the ladder never learns about engine
// tables. Two files make both doctrines mechanically true: each side
// owns its DDL outright, and the backup unit's dataDir copy carries
// both without any coupling.
//
// The package stays framework-free (no samber/do imports): lifecycle
// conformance is asserted adapter-side in internal/app, exactly like
// store.Database.
package readmodels

import (
	"context"
	"database/sql"
	"path/filepath"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	sqliteengine "github.com/larsartmann/go-cqrs-lite/metaengine/sqliteengine/v4"
	"github.com/larsartmann/go-error-family"
)

// fileName is the engine-owned database inside the data dir. The name
// is this package's single fact: callers hand over the data dir, never
// a full path.
const fileName = "readmodels.db"

// Engine is the container-adapted metaengine engine: the sqlite handle
// plus the engine that owns its tables. Constructed once in the
// composition root (named "metaengine"), eagerly invoked, closed on
// the container shutdown cascade.
type Engine struct {
	eng    metaengine.Engine
	dedup  metaengine.DedupStore
	db     *sql.DB
	closed bool
}

// Open creates (or opens) the read-model database inside dataDir and
// builds the sqlite engine over it. The engine applies its production
// pragmas (WAL, busy_timeout) and creates its own tables — one call,
// fail-fast at boot.
func Open(dataDir string) (*Engine, error) {
	eng, db, err := sqliteengine.NewFromDSN(filepath.Join(dataDir, fileName))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.open", "open read-model engine in %s", dataDir)
	}
	// Asserted once here, not per call: the sqlite engine implements
	// the DedupStore capability natively, and a future engine swap that
	// silently dropped it must fail the boot, not the first webhook.
	dedup, ok := eng.(metaengine.DedupStore)
	if !ok {
		_ = db.Close()
		return nil, errorfamily.Newf(errorfamily.Infrastructure, "readmodels.open", "engine %T lacks the dedup capability", eng)
	}
	return &Engine{eng: eng, dedup: dedup, db: db}, nil
}

// Engine exposes the underlying metaengine engine for the Store/Query
// wiring that grows with the dogfooding train. The composition root is
// the only caller; services receive derived seams (dedup windows,
// readers), never the engine itself.
func (e *Engine) Engine() metaengine.Engine { return e.eng }

// DedupStore returns the engine's dedup capability (asserted at Open).
// The composition root adapts it into the server-facing DedupWindow
// port — readmodels itself never imports server.
func (e *Engine) DedupStore() metaengine.DedupStore { return e.dedup }

// SQL exposes the engine's own sqlite handle (probe pings; future
// engine-side maintenance). Callers must not close it — Shutdown owns
// that.
func (e *Engine) SQL() *sql.DB { return e.db }

// HealthCheck pings the read-model database: a closed handle, a lost
// file, or a wedged single-writer connection fails here.
func (e *Engine) HealthCheck(ctx context.Context) error {
	if err := e.db.PingContext(ctx); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.ping", "ping read-model sqlite")
	}
	return nil
}

// Shutdown closes the engine (its in-memory state) and then the sqlite
// handle it owns. Idempotent; the second call is inert.
func (e *Engine) Shutdown() error {
	if e.closed {
		return nil
	}
	e.closed = true
	if err := e.eng.Close(); err != nil {
		_ = e.db.Close()
		return errorfamily.WrapInfrastructuref(err, "readmodels.close", "close read-model engine")
	}
	if err := e.db.Close(); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.close", "close read-model sqlite")
	}
	return nil
}
