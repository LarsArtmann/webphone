package config

import (
	"os"
	"path/filepath"
	"testing"
)

// validPasskey builds the minimal complete passkey config pointing at a
// password file that exists.
func validPasskey(t *testing.T) Passkey {
	t.Helper()
	dir := t.TempDir()
	passFile := filepath.Join(dir, "ext1000")
	if err := os.WriteFile(passFile, []byte("secret\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	return Passkey{
		RPID:                   "localhost",
		RPOrigins:              []string{"http://localhost:18099", "https://localhost"},
		Users:                  map[string]PasskeyUser{"lars@example.com": {Extensions: []string{"1000"}}},
		ExtensionPasswordFiles: map[string]string{"1000": passFile},
	}
}

func TestValidatePasskeyAcceptsPortCarryingOrigins(t *testing.T) {
	// WebAuthn RP IDs are domain-scoped: a loopback dev origin carries a
	// port and must not fail the rp_id match.
	if err := validatePasskey(validPasskey(t)); err != nil {
		t.Fatalf("valid config with ported origins rejected: %v", err)
	}
}

func TestValidatePasskeyRejectsDomainMismatch(t *testing.T) {
	cfg := validPasskey(t)
	cfg.RPID = "pbx.example.org"
	if err := validatePasskey(cfg); err == nil {
		t.Fatal("rp_id unequal to every origin host must reject")
	}
}

func TestValidatePasskeyIsAllOrNothing(t *testing.T) {
	if err := validatePasskey(Passkey{}); err != nil {
		t.Fatalf("empty passkey config (mode off) must pass: %v", err)
	}
	half := validPasskey(t)
	half.Users = nil
	if err := validatePasskey(half); err == nil {
		t.Fatal("half a passkey config (no users) must reject — fail closed, not half-live")
	}
}

func TestValidatePasskeyRequiresPasswordFilePerMappedExtension(t *testing.T) {
	cfg := validPasskey(t)
	cfg.ExtensionPasswordFiles = map[string]string{"1001": "/tmp/other"}
	if err := validatePasskey(cfg); err == nil {
		t.Fatal("a mapped extension without its own password file must reject")
	}
}

func TestValidatePasskeyRejectsUnnormalizedExtensions(t *testing.T) {
	cfg := validPasskey(t)
	users := cfg.Users
	users["lars@example.com"] = PasskeyUser{Extensions: []string{" 1000"}}
	cfg.Users = users
	if err := validatePasskey(cfg); err == nil {
		t.Fatal("unnormalized extension keys must reject (silently unmatchable lookups otherwise)")
	}
}
