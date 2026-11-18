package authmap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/webphone/internal/domain"
)

func line(ext, did string) Line {
	return Line{Extension: domain.MustParseExtension(ext), DID: did}
}

func TestResolveNormalizesCaseAndWhitespace(t *testing.T) {
	m, err := New(map[string]User{
		"lars@example.com": {Email: "lars@example.com", DisplayName: "Lars", Lines: []Line{line("1000", "+49 30 000")}},
	}, map[string]string{"1000": fileWith(t, "secret")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	user, ok := m.Resolve("  LARS@Example.COM ")
	if !ok {
		t.Fatal("mapping lookup must normalize case and whitespace")
	}
	if user.SessionLine().Extension.String() != "1000" {
		t.Errorf("session line extension = %q, want 1000", user.SessionLine().Extension.String())
	}
	if _, ok := m.Resolve("nobody@example.com"); ok {
		t.Fatal("unmapped email must not resolve")
	}
}

func TestMappedByExtensionReverseResolves(t *testing.T) {
	m, err := New(map[string]User{
		"alice@example.com": {Email: "alice@example.com", DisplayName: "Alice", Lines: []Line{line("1001", "")}},
	}, map[string]string{"1001": fileWith(t, "secret")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	user, ok := m.MappedByExtension(domain.MustParseExtension("1001"))
	if !ok || user.Email != "alice@example.com" {
		t.Fatalf("MappedByExtension = (%v, %v), want alice", user.Email, ok)
	}
	if _, ok := m.MappedByExtension(domain.MustParseExtension("1009")); ok {
		t.Fatal("unmapped extension must not resolve")
	}
}

// TestMappedByLoginAndSharedLineDeterminism pins the multi-user rules
// (2026-10-28): on a SHARED line the plain reverse lookup is
// deterministic (sorted emails — Go map order must never pick whoami),
// while a session that knows its own login email resolves THAT human —
// each session names its own person, never a coin flip.
func TestMappedByLoginAndSharedLineDeterminism(t *testing.T) {
	m, err := New(map[string]User{
		"zoe@example.com":   {Email: "zoe@example.com", DisplayName: "Zoe", Lines: []Line{line("1001", "")}},
		"alice@example.com": {Email: "alice@example.com", DisplayName: "Alice", Lines: []Line{line("1001", "")}},
	}, map[string]string{"1001": fileWith(t, "secret")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ext := domain.MustParseExtension("1001")

	for i := range 20 {
		user, ok := m.MappedByExtension(ext)
		if !ok || user.Email != "alice@example.com" {
			t.Fatalf("MappedByExtension iteration %d = (%v, %v), want deterministic alice", i, user.Email, ok)
		}
	}

	if user, ok := m.MappedByLogin(ext, "zoe@example.com"); !ok || user.DisplayName != "Zoe" {
		t.Fatalf("MappedByLogin(zoe) = (%v, %v), want Zoe", user.DisplayName, ok)
	}
	if user, ok := m.MappedByLogin(ext, ""); !ok || user.DisplayName != "Alice" {
		t.Fatalf("MappedByLogin(empty) = (%v, %v), want the deterministic fallback Alice", user.DisplayName, ok)
	}
	if user, ok := m.MappedByLogin(ext, "stale@example.com"); !ok || user.DisplayName != "Alice" {
		t.Fatalf("MappedByLogin(stale email) = (%v, %v), want fallback Alice — a stale identity must not win", user.DisplayName, ok)
	}
	other := domain.MustParseExtension("1002")
	if user, ok := m.MappedByLogin(other, "alice@example.com"); ok {
		t.Fatalf("MappedByLogin(email not owning the session line) = %v, want no match", user.Email)
	}
}

// TestLineModelPinsOwnershipShape pins the line-ownership model itself
// (2026-11-18): lines ride WITH their DID (one join, not parallel maps),
// the FIRST line is the session default, Has answers ownership, and
// Display names the ONE fallback rule (DID when declared, extension
// otherwise).
func TestLineModelPinsOwnershipShape(t *testing.T) {
	user := User{
		Email: "lars@example.com", DisplayName: "Lars",
		Lines: []Line{line("1000", "+49 30 111"), line("1001", "")},
	}
	if got := user.SessionLine().Extension.String(); got != "1000" {
		t.Errorf("SessionLine extension = %q, want the first mapped line 1000", got)
	}
	if !user.Has(domain.MustParseExtension("1001")) {
		t.Error("Has(1001) = false, want true — the second owned line")
	}
	if user.Has(domain.MustParseExtension("1002")) {
		t.Error("Has(1002) = true, want false — nobody's line")
	}
	if got := user.Extensions()[1].String(); got != "1001" {
		t.Errorf("Extensions()[1] = %q, want mapping order preserved", got)
	}
	if got := user.Lines[0].Display(); got != "+49 30 111" {
		t.Errorf("Display with DID = %q, want the DID verbatim", got)
	}
	if got := user.Lines[1].Display(); got != "1001" {
		t.Errorf("Display without DID = %q, want the bare extension", got)
	}
	if (Line{}).IsZero() != true || user.Lines[0].IsZero() != false {
		t.Error("IsZero must answer true only for the extension-less line")
	}
}

func TestSIPPasswordTrimsAndFailsClosed(t *testing.T) {
	file := fileWith(t, "  dir-password\n")
	m, err := New(map[string]User{
		"lars@example.com": {Email: "lars@example.com", Lines: []Line{line("1000", "")}},
	}, map[string]string{"1000": file})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ext := domain.MustParseExtension("1000")
	if got, err := m.SIPPassword(ext); err != nil || got != "dir-password" {
		t.Fatalf("SIPPassword = (%q, %v), want the trimmed password", got, err)
	}

	// Empty file: the 2026-10-01 outage class.
	if err := os.WriteFile(file, []byte(" \n"), 0o640); err != nil {
		t.Fatal(err)
	}
	_, err = m.SIPPassword(ext)
	errorfamilytest.AssertCode(t, err, "authmap.password_file.empty")

	// Unreadable file: the read code, never a silent empty password.
	if err := os.WriteFile(file, []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(file, 0o640) })
	_, err = m.SIPPassword(ext)
	errorfamilytest.AssertCode(t, err, "authmap.password_file.read")

	// Extension with no configured file at all.
	_, err = m.SIPPassword(domain.MustParseExtension("1009"))
	errorfamilytest.AssertCode(t, err, "authmap.password_file.missing")
}

func TestNewRejectsMappedExtensionWithoutPasswordFile(t *testing.T) {
	_, err := New(map[string]User{
		"lars@example.com": {Email: "lars@example.com", Lines: []Line{line("1000", "")}},
	}, nil)
	errorfamilytest.AssertCode(t, err, "authmap.password_file.missing")
}

func TestNewRejectsUnownableOwnershipShapes(t *testing.T) {
	_, err := New(map[string]User{
		"lars@example.com": {Email: "lars@example.com", Lines: nil},
	}, map[string]string{"1000": fileWith(t, "secret")})
	errorfamilytest.AssertCode(t, err, "authmap.user.lines")

	_, err = New(map[string]User{
		"lars@example.com": {Email: "lars@example.com", Lines: []Line{line("1000", ""), line("1000", "+49 30 000")}},
	}, map[string]string{"1000": fileWith(t, "secret")})
	errorfamilytest.AssertCode(t, err, "authmap.user.duplicate_line")
}

func TestNilMapFailsClosed(t *testing.T) {
	var m *Map
	if _, ok := m.Resolve("lars@example.com"); ok {
		t.Fatal("nil map must not resolve")
	}
	if _, err := m.SIPPassword(domain.MustParseExtension("1000")); err == nil {
		t.Fatal("nil map must reject a password read")
	}
}

func fileWith(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte(content), 0o640); err != nil {
		t.Fatal(err)
	}
	return path
}
