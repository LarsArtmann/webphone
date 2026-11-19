package cqrs

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The call-notes/disposition slot is a state machine (mint → echo-drop →
// correction → merge → delete frees the cap), so per the repo's BDD
// posture it earns the Ginkgo lane beside the mechanics-level table
// tests in callnotes_test.go. The specs assert ONLY the exported Notes
// gate (Set/SetDisposition/Get/Count/Delete) — the boot helpers are
// shared with the table tests, the behaviors are black-box.
func TestCallNotesSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	notesSuiteT = t
	RunSpecs(t, "webphone call-notes slot suite")
}

// notesSuiteT hands the real *testing.T to the shared table-test boot
// helpers (bootNotes/awaitNote/mustExt) — GinkgoT() is an interface the
// helpers' signatures predate.
var notesSuiteT *testing.T
