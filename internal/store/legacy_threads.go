package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// Legacy-threads seam (SUPERB A35): the thread/message/attachment truth
// moved to the journal streams + system.db projections. The SQL
// threads/messages/attachments tables stay untouched in webphone.db
// (the migration ladder never rewrites shipped shape), so an upgraded
// deployment's conversation history rides this ONE transitional read at
// boot: the composition root imports the rows into the journal once and
// stamps the app_flags row. This file exists to be deleted when the
// transitional window closes.

// LegacyThread is one row of the retired SQL threads table, kept as
// plain values: import-time parsing decides what survives.
// brandid-lint:ignore(BD003) retired SQL row kept as plain values; import-time parsing decides what survives
type LegacyThread struct {
	ID           string
	Owner        string
	Network      string
	Remote       string
	LastActivity int64 // unix seconds, the SQL column's shape
	Unread       int
	Pinned       bool
	Archived     bool
	Muted        bool
}

// LegacyMessage is one row of the retired SQL messages table.
// brandid-lint:ignore(BD003) retired SQL row kept as plain values; import-time parsing decides what survives
type LegacyMessage struct {
	ID            string
	ThreadID      string
	Owner         string
	Network       string
	Remote        string
	Direction     string
	Channel       string
	Body          string
	Status        string
	ProviderRef   string
	FailureKind   string
	FailureDetail string
	CreatedAt     int64 // unix seconds, the SQL column's shape
}

// LegacyAttachment is one row of the retired SQL attachments table
// (owner rides the owning message row; the table never carried it).
// brandid-lint:ignore(BD003) retired SQL row kept as plain values; import-time parsing decides what survives
type LegacyAttachment struct {
	ID        string
	MessageID string
	Name      string
	MimeType  string
	SizeBytes int64
	Path      string
}

// ReadLegacyThreads reads every retired thread row oldest-activity
// first (import streams build oldest-first; the rowid tiebreak is the
// legacy insertion order).
func ReadLegacyThreads(ctx context.Context, db *sql.DB) ([]LegacyThread, error) {
	return listRows(ctx, db, "read legacy threads", `
		SELECT id, owner, network, remote, last_activity_at, unread, pinned, archived, muted
		FROM threads
		ORDER BY last_activity_at ASC, rowid ASC
	`, nil, func(row rowScanner) (LegacyThread, error) {
		var t LegacyThread
		var unread int
		var pinned, archived, muted int
		if err := row.Scan(&t.ID, &t.Owner, &t.Network, &t.Remote, &t.LastActivity, &unread,
			&pinned, &archived, &muted); err != nil {
			return t, err
		}
		t.Unread, t.Pinned, t.Archived, t.Muted = unread, pinned == 1, archived == 1, muted == 1
		return t, nil
	})
}

// ReadLegacyMessages reads every retired message row in the legacy
// list's own order — created_at ASC, rowid ASC — so the importer's
// journal sequence IS the legacy transcript order (timestamp fidelity:
// a clock-skewed rowid order must not win).
func ReadLegacyMessages(ctx context.Context, db *sql.DB) ([]LegacyMessage, error) {
	return listRows(ctx, db, "read legacy messages", `
		SELECT id, thread_id, owner, network, remote, direction, channel, body,
		       status, provider_ref, failure_kind, failure_detail, created_at
		FROM messages
		ORDER BY created_at ASC, rowid ASC
	`, nil, func(row rowScanner) (LegacyMessage, error) {
		var m LegacyMessage
		return m, row.Scan(&m.ID, &m.ThreadID, &m.Owner, &m.Network, &m.Remote, &m.Direction, &m.Channel,
			&m.Body, &m.Status, &m.ProviderRef, &m.FailureKind, &m.FailureDetail, &m.CreatedAt)
	})
}

// ReadLegacyAttachments reads every retired attachment row in insertion
// order (the owning message's row gives the owner; per-message name
// order is the reader's sort, preserved by the projection fan-in).
func ReadLegacyAttachments(ctx context.Context, db *sql.DB) ([]LegacyAttachment, error) {
	return listRows(ctx, db, "read legacy attachments", `
		SELECT id, message_id, name, mime_type, size_bytes, path
		FROM attachments
		ORDER BY rowid ASC
	`, nil, func(row rowScanner) (LegacyAttachment, error) {
		var a LegacyAttachment
		return a, row.Scan(&a.ID, &a.MessageID, &a.Name, &a.MimeType, &a.SizeBytes, &a.Path)
	})
}

// ToThread validates one legacy thread row. A row whose ids, owner or
// address do not parse is skipped (the caller logs it) — a corrupt
// legacy row must not brick a boot, and a foreign-network row would
// poison the stream key.
func (t LegacyThread) ToThread() (domain.Thread, bool) {
	id, err := domain.ParseThreadID(t.ID)
	if err != nil {
		return domain.Thread{}, false
	}
	owner, err := domain.ParseExtension(t.Owner)
	if err != nil {
		return domain.Thread{}, false
	}
	network, err := domain.ParseNetwork(t.Network)
	if err != nil {
		return domain.Thread{}, false
	}
	remote, err := domain.ParseAddress(network, t.Remote)
	if err != nil {
		return domain.Thread{}, false
	}
	return domain.Thread{
		ID:             id,
		Owner:          owner,
		Remote:         remote,
		LastActivityAt: time.Unix(t.LastActivity, 0),
		Unread:         t.Unread,
		Pinned:         t.Pinned,
		Archived:       t.Archived,
		Muted:          t.Muted,
	}, true
}

// ToMessage validates one legacy message row against the thread it
// claims (identity disagreements skip: the stream key and the fact must
// agree or the projection would split-brain).
func (m LegacyMessage) ToMessage(thread domain.Thread) (domain.Message, bool) {
	id, err := domain.ParseMessageID(m.ID)
	if err != nil {
		return domain.Message{}, false
	}
	if m.ThreadID != thread.ID.String() {
		return domain.Message{}, false
	}
	if m.Owner != thread.Owner.String() || m.Network != string(thread.Remote.Network()) ||
		m.Remote != thread.Remote.String() {
		return domain.Message{}, false
	}
	direction, err := domain.ParseDirection(m.Direction)
	if err != nil {
		return domain.Message{}, false
	}
	channel, err := domain.ParseChannel(m.Channel)
	if err != nil {
		return domain.Message{}, false
	}
	status, err := domain.ParseOutboundStatus(m.Status)
	if err != nil {
		return domain.Message{}, false
	}
	return domain.Message{
		ID:            id,
		ThreadID:      thread.ID,
		Owner:         thread.Owner,
		Remote:        thread.Remote,
		Direction:     direction,
		Channel:       channel,
		Body:          m.Body,
		Status:        status,
		ProviderRef:   m.ProviderRef,
		FailureKind:   domain.FailureKind(m.FailureKind),
		FailureDetail: m.FailureDetail,
		CreatedAt:     time.Unix(m.CreatedAt, 0),
	}, true
}

// ToAttachment validates one legacy attachment row; the owner and
// message link arrive from the owning validated message.
func (a LegacyAttachment) ToAttachment(message domain.Message) (domain.Attachment, bool) {
	id, err := domain.ParseAttachmentID(a.ID)
	if err != nil {
		return domain.Attachment{}, false
	}
	if a.MessageID != message.ID.String() {
		return domain.Attachment{}, false
	}
	return domain.Attachment{
		ID:        id,
		MessageID: message.ID,
		Name:      a.Name,
		MimeType:  a.MimeType,
		SizeBytes: a.SizeBytes,
		Path:      a.Path,
	}, true
}
