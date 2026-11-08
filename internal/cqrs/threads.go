// The thread aggregate (SUPERB Phase 5, A32): one stream per thread keyed
// (owner, network, remote) — the legacy unique index as a stream key. The
// append invariants stay in domain.AppendToThread (consumed verbatim as
// the decider's apply), delivery legality stays in domain.DeliveryAdvance
// (deliveryMoves consumed at the decision), and the decider additionally
// rejects a message id the stream already carries — the importer-retry and
// webhook-replay guard. ThreadDeleted embeds every id the stream ever
// carried so every projection fold is a pure keyed Remove and replay is
// exact.
package cqrs

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/snapshot/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// ThreadsStreamType keys per-thread streams.
const ThreadsStreamType = "thread"

// ThreadStreamKey builds the per-thread stream key from the thread's
// identity axis. The remote is hex-encoded: matrix addresses contain the
// separator-rich forms (`@local:server`) and every remote shape must map
// to exactly one key. The key is never parsed back — the facts carry the
// readable fields.
func ThreadStreamKey(owner domain.Extension, network domain.Network, remote domain.Address) string {
	return owner.String() + "|" + string(network) + "|" + hex.EncodeToString([]byte(remote.String()))
}

// Thread stream commands.

// AppendMessage appends one message (inbound or outbound); on an empty
// stream this is thread creation and the fact adopts the message's
// ThreadID (the composer pre-generates; the importer carries legacy ids).
type AppendMessage struct {
	*command.BasicCommand
	Msg domain.Message
}

// AdvanceDelivery moves one outbound message's delivery machine. The step
// is the wire form of the DeliveryEvent; an illegal move is a zero-event
// dispatch (the settled state answers).
type AdvanceDelivery struct {
	*command.BasicCommand
	Owner     domain.Extension
	Network   domain.Network
	Remote    domain.Address
	MessageID domain.MessageID
	Step      domain.DeliveryStep
}

// SetThreadFlag sets or clears one organization axis to the desired state.
type SetThreadFlag struct {
	*command.BasicCommand
	Owner    domain.Extension
	Network  domain.Network
	Remote   domain.Address
	Flag     string // "pinned" | "archived" | "muted"
	On       bool
}

// MarkThreadRead zeroes the unread counter (already-zero is a legal
// zero-event dispatch).
type MarkThreadRead struct {
	*command.BasicCommand
	Owner   domain.Extension
	Network domain.Network
	Remote  domain.Address
}

// DeleteThread is the retention/owner tombstone: the decision gathers
// every message and attachment id the stream carries so the folds Remove
// them all.
type DeleteThread struct {
	*command.BasicCommand
	Owner   domain.Extension
	Network domain.Network
	Remote  domain.Address
	Reason  string
}

// threadFlagLegal is the flag vocabulary's one home (the store's
// ThreadFlag spelled as journal data).
func threadFlagLegal(flag string) bool {
	return flag == "pinned" || flag == "archived" || flag == "muted"
}

// threadState is the full aggregate: the legacy thread row plus the id
// bookkeeping the deletes and delivery advances need.
type threadState struct {
	ThreadID  string
	Owner     string
	Network   string
	Remote    string
	LastAct   int64 // unix nanos
	Unread    int
	Pinned    bool
	Archived  bool
	Muted     bool
	Messages  []string // every appended message id, journal order
	Outbound  map[string]domain.Delivery
	AttachIDs []string // every attachment id, journal order
}

func (threadState) initial() threadState {
	return threadState{Outbound: map[string]domain.Delivery{}}
}

// threadDomain rebuilds the domain.Thread the ONE home (AppendToThread)
// consumes. Journal-trusted: only parsed values were ever appended.
func (s threadState) threadDomain() domain.Thread {
	if s.ThreadID == "" {
		return domain.Thread{}
	}
	return domain.Thread{
		ID:             domain.MustThreadID(s.ThreadID),
		Owner:          domain.MustParseExtension(s.Owner),
		Remote:         domain.MustParseAddress(domain.MustNetwork(s.Network), s.Remote),
		LastActivityAt: time.Unix(0, s.LastAct),
		Unread:         s.Unread,
		Pinned:         s.Pinned,
		Archived:       s.Archived,
		Muted:          s.Muted,
	}
}

// factMessage renders the appended message fact payload.
func factMessage(msg domain.Message) domain.MessageAppended {
	fact := domain.MessageAppended{
		MessageID:     msg.ID.String(),
		ThreadID:      msg.ThreadID.String(),
		Owner:         msg.Owner.String(),
		Network:       string(msg.Remote.Network()),
		Remote:        msg.Remote.String(),
		Direction:     string(msg.Direction),
		Channel:       string(msg.Channel),
		Body:          msg.Body,
		Status:        string(msg.Status),
		ProviderRef:   msg.ProviderRef,
		FailureKind:   string(msg.FailureKind),
		FailureDetail: msg.FailureDetail,
		CreatedAt:     msg.CreatedAt,
	}
	for _, att := range msg.Attachments {
		fact.Attachments = append(fact.Attachments, domain.AttachmentFact{
			ID:        att.ID.String(),
			MessageID: att.MessageID.String(),
			Name:      att.Name,
			MimeType:  att.MimeType,
			SizeBytes: att.SizeBytes,
			Path:      att.Path,
		})
	}
	return fact
}

// messageDomain rebuilds the domain.Message a fact carries (journal-trusted
// re-parse, the store scan twin).
func messageDomain(f domain.MessageAppended) domain.Message {
	msg := domain.Message{
		ID:            domain.MustMessageID(f.MessageID),
		ThreadID:      domain.MustThreadID(f.ThreadID),
		Owner:         domain.MustParseExtension(f.Owner),
		Remote:        domain.MustParseAddress(domain.MustNetwork(f.Network), f.Remote),
		Direction:     domain.MustDirection(f.Direction),
		Channel:       domain.MustChannel(f.Channel),
		Body:          f.Body,
		Status:        domain.MustOutboundStatus(f.Status),
		ProviderRef:   f.ProviderRef,
		FailureKind:   domain.FailureKind(f.FailureKind),
		FailureDetail: f.FailureDetail,
		CreatedAt:     f.CreatedAt,
	}
	for _, att := range f.Attachments {
		msg.Attachments = append(msg.Attachments, domain.Attachment{
			ID:        domain.MustAttachmentID(att.ID),
			MessageID: domain.MustMessageID(att.MessageID),
			Name:      att.Name,
			MimeType:  att.MimeType,
			SizeBytes: att.SizeBytes,
			Path:      att.Path,
		})
	}
	return msg
}

func applyThreadState(s threadState, evt event.Event) (threadState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.MessageAppended{})):
		fact, err := event.DecodePayloadAuto[domain.MessageAppended](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.thread_fold", "decode MessageAppended")
		}
		// The decider's apply IS domain.AppendToThread: the legacy row's
		// every invariant (creation adoption, last-activity stamp, inbound
		// unread+unarchive) computed by the ONE home, never a parallel copy.
		msg := messageDomain(fact)
		thread := domain.AppendToThread(s.threadDomain(), msg)
		s.ThreadID, s.Owner, s.Network, s.Remote = thread.ID.String(), thread.Owner.String(),
			string(thread.Remote.Network()), thread.Remote.String()
		s.LastAct, s.Unread, s.Archived = thread.LastActivityAt.UnixNano(), thread.Unread, thread.Archived
		s.Messages = append(s.Messages, msg.ID.String())
		if msg.Direction == domain.DirectionOutbound {
			s.Outbound[msg.ID.String()] = msg.DeliveryOf()
		}
	case event.Type(metaengine.EventTypeName(domain.AttachmentStored{})):
		attFact, err := event.DecodePayloadAuto[domain.AttachmentStored](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.thread_fold", "decode AttachmentStored")
		}
		s.AttachIDs = append(s.AttachIDs, attFact.ID)
	case event.Type(metaengine.EventTypeName(domain.OutboundAdvanced{})):
		fact, err := event.DecodePayloadAuto[domain.OutboundAdvanced](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.thread_fold", "decode OutboundAdvanced")
		}
		if d, ok := s.Outbound[fact.MessageID]; ok {
			d.Status = domain.MustOutboundStatus(fact.Status)
			d.ProviderRef = fact.ProviderRef
			d.FailureKind = domain.FailureKind(fact.FailureKind)
			d.FailureDetail = fact.FailureDetail
			s.Outbound[fact.MessageID] = d
		}
	case event.Type(metaengine.EventTypeName(domain.ThreadFlagSet{})):
		fact, err := event.DecodePayloadAuto[domain.ThreadFlagSet](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.thread_fold", "decode ThreadFlagSet")
		}
		switch fact.Flag {
		case "pinned":
			s.Pinned = fact.On
		case "archived":
			s.Archived = fact.On
		case "muted":
			s.Muted = fact.On
		}
	case event.Type(metaengine.EventTypeName(domain.ThreadUnreadCleared{})):
		s.Unread = 0
	case event.Type(metaengine.EventTypeName(domain.ThreadDeleted{})):
		s = threadState{}.initial()
	}
	return s, nil
}

func decideAppendMessage(s threadState, ver event.Version, cmd AppendMessage) ([]event.Event, error) {
	msgID := cmd.Msg.ID.String()
	for _, prev := range s.Messages {
		if prev == msgID {
			return nil, errorfamily.NewRejection("cqrs.thread_duplicate_message",
				fmt.Sprintf("message %s already on thread %s", msgID, cmd.Msg.ThreadID))
		}
	}
	// Creation legality: an empty stream adopts the message's thread
	// identity (AppendToThread); a live thread must agree with the message.
	if s.ThreadID != "" && s.ThreadID != cmd.Msg.ThreadID.String() {
		return nil, errorfamily.NewRejection("cqrs.thread_identity",
			fmt.Sprintf("message %s targets thread %s, stream holds %s", msgID, cmd.Msg.ThreadID, s.ThreadID))
	}
	fact := factMessage(cmd.Msg)
	evts := make([]event.Event, 0, 1+len(cmd.Msg.Attachments))
	appendEvt, err := event.New(event.Type(metaengine.EventTypeName(domain.MessageAppended{})),
		cmd.StreamID(), ThreadsStreamType, ver+1, fact)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.thread_append", "build MessageAppended")
	}
	evts = append(evts, appendEvt)
	for _, att := range cmd.Msg.Attachments {
		attEvt, attErr := event.New(event.Type(metaengine.EventTypeName(domain.AttachmentStored{})),
			cmd.StreamID(), ThreadsStreamType, ver+event.Version(len(evts))+1, domain.AttachmentStored{
				ID:        att.ID.String(),
				MessageID: att.MessageID.String(),
				Owner:     cmd.Msg.Owner.String(),
				Name:      att.Name,
				MimeType:  att.MimeType,
				SizeBytes: att.SizeBytes,
				Path:      att.Path,
			})
		if attErr != nil {
			return nil, errorfamily.WrapInfrastructuref(attErr, "cqrs.thread_append", "build AttachmentStored")
		}
		evts = append(evts, attEvt)
	}
	return evts, nil
}

func decideAdvanceDelivery(s threadState, ver event.Version, cmd AdvanceDelivery) ([]event.Event, error) {
	current, ok := s.Outbound[cmd.MessageID.String()]
	if !ok {
		return nil, errorfamily.NewRejection("cqrs.thread_message_missing",
			fmt.Sprintf("outbound message %s", cmd.MessageID))
	}
	next, moved := domain.DeliveryAdvance(current, cmd.Step.Event())
	if !moved {
		return nil, nil // the settled state answers; verdicts never land here while deliveryMoves stands
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.OutboundAdvanced{})),
		cmd.StreamID(), ThreadsStreamType, ver+1, domain.OutboundAdvanced{
			MessageID:     cmd.MessageID.String(),
			Status:        string(next.Status),
			ProviderRef:   next.ProviderRef,
			FailureKind:   string(next.FailureKind),
			FailureDetail: next.FailureDetail,
		})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.thread_advance", "build OutboundAdvanced")
	}
	return []event.Event{evt}, nil
}

func decideSetThreadFlag(s threadState, ver event.Version, cmd SetThreadFlag) ([]event.Event, error) {
	if s.ThreadID == "" {
		return nil, errorfamily.NewRejection("cqrs.thread_missing", "thread for flag "+cmd.Flag)
	}
	if !threadFlagLegal(cmd.Flag) {
		return nil, errorfamily.NewRejection("cqrs.thread_flag", "unknown thread flag "+cmd.Flag)
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.ThreadFlagSet{})),
		cmd.StreamID(), ThreadsStreamType, ver+1, domain.ThreadFlagSet{
			Owner:    cmd.Owner.String(),
			ThreadID: s.ThreadID,
			Flag:     cmd.Flag,
			On:       cmd.On,
		})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.thread_flag", "build ThreadFlagSet")
	}
	return []event.Event{evt}, nil
}

func decideMarkThreadRead(s threadState, ver event.Version, cmd MarkThreadRead) ([]event.Event, error) {
	if s.Unread == 0 {
		return nil, nil
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.ThreadUnreadCleared{})),
		cmd.StreamID(), ThreadsStreamType, ver+1, domain.ThreadUnreadCleared{
			Owner:    cmd.Owner.String(),
			ThreadID: s.ThreadID,
		})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.thread_read", "build ThreadUnreadCleared")
	}
	return []event.Event{evt}, nil
}

func decideDeleteThread(s threadState, ver event.Version, cmd DeleteThread) ([]event.Event, error) {
	if s.ThreadID == "" {
		return nil, nil // idempotent: a vanished thread needs no tombstone
	}
	deleted := domain.ThreadDeleted{
		Owner:         s.Owner,
		ThreadID:      s.ThreadID,
		Network:       s.Network,
		Remote:        s.Remote,
		MessageIDs:    append([]string(nil), s.Messages...),
		AttachmentIDs: append([]string(nil), s.AttachIDs...),
		Reason:        cmd.Reason,
	}
	evts := make([]event.Event, 0, 1+len(s.Messages)+len(s.AttachIDs))
	next := func() event.Version { return ver + event.Version(len(evts)) + 1 }
	tombstone, err := event.New(event.Type(metaengine.EventTypeName(domain.ThreadDeleted{})),
		cmd.StreamID(), ThreadsStreamType, next(), deleted)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.thread_delete", "build ThreadDeleted")
	}
	evts = append(evts, tombstone)
	for _, msgID := range s.Messages {
		rem, remErr := event.New(event.Type(metaengine.EventTypeName(domain.MessageRemoved{})),
			cmd.StreamID(), ThreadsStreamType, next(), domain.MessageRemoved{MessageID: msgID})
		if remErr != nil {
			return nil, errorfamily.WrapInfrastructuref(remErr, "cqrs.thread_delete", "build MessageRemoved")
		}
		evts = append(evts, rem)
	}
	for _, attID := range s.AttachIDs {
		rem, remErr := event.New(event.Type(metaengine.EventTypeName(domain.AttachmentRemoved{})),
			cmd.StreamID(), ThreadsStreamType, next(), domain.AttachmentRemoved{ID: attID})
		if remErr != nil {
			return nil, errorfamily.WrapInfrastructuref(remErr, "cqrs.thread_delete", "build AttachmentRemoved")
		}
		evts = append(evts, rem)
	}
	return evts, nil
}

// snapshotEveryNEvents is D8's conservative start: a phone thread is
// hundreds of events, not millions, and the state carries the full id
// lists — snapshots bound the fold cost; A38's perf data tunes the number.
const snapshotEveryNEvents = 500

// threadSnapshotStrategy is the resolved D8 strategy (EveryNEvents only
// rejects non-positive N, a compile-adjacent miswire — panic loudly).
var threadSnapshotStrategy = func() snapshot.SnapshotStrategy {
	strat, err := snapshot.EveryNEvents(snapshotEveryNEvents)
	if err != nil {
		panic(fmt.Sprintf("cqrs: thread snapshot strategy: %v", err))
	}
	return strat
}()

// RegisterThreads wires the thread aggregate into a System (D8: the
// conservative snapshot strategy bounds the id-list fold cost of
// long-lived threads).
func RegisterThreads(sys *system.System) error {
	if err := system.RegisterDecider[threadState](sys, ThreadsStreamType, decider.Decider[threadState]{
		Initial: threadState{}.initial(),
		Apply:   applyThreadState,
	}, system.WithSnapshotStrategy(threadSnapshotStrategy)); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread decider")
	}
	if err := registerThreadCommands(sys); err != nil {
		return err
	}
	return nil
}

func registerThreadCommands(sys *system.System) error {
	if err := system.RegisterCommand[*AppendMessage, threadState](sys, "thread.append",
		func(_ context.Context, cmd *AppendMessage) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideAppendMessage(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.append")
	}
	if err := system.RegisterCommand[*AdvanceDelivery, threadState](sys, "thread.advance_delivery",
		func(_ context.Context, cmd *AdvanceDelivery) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideAdvanceDelivery(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.advance_delivery")
	}
	if err := system.RegisterCommand[*SetThreadFlag, threadState](sys, "thread.set_flag",
		func(_ context.Context, cmd *SetThreadFlag) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideSetThreadFlag(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.set_flag")
	}
	if err := system.RegisterCommand[*MarkThreadRead, threadState](sys, "thread.mark_read",
		func(_ context.Context, cmd *MarkThreadRead) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideMarkThreadRead(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.mark_read")
	}
	if err := system.RegisterCommand[*DeleteThread, threadState](sys, "thread.delete",
		func(_ context.Context, cmd *DeleteThread) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideDeleteThread(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.delete")
	}
	return nil
}

// dispatchThreadCommand dispatches a thread command keyed by the thread's
// stream key (the dispatchOwnerCommand twin for a non-owner key).
func dispatchThreadCommand[C command.Command](ctx context.Context, sys *system.System, name, key string,
	build func(*command.BasicCommand) C,
) error {
	streamID, err := id.ParseStreamID(key)
	if err != nil {
		return errorfamily.WrapRejectionf(err, "cqrs.thread_stream", "parse thread stream key")
	}
	base, err := command.New(command.Type(name), streamID)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.command_build", "build %s", name)
	}
	return sys.CommandDispatcher().Dispatch(ctx, build(base))
}
