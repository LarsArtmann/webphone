// Thread READ gates (SUPERB Phase 5): the list/search/page/count
// family and the attachment lookups the panels, the SSE fragments
// and the metrics surface call — every read a projection scan with
// its ceiling LIFTED (WithLimit(0)) or bounded by the caller.
package cqrs

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

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
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithLimit(0))
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
// first (the raw engine order the callers page/reverse from). The
// thread_id filter is the narrow index; WithLimit(0) lifts the engine's
// default 100-row ceiling (the paging contract needs every row).
func (s *Threads) threadMessages(ctx context.Context, owner domain.Extension, threadID domain.ThreadID) ([]readmodels.MessageView, error) {
	views, err := s.messages.Scan(ctx,
		metaengine.WithFilter("thread_id", metaengine.FilterEq, threadID.String()),
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("created_at", true),
		metaengine.WithLimit(0))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.thread_messages", "list messages of thread %s", threadID)
	}
	return views, nil
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
			metaengine.WithFilter("owner", metaengine.FilterEq, owner),
			metaengine.WithLimit(0))
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

// CountMessages reports the total stored messages via the engine's
// native COUNT.
func (s *Threads) CountMessages(ctx context.Context) (int64, error) {
	n, err := engineCount(ctx, s.sys, messagesCollection)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "cqrs.message_count", "count messages")
	}
	return n, nil
}

// CountThreads reports the total stored threads via the engine's
// native COUNT (the metrics surface's webphone_threads_total).
func (s *Threads) CountThreads(ctx context.Context) (int64, error) {
	n, err := engineCount(ctx, s.sys, threadsCollection)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "cqrs.thread_count", "count threads")
	}
	return n, nil
}
