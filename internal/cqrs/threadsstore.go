// The thread backend (SUPERB Phase 5, A33): the projections that mirror
// the legacy threads/messages/attachments rows into system.db, and the
// dispatch-backed Threads store implementing readmodels.ThreadGate. The
// list semantics (pinned first, then most recently active; archived
// excluded from the default view; ASCII case-insensitive search) are the
// SQL's, computed over the owner-filtered projection scans.
package cqrs

import (
	"context"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	errorfamily "github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

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

func (s *Threads) threadByID(owner domain.Extension, id domain.ThreadID) (readmodels.ThreadSummaryView, error) {
	views, err := s.threads.Scan(context.Background(),
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithLimit(0))
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
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithLimit(0))
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
