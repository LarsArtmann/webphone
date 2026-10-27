package views

import (
	"os"
	"testing"
)

func TestDumpDispositionSelect(t *testing.T) {
	out := renderComponent(t, HistoryPanel(HistoryPanelProps{Enabled: true, Lang: LangEN, Disposition: "missed"}))
	os.WriteFile("/tmp/history-render.html", []byte(out), 0o644)
}
