// Thread ADMIN seams (SUPERB Phase 5): the deliberate cross-owner
// reads — the admin inventory list and the any-owner thread and
// attachment lookups (the read-only Admin surface's fallback lane,
// pinned by store/readmodels admin_reads twins).
package cqrs

import (
	"context"
	"errors"
	"sort"

	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

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
