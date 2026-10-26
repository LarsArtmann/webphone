package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/textagent"
)

// summaryDraftTimeout bounds one summarize ask: a synchronous LLM
// round trip inside the agent, the same generation-sized budget the
// on-demand compose action rides (the read-path 5 s ceiling would
// abort nearly every real summary).
const summaryDraftTimeout = 95 * time.Second

// summaryCallIDMax bounds the opaque island Call-ID before it reaches
// the store scan (a malformed path is a client bug, not a lookup).
const summaryCallIDMax = 128

// SummaryDraftRequest is the POST /api/calls/{id}/summary-draft wire
// shape — the call id rides the path, the body is empty. Named (not
// inline) so the machine-API contract derives its schema from this
// type (internal/server/apidocs.go).
type SummaryDraftRequest struct{}

// SummaryDraftResponse is the 200 answer: the agent's draft, already
// in the review queue. Duplicate reports the idempotency window hit
// (a double-click or a retry whose first attempt succeeded): the
// island toasts "already drafted" and points at Messages.
type SummaryDraftResponse struct {
	ID        string `description:"The draft's review id (empty when duplicate)." json:"id"`
	Contact   string `description:"The number the draft addresses." json:"contact"`
	Text      string `description:"The drafted follow-up text (empty when duplicate)." json:"text"`
	Duplicate bool   `description:"True when this call was already summarized in the idempotency window." json:"duplicate,omitempty"`
}

// SummaryDraft turns one call's transcript into an AI follow-up draft
// in the review queue (the post-call summarize chain). Contract:
//
//   - 200 with the draft (or duplicate:true for an idempotent replay).
//   - 404 when the call has no transcript for THIS extension (another
//     owner's call id is the same 404 — owner scoping at the store).
//   - 404 when the textagent seam is off (the disabled-seam convention).
//   - 503 when the transcript store is nil (test compositions).
//   - 502/agent-status on agent failure; the idempotency key records
//     only on success, so a failed attempt stays retryable.
func (h *Handlers) SummaryDraft(w http.ResponseWriter, r *http.Request) {
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
		writeJSONBody(w, SummaryDraftResponse{Duplicate: true})
		return
	}

	transcript, err := h.deps.Transcripts.Call(r.Context(), sess.Extension, callID)
	if err != nil {
		slog.Warn("summary: transcript fetch failed", "error", err, "extension", sess.Extension.String())
		http.Error(w, "could not read the call transcript", http.StatusInternalServerError)
		return
	}
	if transcript == nil || len(transcript.Lines) == 0 {
		http.Error(w, "no transcript for that call", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), summaryDraftTimeout)
	defer cancel()
	draft, err := h.deps.TextAgent.DraftFromTranscript(ctx, transcript.Remote, strings.Join(transcript.Lines, "\n"))
	if err != nil {
		var agentErr *textagent.Error
		if errors.As(err, &agentErr) {
			slog.Warn("summary: agent refused the draft",
				"status", agentErr.Status, "detail", agentErr.Message, "extension", sess.Extension.String())
			http.Error(w, agentErr.Message, agentErr.Status)
			return
		}
		// The agent's transport wrapped its own codes; this seam's own
		// wrap names the summarize op (the family pin rides the pins file).
		wrapped := errorfamily.WrapTransientf(err, "agent.summarize", "summarize call %s", callID)
		slog.Warn("summary: agent unreachable", "error", wrapped, "extension", sess.Extension.String())
		http.Error(w, "the AI agent is unreachable right now", http.StatusBadGateway)
		return
	}

	h.recordSummaryIdem(idemKey)
	writeJSONBody(w, SummaryDraftResponse{ID: draft.ID, Contact: draft.Contact, Text: draft.Text})
}

// recordSummaryIdem marks a summarize attempt processed (successes
// only — a failure must stay retryable).
func (h *Handlers) recordSummaryIdem(key string) {
	h.deps.CallsIdem.Record(key)
}
