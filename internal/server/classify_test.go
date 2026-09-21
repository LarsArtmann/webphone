// classifyForUser pins (SUPERB error-excellence T04): the family→status
// half of the failure→feedback table. Rejection answers 422, every
// system-side family answers 502 with the transport copy — including the
// fail-open default for unknown errors (an untagged error classifies as
// Transient, never as a user mistake).
package server

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/messaging"
)

func TestClassifyForUserMapsFamiliesToStatuses(t *testing.T) {
	providerRefused := &gateway.ErrProviderRejected{Status: http.StatusForbidden, Detail: "no"}
	providerDown := &gateway.ErrProviderRejected{Status: http.StatusBadGateway, Detail: "down"}
	transport := errorfamily.WrapTransientf(errors.New("dial refused"), "gateway.transport", "provider call")
	infrastructure := errorfamily.WrapInfrastructuref(errors.New("disk full"), "store.message_append", "persist message")
	unknown := errors.New("boom") // untagged: the library default applies

	for _, tc := range []struct {
		name string
		err  error
		want int
	}{
		{"validation is 422", &messaging.ErrInvalidSend{Reason: "no text"}, http.StatusUnprocessableEntity},
		{"fax validation is 422", &fax.ErrInvalidFax{Reason: "no pdf"}, http.StatusUnprocessableEntity},
		{"provider 4xx stays 502 with its detail fast path", providerRefused, http.StatusBadGateway},
		{"provider 5xx is 502", providerDown, http.StatusBadGateway},
		{"transport is 502", transport, http.StatusBadGateway},
		{"infrastructure is 502", infrastructure, http.StatusBadGateway},
		{"unknown error is 502 generic, never a client error", unknown, http.StatusBadGateway},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyForUser(tc.err); got != tc.want {
				t.Errorf("classifyForUser(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}

	t.Run("classification survives the service wrap", func(t *testing.T) {
		wrapped := fmt.Errorf("gateway: %w", providerRefused)
		if got := classifyForUser(wrapped); got != http.StatusBadGateway {
			t.Errorf("wrapped rejection: %d, want 502", got)
		}
		wrapped = fmt.Errorf("gateway: %w", transport)
		if got := classifyForUser(wrapped); got != http.StatusBadGateway {
			t.Errorf("wrapped transport: %d, want 502", got)
		}
	})
}
