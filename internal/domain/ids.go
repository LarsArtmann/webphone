// Package domain holds the pure webphone domain types: no external
// dependencies, no I/O. Identifiers are branded so a ThreadID can never be
// passed where a FaxID is expected.
package domain

import (
	"fmt"
	"strings"

	id "github.com/larsartmann/go-branded-id"
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
		return Extension{}, fmt.Errorf("extension %q has no dialable characters", raw)
	}
	if len(clean) > 32 {
		return Extension{}, fmt.Errorf("extension %q longer than 32 characters", raw)
	}
	return Extension{value: clean}, nil
}

// MustParseExtension is ParseExtension for literals in tests and seeds.
func MustParseExtension(raw string) Extension {
	ext, err := ParseExtension(raw)
	if err != nil {
		panic(err)
	}
	return ext
}

// String returns the raw extension.
func (e Extension) String() string { return e.value }

// IsZero reports whether the extension is the zero value.
func (e Extension) IsZero() bool { return e.value == "" }

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
		return Phone{}, fmt.Errorf("number %q has no dialable characters", raw)
	}
	if len(clean) > 32 {
		return Phone{}, fmt.Errorf("number %q longer than 32 characters", raw)
	}
	return Phone{value: clean}, nil
}

// MustParsePhone is ParsePhone for literals in tests and seeds.
func MustParsePhone(raw string) Phone {
	phone, err := ParsePhone(raw)
	if err != nil {
		panic(err)
	}
	return phone
}

// String returns the sanitized number.
func (p Phone) String() string { return p.value }

// IsZero reports whether the phone is the zero value.
func (p Phone) IsZero() bool { return p.value == "" }

// sanitizeDialable strips invisible Unicode direction marks and formatting
// that pasted numbers carry, keeping only dialable characters. Mirrors the
// island's `raw.replace(/[^\d+*#]/g, "")` exactly.
func sanitizeDialable(raw string) string {
	clean := make([]rune, 0, len(raw))
	for _, r := range raw {
		switch {
		case r >= '0' && r <= '9', r == '+', r == '*', r == '#',
			r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			clean = append(clean, r)
		}
	}
	return string(clean)
}

// ThreadBrand brands thread identifiers.
type ThreadBrand struct{}

// Name implements the id.Brand interface.
func (ThreadBrand) Name() string { return "Thread" }

// ThreadID identifies a message thread.
type ThreadID = id.ID[ThreadBrand, nanoid.ID]

// GenerateThreadID mints a new thread identifier.
func GenerateThreadID() ThreadID { return id.NewID[ThreadBrand](nanoid.Must()) }

// MustThreadID parses a stored thread id; it panics on malformed input,
// which can only come from a corrupted database.
func MustThreadID(s string) ThreadID { return mustID[ThreadBrand](s, "thread") }

// MessageBrand brands message identifiers.
type MessageBrand struct{}

// Name implements the id.Brand interface.
func (MessageBrand) Name() string { return "Message" }

// MessageID identifies a single message.
type MessageID = id.ID[MessageBrand, nanoid.ID]

// GenerateMessageID mints a new message identifier.
func GenerateMessageID() MessageID { return id.NewID[MessageBrand](nanoid.Must()) }

// MustMessageID parses a stored message id.
func MustMessageID(s string) MessageID { return mustID[MessageBrand](s, "message") }

// AttachmentBrand brands attachment identifiers.
type AttachmentBrand struct{}

// Name implements the id.Brand interface.
func (AttachmentBrand) Name() string { return "Attachment" }

// AttachmentID identifies a message attachment.
type AttachmentID = id.ID[AttachmentBrand, nanoid.ID]

// GenerateAttachmentID mints a new attachment identifier.
func GenerateAttachmentID() AttachmentID { return id.NewID[AttachmentBrand](nanoid.Must()) }

// MustAttachmentID parses a stored attachment id.
func MustAttachmentID(s string) AttachmentID { return mustID[AttachmentBrand](s, "attachment") }

// FaxBrand brands fax job identifiers.
type FaxBrand struct{}

// Name implements the id.Brand interface.
func (FaxBrand) Name() string { return "Fax" }

// FaxID identifies a fax job.
type FaxID = id.ID[FaxBrand, nanoid.ID]

// GenerateFaxID mints a new fax identifier.
func GenerateFaxID() FaxID { return id.NewID[FaxBrand](nanoid.Must()) }

// MustFaxID parses a stored fax id.
func MustFaxID(s string) FaxID { return mustID[FaxBrand](s, "fax") }

// ContactBrand brands contact identifiers.
type ContactBrand struct{}

// Name implements the id.Brand interface.
func (ContactBrand) Name() string { return "Contact" }

// ContactID identifies a personal contact.
type ContactID = id.ID[ContactBrand, nanoid.ID]

// GenerateContactID mints a new contact identifier.
func GenerateContactID() ContactID { return id.NewID[ContactBrand](nanoid.Must()) }

// MustContactID parses a stored contact id.
func MustContactID(s string) ContactID { return mustID[ContactBrand](s, "contact") }

// mustID re-brands a stored nanoid-backed identifier, panicking on
// corruption — malformed ids can only come from a broken database. Both
// the branded ("Thread:xxx") and raw ("xxx") forms parse.
func mustID[B any](s string, kind string) id.ID[B, nanoid.ID] {
	raw := s
	if _, rest, found := strings.Cut(s, ":"); found {
		raw = rest
	}
	if len(raw) != 21 {
		panic(fmt.Sprintf("corrupt %s id %q: want 21 nanoid chars", kind, s))
	}
	return id.NewID[B](nanoid.ID(raw))
}
