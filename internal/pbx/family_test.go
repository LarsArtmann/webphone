// PBX family pins (family-adoption train, 2026-09-30): the sentinels
// classify via registration (Infrastructure for disabled, Rejection
// for bad credentials) and keep their errors.Is identity; transport
// and decode failures are Transient; a provider 4xx is a Rejection and
// a 5xx Transient — mirroring the gateway seam's split.
package pbx_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/pbx"
)

func TestSentinelsClassify(t *testing.T) {
	errorfamilytest.AssertFamily(t, pbx.ErrDisabled, errorfamily.Infrastructure)
	errorfamilytest.AssertFamily(t, pbx.ErrUnauthorized, errorfamily.Rejection)
}

func TestTransportFailureIsTransient(t *testing.T) {
	// A server that is already closed makes client.Do fail.
	closed := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := closed.URL
	closed.Close()

	client, err := pbx.NewClient(url)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.History(t.Context(), pbx.Credentials{Extension: "1001", Password: "pw"}, 10)
	if err == nil {
		t.Fatal("dial refused: want error")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	errorfamilytest.AssertCode(t, err, "pbx.transport")
	errorfamilytest.AssertRetryable(t, err, true)
	errorfamilytest.AssertFamily(t, fmt.Errorf("verify: %w", err), errorfamily.Transient)
}

func TestStatusSplit(t *testing.T) {
	for _, tc := range []struct {
		status int
		family errorfamily.Family
	}{
		{http.StatusNotFound, errorfamily.Rejection},
		{http.StatusInternalServerError, errorfamily.Transient},
	} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			client, err := pbx.NewClient(srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.History(t.Context(), pbx.Credentials{Extension: "1001", Password: "pw"}, 10)
			if err == nil {
				t.Fatal("non-2xx: want error")
			}
			if errors.Is(err, pbx.ErrUnauthorized) {
				t.Fatalf("status %d must not classify as unauthorized", tc.status)
			}
			errorfamilytest.AssertFamily(t, err, tc.family)
			errorfamilytest.AssertCode(t, err, "pbx.http")
		})
	}
}
