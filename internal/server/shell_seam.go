package server

import (
	"net/http"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/web/views"
)

// The auth.Shell seam (the carved auth package rides on this hub):
// exported wrappers so internal/server/auth can borrow the session
// gate and the render plumbing without reaching into unexported
// methods — the wording and helpers keep their one home here.

// IdentityFor is the auth seam's identity read.
func (h *handlers) IdentityFor(ext domain.Extension) string { return h.identityFor(ext) }

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
