package fax_test

import (
	"testing"

	"github.com/larsartmann/webphone/internal/readmodels"
)

// newTestFaxes opens a throwaway read-model engine and returns its fax
// store (the engine seam replaced the SQL store behind the same
// contract during the metaengine train).
func newTestFaxes(t *testing.T) *readmodels.Faxes {
	t.Helper()
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	return engine.Faxes()
}
