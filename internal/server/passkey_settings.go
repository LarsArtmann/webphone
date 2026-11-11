package server

import (
	"net/http"

	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/web/views"
)

// The Settings-tab passkey management surface (A2 slice): the signed-in
// session lists and revokes the passkeys of ITS OWN mapped account.
// Owner-match is structural — the email comes from the session's login
// identity (MappedByLogin prefers it; the shared-extension fallback is
// deterministic but cannot know which human a line-password session is),
// never from the request. A classic extension login authenticated the
// LINE, not a human: it has no own passkeys, so the section stays
// hidden and revocation answers 404 — a line-password session must not
// manage the sorted-first human's credentials on a shared extension.

// passkeySectionProps shapes the Settings passkeys section. Available is
// false when the identity layer is off, the session carries no login
// identity (extension login), or the account has no mapped extension
// (the section renders nothing); Empty marks a mapped account with zero
// enrolled passkeys.
func (h *handlers) passkeySectionProps(sess session.Session) views.PasskeysSectionProps {
	if h.deps.UserAuth == nil || sess.LoginEmail == "" {
		return views.PasskeysSectionProps{}
	}
	mapped, ok := h.deps.UserAuth.MappedByLogin(sess.Extension, sess.LoginEmail)
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
// An extension login (no identity) has no own passkeys to revoke — the
// uniform 404 is the anti-enumeration answer, not an error shape.
func (h *handlers) revokePasskey(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.RequireSession(w, r)
	if !ok {
		return
	}
	if sess.LoginEmail == "" {
		http.Error(w, "no passkey account", http.StatusNotFound)
		return
	}
	mapped, ok := h.deps.UserAuth.MappedByLogin(sess.Extension, sess.LoginEmail)
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
