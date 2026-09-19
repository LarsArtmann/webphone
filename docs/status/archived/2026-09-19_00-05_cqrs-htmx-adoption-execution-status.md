# Execution Session Status — cqrs-htmx Adoption Plan, Executed

**Date:** 2026-09-18/19 (session ended 00:05 CEST) · **Trigger:** user approved the
Pareto plan with Full Execution Mode ("NOW GET SHIT DONE") · **Input:**
`docs/planning/2026-09-18_21-45_cqrs-htmx-adoption-pareto-execution-plan.md`

## a) FULLY DONE

1. **P0 — baseline + tag verification.** Tree clean, full suite green before any change.
   G1/G2 re-verified every library claim against the _consumed_ v4.9.0 bytes (module
   cache), not master. One audit claim died here: `ServeSSE` at v4.9.0 sends **no**
   `retry:` hint; `transport.DefaultRetryHintMillis` does not exist at this tag.
2. **P1 — the 1% (51% of the result).** `recovery()` deleted, `cqrshtmx.RecoveryMiddleware`
   in (full stack + method/path in the log, `http.ErrAbortHandler` re-raised — a real
   correctness bug fixed); `cqrshtmx.RequestLoggingSlog` outermost — the server finally
   logs every request. Panic-path, header-survival, log-shape, and no-credentials tests
   shipped (`middleware_test.go`, `requestlog_test.go`).
3. **P2 — the 4%.** `ratelimit.go` (82 lines) deleted; login/hooks ride
   `httputil.KeyedRateLimiter` with port-stripped peer-host keys (`remoteHostKey`,
   flip rule documented), computed `Retry-After` pinned by test. `/healthz` no longer
   lies: `ReadinessHandler` with named `sqlite` + `blob-dir` checks; 503 names the
   failing check (`healthz_test.go`, incl. fault injection via new Deps mutators).
4. **P3 — ServeSSE collapse + all four gates.** The 36-line hand loop is gone;
   `events` = session gate + language memory + `hub.ServeSSE`. Wire shape byte-verified
   (`TestSSEStreamCarriesConnectedThenEvents`): connected frame first, payloads
   unchanged. V1 (full suite + DOM contract), V2 (vet/gofmt), B1 (buildflow),
   B2 (`nix flake check`) all green. E2E verdict: zero markup/ids changed, the additive
   `connected` frame is listener-inert → upstream E2E skip documented.
5. **Docs + HARVEST.** AGENTS.md middleware-chain invariant + adoption posture +
   tag-verification lesson; CHANGELOG entry; plan items routed to TODO_LIST (bounded)
   and ROADMAP (long tail) — commit `6015051`.
6. **P4 — audit-integrity closure.** All previously unread server files read
   (webhooks bodies, actions tail, panels, proxy, configjs, assets, unread): **no
   additional duplication** — sweep recorded in the report's rubric appendix. Symbol
   map complete (ack/notify/decoder/partial/redirect/security/event-catalog/openapi/
   errors_status). Report amended ANNOTATE-style (inline, dated, originals struck):
   score 62 → **92**, 0/4 → 4/4 middleware, F4's retry claim corrected in place,
   rubric appendix added, action table gained effort-minutes + status columns.
7. **P5 — the shippable enhancements.** `GET /events` rate-limited (EL1) with the
   limiter/gate ordering pinned by test (OR1); hub idle reaper (HT1/HT2, 10-min TTL,
   subscriber-count + seen-stamp double guard, deterministic tests); `/version`
   (VE1, DebugHandler pattern); root stack composed via `cqrshtmx.Chain` with a
   behavioral parity test (CH1); opt-in Server-Timing (`WEBPHONE_DEBUG_TIMING`, ST1);
   webhook idempotency (ID1-ID3 — replays answer 202 inertly, failures stay retryable,
   end-to-end single-write proof); toasts over the ToastDetail wire shape on all six
   tab actions with en+de i18n (TO1-TO3); SSE live pill (CO1/CO2, JS-created, zero
   markup change); island 429/Retry-After surfacing (RA1); OpenAPI 3.1 for the session
   API at `/openapi.json` (OA1/OA2); session TTL/sweep interaction tests (TT1);
   webhook-decode fuzz target with 30 s live run — 844k execs, zero findings (FZ1/FZ2);
   hub fan-out benchmark + baseline doc (HB1/HB2); vulnix runtime closure re-verified
   — still 8 derivations, unchanged glibc (VL1); dep drift checked — no newer releases
   (DR1); master re-diffed — the SSE retry hint arrives with the next root tag, bump
   checklist written to ROADMAP (MD1); OOB spike verdict **PARKED** with written
   adoption criteria (OO1-OO3); hx-boost/sync/DecodePagination/StructuredError/
   formatter/readiness-body/notify-ack N.A. records live in the rubric appendix and
   ROADMAP (P7).

## b) PARTIALLY DONE

1. **Stack browser E2E** — skipped by documented verdict (zero markup change), not run.
   The re-run trigger stays: any future SSE payload-shape or markup change.
2. **OOB badge push** — parked, prototype specified in
   `docs/reviews/2026-09-18_oob-badge-spike-verdict.md`; UB1 gated on it.
3. **JS-side tests** — the island has no JS test runner; island behaviors
   (toasts listener, live pill, 429 surfacing) are pinned by Go asset tripwires +
   code review only. Recorded as a standing gap.
4. ~~**Push to origin** — not performed; branch is ahead by ~30 commits. The TODO_LIST
   row for pushing the accept/reject fix (`00f13fe`) plus these commits is the
   operator's move.~~ done: the daemon pushed, and the 07:43 session verified `44db922` at origin (push range analysis §b.1 there); the daemon push behavior is now documented in AGENTS.md (2026-09-19).

## c) NOT STARTED

Nothing from plan phases P0-P7. Everything listed under b) is a deliberate gate or
verdict, not undone work.

## d) TOTALLY FUCKED UP

1. **The audit's retry-hint claim was master-cited** — caught and corrected, but it
   means the original report shipped one wrong capability claim. Fix: rubric appendix
   - F4 strike-through + the tag-verification lesson in AGENTS.md.
2. **`rg -rn` earlier read as recursive** — it is `-r n` (replace); briefly produced
   misleading grep output before being caught and re-run correctly.
3. **Two commits lost their intended messages to the auto-commit daemon** (P2's swap
   and parts of P1/P3 landed as `chore: auto-commit ...` heuristic commits; my
   detailed messages then describe only the remainder). Content is correct and
   verified; the _history_ is noisier than the per-task invariant demands. The daemon
   race is documented in AGENTS.md, but in hindsight I should have committed each
   task the moment its tests went green instead of batching verification.
4. **One bash append landed mangled** (healthz readiness block with a stray comment +
   half a function) — caught by the immediate build, fixed before any commit.

## e) WHAT WE SHOULD IMPROVE (process notes for the next session)

1. `hookFaxStatus` maps _every_ store error to 404 (message counterpart
   distinguishes 404/500). Pre-existing wart, noted here; fix separately with the
   provider-retry contract in mind.
2. The `client`/`testServer` helpers grew organically (variadic mutators); a small
   builder would slow the next test addition.
3. `runbook` strings: the new request-log lines are English by policy; toast copy is
   i18n — both enforced by tests, keep it that way.
4. JS: when the island ever gains a test runner, port the asset tripwires into real
   DOM tests.

## f) TOP THINGS TO DO NEXT

1. ~~Push `main` (~30 commits) and bump the telephony stack's `webphone` input lock
   (TODO_LIST row 1 — includes the earlier accept/reject fix).~~ done (daemon push + stack bump `fc6bc81`, 07:43 report §a.10)
2. ~~Tag v2.0.0 (TODO_LIST row 2) — the CHANGELOG's Unreleased section now has real
   content; release it.~~ done at `d9d6d03` (07:43 report §a.9)
3. ~~On the next cqrs-htmx root tag: follow the bump checklist (ROADMAP, MD1 finding).~~ routed (ROADMAP MD1 bump trigger)
4. ~~OOB badge push: un-park only with the stack E2E loop available.~~ **Won't implement for now —** parked with written adoption criteria (`docs/reviews/2026-09-18_oob-badge-spike-verdict.md`)

## g) OPEN QUESTIONS

1. ~~XFF sanitization upstream (flips `remoteHostKey` → `KeyExtractorFromClientIP`)
   — still unanswered; safe default shipped, flip rule documented.~~ Still open — now tracked in ROADMAP "Open questions" (rate-limit keys).
2. None of the original three remain: Q2 (healthz semantics) resolved by
   implementation, Q3 (SSE stream freeze) resolved by verdict (see the annotated
   status report of 21:38).
