package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// Legacy-snippets seam (2026-10-15): the reply-snippets truth moved to
// the metaengine read models (readmodels.db). The SQL snippets table
// stays untouched in webphone.db (the migration ladder never rewrites
// shipped shape), so an upgraded deployment's snippets ride this ONE
// transitional read at boot: the composition root imports them into
// the engine store once and stamps the app_flags row. This file exists
// to be deleted when the transitional window closes.

// FlagLegacySnippetsImported marks the one-time legacy-snippets import
// as complete (value carries the import date for operators).
const FlagLegacySnippetsImported = "legacy_snippets_imported"

// LegacySnippet is one row of the retired SQL snippets table, kept as
// plain values: import-time parsing decides what survives.
//brandid-lint:ignore(BD003) retired SQL row kept as plain values; import-time parsing decides what survives
type LegacySnippet struct {
	ID        string
	Owner     string
	Body      string
	Quick     int
	CreatedAt int64
}

// ReadLegacySnippets reads the whole retired table oldest-first (the
// engine list is oldest-first too; importing in order preserves the
// legacy ordering semantics).
func ReadLegacySnippets(ctx context.Context, db *sql.DB) ([]LegacySnippet, error) {
	return listRows(ctx, db, "read legacy snippets", `
		SELECT id, owner, body, quick, created_at
		FROM snippets
		ORDER BY created_at ASC, rowid ASC
	`, nil, func(row rowScanner) (LegacySnippet, error) {
		var s LegacySnippet
		return s, row.Scan(&s.ID, &s.Owner, &s.Body, &s.Quick, &s.CreatedAt)
	})
}

// ToSnippet validates one legacy row into a domain.Snippet. A bad id
// demotes to a freshly minted one; a row whose owner does not parse is
// skipped (the caller logs it) — a corrupt legacy row must not brick a
// boot.
func (s LegacySnippet) ToSnippet() (domain.Snippet, bool) {
	owner, err := domain.ParseExtension(s.Owner)
	if err != nil {
		return domain.Snippet{}, false
	}
	id, err := domain.ParseSnippetID(s.ID)
	if err != nil {
		id = domain.GenerateSnippetID()
	}
	return domain.Snippet{
		ID:        id,
		Owner:     owner,
		Body:      s.Body,
		Quick:     s.Quick == 1,
		CreatedAt: time.Unix(s.CreatedAt, 0),
	}, true
}
