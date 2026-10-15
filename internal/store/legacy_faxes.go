package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// Legacy-faxes seam (2026-10-15): the fax-job truth moved to the
// metaengine read models (readmodels.db). The SQL fax_jobs table stays
// untouched in webphone.db (the migration ladder never rewrites
// shipped shape), so an upgraded deployment's fax history rides this
// ONE transitional read at boot: the composition root imports the rows
// into the engine store once and stamps the app_flags row. This file
// exists to be deleted when the transitional window closes.

// FlagLegacyFaxesImported marks the one-time legacy-faxes import as
// complete (value carries the import date for operators).
const FlagLegacyFaxesImported = "legacy_faxes_imported"

// LegacyFax is one row of the retired SQL fax_jobs table, kept as
// plain values: import-time parsing decides what survives.
// brandid-lint:ignore(BD003) retired SQL row kept as plain values; import-time parsing decides what survives
type LegacyFax struct {
	ID           string
	Owner        string
	Remote       string
	Direction    string
	Status       string
	Pages        int
	DocumentPath string
	ProviderRef  string
	Error        string
	CreatedAt    int64
	UpdatedAt    int64
}

// ReadLegacyFaxes reads the whole retired table oldest-first (the
// engine list is newest-first; importing in order preserves the
// legacy ordering semantics).
func ReadLegacyFaxes(ctx context.Context, db *sql.DB) ([]LegacyFax, error) {
	return listRows(ctx, db, "read legacy faxes", `
		SELECT id, owner, remote, direction, status, pages, document_path, provider_ref, error, created_at, updated_at
		FROM fax_jobs
		ORDER BY created_at ASC, rowid ASC
	`, nil, func(row rowScanner) (LegacyFax, error) {
		var f LegacyFax
		return f, row.Scan(&f.ID, &f.Owner, &f.Remote, &f.Direction, &f.Status, &f.Pages,
			&f.DocumentPath, &f.ProviderRef, &f.Error, &f.CreatedAt, &f.UpdatedAt)
	})
}

// ToFax validates one legacy row into a domain.FaxJob. A bad id
// demotes to a freshly minted one; a row whose owner or remote does
// not parse, or whose direction/status is not a known enum value, is
// skipped (the caller logs it) — a corrupt legacy row must not brick a
// boot, and an unknown lifecycle value would poison the projection.
func (f LegacyFax) ToFax() (domain.FaxJob, bool) {
	owner, err := domain.ParseExtension(f.Owner)
	if err != nil {
		return domain.FaxJob{}, false
	}
	remote, err := domain.ParsePhone(f.Remote)
	if err != nil {
		return domain.FaxJob{}, false
	}
	direction := domain.FaxDirection(f.Direction)
	if direction != domain.FaxOutbound && direction != domain.FaxInbound {
		return domain.FaxJob{}, false
	}
	status := domain.FaxStatus(f.Status)
	switch status {
	case domain.FaxQueued, domain.FaxSending, domain.FaxTransmitted, domain.FaxReceived, domain.FaxFailed:
	default:
		return domain.FaxJob{}, false
	}
	id, err := domain.ParseFaxID(f.ID)
	if err != nil {
		id = domain.GenerateFaxID()
	}
	return domain.FaxJob{
		ID:           id,
		Owner:        owner,
		Remote:       remote,
		Direction:    direction,
		Status:       status,
		Pages:        f.Pages,
		DocumentPath: f.DocumentPath,
		ProviderRef:  f.ProviderRef,
		Error:        f.Error,
		CreatedAt:    time.Unix(f.CreatedAt, 0),
		UpdatedAt:    time.Unix(f.UpdatedAt, 0),
	}, true
}
