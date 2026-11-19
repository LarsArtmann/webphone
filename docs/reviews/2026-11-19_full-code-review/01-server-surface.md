# Full code review, part 1 of 4 — HTTP/server surface — 2026-11-19

> SNAPSHOT-ROT HEADER: claims verified as of **2026-11-19 07:59 UTC** against tree `54a93e86`.
> Line numbers and grep counts drift; re-verify before citing downstream.
> READ-ONLY REVIEW — no fixes applied (fixes are adjudicated centrally afterward).

Scope (PART 1 of 4): `internal/server/**` (root, `hooks/`, `api/`, `auth/`, `platform/`),
`internal/web/views/**` (`.templ` sources + `helpers.go` + `i18n.go` + `types.go` + `crm.go`),
`cmd/webphone/**`. Generated `*_templ.go` files skipped beyond spot-checks (findings target
`.templ` sources); `internal/web/assets/**` owned by another reviewer. 80 non-test code files
read in full, ~78 test files visited (pin enumeration + targeted reads of the suites guarding
each finding; 362 unique tests exist under `internal/server` alone).

Cross-referenced, not re-litigated: the 2026-11-19 naming review (Shell×2, `CDROff`,
bool-trap templ params stay on its table), the 2026-10-25 data-model roadmap
(`machineSurface bool-last` — note: the `//nolint:branching-flow` comment at
`platform/internal_directory.go:39` claims the position-capture trap is already closed),
the dedup registry's accepted clones, and every AGENTS-documented intentional pattern
(the four 401 writers, `*Any`/`ListAll*` admin seams, byte-pinned literals, English shell
copy D3, the one-home helper list, templ bucketing grades).

## 1. Per-area verdicts

| Area | Verdict |
| --- | --- |
| `internal/server/server.go` (route table, middleware chain, limiters, health wiring) | **Strong.** Chain order matches AGENTS exactly (incl. `ReportingEndpointsMiddleware` beside SecurityHeaders). Registrations-as-route-table (B007) is greppable. Two real findings below (dual mux registration, login CSP override). |
| `internal/server/actions.go` | **Good with defects.** `requireSessionMultipart` prologue and the `sendFailure` ladder are exemplary; import-filename escaping and the contacts/snippets failure shapes are findings. |
| `internal/server/panels.go`, `pages.go`, `notfound.go`, `toast.go`, `unread.go` | **Clean.** Minor: discarded MarkRead error (panels.go:98), repeated `h.lang(r)`/`h.T` cookie parsing per call (cosmetic). |
| `internal/server/sse.go`, `notifier.go` | **Good.** Hub reaper and notification-policy table are well-built; Shutdown log bug + stale sweep comment below. |
| `internal/server/crm.go`, `admin.go`, `reports.go`, `livecontext.go`, `auth_audit.go`, `shell_seam.go`, `wirefixtures.go` | **Clean.** Admin degrade-per-section, `crmContextFor` nil-safe by resolver construction (verified `internal/crm` nil-guards). |
| `internal/server/apidocs.go`, `wirecontracts.go` | **Good.** Declarative catalog is the right shape; stale in-file comment re response-header support. |
| `internal/server/textagent_api.go` | **Clean.** Prologue/gate/action helper trio holds the error contract in one place. |
| `internal/server/hooks/` | **Strong.** `decodeJSON` cap-in-one-home, `applyStatusWebhook` idem ladder, `BearerGate` constant-time — all match the documented contract. |
| `internal/server/api/` | **Good.** Consistent 401/403/404/503 grammar via the Server seam; call-ID bound fragmentation and a bytes-vs-runes inconsistency below. |
| `internal/server/auth/` | **Strong.** `mintSession` birth invariant, SSO cookie hygiene (path-scoped, cleared on every exit), anti-enumeration passkey 401s. |
| `internal/server/platform/` | **Good.** `machineSurface` doctrine correctly applied; `metrics.go` nil-contract violation is the one defect. ETag-per-request hashing is a note, not a finding. |
| `internal/web/views/*.templ` + helpers | **Strong.** Master–detail split, capability-honest fax states, morph-stable ids, responseHandling config and `.wp-error` wire are coherent. One hardcoded English string, minor LSP diagnostics (`templ slicesbackward` at messages.templ:835, `embedlit` hints) — advisory only. |
| `cmd/webphone/` | **Strong.** `bootJoin`/`bootStepError` taxonomy, enum-forced boot copy table, SSE-first shutdown order — all as documented; no findings. |

## 2. Findings

Severity: 🔴 honesty/correctness (name or behavior lies), 🟠 real defect, 🟡 consistency/debt, 🔵 nit.

| # | Sev | Location | Finding | Suggested fix |
| --- | --- | --- | --- | --- |
| 1 | 🟠 | `internal/server/sse.go:123-137` | `Shutdown` logs `"hubs", len(h.hubs)` AFTER the loop deleted every entry — the greppable incident line always reports `hubs=0`. The `streams` count is correct; the hub count is permanently wrong (no test pins it, verified). | Capture `hubCount := len(h.hubs)` before the loop and log that. |
| 2 | 🟠 | `internal/server/actions.go:788-789` (+ `actions.go:545-562`, `toast.go:20-28`) | `importContacts` pre-escapes the filename (`templ.EscapeString(header.Filename)`) and then hands the string to `renderPanelError`, which escapes AGAIN at `actions.go:559` → double-escaped `.wp-error` banner, and the `notifyToast` JSON header carries the once-escaped string → the toast displays literal `&amp;`/`&lt;` entities for filenames containing `& < >`. `renderPanelError` is the documented one-home escaper; every other caller (e.g. `textAgentPanelAction:159`) correctly passes raw text. | Drop the `templ.EscapeString` at the call site; interpolate the raw filename. |
| 3 | 🟠 | `internal/server/auth/login_page.go:31-36` vs `internal/server/platform/csp_report.go:10-15` | The `/login` CSP override builds a fresh policy string that OMITS `report-uri /csp-report` and `report-to csp-endpoint`. The Reporting-Endpoints header still rides the response (global middleware) but without the directives the browser never reports — CSP violations on the login page (the one surface with nonce'd inline scripts, i.e. the highest-value XSS observation target) are invisible to the journal. `csp_report.go`'s own comment claims "the login nonce lane" flows to the one journal — the code contradicts it. | Append `report-uri /csp-report; report-to csp-endpoint` to `loginContentSecurityPolicy`. |
| 4 | 🟠 | `internal/server/platform/metrics.go:66` vs `platform/platform.go:39-41` | `Metrics` calls `p.deps.Transcripts.Count(...)` unguarded, while `platform.Deps` documents Transcripts as nil-safe ("the export degrades to no transcripts.json leg when a hand-composed test composition omits it") and `export.go:105` honors that. A composition omitting Transcripts panics on `/metrics` (nil interface method call). Prod wires it; the documented contract is what's violated. | Guard like export does (`if p.deps.Transcripts != nil`) and render the counter only when present, or drop the nil-safety claim from the Deps doc. |
| 5 | 🟡 | `internal/server/actions.go:299-306, 380-385` (and the 500 paths at `:316`, `:399`) | Contacts/snippets validation failures answer PLAIN `http.Error` text with no `.wp-error` element and no `HX-Trigger`. htmx's `responseHandling` (`select: ".wp-error"` → `#wp-tab-error`, layout.templ:423) finds nothing to swap and shell.js's `htmx:responseError` (shell.js:1680-1699) skips HX-Trigger-carrying responses — so the user sees the GENERIC throttled toast "The request failed (HTTP 422)" and the specific copy ("enter a valid number", "the snippet is too long") never reaches a surface. The messages/fax/drafts lanes answer the same class through `renderPanelError` (specific banner + toast). Two failure vocabularies for one templ form-POST family. | Route these through `renderPanelError` (or a lighter `.wp-error`-carrying writer) so the validation copy lands; English copy is fine per D3. |
| 6 | 🟡 | `internal/server/api/summary_api.go:26,85`, `review_api.go:57,205`, `note_api.go:20,38-40`, `transcripts_api.go:72,134` | The "opaque island Call-ID ≤ 128" bound has FOUR homes: `summaryCallIDMax` const (used inline at 3 sites), a second `noteCallIDMax` const + `validNoteCallID` helper (used by note + disposition lanes), and the bare literal `128` twice in `transcripts_api.go` (the DTO description even says "max 128 chars"). One rule, four spellings that can drift independently. | One `validCallID(id string) bool` helper + one const; the note lane's variant reads through it. |
| 7 | 🟡 | `internal/server/api/transcripts_api.go:93` vs `note_api.go:98-105` | `SaveTranscript` bounds the segment in BYTES (`len(body.Text) > transcriptTextMax`) while `SetCallNote` deliberately bounds in RUNES with a comment explaining exactly this trap (a 4000-char CJK note is 2-4 bytes/char). Same "4000 chars" promise, different units: a full-length non-Latin transcript segment is 413-rejected where the identical note is accepted. | Count runes (`len([]rune(...))`) or lower the byte bound with a comment naming the unit choice. |
| 8 | 🟡 | `internal/server/server.go:700-913` | Root/open DUAL route registration: every pattern registered on `open` must be hand-copied onto `root` (15 entries), else it falls through `session.Attach(..., protected)` into the styled 404 with session/CSRF wrappers it never asked for. Only the `/api/*` lanes have a mount-parity pin (`TestAPIContractPathsAreMounted`); `/health/*` had to be special-cased twice (once for dashboard-on redirect, once for both-404 shapes) — the shape that keeps needing by-hand repairs. | A construction-time assertion that every `open` mux pattern (bar the catch-all) also exists on `root`, or a single `mount(root, open, pattern)` helper that registers both. |
| 9 | 🟡 | `internal/web/views/messages.templ:268` | `ThreadHead`'s back link is hardcoded `← All messages` while the `threads.back` dictionary key ("Back to messages"/"Zurück zu den Nachrichten") exists and is used by the archived toggle ~180 lines above (messages.templ:90). A German UI gets an English back-link on every open conversation. No test or dom-contract entry pins the literal (verified by grep). | `{ "← " }{ T(props.Lang, "threads.back") }` (or a dedicated key if the wording matters). |
| 10 | 🟡 | `internal/server/panels.go:547-554`, `auth/session_identity.go:159-174`, `textagent_api.go:61-72` | Three surfaces fetch the FULL `TextAgent.Reviews()` payload (drafts + frozen + suppressed) just to read 1-4 scalar mode fields — and one of them is `fillAgentMode`, which rides EVERY session mint AND every `GET /api/session` boot probe. `textagent.Client.Reviews` is an uncached loopback GET (verified `internal/textagent/client.go:178`), so every page load with the seam on pays a full-queue roundtrip inside a 3 s budget. | A `Mode()`-shaped ask on the client (or a short TTL memo in the client) for the two decoration sites; `messagesPanelPrefilled` legitimately wants the drafts. |
| 11 | 🟡 | `internal/server/actions.go:66-87, 119-124` | `sendMessage` parses `?thread=` twice: the second `ParseThreadID` + `h.notFound` branch (121-124) is dead code (the first parse already returned) and the id could simply be kept. Small, but it is the exact "parse-at-the-door" shape the repo centralizes elsewhere. | Hoist `threadID` from the first branch (parse once, reuse). |
| 12 | 🟡 | `internal/server/panels.go:98-101` | `threadPanelPrefilled` swallows the MarkRead error completely (`if err := ...; err != nil { _ = err }`) — no log, unlike the rail failure 12 lines below which gets `LogErrorContext`. The error-surface owner rule wants either a visible surface or a written justification; the comment justifies degradation but nothing observes it. | `slog.DebugContext` (badge-cosmetic class) — mirrors countUnread's Warn-at-debug posture. |
| 13 | 🟡 | `internal/server/apidocs.go:23-26` (+ AGENTS.md §Machine-contract) | The file header says "the exporter cannot express response headers at catalog v4.9.0 … the Retry-After contract lives in 429 description TEXT, pinned" — but the same file calls `catalog.WithResponseHeader` 19× and the pinned catalog is v4.15.0 (which has the option, verified). The comment and AGENTS both describe a constraint that no longer exists; the description-text duplication remains as harmless hedging. | Update the comment (and the AGENTS line) to describe description-text + typed header as belt-and-braces. |
| 14 | 🔵 | `internal/server/sse.go:139-143` | The sweep comment claims "get() refreshed its seen stamp microseconds before ServeSSE subscribes" — `get()` only stamps `seen` on CREATE (sse.go:111), not on the exists-branch. Behavior is still safe (SubscriberCount protects mid-connect), the comment misstates the mechanism. | Reword to cite SubscriberCount as the sole mid-connect guard. |
| 15 | 🔵 | `internal/server/auth/passkey_settings.go:34` | `creds, _ := a.deps.UserAuth.Credentials(...)` — the error is dropped with no log; a broken store renders "no passkeys enrolled" (a false claim) silently. Same fail-open class as the Settings decorations, but those log. | Log at Warn like `agentAuditRows` does. |
| 16 | 🔵 | `internal/server/admin.go:296-315` | `fillAdminCalls` is a sequential per-line fan: password-file read + one `PhoneAPI.History` roundtrip per mapped extension. Panel latency scales linearly with line count (admin-only, bounded by the PBX client's own timeout). | Fine at fleet scale; note for a future parallel fan if the inventory grows. |
| 17 | 🔵 | `internal/server/panels.go:410-425` | `recentTranscripts` is an N+1 (Notes.Get + Reviews.Get per transcript group; ≤10 groups unfiltered, ≤100 filtered) — local SQLite, bounded, but the widest per-render query fan in the package. | Acceptable; a joined read later if History render latency ever matters. |
| 18 | 🔵 | `internal/server/platform/assets.go:108-129` | `assetHeaders` re-reads and re-hashes every asset on every request (fs.ReadFile for the ETag, then the file server reads it again). Embedded-FS cheap, but precomputing the tag map at construction is free. | Precompute `map[path]etag` in `Assets()`. |

### Automated observations

- `templ generate` LSP diagnostics report: `slicesbackward` at `messages.templ:835` (firstUnreadIndex could use `slices.Backward`), `embedlit` hints in contacts/fax/login_page/settings, one `QF1003` tagged-switch hint in settings — all advisory; the committed `*_templ.go` artifacts are current against the sources (spot-checked `contactLabel`, `All messages` renders).
- No `go vet`/compile issues encountered in the read; full `go test` not run (read-only review pass; suites enumerated instead).

## 3. Split-brain candidates (for the central adjudication)

1. **Root/open dual mux registration** (finding 8) — the route table exists twice; sync is by hand.
2. **Call-ID bound ×4** (finding 6) — one rule, four homes.
3. **Two failure vocabularies for templ form POSTs** (finding 5) — `.wp-error` panel-error ladder vs plain `http.Error`.
4. **Escaping layer split** (finding 2) — `renderPanelError` is the one-home escaper; `importContacts` keeps a second escape.
5. **CSP observability claim vs login override** (finding 3) — comment says all surfaces report; the login policy opts out.
6. **bytes-vs-runes text bounds** (finding 7) — same "4000 chars" promise, two units.

## 4. Explicit exemptions (checked, not findings)

- The allowlisted four 401 writers, the `*Any`/`ListAll*` admin seams, `.wp-error` as the htmx
  responseHandling wire, `wp-status-*` pins, English shell copy (D3), `RetryAfterHint`,
  per-package `writeJSON` one-liners (documented carve precedent), the deliberate
  one-home helpers — all present as documented and none misapplied at new sites.
- `fax.templ:168` reusing `.wp-error` inside a 200-rendered row: class overload, not a
  responseHandling hazard (selection only happens on 4xx/5xx).
- `LogCall`/`crmContextFor` calling nil-`*crm.Resolver` methods: verified nil-guarded in
  `internal/crm` (pointer receivers with `r == nil` checks) — not panics.
- `enrollPasskey`'s `RPOrigins[0]`: safe — `config.validatePasskey` refuses an enabled
  passkey config with empty `rp_origins` (config.go:1256).
- Export's unbounded `ListThreads`: verified `cqrs.ownerSummaries` lifts the engine limit
  (`WithLimit(0)`), so "everything the extension owns" is true.
- Login page English-only copy, `AllTabs`/`navGroups` ordering, the double CSP
  (description text + header) in apidocs — deliberate, documented.

## 5. Strengths to preserve

- The B007 route-table registrations: every gate (limiter, body cap, CSRF, session) visible at the mount site — this review could audit the whole auth posture by reading one function.
- `renderPanelError`/`internalError`/`safeDetail` error ladder: full detail to the operator log via `errorfamily`, family-default copy to the browser — consistently applied everywhere except the two findings above.
- `directoryFor` shared by request renders AND the SSE push lane (the morph-rename split-brain prevented by construction).
- The notifier's event→SSE policy table (`Notifier.Project`) — exactly the machine-table doctrine AGENTS prescribes.
- `bootreport.go`'s enum-forced copy table (new class without copy = compile/test failure) and `bootJoin`'s errors.Join-compatible traversal.
- `wirefixtures.go`'s `json.Deterministic` + stub-keeps-render-panic-free posture.

## 6. Verdict

The HTTP/server surface is in strong shape: the architecture invariants AGENTS documents are
actually enforced in the code (route gates, nil-seam doctrine, one-home helpers, anti-enumeration
postures), and the test surface is unusually dense (362 server tests pinning the documented
contracts). Eighteen findings, none architectural: four genuine defects (an always-zero
observability count, a double-escaping path, a CSP-reporting gap on /login, one nil-contract
violation on /metrics), the rest consistency debt — duplicated bounds, two failure shapes for
templ form POSTs, and a handful of unlogged degradations. Nothing blocks; findings 1-4 are
cheap, surgical fixes and findings 5-13 are small consistency trains.
