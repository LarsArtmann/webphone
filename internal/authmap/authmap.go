// Package authmap is the ONE model of phone-number ownership: a Line
// (PBX extension + presented DID) is owned by a User (login email), and
// every passwordless login mode (passkey, OIDC/SSO) resolves identities
// and sources SIP directory passwords through exactly one implementation
// of it — a second copy is how one mode silently starts allowing an
// account the other refuses.
//
// It stays config-free — the composition root joins the three operator
// feeds (auth.users, identities, extension_password_files) into these
// types, so the LINE join happens exactly once.
package authmap

import (
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// Line is one telephony identity: the PBX extension its SIP traffic
// registers as, plus the PSTN number (DID) it presents to the outside
// world. The DID is display-verbatim owner formatting (config
// identities); "" declares a line whose presented number the deployment
// did not name. A Line is a value object: it has no lifecycle apart
// from its owning User's config entry.
type Line struct {
	Extension domain.Extension
	DID       string
}

// IsZero reports the zero Line (no extension).
func (l Line) IsZero() bool { return l.Extension.IsZero() }

// Display is the line's human-facing name: the DID when the deployment
// declared one, the bare extension otherwise — the ONE fallback rule
// every surface (whoami, switcher, Settings, Admin) shares. The island
// mirrors it client-side over the same wire fields.
func (l Line) Display() string {
	if l.DID != "" {
		return l.DID
	}
	return l.Extension.String()
}

// User is one email's resolved ownership: the human identity behind a
// login and the lines they may act on. A mapped user owns AT LEAST one
// line (New rejects the rest), and the FIRST line is the session's
// default — the one a fresh login binds to before any switch.
type User struct {
	Email       string
	DisplayName string
	Lines       []Line
	// Admin grants read-only cross-line visibility (the Admin tab). It
	// is a property of the IDENTITY, resolved per request from the live
	// mapping — never minted into the session, so a config demotion
	// takes effect on the admin's next request, not their next login.
	Admin bool
}

// SessionLine is the line a passwordless login binds its session to
// (the first mapped one — all stores are extension-scoped). The zero
// Line when the user owns none, which New makes unrepresentable.
func (u User) SessionLine() Line {
	if len(u.Lines) == 0 {
		return Line{}
	}
	return u.Lines[0]
}

// Has reports whether the user owns the extension (any of their lines).
func (u User) Has(extension domain.Extension) bool {
	return slices.ContainsFunc(u.Lines, func(l Line) bool { return l.Extension == extension })
}

// Extensions lists the owned lines' extensions in mapping order — the
// derived bare form for consumers that only loop (live-session counts,
// the all-lines inventory).
func (u User) Extensions() []domain.Extension {
	out := make([]domain.Extension, 0, len(u.Lines))
	for _, line := range u.Lines {
		out = append(out, line.Extension)
	}
	return out
}

// Map resolves authenticated emails to their owned lines and the
// runtime file carrying each line's SIP directory password.
type Map struct {
	users         map[string]User
	passwordFiles map[string]string
}

// New builds a Map. Both the user map and the password-file map are
// required: a mapped line without a password file would fail every
// login at the last step, so the misconfiguration surfaces here instead.
// Empty maps are legal (the mode is off); a user without lines, a user
// naming one extension twice, or a mapped line without a file is not.
func New(users map[string]User, passwordFiles map[string]string) (*Map, error) {
	normalized := make(map[string]User, len(users))
	for email, user := range users {
		key := strings.ToLower(strings.TrimSpace(email))
		user.Email = key
		normalized[key] = user
		if len(user.Lines) == 0 {
			return nil, errorfamily.Newf(errorfamily.Rejection, "authmap.user.lines",
				"user %q owns no lines: a login mapping must name at least one extension", email)
		}
		seen := make(map[string]bool, len(user.Lines))
		for _, line := range user.Lines {
			ext := line.Extension.String()
			if seen[ext] {
				return nil, errorfamily.Newf(errorfamily.Rejection, "authmap.user.duplicate_line",
					"user %q maps extension %q twice: one line per extension (switching needs them distinct)", email, ext)
			}
			seen[ext] = true
			if file := passwordFiles[ext]; file == "" {
				return nil, errorfamily.Newf(errorfamily.Rejection, "authmap.password_file.missing",
					"no password file for extension %q (mapped by %q): passwordless login must source the SIP directory password server-side", ext, email)
			}
		}
	}
	files := make(map[string]string, len(passwordFiles))
	maps.Copy(files, passwordFiles)
	return &Map{users: normalized, passwordFiles: files}, nil
}

// Resolve maps an email to its owned lines; ok=false when the email has
// no mapping (login fails closed — an IdP identity alone must never
// mint a session for an unmapped account).
func (m *Map) Resolve(email string) (User, bool) {
	if m == nil {
		return User{}, false
	}
	user, ok := m.users[strings.ToLower(strings.TrimSpace(email))]
	return user, ok
}

// MappedByExtension reverse-resolves: which mapped user (if any) a line
// belongs to. Deterministic by construction (emails iterated in sorted
// order — Go map order is random, and a shared line must not change
// whoami identity between boots). This is the FALLBACK resolver: a
// session that knows its own login email resolves through MappedByLogin
// instead.
func (m *Map) MappedByExtension(extension domain.Extension) (User, bool) {
	if m == nil {
		return User{}, false
	}
	for _, email := range slices.Sorted(maps.Keys(m.users)) {
		if m.users[email].Has(extension) {
			return m.users[email], true
		}
	}
	return User{}, false
}

// MappedByLogin resolves the identity a SESSION should display: the
// session's own login email when it owns the session's line (the
// multi-user substrate — on a shared line, each session names its own
// human), falling back to the deterministic line lookup for extension
// logins, stale emails, and unmapped accounts.
func (m *Map) MappedByLogin(extension domain.Extension, loginEmail string) (User, bool) {
	if m == nil {
		return User{}, false
	}
	if loginEmail != "" {
		if user, ok := m.Resolve(loginEmail); ok && user.Has(extension) {
			return user, true
		}
	}
	return m.MappedByExtension(extension)
}

// Users enumerates the configured mappings deterministically (sorted
// by email) — the admin surface's user listing. Go map order is random;
// a shared-line or multi-account fleet must not reshuffle between
// requests.
func (m *Map) Users() []User {
	if m == nil {
		return nil
	}
	out := make([]User, 0, len(m.users))
	for _, email := range slices.Sorted(maps.Keys(m.users)) {
		out = append(out, m.users[email])
	}
	return out
}

// SIPPassword reads the line's directory password from its
// operator-managed file — at login time, never cached, so rotation via
// the secrets pipeline takes effect on the next login. Empty or
// whitespace-only files are Rejections (the 2026-10-01 outage class: an
// empty secret must fail loudly, never mint a passwordless session).
func (m *Map) SIPPassword(extension domain.Extension) (string, error) {
	if m == nil {
		return "", errorfamily.NewRejection("authmap.password_file.missing",
			"no password-file mapping configured for passwordless login")
	}
	file, ok := m.passwordFiles[extension.String()]
	if !ok || file == "" {
		return "", errorfamily.NewRejection("authmap.password_file.missing",
			"no password file configured for extension %q (auth.extension_password_files)").WithContext("extension", extension.String())
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return "", errorfamily.WrapRejection(err, "authmap.password_file.read",
			"read password file for extension %q (check the secrets pipeline and the secrets-perms grants)").WithContext("extension", extension.String())
	}
	password := strings.TrimSpace(string(raw))
	if password == "" {
		return "", errorfamily.NewRejection("authmap.password_file.empty",
			"password file for extension %q is empty").WithContext("extension", extension.String())
	}
	return password, nil
}
