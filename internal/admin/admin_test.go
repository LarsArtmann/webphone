package admin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/domain"
)

func ext(t *testing.T, raw string) domain.Extension {
	t.Helper()
	parsed, err := domain.ParseExtension(raw)
	if err != nil {
		t.Fatalf("parse extension %q: %v", raw, err)
	}
	return parsed
}

// passwordFile writes a one-line password file and returns its path.
func passwordFile(t *testing.T, dir, password string) string {
	t.Helper()
	path := filepath.Join(dir, "pw-"+password+".txt")
	if err := os.WriteFile(path, []byte(password+"\n"), 0o600); err != nil {
		t.Fatalf("write password file: %v", err)
	}
	return path
}

func newTestService(t *testing.T, identities map[string]string, users map[string]authmap.User, passwordFiles map[string]string) *Service {
	t.Helper()
	mapped, err := authmap.New(users, passwordFiles)
	if err != nil {
		t.Fatalf("authmap.New: %v", err)
	}
	service, err := New(mapped, identities)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return service
}

func TestIsAdminFailClosed(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{"1001": passwordFile(t, dir, "alpha"), "1002": passwordFile(t, dir, "beta")}
	service := newTestService(t, nil, map[string]authmap.User{
		"boss@example.com":   {Email: "boss@example.com", Extensions: []domain.Extension{ext(t, "1001")}, Admin: true},
		"worker@example.com": {Email: "worker@example.com", Extensions: []domain.Extension{ext(t, "1002")}},
	}, files)

	if !service.IsAdmin("boss@example.com") {
		t.Error("mapped admin email must resolve as admin")
	}
	if !service.IsAdmin("Boss@Example.COM") {
		t.Error("email normalization must match the authmap contract (case-insensitive)")
	}
	if service.IsAdmin("worker@example.com") {
		t.Error("mapped non-admin must not resolve as admin")
	}
	if service.IsAdmin("stranger@example.com") {
		t.Error("unmapped email must not resolve as admin")
	}
	if service.IsAdmin("") {
		t.Error("empty email (extension login) must never resolve as admin")
	}

	var nilService *Service
	if nilService.IsAdmin("boss@example.com") {
		t.Error("nil service must answer false (mode off)")
	}
	if nilService.Users() != nil || nilService.Numbers() != nil || nilService.Extensions() != nil {
		t.Error("nil service must list nothing")
	}
}

func TestNumbersInventoriesMappedAndIdentityLines(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{"1001": passwordFile(t, dir, "alpha"), "1002": passwordFile(t, dir, "beta")}
	service := newTestService(t,
		map[string]string{"1001": "+49301111", "1099": "+49309999"},
		map[string]authmap.User{
			"boss@example.com":   {Email: "boss@example.com", Extensions: []domain.Extension{ext(t, "1001")}, Admin: true},
			"worker@example.com": {Email: "worker@example.com", Extensions: []domain.Extension{ext(t, "1001"), ext(t, "1002")}},
		}, files)

	numbers := service.Numbers()
	var got []string
	for _, number := range numbers {
		got = append(got, number.Extension.String())
	}
	want := []string{"1001", "1002", "1099"}
	if len(got) != len(want) {
		t.Fatalf("extensions = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("extensions = %v, want %v (sorted union of mapping + identities)", got, want)
		}
	}
	byExtension := make(map[string]Number, len(numbers))
	for _, number := range numbers {
		byExtension[number.Extension.String()] = number
	}
	if got, want := byExtension["1001"].DID, "+49301111"; got != want {
		t.Errorf("1001 DID = %q, want %q", got, want)
	}
	if got, want := byExtension["1002"].DID, ""; got != want {
		t.Errorf("1002 (mapped, no identity) DID = %q, want empty", got)
	}
	if got, want := byExtension["1099"].Users, []string(nil); len(got) != len(want) {
		t.Errorf("1099 (identity, unmapped) users = %v, want none", got)
	}
	users := byExtension["1001"].Users
	if len(users) != 2 || users[0] != "boss@example.com" || users[1] != "worker@example.com" {
		t.Errorf("1001 users = %v, want sorted [boss worker]", users)
	}
}

func TestExtensionsSortedDeduplicated(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{"1002": passwordFile(t, dir, "beta"), "1001": passwordFile(t, dir, "alpha")}
	service := newTestService(t, nil, map[string]authmap.User{
		"boss@example.com":   {Email: "boss@example.com", Extensions: []domain.Extension{ext(t, "1002"), ext(t, "1001")}, Admin: true},
		"worker@example.com": {Email: "worker@example.com", Extensions: []domain.Extension{ext(t, "1002")}},
	}, files)

	extensions := service.Extensions()
	if len(extensions) != 2 || extensions[0].String() != "1001" || extensions[1].String() != "1002" {
		t.Errorf("extensions = %v, want sorted deduplicated [1001 1002]", extensions)
	}
}

func TestSIPPasswordDelegatesToAuthmap(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{"1001": passwordFile(t, dir, "alpha")}
	service := newTestService(t, nil, map[string]authmap.User{
		"boss@example.com": {Email: "boss@example.com", Extensions: []domain.Extension{ext(t, "1001")}, Admin: true},
	}, files)

	password, err := service.SIPPassword(ext(t, "1001"))
	if err != nil || password != "alpha" {
		t.Errorf("SIPPassword = %q, %v; want alpha, nil", password, err)
	}
	if _, err := service.SIPPassword(ext(t, "4321")); err == nil {
		t.Error("unmapped extension must fail closed")
	}
}

func TestNewRejectsUnnormalizedIdentityKey(t *testing.T) {
	mapped, err := authmap.New(nil, nil)
	if err != nil {
		t.Fatalf("authmap.New: %v", err)
	}
	if _, err := New(mapped, map[string]string{"10 01": "+4930"}); err == nil {
		t.Error("a non-extension identities key must fail loudly (composition defect)")
	}
}
