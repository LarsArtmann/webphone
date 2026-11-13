package readmodels

import (
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// SummaryOfView re-parses a system.db projection view into the render
// shape (raw-at-rest wire, branded-at-gate — the doctrine the store's
// scans performed before the journal swap).
func SummaryOfView(v ThreadSummaryView) ThreadSummary {
	return ThreadSummary{
		Thread: domain.Thread{
			ID:             domain.MustThreadID(v.ThreadID),
			Owner:          domain.MustParseExtension(v.Owner),
			Remote:         domain.MustParseAddress(domain.MustNetwork(v.Network), v.Remote),
			LastActivityAt: time.Unix(0, v.LastActivity),
			Unread:         v.Unread,
			Pinned:         v.Pinned,
			Archived:       v.Archived,
			Muted:          v.Muted,
		},
		LastBody:       v.LastBody,
		LastDirection:  domain.MustDirection(v.LastDirection),
		LastChannel:    domain.MustChannel(v.LastChannel),
		LastAttachment: v.LastAttachment,
	}
}

// SummariesOfView converts a view slice wholesale.
func SummariesOfView(views []ThreadSummaryView) []ThreadSummary {
	out := make([]ThreadSummary, 0, len(views))
	for _, v := range views {
		out = append(out, SummaryOfView(v))
	}
	return out
}
