// Service-level classification pins (SUPERB error-excellence T02): fax
// validation refusals must classify as Rejection and survive the %w
// wraps the Send path adds on its way to the server switch.
package fax_test

import (
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/fax"
)

func TestValidationRefusalsAreRejections(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"no document", &fax.ErrInvalidFax{Reason: "attach a PDF to send"}},
		{"not a pdf", &fax.ErrInvalidFax{Reason: "only PDF documents can be faxed"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errorfamilytest.AssertFamily(t, tc.err, errorfamily.Rejection)
			wrapped := fmt.Errorf("send: %w", tc.err)
			errorfamilytest.AssertFamily(t, wrapped, errorfamily.Rejection)
		})
	}
}
