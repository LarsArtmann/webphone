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

	sqliteengine "github.com/larsartmann/go-cqrs-lite/metaengine/sqliteengine/v4"
	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-error-family"
)

// fileName is the engine-owned database inside the data dir. The name
// is this package's single fact: callers hand over the data dir, never
// a full path.
const fileName = "readmodels.db"

// humanScaleVolume is the Volume hint every query declaration carries:
// webphone is single-tenant, and even its highest-rate facts (live-call
// transcript segments) average well under one per second — the
// planner's 1000/sec default would overestimate by three orders.
const humanScaleVolume = 1

// Engine is the container-adapted metaengine engine: the sqlite
// handle, the engine that owns its tables, and the planned model
// store over it. Constructed once in the composition root (named
// "metaengine"), eagerly invoked, closed on the container shutdown
// cascade.
type Engine struct {
	eng      metaengine.Engine
	dedup    metaengine.DedupStore
	agg      metaengine.AggregateReader
	models   *metaengine.Store
	contacts *Contacts
	snippets *Snippets
	faxes    *Faxes
	calls    *Transcripts
	notes    *Notes
	reviews  *Reviews
	enroll   *EnrollTokens
	devices  *PushDevices
	db       *sql.DB
	closed   bool
}

// Open creates (or opens) the read-model database inside dataDir,
// builds the sqlite engine over it, and plans every read-model query
// the binary declares (the list grows per train step). The engine
// applies its production pragmas (WAL, busy_timeout) and creates its
// own tables — one call, fail-fast at boot.
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
	// Same doctrine for the aggregate reader: the metrics surface counts
	// contacts through it, so a swap that dropped it fails the boot.
	agg, ok := eng.(metaengine.AggregateReader)
	if !ok {
		_ = db.Close()
		return nil, errorfamily.Newf(errorfamily.Infrastructure, "readmodels.open", "engine %T lacks the aggregate capability", eng)
	}
	models, err := metaengine.Plan([]metaengine.Engine{eng}, contactsQuery, snippetsQuery, faxesQuery, transcriptsQuery, notesQuery, enrollTokensQuery, pushDevicesQuery)
	if err != nil {
		_ = db.Close()
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.open", "plan read models")
	}
	return &Engine{
		eng:      eng,
		dedup:    dedup,
		agg:      agg,
		models:   models,
		contacts: newContacts(models, agg),
		snippets: newSnippets(models),
		faxes:    newFaxes(models, agg),
		calls:    newTranscripts(models, agg),
		notes:    newNotes(models),
		reviews:  newReviews(models),
		enroll:   newEnrollTokens(models),
		devices:  newPushDevices(models),
		db:       db,
	}, nil
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

// Contacts returns the engine-backed personal-contacts store (the
// SQL store's replacement behind the same method contract).
func (e *Engine) Contacts() *Contacts { return e.contacts }

// Snippets returns the engine-backed reply-snippet store (the SQL
// store's replacement behind the same method contract).
func (e *Engine) Snippets() *Snippets { return e.snippets }

// Faxes returns the engine-backed fax store (the SQL store's
// replacement behind the same method contract).
func (e *Engine) Faxes() *Faxes { return e.faxes }

// Transcripts returns the engine-backed transcript store (the SQL
// store's replacement behind the same method contract).
func (e *Engine) Transcripts() *Transcripts { return e.calls }

// Notes returns the engine-backed post-call-note store (M16).
func (e *Engine) Notes() *Notes { return e.notes }

// Reviews returns the engine-backed post-call-review store (issue #8:
// the automatic review's durable artifact slot).
func (e *Engine) Reviews() *Reviews { return e.reviews }

// EnrollTokens returns the engine-backed enrollment-token store (the
// userauth SQL table's replacement; token classification stays in the
// userauth service).
func (e *Engine) EnrollTokens() *EnrollTokens { return e.enroll }

// PushDevices returns the engine-backed registered-device store (the
// push seam's ring fan-out source; owner-scoped everywhere).
func (e *Engine) PushDevices() *PushDevices { return e.devices }

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

// Shutdown closes the model store, the engine (its in-memory state),
// and the sqlite handle it owns. Idempotent; the second call is inert.
func (e *Engine) Shutdown() error {
	if e.closed {
		return nil
	}
	e.closed = true
	if err := e.models.Close(); err != nil {
		_ = e.eng.Close()
		_ = e.db.Close()
		return errorfamily.WrapInfrastructuref(err, "readmodels.close", "close read-model store")
	}
	if err := e.eng.Close(); err != nil {
		_ = e.db.Close()
		return errorfamily.WrapInfrastructuref(err, "readmodels.close", "close read-model engine")
	}
	if err := e.db.Close(); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.close", "close read-model sqlite")
	}
	return nil
}
