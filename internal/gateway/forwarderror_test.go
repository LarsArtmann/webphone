package gateway

import (
	"errors"
	"testing"

	"github.com/larsartmann/go-error-family"
)

func TestForwardErrorPrefixAndTraversal(t *testing.T) {
	cause := errorfamily.WrapTransientf(errors.New("dial failed"), "gateway.transport", "provider call")
	fwd := &ForwardError{Err: cause}

	if fwd.Error() != "gateway: "+cause.Error() {
		t.Fatalf("message = %q, want gateway-prefixed cause", fwd.Error())
	}
	if !errors.Is(fwd, cause) {
		t.Fatal("cause not reachable through Unwrap")
	}
	var coded *errorfamily.Error
	if !errors.As(fwd, &coded) || coded.ErrorCode() != "gateway.transport" {
		t.Fatalf("coded cause lost: %v", fwd)
	}
}

func TestFormErrorPassthrough(t *testing.T) {
	cause := errors.New("open attachment x: no such file")
	form := formErrf("open attachment %s: %w", "x", cause)

	if form.Error() != cause.Error() {
		t.Fatalf("message = %q, want cause text", form.Error())
	}
	if !errors.Is(form, cause) {
		t.Fatal("cause not reachable through Unwrap")
	}
}
