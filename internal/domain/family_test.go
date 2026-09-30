// Domain family pins (family-adoption train, 2026-09-30): validation
// refusals classify Rejection (they feed the 400/422 lanes), while a
// malformed stored id is Corruption (data at rest, never user input —
// see the must/parseID doc). The families must survive the handler
// wraps these errors cross on their way out.
package domain_test

import (
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/domain"
)

func TestValidationRefusalsAreRejections(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		code string
	}{
		{"empty extension", mustErr(domain.ParseExtension("?!")), "domain.extension"},
		{"long extension", mustErr(domain.ParseExtension("123456789012345678901234567890123")), "domain.extension"},
		{"empty number", mustErr(domain.ParsePhone("")), "domain.phone"},
		{"long number", mustErr(domain.ParsePhone("123456789012345678901234567890123")), "domain.phone"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errorfamilytest.AssertFamily(t, tc.err, errorfamily.Rejection)
			errorfamilytest.AssertCode(t, tc.err, tc.code)
			errorfamilytest.AssertRetryable(t, tc.err, false)
			errorfamilytest.AssertFamily(t, fmt.Errorf("handler: %w", tc.err), errorfamily.Rejection)
		})
	}
}

func TestCorruptStoredIDIsCorruption(t *testing.T) {
	_, err := domain.ParseThreadID("Thread:short")
	if err == nil {
		t.Fatal("short id: want error")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Corruption)
	errorfamilytest.AssertCode(t, err, "domain.id")
}

// mustErr unwraps a Parse result for table entries; the inputs above
// are chosen to always fail.
func mustErr[T any](_ T, err error) error { return err }
