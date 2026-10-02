package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStylesUseLogicalProperties keeps the two owned stylesheets
// RTL-ready (M24.5, idea L4): reading-flow spacing, borders and
// alignment use LOGICAL properties (margin-inline-*, padding-inline-*,
// border-inline-*, text-align: start) so a future RTL locale mirrors
// correctly. Position offsets for fixed chrome (left:/right: anchors)
// are exempt — they are deliberately physical placements, and the two
// snippet/lightbox offsets already use inset-inline-*.
func TestStylesUseLogicalProperties(t *testing.T) {
	banned := []string{
		"margin-left", "margin-right",
		"padding-left", "padding-right",
		"border-left", "border-right",
		"text-align: left", "text-align: right",
	}
	for _, name := range []string{"app.css", filepath.Join("island", "style.css")} {
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for i, line := range strings.Split(string(raw), "\n") {
			trimmed := strings.TrimSpace(line)
			for _, physical := range banned {
				if strings.HasPrefix(trimmed, physical) {
					t.Errorf("%s:%d uses physical %q — use the logical property (RTL readiness, M24.5)", name, i+1, physical)
				}
			}
		}
	}
}
