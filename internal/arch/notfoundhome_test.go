package arch

import (
	"os"
	"strings"
	"testing"
)

// TestHTMLRouterFortyFoursRouteThroughTheStyledHome pins the prose rule
// at notfound.go/actions.go as a checkable invariant (the T22 open leg
// f47): every HTML-router MISS lands on the styled 404 home — the
// catch-all registration routes through the notfound.go handlers, and
// the ROUTER REGISTRATION files never write a bare stdlib
// http.NotFound. actions.go's two http.NotFound sites are byte-stream
// lanes (attachment/fax-document fetches — the terse-answer doctrine's
// own list in TODO_LIST's quarterly re-rule row), which is why the
// scan covers server.go/pages.go (registration + page surfaces) and
// not the stream actions.
func TestHTMLRouterFortyFoursRouteThroughTheStyledHome(t *testing.T) {
	routerFiles := []string{
		"../server/server.go",
		"../server/pages.go",
	}
	var catchAll bool
	for _, path := range routerFiles {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(raw)
		if strings.Contains(src, "http.NotFound(") {
			t.Errorf("%s writes a bare stdlib 404 — the HTML surface answers through the notfound.go home (page or .wp-error by HX-Request), never http.NotFound", path)
		}
		if strings.Contains(src, `HandleFunc("/", h.notFoundPage)`) {
			catchAll = true
		}
	}
	if !catchAll {
		t.Error("the protected HTML router must register its catch-all on h.notFoundPage (the styled 404) — an unregistered path would fall through to the mux's default NotFound")
	}
}
