package readmodels

import (
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// SummaryOfView converts a system.db projection view into the summary
// shape the store's public contract has always carried (the ES
// migration's read path answers through the same type — one shape, two
// stores, field-for-field).
func SummaryOfView(v ThreadSummaryView) store.ThreadSummary {
	return store.ThreadSummary{
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
func SummariesOfView(views []ThreadSummaryView) []store.ThreadSummary {
	out := make([]store.ThreadSummary, 0, len(views))
	for _, v := range views {
		out = append(out, SummaryOfView(v))
	}
	return out
}
