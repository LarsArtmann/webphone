package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/textagent"
)

// reviewDispatchResponse is the 202 answer: the automatic review job
// was accepted (or already exists — duplicate names the replay).
type reviewDispatchResponse struct {
	Dispatched bool `description:"True when the review job was accepted." json:"dispatched"`
	Duplicate  bool `description:"True when this call already has a review (done or in flight)." json:"duplicate,omitempty"`
}

// DispatchCallReview accepts the post-settle AUTOMATIC review ask
// (issue #8): the island fires it once when a media call settles, and
// the job runs WITHOUT the request — context.WithoutCancel plus its own
// budget, so closing the tab cannot abort a review the operator asked
// for by making a call. Contract:
//
//   - 202 with the job accepted (or duplicate:true when the idem window
//     or an in-flight guard already covers this call — the manual
//     summarize button shares the idempotency key, so exactly one of
//     them ever drafts).
//   - 404 when the seam is off (disabled-seam convention) or the call
//     has no transcript for THIS extension (owner-scoped, fast-fail
//     before the dispatch — garbage never enters the queue).
//   - 503 when the transcript store is nil (test compositions).
//   - Failure is non-blocking by design: the job logs, audits, and
//     leaves the idempotency key unrecorded, so the manual button (or a
//     re-settle) stays retryable.
func (h *Handlers) DispatchCallReview(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return
	}
	if h.deps.TextAgent == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if h.deps.Transcripts == nil {
		http.Error(w, "transcript store unavailable", http.StatusServiceUnavailable)
		return
	}
	callID := r.PathValue("id")
	if callID == "" || len(callID) > summaryCallIDMax || strings.ContainsAny(callID, "/%") {
		http.Error(w, "invalid call id", http.StatusBadRequest)
		return
	}

	idemKey := sess.Extension.String() + ":summary:" + callID
	if h.deps.CallsIdem.Seen(idemKey) {
		writeJSONBody(w, reviewDispatchResponse{Dispatched: true, Duplicate: true})
		return
	}
	// In-flight guard: the idem window records successes only, so two
	// dispatches racing (auto settle + a manual click) would both pass
	// Seen and draft twice. The guard holds the slot for the job's whole
	// run; the job releases it on every exit path.
	if _, busy := h.reviewInFlight.LoadOrStore(idemKey, struct{}{}); busy {
		writeJSONBody(w, reviewDispatchResponse{Dispatched: true, Duplicate: true})
		return
	}

	transcript, err := h.deps.Transcripts.Call(r.Context(), sess.Extension, callID)
	if err != nil {
		slog.Warn("review: transcript fetch failed", "error", err, "extension", sess.Extension.String())
		http.Error(w, "could not read the call transcript", http.StatusInternalServerError)
		h.reviewInFlight.Delete(idemKey)
		return
	}
	if transcript == nil || len(transcript.Lines) == 0 {
		http.Error(w, "no transcript for that call", http.StatusNotFound)
		h.reviewInFlight.Delete(idemKey)
		return
	}

	jobCtx := context.WithoutCancel(r.Context())
	go h.runReviewJob(jobCtx, sess.Extension, callID, transcript.Remote, strings.Join(transcript.Lines, "\n"), idemKey)
	writeJSONBody(w, reviewDispatchResponse{Dispatched: true})
}

// runReviewJob is the detached automatic review: draft from the
// transcript exactly like the manual button's synchronous path (ONE
// drafting path — the job adds no second vocabulary), then record idem
// and fan the draft-ready push. Every exit releases the in-flight slot;
// a panic is contained (the goroutine doctrine) and audited as failed.
func (h *Handlers) runReviewJob(ctx context.Context, owner domain.Extension, callID, remote, text, idemKey string) {
	defer h.reviewInFlight.Delete(idemKey)
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("review: job PANIC recovered; goroutine contained, process survives",
				"panic", rec, "call", callID, "extension", owner.String())
			agentAudit("auto_review", owner.String(), callID, "failed")
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, summaryDraftTimeout)
	defer cancel()

	draft, err := h.deps.TextAgent.DraftFromTranscript(ctx, remote, text)
	if err != nil {
		outcome := "failed"
		if agentErr, isAgentErr := errors.AsType[*textagent.Error](err); isAgentErr {
			slog.Warn("review: agent refused the automatic draft",
				"status", agentErr.Status, "detail", agentErr.Message, "extension", owner.String())
			outcome = "refused"
		} else {
			wrapped := errorfamily.WrapTransientf(err, "agent.summarize", "automatic summarize call %s", callID)
			slog.Warn("review: agent unreachable for the automatic draft", "error", wrapped, "extension", owner.String())
		}
		agentAudit("auto_review", owner.String(), callID, outcome)
		return
	}

	h.recordSummaryIdem(idemKey)
	agentAudit("auto_review", owner.String(), draft.ID, "ok")
	h.notifyDraftReady(ctx, owner, draft.ID, remote)
}
