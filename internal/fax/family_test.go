// Service-level classification pins (SUPERB error-excellence T02): fax
// validation refusals must classify as Rejection and survive the %w
// wraps the Send path adds on its way to the server switch.
package fax_test

import (
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
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
			if family := errorfamily.Classify(tc.err); family != errorfamily.Rejection {
				t.Errorf("Classify(%v) = %s, want rejection", tc.err, family)
			}
			wrapped := fmt.Errorf("send: %w", tc.err)
			if family := errorfamily.Classify(wrapped); family != errorfamily.Rejection {
				t.Errorf("Classify through wrap = %s, want rejection", family)
			}
		})
	}
}
