// Package domain holds the pure webphone domain types: no I/O, no
// sibling-internal imports (enforced by internal/arch). Identifiers
// are branded so a ThreadID can never be passed where a FaxID is
// expected; validation failures classify via go-error-family (itself
// stdlib-only, so the transitive dependency surface stays unchanged).
package domain

import (
	"encoding/json"
	"strings"

	id "github.com/larsartmann/go-branded-id"
	"github.com/larsartmann/go-error-family"
	"github.com/sixafter/nanoid"
)

// Extension is a PBX extension (the SIP user part), e.g. "1001". A defined
// struct (not an alias) so it cannot be confused with a Phone or a raw
// string; only ParseExtension can mint one.
type Extension struct {
	value string
}

// ParseExtension validates and brands an extension: 1..32 dialable
// characters (digits, letters, +, *, #).
func ParseExtension(raw string) (Extension, error) {
	clean := sanitizeDialable(raw)
	if clean == "" {
		return Extension{}, errorfamily.Newf(errorfamily.Rejection, "domain.extension", "extension %q has no dialable characters", raw)
	}
	if len(clean) > 32 {
		return Extension{}, errorfamily.Newf(errorfamily.Rejection, "domain.extension", "extension %q longer than 32 characters", raw)
	}
	return Extension{value: clean}, nil
}

// MustParseExtension is ParseExtension for literals in tests and seeds.
func MustParseExtension(raw string) Extension {
	return must(ParseExtension(raw))
}

// String returns the raw extension.
func (e Extension) String() string { return e.value }

// IsZero reports whether the extension is the zero value.
func (e Extension) IsZero() bool { return e.value == "" }

// MarshalJSON renders the raw extension (the id.ID doctrine: branded
// values JSON-marshal their raw value, never a wrapper object). Under
// encoding/json v2 semantics an unexported-only struct is a marshal
// ERROR rather than `{}`, so identity shapes carrying Extensions need
// this to stay serializable in both modes.
func (e Extension) MarshalJSON() ([]byte, error) { return json.Marshal(e.String()) }

// Phone is a sanitized remote number (digits with an optional leading +,
// plus * and # which some dialplans use). The island's dial field uses the
// same rule, so history rows and threads key identically.
type Phone struct {
	value string
}

// ParsePhone sanitizes and brands a remote number.
func ParsePhone(raw string) (Phone, error) {
	clean := sanitizeDialable(raw)
	if clean == "" {
		return Phone{}, errorfamily.Newf(errorfamily.Rejection, "domain.phone", "number %q has no dialable characters", raw)
	}
	if len(clean) > 32 {
		return Phone{}, errorfamily.Newf(errorfamily.Rejection, "domain.phone", "number %q longer than 32 characters", raw)
	}
	return Phone{value: clean}, nil
}

// MustParsePhone is ParsePhone for literals in tests and seeds.
func MustParsePhone(raw string) Phone {
	return must(ParsePhone(raw))
}

// String returns the sanitized number.
func (p Phone) String() string { return p.value }

// IsZero reports whether the phone is the zero value.
func (p Phone) IsZero() bool { return p.value == "" }

// keepRunes is the ONE home of the "keep only the runes of s for which
// keep holds, order preserved" loop: the dialable sanitize, the contact
// digit key and the avatar digit tail share the shape, and second
// hand-kept copies drift the day one of them learns a new character class.
func keepRunes(s string, keep func(rune) bool) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if keep(r) {
			out = append(out, r)
		}
	}
	return string(out)
}

// isASCIIDigit reports whether r is one of '0'-'9'. Deliberately not
// unicode.IsDigit: non-ASCII digit glyphs must never enter digit keys.
func isASCIIDigit(r rune) bool { return r >= '0' && r <= '9' }

// isDialable reports whether r is in the dialable alphabet (digits,
// letters, +, *, #), exactly the island's
// `raw.replace(/[^\d+*#a-zA-Z]/g, "")`.
func isDialable(r rune) bool {
	switch {
	case isASCIIDigit(r), r == '+', r == '*', r == '#',
		r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		return true
	}
	return false
}

// DigitsOf returns s reduced to its ASCII digits ('0'-'9'), order
// preserved: the contact upsert key and the avatar digit tail are two
// consumers of the same digits-only reduction, so one home carries the loop.
func DigitsOf(s string) string { return keepRunes(s, isASCIIDigit) }

// sanitizeDialable strips invisible Unicode direction marks and formatting
// that pasted numbers carry, keeping only dialable characters. The
// dialable alphabet (digits, letters, +, *, #) is exactly the island's
// `raw.replace(/[^\d+*#a-zA-Z]/g, "")` (DECIDED 2026-09-20: the island
// keeps letters, like this side — some PBXs use alphanumeric SIP user
// parts). Both sides are pinned: this test file for Go, the served-asset
// table in internal/server for the island literal.
func sanitizeDialable(raw string) string { return keepRunes(raw, isDialable) }

// ThreadBrand brands thread identifiers.
type ThreadBrand struct{}

// Name implements the id.Brand interface.
func (ThreadBrand) Name() string { return "Thread" }

// ThreadID identifies a message thread.
type ThreadID = id.ID[ThreadBrand, nanoid.ID]

// GenerateThreadID mints a new thread identifier.
func GenerateThreadID() ThreadID { return id.NewID[ThreadBrand](nanoid.Must()) }

// ParseThreadID parses a stored thread id (branded "Thread:xxx" or raw
// nanoid form). Handlers use the Parse form on client-supplied ids so a
// forged or typoed id answers 404 instead of panicking.
func ParseThreadID(s string) (ThreadID, error) { return parseID[ThreadBrand](s, "thread") }

// MustThreadID is ParseThreadID for ids read back from the database, where
// a malformed id means corruption and panicking is the honest response.
func MustThreadID(s string) ThreadID { return must(ParseThreadID(s)) }

// MessageBrand brands message identifiers.
type MessageBrand struct{}

// Name implements the id.Brand interface.
func (MessageBrand) Name() string { return "Message" }

// MessageID identifies a single message.
type MessageID = id.ID[MessageBrand, nanoid.ID]

// GenerateMessageID mints a new message identifier.
func GenerateMessageID() MessageID { return id.NewID[MessageBrand](nanoid.Must()) }

// ParseMessageID parses a stored message id; see ParseThreadID.
func ParseMessageID(s string) (MessageID, error) { return parseID[MessageBrand](s, "message") }

// MustMessageID is ParseMessageID for database rows; see MustThreadID.
func MustMessageID(s string) MessageID { return must(ParseMessageID(s)) }

// AttachmentBrand brands attachment identifiers.
type AttachmentBrand struct{}

// Name implements the id.Brand interface.
func (AttachmentBrand) Name() string { return "Attachment" }

// AttachmentID identifies a message attachment.
type AttachmentID = id.ID[AttachmentBrand, nanoid.ID]

// GenerateAttachmentID mints a new attachment identifier.
func GenerateAttachmentID() AttachmentID { return id.NewID[AttachmentBrand](nanoid.Must()) }

// ParseAttachmentID parses a stored attachment id; see ParseThreadID.
func ParseAttachmentID(s string) (AttachmentID, error) {
	return parseID[AttachmentBrand](s, "attachment")
}

// MustAttachmentID is ParseAttachmentID for database rows; see MustThreadID.
func MustAttachmentID(s string) AttachmentID { return must(ParseAttachmentID(s)) }

// FaxBrand brands fax job identifiers.
type FaxBrand struct{}

// Name implements the id.Brand interface.
func (FaxBrand) Name() string { return "Fax" }

// FaxID identifies a fax job.
type FaxID = id.ID[FaxBrand, nanoid.ID]

// GenerateFaxID mints a new fax identifier.
func GenerateFaxID() FaxID { return id.NewID[FaxBrand](nanoid.Must()) }

// ParseFaxID parses a stored fax id; see ParseThreadID.
func ParseFaxID(s string) (FaxID, error) { return parseID[FaxBrand](s, "fax") }

// MustFaxID is ParseFaxID for database rows; see MustThreadID.
func MustFaxID(s string) FaxID { return must(ParseFaxID(s)) }

// ContactBrand brands contact identifiers.
type ContactBrand struct{}

// Name implements the id.Brand interface.
func (ContactBrand) Name() string { return "Contact" }

// ContactID identifies a personal contact.
type ContactID = id.ID[ContactBrand, nanoid.ID]

// GenerateContactID mints a new contact identifier.
func GenerateContactID() ContactID { return id.NewID[ContactBrand](nanoid.Must()) }

// ParseContactID parses a stored contact id; see ParseThreadID.
func ParseContactID(s string) (ContactID, error) { return parseID[ContactBrand](s, "contact") }

// MustContactID is ParseContactID for database rows; see MustThreadID.
func MustContactID(s string) ContactID { return must(ParseContactID(s)) }

// SnippetBrand brands snippet identifiers.
type SnippetBrand struct{}

// Name implements the id.Brand interface.
func (SnippetBrand) Name() string { return "Snippet" }

// SnippetID identifies a reply snippet.
type SnippetID = id.ID[SnippetBrand, nanoid.ID]

// GenerateSnippetID mints a new snippet identifier.
func GenerateSnippetID() SnippetID { return id.NewID[SnippetBrand](nanoid.Must()) }

// ParseSnippetID parses a stored snippet id; see ParseThreadID.
func ParseSnippetID(s string) (SnippetID, error) {
	return parseID[SnippetBrand](s, "snippet")
}

// MustSnippetID is ParseSnippetID for database rows; see MustThreadID.
func MustSnippetID(s string) SnippetID { return must(ParseSnippetID(s)) }

// parseID re-brands a stored nanoid-backed identifier. Both the branded
// ("Thread:xxx") and raw ("xxx") forms parse; anything else is an error.
//
// Parked adoption decision (stack audit finding #8, 2026-09-24): the
// go-branded-id library has no prefix-aware parser/Valuer/Scanner — it
// round-trips its own "Brand:value" form only, while rows at rest carry
// raw nano ids. This hand-rolled parseID stays the one home for the
// both-forms tolerance; adopt the library's Valuer/Scanner seam only on
// the next storage-format touch (a migration that rewrites id columns
// can unify the wire format at the same time).
func parseID[B any](s string, kind string) (id.ID[B, nanoid.ID], error) {
	raw := s
	if _, rest, found := strings.Cut(s, ":"); found {
		raw = rest
	}
	if len(raw) != 21 {
		return id.ID[B, nanoid.ID]{}, errorfamily.Newf(errorfamily.Corruption, "domain.id", "corrupt %s id %q: want 21 nanoid chars", kind, s)
	}
	return id.NewID[B](nanoid.ID(raw)), nil
}

// must unwraps a Parse result, panicking on error. The Must forms are fed
// literals (tests, seeds) or database rows: a failure means a broken
// literal or a corrupt database, never bad user input.
func must[T any](parsed T, err error) T {
	if err != nil {
		panic(err)
	}
	return parsed
}
