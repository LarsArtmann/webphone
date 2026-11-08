// Thread-aggregate STATE (SUPERB Phase 5): the decider's fold — the
// thread row image, the id bookkeeping the deletes and advances
// need, and the apply/adopt twins the replay tests pin. Commands
// and decisions live in threads.go.
package cqrs

import (
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// threadState is the full aggregate: the legacy thread row plus the id
// bookkeeping the deletes and delivery advances need.
type threadState struct {
	ThreadID   string
	Owner      string
	Network    string
	Remote     string
	LastAct    int64 // unix nanos
	Unread     int
	Pinned     bool
	Archived   bool
	Muted      bool
	MastBody   string // last-by-created-at preview body
	MastDir    string
	MastChan   string
	MastAttach bool
	Messages   []string // every appended message id, journal order
	Outbound   map[string]domain.Delivery
	AttachIDs  []string // every attachment id, journal order
}

// summaryImage renders the row image the facts carry whole.
func (s threadState) summaryImage() domain.ThreadSummaryImage {
	return domain.ThreadSummaryImage{
		ThreadID:       s.ThreadID,
		Owner:          s.Owner,
		Network:        s.Network,
		Remote:         s.Remote,
		LastActivity:   s.LastAct,
		Unread:         s.Unread,
		Pinned:         s.Pinned,
		Archived:       s.Archived,
		Muted:          s.Muted,
		LastBody:       s.MastBody,
		LastDirection:  s.MastDir,
		LastChannel:    s.MastChan,
		LastAttachment: s.MastAttach,
	}
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
		// The decision computed the row image through domain.AppendToThread
		// (the ONE home); the apply adopts it and tracks the ids the later
		// decisions need. appendToState (the decision twin) and this branch
		// are one rule — the battery's replay tests pin the equality.
		msg := messageDomain(fact)
		adoptSummaryImage(&s, fact.Summary)
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
		adoptSummaryImage(&s, fact.Summary)
	case event.Type(metaengine.EventTypeName(domain.ThreadUnreadCleared{})):
		fact, err := event.DecodePayloadAuto[domain.ThreadUnreadCleared](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.thread_fold", "decode ThreadUnreadCleared")
		}
		adoptSummaryImage(&s, fact.Summary)
	case event.Type(metaengine.EventTypeName(domain.ThreadDeleted{})):
		s = threadState{}.initial()
	}
	return s, nil
}

// appendToState is applyThreadState's MessageAppended transition as a
// pure function over the state (the image the fact carries must equal
// what the apply will fold — one rule, two sights, pinned by the
// battery's replay tests).
func appendToState(s threadState, msg domain.Message) threadState {
	thread := domain.AppendToThread(s.threadDomain(), msg)
	s.ThreadID, s.Owner, s.Network, s.Remote = thread.ID.String(), thread.Owner.String(),
		string(thread.Remote.Network()), thread.Remote.String()
	s.LastAct, s.Unread, s.Archived = thread.LastActivityAt.UnixNano(), thread.Unread, thread.Archived
	// Journal order is creation order for both the live lane (CreatedAt =
	// now, monotonic) and the sorted importer, so the appended message is
	// the newest — the legacy list's ORDER BY created_at DESC twin.
	s.MastBody, s.MastDir, s.MastChan, s.MastAttach = msg.Body,
		string(msg.Direction), string(msg.Channel), len(msg.Attachments) > 0
	s.Messages = append(s.Messages, msg.ID.String())
	if msg.Direction == domain.DirectionOutbound {
		s.Outbound[msg.ID.String()] = msg.DeliveryOf()
	}
	return s
}

// adoptSummaryImage writes a fact's row image into the state.
func adoptSummaryImage(s *threadState, img domain.ThreadSummaryImage) {
	s.ThreadID, s.Owner, s.Network, s.Remote = img.ThreadID, img.Owner, img.Network, img.Remote
	s.LastAct, s.Unread = img.LastActivity, img.Unread
	s.Pinned, s.Archived, s.Muted = img.Pinned, img.Archived, img.Muted
	s.MastBody, s.MastDir, s.MastChan, s.MastAttach = img.LastBody, img.LastDirection, img.LastChannel, img.LastAttachment
}
