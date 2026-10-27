package readmodels

import (
	"errors"
	"testing"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/store"
)

func TestNotFoundKeepsSentinelAndCodedHalves(t *testing.T) {
	verdict := notFound(errorfamily.Newf(errorfamily.Rejection, "readmodels.fax_missing", "fax %s", "fax-1"))

	if !errors.Is(verdict, store.ErrNotFound) {
		t.Fatal("sentinel not errors.Is-matchable through the envelope")
	}
	coded, ok := errors.AsType[*errorfamily.Error](verdict)
	if !ok || coded.ErrorCode() != "readmodels.fax_missing" {
		t.Fatalf("coded half lost through the envelope: %v", verdict)
	}
	if coded.ErrorFamily() != errorfamily.Rejection {
		t.Fatalf("family = %v, want Rejection", coded.ErrorFamily())
	}
	want := store.ErrNotFound.Error() + "\n" + coded.Error()
	if verdict.Error() != want {
		t.Fatalf("message = %q, want %q", verdict.Error(), want)
	}
	single := &notFoundError{errs: []error{store.ErrNotFound}}
	if single.Error() != store.ErrNotFound.Error() {
		t.Fatalf("single-cause message = %q, want bare cause", single.Error())
	}
}
