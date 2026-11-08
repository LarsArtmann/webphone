package readmodels

import (
	"context"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// Thread-stream views (SUPERB Phase 5): the system.db projection shapes
// the server renders — the legacy store.ThreadSummary / message /
// attachment rows as engine views. Field-for-field mirrors; the gates
// re-parse into domain types exactly as the store's scans did.

// ThreadSummaryView is one thread row as the list renders it: the thread
// plus its last message's preview.
type ThreadSummaryView struct {
	ThreadID       string `json:"thread_id"`
	Owner          string `json:"owner"`
	Network        string `json:"network"`
	Remote         string `json:"remote"`
	LastActivity   int64  `json:"last_activity"` // unix nanos
	Unread         int    `json:"unread"`
	Pinned         bool   `json:"pinned"`
	Archived       bool   `json:"archived"`
	Muted          bool   `json:"muted"`
	LastBody       string `json:"last_body"`
	LastDirection  string `json:"last_direction"`
	LastChannel    string `json:"last_channel"`
	LastAttachment bool   `json:"last_attachment"`
}

// MessageView is one message row (the transcript read).
type MessageView struct {
	MessageID     string `json:"message_id"`
	ThreadID      string `json:"thread_id"`
	Owner         string `json:"owner"`
	Network       string `json:"network"`
	Remote        string `json:"remote"`
	Direction     string `json:"direction"`
	Channel       string `json:"channel"`
	Body          string `json:"body"`
	Status        string `json:"status"`
	ProviderRef   string `json:"provider_ref"`
	FailureKind   string `json:"failure_kind"`
	FailureDetail string `json:"failure_detail"`
	CreatedAt     int64  `json:"created_at"` // unix nanos — the sort key
}

// AttachmentView is one attachment row.
type AttachmentView struct {
	ID        string `json:"id"`
	MessageID string `json:"message_id"`
	Owner     string `json:"owner"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	Path      string `json:"path"`
}

// ThreadFlag is the persistence-layer spelling of the three organization
// axes (the store's ThreadFlag, journal-side).
type ThreadFlag string

const (
	FlagPinned   ThreadFlag = "pinned"
	FlagArchived ThreadFlag = "archived"
	FlagMuted    ThreadFlag = "muted"
)

// ThreadGate is the thread/message/attachment seam: every store.Messages
// method the server and the messaging service call, answered by the
// journal-backed backend (the store keeps serving as the import source
// until the ladder retires). ErrNotFound keeps its store meaning.
type ThreadGate interface {
	AppendMessage(ctx context.Context, msg domain.Message) error
	AdvanceOutbound(ctx context.Context, id domain.MessageID, step domain.DeliveryStep) error
	MessageByProviderRef(ctx context.Context, ref string) (domain.Message, error)
	FindThread(ctx context.Context, owner domain.Extension, remote domain.Address, now time.Time) (domain.ThreadID, error)
	GetThread(ctx context.Context, owner domain.Extension, id domain.ThreadID) (domain.Thread, error)
	MarkThreadRead(ctx context.Context, owner domain.Extension, id domain.ThreadID) error
	SetThreadFlag(ctx context.Context, owner domain.Extension, id domain.ThreadID, flag ThreadFlag, on bool) error
	ListThreads(ctx context.Context, owner domain.Extension) ([]ThreadSummaryView, error)
	ListArchivedThreads(ctx context.Context, owner domain.Extension) ([]ThreadSummaryView, error)
	CountArchived(ctx context.Context, owner domain.Extension) (int, error)
	SearchThreads(ctx context.Context, owner domain.Extension, needle string) ([]ThreadSummaryView, error)
	ListMessages(ctx context.Context, owner domain.Extension, threadID domain.ThreadID, limit int) ([]domain.Message, error)
	ListMessagesPage(
		ctx context.Context, owner domain.Extension, threadID domain.ThreadID, page, limit int,
	) ([]domain.Message, bool, error)
	ListAllThreads(ctx context.Context, limit int) ([]ThreadSummaryView, error)
	GetThreadAny(ctx context.Context, id domain.ThreadID) (domain.Thread, error)
	AttachmentByID(ctx context.Context, owner domain.Extension, id domain.AttachmentID) (domain.Attachment, error)
	AttachmentByIDAny(ctx context.Context, id domain.AttachmentID) (domain.Attachment, error)
	// CountMessages reports the total stored messages across all owners
	// (the metrics surface's webphone_messages_total).
	CountMessages(ctx context.Context) (int64, error)
	// CountThreads reports the total stored threads across all owners
	// (the metrics surface's webphone_threads_total).
	CountThreads(ctx context.Context) (int64, error)
	// DeleteThread is the retention/owner removal seam: it dispatches the
	// tombstone (folds Remove every row), awaits visibility, then erases
	// the stream's journal bytes (sys.TrimStream) — the sweep's one call.
	DeleteThread(ctx context.Context, owner domain.Extension, id domain.ThreadID, reason string) error
}
