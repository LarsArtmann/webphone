package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

// bootReportMarkers are the five contract parts; the smoke suite
// greps the same five on a live boot failure.
var bootReportMarkers = []string{"WHAT:", "REASSURE:", "WHY:", "FIX:", "ESCAPE:"}

// TestRenderBootFailureCoversEveryClass is the enum-coverage guard: a
// new bootClass value with missing copy renders a zero row and fails
// here (the copyRows array bound makes forgetting the copy a
// compile-time size error only when bootClassCount is updated; this
// test catches the partial case). Every class must render the header
// (class + phase + version), the cause line, and all five parts
// non-empty.
func TestRenderBootFailureCoversEveryClass(t *testing.T) {
	for i := 0; i < int(bootClassCount); i++ {
		class := bootClass(i)
		if copyRows[class].name == "" {
			t.Errorf("bootClass(%d) has no copy row: fill copyRows in bootreport.go", i)
			continue
		}
		out := renderBootFailure(bootFailure{class: class, cause: "synthetic cause"}, "v9.9.9-test")
		if !strings.Contains(out, "webphone boot failed (class="+class.String()+", phase=") {
			t.Errorf("class %s: header missing class or phase\n%s", class, out)
		}
		if !strings.Contains(out, "version=v9.9.9-test") {
			t.Errorf("class %s: header missing the version stamp\n%s", class, out)
		}
		if !strings.HasPrefix(out, "webphone boot failed ") || !strings.Contains(out, "\nerror: synthetic cause\n") {
			t.Errorf("class %s: cause line malformed\n%s", class, out)
		}
		for _, marker := range bootReportMarkers {
			idx := strings.Index(out, marker)
			if idx < 0 {
				t.Errorf("class %s: contract part %s missing\n%s", class, marker, out)
				continue
			}
			body := out[idx+len(marker):]
			if nl := strings.IndexByte(body, '\n'); nl >= 0 {
				body = body[:nl]
			}
			if strings.TrimSpace(body) == "" {
				t.Errorf("class %s: contract part %s is empty\n%s", class, marker, out)
			}
		}
	}
}

// TestRenderBootFailureGoldenListen byte-pins one full render (the
// listen class: the most common real operator failure). Deliberate
// copy changes update this literal and copyRows in the same commit.
func TestRenderBootFailureGoldenListen(t *testing.T) {
	got := renderBootFailure(bootFailure{
		class: bootListen,
		cause: "listen tcp 127.0.0.1:8080: bind: address already in use",
	}, "v9.9.9-test")
	want := `webphone boot failed (class=listen, phase=listen, version=v9.9.9-test)
error: listen tcp 127.0.0.1:8080: bind: address already in use

WHAT:     webphone could not bind or serve its HTTP address, so nothing is listening.
REASSURE: No requests were accepted and nothing was written; the port conflict harms no data.
WHY:      The configured WEBPHONE_ADDR is already held by another process (often a second or leftover webphone) or is not bindable on this host.
FIX:      Find the holder with ` + "`sudo ss -ltnp 'sport = :<port>'`" + `, then stop that service or change WEBPHONE_ADDR in the unit's environment file to a free address (e.g. 127.0.0.1:8080).
ESCAPE:   Exit code 1; systemd retries every 5s. ` + "`systemctl stop webphone`" + ` stops the loop while you free the port.
`
	if got != want {
		t.Errorf("golden render drifted:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestBootCopyIsEnglishOnly pins ruling D2 for the boot journal: the
// copy carries no German script and no leftover template braces.
func TestBootCopyIsEnglishOnly(t *testing.T) {
	for i := 0; i < int(bootClassCount); i++ {
		row := copyRows[bootClass(i)]
		for _, field := range []string{row.what, row.reassure, row.why, row.fix, row.escape} {
			if strings.ContainsAny(field, "äöüßÄÖÜ") {
				t.Errorf("class %s: boot copy must be English-only (D2): %q", row.name, field)
			}
			if strings.Contains(field, "{{") {
				t.Errorf("class %s: unrendered template braces: %q", row.name, field)
			}
		}
	}
}

// TestClassifyBootError pins the designed-error tagging and the
// app.New wrap-prefix seam: a rename of New's wrap literals must flip
// this test (and TestAppWrapPrefixesStillExistInAppSource below).
func TestClassifyBootError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bootClass
	}{
		{"plain error falls back generic", errors.New("boom"), bootGeneric},
		{"config site tagged", bootErr(bootConfigInvalid, errors.New("load config: bad addr")), bootConfigInvalid},
		{"start site tagged", bootErr(bootGeneric, errors.New("start app: probe refused")), bootGeneric},
		{"data dir prefix", errors.New("create data dir: mkdir /var/lib/webphone: permission denied"), bootDataDir},
		{"timezone prefix", errors.New("load timezone: unknown time zone Nowhere/Nowhere"), bootTimezone},
		{"paperless prefix", errors.New("paperless: both-or-neither violated"), bootPaperless},
		{"unclassified New wrap stays generic", errors.New("wire server: store exploded"), bootGeneric},
		{"tagged inside a join wins", errors.Join(bootErr(bootListen, errors.New("serve: bind: busy")), errors.New("shutdown hiccup")), bootListen},
		{"wrapped tagged error unwraps", bootErr(bootConfigInvalid, errors.New("load config: x")), bootConfigInvalid},
	}
	for _, tc := range cases {
		if got := classifyBootError(tc.err); got != tc.want {
			t.Errorf("%s: classifyBootError = %s, want %s", tc.name, got, tc.want)
		}
	}
}

// TestBootExitCodeTaxonomy pins D5: 1 = designed failure, 2 = panic
// (Go's historic default, now a deliberate contract).
func TestBootExitCodeTaxonomy(t *testing.T) {
	if exitDesignedBootFailure != 1 {
		t.Errorf("exitDesignedBootFailure = %d, want 1 (docs/error-contract.md Boot surface)", exitDesignedBootFailure)
	}
	if exitPanicBootFailure != 2 {
		t.Errorf("exitPanicBootFailure = %d, want 2 (Go's default panic code, now deliberate)", exitPanicBootFailure)
	}
}

// swapBootSink redirects the report shells' writer and exit capture for
// one test and restores them on cleanup.
func swapBootSink(t *testing.T) (*bytes.Buffer, *[]int) {
	t.Helper()
	buf := &bytes.Buffer{}
	exits := &[]int{}
	oldOut, oldExit := bootOut, bootExit
	bootOut = buf
	bootExit = func(code int) { *exits = append(*exits, code) }
	t.Cleanup(func() { bootOut, bootExit = oldOut, oldExit })
	return buf, exits
}

// TestReportBootFailureRendersAndExitsOne covers the designed path
// end to end: one journal-worthy render, exactly one exit, code 1.
func TestReportBootFailureRendersAndExitsOne(t *testing.T) {
	buf, exits := swapBootSink(t)
	reportBootFailure(errors.New("create data dir: permission denied"))
	if got := (*exits); len(got) != 1 || got[0] != exitDesignedBootFailure {
		t.Fatalf("exits = %v, want exactly [%d]", got, exitDesignedBootFailure)
	}
	out := buf.String()
	for _, marker := range bootReportMarkers {
		if !strings.Contains(out, marker) {
			t.Errorf("render missing %s:\n%s", marker, out)
		}
	}
	if !strings.Contains(out, "class=data-dir") {
		t.Errorf("prefix classification failed, header:\n%s", out)
	}
}

// TestReportBootPanicRendersBlockTraceAndExitsTwo covers the recover
// arm: contract block first, trace marker, raw stack below, exit 2.
func TestReportBootPanicRendersBlockTraceAndExitsTwo(t *testing.T) {
	buf, exits := swapBootSink(t)
	reportBootPanic("do: service not found: sqlite")
	if got := (*exits); len(got) != 1 || got[0] != exitPanicBootFailure {
		t.Fatalf("exits = %v, want exactly [%d]", got, exitPanicBootFailure)
	}
	out := buf.String()
	for _, marker := range bootReportMarkers {
		if !strings.Contains(out, marker) {
			t.Errorf("render missing %s:\n%s", marker, out)
		}
	}
	for _, want := range []string{"class=panic", "do: service not found: sqlite", traceMarker, "goroutine"} {
		if !strings.Contains(out, want) {
			t.Errorf("panic report missing %q:\n%s", want, out)
		}
	}
	traceAt := strings.Index(out, traceMarker)
	blockAt := strings.LastIndex(out, "ESCAPE:")
	if traceAt < 0 || blockAt < 0 || traceAt < blockAt {
		t.Errorf("trace must render BELOW the contract block:\n%s", out)
	}
}

// TestRecoverBootPanicOnlyFiresOnPanic proves the deferred wrapper is
// inert on the normal path (no render, no exit) and routes a real
// panic to the reporter.
func TestRecoverBootPanicOnlyFiresOnPanic(t *testing.T) {
	_, exits := swapBootSink(t)
	recoverBootPanic()
	if len(*exits) != 0 {
		t.Fatalf("inert path exited: %v", *exits)
	}
	func() {
		defer recoverBootPanic()
		panic("synthetic boot panic")
	}()
	if got := *exits; len(got) != 1 || got[0] != exitPanicBootFailure {
		t.Fatalf("panic path exits = %v, want [%d]", got, exitPanicBootFailure)
	}
}

// TestAppWrapPrefixesStillExistInAppSource is the drift pin for the
// classification seam: app.New's wrap literals are the ONE home; if
// they move or rename, this test forces the appWrap* constants (and
// the copy that depends on them) to follow.
func TestAppWrapPrefixesStillExistInAppSource(t *testing.T) {
	src, err := os.ReadFile("../../internal/app/app.go")
	if err != nil {
		t.Skipf("app.go not readable here (exported tree?): %v", err)
	}
	for _, p := range []struct{ class, prefix string }{
		{"data-dir", appWrapDataDir},
		{"timezone", appWrapTimezone},
		{"paperless", appWrapPaperless},
	} {
		call := `wrapf("` + p.prefix
		if !strings.Contains(string(src), call) {
			t.Errorf("classification prefix %q (class %s) no longer matches a wrapf call in internal/app/app.go: update appWrap* + copy in bootreport.go or fix the classification", p.prefix, p.class)
		}
	}
}
