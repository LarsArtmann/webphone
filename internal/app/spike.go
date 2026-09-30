package app

import (
	"net/http"

	dashboard "github.com/larsartmann/go-health-dashboard"
)

// Spike: footprint measurement only.
var _ = dashboard.New

func spikeHandler() http.Handler { return nil }
