package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// Legacy-transcripts seam (2026-10-15): the call-transcript truth moved
// to the metaengine read models (readmodels.db). The SQL
// call_transcripts table stays untouched in webphone.db (the migration
// ladder never rewrites shipped shape), so an upgraded deployment's
// transcript history rides this ONE transitional read at boot: the
// composition root imports the rows into the engine store once and
// stamps the app_flags row. This file exists to be deleted when the
// transitional window closes.

// FlagLegacyTranscriptsImported marks the one-time legacy-transcripts
// import as complete (value carries the import date for operators).
const FlagLegacyTranscriptsImported = "legacy_transcripts_imported"

// LegacyTranscript is one row of the retired SQL call_transcripts
// table, kept as plain values: import-time parsing decides what
// survives.
// brandid-lint:ignore(BD003) retired SQL row; CallID is the opaque island token kept verbatim (strong-id wire ruling)
type LegacyTranscript struct {
	Owner     string
	CallID    string
	Direction string
	Remote    string
	StartedAt int64
	Text      string
	CreatedAt int64
}

// ReadLegacyTranscripts reads the whole retired table oldest-first (the
// engine keeps newest-first lists; importing in order preserves the
// legacy ordering semantics).
func ReadLegacyTranscripts(ctx context.Context, db *sql.DB) ([]LegacyTranscript, error) {
	return listRows(ctx, db, "read legacy transcripts", `
		SELECT owner, call_id, direction, remote, started_at, text, created_at
		FROM call_transcripts
		ORDER BY created_at ASC, rowid ASC
	`, nil, func(row rowScanner) (LegacyTranscript, error) {
		var t LegacyTranscript
		return t, row.Scan(&t.Owner, &t.CallID, &t.Direction, &t.Remote, &t.StartedAt, &t.Text, &t.CreatedAt)
	})
}

// ToSegment validates one legacy row into a domain.CallTranscriptSegment.
// A row whose owner does not parse, or whose direction is not a known
// enum value, is skipped (the caller logs it) — a corrupt legacy row
// must not brick a boot.
func (t LegacyTranscript) ToSegment() (domain.CallTranscriptSegment, bool) {
	owner, err := domain.ParseExtension(t.Owner)
	if err != nil {
		return domain.CallTranscriptSegment{}, false
	}
	direction := domain.Direction(t.Direction)
	if direction != domain.DirectionInbound && direction != domain.DirectionOutbound {
		return domain.CallTranscriptSegment{}, false
	}
	return domain.CallTranscriptSegment{
		Owner:     owner,
		CallID:    t.CallID,
		Direction: direction,
		Remote:    t.Remote,
		StartedAt: time.Unix(t.StartedAt, 0),
		Text:      t.Text,
	}, true
}
