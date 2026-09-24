// Service-level classification pins (SUPERB error-excellence T02): the
// validation refusals the server renders as 422 must classify as
// Rejection, and the classification must survive the %w wraps the Send
// path adds on its way up (the server switch classifies the error it
// receives, not the one the service constructed).
package messaging_test

import (
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/messaging"
)

func TestValidationRefusalsAreRejections(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"empty send", &messaging.ErrInvalidSend{Reason: "a message needs text or an attachment"}},
		{"body too long", &messaging.ErrInvalidSend{Reason: "message longer than 1600 characters"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errorfamilytest.AssertFamily(t, tc.err, errorfamily.Rejection)
			wrapped := fmt.Errorf("send: %w", tc.err)
			errorfamilytest.AssertFamily(t, wrapped, errorfamily.Rejection)
		})
	}
}
