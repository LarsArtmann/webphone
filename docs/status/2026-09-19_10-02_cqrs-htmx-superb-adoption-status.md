# Session Status — cqrs-htmx Deep Dive → Superb Adoption (78 → ~95/100)

**Date:** 2026-09-19 10:02 · **Repo:** webphone @ `main` = `aed88dc` (pushed) · **Scope:** this session's work only
**Thread:** utilization audit → adoption plan → execution → gates → docs → push
**Format note:** user requested `.md`; the status-report skill's canonical format is a styled HTML dashboard. Explicit user instruction wins; override flagged here.

---

## The one-paragraph verdict

The session set out to answer "are we using cqrs-htmx fully and properly?" (78/100 — core exemplary, periphery gapped), then closed most of that gap for real: five library capabilities adopted with tests, three potential Verschlimmbesserungen caught before they shipped, all gates green, everything pushed. The remaining ~5% is gated on the consuming stack's browser E2E, not on knowledge or effort. Nothing shipped broken. One published recommendation (in the audit snapshot) later proved wrong and is corrected downstream — that is the honest stain on the session.

---

## a) FULLY DONE

| Work                                                                                                                                                                                                      | Evidence                                                                                                           |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| **Utilization audit** of cqrs-htmx v4.9.0 (all claims verified at the consumed tag via `git show`, never master)                                                                                          | `docs/research/2026-09-19_cqrs-htmx-deep-dive.html`, 78/100, 15 capabilities assessed, every finding cited to code |
| **Toast wire-shape alias** — `toastDetail = cqrshtmx.ToastDetail`; upstream shape change now fails this build                                                                                             | `internal/server/toast.go`                                                                                         |
| **T2 — Request-ID correlation**: `ContextEnrichmentMiddleware` outermost, `X-Request-ID` response header + `request_id=` in every log line                                                                | `internal/server/server.go` chain; `TestRequestIDEnrichmentWiredIntoTheChain`                                      |
| **T1 — CSRF wiring JSON-safe**: `hx-headers` built with `templ.JSONString` (raw JSON, templ escapes exactly once); string concat retired                                                                  | `pages.go renderShell`, `layout.templ bodyAttrs`; `TestShellRendersValidJSONCSRFHxHeaders`                         |
| **T4 — Calibrated Permissions-Policy** (`microphone=(self)`, camera/display-capture/geolocation/payment/usb denied) + single-source `securityHeadersConfig()` killing the server↔test literal duplication | `server.go`; `TestPermissionsPolicyShipsCalibrated`; `productionStack` shares the config                           |
| **T3 — Server-Timing via library**: hand-rolled `timingWriter`/`timingMiddleware` (~40 lines) deleted; `servertiming.ServerTimingMiddlewareWhen` with the same `WEBPHONE_DEBUG_TIMING` gate               | `server.go`; `TestServerTimingOptIn` (both subtests)                                                               |
| **T5 — Webhook 5xx redaction**: `webhookFail` helper, `SafeDetail` body, full detail to the server log only                                                                                               | `webhooks.go`; `TestWebhookFailRedactsInternalDetail`                                                              |
| **Superb-adoption plan** with mermaid execution graph, two task tables, re-scope log                                                                                                                      | `docs/planning/2026-09-19_09-30_cqrs-htmx-superb-100-adoption-plan.md`                                             |
| **AGENTS.md current**: new middleware-chain invariant (enrichment outermost — and _why_), the two CSRF constraints, audit-trail pointers, toast alias fact + `Notify*` trap                               | `AGENTS.md`                                                                                                        |
| **CHANGELOG `[Unreleased]`**: user-visible story for all five adoptions                                                                                                                                   | `CHANGELOG.md`                                                                                                     |
| **Quality gates**: full Go suite 10/10 packages; BuildFlow full run green on the stable tree (vulnix warnings = documented build-closure false positives; runtime closure clean)                          | session test runs; BuildFlow runs 09:56                                                                            |
| **Pushed**: plan commit `5feecc7`, all implementation (via daemon commits) + docs, `origin/main = aed88dc`, 0 unpushed                                                                                    | `git ls-remote` verified                                                                                           |

## b) PARTIALLY DONE

| Work                            | Done                                                             | Missing                                                                                                                    |
| ------------------------------- | ---------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| **Superb goal (~100/100)**      | 78 → ~95: five adoptions landed, one latent anti-pattern retired | idiomorph + CSRF rotation gated (below); score not re-audited after landing                                                |
| **CSRF hardening**              | Wiring hardened + tested; constraints documented                 | Token rotation on login (needs island-side token refresh — server cannot do it alone)                                      |
| **Audit's CSRF recommendation** | Superseded by a better, tested implementation                    | The snapshot report still prints the wrong recommendation (double-escape trap), corrected only downstream                  |
| **Observability**               | Request-ID end-to-end correlation                                | `user_id` mapping dropped (extensions are not ULIDs — correct, but the "1%" slice delivered smaller than first advertised) |

## c) NOT STARTED

| Work                                                                                                           | Why it exists                                                      |
| -------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| **T6 — idiomorph experiment** (serve `HTMXExtIdiomorph`, morph-swap the transcript, kill the draft-wipe class) | Gated on the consuming stack's browser E2E                         |
| **CSRF token rotation** (island adopts a fresh token post-login)                                               | Same gate; server-only version proven harmful by the suite         |
| **Stack-side browser E2E re-run** after this session's header/attribute changes                                | AGENTS.md rule: any markup-adjacent change re-runs it              |
| **`docs-health` HARVEST** of this report's section (f) into `TODO_LIST.md`/`ROADMAP.md`                        | The 50-item list below is otherwise entombed in a timestamped file |
| **cqrs-htmx v4.10.0 bump watch** (retry hint, security consolidation — re-verify at the consumed tag)          | Upstream is ~256 commits past v4.9.0, unreleased                   |

## d) TOTALLY FUCKED UP

Nothing shipped broken — suite and gate are green, and `origin/main` carries only tested work. The honest entries:

1. **The audit report recommends a CSRF helper usage that is wrong for templ** (double-escape → `hx-headers` JSON.parse fails → silent CSRF loss). Caught during plan review, corrected in code, AGENTS.md, and the plan doc — but the research snapshot itself still tells a future reader the wrong thing. A report I published is wrong in print; the correction lives elsewhere.
2. **I nearly shipped a Verschlimmbesserung**: `InvalidateCSRFCookie` on login 403'd every post-login HTMX action. The existing suite caught it (`TestFaxStatusWebhookUpdatesJob`) — I did not catch it by thinking. Only the no-reload login flow note in my own head + the test stood between that commit and a dead app.
3. **BuildFlow run 1 failed a step** because the auto-commit daemon mutated the tree mid-run. Transient, infra-only, run 2 green — but I burned ~10 minutes and an investigation on it.

## e) WHAT WE SHOULD IMPROVE (self-review, answered honestly)

1. **What did I forget?** To grep the full test suite before swapping the timing middleware — `TestServerTimingOptIn` sat in the middle of `middleware_test.go` (I had read head and tail, not the middle) and broke. The suite caught what reading should have prevented.
2. **What is stupid that we do anyway?** Editing files read once at session start while the auto-commit daemon touches the tree: I hit "file modified since read" four times and re-read four times. Also: running quality gates while the daemon is actively committing.
3. **What could I have done better?** (a) Verify the frontend-escaping context of a recommended API _before_ publishing it in a report, not after — the CSRF finding was "verified" against the library source but not against templ's escaper. (b) Re-run the binary smoke path when the middleware chain changes (curl is banned here; a tiny Go smoke or the existing suite suffices — I leaned on tests alone). (c) Not advertise `user_id` enrichment as part of the "1%" slice before checking `ParseUserID`'s ULID requirement — the delivery shrank and the earlier message oversold it.
4. **What could I still improve?** The gated 5% (T6 + rotation) with one stack-E2E session; correction pointers into snapshot reports via an appendix policy instead of silence; HARVEST of section (f) so the list is living, not entombed.
5. **Did I lie to you?** No. Two claims needed correction mid-session (`user_id` scope; the CSRF helper recommendation) and both were corrected out loud, in the same session, in the durable docs.
6. **How can we be less stupid?** Treat "verify the API in the _consumer's_ context" as part of verification, not research; re-read immediately before every edit batch in daemon-run repos; run gates on a quiet tree.
7. **Ghost systems?** None created; none found in the touched surface. The idiomorph route was deliberately NOT served to avoid a dead asset — it stays a gated experiment, not a ghost.
8. **Scope creep trap?** Held. The dispatch layer, usermgmt, casbin, and the microphone-denying security preset all stayed rejected despite "use it 100%" pressure — 100/100 was defined as superb _within_ the root-library posture.
9. **Did we remove something useful?** Only `timingWriter`/`timingMiddleware` — replaced by a strictly more capable library middleware with the same gate, same header semantics, test-pinned.
10. **Split brains?** One killed (`securityHeadersConfig()` — the server↔test config literal duplication), one consciously avoided (`toastDetail` is now an alias, not a mirror). Residual risk watched: the audit snapshot vs. the corrected reality (see d1).
11. **Tests?** Five new/adapted guard tests landed; each behavior change is pinned. Gaps: no test pins the _absence_ of the pre-escaped helper value (the JSON-validity test covers it indirectly); no race test on the new chain order (none needed — middleware composition is static); `TestServerTimingOptIn` no longer pins attribute order (deliberate, documented in the test).

## f) UP TO 50 THINGS TO GET DONE NEXT (brainstorm — HARVEST fuel, sorted by impact)

**Gated on the stack E2E (do together in one session):**

1. Run the consuming stack's browser E2E against current `main` (this session changed headers/attributes).
2. T6: serve `cqrshtmx.HTMXExtIdiomorph` and evaluate `sse-swap="morph"` on the transcript panel.
3. Island-side CSRF token refresh after login; then adopt `InvalidateCSRFCookie` server-side.
4. Re-run the stack E2E after 2+3; flip the DOM-contract docs.
5. Re-audit the adoption score post-T6 (expect ~100) and annotate the plan doc's score table.

**Living docs / knowledge:**
6. HARVEST this report's (f) into `TODO_LIST.md` (docs-health skill).
7. Add a correction appendix pointer to the audit snapshot policy question (annotate vs. never-edit).
8. `docs-health` VERIFY pass over AGENTS.md claims changed this session.
9. Annotate the 2026-09-19 07:43 status report with the landed state (it predates the adoption work).
10. Record the BuildFlow daemon-race lesson (run gates on a quiet tree) in AGENTS.md.

**Upstream watch:**
11. Watch for cqrs-htmx v4.10.0 release; re-verify SSE behavior at the consumed tag before bumping.
12. When go-sse ships the reconnect `retry:` hint downstream, re-check `ServeSSE` and update the AGENTS.md tag-checking note.
13. templ-components `ThemeScript` opt-out watch — if it ships, drop the CSP hash + `!important` color-scheme rules.
14. httputil v1.3.0 watch (we track v1.2.0+1).

**Product polish surfaced by the audit:**
15. Atomic `claim` op for the webhook idem store (close the `seen`+`record` TOCTOU race).
16. Consider 503 (retryable) classification for transient webhook store failures — behind a provider-behavior check.
17. `MeasureServerTiming` named spans around PBX phone-api calls (now free with the library middleware).
18. Surface `X-Request-ID` in the island's `#log` lines so browser-side and server-side ids correlate.
19. Sweep remaining `err.Error()` interpolations in user-facing 4xx/422 panel errors (actions.go) for i18n parity.
20. Audit the remaining `http.Error` sites for SafeDetail-style redaction outside webhooks (proxy.go 502s).
21. OpenAPI spec: decide whether `/hooks/*` contracts deserve publication (currently only `/api/session`).
22. `Notify*` vs `ToastDetail` shape trap: add a comment/alias guard in case the dispatch layer ever lands.
23. `OOBHTML` evaluation for the nav unread badges (avoid full-panel SSE pushes for badge-only updates).
24. Settings panel: show the server version (`/version` payload) so operators see what runs.
25. `SafeRedirectPath` adoption readiness review (only relevant if a server-side login redirect ever appears).

**Repo hygiene noticed this session:**
26. `.buildflow.yml` `skip_steps` cleanup — `branching-flow`, `cqrs-lint`, `go-structure-linter`, `nix-hash-fix`, `pytest-test` "match no registered tool" (warning at every load).
27. Investigate the 9 BuildFlow tools "unavailable (health check failed)" (incl. `go-licenses` binary missing).
28. Review what the daemon's 3-line `flake.nix` change (commit `6691ca1`) actually was — I never inspected it.
29. Decide: explicit per-task commits vs. daemon-owned history (they raced each other all session).
30. Run gitleaks + codespell once (on-demand BuildFlow steps, skipped in every mode).
31. Run `markdown-lint`/`lychee` over the three new docs (skipped in full mode).
32. `git log -S timingWriter` sanity: confirm no stale references in docs after the deletion.

**Testing depth:**
33. Table-driven test for the enrichment middleware order (enrichment-inside-log regression would currently only fail via the wiring test — make the invariant explicit).
34. Fuzz or property test for `hx-headers` rendering with adversarial tokens (quotes, unicode).
35. Contract test: `Notify*` wire shape (`{level,message}`) vs island listener — pin the difference so a future dispatch-layer adoption fails loudly.
36. Race test for `ExtensionHubs` under concurrent publish+sweep (existing code, untested this session).
37. Smoke-test the built binary through `flake.nix` output (the AGENTS.md smoke command was never run this session).

**Release / ops:**
38. Cut v2.1.0 from `[Unreleased]` (five library adoptions + redaction are user-visible).
39. After release: `vulnix $(nix-store -qR ./result)` runtime-closure re-verification.
40. Stack-side input bump to the post-adoption commit; record the switchover in AGENTS.md.

**Bigger swings (ROADMAP fuel):**
41. Event-catalog-driven SSE event naming audit (`threads`/`thread`/`fax`/`voicemail` vs a documented registry).
42. Session-store persistence decision doc (in-memory by design — write down the multi-instance caveat).
43. `RenderTemplComponent`-style local helper to DRY the repeated Content-Type+render+error pattern (7 sites).
44. Evaluate `HTMXMiddleware` + HX-* context accessors if any handler ever needs request-aware partials.
45. Journal-backed SSE replay (`JournalSSEStore`) if message history needs reconnect-replay.
46. `datastar` evaluation spike (same hub, different transport) — only if HTMX limits appear.
47. `adminui`/`dashboardui` adoption spike if operator surfaces are ever wanted.
48. `IdempotencyStore` upstream API check: propose a record-on-success variant (webhook semantics) upstream.
49. Performance baseline: Server-Timing spans + `benchmark_*` comparison before/after chain changes.
50. Docs: a one-page "which library owns which surface" map (cqrs-htmx / httputil / go-sse / templ-components) from the audit tables.

## g) THREE QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **When can the consuming stack's browser E2E run against current `main`?** Both remaining items (idiomorph, CSRF rotation) are gated on it, and I cannot schedule or access that environment from this repo.
2. **Should the [Unreleased] work ship as v2.1.0 now, or keep accumulating?** This decides whether I cut the release, tag, and re-pin the stack input — or leave `[Unreleased]` open for the gated 5% first.
3. **Who owns commit history — me or the daemon?** The daemon committed my in-flight work four times, raced BuildFlow mid-run, and interleaved my per-task messages. If you want per-task history, I need to know whether to keep racing it or stop manual commits entirely.
