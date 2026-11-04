// Package admin owns the admin surface's read-only inventory: who may
// sign in (the shared email→extension mapping), which lines exist (the
// identities feed), and whether one login email carries the admin flag.
//
// Admin is a property of the IDENTITY, not the session: IsAdmin resolves
// the live mapping per request, so a config demotion takes effect on the
// admin's next request — nothing admin-shaped is ever minted into a
// session row. The package stays config-free (the authmap doctrine): the
// composition root translates config into the authmap + identities this
// service reads.
package admin

import (
	"maps"
	"slices"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/domain"
)

// Service answers the admin surface's identity-side reads. A nil
// *Service is the mode-off shape: every method is nil-safe and answers
// "no admin, nothing listed" — the same nil-gated seam doctrine as Deps'
// other optional services.
type Service struct {
	users      *authmap.Map
	identities map[string]string
}

// New builds the service over the resolved mapping and the config
// identities feed (extension → presented DID). Identity keys must be
// normalized extensions — config.Load guarantees it, so a violation here
// is a composition defect and fails loudly.
func New(users *authmap.Map, identities map[string]string) (*Service, error) {
	parsed := make(map[string]string, len(identities))
	for ext, did := range identities {
		// The config identities validation contract: keys must be
		// NORMALIZED extensions (sanitize would change a non-normalized
		// key, silently splitting one line into two spellings).
		if normalized, err := domain.ParseExtension(ext); err != nil || normalized.String() != ext {
			return nil, errorfamily.Newf(errorfamily.Rejection, "admin.identities",
				"identities key %q is not a normalized extension (config validation owns this invariant)", ext)
		}
		parsed[ext] = did
	}
	return &Service{users: users, identities: parsed}, nil
}

// IsAdmin answers whether the login email maps to an account carrying
// the admin flag. Fail-closed on every seam: a nil service, an empty
// email (the classic extension login has no identity), and an unmapped
// email all answer false.
func (s *Service) IsAdmin(loginEmail string) bool {
	if s == nil || loginEmail == "" {
		return false
	}
	user, ok := s.users.Resolve(loginEmail)
	return ok && user.Admin
}

// Users lists the configured accounts deterministically (email-sorted;
// the authmap enumerator owns the ordering contract).
func (s *Service) Users() []authmap.User {
	if s == nil {
		return nil
	}
	return s.users.Users()
}

// Extensions lists every extension any account maps, sorted — the
// all-lines loop's shape (live sessions, CDR pages).
func (s *Service) Extensions() []domain.Extension {
	if s == nil {
		return nil
	}
	seen := make(map[string]bool)
	for _, user := range s.users.Users() {
		for _, ext := range user.Extensions {
			seen[ext.String()] = true
		}
	}
	out := make([]domain.Extension, 0, len(seen))
	for _, raw := range slices.Sorted(maps.Keys(seen)) {
		out = append(out, domain.MustParseExtension(raw))
	}
	return out
}

// Number is one line of the numbers inventory: the extension, its
// presented DID when configured, and every account mapped to it.
type Number struct {
	Extension domain.Extension
	DID       string
	Users     []string
}

// Numbers inventories the phone system's lines: the union of the
// identities feed and the mapped extensions (a mapped line without a
// DID is a line that exists; an identity without a mapping is a number
// nobody can sign in as — both belong in an honest inventory).
// Deterministic: extension-sorted, user emails sorted.
func (s *Service) Numbers() []Number {
	if s == nil {
		return nil
	}
	usersByExtension := make(map[string][]string)
	for _, user := range s.users.Users() {
		for _, ext := range user.Extensions {
			usersByExtension[ext.String()] = append(usersByExtension[ext.String()], user.Email)
		}
	}
	seen := make(map[string]bool)
	for ext := range usersByExtension {
		seen[ext] = true
	}
	for ext := range s.identities {
		seen[ext] = true
	}
	out := make([]Number, 0, len(seen))
	for _, raw := range slices.Sorted(maps.Keys(seen)) {
		ext := domain.MustParseExtension(raw)
		users := slices.Clone(usersByExtension[raw])
		slices.Sort(users)
		out = append(out, Number{Extension: ext, DID: s.identities[raw], Users: users})
	}
	return out
}

// SIPPassword sources an extension's directory password from its
// operator-managed file (read per call, never cached — the authmap
// doctrine). The admin CDR loop uses it to fetch one page of call
// records per line with that line's own credentials.
func (s *Service) SIPPassword(extension domain.Extension) (string, error) {
	if s == nil {
		return "", errorfamily.NewRejection("admin.password_file.missing",
			"admin service is not wired")
	}
	return s.users.SIPPassword(extension)
}
