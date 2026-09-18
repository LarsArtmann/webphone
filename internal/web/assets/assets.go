// Package assets embeds every static file the webphone serves: the SIP call
// island (ES modules + stylesheet), the vendored sip.js browser bundle, and
// the compiled Tailwind stylesheet. Everything ships same-origin from the
// binary — the strict CSP of the serving vhost allows nothing else.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:island all:vendor app.css
var embedded embed.FS

// FS returns the embedded assets rooted at the package directory, so paths
// are "island/...", "vendor/sip.min.js", "app.css".
func FS() fs.FS {
	return embedded
}
