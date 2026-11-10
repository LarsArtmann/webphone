// Thread projections and domain declaration (SUPERB Phase 5): the
// system.db collections the server renders (threads, thread messages,
// message attachments) — full-image inserts, keyed removes, the
// declared filters and sorts the backend's scans ride.
package cqrs

import (
	"fmt"

	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// The three projection collections (query names in system.db).
const (
	threadsCollection  = "threads"
	messagesCollection = "thread_messages"
	attachmentsColName = "message_attachments"
)

// summaryViewOf converts the fact's row image into the read view
// (field-for-field; the decision computed it through the ONE homes).
func summaryViewOf(img domain.ThreadSummaryImage) readmodels.ThreadSummaryView {
	return readmodels.ThreadSummaryView{
		ThreadID:       img.ThreadID,
		Owner:          img.Owner,
		Network:        img.Network,
		Remote:         img.Remote,
		LastActivity:   img.LastActivity,
		Unread:         img.Unread,
		Pinned:         img.Pinned,
		Archived:       img.Archived,
		Muted:          img.Muted,
		LastBody:       img.LastBody,
		LastDirection:  img.LastDirection,
		LastChannel:    img.LastChannel,
		LastAttachment: img.LastAttachment,
	}
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

// The collections' key shapes are the domain row-key types (type-matched
// by the remove/update key inference — every keyed fold's fact carries
// its Key field).

// threadProjections declares the three collections. Update folds (the
// prev-aware shape) carry only their delta — the row's other fields
// survive through prev, exactly as the SQL UPDATE touched one column;
// removes are keyed by the fact's id field.
func threadProjections() []system.ProjectionDeclaration {
	threads := metaengine.Query[domain.ThreadRowKey, readmodels.ThreadSummaryView](threadsCollection,
		metaengine.OnRecord(domain.MessageAppended{},
			func(_ record.Record, e domain.MessageAppended) (domain.ThreadRowKey, readmodels.ThreadSummaryView) {
				return domain.ThreadRowKey{ThreadID: e.ThreadID}, summaryViewOf(e.Summary)
			}),
		metaengine.OnRecord(domain.ThreadFlagSet{},
			func(_ record.Record, e domain.ThreadFlagSet) (domain.ThreadRowKey, readmodels.ThreadSummaryView) {
				return e.Key, summaryViewOf(e.Summary)
			}),
		metaengine.OnRecord(domain.ThreadUnreadCleared{},
			func(_ record.Record, e domain.ThreadUnreadCleared) (domain.ThreadRowKey, readmodels.ThreadSummaryView) {
				return e.Key, summaryViewOf(e.Summary)
			}),
		metaengine.OnRecord(domain.ThreadDeleted{}, metaengine.Remove[readmodels.ThreadSummaryView]()),
		metaengine.FilterOnField[readmodels.ThreadSummaryView]("owner", metaengine.FilterEq),
	)

	messages := metaengine.Query[domain.MessageRowKey, readmodels.MessageView](messagesCollection,
		metaengine.OnRecord(domain.MessageAppended{},
			func(_ record.Record, e domain.MessageAppended) (domain.MessageRowKey, readmodels.MessageView) {
				return domain.MessageRowKey{MessageID: e.MessageID}, messageViewOf(e)
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
		metaengine.FilterOnField[readmodels.MessageView]("thread_id", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.MessageView]("created_at", true),
	)

	attachments := metaengine.Query[domain.AttachmentRowKey, readmodels.AttachmentView](attachmentsColName,
		metaengine.OnRecord(domain.AttachmentStored{},
			func(_ record.Record, e domain.AttachmentStored) (domain.AttachmentRowKey, readmodels.AttachmentView) {
				return domain.AttachmentRowKey{ID: e.ID}, readmodels.AttachmentView{
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
				panic(fmt.Sprintf("cqrs: register thread aggregate: %v", err)) //cqrs-lint:ignore(C009) boot-time fail-fast on a composition miswire (DO-1; the decl comment owns the why)
			}
		},
		Projections: threadProjections(),
		Events: threadStreamEventTypes,
	}
}
