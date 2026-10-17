package domain

import (
	"strings"

	"github.com/larsartmann/go-error-family"
)

// RelPath is a stored blob's location relative to the blob root: the
// database persists exactly this string and every disk touch joins it
// onto the root. A defined struct (not an alias) so an absolute path or
// an escaping ".." segment cannot be passed where a root-relative one
// is expected — only ParseRelPath can mint one, and it refuses both
// escape shapes up front. The blob store's Open/Abs/Remove take this
// type, which moves the old runtime escape check to the type boundary:
// by construction, nothing that did not cross ParseRelPath reaches the
// disk join.
type RelPath struct {
	value string
}

// ParseRelPath validates and brands a stored relative path. Rejected
// shapes: empty, absolute (leading "/"), backslash separators, and any
// ".." segment.
func ParseRelPath(raw string) (RelPath, error) {
	if raw == "" {
		return RelPath{}, errorfamily.Newf(errorfamily.Rejection, "domain.relpath", "empty blob path")
	}
	if strings.HasPrefix(raw, "/") || strings.ContainsRune(raw, '\\') {
		return RelPath{}, errorfamily.Newf(errorfamily.Rejection, "domain.relpath", "path %q is not relative to the blob root", raw)
	}
	for _, seg := range strings.Split(raw, "/") {
		if seg == ".." {
			return RelPath{}, errorfamily.Newf(errorfamily.Rejection, "domain.relpath", "path %q escapes the blob root", raw)
		}
	}
	return RelPath{value: raw}, nil
}

// MustParseRelPath is ParseRelPath for row-scan and importer sites: the
// writer only ever stores Save-produced paths, so an unparsable value
// is a corrupt row the operator must see at first read.
func MustParseRelPath(raw string) RelPath { return must(ParseRelPath(raw)) }

// String returns the stored relative path.
func (p RelPath) String() string { return p.value }

// IsZero reports whether the path is the zero value.
func (p RelPath) IsZero() bool { return p.value == "" }
