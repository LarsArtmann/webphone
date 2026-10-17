package gateway

import (
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
)

// The dispatch is exhaustive on purpose: an unknown mode reaching a
// constructor is a wiring bug (config.validate is the rejection home),
// and the panic must NAME the mode so the miswire is greppable in a
// boot report instead of degrading to a pretend loopback gateway.
func TestNewGatewayPanicsOnUnknownMode(t *testing.T) {
	cfg := config.Gateway{Mode: config.GatewayMode("carrier-pigeon")}

	assertPanicNamesMode := func(name string, construct func()) {
		t.Helper()
		defer func() {
			r := recover()
			if r == nil {
				t.Fatalf("%s must panic on an unknown mode, not fall back to loopback", name)
			}
			msg, ok := r.(string)
			if !ok || !strings.Contains(msg, "carrier-pigeon") || !strings.Contains(msg, "loopback|webhook") {
				t.Errorf("%s panic = %v, want a message naming the mode and the legal set", name, r)
			}
		}()
		construct()
	}

	assertPanicNamesMode("NewMessageGateway", func() { NewMessageGateway(cfg, nil) })
	assertPanicNamesMode("NewFaxGateway", func() { NewFaxGateway(cfg, nil) })
}
