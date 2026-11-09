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
	"encoding/hex"
	"fmt"
	"slices"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	errorfamily "github.com/larsartmann/go-error-family"

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
	Owner   domain.Extension
	Network domain.Network
	Remote  domain.Address
	Flag    string // "pinned" | "archived" | "muted"
	On      bool
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

func decideAppendMessage(s threadState, ver event.Version, cmd AppendMessage) ([]event.Event, error) {
	msgID := cmd.Msg.ID.String()
	if slices.Contains(s.Messages, msgID) {
		return nil, errorfamily.NewRejection("cqrs.thread_duplicate_message",
			fmt.Sprintf("message %s already on thread %s", msgID, cmd.Msg.ThreadID))
	}
	// Creation legality: an empty stream adopts the message's thread
	// identity (AppendToThread); a live thread must agree with the message.
	if s.ThreadID != "" && s.ThreadID != cmd.Msg.ThreadID.String() {
		return nil, errorfamily.NewRejection("cqrs.thread_identity",
			fmt.Sprintf("message %s targets thread %s, stream holds %s", msgID, cmd.Msg.ThreadID, s.ThreadID))
	}
	state := s
	state = appendToState(state, cmd.Msg)
	fact := factMessage(cmd.Msg)
	fact.Key = domain.ThreadRowKey{ThreadID: fact.ThreadID}
	fact.Summary = state.summaryImage()
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
			Key:           domain.MessageRowKey{MessageID: cmd.MessageID.String()},
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
	next := s
	switch cmd.Flag {
	case "pinned":
		next.Pinned = cmd.On
	case "archived":
		next.Archived = cmd.On
	case "muted":
		next.Muted = cmd.On
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.ThreadFlagSet{})),
		cmd.StreamID(), ThreadsStreamType, ver+1, domain.ThreadFlagSet{
			Key:      domain.ThreadRowKey{ThreadID: s.ThreadID},
			Summary:  next.summaryImage(),
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
	next := s
	next.Unread = 0
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.ThreadUnreadCleared{})),
		cmd.StreamID(), ThreadsStreamType, ver+1, domain.ThreadUnreadCleared{
			Key:      domain.ThreadRowKey{ThreadID: s.ThreadID},
			Summary:  next.summaryImage(),
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
		Key:           domain.ThreadRowKey{ThreadID: s.ThreadID},
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
			cmd.StreamID(), ThreadsStreamType, next(), domain.MessageRemoved{
				Key:       domain.MessageRowKey{MessageID: msgID},
				MessageID: msgID,
			})
		if remErr != nil {
			return nil, errorfamily.WrapInfrastructuref(remErr, "cqrs.thread_delete", "build MessageRemoved")
		}
		evts = append(evts, rem)
	}
	for _, attID := range s.AttachIDs {
		rem, remErr := event.New(event.Type(metaengine.EventTypeName(domain.AttachmentRemoved{})),
			cmd.StreamID(), ThreadsStreamType, next(), domain.AttachmentRemoved{
				Key: domain.AttachmentRowKey{ID: attID},
				ID:  attID,
			})
		if remErr != nil {
			return nil, errorfamily.WrapInfrastructuref(remErr, "cqrs.thread_delete", "build AttachmentRemoved")
		}
		evts = append(evts, rem)
	}
	return evts, nil
}
