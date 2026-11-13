// Readmodels family pins (T07 swap, 2026-10-15): the contacts read
// model owns the ErrListFull sentinel — the wrap keeps its errors.Is
// identity (the contacts_api 422 branch depends on it). The snippets
// swap (T09) homed ErrSnippetListFull the same way. The live-engine
// delete-miss pins retired with the engine stores (2026-11-11 train C:
// the cqrs backend's own delete tests carry the contract).
package readmodels_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

func TestErrListFullClassifiesRejection(t *testing.T) {
	errorfamilytest.AssertFamily(t, readmodels.ErrListFull, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, readmodels.ErrSnippetListFull, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, readmodels.ErrReviewListFull, errorfamily.Rejection)
}

// TestErrNoteListFullClassifiesConflict pins the one 409-shaped cap:
// the note list is the only cap whose HTTP surface answers 409, so the
// family must say Conflict (the owner's STATE conflicts and can be
// resolved by deleting another note) — the siblings answer 422 and
// stay Rejections (Conflict-family train, 2026-11-04).
func TestErrNoteListFullClassifiesConflict(t *testing.T) {
	errorfamilytest.AssertFamily(t, readmodels.ErrNoteListFull, errorfamily.Conflict)
}

func TestListCapWrapKeepsIdentityAndFamily(t *testing.T) {
	owner := domain.MustParseExtension("1001")
	full := errorfamily.WrapRejectionf(readmodels.ErrListFull, "readmodels.contact_list_full", "%s at %d contacts", owner, readmodels.ContactsMaxPerExtension)
	errorfamilytest.AssertFamily(t, full, errorfamily.Rejection)
	if !errors.Is(full, readmodels.ErrListFull) {
		t.Fatal("the list-full wrap must keep errors.Is identity")
	}
	wrapped := fmt.Errorf("save contact: %w", full)
	if !errors.Is(wrapped, readmodels.ErrListFull) {
		t.Fatal("identity must survive service wraps")
	}
	errorfamilytest.AssertFamily(t, wrapped, errorfamily.Rejection)
}
