package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// ErrNotFound is returned when a row the caller asked for does not exist.
var ErrNotFound = errors.New("not found")

// Messages is the thread/message persistence.
type Messages struct {
	db *sql.DB
}

// NewMessages builds the message store.
func NewMessages(db *sql.DB) *Messages { return &Messages{db: db} }

// AppendMessage inserts a message inside one transaction: it upserts the
// owner/remote thread, inserts the message and its attachments, and updates
// the thread's activity timestamp (and unread count for inbound).
func (s *Messages) AppendMessage(ctx context.Context, msg domain.Message) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }() //nolint:erraudit // best-effort write; the response is already committed

	threadID := msg.ThreadID.String()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO threads (id, owner, remote, last_activity_at, unread)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(owner, remote) DO UPDATE SET
			last_activity_at = excluded.last_activity_at,
			unread = threads.unread + excluded.unread
	`, threadID, msg.Owner.String(), msg.Remote.String(), msg.CreatedAt.Unix(),
		incrementIf(domain.DirectionInbound, msg.Direction)); err != nil {
		return fmt.Errorf("upsert thread %s for %s/%s: %w", threadID, msg.Owner, msg.Remote, err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO messages (id, thread_id, owner, remote, direction, channel, body, status, provider_ref, failure_kind, failure_detail, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, msg.ID.String(), threadID, msg.Owner.String(), msg.Remote.String(),
		string(msg.Direction), string(msg.Channel), msg.Body,
		string(msg.Status), msg.ProviderRef, msg.FailureKind, msg.FailureDetail,
		msg.CreatedAt.Unix()); err != nil {
		return fmt.Errorf("insert message %s (thread %s): %w", msg.ID, threadID, err)
	}

	for _, att := range msg.Attachments {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO attachments (id, message_id, name, mime_type, size_bytes, path)
			VALUES (?, ?, ?, ?, ?, ?)
		`, att.ID.String(), msg.ID.String(), att.Name, att.MimeType, att.SizeBytes, att.Path); err != nil {
			return fmt.Errorf("insert attachment %s of message %s: %w", att.ID, msg.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func incrementIf(want, got domain.Direction) int {
	if want == got {
		return 1
	}
	return 0
}

// UpdateOutboundStatus advances an outbound message's delivery status,
// provider reference and failure story. A delivered verdict CLEARS any
// earlier failure (kind and detail revert to empty); a failed verdict
// records kind + detail for the bubble's disclosure and retry decision.
func (s *Messages) UpdateOutboundStatus(
	ctx context.Context, id domain.MessageID, status domain.OutboundStatus, providerRef string,
	failureKind, failureDetail string,
) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE messages
		SET status = ?, provider_ref = ?, failure_kind = ?, failure_detail = ?
		WHERE id = ? AND direction = ?
	`, string(status), providerRef, failureKind, failureDetail,
		id.String(), string(domain.DirectionOutbound))
	if err != nil {
		return fmt.Errorf("update status of message %s: %w", id, err)
	}
	return updatedOrNotFound(res)
}

// MessageByProviderRef resolves an outbound message by its gateway
// correlation id — the delivery-status webhook's lookup path.
func (s *Messages) MessageByProviderRef(ctx context.Context, ref string) (domain.Message, error) {
	if ref == "" {
		return domain.Message{}, ErrNotFound
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT id, thread_id, owner, remote, direction, channel, body, status, provider_ref, failure_kind, failure_detail, created_at
		FROM messages WHERE provider_ref = ? AND direction = ?
	`, ref, string(domain.DirectionOutbound))
	msg, err := scanMessage(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Message{}, ErrNotFound
	}
	if err != nil {
		return domain.Message{}, err
	}
	return msg, nil
}

// ThreadSummary is a thread row as shown in the thread list: the thread
// plus a preview of its last message.
type ThreadSummary struct {
	Thread        domain.Thread
	LastBody      string
	LastDirection domain.Direction
	LastChannel   domain.Channel
}

// ListThreads returns the owner's threads, most recently active first.
func (s *Messages) ListThreads(ctx context.Context, owner domain.Extension) ([]ThreadSummary, error) {
	return listRows(ctx, s.db, "list threads", `
		SELECT t.id, t.owner, t.remote, t.last_activity_at, t.unread,
		       m.body, m.direction, m.channel
		FROM threads t
		LEFT JOIN messages m ON m.id = (
			SELECT id FROM messages WHERE thread_id = t.id ORDER BY created_at DESC, rowid DESC LIMIT 1
		)
		WHERE t.owner = ?
		ORDER BY t.last_activity_at DESC
	`, []any{owner.String()}, scanThreadSummary)
}

// SearchThreads returns the owner's threads whose remote number or ANY
// message body matches the query (SQLite LIKE: ASCII case-insensitive),
// most recently active first. LIKE metacharacters in the query are
// escaped, so a search for "50%" finds "50%", not "50" followed by
// anything.
func (s *Messages) SearchThreads(ctx context.Context, owner domain.Extension, query string) ([]ThreadSummary, error) {
	pattern := "%" + likeEscape(query) + "%"
	return listRows(ctx, s.db, "search threads", `
		SELECT t.id, t.owner, t.remote, t.last_activity_at, t.unread,
		       m.body, m.direction, m.channel
		FROM threads t
		LEFT JOIN messages m ON m.id = (
			SELECT id FROM messages WHERE thread_id = t.id ORDER BY created_at DESC, rowid DESC LIMIT 1
		)
		WHERE t.owner = ?
		  AND (t.remote LIKE ? ESCAPE '\' OR EXISTS (
			SELECT 1 FROM messages sm
			WHERE sm.thread_id = t.id AND sm.owner = t.owner AND sm.body LIKE ? ESCAPE '\'
		  ))
		ORDER BY t.last_activity_at DESC
	`, []any{owner.String(), pattern, pattern}, scanThreadSummary)
}

// likeEscape escapes SQL LIKE metacharacters so the query matches them
// literally (paired with ESCAPE '\' in the statement).
func likeEscape(query string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(query)
}

func scanThreadSummary(row rowScanner) (ThreadSummary, error) {
	var (
		id, owner, remote         string
		lastActivity              int64
		unread                    int
		lastBody, lastDir, lastCh sql.NullString
	)
	if err := row.Scan(&id, &owner, &remote, &lastActivity, &unread, &lastBody, &lastDir, &lastCh); err != nil {
		return ThreadSummary{}, fmt.Errorf("scan thread row: %w", err)
	}

	sum := ThreadSummary{
		Thread: domain.Thread{
			ID:             domain.MustThreadID(id),
			Owner:          domain.MustParseExtension(owner),
			Remote:         domain.MustParsePhone(remote),
			LastActivityAt: time.Unix(lastActivity, 0),
			Unread:         unread,
		},
	}
	if lastBody.Valid {
		sum.LastBody = lastBody.String
		sum.LastDirection = domain.Direction(lastDir.String)
		sum.LastChannel = domain.Channel(lastCh.String)
	}

	return sum, nil
}

// ListMessages returns the messages of one thread, oldest first.
func (s *Messages) ListMessages(
	ctx context.Context, owner domain.Extension, threadID domain.ThreadID, limit int,
) ([]domain.Message, error) {
	msgs, err := listRows(ctx, s.db, fmt.Sprintf("list messages of thread %s", threadID), `
		SELECT id, thread_id, owner, remote, direction, channel, body, status, provider_ref, failure_kind, failure_detail, created_at
		FROM messages
		WHERE owner = ? AND thread_id = ?
		ORDER BY created_at DESC, rowid DESC
		LIMIT ?
	`, []any{owner.String(), threadID.String(), limit}, scanMessage)
	if err != nil {
		return nil, err
	}

	if err := s.attachAttachments(ctx, msgs); err != nil {
		return nil, err
	}

	// Query was newest-first for the LIMIT; the UI wants a chat transcript,
	// oldest at the top.
	slices.Reverse(msgs)

	return msgs, nil
}

func scanMessage(row rowScanner) (domain.Message, error) {
	var (
		id, threadID, owner, remote, direction, channel string
		body, status, providerRef                       string
		failureKind, failureDetail                      string
		createdAt                                       int64
	)
	if err := row.Scan(&id, &threadID, &owner, &remote, &direction, &channel,
		&body, &status, &providerRef, &failureKind, &failureDetail, &createdAt); err != nil {
		return domain.Message{}, fmt.Errorf("scan message %s of thread %s: %w", id, threadID, err)
	}
	return domain.Message{
		ID:            domain.MustMessageID(id),
		ThreadID:      domain.MustThreadID(threadID),
		Owner:         domain.MustParseExtension(owner),
		Remote:        domain.MustParsePhone(remote),
		Direction:     domain.Direction(direction),
		Channel:       domain.Channel(channel),
		Body:          body,
		Status:        domain.OutboundStatus(status),
		ProviderRef:   providerRef,
		FailureKind:   failureKind,
		FailureDetail: failureDetail,
		CreatedAt:     time.Unix(createdAt, 0),
	}, nil
}

func (s *Messages) attachAttachments(ctx context.Context, msgs []domain.Message) error {
	if len(msgs) == 0 {
		return nil
	}
	for i, msg := range msgs {
		rows, err := s.db.QueryContext(ctx, `
			SELECT id, message_id, name, mime_type, size_bytes, path
			FROM attachments WHERE message_id = ? ORDER BY name
		`, msg.ID.String())
		if err != nil {
			return fmt.Errorf("list attachments: %w", err)
		}
		for rows.Next() {
			var (
				id, messageID, name, mime, path string
				size                            int64
			)
			if err := rows.Scan(&id, &messageID, &name, &mime, &size, &path); err != nil {
				_ = rows.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
				return fmt.Errorf("scan attachment %s of message %s: %w", id, messageID, err)
			}
			msgs[i].Attachments = append(msgs[i].Attachments, domain.Attachment{
				ID:        domain.MustAttachmentID(id),
				MessageID: domain.MustMessageID(messageID),
				Name:      name,
				MimeType:  mime,
				SizeBytes: size,
				Path:      path,
			})
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
			return fmt.Errorf("attachments rows: %w", err)
		}
		_ = rows.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
	}
	return nil
}

// FindThread returns the owner's thread for a remote number, creating it
// (persisted) when absent. Callers use it to resolve where a message goes.
func (s *Messages) FindThread(
	ctx context.Context, owner domain.Extension, remote domain.Phone, now time.Time,
) (domain.ThreadID, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM threads WHERE owner = ? AND remote = ?
	`, owner.String(), remote.String()).Scan(&id)
	if err == nil {
		return domain.MustThreadID(id), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domain.ThreadID{}, fmt.Errorf("find thread for %s/%s: %w", owner, remote, err)
	}

	newID := domain.GenerateThreadID()
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO threads (id, owner, remote, last_activity_at, unread)
		VALUES (?, ?, ?, ?, 0)
		ON CONFLICT(owner, remote) DO UPDATE SET last_activity_at = last_activity_at
	`, newID.String(), owner.String(), remote.String(), now.Unix()); err != nil {
		return domain.ThreadID{}, fmt.Errorf("insert thread %s: %w", newID, err)
	}

	// The ON CONFLICT above is a no-op update so a concurrent insert cannot
	// fail the call; re-read to return the winner's id.
	if err := s.db.QueryRowContext(ctx, `
		SELECT id FROM threads WHERE owner = ? AND remote = ?
	`, owner.String(), remote.String()).Scan(&id); err != nil {
		return domain.ThreadID{}, fmt.Errorf("re-read thread for %s/%s: %w", owner, remote, err)
	}

	return domain.MustThreadID(id), nil
}

// GetThread fetches one thread by id, scoped to the owner.
func (s *Messages) GetThread(
	ctx context.Context, owner domain.Extension, id domain.ThreadID,
) (domain.Thread, error) {
	var (
		ownerStr, remoteStr string
		lastActivity        int64
		unread              int
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT owner, remote, last_activity_at, unread FROM threads WHERE id = ? AND owner = ?
	`, id.String(), owner.String()).Scan(&ownerStr, &remoteStr, &lastActivity, &unread)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Thread{}, ErrNotFound
	}
	if err != nil {
		return domain.Thread{}, fmt.Errorf("get thread %s: %w", id, err)
	}
	return domain.Thread{
		ID:             id,
		Owner:          domain.MustParseExtension(ownerStr),
		Remote:         domain.MustParsePhone(remoteStr),
		LastActivityAt: time.Unix(lastActivity, 0),
		Unread:         unread,
	}, nil
}

// MarkThreadRead zeroes the unread counter.
func (s *Messages) MarkThreadRead(ctx context.Context, owner domain.Extension, id domain.ThreadID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE threads SET unread = 0 WHERE id = ? AND owner = ?
	`, id.String(), owner.String())
	if err != nil {
		return fmt.Errorf("mark thread %s read: %w", id, err)
	}
	return nil
}

// AttachmentByID resolves one attachment scoped to the owner (the
// attachment's message must belong to her).
func (s *Messages) AttachmentByID(
	ctx context.Context, owner domain.Extension, id domain.AttachmentID,
) (domain.Attachment, error) {
	var (
		attachmentID, messageID, name, mime, path string
		size                                      int64
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT a.id, a.message_id, a.name, a.mime_type, a.size_bytes, a.path
		FROM attachments a
		JOIN messages m ON m.id = a.message_id
		WHERE a.id = ? AND m.owner = ?
	`, id.String(), owner.String()).Scan(&attachmentID, &messageID, &name, &mime, &size, &path)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Attachment{}, ErrNotFound
	}
	if err != nil {
		return domain.Attachment{}, fmt.Errorf("get attachment %s (row %s of message %s): %w", id, attachmentID, messageID, err)
	}
	return domain.Attachment{
		ID:        domain.MustAttachmentID(attachmentID),
		MessageID: domain.MustMessageID(messageID),
		Name:      name,
		MimeType:  mime,
		SizeBytes: size,
		Path:      path,
	}, nil
}

// ListMessagesPage returns one page of a thread's messages, oldest
// first. Page 0 is the newest window; hasMore reports whether older
// pages exist beyond it.
func (s *Messages) ListMessagesPage(
	ctx context.Context, owner domain.Extension, threadID domain.ThreadID, page, limit int,
) ([]domain.Message, bool, error) {
	if page < 0 {
		page = 0
	}
	msgs, err := listRows(ctx, s.db, fmt.Sprintf("list messages of thread %s", threadID), `
		SELECT id, thread_id, owner, remote, direction, channel, body, status, provider_ref, failure_kind, failure_detail, created_at
		FROM messages
		WHERE owner = ? AND thread_id = ?
		ORDER BY created_at DESC, rowid DESC
		LIMIT ? OFFSET ?
	`, []any{owner.String(), threadID.String(), limit + 1, page * limit}, scanMessage)
	if err != nil {
		return nil, false, fmt.Errorf("list messages page %d of thread %s: %w", page, threadID, err)
	}

	hasMore := len(msgs) > limit
	if hasMore {
		msgs = msgs[:limit]
	}
	if err := s.attachAttachments(ctx, msgs); err != nil {
		return nil, false, fmt.Errorf("attach attachments to thread %s messages: %w", threadID, err)
	}

	// Query was newest-first for the LIMIT; the UI wants a chat transcript,
	// oldest at the top.
	slices.Reverse(msgs)

	return msgs, hasMore, nil
}
