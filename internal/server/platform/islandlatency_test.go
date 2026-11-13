package platform

import (
	"strings"
	"testing"
)

func TestIslandLatencyRecordFoldsValidLegs(t *testing.T) {
	m := NewIslandLatency()
	m.RecordBeacon("dial_to_audio", map[string]int64{
		"island_overhead": 12,
		"signaling":       1500,
		"media_attach":    80,
		"audio_gap":       240,
	}, nil)
	var b strings.Builder
	m.writePrometheus(&b)
	out := b.String()
	for _, want := range []string{
		`webphone_island_latency_seconds_bucket{interval="dial_to_audio",leg="audio_gap",le="0.25"} 1`,
		`webphone_island_latency_seconds_bucket{interval="dial_to_audio",leg="island_overhead",le="0.025"} 1`,
		`webphone_island_latency_seconds_count{interval="dial_to_audio",leg="signaling"} 1`,
		`webphone_island_latency_seconds_sum{interval="dial_to_audio",leg="media_attach"} 0.08`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("exposition missing %q\ngot:\n%s", want, out)
		}
	}
	// Sorted (interval, leg) order: audio_gap renders before island_overhead.
	agap := strings.Index(out, `leg="audio_gap"`)
	oh := strings.Index(out, `leg="island_overhead"`)
	if agap < 0 || oh < 0 || agap > oh {
		t.Errorf("series not in sorted (interval, leg) order (audio_gap@%d island_overhead@%d)", agap, oh)
	}
}

func TestIslandLatencyDropsMalformedBeaconsSilently(t *testing.T) {
	m := NewIslandLatency()
	m.RecordBeacon("totally_unknown", map[string]int64{"leg": 5}, nil)
	m.RecordBeacon("dial_to_audio", map[string]int64{"rogue_leg": 5}, nil)
	m.RecordBeacon("asr_round_trip", map[string]int64{"capture_to_post": -1}, nil)
	m.RecordBeacon("asr_round_trip", map[string]int64{"capture_to_post": islandMaxLegMS + 1}, nil)
	m.RecordBeacon("dial_to_audio", map[string]int64{}, nil)
	var b strings.Builder
	m.writePrometheus(&b)
	if out := b.String(); out != "" {
		t.Errorf("malformed beacons must render nothing, got:\n%s", out)
	}
}

func TestIslandLatencyNegativeBytesDropButLegSurvives(t *testing.T) {
	m := NewIslandLatency()
	m.RecordBeacon("dial_to_audio", map[string]int64{"signaling": 10}, new(int64(-3)))
	var b strings.Builder
	m.writePrometheus(&b)
	out := b.String()
	if !strings.Contains(out, `webphone_island_latency_seconds_count{interval="dial_to_audio",leg="signaling"} 1`) {
		t.Errorf("valid leg must survive a negative size:\n%s", out)
	}
	if strings.Contains(out, "webphone_island_asr_segment_bytes") {
		t.Errorf("a negative size must not create the bytes family:\n%s", out)
	}
}

func TestIslandLatencyEmptyRendersNoFamily(t *testing.T) {
	m := NewIslandLatency()
	var b strings.Builder
	m.writePrometheus(&b)
	if out := b.String(); out != "" {
		t.Errorf("an empty recorder must publish no family, got:\n%s", out)
	}
}

func TestIslandLatencyASRBytesHistogram(t *testing.T) {
	m := NewIslandLatency()
	size := int64(4000)
	m.RecordBeacon("asr_round_trip", map[string]int64{"parse_to_insert": 3}, &size)
	var b strings.Builder
	m.writePrometheus(&b)
	out := b.String()
	if !strings.Contains(out, "webphone_island_asr_segment_bytes_bucket{le=\"4096\"} 1") {
		t.Errorf("segment-size histogram missing the 4 KiB bucket:\n%s", out)
	}
	if strings.Contains(out, `interval="dial_to_audio"`) {
		t.Errorf("ASR beacon must not create dial series:\n%s", out)
	}
}
