package server

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/a-h/templ"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
	"github.com/larsartmann/webphone/internal/web/views"
)

// The Admin tab: read-only visibility across EVERY line of the phone
// system — accounts, numbers, conversations, faxes and calls. Admin is a
// property of the identity (the shared auth.users mapping's admin flag),
// resolved per request from the LIVE mapping: nothing admin-shaped is
// minted into the session, so a config demotion bites on the admin's
// next request, and a deployment with no admin configured is
// byte-identical to one without the feature (every admin route answers
// the styled 404 for everyone).

// Admin section sizes: bounded windows, deliberately generous (an
// inventory is the point) but capped so a flooded line cannot produce an
// unbounded page.
const (
	adminThreadsCap = 50
	adminFaxesCap   = 50
	adminCallsCap   = 10
	adminCDRPerLine = 25
	adminCDRCap     = 60
)

// isAdminSession answers whether this session may see the admin
// surface: the seam wired AND the session's login email carries the
// admin flag. Fail-closed on every leg — a nil seam, an extension login
// (no identity), an unmapped email and a mapped non-admin all answer
// false.
func (h *handlers) isAdminSession(sess session.Session) bool {
	return h.deps.Admin != nil && h.deps.Admin.IsAdmin(sess.LoginEmail)
}

// requireAdmin gates an admin handler: the session check first (401),
// the admin check second. A non-admin gets the SAME styled 404 an
// anonymous visitor gets — the surface never confirms it exists
// (anti-enumeration, the dial-gate 404 doctrine).
func (h *handlers) requireAdmin(w http.ResponseWriter, r *http.Request) (session.Session, bool) {
	sess, ok := h.RequireSession(w, r)
	if !ok {
		return session.Session{}, false
	}
	if !h.isAdminSession(sess) {
		h.notFound(w, r)
		return session.Session{}, false
	}
	return sess, true
}

// adminPage renders the full shell for the Admin tab (deep link).
func (h *handlers) adminPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	h.renderShell(w, r, views.TabAdmin)
}

// adminPartial renders just the Admin tab region for an HTMX swap.
func (h *handlers) adminPartial(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	h.partial(w, r, views.TabAdmin)
}

// adminThreadPage renders the read-only cross-line conversation (deep
// link; the nav rows' href).
func (h *handlers) adminThreadPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	component, ok := h.adminThreadComponent(w, r)
	if !ok {
		return
	}
	h.renderShell(w, r, views.TabAdmin, component)
}

// adminThreadPartial renders the conversation region for an HTMX swap
// (the nav rows' hx-get).
func (h *handlers) adminThreadPartial(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	component, ok := h.adminThreadComponent(w, r)
	if !ok {
		return
	}
	renderHTML(w, r, component)
}

// adminThreadComponent builds the read-only conversation view for any
// owner's thread: the by-id lookup crosses the line once (ThreadAny),
// then the read re-enters the OWNER-SCOPED window with the thread's own
// owner. ok=false means the response is already answered (not-found or
// 500).
func (h *handlers) adminThreadComponent(w http.ResponseWriter, r *http.Request) (templ.Component, bool) {
	id, err := domain.ParseThreadID(r.PathValue("id"))
	if err != nil {
		h.notFound(w, r)
		return nil, false
	}
	thread, err := h.deps.Messaging.ThreadAny(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		h.notFound(w, r)
		return nil, false
	}
	if err != nil {
		h.internalError(w, r, "load conversation", err)
		return nil, false
	}
	thread, messages, _, err := h.deps.Messaging.ThreadWindow(r.Context(), thread.Owner, id, 0)
	if err != nil {
		h.internalError(w, r, "load conversation", err)
		return nil, false
	}
	return views.AdminThread(views.AdminThreadProps{
		Thread: thread, Messages: messages, Lang: h.lang(r),
	}), true
}

// adminPanel builds the Admin tab: the identity inventory (accounts,
// numbers, live sessions) plus the cross-owner data windows. Every data
// section degrades independently — one unreachable seam (the phone API,
// a store failure) shrinks its section with an honest notice instead of
// dying the panel.
func (h *handlers) adminPanel(r *http.Request, sess session.Session) (templ.Component, error) {
	// Defense in depth: the routes gate first; a hand-routed TabAdmin for
	// a non-admin session (a future wiring mistake) must not leak the
	// panel either.
	if !h.isAdminSession(sess) {
		return views.NotFoundPanel(r.URL.Path, h.lang(r)), nil
	}
	props := views.AdminPanelProps{Lang: h.lang(r)}
	h.fillAdminIdentity(r, &props)
	h.fillAdminThreads(r, &props)
	h.fillAdminFaxes(r, &props)
	h.fillAdminCalls(r, &props)
	return views.AdminPanel(props), nil
}

// fillAdminIdentity fills the users and numbers inventories from the
// admin seam (nil-safe: a hand-composed test Deps renders both sections
// empty with their empty states).
func (h *handlers) fillAdminIdentity(r *http.Request, props *views.AdminPanelProps) {
	if h.deps.Admin == nil {
		return
	}
	for _, user := range h.deps.Admin.Users() {
		row := views.AdminUserRow{
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Admin:       user.Admin,
		}
		for _, ext := range user.Extensions {
			row.Extensions = append(row.Extensions, ext.String())
			// Live sessions count across the account's lines: a second
			// line signed in is a second live session of the same human.
			row.LiveSessions += len(h.deps.Sessions.ActiveTokens(ext))
		}
		props.Users = append(props.Users, row)
	}
	for _, number := range h.deps.Admin.Numbers() {
		props.Numbers = append(props.Numbers, views.AdminNumberRow{
			Extension: number.Extension.String(),
			DID:       number.DID,
			Users:     number.Users,
		})
	}
}

// fillAdminThreads fills the cross-owner conversation inventory; a read
// failure degrades the section, the panel stays.
func (h *handlers) fillAdminThreads(r *http.Request, props *views.AdminPanelProps) {
	threads, err := h.deps.Messaging.AllThreads(r.Context(), adminThreadsCap)
	if err != nil {
		h.adminSectionFailed(r, "admin threads", err)
		props.ThreadsError = h.T(r, "admin.sectionUnavailable")
		return
	}
	props.Threads = threads
}

// fillAdminFaxes fills the cross-owner fax inventory; same degrade.
func (h *handlers) fillAdminFaxes(r *http.Request, props *views.AdminPanelProps) {
	jobs, err := h.deps.Faxes.ListAll(r.Context(), adminFaxesCap)
	if err != nil {
		h.adminSectionFailed(r, "admin faxes", err)
		props.FaxesError = h.T(r, "admin.sectionUnavailable")
		return
	}
	props.Faxes = jobs
}

// fillAdminCalls fills both halves of the calls section: the cross-owner
// transcribed calls (the transcripts store) and the merged per-line CDR
// page (each line's own directory credentials, sourced fresh from its
// password file — never cached, the authmap doctrine).
func (h *handlers) fillAdminCalls(r *http.Request, props *views.AdminPanelProps) {
	if h.deps.Transcripts != nil {
		recent, err := h.deps.Transcripts.RecentAll(r.Context(), adminCallsCap)
		if err != nil {
			h.adminSectionFailed(r, "admin transcripts", err)
			props.TranscriptsError = h.T(r, "admin.sectionUnavailable")
		} else {
			for _, tx := range recent {
				props.Transcripts = append(props.Transcripts, views.AdminCallRow{
					Owner:     tx.Owner.String(),
					Remote:    tx.Remote,
					Direction: string(tx.Direction),
					Lines:     tx.Lines,
				})
			}
		}
	}
	if !h.deps.PhoneAPI.Enabled() || h.deps.Admin == nil {
		props.CDROff = true
		return
	}
	var rows []views.AdminCDRRow
	for _, ext := range h.deps.Admin.Extensions() {
		password, err := h.deps.Admin.SIPPassword(ext)
		if err != nil {
			h.adminSectionFailed(r, "admin cdr credentials", err)
			props.CDRFailed = append(props.CDRFailed, ext.String())
			continue
		}
		page, err := h.deps.PhoneAPI.History(r.Context(), pbx.Credentials{
			Extension: ext.String(), Password: password,
		}, adminCDRPerLine)
		if err != nil {
			h.adminSectionFailed(r, "admin cdr page", err)
			props.CDRFailed = append(props.CDRFailed, ext.String())
			continue
		}
		for _, cdr := range page.Entries {
			rows = append(rows, views.AdminCDRRow{CDR: cdr, Extension: ext.String()})
		}
	}
	// The FreeSWITCH stamp ("YYYY-MM-DD HH:MM:SS") sorts lexically =
	// chronologically; newest first, then the display cap.
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].CDR.Start > rows[j].CDR.Start })
	if len(rows) > adminCDRCap {
		rows = rows[:adminCDRCap]
	}
	props.CDRs = rows
}

// adminSectionFailed logs one degraded section with the operator
// context (the internalError posture: full detail in the log, generic
// copy in the panel) — the section's notice stays generic, the detail
// stays out of the browser.
func (h *handlers) adminSectionFailed(r *http.Request, op string, err error) {
	errorfamily.LogErrorContext(r.Context(), fmt.Errorf("%s: %w", op, err), nil)
}
