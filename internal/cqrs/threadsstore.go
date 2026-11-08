// The thread backend (SUPERB Phase 5, A33): the projections that mirror
// the legacy threads/messages/attachments rows into system.db, and the
// dispatch-backed Threads store implementing readmodels.ThreadGate. The
// list semantics (pinned first, then most recently active; archived
// excluded from the default view; ASCII case-insensitive search) are the
// SQL's, computed over the owner-filtered projection scans.
package cqrs

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// The three projection collections (query names in system.db).
const (
	threadsCollection    = "threads"
	messagesCollection   = "thread_messages"
	attachmentsColName   = "message_attachments"
)

// threadsKeyInput is the threads projection's key shape.
type threadsKeyInput struct {
	ThreadID string
}

// threadSummaryOf renders the summary view the append leaves behind: the
// legacy row computed by domain.AppendToThread plus this message's
// preview when it is the newest (journal order is creation order for the
// live lane and the sorted importer alike — the A37 battery pins it).
func threadSummaryOf(prev readmodels.ThreadSummaryView, f domain.MessageAppended) readmodels.ThreadSummaryView {
	view := prev
	view.ThreadID = f.ThreadID
	view.Owner = f.Owner
	view.Network = f.Network
	view.Remote = f.Remote
	view.Unread = prev.Unread
	view.Archived = prev.Archived
	if f.Direction == string(domain.DirectionInbound) {
		view.Unread++
		view.Archived = false
	}
	if prev.ThreadID == "" || f.CreatedAt.UnixNano() >= prev.LastActivity {
		view.LastActivity = f.CreatedAt.UnixNano()
		view.LastBody = f.Body
		view.LastDirection = f.Direction
		view.LastChannel = f.Channel
		view.LastAttachment = len(f.Attachments) > 0
	} else {
		view.LastActivity = prev.LastActivity
	}
	return view
}

// messageViewOf renders one transcript row.
func messageViewOf(f domain.MessageAppended) readmodels.MessageView {
	return readmodels.MessageView{
		MessageID:     f.MessageID,
		ThreadID:      f.ThreadID,
		Owner:         f.Owner,
		Network:       f.Network,
		Remote:        f.Remote,
		Direction:     f.Direction,
		Channel:       f.Channel,
		Body:          f.Body,
		Status:        f.Status,
		ProviderRef:   f.ProviderRef,
		FailureKind:   f.FailureKind,
		FailureDetail: f.FailureDetail,
		CreatedAt:     f.CreatedAt.UnixNano(),
	}
}

// messagesKey and attachmentsKey are the keyed collections' key shapes
// (field-name-matched by the remove/update key inference — every fact
// that folds them carries the matching field).
type messagesKey struct {
	MessageID string
}

type attachmentsKey struct {
	ID string
}

// threadProjections declares the three collections. Update folds (the
// prev-aware shape) carry only their delta — the row's other fields
// survive through prev, exactly as the SQL UPDATE touched one column;
// removes are keyed by the fact's id field.
func threadProjections() []system.ProjectionDeclaration {
	threads := metaengine.Query[threadsKeyInput, readmodels.ThreadSummaryView](threadsCollection,
		metaengine.OnRecord(domain.MessageAppended{},
			func(_ record.Record, e domain.MessageAppended, prev readmodels.ThreadSummaryView) readmodels.ThreadSummaryView {
				return threadSummaryOf(prev, e)
			}),
		metaengine.OnRecord(domain.ThreadFlagSet{},
			func(_ record.Record, e domain.ThreadFlagSet, prev readmodels.ThreadSummaryView) readmodels.ThreadSummaryView {
				switch e.Flag {
				case "pinned":
					prev.Pinned = e.On
				case "archived":
					prev.Archived = e.On
				case "muted":
					prev.Muted = e.On
				}
				return prev
			}),
		metaengine.OnRecord(domain.ThreadUnreadCleared{},
			func(_ record.Record, _ domain.ThreadUnreadCleared, prev readmodels.ThreadSummaryView) readmodels.ThreadSummaryView {
				prev.Unread = 0
				return prev
			}),
		metaengine.OnRecord(domain.ThreadDeleted{}, metaengine.Remove[readmodels.ThreadSummaryView]()),
		metaengine.FilterOnField[readmodels.ThreadSummaryView]("owner", metaengine.FilterEq),
	)

	messages := metaengine.Query[messagesKey, readmodels.MessageView](messagesCollection,
		metaengine.OnRecord(domain.MessageAppended{},
			func(_ record.Record, e domain.MessageAppended) (string, readmodels.MessageView) {
				return e.MessageID, messageViewOf(e)
			}),
		metaengine.OnRecord(domain.OutboundAdvanced{},
			func(_ record.Record, e domain.OutboundAdvanced, prev readmodels.MessageView) readmodels.MessageView {
				prev.Status = e.Status
				prev.ProviderRef = e.ProviderRef
				prev.FailureKind = e.FailureKind
				prev.FailureDetail = e.FailureDetail
				return prev
			}),
		metaengine.OnRecord(domain.MessageRemoved{}, metaengine.Remove[readmodels.MessageView]()),
		metaengine.FilterOnField[readmodels.MessageView]("owner", metaengine.FilterEq),
		metaengine.FilterOnField[readmodels.MessageView]("provider_ref", metaengine.FilterEq),
		metaengine.FilterOnField[readmodels.MessageView]("message_id", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.MessageView]("created_at", true),
	)

	attachments := metaengine.Query[attachmentsKey, readmodels.AttachmentView](attachmentsColName,
		metaengine.OnRecord(domain.AttachmentStored{},
			func(_ record.Record, e domain.AttachmentStored) (string, readmodels.AttachmentView) {
				return e.ID, readmodels.AttachmentView{
					ID:        e.ID,
					MessageID: e.MessageID,
					Owner:     e.Owner,
					Name:      e.Name,
					MimeType:  e.MimeType,
					SizeBytes: e.SizeBytes,
					Path:      e.Path,
				}
			}),
		metaengine.OnRecord(domain.AttachmentRemoved{}, metaengine.Remove[readmodels.AttachmentView]()),
		metaengine.FilterOnField[readmodels.AttachmentView]("owner", metaengine.FilterEq),
		metaengine.FilterOnField[readmodels.AttachmentView]("id", metaengine.FilterEq),
	)

	decls := make([]system.ProjectionDeclaration, 0, 3)
	for _, q := range []any{threads, messages, attachments} {
		decls = append(decls, system.RawQuery(q))
	}
	return decls
}

// ThreadsDomainDecl is the thread aggregate's share of the system's
// DomainConfig (a registration failure is a boot-time miswire and panics
// loudly).
func ThreadsDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterThreads(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register thread aggregate: %v", err))
			}
		},
		Projections: threadProjections(),
		Events: []event.Type{
			event.Type(metaengine.EventTypeName(domain.MessageAppended{})),
			event.Type(metaengine.EventTypeName(domain.OutboundAdvanced{})),
			event.Type(metaengine.EventTypeName(domain.ThreadFlagSet{})),
			event.Type(metaengine.EventTypeName(domain.ThreadUnreadCleared{})),
			event.Type(metaengine.EventTypeName(domain.ThreadDeleted{})),
			event.Type(metaengine.EventTypeName(domain.MessageRemoved{})),
			event.Type(metaengine.EventTypeName(domain.AttachmentRemoved{})),
		},
	}
}

// ─── the backend the server keeps calling ──────────────────────────

// Threads is the dispatch-backed thread/message store: the store.Messages
// contract answered by the journal and the system.db projections.
type Threads struct {
	sys      *system.System
	threads  *metaengine.TypedReader[readmodels.ThreadSummaryView]
	messages *metaengine.TypedReader[readmodels.MessageView]
	attachs  *metaengine.TypedReader[readmodels.AttachmentView]
	vers     streamVersions
	mu       sync.Mutex
}

var _ readmodels.ThreadGate = (*Threads)(nil)

// NewThreads builds the backend over a System.
func NewThreads(sys *system.System) *Threads {
	return &Threads{
		sys:      sys,
		threads:  metaengine.NewReader[readmodels.ThreadSummaryView](sys.MetaEngine(), threadsCollection),
		messages: metaengine.NewReader[readmodels.MessageView](sys.MetaEngine(), messagesCollection),
		attachs:  metaengine.NewReader[readmodels.AttachmentView](sys.MetaEngine(), attachmentsColName),
		vers:     streamVersions{},
	}
}

// key resolves the thread stream key for a stored thread id (any of the
// three identity fields answers; the summary carries them all).
func (s *Threads) key(owner domain.Extension, id domain.ThreadID) (string, error) {
	view, err := s.threadByID(owner, id)
	if err != nil {
		return "", err
	}
	return ThreadStreamKey(owner, domain.MustNetwork(view.Network),
		domain.MustParseAddress(domain.MustNetwork(view.Network), view.Remote)), nil
}

func (s *Threads) threadByID(owner domain.Extension, id domain.ThreadID) (readmodels.ThreadSummaryView, error) {
	views, err := s.threads.Scan(context.Background(),
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()))
	if err != nil {
		return readmodels.ThreadSummaryView{}, errorfamily.WrapInfrastructuref(err, "cqrs.thread_lookup",
			"lookup thread %s for %s", id, owner)
	}
	for _, v := range views {
		if v.ThreadID == id.String() {
			return v, nil
		}
	}
	return readmodels.ThreadSummaryView{}, store.ErrNotFound
}

// awaitOwnWrite waits until the projection worker folded the just-
// dispatched batch (D7): the messages handlers re-render list and
// transcript partials in the SAME request.
func (s *Threads) awaitOwnWrite(ctx context.Context, key string) {
	awaitOwnWrite(ctx, s.sys, ThreadsStreamType, key, threadsCollection, s.vers)
}

// AppendMessage appends one message, creating the thread on first sight.
func (s *Threads) AppendMessage(ctx context.Context, msg domain.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := ThreadStreamKey(msg.Owner, msg.Remote.Network(), msg.Remote)
	if err := dispatchThreadCommand(ctx, s.sys, "thread.append", key,
		func(base *command.BasicCommand) *AppendMessage {
			return &AppendMessage{BasicCommand: base, Msg: msg}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, key)
	return nil
}

// AdvanceOutbound moves one outbound message's delivery machine; an
// illegal move is inert (the settled state answers, the legacy
// updatedOrNotFound contract preserved for a missing message).
func (s *Threads) AdvanceOutbound(ctx context.Context, id domain.MessageID, step domain.DeliveryStep) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg, err := s.messageByID(id)
	if err != nil {
		return err
	}
	key := ThreadStreamKey(domain.MustParseExtension(msg.Owner),
		domain.MustNetwork(msg.Network), domain.MustParseAddress(domain.MustNetwork(msg.Network), msg.Remote))
	if err := dispatchThreadCommand(ctx, s.sys, "thread.advance_delivery", key,
		func(base *command.BasicCommand) *AdvanceDelivery {
			return &AdvanceDelivery{
				BasicCommand: base,
				Owner:        domain.MustParseExtension(msg.Owner),
				Network:      domain.MustNetwork(msg.Network),
				Remote:       domain.MustParseAddress(domain.MustNetwork(msg.Network), msg.Remote),
				MessageID:    id,
				Step:         step,
			}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, key)
	return nil
}

// UpdateOutboundStatus is the store-era signature over the step seam: the
// service already advanced the machine; the decider re-derives the move
// from the delivered target (every caller's next state came from
// DeliveryAdvance, so the derived step is the same move).
func (s *Threads) UpdateOutboundStatus(ctx context.Context, id domain.MessageID, delivery domain.Delivery) error {
	step := stepForTarget(delivery)
	if step.Kind == "" {
		return nil
	}
	return s.AdvanceOutbound(ctx, id, step)
}

// stepForTarget derives a DeliveryStep whose DeliveryAdvance lands on the
// target slice (the store-era callers' next values are always machine
// products, so the reverse mapping is exact).
func stepForTarget(target domain.Delivery) domain.DeliveryStep {
	switch target.Status {
	case domain.StatusSent:
		return domain.SendAccepted(target.ProviderRef).Step()
	case domain.StatusFailed:
		return domain.SendRefused(target.FailureKind, target.FailureDetail).Step()
	case domain.StatusDelivered:
		return domain.ProviderVerdict(domain.VerdictDelivered, "").Step()
	default:
		return domain.DeliveryStep{}
	}
}

// MessageByProviderRef resolves an outbound message by its gateway
// correlation id (the webhook lane's lookup). The provider_ref filter is
// declared on the messages projection.
func (s *Threads) MessageByProviderRef(ctx context.Context, ref string) (domain.Message, error) {
	if ref == "" {
		return domain.Message{}, store.ErrNotFound
	}
	views, err := s.messages.Scan(ctx,
		metaengine.WithFilter("provider_ref", metaengine.FilterEq, ref),
		metaengine.WithLimit(2),
	)
	if err != nil {
		return domain.Message{}, errorfamily.WrapInfrastructuref(err, "cqrs.message_by_ref", "resolve provider ref %s", ref)
	}
	for _, v := range views {
		if v.Direction == string(domain.DirectionOutbound) {
			return s.messageDomain(v), nil
		}
	}
	return domain.Message{}, store.ErrNotFound
}

// FindThread resolves the thread id for an address; creation is the first
// append's fact (the composer's candidate id is adopted then), so an
// unknown address answers a FRESH id — the legacy empty-row insert
// disappears by design (Errata: no thread row exists until a message
// lands).
func (s *Threads) FindThread(
	ctx context.Context, owner domain.Extension, remote domain.Address, now time.Time,
) (domain.ThreadID, error) {
	key := ThreadStreamKey(owner, remote.Network(), remote)
	_ = now
	views, err := s.threads.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()))
	if err != nil {
		return domain.ThreadID{}, errorfamily.WrapInfrastructuref(err, "cqrs.thread_find", "find thread for %s/%s", owner, remote)
	}
	for _, v := range views {
		if v.Network == string(remote.Network()) && v.Remote == remote.String() {
			return domain.MustThreadID(v.ThreadID), nil
		}
	}
	_ = key
	return domain.GenerateThreadID(), nil
}

// GetThread fetches one thread scoped to the owner.
func (s *Threads) GetThread(ctx context.Context, owner domain.Extension, id domain.ThreadID) (domain.Thread, error) {
	view, err := s.threadByID(owner, id)
	if err != nil {
		return domain.Thread{}, err
	}
	return domain.Thread{
		ID:             domain.MustThreadID(view.ThreadID),
		Owner:          domain.MustParseExtension(view.Owner),
		Remote:         domain.MustParseAddress(domain.MustNetwork(view.Network), view.Remote),
		LastActivityAt: time.Unix(0, view.LastActivity),
		Unread:         view.Unread,
		Pinned:         view.Pinned,
		Archived:       view.Archived,
		Muted:          view.Muted,
	}, nil
}

// MarkThreadRead zeroes the unread counter.
func (s *Threads) MarkThreadRead(ctx context.Context, owner domain.Extension, id domain.ThreadID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	view, err := s.threadByID(owner, id)
	if err != nil {
		return err
	}
	remote := domain.MustParseAddress(domain.MustNetwork(view.Network), view.Remote)
	key := ThreadStreamKey(owner, remote.Network(), remote)
	if err := dispatchThreadCommand(ctx, s.sys, "thread.mark_read", key,
		func(base *command.BasicCommand) *MarkThreadRead {
			return &MarkThreadRead{BasicCommand: base, Owner: owner,
				Network: remote.Network(), Remote: remote}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, key)
	return nil
}

// SetThreadFlag sets or clears one organization flag on the owner's
// thread; a missing thread (or another owner's) is ErrNotFound.
func (s *Threads) SetThreadFlag(
	ctx context.Context, owner domain.Extension, id domain.ThreadID, flag readmodels.ThreadFlag, on bool,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	view, err := s.threadByID(owner, id)
	if err != nil {
		return err
	}
	remote := domain.MustParseAddress(domain.MustNetwork(view.Network), view.Remote)
	key := ThreadStreamKey(owner, remote.Network(), remote)
	if err := dispatchThreadCommand(ctx, s.sys, "thread.set_flag", key,
		func(base *command.BasicCommand) *SetThreadFlag {
			return &SetThreadFlag{BasicCommand: base, Owner: owner,
				Network: remote.Network(), Remote: remote, Flag: string(flag), On: on}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, key)
	return nil
}

// DeleteThread dispatches the tombstone batch (folds Remove every row of
// the thread), awaits visibility, then erases the stream's journal bytes
// and snapshot — the sweep's one call (retention_days semantics, A36).
func (s *Threads) DeleteThread(ctx context.Context, owner domain.Extension, id domain.ThreadID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	view, err := s.threadByID(owner, id)
	if err != nil {
		return err
	}
	remote := domain.MustParseAddress(domain.MustNetwork(view.Network), view.Remote)
	key := ThreadStreamKey(owner, remote.Network(), remote)
	if err := dispatchThreadCommand(ctx, s.sys, "thread.delete", key,
		func(base *command.BasicCommand) *DeleteThread {
			return &DeleteThread{BasicCommand: base, Owner: owner,
				Network: remote.Network(), Remote: remote, Reason: reason}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, key)
	ref, err := streamRef(ThreadsStreamType, key)
	if err != nil {
		return err
	}
	if _, err := s.sys.TrimStream(ctx, ref); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_trim", "trim journal of thread %s", id)
	}
	return nil
}

// messageDomain rehydrates the domain.Message from a view row.
func (s *Threads) messageDomain(v readmodels.MessageView) domain.Message {
	return domain.Message{
		ID:            domain.MustMessageID(v.MessageID),
		ThreadID:      domain.MustThreadID(v.ThreadID),
		Owner:         domain.MustParseExtension(v.Owner),
		Remote:        domain.MustParseAddress(domain.MustNetwork(v.Network), v.Remote),
		Direction:     domain.MustDirection(v.Direction),
		Channel:       domain.MustChannel(v.Channel),
		Body:          v.Body,
		Status:        domain.MustOutboundStatus(v.Status),
		ProviderRef:   v.ProviderRef,
		FailureKind:   domain.FailureKind(v.FailureKind),
		FailureDetail: v.FailureDetail,
		CreatedAt:     time.Unix(0, v.CreatedAt),
	}
}

func (s *Threads) messageByID(id domain.MessageID) (readmodels.MessageView, error) {
	views, err := s.messages.Scan(context.Background(),
		metaengine.WithFilter("message_id", metaengine.FilterEq, id.String()),
		metaengine.WithLimit(1),
	)
	if err != nil {
		return readmodels.MessageView{}, errorfamily.WrapInfrastructuref(err, "cqrs.message_lookup", "lookup message %s", id)
	}
	if len(views) == 0 {
		return readmodels.MessageView{}, store.ErrNotFound
	}
	return views[0], nil
}

// sortSummaries orders the owner's rows as the legacy list SQL did:
// pinned first, then most recently active.
func sortSummaries(views []readmodels.ThreadSummaryView) {
	sort.SliceStable(views, func(i, j int) bool {
		if views[i].Pinned != views[j].Pinned {
			return views[i].Pinned
		}
		return views[i].LastActivity > views[j].LastActivity
	})
}

// ownerSummaries scans the owner's thread rows (every list variant's
// base).
func (s *Threads) ownerSummaries(ctx context.Context, owner domain.Extension) ([]readmodels.ThreadSummaryView, error) {
	views, err := s.threads.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.thread_list", "list threads for %s", owner)
	}
	return views, nil
}

// ListThreads returns the owner's active threads.
func (s *Threads) ListThreads(ctx context.Context, owner domain.Extension) ([]readmodels.ThreadSummaryView, error) {
	views, err := s.ownerSummaries(ctx, owner)
	if err != nil {
		return nil, err
	}
	out := make([]readmodels.ThreadSummaryView, 0, len(views))
	for _, v := range views {
		if !v.Archived {
			out = append(out, v)
		}
	}
	sortSummaries(out)
	return out, nil
}

// ListArchivedThreads returns the owner's archived threads (activity
// order only — the filing cabinet).
func (s *Threads) ListArchivedThreads(ctx context.Context, owner domain.Extension) ([]readmodels.ThreadSummaryView, error) {
	views, err := s.ownerSummaries(ctx, owner)
	if err != nil {
		return nil, err
	}
	out := make([]readmodels.ThreadSummaryView, 0, len(views))
	for _, v := range views {
		if v.Archived {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].LastActivity > out[j].LastActivity })
	return out, nil
}

// CountArchived reports how many of the owner's threads are archived.
func (s *Threads) CountArchived(ctx context.Context, owner domain.Extension) (int, error) {
	views, err := s.ownerSummaries(ctx, owner)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, v := range views {
		if v.Archived {
			n++
		}
	}
	return n, nil
}

// SearchThreads matches the owner's active threads whose remote or ANY
// message body contains the needle (ASCII case-insensitive; the legacy
// LIKE-escape contract is a literal contains).
func (s *Threads) SearchThreads(ctx context.Context, owner domain.Extension, needle string) ([]readmodels.ThreadSummaryView, error) {
	active, err := s.ListThreads(ctx, owner)
	if err != nil {
		return nil, err
	}
	msgs, err := s.messages.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.thread_search", "search messages of %s", owner)
	}
	lower := strings.ToLower(needle)
	hitThreads := map[string]bool{}
	for _, m := range msgs {
		if strings.Contains(strings.ToLower(m.Body), lower) {
			hitThreads[m.ThreadID] = true
		}
	}
	out := make([]readmodels.ThreadSummaryView, 0, len(active))
	for _, v := range active {
		if hitThreads[v.ThreadID] || strings.Contains(strings.ToLower(v.Remote), lower) {
			out = append(out, v)
		}
	}
	return out, nil
}

// threadMessages returns the owner's messages of one thread, newest
// first (the raw engine order the callers page/reverse from).
func (s *Threads) threadMessages(ctx context.Context, owner domain.Extension, threadID domain.ThreadID) ([]readmodels.MessageView, error) {
	views, err := s.messages.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("created_at", true))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.thread_messages", "list messages of thread %s", threadID)
	}
	out := make([]readmodels.MessageView, 0, len(views))
	for _, v := range views {
		if v.ThreadID == threadID.String() {
			out = append(out, v)
		}
	}
	return out, nil
}

// ListMessages returns the messages of one thread, oldest first.
func (s *Threads) ListMessages(
	ctx context.Context, owner domain.Extension, threadID domain.ThreadID, limit int,
) ([]domain.Message, error) {
	views, err := s.threadMessages(ctx, owner, threadID)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(views) > limit {
		views = views[:limit]
	}
	msgs := make([]domain.Message, 0, len(views))
	for _, v := range views {
		msgs = append(msgs, s.messageDomain(v))
	}
	if err := s.attachAttachments(ctx, msgs); err != nil {
		return nil, err
	}
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// ListMessagesPage returns one page of a thread's messages, oldest
// first; page 0 is the newest window.
func (s *Threads) ListMessagesPage(
	ctx context.Context, owner domain.Extension, threadID domain.ThreadID, page, limit int,
) ([]domain.Message, bool, error) {
	if page < 0 {
		page = 0
	}
	views, err := s.threadMessages(ctx, owner, threadID)
	if err != nil {
		return nil, false, err
	}
	start := page * limit
	if start >= len(views) {
		return []domain.Message{}, false, nil
	}
	window := views[start:]
	hasMore := len(window) > limit
	if hasMore {
		window = window[:limit]
	}
	msgs := make([]domain.Message, 0, len(window))
	for _, v := range window {
		msgs = append(msgs, s.messageDomain(v))
	}
	if err := s.attachAttachments(ctx, msgs); err != nil {
		return nil, false, err
	}
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, hasMore, nil
}

// attachAttachments loads the owner's attachment rows once and folds them
// onto their messages (the legacy one-query fan-out; name order per
// message preserved by the sort).
func (s *Threads) attachAttachments(ctx context.Context, msgs []domain.Message) error {
	if len(msgs) == 0 {
		return nil
	}
	byOwner := make(map[string]struct{}, len(msgs))
	for _, m := range msgs {
		byOwner[m.Owner.String()] = struct{}{}
	}
	rows := make([]readmodels.AttachmentView, 0)
	for owner := range byOwner {
		views, err := s.attachs.Scan(ctx,
			metaengine.WithFilter("owner", metaengine.FilterEq, owner))
		if err != nil {
			return errorfamily.WrapInfrastructuref(err, "cqrs.attachments_list", "list attachments of %s", owner)
		}
		rows = append(rows, views...)
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	byMessage := map[string][]domain.Attachment{}
	for _, v := range rows {
		byMessage[v.MessageID] = append(byMessage[v.MessageID], domain.Attachment{
			ID:        domain.MustAttachmentID(v.ID),
			MessageID: domain.MustMessageID(v.MessageID),
			Name:      v.Name,
			MimeType:  v.MimeType,
			SizeBytes: v.SizeBytes,
			Path:      v.Path,
		})
	}
	for i := range msgs {
		msgs[i].Attachments = byMessage[msgs[i].ID.String()]
	}
	return nil
}

// ListAllThreads returns every owner's threads, most recently active
// first, bounded by limit (the admin inventory; archived included).
func (s *Threads) ListAllThreads(ctx context.Context, limit int) ([]readmodels.ThreadSummaryView, error) {
	views, err := s.threads.Scan(ctx, metaengine.WithLimit(limit))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.threads_list_all", "list all threads")
	}
	sort.SliceStable(views, func(i, j int) bool { return views[i].LastActivity > views[j].LastActivity })
	if limit > 0 && len(views) > limit {
		views = views[:limit]
	}
	return views, nil
}

// GetThreadAny fetches one thread by id across owners (admin seam).
func (s *Threads) GetThreadAny(ctx context.Context, id domain.ThreadID) (domain.Thread, error) {
	views, err := s.threads.Scan(ctx, metaengine.WithLimit(0))
	if err != nil {
		return domain.Thread{}, errorfamily.WrapInfrastructuref(err, "cqrs.thread_get_any", "get thread %s (admin)", id)
	}
	for _, v := range views {
		if v.ThreadID == id.String() {
			owner := domain.MustParseExtension(v.Owner)
			return s.GetThread(ctx, owner, id)
		}
	}
	return domain.Thread{}, store.ErrNotFound
}

// attachmentByID resolves one attachment row by key.
func (s *Threads) attachmentByID(ctx context.Context, id domain.AttachmentID) (readmodels.AttachmentView, error) {
	views, err := s.attachs.Scan(ctx,
		metaengine.WithFilter("id", metaengine.FilterEq, id.String()),
		metaengine.WithLimit(1))
	if err != nil {
		return readmodels.AttachmentView{}, errorfamily.WrapInfrastructuref(err, "cqrs.attachment_get", "get attachment %s", id)
	}
	if len(views) == 0 {
		return readmodels.AttachmentView{}, store.ErrNotFound
	}
	return views[0], nil
}

// AttachmentByID resolves one attachment scoped to the owner.
func (s *Threads) AttachmentByID(ctx context.Context, owner domain.Extension, id domain.AttachmentID) (domain.Attachment, error) {
	view, err := s.attachmentByID(ctx, id)
	if err != nil {
		return domain.Attachment{}, err
	}
	if view.Owner != owner.String() {
		return domain.Attachment{}, store.ErrNotFound
	}
	return attachmentDomain(view), nil
}

// AttachmentByIDAny resolves one attachment across owners (admin seam).
func (s *Threads) AttachmentByIDAny(ctx context.Context, id domain.AttachmentID) (domain.Attachment, error) {
	view, err := s.attachmentByID(ctx, id)
	if err != nil {
		return domain.Attachment{}, err
	}
	return attachmentDomain(view), nil
}

func attachmentDomain(v readmodels.AttachmentView) domain.Attachment {
	return domain.Attachment{
		ID:        domain.MustAttachmentID(v.ID),
		MessageID: domain.MustMessageID(v.MessageID),
		Name:      v.Name,
		MimeType:  v.MimeType,
		SizeBytes: v.SizeBytes,
		Path:      v.Path,
	}
}

// compile guard: the ErrNotFound passthrough keeps the handler 404 map.
var _ = errors.Is
