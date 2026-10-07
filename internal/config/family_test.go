// Config family pins (family-adoption train, 2026-09-30): operator
// input problems classify Rejection — the startup path can one day
// branch on the family instead of parsing prose. The both-sources
// sentinel is a classified *Error and survives Load's neutral wrap.
package config_test

import (
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/config"
)

func TestConfigValidationFailuresAreRejections(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  map[string]string
		code string
	}{
		{"bad gateway mode", map[string]string{"WEBPHONE_GATEWAY__MODE": "carrier-pigeon"}, "config.gateway.mode"},
		{"webhook without url", map[string]string{"WEBPHONE_GATEWAY__MODE": "webhook"}, "config.gateway.webhook_url"},
		{"negative ttl", map[string]string{"WEBPHONE_SESSION_TTL": "-1s"}, "config.session_ttl"},
		// The passkey mode is env-armable with one scalar (any set field
		// enables it); the then-missing list fields reject fail-closed.
		{"half a passkey config", map[string]string{"WEBPHONE_AUTH__PASSKEY__RP_ID": "pbx.example.org"}, "config.auth.passkey.rp_origins"},
		// ASR: a token without a URL (nowhere to send it) rejects; so
		// does a non-absolute URL.
		{"asr token without url", map[string]string{"WEBPHONE_ASR__TOKEN": "secret"}, "config.asr.url"},
		{"asr relative url", map[string]string{"WEBPHONE_ASR__URL": "not-a-url"}, "config.asr.url"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			t.Setenv("WEBPHONE_DATA_DIR", t.TempDir())
			_, err := config.Load()
			if err == nil {
				t.Fatal("Load: want error")
			}
			errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
			errorfamilytest.AssertCode(t, err, tc.code)
			errorfamilytest.AssertFamily(t, fmt.Errorf("load config: %w", err), errorfamily.Rejection)
		})
	}
}

func TestSecretBothSourcesClassifiesAndSurvivesTheWrap(t *testing.T) {
	t.Setenv("WEBPHONE_GATEWAY__MODE", "webhook")
	t.Setenv("WEBPHONE_GATEWAY__WEBHOOK_URL", "http://127.0.0.1:9")
	t.Setenv("WEBPHONE_GATEWAY__WEBHOOK_SECRET", "a")
	t.Setenv("WEBPHONE_GATEWAY__WEBHOOK_SECRET_FILE", "/anywhere")
	t.Setenv("WEBPHONE_DATA_DIR", t.TempDir())

	_, err := config.Load()
	if err == nil {
		t.Fatal("both secret sources: want error")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "config.gateway_secret_sources")
}
