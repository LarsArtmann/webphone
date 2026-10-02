// Boot-failure reporting: the operator surface for every way the boot
// can die. The operator is a user (ruling 2026-10-02), so each failure
// renders the five-part error contract (what / reassure / why / fix /
// escape) plus the underlying error, version-stamped and phase-tagged,
// English-only (D3 precedent: the journal is the operator trail).
//
// The renderer is pure (D4); the thin report/recover shells own the
// side effects (stderr, slog, exit). The panic MECHANISM at the
// composition root stays fail-fast by design — this file only changes
// its PRESENTATION. Contract: docs/error-contract.md, "Boot surface".
package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"github.com/larsartmann/webphone/internal/server"
)

// Boot exit codes (taxonomy D5, docs/error-contract.md "Boot surface"):
//   1 designed boot failure (config, storage, timezone, paperless,
//     listen, or an unclassified step)
//   2 a panic escaped run() (Go's default panic code, now deliberate)
//
// systemd restarts either way (Restart=on-failure, RetrySec 5) — the
// module ships that policy; each attempt renders exactly once.
const (
	exitDesignedBootFailure = 1
	exitPanicBootFailure    = 2
)

// bootClass enumerates every failure class the boot path can produce.
// The copy rows below are the single home for the per-class wording;
// bootClassCount must stay last so the copyRows array forces a new
// class to carry copy (missing copy = zero row = test failure).
type bootClass int

const (
	bootPanicMiswire bootClass = iota
	bootConfigInvalid
	bootDataDir
	bootTimezone
	bootPaperless
	bootListen
	bootGeneric
	bootClassCount
)

// app.New's wrap prefixes (internal/app/app.go) are the classification
// seam for New's sub-phases: config.Load already validates timezone and
// paperless, so these wraps are second-line defense, but they must
// classify correctly when they DO fire. bootreport_test.go pins the
// literals against app.go so a rename cannot silently degrade them to
// the generic class.
const (
	appWrapDataDir   = "create data dir:"
	appWrapTimezone  = "load timezone:"
	appWrapPaperless = "paperless:"
)

// bootCopy is one class's five-part contract copy (static, English).
type bootCopy struct {
	name     string
	phase    string
	what     string
	reassure string
	why      string
	fix      string
	escape   string
}

// copyRows is indexed by bootClass; the array bound enforces that a new
// enum value cannot compile without touching this table, and the
// enum-coverage test fails on a zero-value (copy-less) row.
var copyRows = [bootClassCount]bootCopy{
	bootPanicMiswire: {
		name:     "panic",
		phase:    "wire",
		what:     "webphone hit an internal error while wiring its services (a panic escaped the startup path) and stopped.",
		reassure: "Nothing was served and no data was written; stored messages, faxes, and contacts are untouched.",
		why:      "A service in the composition root panicked while being built or started. This is a webphone bug or a service miswire, not an operator configuration mistake.",
		fix:      "Re-run once to confirm it reproduces. If it does, attach this block and the trace below (secrets removed) to a bug report against this webphone version.",
		escape:   "Exit code 2. systemd restarts every 5s (Restart=on-failure); stop the loop with `systemctl stop webphone`, or `systemctl disable --now webphone` to keep it stopped across reboots.",
	},
	bootConfigInvalid: {
		name:     "config",
		phase:    "config",
		what:     "webphone rejected its configuration and did not start.",
		reassure: "Nothing was written: no database, no files, no listener. Repeating with a fixed config is safe.",
		why:      "The WEBPHONE_* environment and/or the WEBPHONE_CONFIG JSON file failed validation (unknown or missing keys, bad values, malformed nesting).",
		fix:      "Read the error line above. Env keys nest with a double underscore (WEBPHONE_GATEWAY__MODE); lists and maps live in the JSON file; README documents every key.",
		escape:   "Exit code 1. systemd retries every 5s until fixed; `systemctl stop webphone` pauses the loop while you edit the unit's environment file.",
	},
	bootDataDir: {
		name:     "data-dir",
		phase:    "storage",
		what:     "webphone could not create or open its data directory, so SQLite and the file store cannot start.",
		reassure: "Existing data is intact: a permission or path problem blocks access, it does not delete anything.",
		why:      "The configured data dir (WEBPHONE_DATA_DIR, or the systemd StateDirectory) is missing, unwritable, or sits under a non-directory path.",
		fix:      "Give the webphone user a writable directory: `sudo install -d -o webphone -g webphone -m 0750 /var/lib/webphone` (module deployments own this via StateDirectory; bare binaries must set WEBPHONE_DATA_DIR).",
		escape:   "Exit code 1; systemd retries every 5s. `systemctl stop webphone` stops the retry loop while you fix permissions.",
	},
	bootTimezone: {
		name:     "timezone",
		phase:    "timezone",
		what:     "webphone could not load the configured IANA time zone and refused to start rather than render wrong times.",
		reassure: "Config-only: nothing was written, stored timestamps are untouched.",
		why:      "WEBPHONE_TIMEZONE (config key `timezone`) is not a valid IANA zone on this host (a typo, or a slim system without tzdata).",
		fix:      "Set a valid zone, e.g. WEBPHONE_TIMEZONE=Europe/Berlin (`timedatectl list-timezones` lists them), or unset it to inherit the system zone.",
		escape:   "Exit code 1; systemd retries every 5s until the value is fixed.",
	},
	bootPaperless: {
		name:     "paperless",
		phase:    "paperless",
		what:     "webphone's Paperless-ngx archive configuration is unusable, so inbound fax archiving could not be set up.",
		reassure: "Optional integration: nothing was written, and faxes keep working without archiving.",
		why:      "paperless.url and paperless.token are BOTH required (both-or-neither): one is missing or empty, or the URL is not an absolute http(s) URL.",
		fix:      "Either set both WEBPHONE_PAPERLESS__URL and WEBPHONE_PAPERLESS__TOKEN, or remove both to run without archiving.",
		escape:   "Exit code 1; systemd retries every 5s until the pair is completed or removed.",
	},
	bootListen: {
		name:     "listen",
		phase:    "listen",
		what:     "webphone could not bind or serve its HTTP address, so nothing is listening.",
		reassure: "No requests were accepted and nothing was written; the port conflict harms no data.",
		why:      "The configured WEBPHONE_ADDR is already held by another process (often a second or leftover webphone) or is not bindable on this host.",
		fix:      "Find the holder with `sudo ss -ltnp 'sport = :<port>'`, then stop that service or change WEBPHONE_ADDR in the unit's environment file to a free address (e.g. 127.0.0.1:8080).",
		escape:   "Exit code 1; systemd retries every 5s. `systemctl stop webphone` stops the loop while you free the port.",
	},
	bootGeneric: {
		name:     "generic",
		phase:    "unclassified",
		what:     "webphone stopped with an error outside the classified boot phases (config, storage, timezone, paperless, listen).",
		reassure: "Diagnostic only: acting on this report cannot damage stored data.",
		why:      "An unexpected startup or shutdown step failed; the error line above carries the specifics.",
		fix:      "Read the surrounding journal context: `journalctl -u webphone -n 100 --no-pager`. Fix what the error names, then `systemctl start webphone`.",
		escape:   "Exit code 1; systemd retries every 5s while enabled; `systemctl stop webphone` stops the loop.",
	},
}

// String names a class for the journal attr and the report header.
func (c bootClass) String() string {
	if c >= 0 && int(c) < len(copyRows) && copyRows[c].name != "" {
		return copyRows[c].name
	}
	return fmt.Sprintf("bootClass(%d)", int(c))
}

// bootFailure is one failure handed to the renderer: the class and the
// dynamic cause text. The five-part copy itself is static data.
type bootFailure struct {
	class bootClass
	cause string
}

// renderBootFailure renders the whole operator block (pure; the report
// shells own stderr/slog/exit). One line per contract part keeps the
// journal greppable (the smoke suite asserts the five markers).
func renderBootFailure(f bootFailure, version string) string {
	row := copyRows[f.class]
	var b strings.Builder
	fmt.Fprintf(&b, "webphone boot failed (class=%s, phase=%s, version=%s)\n", f.class, row.phase, version)
	fmt.Fprintf(&b, "error: %s\n\n", f.cause)
	fmt.Fprintf(&b, "WHAT:     %s\n", row.what)
	fmt.Fprintf(&b, "REASSURE: %s\n", row.reassure)
	fmt.Fprintf(&b, "WHY:      %s\n", row.why)
	fmt.Fprintf(&b, "FIX:      %s\n", row.fix)
	fmt.Fprintf(&b, "ESCAPE:   %s\n", row.escape)
	return b.String()
}

// traceMarker sits between the contract block and the raw stack trace
// (panic path only): the report above is the whole operator contract;
// the trace is bug-report material printed after it for log tooling.
const traceMarker = "--- panic trace (attach this and the report above when filing a bug) ---"

// bootError tags a run() failure with its class so the report names the
// phase without string-matching the cause. It wraps transparently: the
// inner error keeps its family, text, and join behavior.
type bootError struct {
	class bootClass
	err   error
}

func (e *bootError) Error() string { return e.err.Error() }
func (e *bootError) Unwrap() error { return e.err }

// bootErr is the one tagging helper for run()'s designed-error sites.
func bootErr(class bootClass, err error) error {
	return &bootError{class: class, err: err}
}

// classifyBootError maps a run() error to its class: tagged sites win,
// then app.New's stable wrap prefixes, then the generic fallback.
func classifyBootError(err error) bootClass {
	var tagged *bootError
	if errors.As(err, &tagged) {
		return tagged.class
	}
	msg := err.Error()
	switch {
	case strings.HasPrefix(msg, appWrapDataDir):
		return bootDataDir
	case strings.HasPrefix(msg, appWrapTimezone):
		return bootTimezone
	case strings.HasPrefix(msg, appWrapPaperless):
		return bootPaperless
	default:
		return bootGeneric
	}
}

// The report shells' side effects are injectable so tests intercept the
// exit path instead of killing the test process (D4: pure renderer,
// thin impure edge).
var (
	bootOut     io.Writer = os.Stderr
	bootExit              = os.Exit
	bootVersion           = server.DisplayVersion
)

// reportBootFailure renders a designed run() error and exits 1. It
// never returns.
func reportBootFailure(err error) {
	class := classifyBootError(err)
	slog.Error("webphone boot failed", "class", class.String(), "error", err.Error())
	fmt.Fprintf(bootOut, "%s", renderBootFailure(bootFailure{class: class, cause: err.Error()}, bootVersion()))
	bootExit(exitDesignedBootFailure)
}

// recoverBootPanic is deferred first in run(): a panic escaping the
// boot (composition-root miswire, programming error) renders the
// contract block, then the raw trace below the marker, then exits 2.
func recoverBootPanic() {
	if r := recover(); r != nil {
		reportBootPanic(r)
	}
}

// reportBootPanic is the recover arm, split out so tests exercise it
// without a real panic.
func reportBootPanic(panicValue any) {
	cause := fmt.Sprint(panicValue)
	slog.Error("webphone boot panicked", "panic", cause)
	fmt.Fprintf(bootOut, "%s", renderBootFailure(bootFailure{class: bootPanicMiswire, cause: cause}, bootVersion()))
	fmt.Fprintf(bootOut, "\n%s\n%s", traceMarker, debug.Stack())
	bootExit(exitPanicBootFailure)
}
