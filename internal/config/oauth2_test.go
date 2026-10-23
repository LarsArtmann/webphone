package config

import (
	"testing"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
)

func validOAuth2() OAuth2 {
	return OAuth2{
		Issuer:           "https://auth.example.org",
		ClientID:         "webphone",
		ClientSecretFile: "/run/secrets/webphone-oidc",
		RedirectURL:      "https://pbx.example.org/api/auth/oauth2/callback",
		Scopes:           []string{"openid", "email", "profile"},
	}
}

func TestValidateOAuth2AcceptsTheCanonicalShape(t *testing.T) {
	if err := validateOAuth2(validOAuth2()); err != nil {
		t.Fatalf("validateOAuth2: %v", err)
	}
	if err := validateOAuth2(OAuth2{}); err != nil {
		t.Fatalf("disabled mode must validate clean: %v", err)
	}
}

func TestValidateOAuth2FailsClosedOnEveryHalfShape(t *testing.T) {
	noIssuer := validOAuth2()
	noIssuer.Issuer = ""
	httpIssuer := validOAuth2()
	httpIssuer.Issuer = "http://auth.example.org"
	queryIssuer := validOAuth2()
	queryIssuer.Issuer = "https://auth.example.org?x=1"
	noClient := validOAuth2()
	noClient.ClientID = ""
	noSecret := validOAuth2()
	noSecret.ClientSecretFile = ""
	noRedirect := validOAuth2()
	noRedirect.RedirectURL = ""
	plainRedirect := validOAuth2()
	plainRedirect.RedirectURL = "http://pbx.example.org/api/auth/oauth2/callback"
	noOpenID := validOAuth2()
	noOpenID.Scopes = []string{"email", "profile"}

	for _, tc := range []struct {
		name string
		cfg  OAuth2
		code string
	}{
		{"missing issuer", noIssuer, "config.auth.oauth2.issuer"},
		{"http issuer", httpIssuer, "config.auth.oauth2.issuer"},
		{"issuer with query", queryIssuer, "config.auth.oauth2.issuer"},
		{"missing client id", noClient, "config.auth.oauth2.client_id"},
		{"missing client secret file", noSecret, "config.auth.oauth2.client_secret_file"},
		{"missing redirect", noRedirect, "config.auth.oauth2.redirect_url"},
		{"plain-http redirect", plainRedirect, "config.auth.oauth2.redirect_url"},
		{"scopes without openid", noOpenID, "config.auth.oauth2.scopes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOAuth2(tc.cfg)
			if err == nil {
				t.Fatal("validateOAuth2: want error")
			}
			errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
			errorfamilytest.AssertCode(t, err, tc.code)
		})
	}
}

func TestResolvedAuthFoldsTheSharedMappingDown(t *testing.T) {
	auth := Auth{
		Users: map[string]PasskeyUser{
			"lars@example.com": {Extensions: []string{"1000"}, DisplayName: "Lars"},
		},
		ExtensionPasswordFiles: map[string]string{"1000": "/run/secrets/ext1000"},
	}
	resolved, err := resolvedAuth(auth)
	if err != nil {
		t.Fatalf("resolvedAuth: %v", err)
	}
	if _, ok := resolved.Passkey.Users["lars@example.com"]; !ok {
		t.Fatal("the shared map must reach the passkey shape")
	}
	if resolved.Passkey.ExtensionPasswordFiles["1000"] != "/run/secrets/ext1000" {
		t.Fatal("the shared password files must reach the passkey shape")
	}
}

func TestResolvedAuthRejectsDivergentSources(t *testing.T) {
	_, err := resolvedAuth(Auth{
		Users: map[string]PasskeyUser{
			"lars@example.com": {Extensions: []string{"1000"}},
		},
		Passkey: Passkey{
			Users: map[string]PasskeyUser{
				"lars@example.com": {Extensions: []string{"1001"}},
			},
		},
	})
	errorfamilytest.AssertCode(t, err, "config.auth.users")

	_, err = resolvedAuth(Auth{
		Users:                  map[string]PasskeyUser{"lars@example.com": {Extensions: []string{"1000"}}},
		ExtensionPasswordFiles: map[string]string{"1000": "/run/secrets/a"},
		Passkey: Passkey{
			Users:                  map[string]PasskeyUser{"lars@example.com": {Extensions: []string{"1000"}}},
			ExtensionPasswordFiles: map[string]string{"1000": "/run/secrets/b"},
		},
	})
	errorfamilytest.AssertCode(t, err, "config.auth.extension_password_files")
}

func TestResolvedAuthKeepsTheLegacyPasskeyShapeWorking(t *testing.T) {
	auth := Auth{Passkey: Passkey{Users: map[string]PasskeyUser{
		"lars@example.com": {Extensions: []string{"1000"}},
	}}}
	resolved, err := resolvedAuth(auth)
	if err != nil {
		t.Fatalf("resolvedAuth: %v", err)
	}
	if _, ok := resolved.Passkey.Users["lars@example.com"]; !ok {
		t.Fatal("auth.passkey.users must stay the historical home")
	}
}
