package store

import (
	"strings"
	"testing"
)

// fakeRow feeds scanMessage a fully-formed column set without a
// database — the scan's own contract is under test, not SQLite's.
type fakeRow []any

func (r fakeRow) Scan(dest ...any) error {
	for i, src := range r {
		switch d := dest[i].(type) {
		case *string:
			*d = src.(string)
		case *int64:
			*d = src.(int64)
		default:
			panic("fakeRow: unsupported dest type")
		}
	}
	return nil
}

// wellFormedMessageRow mirrors one stored outbound SMS row.
func wellFormedMessageRow() []any {
	return []any{
		"msg_AAAAAAAAEEEEEEEEEEEEEEEE", "thr_AAAAAAAAEEEEEEEEEEEEEEEE", "1001", "+441632960961",
		"out", "sms", "hello", "sent", "ref-1", "", "", int64(1760000000),
	}
}

// TestScanMessageRejectsCorruptEnumRows pins the corrupt-row contract:
// the enum columns ride the Must* parsers, so a row whose enum value
// the writer never writes PANICS with the seam code instead of
// rendering a lying state to a user (the 2026-10-17 data-model batch,
// F37).
func TestScanMessageRejectsCorruptEnumRows(t *testing.T) {
	corrupt := map[string]int{"direction": 4, "channel": 5, "status": 7}
	for column, idx := range corrupt {
		t.Run(column, func(t *testing.T) {
			row := wellFormedMessageRow()
			row[idx] = "??corrupt??"
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("a corrupt %s column must panic the scan, not pass through", column)
				}
				if msg, ok := r.(error); !ok || !strings.Contains(msg.Error(), "domain.") {
					t.Fatalf("panic must name the domain seam, got %v", r)
				}
			}()
			_, _ = scanMessage(fakeRow(row))
		})
	}
}

// TestScanMessageAcceptsEveryStoredEnumValue pins that the honest
// vocabulary still scans: inbound rows carry the empty status, every
// channel/direction/status the writer emits parses.
func TestScanMessageAcceptsEveryStoredEnumValue(t *testing.T) {
	for _, direction := range []string{"in", "out"} {
		for _, status := range []string{"", "queued", "sent", "delivered", "failed"} {
			row := wellFormedMessageRow()
			row[4] = direction
			row[7] = status
			msg, err := scanMessage(fakeRow(row))
			if err != nil {
				t.Fatalf("scan %s/%s: %v", direction, status, err)
			}
			if string(msg.Direction) != direction || string(msg.Status) != status {
				t.Fatalf("roundtrip %s/%s got %s/%s", direction, status, msg.Direction, msg.Status)
			}
		}
	}
}
