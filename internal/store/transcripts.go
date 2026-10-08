package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// Transcripts persists per-extension live-call transcript segments.
type Transcripts struct {
	db *sql.DB
}

// TranscriptSegmentsMaxPerExtension bounds one extension's transcript
// history (≈5.5 h of continuous speech at the 4 s segment cadence).
// The cap is the durability story like ContactsMaxPerExtension: Recent
// renders the newest window, so trimming the oldest rows costs the
// deep past, never the current view.
const TranscriptSegmentsMaxPerExtension = 5000

// TranscriptSegmentsMaxPerCall bounds ONE call's transcript (≈100 min of
// continuous speech at the 4 s cadence): a single marathon call must not
// eat the owner's whole budget before the other calls' history. Same trim
// shape as the owner cap — the interaction (3+ long calls trimming each
// other's oldest rows) is intended, the per-call floor keeps every call
// reassemblable.
const TranscriptSegmentsMaxPerCall = 1500

// NewTranscripts builds the transcript store.
func NewTranscripts(db *sql.DB) *Transcripts { return &Transcripts{db: db} }

// Append stores one segment and trims past both caps: the owner's
// oldest rows past the per-extension cap, then the call's oldest past
// the per-call cap. The trims are scalar-bounded range deletes (the
// keep-set's MIN id via the indexed top-N) — a NOT IN keep-list would
// materialize the whole cap per append. Best-effort by design: a failed
// DELETE must not fail the write it accompanies (the cap re-arms on the
// next append).
func (s *Transcripts) Append(ctx context.Context, seg domain.CallTranscriptSegment) error {
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO call_transcripts (owner, call_id, direction, remote, started_at, text, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, seg.Owner.String(), seg.CallID, string(seg.Direction), seg.Remote,
		seg.StartedAt.Unix(), seg.Text, time.Now().Unix()); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.transcript_append", "append transcript")
	}
	if _, err := s.db.ExecContext(ctx, `
		DELETE FROM call_transcripts
		WHERE owner = ? AND id < (
			SELECT MIN(id) FROM (
				SELECT id FROM call_transcripts WHERE owner = ? ORDER BY id DESC LIMIT ?
			)
		)
	`, seg.Owner.String(), seg.Owner.String(), TranscriptSegmentsMaxPerExtension); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.transcript_trim", "trim transcripts")
	}
	if _, err := s.db.ExecContext(ctx, `
		DELETE FROM call_transcripts
		WHERE owner = ? AND call_id = ? AND id < (
			SELECT MIN(id) FROM (
				SELECT id FROM call_transcripts
				WHERE owner = ? AND call_id = ? ORDER BY id DESC LIMIT ?
			)
		)
	`, seg.Owner.String(), seg.CallID, seg.Owner.String(), seg.CallID, TranscriptSegmentsMaxPerCall); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.transcript_trim", "trim transcripts")
	}
	return nil
}

// Recent reassembles the owner's newest transcribed calls: segments are
// fetched newest-first, grouped by call (newest call first, its lines
// in spoken order), and the group list is capped at maxCalls.
func (s *Transcripts) Recent(ctx context.Context, owner domain.Extension, maxCalls int) ([]domain.CallTranscript, error) {
	rows, err := listRows(ctx, s.db, "list transcripts", `
		SELECT owner, call_id, direction, remote, started_at, text
		FROM call_transcripts WHERE owner = ?
		ORDER BY started_at DESC, rowid DESC
		LIMIT ?
	`, []any{owner.String(), transcriptFetchWindow}, scanTranscriptRow)
	if err != nil {
		return nil, err
	}
	groups := make([]domain.CallTranscript, 0, maxCalls)
	byCall := make(map[string]int, maxCalls)
	for _, seg := range rows {
		idx, seen := byCall[seg.CallID]
		if !seen {
			if len(groups) >= maxCalls {
				// Segments of already-open calls keep landing (the
				// oldest included call must reassemble whole); only NEW
				// groups are refused.
				continue
			}
			idx = len(groups)
			byCall[seg.CallID] = idx
			groups = append(groups, domain.CallTranscript{
				Owner:     seg.Owner,
				CallID:    seg.CallID,
				Direction: seg.Direction,
				Remote:    seg.Remote,
				StartedAt: seg.StartedAt,
			})
		}
		// Newest-first fetch, spoken-order lines: prepend within the group.
		group := &groups[idx]
		group.Lines = append([]string{seg.Text}, group.Lines...)
		if group.StartedAt.Before(seg.StartedAt) {
			group.StartedAt = seg.StartedAt
		}
	}
	return groups, nil
}

// transcriptFetchWindow is the segment read window behind Recent: wide
// enough that the maxCalls newest calls reassemble whole even when one
// call alone produced hundreds of segments.
const transcriptFetchWindow = 2000

// DeleteCall erases one call's transcript for one owner (the end-user
// leg of the retention story): every segment of the call goes in one
// statement. Idempotent — deleting an unknown call affects zero rows,
// which is the same success as deleting an already-deleted one.
func (s *Transcripts) DeleteCall(ctx context.Context, owner domain.Extension, callID string) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM call_transcripts WHERE owner = ? AND call_id = ?
	`, owner.String(), callID)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "store.transcript_delete", "delete call transcript")
	}
	n, _ := res.RowsAffected() //nolint:erraudit // count is informational
	return n, nil
}

func scanTranscriptRow(row rowScanner) (domain.CallTranscriptSegment, error) {
	var (
		owner, callID, direction, remote, text string
		started                                int64
	)
	if err := row.Scan(&owner, &callID, &direction, &remote, &started, &text); err != nil {
		return domain.CallTranscriptSegment{}, errorfamily.WrapInfrastructuref(err, "store.transcript_scan", "scan transcript")
	}
	return domain.CallTranscriptSegment{
		Owner:     domain.MustParseExtension(owner),
		CallID:    callID,
		Direction: domain.Direction(direction),
		Remote:    remote,
		StartedAt: time.Unix(started, 0),
		Text:      text,
	}, nil
}
