//go:build threadsdual

// T20 dual-run spike (2026-10-15 remainder plan M15/μ57): run the SAME
// thread fixtures through the production SQL list AND an equivalent
// metaengine projection ("planned tables, read side only"), byte-compare
// the JSON, and surface the mechanics that do NOT port for free. Build
// tag-gated so production and CI never pay for it:
//
//	go test -tags threadsdual -run TestDualRunThreads ./internal/store/
//
// The report this produces lives in
// docs/planning/2026-10-15_20-35_T20-dual-run-report.md.
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	sqliteengine "github.com/larsartmann/go-cqrs-lite/metaengine/sqliteengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"

	"github.com/larsartmann/webphone/internal/domain"
)

// dualThreadUpserted sets one thread's row state (flags, unread,
// last_activity — the read-side projection of the SQL threads table).
type dualThreadUpserted struct {
	Key      domain.ThreadID
	Owner    string
	Remote   string
	Unread   int
	Pinned   bool
	Archived bool
	Muted    bool
	LastUnix int64
}

func (dualThreadUpserted) EventName() string { return "dual.thread_upserted" }

// dualMessageAppended carries one message row; the fold keeps the
// last-by-(created_at, seq) as the thread's last message and bumps
// last_activity, mirroring the SQL correlated subquery
// (ORDER BY created_at DESC, rowid DESC LIMIT 1) and the threads row's
// last_activity_at column.
type dualMessageAppended struct {
	Key         domain.ThreadID
	Body        string
	Direction   string
	Channel     string
	HasAttach   bool
	CreatedUnix int64
	Seq         int64
}

func (dualMessageAppended) EventName() string { return "dual.message_appended" }

// dualThreadView is the engine-side thread row.
type dualThreadView struct {
	ID            string `json:"id"`
	Owner         string `json:"owner"`
	Remote        string `json:"remote"`
	LastUnix      int64  `json:"last_activity_at"`
	Unread        int    `json:"unread"`
	Pinned        bool   `json:"pinned"`
	Archived      bool   `json:"archived"`
	Muted         bool   `json:"muted"`
	LastBody      string `json:"last_body"`
	LastDirection string `json:"last_direction"`
	LastChannel   string `json:"last_channel"`
	LastHasAttach bool   `json:"last_has_attach"`
	LastSet       bool   `json:"last_set"`
	LastCreated   int64  `json:"last_created"`
	LastSeq       int64  `json:"last_seq"`
}

type dualInput struct{ Key domain.ThreadID }

var dualQuery = metaengine.Query[dualInput, dualThreadView]("dual_threads",
	metaengine.OnRecord(dualThreadUpserted{},
		func(_ record.Record, e dualThreadUpserted) (domain.ThreadID, dualThreadView) {
			return e.Key, dualThreadView{
				ID: e.Key.String(), Owner: e.Owner, Remote: e.Remote,
				LastUnix: e.LastUnix, Unread: e.Unread,
				Pinned: e.Pinned, Archived: e.Archived, Muted: e.Muted,
			}
		}),
	metaengine.OnRecord(dualMessageAppended{},
		func(_ record.Record, e dualMessageAppended, prev dualThreadView) dualThreadView {
			view := prev
			view.LastUnix = e.CreatedUnix
			if !prev.LastSet || e.CreatedUnix > prev.LastCreated ||
				(e.CreatedUnix == prev.LastCreated && e.Seq > prev.LastSeq) {
				view.LastBody = e.Body
				view.LastDirection = e.Direction
				view.LastChannel = e.Channel
				view.LastHasAttach = e.HasAttach
				view.LastSet = true
				view.LastCreated = e.CreatedUnix
				view.LastSeq = e.Seq
			}
			return view
		}),
)

// TestDualRunThreads is the spike: seed both stores identically, compare
// byte-for-byte, and print the per-case verdicts (the test PASSES when the
// comparison completes; the REPORT carries the mismatches — this is
// evidence-gathering for a NO-GO, not a parity gate).
func TestDualRunThreads(t *testing.T) {
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	messages := newTestDB(t)

	eng, _, err := sqliteengine.NewFromDSN(filepath.Join(t.TempDir(), "dual.db"))
	if err != nil {
		t.Fatal(err)
	}
	models, err := metaengine.Plan([]metaengine.Engine{eng}, dualQuery)
	if err != nil {
		t.Fatal(err)
	}

	base := time.Unix(1_700_000_000, 0)
	type fixture struct {
		name     string
		remote   string
		pinned   bool
		archived bool
		muted    bool
		unread   int
		msgs     []domain.Message
		id       domain.ThreadID
	}
	// unread mirrors what the STORE computes (domain.AppendToThread
	// increments on inbound) — the engine side gets the same number.
	fixtures := []fixture{
		{
			name: "plain-two-msgs", remote: "+441632960001", unread: 1, muted: true,
			msgs: []domain.Message{
				{Direction: domain.DirectionOutbound, Body: "first outbound", CreatedAt: base},
				{Direction: domain.DirectionInbound, Body: "reply 🎉", CreatedAt: base.Add(time.Minute)},
			},
		},
		{
			name: "pinned-late", remote: "+441632960002", pinned: true, unread: 1,
			msgs: []domain.Message{
				{Direction: domain.DirectionInbound, Body: "pinned thread body", CreatedAt: base.Add(time.Hour)},
			},
		},
		{
			name: "archived-hidden", remote: "+441632960003", archived: true,
			msgs: []domain.Message{
				{Direction: domain.DirectionOutbound, Body: "archived out", CreatedAt: base.Add(2 * time.Hour)},
			},
		},
		{
			name: "tie-broken-by-rowid", remote: "+441632960004",
			msgs: []domain.Message{
				{Direction: domain.DirectionInbound, Body: "older same-second", CreatedAt: base},
				{Direction: domain.DirectionOutbound, Body: "newer same-second", CreatedAt: base},
			},
		},
		{
			name: "attachment-last", remote: "+441632960005", unread: 1,
			msgs: []domain.Message{
				{Direction: domain.DirectionOutbound, Body: "text then file", CreatedAt: base},
				{Direction: domain.DirectionInbound, Body: "", CreatedAt: base.Add(time.Minute),
					Attachments: []domain.Attachment{{ID: domain.GenerateAttachmentID(), Name: "a.pdf", MimeType: "application/pdf", SizeBytes: 3}}},
			},
		},
		{
			name: "search-target-50%", remote: "+441632960006",
			msgs: []domain.Message{
				{Direction: domain.DirectionOutbound, Body: "invoice for 50% ready", CreatedAt: base.Add(3 * time.Hour)},
			},
		},
	}

	var seq int64
	for i := range fixtures {
		fx := &fixtures[i]
		remote := domain.MustParsePhone(fx.remote)
		threadID, err := messages.FindThread(ctx, owner, remote, fx.msgs[0].CreatedAt)
		if err != nil {
			t.Fatal(err)
		}
		fx.id = threadID
		if err := models.Apply(ctx, metaengine.EventTypeName(dualThreadUpserted{}), dualThreadUpserted{
			Key: threadID, Owner: owner.String(), Remote: fx.remote,
			Unread: fx.unread, Pinned: fx.pinned, Archived: fx.archived, Muted: fx.muted,
			LastUnix: fx.msgs[0].CreatedAt.Unix(),
		}); err != nil {
			t.Fatal(err)
		}
		for _, m := range fx.msgs {
			m.ID = domain.GenerateMessageID()
			m.ThreadID = threadID
			m.Owner = owner
			m.Remote = remote
			m.Channel = domain.ChannelSMS
			m.Status = domain.StatusSent
			seq++
			if err := messages.AppendMessage(ctx, m); err != nil {
				t.Fatal(err)
			}
			if err := models.Apply(ctx, metaengine.EventTypeName(dualMessageAppended{}), dualMessageAppended{
				Key: threadID, Body: m.Body, Direction: string(m.Direction), Channel: string(m.Channel),
				HasAttach: len(m.Attachments) > 0, CreatedUnix: m.CreatedAt.Unix(),
				Seq: seq,
			}); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Mirror the flags the SQL side carries only via SetThreadFlag
	// (archived/pinned/muted are read-side state).
	flag := func(i int, f ThreadFlag, on bool) {
		if err := messages.SetThreadFlag(ctx, owner, fixtures[i].id, f, on); err != nil {
			t.Fatal(err)
		}
	}
	flag(1, FlagPinned, true)
	flag(2, FlagArchived, true)
	flag(0, FlagMuted, true)

	report := &strings.Builder{}
	report.WriteString("# T20 dual-run report (threads list, read side only)\n\n")
	report.WriteString("Generated by `go test -tags threadsdual -run TestDualRunThreads ./internal/store/` (2026-10-15 spike).\n\n")
	report.WriteString("Both sides see the same rows; the engine side is a fold projection fed by two facts. JSON byte-compare of the `[]ThreadSummary` values.\n\n")

	compare := func(label string, sqlRows []ThreadSummary, engineRows []ThreadSummary) {
		sqlJSON, _ := json.Marshal(sqlRows)
		engJSON, _ := json.Marshal(engineRows)
		verdict := "PASS (byte-identical)"
		if len(sqlRows) == 0 && len(engineRows) == 0 {
			verdict = "PASS (both empty)"
		}
		if string(sqlJSON) != string(engJSON) {
			verdict = "MISMATCH"
		}
		fmt.Fprintf(report, "## %s — %s\n\n", label, verdict)
		if verdict == "MISMATCH" {
			fmt.Fprintf(report, "```json\n// sql:    %s\n// engine: %s\n```\n\n", sqlJSON, engJSON)
		}
		t.Logf("%s: %s", label, verdict)
	}

	sqlList, err := messages.ListThreads(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	compare("ListThreads", sqlList, engineSummaries(t, models, false))

	sqlArch, err := messages.ListArchivedThreads(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	compare("ListArchivedThreads", sqlArch, engineSummaries(t, models, true))

	for _, needle := range []string{"50%", "reply", "PINNED", "", "no-such-body"} {
		sqlSearch, err := messages.SearchThreads(ctx, owner, needle)
		if err != nil {
			t.Fatal(err)
		}
		compare(fmt.Sprintf("SearchThreads(%q)", needle), sqlSearch, engineSearch(t, models, needle))
	}

	out := "../../docs/planning/2026-10-15_20-35_T20-dual-run-report.md"
	if err := os.WriteFile(out, []byte(report.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("report written to %s", out)
}

// engineSummaries materializes the engine view into ThreadSummary rows in
// SQL's order (pinned DESC, last_activity DESC) — the in-Go re-sort a real
// read model needs because the sqlite pushdown applies only the first
// compound-sort column (WithSortColumns godoc, metaengine v4.19.0).
func engineSummaries(t *testing.T, models *metaengine.Store, archived bool) []ThreadSummary {
	t.Helper()
	views := scanAll(t, models)
	var out []ThreadSummary
	for _, v := range views {
		if v.Archived != archived {
			continue
		}
		out = append(out, engineToSummary(v))
	}
	sortDual(out)
	return out
}

// engineSearch mirrors SearchThreads' semantics in Go: ASCII
// case-insensitive contains on remote or last body, archived excluded,
// SQL order. This is the port cost the SQL query hides (LIKE + ESCAPE +
// correlated EXISTS).
func engineSearch(t *testing.T, models *metaengine.Store, needle string) []ThreadSummary {
	t.Helper()
	pattern := "%" + likeEscape(needle) + "%"
	var out []ThreadSummary
	for _, v := range scanAll(t, models) {
		if v.Archived {
			continue
		}
		if !likeMatch(pattern, v.Remote) && !likeMatch(pattern, v.LastBody) {
			continue
		}
		out = append(out, engineToSummary(v))
	}
	sortDual(out)
	return out
}

func scanAll(t *testing.T, models *metaengine.Store) []dualThreadView {
	t.Helper()
	views, err := metaengine.NewReader[dualThreadView](models, "dual_threads").Scan(context.Background(), metaengine.WithLimit(0))
	if err != nil {
		t.Fatal(err)
	}
	return views
}

func sortDual(out []ThreadSummary) {
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Thread.Pinned != out[j].Thread.Pinned {
			return out[i].Thread.Pinned
		}
		return out[i].Thread.LastActivityAt.After(out[j].Thread.LastActivityAt)
	})
}

// likeMatch evaluates the escaped LIKE pattern (always wrapped in %...% by
// SearchThreads) as ordered contains-segments, ASCII case-insensitive —
// the same semantics SQLite's LIKE gives the SQL side.
func likeMatch(pattern, value string) bool {
	parts := strings.Split(pattern, "%")
	if len(parts) < 2 {
		return false
	}
	v := strings.ToLower(value)
	pos := 0
	for _, part := range parts[1 : len(parts)-1] {
		if part == "" {
			continue
		}
		part = strings.ToLower(strings.NewReplacer(`\\`, `\`, `\%`, `%`, `\_`, `_`).Replace(part))
		idx := strings.Index(v[pos:], part)
		if idx < 0 {
			return false
		}
		pos += idx + len(part)
	}
	return true
}

func engineToSummary(v dualThreadView) ThreadSummary {
	sum := ThreadSummary{
		Thread: domain.Thread{
			ID:             domain.MustThreadID(v.ID),
			Owner:          domain.MustParseExtension(v.Owner),
			Remote:         domain.MustParsePhone(v.Remote),
			LastActivityAt: time.Unix(v.LastUnix, 0),
			Unread:         v.Unread,
			Pinned:         v.Pinned,
			Archived:       v.Archived,
			Muted:          v.Muted,
		},
		LastAttachment: v.LastHasAttach,
	}
	if v.LastSet {
		sum.LastBody = v.LastBody
		sum.LastDirection = domain.Direction(v.LastDirection)
		sum.LastChannel = domain.Channel(v.LastChannel)
	}
	return sum
}
