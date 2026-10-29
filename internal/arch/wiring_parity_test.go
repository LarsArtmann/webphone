package arch

import (
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/larsartmann/webphone/internal/server"
)

// TestAppWiresEveryServerDep pins the composition-root parity the M16
// Notes bug class burned: server.Deps grew a field (Notes), the test
// harness wired it, and app.go — the ONLY production composition — sat
// silently without it, so every production note endpoint answered 503
// while the suite stayed green. Every field the server.Deps struct
// declares (minus the explicit test-override seams below) MUST appear
// by name in app.go's server.Deps literal; adding a Deps field without
// landing it in the composition root now fails this test instead of
// production.
func TestAppWiresEveryServerDep(t *testing.T) {
	// Deliberately NOT wired by app.go: Deps.ProjectionStatus is the
	// projection-host override for tests (prod reads the status from
	// the container-resolved projection host instead).
	omitted := map[string]bool{"ProjectionStatus": true}

	src, err := os.ReadFile("../app/app.go")
	if err != nil {
		t.Fatalf("read app.go: %v", err)
	}
	start := regexp.MustCompile(`server\.Deps\{`).FindIndex(src)
	if start == nil {
		t.Fatal("no server.Deps literal found in app.go")
	}
	// Scan to the MATCHING close brace: fields like Shared carry nested
	// composite literals, so a non-greedy `[^}]*` would truncate the
	// match mid-literal and phantom-miss every field after it.
	depth := 0
	end := -1
	for i := start[1] - 1; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i
			}
		}
		if end >= 0 {
			break
		}
	}
	if end < 0 {
		t.Fatal("unterminated server.Deps literal in app.go")
	}
	literal := string(src[start[1]:end])
	wired := map[string]bool{}
	for _, name := range regexp.MustCompile(`([A-Z][A-Za-z0-9]*):`).FindAllStringSubmatch(literal, -1) {
		wired[name[1]] = true
	}

	typ := reflect.TypeFor[server.Deps]()
	for field := range typ.Fields() {
		name := field.Name
		if omitted[name] || wired[name] {
			continue
		}
		t.Errorf("server.Deps.%s is not wired in app.go's composition root — production answers the zero-value path while tests stay green (the M16 Notes 503 class)", name)
	}
}
