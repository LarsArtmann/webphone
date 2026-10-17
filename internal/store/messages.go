package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// ErrNotFound is returned when a row the caller asked for does not exist.
var ErrNotFound = errors.New("not found") //nolint:erraudit // sentinel: absence is not an error family (plan guardrail #4); classified via init registration

// Messages is the thread/message persistence.
type Messages struct {
	db *sql.DB
}

// NewMessages builds the message store.
func NewMessages(db *sql.DB) *Messages { return &Messages{db: db} }

// AppendMessage inserts a message inside one transaction: it loads (or
// creates) the owner/remote thread, folds the message into it via
// domain.AppendToThread — the one home of the append invariants — and
// writes the advanced row back, then inserts the message and its
// attachments.
func (s *Messages) AppendMessage(ctx context.Context, msg domain.Message) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.tx_begin", "begin")
	}
	defer func() { _ = tx.Rollback() }() //nolint:erraudit // best-effort write; the response is already committed

	var existing domain.Thread
	var (
		rowID        string
		lastActivity int64
		unread       int
		archived     int
	)
	err = tx.QueryRowContext(ctx, `
		SELECT id, last_activity_at, unread, archived FROM threads WHERE owner = ? AND remote = ?
	`, msg.Owner.String(), msg.Remote.String()).Scan(&rowID, &lastActivity, &unread, &archived)
	switch {
	case err == nil:
		existing = domain.Thread{
			ID:             domain.MustThreadID(rowID),
			Owner:          msg.Owner,
			Remote:         msg.Remote,
			LastActivityAt: time.Unix(lastActivity, 0),
			Unread:         unread,
			Archived:       archived != 0,
		}
	case errors.Is(err, sql.ErrNoRows):
		// Creation: the zero thread adopts the message's identity.
	default:
		return errorfamily.WrapInfrastructuref(err, "store.thread_lookup", "load thread %s/%s", msg.Owner, msg.Remote)
	}

	thread := domain.AppendToThread(existing, msg)
	if existing.ID == (domain.ThreadID{}) {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO threads (id, owner, remote, last_activity_at, unread, archived)
			VALUES (?, ?, ?, ?, ?, ?)
		`, thread.ID.String(), msg.Owner.String(), msg.Remote.String(),
			thread.LastActivityAt.Unix(), thread.Unread, boolInt(thread.Archived)); err != nil {
			return errorfamily.WrapInfrastructuref(err, "store.thread_insert", "insert thread %s for %s/%s", thread.ID, msg.Owner, msg.Remote)
		}
	} else if _, err := tx.ExecContext(ctx, `
		UPDATE threads SET last_activity_at = ?, unread = ?, archived = ? WHERE id = ? AND owner = ?
	`, thread.LastActivityAt.Unix(), thread.Unread, boolInt(thread.Archived),
		thread.ID.String(), msg.Owner.String()); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.thread_update", "update thread %s for %s", thread.ID, msg.Owner)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO messages (id, thread_id, owner, remote, direction, channel, body, status, provider_ref, failure_kind, failure_detail, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, msg.ID.String(), thread.ID.String(), msg.Owner.String(), msg.Remote.String(),
		string(msg.Direction), string(msg.Channel), msg.Body,
		string(msg.Status), msg.ProviderRef, string(msg.FailureKind), msg.FailureDetail,
		msg.CreatedAt.Unix()); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.message_insert", "insert message %s (thread %s)", msg.ID, thread.ID)
	}

	for _, att := range msg.Attachments {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO attachments (id, message_id, name, mime_type, size_bytes, path)
			VALUES (?, ?, ?, ?, ?, ?)
		`, att.ID.String(), msg.ID.String(), att.Name, att.MimeType, att.SizeBytes, att.Path); err != nil {
			return errorfamily.WrapInfrastructuref(err, "store.attachment_insert", "insert attachment %s of message %s", att.ID, msg.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.tx_commit", "commit")
	}

	return nil
}

// UpdateOutboundStatus advances an outbound message's delivery status,
// provider reference and failure story. A delivered verdict CLEARS any
// earlier failure (kind and detail revert to empty); a failed verdict
// records kind + detail for the bubble's disclosure and retry decision.
func (s *Messages) UpdateOutboundStatus(
	ctx context.Context, id domain.MessageID, status domain.OutboundStatus, providerRef string, //nolint:branching-flow // provider correlation id + failure vocabulary (documented at FailureKind); a ProviderRef type would ripple into webhook DTOs, the partial unique index and the golden-pinned rows
	failureKind, failureDetail string,
) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE messages
		SET status = ?, provider_ref = ?, failure_kind = ?, failure_detail = ?
		WHERE id = ? AND direction = ?
	`, string(status), providerRef, failureKind, failureDetail,
		id.String(), string(domain.DirectionOutbound))
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.outbound_status", "update status of message %s", id)
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
	Thread         domain.Thread
	LastBody       string
	LastDirection  domain.Direction
	LastChannel    domain.Channel
	LastAttachment bool // the last message carries at least one attachment
}

// ThreadFlag names one organization axis of a thread (M22). It is the
// persistence-layer spelling of the three bools on domain.Thread; the
// handlers map their action path onto it.
type ThreadFlag string

const (
	FlagPinned   ThreadFlag = "pinned"
	FlagArchived ThreadFlag = "archived"
	FlagMuted    ThreadFlag = "muted"
)

// threadFlagColumns maps each flag to its column; one home for the
// toggle's UPDATE target.
var threadFlagColumns = map[ThreadFlag]string{
	FlagPinned:   "pinned",
	FlagArchived: "archived",
	FlagMuted:    "muted",
}

// SetThreadFlag sets or clears one organization flag on the owner's
// thread. A missing thread (or another owner's) is ErrNotFound.
func (s *Messages) SetThreadFlag(
	ctx context.Context, owner domain.Extension, id domain.ThreadID, flag ThreadFlag, on bool,
) error {
	column, ok := threadFlagColumns[flag]
	if !ok {
		return errorfamily.NewRejection("store.thread_flag", "unknown thread flag "+string(flag))
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE threads SET `+column+` = ? WHERE id = ? AND owner = ?
	`, boolInt(on), id.String(), owner.String())
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.thread_flag", "set %s = %v on thread %s", flag, on, id)
	}
	return updatedOrNotFound(res)
}

func boolInt(on bool) int {
	if on {
		return 1
	}
	return 0
}

// ListThreads returns the owner's active threads: pinned first, then
// most recently active. Archived threads are excluded (design D6) —
// ListArchivedThreads serves the archived view.
func (s *Messages) ListThreads(ctx context.Context, owner domain.Extension) ([]ThreadSummary, error) {
	return listRows(ctx, s.db, "list threads", threadSummaryQuery("t.archived = 0"),
		[]any{owner.String()}, scanThreadSummary)
}

// ListArchivedThreads returns the owner's archived threads, most
// recently active first (no pinned ordering — the archived view is a
// filing cabinet, not a dashboard).
func (s *Messages) ListArchivedThreads(ctx context.Context, owner domain.Extension) ([]ThreadSummary, error) {
	return listRows(ctx, s.db, "list archived threads", threadSummaryQuery("t.archived = 1"),
		[]any{owner.String()}, scanThreadSummary)
}

// CountArchived reports how many of the owner's threads are archived —
// the toggle link only renders when there is something to show.
func (s *Messages) CountArchived(ctx context.Context, owner domain.Extension) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM threads WHERE owner = ? AND archived = 1`, owner.String()).Scan(&count); err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "store.count_archived", "count archived threads of %s", owner)
	}
	return count, nil
}

// threadSummaryQuery builds the thread-list SELECT with the given
// archived filter. Pinned threads lead (design D5); the last-message
// join and the attachment preview EXISTS are shared by every variant.
func threadSummaryQuery(archivedFilter string) string {
	return `
		SELECT t.id, t.owner, t.remote, t.last_activity_at, t.unread,
		       t.pinned, t.archived, t.muted,
		       m.body, m.direction, m.channel,
		       EXISTS (SELECT 1 FROM attachments a WHERE a.message_id = m.id)
		FROM threads t
		LEFT JOIN messages m ON m.id = (
			SELECT id FROM messages WHERE thread_id = t.id ORDER BY created_at DESC, rowid DESC LIMIT 1
		)
		WHERE t.owner = ? AND ` + archivedFilter + `
		ORDER BY t.pinned DESC, t.last_activity_at DESC
	`
}

// SearchThreads returns the owner's active threads whose remote number
// or ANY message body matches the needle (SQLite LIKE: ASCII
// case-insensitive), pinned first then most recently active. Archived
// threads stay hidden (design D6 — search must not resurrect what
// archive hid). LIKE metacharacters in the needle are escaped, so a
// search for "50%" finds "50%", not "50" followed by anything.
func (s *Messages) SearchThreads(ctx context.Context, owner domain.Extension, needle string) ([]ThreadSummary, error) {
	pattern := "%" + likeEscape(needle) + "%"
	return listRows(ctx, s.db, "search threads", `
		SELECT t.id, t.owner, t.remote, t.last_activity_at, t.unread,
		       t.pinned, t.archived, t.muted,
		       m.body, m.direction, m.channel,
		       EXISTS (SELECT 1 FROM attachments a WHERE a.message_id = m.id)
		FROM threads t
		LEFT JOIN messages m ON m.id = (
			SELECT id FROM messages WHERE thread_id = t.id ORDER BY created_at DESC, rowid DESC LIMIT 1
		)
		WHERE t.owner = ? AND t.archived = 0
		  AND (t.remote LIKE ? ESCAPE '\' OR EXISTS (
			SELECT 1 FROM messages sm
			WHERE sm.thread_id = t.id AND sm.owner = t.owner AND sm.body LIKE ? ESCAPE '\'
		  ))
		ORDER BY t.pinned DESC, t.last_activity_at DESC
	`, []any{owner.String(), pattern, pattern}, scanThreadSummary)
}

// likeEscape escapes SQL LIKE metacharacters so the needle matches them
// literally (paired with ESCAPE '\' in the statement).
func likeEscape(needle string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(needle)
}

func scanThreadSummary(row rowScanner) (ThreadSummary, error) {
	var (
		id, owner, remote         string
		lastActivity              int64
		unread                    int
		pinned, archived, muted   int
		lastAttachment            int
		lastBody, lastDir, lastCh sql.NullString
	)
	if err := row.Scan(&id, &owner, &remote, &lastActivity, &unread,
		&pinned, &archived, &muted, &lastBody, &lastDir, &lastCh, &lastAttachment); err != nil {
		return ThreadSummary{}, errorfamily.WrapInfrastructuref(err, "store.thread_scan", "scan thread row") //nolint:erraudit // context_loss FP: lastDir is an OUT param, garbage exactly when the scan failed — no honest context to include
	}

	sum := ThreadSummary{
		Thread: domain.Thread{
			ID:             domain.MustThreadID(id),
			Owner:          domain.MustParseExtension(owner),
			Remote:         domain.MustParsePhone(remote),
			LastActivityAt: time.Unix(lastActivity, 0),
			Unread:         unread,
			Pinned:         pinned == 1,
			Archived:       archived == 1,
			Muted:          muted == 1,
		},
		LastAttachment: lastAttachment == 1,
	}
	if lastBody.Valid {
		sum.LastBody = lastBody.String
		sum.LastDirection = domain.MustDirection(lastDir.String)
		sum.LastChannel = domain.MustChannel(lastCh.String)
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
		return domain.Message{}, errorfamily.WrapInfrastructuref(err, "store.message_scan", "scan message %s of thread %s", id, threadID)
	}
	return domain.Message{
		ID:            domain.MustMessageID(id),
		ThreadID:      domain.MustThreadID(threadID),
		Owner:         domain.MustParseExtension(owner),
		Remote:        domain.MustParsePhone(remote),
		Direction:     domain.MustDirection(direction),
		Channel:       domain.MustChannel(channel),
		Body:          body,
		Status:        domain.MustOutboundStatus(status),
		ProviderRef:   providerRef,
		FailureKind:   domain.FailureKind(failureKind),
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
			return errorfamily.WrapInfrastructuref(err, "store.attachments_list", "list attachments")
		}
		for rows.Next() {
			var (
				id, messageID, name, mime, path string
				size                            int64
			)
			if err := rows.Scan(&id, &messageID, &name, &mime, &size, &path); err != nil {
				_ = rows.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
				return errorfamily.WrapInfrastructuref(err, "store.attachment_scan", "scan attachment %s of message %s", id, messageID)
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
			return errorfamily.WrapInfrastructuref(err, "store.attachments_rows", "attachments rows")
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
		return domain.ThreadID{}, errorfamily.WrapInfrastructuref(err, "store.thread_find", "find thread for %s/%s", owner, remote)
	}

	newID := domain.GenerateThreadID()
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO threads (id, owner, remote, last_activity_at, unread)
		VALUES (?, ?, ?, ?, 0)
		ON CONFLICT(owner, remote) DO UPDATE SET last_activity_at = last_activity_at
	`, newID.String(), owner.String(), remote.String(), now.Unix()); err != nil {
		return domain.ThreadID{}, errorfamily.WrapInfrastructuref(err, "store.thread_insert", "insert thread %s", newID)
	}

	// The ON CONFLICT above is a no-op update so a concurrent insert cannot
	// fail the call; re-read to return the winner's id.
	if err := s.db.QueryRowContext(ctx, `
		SELECT id FROM threads WHERE owner = ? AND remote = ?
	`, owner.String(), remote.String()).Scan(&id); err != nil {
		return domain.ThreadID{}, errorfamily.WrapInfrastructuref(err, "store.thread_reread", "re-read thread for %s/%s", owner, remote)
	}

	return domain.MustThreadID(id), nil
}

// GetThread fetches one thread by id, scoped to the owner.
func (s *Messages) GetThread(
	ctx context.Context, owner domain.Extension, id domain.ThreadID,
) (domain.Thread, error) {
	var (
		ownerStr, remoteStr     string
		lastActivity            int64
		unread                  int
		pinned, archived, muted int
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT owner, remote, last_activity_at, unread, pinned, archived, muted
		FROM threads WHERE id = ? AND owner = ?
	`, id.String(), owner.String()).Scan(&ownerStr, &remoteStr, &lastActivity, &unread, &pinned, &archived, &muted)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Thread{}, ErrNotFound
	}
	if err != nil {
		return domain.Thread{}, errorfamily.WrapInfrastructuref(err, "store.thread_get", "get thread %s", id)
	}
	return domain.Thread{
		ID:             id,
		Owner:          domain.MustParseExtension(ownerStr),
		Remote:         domain.MustParsePhone(remoteStr),
		LastActivityAt: time.Unix(lastActivity, 0),
		Unread:         unread,
		Pinned:         pinned == 1,
		Archived:       archived == 1,
		Muted:          muted == 1,
	}, nil
}

// MarkThreadRead zeroes the unread counter.
func (s *Messages) MarkThreadRead(ctx context.Context, owner domain.Extension, id domain.ThreadID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE threads SET unread = 0 WHERE id = ? AND owner = ?
	`, id.String(), owner.String())
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.thread_mark_read", "mark thread %s read", id)
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
		return domain.Attachment{}, errorfamily.WrapInfrastructuref(err, "store.attachment_get", "get attachment %s (row %s of message %s)", id, attachmentID, messageID)
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
	ctx context.Context, owner domain.Extension, threadID domain.ThreadID, page, limit int, //nolint:branching-flow // mis-order degrades visibly (a short window) and page < 0 is guarded; two callers
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
		return nil, false, fmt.Errorf("list messages page %d of thread %s: %w", page, threadID, err) //nolint:erraudit // family-neutral propagation: the inner error owns the family
	}

	hasMore := len(msgs) > limit
	if hasMore {
		msgs = msgs[:limit]
	}
	if err := s.attachAttachments(ctx, msgs); err != nil {
		return nil, false, fmt.Errorf("attach attachments to thread %s messages: %w", threadID, err) //nolint:erraudit // family-neutral propagation: the inner error owns the family
	}

	// Query was newest-first for the LIMIT; the UI wants a chat transcript,
	// oldest at the top.
	slices.Reverse(msgs)

	return msgs, hasMore, nil
}
