// Readmodels family pins (T07 swap, 2026-10-15): the contacts read
// model owns the ErrListFull sentinel now — the wrap keeps its
// errors.Is identity (the contacts_api 422 branch depends on it), and
// the delete-missing Join stays ErrNotFound-matchable with the
// Rejection family. The snippets swap (T09) homed ErrSnippetListFull
// the same way; its pins ride the same shape.
package readmodels_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

func TestErrListFullClassifiesRejection(t *testing.T) {
	errorfamilytest.AssertFamily(t, readmodels.ErrListFull, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, readmodels.ErrSnippetListFull, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, readmodels.ErrNoteListFull, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, readmodels.ErrReviewListFull, errorfamily.Rejection)
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

func TestDeleteMissingJoinsStayNotFound(t *testing.T) {
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })

	err = engine.Contacts().Delete(context.Background(), domain.MustParseExtension("1001"), domain.GenerateContactID())
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("delete of missing row: want ErrNotFound, got %v", err)
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)

	err = engine.Snippets().Delete(context.Background(), domain.MustParseExtension("1001"), domain.GenerateSnippetID())
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("snippet delete of missing row: want ErrNotFound, got %v", err)
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
}
