package server

import (
	"strconv"
	"testing"

	"github.com/larsartmann/go-sse"

	"github.com/larsartmann/webphone/internal/domain"
)

// BenchmarkHubFanOut measures the broadcast cost the notifier adds per
// tab event: N extensions each with M live tab connections. Baseline
// numbers (2026-09-18): docs/reviews/2026-09-18_hub-fanout-baseline.md
func BenchmarkHubFanOut(b *testing.B) {
	for _, bb := range []struct {
		hubs, subscribers int
	}{
		{hubs: 1, subscribers: 1},
		{hubs: 10, subscribers: 2},
		{hubs: 100, subscribers: 2},
		{hubs: 100, subscribers: 10},
	} {
		b.Run("hubs"+strconv.Itoa(bb.hubs)+"xsubs"+strconv.Itoa(bb.subscribers), func(b *testing.B) {
			hubs := NewHubs()
			for i := 0; i < bb.hubs; i++ {
				ext, _ := domain.ParseExtension(strconv.Itoa(1000 + i))
				hub := hubs.get(ext)
				for j := 0; j < bb.subscribers; j++ {
					ch := hub.Hub().Subscribe()
					b.Cleanup(func() { hub.Hub().Unsubscribe(ch) })
				}
			}
			ext, _ := domain.ParseExtension("1000")
			hub := hubs.get(ext)
			event := sse.Event{Event: "threads", Data: "<div>bench</div>"}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hub.Broadcast(event)
			}
		})
	}
}
