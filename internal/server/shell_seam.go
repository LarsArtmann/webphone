package server

import (
	"net/http"

	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/web/views"
)

// The auth.Shell seam (the carved auth package rides on this hub):
// exported wrappers so internal/server/auth can borrow the session
// gate and the render plumbing without reaching into unexported
// methods — the wording and helpers keep their one home here.

// IdentityFor is the auth seam's identity read.
func (h *handlers) IdentityFor(ext domain.Extension) string { return h.identityFor(ext) }

// MappedIdentityWith is the ONE identity resolution behind every
// session-facing surface (whoami, the line switcher, Settings' account
// section): the session's own login email when it owns the session's
// line, else the deterministic line fallback — passkey's authmap first,
// SSO's second, so an SSO-only deployment resolves the same ownership
// model a passkey one does. ok=false is the classic extension login.
func (h *handlers) MappedIdentityWith(extension domain.Extension, loginEmail string) (authmap.User, bool) {
	if h.deps.UserAuth != nil {
		if mapped, ok := h.deps.UserAuth.MappedByLogin(extension, loginEmail); ok {
			return mapped, true
		}
	}
	if h.deps.SSO != nil {
		if mapped, ok := h.deps.SSO.MappedByLogin(extension, loginEmail); ok {
			return mapped, true
		}
	}
	return authmap.User{}, false
}

// SafeDetail is the auth seam's honest-detail log-and-return.
func (h *handlers) SafeDetail(r *http.Request, op string, err error) string {
	return h.safeDetail(r, op, err)
}

// Partial is the auth seam's tab-partial re-render.
func (h *handlers) Partial(w http.ResponseWriter, r *http.Request, tab views.Tab) {
	h.partial(w, r, tab)
}

// Lang is the auth seam's language resolution.
func (h *handlers) Lang(r *http.Request) views.Lang { return h.lang(r) }

// NotFound is the auth seam's styled 404.
func (h *handlers) NotFound(w http.ResponseWriter, r *http.Request) { h.notFound(w, r) }

// NotifyToast is the auth seam's toast push.
func (h *handlers) NotifyToast(w http.ResponseWriter, kind, msg string) {
	notifyToast(w, kind, msg)
}

// NotFoundPage is the platform seam's styled 404 (asset-tree misses).
func (h *handlers) NotFoundPage(w http.ResponseWriter, r *http.Request) {
	h.notFoundPage(w, r)
}

// InternalError is the platform seam's terse 500 (op logged).
func (h *handlers) InternalError(w http.ResponseWriter, r *http.Request, op string, err error) {
	h.internalError(w, r, op, err)
}

// NotifyVoicemail is the platform seam's post-read voicemail nudge (the
// proxy forwards a 2xx voicemail read; the event vocabulary's one home
// stays here).
func (h *handlers) NotifyVoicemail(extension domain.Extension) {
	h.deps.Hubs.Publish(extension, sseEventVoicemail, "")
}
