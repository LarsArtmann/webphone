package readmodels

import (
	"strings"

	"github.com/larsartmann/webphone/internal/store"
)

// notFoundError is the ONE home for a collection-miss verdict: the
// store.ErrNotFound sentinel (errors.Is-matchable, exactly the shape
// the SQL store answered) rides beside the op's coded Rejection so the
// sentinel drives caller control flow while the code names the seam in
// logs. It reproduces errors.Join's message format (one line per
// cause) and its Unwrap() []error traversal, so sentinel matching and
// family classification see both halves.
type notFoundError struct {
	errs []error
}

func (e *notFoundError) Error() string {
	switch len(e.errs) {
	case 0:
		return ""
	case 1:
		return e.errs[0].Error()
	}
	var b strings.Builder
	for i, err := range e.errs {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

func (e *notFoundError) Unwrap() []error { return e.errs }

// notFound builds the miss verdict for one op: the ErrNotFound sentinel
// plus the op's coded Rejection. The errorfamily constructor stays at
// the call site so the arch registry scan sees every code literal; this
// one home owns only the sentinel pairing.
func notFound(coded error) *notFoundError {
	return &notFoundError{errs: []error{store.ErrNotFound, coded}}
}
