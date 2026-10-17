package server

import (
	"net/http"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/web/views"
)

// The Settings-tab passkey management surface (A2 slice): the signed-in
// extension lists and revokes the passkeys of ITS OWN mapped account.
// Owner-match is structural — the email comes from the session
// extension's mapping (MappedByExtension), never from the request — so
// an extension can only ever touch its own identity account. The
// handlers ride the session gate + the protected mux's CSRF layer.

// passkeySectionProps shapes the Settings passkeys section. Available is
// false when the identity layer is off or the extension has no mapped
// account (the section renders nothing); Empty marks a mapped account
// with zero enrolled passkeys.
func (h *handlers) passkeySectionProps(extension domain.Extension) views.PasskeysSectionProps {
	if h.deps.UserAuth == nil {
		return views.PasskeysSectionProps{}
	}
	mapped, ok := h.deps.UserAuth.MappedByExtension(extension)
	if !ok {
		return views.PasskeysSectionProps{}
	}
	creds, _ := h.deps.UserAuth.Credentials(mapped.Email)
	props := views.PasskeysSectionProps{Available: true, Empty: len(creds) == 0}
	for _, cred := range creds {
		props.Credentials = append(props.Credentials, views.PasskeyCredentialView{
			ID:        cred.ID,
			Name:      cred.Name,
			CreatedAt: cred.CreatedAt,
		})
	}
	return props
}

// revokePasskey removes one passkey of the session's own mapped account
// and re-renders the settings partial with a toast (HX-Trigger
// showMessage — the island's one toast listener); the same whole-panel
// swap the snippets section uses, so the list always shows store truth.
func (h *handlers) revokePasskey(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.RequireSession(w, r)
	if !ok {
		return
	}
	mapped, ok := h.deps.UserAuth.MappedByExtension(sess.Extension)
	if !ok {
		http.Error(w, "no passkey account", http.StatusNotFound)
		return
	}
	if err := h.deps.UserAuth.RemoveCredential(r.Context(), mapped.Email, r.PathValue("id")); err != nil {
		h.safeDetail(r, "revoke passkey", err)
		notifyToast(w, "error", views.T(h.lang(r), "passkeys.revokeFailed"))
	} else {
		authAudit("passkey_revoked", "email", mapped.Email, "credential "+r.PathValue("id"))
		notifyToast(w, "ok", views.T(h.lang(r), "passkeys.revoked"))
	}
	h.partial(w, r, views.TabSettings)
}
