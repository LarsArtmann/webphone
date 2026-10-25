package domain

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/larsartmann/go-cqrs-lite/catalog/v4"
	"github.com/larsartmann/go-cqrs-lite/catalog/v4/eventcatalog"
)

func TestBuildCatalog(t *testing.T) {
	t.Parallel()

	cat := BuildCatalog()

	if cat.Title != CatalogTitle {
		t.Errorf("Title = %q, want %q", cat.Title, CatalogTitle)
	}

	if cat.Version != CatalogVersion {
		t.Errorf("Version = %q, want %q", cat.Version, CatalogVersion)
	}

	if len(cat.Services) != 1 {
		t.Fatalf("len(Services) = %d, want 1", len(cat.Services))
	}

	testCatalogService(t, cat.Services[0])
	testCatalogDomain(t, cat)

	if len(cat.Users) != 1 || cat.Users[0].ID != "lars" {
		t.Errorf("Users = %v, want one user lars", cat.Users)
	}
}

func testCatalogService(t *testing.T, svc catalog.Service) {
	t.Helper()

	if svc.ID != catalogServiceID {
		t.Errorf("Service.ID = %q, want %q", svc.ID, catalogServiceID)
	}

	if !slices.Equal(svc.Owners, []string{"lars"}) {
		t.Errorf("Service.Owners = %v, want [lars]", svc.Owners)
	}

	if svc.Summary == "" {
		t.Error("Service.Summary is empty")
	}

	if len(svc.Commands) != 0 {
		t.Errorf("len(Commands) = %d, want 0 (webphone writes are HTTP, not messages)", len(svc.Commands))
	}

	wantEventIDs := []string{
		"message.received",
		"message.delivery.advanced",
		"thread.flag.changed",
		"fax.advanced",
	}

	if len(svc.Events) != len(wantEventIDs) {
		t.Fatalf("len(Events) = %d, want %d", len(svc.Events), len(wantEventIDs))
	}

	seen := make(map[catalog.MessageID]bool, len(svc.Events))
	for i, want := range wantEventIDs {
		msg := svc.Events[i]

		if string(msg.ID) != want {
			t.Errorf("Events[%d].ID = %q, want %q", i, msg.ID, want)
		}

		if seen[msg.ID] {
			t.Errorf("duplicate message ID in catalog: %s", msg.ID)
		}
		seen[msg.ID] = true

		if msg.Direction != catalog.Sends {
			t.Errorf("Events[%d] (%s).Direction = %q, want %q", i, msg.ID, msg.Direction, catalog.Sends)
		}

		// The zlota44 lesson: an unstamped version fails the hub's
		// governance gate, so pin it here at the source.
		if string(msg.Version) != CatalogVersion {
			t.Errorf("Events[%d] (%s).Version = %q, want %q", i, msg.ID, msg.Version, CatalogVersion)
		}

		if !slices.Equal(msg.Owners, []string{"lars"}) {
			t.Errorf("Events[%d] (%s).Owners = %v, want [lars]", i, msg.ID, msg.Owners)
		}

		if msg.Summary == "" {
			t.Errorf("Events[%d] (%s) has an empty summary", i, msg.ID)
		}

		if msg.Schema == nil {
			t.Errorf("Events[%d] (%s) has no derived schema", i, msg.ID)
		}
	}
}

func testCatalogDomain(t *testing.T, cat *catalog.Catalog) {
	t.Helper()

	if len(cat.Domains) != 1 {
		t.Fatalf("len(Domains) = %d, want 1", len(cat.Domains))
	}

	domain := cat.Domains[0]

	if domain.ID != "telephony" {
		t.Errorf("Domain.ID = %q, want telephony", domain.ID)
	}

	if string(domain.Version) != CatalogVersion {
		t.Errorf("Domain.Version = %q, want %q", domain.Version, CatalogVersion)
	}

	if len(domain.Services) != 1 || domain.Services[0] != catalogServiceID {
		t.Errorf("Domain.Services = %v, want [%s]", domain.Services, catalogServiceID)
	}

	if !slices.Equal(domain.Owners, []string{"lars"}) {
		t.Errorf("Domain.Owners = %v, want [lars]", domain.Owners)
	}
}

// TestCatalogSchemaParity guards the Catalog*Schema mirrors: every
// exported field of the real seam event must appear, by name, on its
// mirror — the mirror exists because the branded types hide their
// values from the schema reflector, and a drifting mirror would
// document a shape the event no longer has.
func TestCatalogSchemaParity(t *testing.T) {
	t.Parallel()

	pairs := []struct {
		event  any
		mirror any
	}{
		{MessageReceived{}, CatalogMessageReceivedSchema{}},
		{MessageDeliveryAdvanced{}, CatalogMessageDeliveryAdvancedSchema{}},
		{ThreadFlagChanged{}, CatalogThreadFlagChangedSchema{}},
		{FaxAdvanced{}, CatalogFaxAdvancedSchema{}},
	}

	for _, pair := range pairs {
		eventFields := exportedFieldNames(reflect.TypeOf(pair.event))
		mirrorFields := exportedFieldNames(reflect.TypeOf(pair.mirror))

		if !slices.Equal(eventFields, mirrorFields) {
			t.Errorf("%T and its catalog mirror %T drifted: fields %v vs %v",
				pair.event, pair.mirror, eventFields, mirrorFields)
		}
	}
}

func exportedFieldNames(t reflect.Type) []string {
	names := make([]string, 0, t.NumField())
	for f := range t.Fields() {
		if f := f; f.IsExported() {
			names = append(names, f.Name)
		}
	}
	return names
}

// TestCatalogExportFrontmatter runs the hub's exact export (default
// tree) against a temp dir and asserts what the federation gate checks:
// every resource file carries the catalog version, and the service and
// domain files exist where the hub's merge expects them.
func TestCatalogExportFrontmatter(t *testing.T) {
	t.Parallel()

	out := t.TempDir()
	tree := filepath.Join(out, "eventcatalog")

	if err := eventcatalog.NewExporter(tree).Export(BuildCatalog()); err != nil {
		t.Fatalf("export: %v", err)
	}

	wantVersionLine := "version: " + CatalogVersion
	checked := 0

	err := filepath.WalkDir(tree, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".mdx") {
			return nil
		}

		rel, err := filepath.Rel(tree, path)
		if err != nil {
			return err
		}

		// Users and teams are unversioned in EventCatalog; every other
		// resource kind carries the catalog version.
		if rel == "users" || rel == "teams" || strings.HasPrefix(rel, "users/") || strings.HasPrefix(rel, "teams/") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		frontmatter := string(content)
		if body, ok := strings.CutPrefix(frontmatter, "---\n"); ok {
			if before, _, ok := strings.Cut(body, "\n---"); ok {
				frontmatter = before
			}
		}
		if !strings.Contains(frontmatter, wantVersionLine) {
			t.Errorf("%s: frontmatter missing %q", rel, wantVersionLine)
		}
		checked++
		return nil
	})
	if err != nil {
		t.Fatalf("walk export tree: %v", err)
	}

	if checked == 0 {
		t.Fatal("export produced no .mdx files")
	}

	for _, want := range []string{
		filepath.Join("services", catalogServiceID, "index.mdx"),
		filepath.Join("domains", "telephony", "index.mdx"),
	} {
		if _, err := os.Stat(filepath.Join(tree, want)); err != nil {
			t.Errorf("expected %s in export tree: %v", want, err)
		}
	}
}
