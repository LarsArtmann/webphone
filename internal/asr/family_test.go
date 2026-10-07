// ASR seam classification pins: the disabled sentinel is Infrastructure
// (nil deps), a rejected token is a Rejection (operator's to fix), and
// validation refusals survive the %w wraps the caller adds.
package asr_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/asr"
)

func TestSentinelsClassify(t *testing.T) {
	errorfamilytest.AssertFamily(t, asr.ErrDisabled, errorfamily.Infrastructure)
	errorfamilytest.AssertFamily(t, asr.ErrUnauthorized, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, fmt.Errorf("transcribe: %w", asr.ErrUnauthorized), errorfamily.Rejection)
}

func TestEmptyAudioIsRejection(t *testing.T) {
	client, err := asr.NewClient("http://127.0.0.1:1", "", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Transcribe(context.Background(), asr.Request{})
	if err == nil {
		t.Fatal("want error")
	}
	if errors.Is(err, asr.ErrDisabled) {
		t.Fatal("empty-audio error must not be the disabled sentinel")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "asr.empty")
}
