package server

import (
	"net/http"

	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/web/views"
)

// liveContextPartial answers the island's live-call context slot (issue
// #10): who is on the line and what is outstanding with them in the CRM.
//
// The contract is deliberately the quietest one in the app. The container
// exists only when the CRM seam is configured; every degraded answer is an
// EMPTY 200 body — the seam off, an unparsable or non-phone remote, no
// contact match, no published context, a CRM that is slow or down. A live
// call is the worst possible moment for a red banner: the operator is
// talking to someone, and a CRM that cannot answer is a fact about
// decoration, not about the call. Nothing here can fail the call, delay
// the island, or render layout chrome nobody asked for.
//
// The number arrives from the island (the call's own remote), so it is
// parsed as a PHONE address and nothing else; the CRM indexes phone
// numbers, and asking it about anything else would be a guaranteed 400 per
// render.
func (h *handlers) liveContextPartial(w http.ResponseWriter, r *http.Request) {
	lang := h.lang(r)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if h.deps.CRM == nil || !h.deps.CRM.Enabled() {
		return
	}

	raw := r.URL.Query().Get("remote")
	if raw == "" {
		return
	}

	address, err := domain.ParseAddress(domain.NetworkSMS, raw)
	if err != nil {
		// A malformed or non-phone remote is not an error surface: the slot
		// simply says nothing.
		return
	}

	person, ok := h.deps.CRM.Resolve(r.Context(), address.String())
	if !ok || person.Name == "" {
		return
	}

	var context *views.CRMContext
	if published, ok := h.deps.CRM.ContextOf(r.Context(), person.Ref); ok {
		context = crmContextView(*published)
	}

	props := views.LiveContextProps{
		Remote:  address.String(),
		CRM:     crmContacts(map[string]crm.Person{address.String(): person}, h.deps.CRM),
		Context: context,
		Lang:    lang,
	}
	if props.Empty() {
		return
	}

	if err := views.LiveContextPanel(props).Render(r.Context(), w); err != nil {
		// The bytes may already be on the wire; a render failure leaves the
		// slot empty, which is the same product as "no context".
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}
