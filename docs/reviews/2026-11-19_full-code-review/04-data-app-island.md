# Full code review 2026-11-19, part 4 of 4: journal/data layer, composition root, arch tests, browser island

> Snapshot-rot header — **claims verified as of 2026-11-19 ~18:40 CET**; every cited
> line was read in source today. Scoped suites run green today, all inside the
> sanctioned environments: `nix develop -c go test -count=1` over
> `internal/{cqrs,readmodels,store,app,arch}` (all ok), host
> `node --test --test-force-exit internal/web/assets/island-tests/*.test.mjs`
> (**364 pass / 0 fail / 0 skipped / 0 todo**), and `nix develop -c tsc -p . --noEmit`
> (**green** — the type gate's own authority). Companions: `03-integrations-identity.md`
> in this directory (cross-referenced, not re-discovered) plus the same-day
> data-model/naming reviews it cites. Repo discipline was READ-ONLY: findings only,
> no code changed, nothing committed. `internal/arch/model_api_test.go` carried a
> concurrent session's uncommitted edits — read as-found, untouched.

## Scope + method

Five Go packages + the whole non-vendor browser-asset tree: **~130 files visited**
(~60 Go code+test, ~30 island/shell/enroll/login/theme JS, ~55 island-tests +
config/decl files). Load-bearing files were read end-to-end (every cqrs decider,
backend, and battery; the whole store ladder; app's composition root; island core
modules calls/connection/session/main/state/ui/auth/config/dialgate/latency and
the type-gate pins). The long tail (small island feature modules, the island-tests
beyond `helpers.mjs` and the largest specs, `types/*.d.ts` beyond `island.d.ts`)
was visited structurally — full export/surface map, doc-comment read, anti-pattern
greps (console.log: none; TODO/FIXME: one prose mention) — and is reported at that
depth. Judged against AGENTS.md doctrine: ONE-journal-world, gates+views purity,
frozen ladder, WithLimit ceiling, depth budget, import-closure + state/auth
acyclicity, honesty contract, English #log, strict CSP pins. Per the brief: the
frozen ladder, JS string-builder patterns, deliberate #log English, and the two
pinned CSP exceptions were not flagged.

## Per-file verdicts

| Area                           | Files                                                                                                                                                          | Verdict                                                                                                                                                                                                                                                                                                                                                                               |
| ------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `internal/cqrs` deciders/state | threads*, contacts, snippets, dial*, transcripts, faxes, pushdevices, agentaudit, callnotes, callreviews, enroll, domain.go, engineaggregate.go, visibility.go | **Exemplary.** Machine tables consumed verbatim at the decision; StrictApply vocabularies; caps reject-or-trim per the data-loss rulings; the D7 `awaitOwnWrite` (250 ms budget, WARN-and-proceed) is a genuinely well-engineered read-your-writes seam; journal-trim deletion pins (bytes gone, fresh replay clean) are the strongest tests in the repo. Nits only (F2, F3, F6, F9). |
| `internal/cqrs` backends       | threadsstore/reads/views/admin/sweep, transcriptsstore, faxesstore, dialwindows, pushdevices, agentaudit, callnotes, callreviews, enroll backends              | **Strong.** Owner scoping everywhere; `*Any`/`ListAll` seams named and pinned; dedup rides `sys.Dedup()` per doctrine. Findings: ceiling-lifted full-scan lookups (F5), Seq-as-clock coupling (F4), clock-seam bypass (F1).                                                                                                                                                           |
| `internal/cqrs` tests          | 16 battery/bench files                                                                                                                                         | **Strong.** Every aggregate has restart-reload, cap-boundary, and idempotence pins; decision-level tests where dispatch would be too slow; cross-process enroll pin; benchmarks documented against the A38 budget. Nits: `facesDrive` wrapper (F10), test-side boot-helper copies (F7).                                                                                               |
| `internal/readmodels`          | 14 gates+views files + family_test                                                                                                                             | **Clean.** Purity holds (arch test verified live); sentinel families pinned incl. the one 409 Conflict. FaxesProjection is the deliberate one-off fold living here (F8); one inverted inline comment (F6); `humanScaleVolume` dual home (F8).                                                                                                                                         |
| `internal/store`               | db.go, messages.go, sweep.go, appflags.go + 10 test files                                                                                                      | **Clean.** Ladder frozen exactly as ruled (v1–v5, no new steps); `listRows`/`updatedOrNotFound` one-home helpers pinned with family tests; messages.go honestly marked TEST-ONLY with its §5 DROP future. Nits: `err == sql.ErrNoRows` (F11), dead legacy const (documented retention), tag-gated dual-run spike can rot silently (F13).                                              |
| `internal/app`                 | app.go, cqrs.go, chat.go, dashboard.go, dedup.go, flightrecorder.go + 9 test files                                                                             | **Strong.** Composition-root rules hold (probe outside the injector it reads; named eager critical pair; typed-nil discipline pinned TWICE — SSO and trace seams); boot-cycle test drives login→send→hook→SSE through the real handler. Nits: hollow-root Warn dead path, 11 sync.Once boilerplate pairs (F12).                                                                       |
| `internal/arch`                | arch, depthbudget, errorregistry, gitignore, model_api, wiring_parity + 5 probe fixtures                                                                       | **Strong.** The registry generator is daemon-reflow-proof (self-tested); wiring-parity closes the M16 class. One real enforcement gap: the services-never-import-server list covers 9 of ~25 service packages (F1 — this review's top finding).                                                                                                                                       |
| island `app/*.js`              | 23 modules                                                                                                                                                     | **Strong.** Honesty contract visibly enforced (CALL_TRANSITIONS table + verdict lane + pending/queued hold machine); watchdog/rebuild in connection.js is battle-documented; every error path lands in a surface per the 2026-10-27 ruling. calls.js is 4.9× the size guideline (F14); one shadowing nit (F15).                                                                       |
| shell.js / theme-preload.js    | 2                                                                                                                                                              | **Sound but monolithic.** shell.js is a 3,316-line single IIFE with 24 numbered sections; the classic-script constraint is doctrine but the section numbering has drifted (F16). theme-preload is exactly right (pre-paint, storage-throw-safe).                                                                                                                                      |
| enroll/, login/                | 3 files + specs                                                                                                                                                | **Clean.** Standalone-entry doctrine held (outside main.js's closure — pinned by the arch test); base64url coercion pure and round-trip pinned.                                                                                                                                                                                                                                       |
| island-tests + type gate       | 53 spec files, helpers.mjs, typecheck_pin_test.go, tsconfig.json, types/_.d.ts, oxlint.json, assets_.go, tokens/styles tests                                   | **Exemplary test discipline.** 364/364 green, zero skips; helpers.mjs documents its deliberate stub asymmetry; the type-gate pin test makes the gate un-gameable (include/exclude/strict/@ts-ignore/nix-lane all refused from the EMBEDDED tree).                                                                                                                                     |

## Findings

Severity scale: **small** (real defect, cheap fix) · **low** (debt/gap with a named
fix) · **info** (observation). Nothing medium or high found in this layer.

### F1 · low · internal/arch/arch_test.go:73-77 — services-never-import-server test covers a stale subset

`TestServicesNeverImportServerOrWeb` hardcodes 9 packages
(`blob config fax gateway messaging pbx session store vcard`). The service family
has since grown to ~25 — `cqrs`, `readmodels`, `push`, `asr`, `crm`, `paperless`,
`textagent`, `matrix`, `signal`, `chatloop`, `userauth`, `authmap`, `admin`,
`sso`, `reports`, `retention` are all OUTSIDE the check, so a future
`internal/cqrs → internal/server` import would pass every gate. No violation
exists today (verified: `go list -deps ./internal/cqrs` pulls no server/web), so
this is an enforcement gap, not a breach — but cqrs is the largest service
package and the newest, exactly where the invariant is least protected.
**Fix:** derive the set from `go list ./internal/...` minus
`{server,web,web/**,app,arch}` instead of the hand list (the file already shells
to `go list` for the other checks).

### F2 · low · internal/cqrs/dialwindows.go:110 — `Active()` bypasses the seam's own test clock

The dial seam defines `var timeNow = time.Now` "overridable for deterministic
tests (window expiry, cap-expunge ordering)" (line 48-50) and
`dial_test.go:175,233` does override it. `Active()` filters expiry with a bare
`now := time.Now()` instead, so the one read whose semantics ARE window expiry
cannot be driven deterministically by the seam's own clock. The existing test
works around it by stamping a 2-hour-stale slot instead of freezing time.
**Fix:** `now := timeNow()` at line 110 (one line); optionally freeze the clock
in `TestDialActiveListsUnexpiredWindows` to pin why the seam exists.

### F3 · low · internal/cqrs/dial.go:216-220 vs 265-269 — duplicated overflow-trim logic

`dialExpungeEvents` (oldest-first insertion sort with remote tie-break, then
expunge the overflow) is re-implemented verbatim inside `decideSaveOverride`
because the helper sorts `map[string]int64` and the override path builds its own
stamp map first. Same comparator, same tie-break, 100% copy.
**Fix:** extract `oldestOverflow(slots map[string]int64, cap int) []string` and
let both decision paths call it (the windows path already maps into the same
shape); keeps the determinism contract in one place. Related info nit:
`dialWindowsMaxSlots`/`dialWindowMaxPerOwner`/`dialOverridesMaxSlots` are three
names for one number (dial.go:35-42) — the alias chain reads as leftover intent.

### F4 · low · internal/cqrs/transcriptsstore.go:92 + :284 — Seq silently doubles as the retention clock

`Append` stamps `Seq: time.Now().UnixNano()`; `SweepBefore` filters
`metaengine.WithFilter("seq", metaengine.FilterLt, cutoff.UnixNano())`. The
retention sweep's correctness therefore rests on an invariant expressed nowhere
at the filter site: "Seq is the append wall-clock." `TranscriptSegmentView`'s
comment says "Seq is the append order (nanos at the seam)" — sort-order framing,
not retention framing. A future writer stamping Seq differently (per-call
counter, provider sequence) silently breaks retention while every test stays
green (tests seed Seq in nanos, matching by accident of the same convention).
**Fix:** one comment at SweepBefore naming the coupling ("Seq IS the append
stamp — the cutoff compares wall-clock nanos"), or stamp an explicit
`AppendedAt`-shaped fact field when the aggregate next evolves. Same-file info
nit: `Call()` (line 208) takes the backend mutex while `Recent`/`Search`/
`Export`/`RecentAll` do not — harmless (the mutex guards dispatch+tail-cache
pairs) but the asymmetry is unexplained.

### F5 · low · internal/cqrs/threadsreads.go:230-264 (and threadsstore.go:50-64/146-164, threadsadmin.go:40-52) — ceiling-lifted whole-corpus reads per request

`attachAttachments` scans the owner's ENTIRE attachment projection
(`WithLimit(0)`) and sorts it, to hydrate one ≤200-message transcript page;
`threadByID`/`FindThread` scan the owner's whole thread set and linear-search;
`GetThreadAny` scans ALL threads (`WithLimit(0)`) and then calls `GetThread`,
which scans the owner's set again (double full scan). All are documented
ceiling-lifts in the SQL-parity spirit, and single-tenant volume makes them
fine today — but the transcript render is the hottest read in the product and
its cost is O(owner attachments), not O(page). The messages projection declares
no `message_id IN` capability; the attachments projection has an unused `id`
filter. **Fix (when measured slow):** declare `FilterOnField("message_id",
FilterEq)` on attachments and query per page chunk (the SQL twin's chunked IN
pattern, messages.go:379-427, is the precedent); or a `thread_id` filter. Not
urgent — record as the A38-follow-up perf ledger row.

### F6 · low · internal/readmodels/transcriptsgate.go:103-108 — inline comment inverted against the code

The function doc (line 70-77) says "newest call first, each call's lines in
spoken order," and the code appends fetch-order lines then `slices.Reverse` per
group — which yields oldest-first lines. The inline comment says the opposite:
"The store yields oldest-first per call … each call's lines are presented
newest-first." Both halves are inverted (the fetch is seq-DESC = newest-first;
the presentation after reverse is oldest-first/spoken). It reads like a stale
port note from the SQL store's prepend era and misleads exactly the reader it
exists for. **Fix:** rewrite the two lines to match the doc ("fetch is
newest-first; the per-group reverse restores spoken order").

### F7 · low · boot-helper duplication across cqrs (4 prod + 5 test copies)

`NewThreadsSystem` (threadsregister.go:135), `NewCoreSystem` (:163),
`NewFaxesSystem` (faxesstore.go:291), `NewEnrollSystem` (enroll.go:192) — and
test-side `bootSnippetsSystem`, `bootDialSystem`, `bootPushDevicesSystem`,
`bootAgentAuditSystem`, `bootEnrollSystem` — all inline the same
`DeploymentConfig{one sqlite engine, source-of-truth + projections, WAL}` +
backoff + best-effort-close shape. `bootDomainSystem` (phase4_test.go:32)
already exists as the shared test home and is used by half the batteries; the
other five re-inline it. **Fix:** collapse the test boots onto
`bootDomainSystem`; on the prod side one `bootSystem(ctx, dsn, decl, code)`
helper would leave each aggregate's boot a two-liner. (~120 duplicated lines.)

### F8 · info · two-name/one-fact drift inside the readmodels/cqrs boundary

1. `humanScaleVolume = 1` is declared in BOTH `cqrs/threadsviews.go:71` and
   `readmodels/faxesgate.go:17`, with different comment text — two homes for a
   planner-tuning doctrine constant.
2. `FaxesProjection` is the only aggregate fold living in readmodels (consumed
   by cqrs's `FaxesDomainDecl`), while every other projection lives beside its
   decider in cqrs. Deliberate ("the folds' one home") but asymmetric — the
   gates+views-only doctrine makes room for it only via the "query decls the
   journal backends consume" clause in doc.go.
3. `AGENTS.md:127` still cites `readmodels/admin_reads_test.go` as the
   transcripts admin-seam pin; the file retired with the engine and the pin now
   lives in cqrs (`TestTranscriptRecentAllCrossesOwners`). Docs-drift fix is a
   one-line AGENTS edit naming the cqrs test.

### F9 · info · cqrs decision nits (no behavior defect)

1. `threadsstore.go:100-108` — `AdvanceOutbound` re-parses Owner/Network/Remote
   from the view up to 4× (once for the key, again inside the command literal);
   a parsed triple reads cleaner.
2. `threads.go:163-192` — `SetThreadFlag` emits an event even when the flag is
   already in the desired state, while `MarkThreadRead` (:194) short-circuits a
   no-op. Replay-safe either way; the asymmetry is unexplained.
3. `threadsadmin.go:100-101` — `var _ = errors.Is` keeps an otherwise-unused
   import alive under a "compile guard" comment; deleting import+guard is
   cleaner.
4. `agentaudit.go:120` — cap-trim sorts by Seq with NO tie-break (dial's
   trim tie-breaks by remote; `sort.Slice` is also non-stable): same-nanosecond
   appends pick a nondeterministic victim. Journal-once semantics make it
   harmless; deterministic victims are the house style (dial.go:215-219).
5. All aggregate command registrations execute under `context.Background()`,
   not the dispatch ctx (e.g. threadsregister.go:70, dial.go:303) — almost
   certainly deliberate (a canceled HTTP request must not tear a journal
   append) but the choice is documented nowhere; one comment at one
   registration site would pin it.
6. `threadssweep.go:57-83` — a message appended between the sweep's scan and
   its tombstone decision lands IN the tombstone (the decider sees live state)
   but its attachments' blob paths were gathered before → orphaned file.
   Bounded (daily sweep, one-message window); worth a comment at the path
   gathering, not code.

### F10 · info · small test/store nits

1. `faxes_test.go:81-84` — `facesDrive` is a typo'd one-line wrapper around
   `UpdateStatus` that adds nothing; inline the call.
2. `store/appflags.go:34,48` — `err == sql.ErrNoRows` where the package
   convention (messages.go, db.go) is `errors.Is`; works with modernc's
   sentinel today, fragile to a wrapped driver.
3. `store/appflags.go:22` — `FlagLegacyContactsImported` has no reader
   (importer deleted 2026-11-09); its comment says it stays "for operators
   inspecting stamped flags" — accepted retention, noted for the §5 DROP.
4. `store/threadsdual_test.go` — the `//go:build threadsdual` T20 spike has no
   CI lane that compiles it; it can rot silently. Either a compile-only CI
   step or a retirement note in the header.
5. `cqrs/cqrs_test.go` etc. use the `t := &testing.T{}` trick to reuse
   helpers in benchmarks — works (helpers only Fatalf/Helper), worth a one-line
   comment at the first use since it surprises readers.

### F11 · info · internal/app/cqrs.go boilerplate and a dead path

1. Lines 55-80 + 109-184: eleven `sync.Once` + field + accessor triplets. All
   are constructed during `New` (the handler provider invokes every `XxxSvc()`),
   so the Once guards a second construction that cannot happen in prod. A tiny
   `lazy[T]` generic or constructing the set once in `newCqrs` collapses ~75
   lines of pairs.
2. Lines 233-235: the "cqrs hollow root" WARN can never fire — `NewCqrs`
   always passes the full 11-aggregate domain. Phase-1 residue; delete or
   gate it behind `newCqrs`'s test-only callers honestly.

### F12 · info · register-thread duplication of the summary-image twin contract

`applyThreadState`'s MessageAppended branch adopts the fact's embedded
`Summary` while `appendToState` recomputes it through `domain.AppendToThread`,
with the equality pinned only by the battery's replay tests. The design is
documented ("one rule, two sights"), and the pins exist — the note is that the
equality is a RUNTIME invariant with no compile-time expression; a
`TestApplyAdoptsDecisionImage`-named pin (if not already inside the battery)
would make the contract greppable.

### F13 · low · internal/web/assets/island/app/calls.js — 1,701 lines (4.9× the 350-line guideline)

The module is coherent and superbly commented, but it carries at least four
separable concerns that already have natural module seams and zero import-cycle
risk (each only consumes state/ui/i18n): the call dock (387-490), the
caption/transcript strip + save queue (677-943), the transfer machinery
(945-1062), and the ringing queue (1614-1705). `postcall.js` already proved the
extraction pattern. No depth-budget pin covers JS, so nothing forces this —
which is exactly why it grew. **Fix:** carve `dock.js` and `ringqueue.js` first
(they are self-contained), keep the re-export surface in calls.js so the
import-closure test and the DOM contract stay untouched. Same class, one notch
bigger: `shell.js` at 3,316 lines (F14).

### F14 · low · internal/web/assets/shell.js — 3,316-line IIFE; section numbering has drifted

The single-classic-script constraint is doctrine (must run when island modules
fail; CSP-strict) and does NOT forbid multiple classic scripts — but 24 numbered
sections in one IIFE is past the point where navigation works: the section map
now has TWO different "2b" sections (line 384 `2b2` under "2b", and line 1267 a
second top-level "2b. data-reload buttons" appearing AFTER "2f" at 1198), and
the 2b-subseries starts at 2b2 with no 2b1/2b3. **Fix:** either renumber the
map once (cheap, mechanical, no behavior), or split at the big natural seams
(voicemail player 3i ≈ 450 lines; optimistic send 3e; theme/island/rail §4x)
into additional served-verbatim classic scripts — the DOM contract and E2E
greps survive because they pin ids/classes, not script tags.

### F15 · info · island JS nits

1. `calls.js:1139` — inside `bindSession`, `const target = …` (a DOM element)
   shadows the function's `target` parameter (the phone number) used at
   :1152/:1246; rename the inner one (`stripEl`).
2. `audio.js` uses `console.warn` for sink/resume failures with an inline
   justification each time — consistent with the visible-surface ruling (the
   surfaces are audible: the tone plays or not); fine, noted for completeness.
3. The i18n formatter-parameter discipline (`defineMessages` identity trick)
   is holding — no `any` leakage found in the read modules.

### F16 · info · cross-reference still open from part 3

Report 03's F2 (data-model F2) — `push.Transport` typed enum stops at the cqrs
device gate — is confirmed still open in this layer:
`cqrs/pushdevices.go:43,228` carries `Transport string` raw, re-parsed
defensively in `app.pushRegistryAdapter.List` (app.go:151-154, skip-on-parse).
The `//nolint:branching-flow // verbatim port of the retired engine store's
Upsert signature (parity rule)` makes it a deliberate deferral, but the
skip-on-parse at fan time is the shape a typed gate would remove.

## Split-brain candidates

1. **Seq-as-clock** (F4) — one field, two roles (sort order + retention time);
   the coupling lives only in convention.
2. **humanScaleVolume dual declaration** + **FaxesProjection's readmodels home**
   (F8) — two small same-concept-two-places drifts across the cqrs/readmodels
   boundary.
3. **AGENTS.md's admin-reads pin citation** (F8.3) — doc names a test file the
   tree no longer carries.
4. **shell.js section numbering** (F14) — the file's own navigation map
   disagrees with itself.
5. Verified NON-candidates: `errorMessage` twin (ui.js/shell.js — documented
   module-system boundary), `dialEvent` vs `transcriptEvent` increment
   conventions (per-file consistent), `contactsCollection` const vs literal
   `"snippets"` strings (inconsistent but single-package), readmodels gates
   (single implementations behind cqrs backends — no engine twin remains).

## Verdict

This quarter is the structural backbone of the SUPERB end-state and it holds:
the ONE-journal-world doctrine is physically true (no second engine anywhere —
the purity arch test is live and the readmodels files are contracts only), the
frozen ladder is untouched, dedup rides `sys.Dedup()`, the depth budget ratchets
hold (store 1,297/1,900 · app 1,606/1,650 · cqrs pinned with its ruling), and
the island's honesty contract is enforced by a declarative transition table
with a verdict lane — the strongest small-scale design in the repo. The test
culture is exceptional: journal-byte deletion pins, cross-process enroll
visibility, daemon-reflow-proof generators, an un-gameable type gate, and 364
island specs with zero skips. The genuine debts are an arch-test coverage gap
that silently excludes every post-2026-09 service package (F1), a bypassed test
clock (F2), duplicated decision logic and boot plumbing (F3/F7), whole-corpus
read patterns on the hottest path (F5, fine at single-tenant scale), and two
oversized, well-commented JS monoliths whose own navigation has begun to drift
(F13/F14). Nothing medium, nothing high, no security regression; ~130 files
visited, all scoped suites green, tree left untouched.

_Unfixed findings are forward-looking work — the actionable rows above (F1…F16)
belong in TODO_LIST.md at the orchestrator's discretion, not entombed here._
