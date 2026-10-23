package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/textagent"
	"github.com/larsartmann/webphone/internal/web/views"
)

// textAgentPanelTimeout bounds the review-API READS a page render or
// panel action makes. The agent is a loopback neighbor, so a healthy
// exchange is milliseconds; an unreachable one must degrade the panel
// to its honest error state instead of stalling the Messages tab for
// the client's full 60 s transport timeout.
const textAgentPanelTimeout = 5 * time.Second

// textAgentDraftTimeout bounds an ON-DEMAND draft, which is a
// synchronous LLM round trip inside the agent (its Gemini client runs a
// 60 s window, and a standby model adds a short attempt on top). The
// read-path 5 s ceiling above would abort nearly every real draft — the
// 2026-10-23 "AI draft is broken" class — so the compose action gets its
// own, generation-sized budget.
const textAgentDraftTimeout = 95 * time.Second

// textAgentOn is the one gate every textagent surface reads (same rule
// as asrOn): a nil client means the seam is off — the routes answer the
// styled 404 and the views render neither the drafts panel nor the
// compose buttons.
func (h *handlers) textAgentOn() bool { return h.deps.TextAgent != nil }

// textAgentReviews fetches the review queue under the panel timeout and
// maps any failure onto the panel's honest error line (translated
// headline + the agent's own phrased detail — never a re-guessed
// generic).
func (h *handlers) textAgentReviews(r *http.Request) (*textagent.Reviews, string) {
	if !h.textAgentOn() {
		return nil, ""
	}
	ctx, cancel := context.WithTimeout(r.Context(), textAgentPanelTimeout)
	defer cancel()
	reviews, err := h.deps.TextAgent.Reviews(ctx)
	if err != nil {
		return nil, h.T(r, "drafts.unavailable") + " " + agentDetail(err)
	}
	return reviews, ""
}

// agentDetail surfaces the review API's own human-phrased message when
// the failure is an agent verdict, and the wrapped error text otherwise
// (transport noise names the unreachable base; the owner surface stays
// honest either way).
func agentDetail(err error) string {
	if agentErr, ok := errors.AsType[*textagent.Error](err); ok {
		return agentErr.Message
	}
	return err.Error()
}

// partialDrafts answers the drafts panel's own 30 s poll: the full
// <section> for an outerHTML swap, so the review loop refreshes without
// re-rendering the whole Messages tab (a half-typed composer survives).
func (h *handlers) partialDrafts(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.RequireSession(w, r); !ok {
		return
	}
	if !h.textAgentOn() {
		http.NotFound(w, r)
		return
	}
	h.draftPanel(w, r)
}

// draftContactNames resolves the drafts' contact numbers against the
// optional CRM feed so the review list shows names like the thread rows
// do. Nil-safe on every leg (no drafts, no resolver).
func (h *handlers) draftContactNames(r *http.Request, reviews *textagent.Reviews) map[string]string {
	if reviews == nil || len(reviews.Drafts) == 0 {
		return nil
	}
	numbers := make([]string, 0, len(reviews.Drafts))
	for _, draft := range reviews.Drafts {
		numbers = append(numbers, draft.Contact)
	}
	return h.crmNames(r.Context(), numbers)
}

// draftPanel renders the drafts panel fragment (approve and discard
// both swap #wp-drafts outerHTML with the fresh list).
func (h *handlers) draftPanel(w http.ResponseWriter, r *http.Request) {
	reviews, fetchErr := h.textAgentReviews(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.DraftsPanel(views.DraftsPanelProps{
		Drafts: reviews, Error: fetchErr, Names: h.draftContactNames(r, reviews), Lang: h.lang(r),
	}).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// textAgentActionError maps a failed review action onto the panel
// error contract: the agent's status when it answered, 502 when the
// transport failed, and the honest detail either way.
func textAgentActionError(err error) (int, string) {
	status := http.StatusBadGateway
	if agentErr, ok := errors.AsType[*textagent.Error](err); ok {
		status = agentErr.Status
	}
	return status, agentDetail(err)
}

// approveDraft sends a pending draft — verbatim, or with the human
// edit the review form carried (the agent records the correction as a
// learned preference). Success re-renders the drafts panel; failure
// rides the panel-error contract so the banner lands where the owner
// looks.
func (h *handlers) approveDraft(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.RequireSession(w, r)
	if !ok {
		return
	}
	if !h.textAgentOn() {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), textAgentPanelTimeout)
	defer cancel()
	if _, err := h.deps.TextAgent.Approve(ctx, r.PathValue("id"), r.FormValue("text")); err != nil {
		status, detail := textAgentActionError(err)
		h.renderPanelError(w, r, sess, views.TabMessages, status, h.T(r, "drafts.approve")+": "+detail)
		return
	}
	notifyToast(w, "ok", h.T(r, "toast.messageSent"))
	h.draftPanel(w, r)
}

// discardDraft drops a pending draft without sending.
func (h *handlers) discardDraft(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.RequireSession(w, r)
	if !ok {
		return
	}
	if !h.textAgentOn() {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), textAgentPanelTimeout)
	defer cancel()
	if err := h.deps.TextAgent.Discard(ctx, r.PathValue("id")); err != nil {
		status, detail := textAgentActionError(err)
		h.renderPanelError(w, r, sess, views.TabMessages, status, h.T(r, "drafts.discard")+": "+detail)
		return
	}
	h.draftPanel(w, r)
}

// aiDraft is the compose button: the signed-in owner asks the agent for
// an on-demand draft for the conversation at hand (the `to` the
// composer carries plus whatever hint is already typed). The answer
// round-trips into the composer as a prefill — nothing is sent until
// the owner presses Send, exactly like a hand-typed reply.
func (h *handlers) aiDraft(w http.ResponseWriter, r *http.Request) {
	sess, to, ok := h.requireMultipartTo(w, r, views.TabMessages, "err.invalidTo")
	if !ok {
		return
	}
	if !h.textAgentOn() {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), textAgentDraftTimeout)
	defer cancel()
	draft, err := h.deps.TextAgent.DraftFor(ctx, to.String(), r.FormValue("body"))
	if err != nil {
		status, detail := textAgentActionError(err)
		h.renderPanelError(w, r, sess, views.TabMessages, status, h.T(r, "compose.ai")+": "+detail)
		return
	}
	// A thread-scoped ask (the reply composer's button carries
	// ?thread=) re-opens the conversation with the draft in the reply
	// box; the new-conversation ask lands it in the compose form.
	if threadParam := r.URL.Query().Get("thread"); threadParam != "" {
		threadID, err := domain.ParseThreadID(threadParam)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		component, err := h.threadPanelPrefilled(r, sess, threadID, 0, draft.Text)
		if err != nil {
			h.internalError(w, r, "load conversation", err)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, "render error", http.StatusInternalServerError)
		}
		return
	}
	component, err := h.messagesPanelPrefilled(r, sess, draft.Text)
	if err != nil {
		h.internalError(w, r, "load messages", err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// setDraftMode is the drafts panel's auto-reply switch: "reset" drops the
// agent-side runtime override and returns to the deployment's configured
// mode; any other value pins the mode the agent persists until reset. The
// safety gates are unchanged — the mode only decides whether a confident
// draft may send itself.
func (h *handlers) setDraftMode(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.RequireSession(w, r)
	if !ok {
		return
	}
	if !h.textAgentOn() {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), textAgentPanelTimeout)
	defer cancel()
	var err error
	if r.FormValue("mode") == "reset" {
		_, err = h.deps.TextAgent.ResetMode(ctx)
	} else {
		_, err = h.deps.TextAgent.SetMode(ctx, r.FormValue("mode"))
	}
	if err != nil {
		status, detail := textAgentActionError(err)
		h.renderPanelError(w, r, sess, views.TabMessages, status, h.T(r, "drafts.mode.label")+": "+detail)
		return
	}
	notifyToast(w, "ok", h.T(r, "drafts.mode.changed"))
	h.draftPanel(w, r)
}
