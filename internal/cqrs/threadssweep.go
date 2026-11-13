// Thread deletion and retention (SUPERB Phase 5): the tombstone the
// decider emits, the journal trim that erases the stream's bytes, and
// the SweepBefore leg the retention pass drives (blob paths out for
// unlinking — only the swept thread's own).
package cqrs

import (
	"context"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

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

// SweepBefore is the retention leg (A36): every thread whose last
// activity precedes the cutoff is tombstoned, folded out of the views,
// and journal-trimmed (the legacy sweep's "empty thread" twin — a
// last-activity-older thread has no message that survives the cutoff,
// so the two semantics coincide). Attachment blob paths are gathered
// BEFORE the rows fold away and returned for the CALLER to unlink (the
// store layer only ever owned rows; the retention pass owns files).
// The Faxes.SweepBefore shape, one seam for both legs.
func (s *Threads) SweepBefore(ctx context.Context, cutoff time.Time) ([]string, int64, error) {
	views, err := s.threads.Scan(ctx, metaengine.WithLimit(0))
	if err != nil {
		return nil, 0, errorfamily.WrapInfrastructuref(err, "cqrs.threads_sweep_scan", "sweep scan threads")
	}
	cutoffNanos := cutoff.UnixNano()
	var paths []string
	var removed int64
	for _, v := range views {
		if v.LastActivity >= cutoffNanos {
			continue
		}
		owner := domain.MustParseExtension(v.Owner)
		threadID := domain.MustThreadID(v.ThreadID)
		msgs, err := s.threadMessages(ctx, owner, threadID)
		if err != nil {
			return paths, removed, err
		}
		if err := s.attachmentsPaths(ctx, msgs, &paths); err != nil {
			return paths, removed, err
		}
		if err := s.DeleteThread(ctx, owner, threadID, "retention"); err != nil {
			return paths, removed, err
		}
		removed++
	}
	return paths, removed, nil
}

// attachmentsPaths folds the owners' attachment rows onto the path
// list for the given messages ONLY (paths gathered before the tombstone
// folds the rows away; a sibling thread's blobs are not this sweep's).
func (s *Threads) attachmentsPaths(ctx context.Context, msgs []readmodels.MessageView, paths *[]string) error {
	ofThread := make(map[string]struct{}, len(msgs))
	byOwner := make(map[string]struct{}, len(msgs))
	for _, m := range msgs {
		ofThread[m.MessageID] = struct{}{}
		byOwner[m.Owner] = struct{}{}
	}
	for owner := range byOwner {
		views, err := s.attachs.Scan(ctx,
			metaengine.WithFilter("owner", metaengine.FilterEq, owner),
			metaengine.WithLimit(0))
		if err != nil {
			return errorfamily.WrapInfrastructuref(err, "cqrs.attachments_sweep", "list attachments of %s", owner)
		}
		for _, v := range views {
			if _, ok := ofThread[v.MessageID]; ok {
				*paths = append(*paths, v.Path)
			}
		}
	}
	return nil
}
