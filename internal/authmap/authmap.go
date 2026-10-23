// Package authmap is the ONE email→extension mapping behind every
// passwordless login mode (passkey today, OIDC/SSO next to it).
//
// It stays config-free — the composition root translates both config
// shapes into these types — so both login modes resolve identities and
// source SIP directory passwords through exactly one implementation: a
// second copy is how one mode silently starts allowing an account the
// other refuses.
package authmap

import (
	"os"
	"strings"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// User is one email's resolved mapping.
type User struct {
	Email       string
	DisplayName string
	Extensions  []domain.Extension
}

// SessionExtension is the extension a passwordless login binds its session
// to (the first mapped one — all stores are extension-scoped).
func (u User) SessionExtension() domain.Extension {
	if len(u.Extensions) == 0 {
		return domain.Extension{}
	}
	return u.Extensions[0]
}

// Map resolves authenticated emails to their mapped identities and the
// runtime file carrying each extension's SIP directory password.
type Map struct {
	users         map[string]User
	passwordFiles map[string]string
}

// New builds a Map. Both the user map and the password-file map are
// required: a mapped extension without a password file would fail every
// login at the last step, so the misconfiguration surfaces here instead.
// Empty maps are legal (the mode is off); a mapped extension without a
// file is not.
func New(users map[string]User, passwordFiles map[string]string) (*Map, error) {
	normalized := make(map[string]User, len(users))
	for email, user := range users {
		key := strings.ToLower(strings.TrimSpace(email))
		normalized[key] = user
		for _, ext := range user.Extensions {
			if file := passwordFiles[ext.String()]; file == "" {
				return nil, errorfamily.Newf(errorfamily.Rejection, "authmap.password_file.missing",
					"no password file for extension %q (mapped by %q): passwordless login must source the SIP directory password server-side", ext.String(), email)
			}
		}
	}
	files := make(map[string]string, len(passwordFiles))
	for ext, file := range passwordFiles {
		files[ext] = file
	}
	return &Map{users: normalized, passwordFiles: files}, nil
}

// Resolve maps an email to its configured extensions; ok=false when the
// email has no mapping (login fails closed — an IdP identity alone must
// never mint a session for an unmapped account).
func (m *Map) Resolve(email string) (User, bool) {
	if m == nil {
		return User{}, false
	}
	user, ok := m.users[strings.ToLower(strings.TrimSpace(email))]
	return user, ok
}

// MappedByExtension reverse-resolves: which mapped user (if any) a session
// extension belongs to. Powers the display identity on the extension login
// too — a mapped user gets their name and numbers in the whoami line
// regardless of how they signed in.
func (m *Map) MappedByExtension(extension domain.Extension) (User, bool) {
	if m == nil {
		return User{}, false
	}
	for _, user := range m.users {
		for _, mapped := range user.Extensions {
			if mapped == extension {
				return user, true
			}
		}
	}
	return User{}, false
}

// SIPPassword reads the extension's directory password from its
// operator-managed file — at login time, never cached, so rotation via the
// secrets pipeline takes effect on the next login. Empty or
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
