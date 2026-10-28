// Package chatloop is the shared skeleton of the chat-network
// receivers (matrix sync, signal socket): the inbound sink contract
// both feed, and the backoff-bounded loop helpers both run. One home,
// two consumers — the receivers stay wire-specific, this package stays
// wire-free.
package chatloop

import (
	"context"
	"log/slog"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// Sink is the inbound pipeline the receivers feed (the messaging
// service satisfies it: persist + notify + auto-unarchive, wholesale).
// ONE declaration: the matrix and signal receivers share the type, not
// a copy of it.
type Sink interface {
	Receive(ctx context.Context, inbound domain.InboundMessage) (domain.Message, error)

	// RemoteRead advances the outbound message the provider ref names to
	// delivered: a peer's read marker (Matrix m.read) is the strongest
	// delivery fact the network offers, so it rides the SAME delivery
	// machine as provider verdicts. An uncorrelated ref answers
	// store.ErrNotFound for the receiver to treat as benign.
	RemoteRead(ctx context.Context, providerRef string) error
}

// Sleep waits d or until ctx ends; false means ctx ended first. The
// receivers' backoff waits ride it so shutdown never blocks on a nap.
func Sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// NextBackoff doubles the wait, capped at max.
func NextBackoff(d, max time.Duration) time.Duration {
	d *= 2
	if d > max {
		return max
	}
	return d
}

// Default resolves a nil logger to the slog default — the receivers'
// constructor guard, one home instead of a per-package copy.
func Default(log *slog.Logger) *slog.Logger {
	if log == nil {
		return slog.Default()
	}
	return log
}

// Backoff is the receivers' shared failure step: log the reason, wait
// out the current backoff, and double it (capped). false means ctx
// ended during the wait — the caller exits its loop.
func Backoff(ctx context.Context, log *slog.Logger, message string, err error, wait, max time.Duration) (time.Duration, bool) {
	log.Warn(message, "error", err)
	if !Sleep(ctx, wait) {
		return wait, false
	}
	return NextBackoff(wait, max), true
}
