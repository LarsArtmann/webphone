// Package readmodels owns webphone's read-model GATE + VIEW surface:
// the interfaces the server consumes and the projection row shapes the
// journal backends (internal/cqrs) fold into system.db. The storage
// itself lives in internal/cqrs — since the 2026-11-11 storage-ladder
// closeout (A43 lineage) NOTHING here opens a database: the files are
// typed contracts only, and the purity is pinned by internal/arch.
//
// Historical note: the package used to also own a standalone
// sqlite-backed engine (readmodels.db) whose stores implemented these
// gates in-process. That engine retired train-by-train across the
// SUPERB storage-ladder plan (transcripts, notes/reviews,
// contacts/snippets/dialwindows, push/agent-audit — the 2026-11-09
// A43 ruling stopped the composition root from opening readmodels.db
// long before the last store fell). On upgraded deployments the
// readmodels.db file is an inert orphan outside the backup set.
//
// The package stays framework-free (no samber/do imports): lifecycle
// conformance is asserted adapter-side in internal/app, exactly like
// store.Database.
package readmodels
