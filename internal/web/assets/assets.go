// Package assets embeds every static file the webphone serves: the SIP call
// island (ES modules + stylesheet), the vendored sip.js browser bundle, the
// shell stylesheet, the shell glue + theme-preload scripts, and the
// standalone passkey-enrollment page module. Everything ships same-origin
// from the binary — the strict CSP of the serving vhost allows nothing
// else.
package assets

import (
	"embed"
	"io/fs"
	"sort"
)

//go:embed all:island all:vendor all:enroll app.css tw.css health.css shell.js theme-preload.js
var embedded embed.FS

// islandModuleURLs is the island's ESM graph as serving URLs: every
// module under island/app. main.js's static import closure covers the
// whole set (pinned by TestIslandModulesMatchImportClosure — no dynamic
// imports), so preloading the list warms exactly the graph the browser
// would otherwise discover one waterfall step at a time. Computed once:
// the embed is immutable.
var islandModuleURLs = func() []string {
	matches, err := fs.Glob(embedded, "island/app/*.js")
	if err != nil {
		panic("assets: island module glob: " + err.Error())
	}
	sort.Strings(matches)
	urls := make([]string, len(matches))
	for i, m := range matches {
		urls[i] = "/assets/" + m
	}
	return urls
}()

// IslandModules returns the sorted serving URLs of the island's ES
// modules; the page head renders one modulepreload link per entry so
// the graph is fetched and compiled while the body still parses (the
// entry script tag sits at the end of body).
func IslandModules() []string {
	return islandModuleURLs
}

// FS returns the embedded assets rooted at the package directory, so paths
// are "island/...", "vendor/sip.min.js", "app.css".
func FS() fs.FS {
	return embedded
}
