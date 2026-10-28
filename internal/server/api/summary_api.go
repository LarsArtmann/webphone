package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/push"
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

// draftPushTimeout bounds one fire-and-forget draft fan-out: the
// summarize answer never waits on push health, so the goroutine owns
// its own budget (the transports' own timeouts sit under it).
const draftPushTimeout = 15 * time.Second

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
		agentErr, isAgentErr := errors.AsType[*textagent.Error](err)
		if isAgentErr {
			slog.Warn("summary: agent refused the draft",
				"status", agentErr.Status, "detail", agentErr.Message, "extension", sess.Extension.String())
			http.Error(w, agentErr.Message, agentErr.Status)
			h.recordReviewArtifact(sess.Extension, callID, domain.ReviewFailed, "", "", "", 0)
			return
		}
		// The agent's transport wrapped its own codes; this seam's own
		// wrap names the summarize op (the family pin rides the pins file).
		wrapped := errorfamily.WrapTransientf(err, "agent.summarize", "summarize call %s", callID)
		slog.Warn("summary: agent unreachable", "error", wrapped, "extension", sess.Extension.String())
		http.Error(w, "the AI agent is unreachable right now", http.StatusBadGateway)
		h.recordReviewArtifact(sess.Extension, callID, domain.ReviewFailed, "", "", "", 0)
		return
	}

	h.recordSummaryIdem(idemKey)
	h.recordReviewArtifact(sess.Extension, callID, domain.ReviewPresent, draft.ID, draft.Text, draft.Language, draft.Confidence)
	h.notifyDraftReady(r.Context(), sess.Extension, draft.ID, transcript.Remote)
	writeJSONBody(w, SummaryDraftResponse{ID: draft.ID, Contact: draft.Contact, Text: draft.Text})
}

// notifyDraftReady fans the draft-ready fact out to the owner's
// registered push devices — fire-and-forget: the summarize answer must
// never wait on push health (the ring hook's 202-inert doctrine), and
// the disabled seam is simply nothing. The contact is the CALL's remote
// (the handler resolved it from the owner-scoped transcript — the one
// truth), not the agent's echo; a transport reporting the device dead
// prunes the registration on the spot, exactly like the ring lane.
func (h *Handlers) notifyDraftReady(ctx context.Context, owner domain.Extension, draftID, contact string) {
	if h.deps.Push == nil || !h.deps.Push.Enabled() || h.deps.Devices == nil {
		return
	}
	ctx = context.WithoutCancel(ctx)
	go func() {
		// Goroutine containment (owner ruling 2026-10-27): RecoveryMiddleware
		// cannot see this goroutine, so a panicking transport or registry
		// would take the whole process down. Recover, log greppably, survive.
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("push: draft fan-out PANIC recovered; goroutine contained, process survives",
					"panic", rec, "draft", draftID, "extension", owner.String())
			}
		}()
		ctx, cancel := context.WithTimeout(ctx, draftPushTimeout)
		defer cancel()
		fact := push.Draft{Event: push.EventDraftReady, Contact: contact, DraftID: draftID}
		targets, err := h.deps.Devices.List(ctx, owner)
		if err != nil {
			wrapped := errorfamily.WrapInfrastructuref(err, "push.draft_devices", "list draft push targets")
			slog.Warn("push: draft fan-out aborted", "error", wrapped, "extension", owner.String())
			return
		}
		for _, target := range targets {
			if err := h.deps.Push.SendDraft(ctx, target.Transport, target.Token, fact); err != nil {
				if errors.Is(err, push.ErrUnregistered) {
					if rmErr := h.deps.Devices.Remove(ctx, owner, target.DeviceID); rmErr != nil {
						slog.Warn("push: prune dead device failed", "device", target.DeviceID, "err", rmErr)
					} else {
						slog.Info("push: pruned dead device", "device", target.DeviceID, "transport", target.Transport)
					}
					continue
				}
				wrapped := errorfamily.WrapInfrastructuref(err, "push.draft_send", "ship draft-ready to %s", target.DeviceID)
				slog.Warn("push: draft fan-out failed", "error", wrapped, "device", target.DeviceID, "transport", target.Transport)
			}
		}
	}()
}

// recordSummaryIdem marks a summarize attempt processed (successes
// only — a failure must stay retryable).
func (h *Handlers) recordSummaryIdem(key string) {
	h.deps.CallsIdem.Record(key)
}
