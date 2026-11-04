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

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-retry"
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

// firstWait is the reconnect ladder's floor: every receiver starts a
// fresh failure streak here. A var (not a const) so tests can shrink it.
var firstWait = time.Second

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

// nextDelay computes attempt n's wait: go-retry's exponential doubling
// with up to 50% additive jitter, hard-capped at max — the fleet
// library's hardened math replaces the hand-rolled deterministic
// ladder (thundering-herd prone when every lane reconnects after a
// shared-server restart). attempt must be >= 1.
func nextDelay(attempt int, max time.Duration) time.Duration {
	delay, err := retry.ComputeDelay(firstWait, max, 2, attempt)
	if err != nil {
		return max // unreachable for attempt >= 1; a caller bug waits the cap
	}
	return delay
}

// Default resolves a nil logger to the slog default — the receivers'
// constructor guard, one home instead of a per-package copy.
func Default(log *slog.Logger) *slog.Logger {
	if log == nil {
		return slog.Default()
	}
	return log
}

// Backoff is the receivers' shared failure step. A Rejection stops the
// loop LOUDLY (retrying changes nothing — a revoked token, a daemon
// that refuses the method); every other family logs the reason, waits
// out the attempt's jittered delay, and arms the next attempt. false
// means the caller exits its loop — the context ended during the wait,
// or the failure is permanent.
func Backoff(ctx context.Context, log *slog.Logger, message string, err error, attempt int, max time.Duration) (int, bool) {
	if errorfamily.Classify(err) == errorfamily.Rejection {
		log.Error(message+" (permanent; receiver stops)", "error", err)
		return attempt, false
	}
	log.Warn(message, "error", err)
	if !Sleep(ctx, nextDelay(attempt, max)) {
		return attempt, false
	}
	return attempt + 1, true
}
