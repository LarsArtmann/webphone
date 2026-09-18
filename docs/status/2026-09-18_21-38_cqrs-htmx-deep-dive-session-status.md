# Status: cqrs-htmx Library Deep-Dive Session

**When:** 2026-09-18 21:38 CEST
**Session scope:** Library-deep-dive audit of `github.com/larsartmann/cqrs-htmx/v4 v4.9.0` usage in webphone; deliverable is an HTML research report. **Zero product code was changed** — this session read code and wrote one report.
**Deliverable:** `docs/research/2026-09-18_cqrs-htmx-deep-dive.html` (adoption score 62/100, 6 graded findings, 6-item action plan).

> **Format note (spec override):** the status-report skill's canonical format is a styled
> HTML dashboard; the user explicitly requested `.md`, so this report is Markdown and the
> override is flagged here per the skill's contract.

---

## a) FULLY DONE

Each item verifiably complete this session (evidence cited; no code changed, so no test/commit evidence applies — evidence is the artifact produced).

1. **Skill-compliant deep-dive process executed.** Loaded `library-deep-dive` SKILL.md +
   `research-methodology.md` + both HTML output guides + editorial template before working.
   Evidence: tool transcript; every phase ran in the prescribed order.
2. **Phase 1 — webphone usage inventory (complete for the import surface).** Found exactly
   **2 files, 4 symbols**: `HTMXScriptHandler()` (internal/server/server.go:101),
   `HTMXExtensionHandler("sse")` (server.go:102), `NewBroadcaster`/`Broadcaster` +
   `Hub().Subscribe/Unsubscribe`/`Broadcast` (internal/server/sse.go). Read the primary
   consumers in full: server.go, sse.go, pages.go, ratelimit.go, session_api.go,
   notifier.go; partial: actions.go (to line ~200), main.go, webhooks.go (head only).
   Evidence: file reads + `rg` symbol census (`cqrshtmx\.` → 6 hits).
3. **Phase 2 — full capability map of the consumed version.** `go doc -all` of v4.9.0
   (2,391 lines) captured; read in depth: `Broadcaster` (+`Hub()`/`Raw()` deprecation),
   `Config`, `HTMXRequest`, all ~40 `HandlerOption`s, `Response` builder, `StructuredError`,
   SSE deprecated aliases, readiness/health types. Cross-read implementations in the
   cqrs-htmx checkout (6cf46e62): `recovery.go` (ErrAbortHandler re-panic + stack logging),
   `sse_broadcaster.go ServeSSE` (retry hint + `connected` + 15 s heartbeat), and
   `httputil/ratelimit_keyed.go` (`MaxKeys` cap, computed Retry-After, ClientIP extractor).
4. **Version currency established.** go.mod pins `v4.9.0` which **is the latest root tag**
   (repo master `6cf46e62` = `loginpage/v4.10.0-256`; the v4.10.0 wave so far tags
   submodules only). Indirect `go-cqrs-lite` pins equal v4.9.0's own requirements.
   Verdict in report: nothing to bump.
5. **Phase 3+4 — evidence-based gap analysis, scoring, prioritization.** 6 findings graded
   (1 misuse, 2 missed-opportunity-critical, 2 partial, 1 fully-leveraged) + 6 verified
   deliberate non-adoptions (incl. the load-bearing nuance that the library's
   `RecommendedPermissionsPolicy` denies `microphone` and would break WebRTC calls).
   Weighted score 62/100 (assets 20/20, broadcaster 25/25, SSE lifecycle 7/15, recovery
   3/10, rate limit 4/10, readiness 3/10, request logging 0/10). Pareto table included.
6. **Phase 5 — report written and mechanically validated.** 1,398-line self-contained HTML
   at `docs/research/2026-09-18_cqrs-htmx-deep-dive.html`. Validation: HTML parser reports
   **zero unclosed/mismatched tags**; **8/8** sidebar anchors ↔ section ids; **zero**
   template placeholders left; `<title>` personalized.

## b) PARTIALLY DONE

1. **Consumer-file coverage is incomplete beyond the import surface.**
   - Works: all files that *import* cqrs-htmx read in full; hand-rolled counterparts of the
     6 findings read in full (recovery, keyed limiter, events loop, healthz, main).
   - Remains: `actions.go` tail (saveContact onward, contacts import/export),
     `webhooks.go` handler bodies (only the contract comment block was read), `panels.go`,
     `proxy.go`, `configjs.go`, `assets.go`, rest of `unread.go` — never opened.
   - Risk: an additional hand-rolled duplication (e.g. WriteJSON-equivalent, body-limit
     handling) could hide there and would make the audit *understate* gaps.
   - Blocker: none — out of the session's self-imposed "don't research unrelated" scope.
   - Effort to close: **S** (<30 min).
2. **Implementation claims cite the master checkout, not the consumed tag.**
   - Works: `recovery.go`/`ServeSSE`/rate-limiter behavior verified against repo master
     `6cf46e62`.
   - Remains: no `git diff v4.9.0..master` check on those files; if they changed since the
     tag, report claims could be stale (confidence high they didn't — unverified).
   - Blocker: none. Effort: **S** (`git show v4.9.0:sse_broadcaster.go` etc.).
3. **Symbol map ~70% deep-read.** Structural index of all exported symbols complete;
   ~30% of doc entries never deep-read (ack, notify, decoder, partial, redirect, security,
   openapi collector, event-catalog handlers). Blocker: none; marginal relevance to a
   consumer that imports 4 symbols. Effort: **S**.
4. **The audit's 6-item action plan is recommendation-only.** No middleware adopted, no
   branch, no tests run (nothing ran at all this session — nothing changed, nothing
   verified by execution). Effort: items 1–4 together ≈ **M** (1 h).
5. **Session knowledge not yet written into AGENTS.md.** The verified facts worth keeping
   (audit location + score; PermissionsPolicy-breaks-mic refusal; partial-by-URL is
   DOM-contract) are only in the report. Effort: **S**.
6. **Skill Phase 6 (cross-skill refs) only implicit** — deduplicate-code relevance is stated
   in prose; data-model-review correctly judged N.A. No explicit action taken. Effort: **S**.

## c) NOT STARTED

| Planned work | Why not started | Still wanted? |
|---|---|---|
| Adopt `cqrshtmx.RecoveryMiddleware` (delete `recovery()`) | Audit was report-only by request; awaiting go-ahead | Yes — top priority (S) |
| Add `cqrshtmx.RequestLoggingSlog` outermost | Same | Yes (S) |
| Replace `keyedLimiter` with `httputil.KeyedRateLimiter` ×2 + port tests | Same; also gated on XFF-trust decision (question 1) | Yes (M) |
| Honest `/healthz` via `ReadinessHandler` + SQLite ping | Same; probe semantics of the stack unknown (question 2) | Yes (S) |
| Collapse `events` loop onto `Broadcaster.ServeSSE` | Same; island stream-shape sensitivity unknown (question 3) | Yes (S) |
| `OOBHTML` unread-badge spike | Needs E2E loop; explicitly gated as future spike | Consider (M) |
| `docs-health` HARVEST of section (f) into TODO_LIST/ROADMAP | Report was just written; user said wait for instructions | Yes — must not skip (S) |
| AGENTS.md memory update (posture + audit pointer) | Listed in (b)5, not yet done | Yes (S) |

## d) TOTALLY FUCKED UP

**In the repo: nothing.** This session changed zero product code; build/tests were green by
non-interference (`git status` clean before, only untracked `docs/research/` after). The
DOM contract, tests, and binary are untouched.

**In this session's *output*, two honest defects** (wrong-or-at-risk artifacts, stated
without softening):

1. **The published report cites implementation evidence from the wrong tree.** Recovery/
   ServeSSE/rate-limiter behavior was verified against cqrs-htmx **master (6cf46e62)**,
   not the **v4.9.0 tag webphone actually consumes**. Severity: low (a stale claim would
   overstate/understate one finding, not break anything), but it is a correctness gap in a
   report whose footer claims "every claim verified against source." Mitigation: re-verify
   the three files via `git show v4.9.0:<path>`; effort S.
2. **The 62/100 score and "2 hand-rolled duplications" count are provisional.** Unread
   consumer files (webhooks.go bodies, actions.go tail, panels.go, proxy.go) could contain
   further duplications — meaning the report may understate the gap it measures. Severity:
   low; direction of error is "audit too generous," which for this audit's purpose (drive
   adoption) is self-correcting on follow-up. Mitigation: finish the file sweep (b1).

## e) WHAT WE SHOULD IMPROVE

1. **Verify implementations against the consumed tag, not a checkout.** This session read
   library internals from master while the consumer pins v4.9.0. Fix: a standing rule —
   when auditing dependency X at version V, read internals via `git show V:path` or the
   module cache. Prevents the entire defect class in (d)1.
2. **Finish the consumer sweep before publishing a count.** "2 duplications" was published
   with ~5 server files unread. Fix: inventory *all* files in the audited layer first,
   then write. Cheap, removes the asterisk from the headline number.
3. **Make the score reproducible.** The 62/100 rests on ad-hoc weights stated in prose.
   Fix: add a rubric table (capability, weight, earned, evidence) to the report appendix
   so a re-score after adoption lands on the same scale.
4. **Memory write-up lag.** Session-durable facts (audit pointer, mic/PermissionsPolicy
   refusal) stayed out of AGENTS.md. Fix: update project AGENTS.md at discovery time, per
   the global memory protocol, not at session end.
5. **Failed-edit round trip on the report title.** The splice created the file via bash;
   the first `edit` was rejected ("read the file first"). Fix: when a file is created
   outside write/edit, view it before the first edit — or do the whole assembly in one
   python pass including the title.
6. **No execution gate even for doc-only changes.** Nothing broke, but running nothing at
   all means the session's only verification was static. Acceptable here; for any future
   adoption work the gates are already named: `TestServedPageHoldsTheDomContract`,
   `buildflow`, `nix flake check`, stack browser E2E on markup changes.

## f) TOP 50 THINGS WE SHOULD GET DONE NEXT

Brainstorm per the skill's rule: items past the first ~10 are ROADMAP fuel; HARVEST
(`docs-health`) must apply routing rigor. Impact/Effort per the quality guide
(S <30 min, M 30 min–2 h, L >2 h).

**Tier 1 — adopt the audit's action plan (code, middleware-seam only):**

| # | Task | Impact | Effort | Category |
|---|---|---|---|---|
| 1 | Swap `recovery()` for `cqrshtmx.RecoveryMiddleware`; delete server.go:131-141 | High | S | Cleanup |
| 2 | Wrap chain outermost with `cqrshtmx.RequestLoggingSlog(slog.Default())`; assert no body/credential logging | High | S | Quality |
| 3 | Replace both `newKeyedLimiter` uses with `httputil.KeyedRateLimiterConfig` middleware; port ratelimit_test.go | High | M | Cleanup |
| 4 | `/healthz` → `cqrshtmx.ReadinessHandler` with SQLite ping (`db.Ping`) + blob-dir write check | High | S | Feature |
| 5 | Collapse `events` loop onto `Broadcaster.ServeSSE` (keep session gate + SetLang in front) | Medium | S | Cleanup |
| 6 | Decide rate-limit key: `KeyExtractorFromRemoteAddr` vs `FromClientIP` (needs stack XFF answer, question 1) | Medium | S | Decision |
| 7 | Run `GOEXPERIMENT=jsonv2 go test ./...` + `TestServedPageHoldsTheDomContract` after 1–5 | High | S | Quality |
| 8 | `buildflow` + `nix flake check` gate run | High | S | Quality |
| 9 | Update AGENTS.md posture + CHANGELOG entry for the adoption | Medium | S | Documentation |
| 10 | Confirm zero markup change → skip stack E2E; if any doubt, re-run `tests/browser-e2e.py` upstream | Medium | S | Quality |

**Tier 2 — close the audit's own gaps:**

| # | Task | Impact | Effort | Category |
|---|---|---|---|---|
| 11 | Re-verify recovery/ServeSSE/rate-limiter claims against `git show v4.9.0:<file>` | Medium | S | Quality |
| 12 | Read webhooks.go bodies, actions.go tail, panels.go, proxy.go, configjs.go; amend report if new duplication found | Medium | S | Quality |
| 13 | Deep-read remaining go doc sections (ack, notify, decoder, partial, redirect, security, openapi) | Low | S | Quality |
| 14 | Append score rubric table to the report (make 62 reproducible) | Low | S | Documentation |
| 15 | HARVEST section (f) into TODO_LIST.md / ROADMAP.md via docs-health | High | S | Documentation |

**Tier 3 — product-grade enhancements unlocked by the library:**

| # | Task | Impact | Effort | Category |
|---|---|---|---|---|
| 16 | OOBHTML spike: live nav unread-badge update inside SSE payloads (E2E-gated) | High | M | Feature |
| 17 | Use the `connected` SSE event to drive an island "live" indicator after #5 | Medium | S | Feature |
| 18 | Toast feedback via HX-Trigger + `cqrshtmx.ToastDetail` for send/save/error UX (island listener needed) | Medium | M | Feature |
| 19 | Rate-limit `/events` itself (unlimited reconnect churn currently unthrottled) | Medium | S | Quality |
| 20 | Per-extension hub teardown after long idle (hubs are never torn down today) | Low | M | Quality |
| 21 | Idempotency guard for `/hooks/*/status` (provider retries double-applying verdicts) | Medium | M | Feature |
| 22 | `/version` build-info endpoint (library DebugHandler pattern) for ops | Low | S | Feature |
| 23 | Server-Timing middleware behind a debug flag (library re-export) | Low | S | Quality |
| 24 | OpenAPI doc for `/api/session` (3 endpoints, optional) | Low | S | Documentation |
| 25 | Use `cqrshtmx.Chain` for the middleware stack composition (readability) | Low | S | Cleanup |

**Tier 4 — hardening / hygiene:**

| # | Task | Impact | Effort | Category |
|---|---|---|---|---|
| 26 | Fuzz `/hooks/*` JSON parsing (json/v2) against malformed payloads | Medium | M | Quality |
| 27 | Test hook-limiter-wraps-secret-gate ordering invariant explicitly | Medium | S | Quality |
| 28 | 429/Retry-After handling test for island fetch wrappers (`phone-api/*`) | Low | S | Quality |
| 29 | Session TTL sweeper interval vs `SessionTTL` config interaction test | Low | S | Quality |
| 30 | Verify `go vet`/lint clean on files touched by Tier 1 | Medium | S | Quality |
| 31 | Indirect-dep drift check on cqrs-htmx's transitives (local equivalent of check-version-drift) | Low | S | Cleanup |
| 32 | Track root `v4.10.0` tag; bump indirect go-cqrs-lite pins when it ships | Low | S | Cleanup |
| 33 | vulnix runtime-closure re-scan at next dep bump (per AGENTS.md method) | Low | S | Quality |

**Tier 5 — documentation / process / longer tail (ROADMAP fuel):**

| # | Task | Impact | Effort | Category |
|---|---|---|---|---|
| 34 | AGENTS.md: record "cqrs-htmx adoption posture" + link the audit + mic/PermissionsPolicy refusal fact | High | S | Documentation |
| 35 | Re-score adoption after Tier 1 (target ~90) and ANNOTATE the 2026-09-18 report (docs-health) | Medium | S | Documentation |
| 36 | Codify "verify dependency internals at the consumed tag" as a library-deep-dive checklist item | Medium | S | Documentation |
| 37 | Write per-item effort estimates (minutes) into the report's action table | Low | S | Documentation |
| 38 | Error envelopes (`StructuredError`) for `/api/session` — only if the island starts branching on codes | Low | S | Feature |
| 39 | Evaluate `sync/` multi-tab module vs island's per-tab SIP UA model (document the N.A. verdict) | Low | S | Documentation |
| 40 | Evaluate `DecodePagination` vs webphone's cursor-style `older=` param (likely N.A.; document) | Low | S | Documentation |
| 41 | Benchmark per-extension hub fan-out (library has broadcaster bench patterns) | Low | M | Quality |
| 42 | Emit badge push from unreadCache invalidation points (ties to #16) | Low | S | Feature |
| 43 | Document hx-boost-vs-island-swap non-adoption in views or AGENTS.md (prevents re-litigation) | Low | S | Documentation |
| 44 | Consider `DefaultLogFormatter` vs `JSONLogFormatter` for the stack's log sink | Low | S | Decision |
| 45 | Add readiness JSON body contract note for the stack's probes (depends on question 2) | Low | S | Documentation |
| 46 | Review `notify.go`/`ack.go` docs for anything the island could exploit later (close the symbol map) | Low | S | Quality |
| 47 | Cross-link the deep-dive from the structural-health HTML report (sibling audit hygiene) | Low | S | Documentation |
| 48 | If XFF is trusted upstream, contribute a ClientIP-trust note upstream (httputil docs) | Low | S | Documentation |
| 49 | Post-adoption: rerun `agentic_fetch`-style community check? No — local library; instead re-diff master for new middleware worth adopting | Low | S | Quality |
| 50 | Decide and record whether `/healthz` stays GET-only-open or moves behind session (security review of readiness detail exposure) | Low | S | Decision |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

What I tried first is stated per question; none is answerable from this repo alone.

1. **Does the consuming stack (nix-international-telephony) sanitize `X-Forwarded-For` /
   `X-Real-IP` before proxying to webphone?** Tried: AGENTS.md (says only "stack fronts
   TLS + WSS proxy"); ratelimit.go's own comment says every request shares the proxy
   socket. The answer decides adoption item 6: `KeyExtractorFromClientIP` (real
   per-client limits) vs `FromRemoteAddr` (proxy-wide buckets, status quo semantics).
2. **How does the stack consume `/healthz` — hard routing gate (k8s-style readiness) or
   informational?** Tried: server.go's healthz is a bare constant `ok`; AGENTS.md doesn't
   name a prober. This decides whether item 4 should fail *closed* (503 removes the
   instance from rotation — desired?) or stay advisory, and whether the JSON check names
   in a readiness body would break the stack's probe parser.
3. **May the `/events` SSE byte stream change at all (a `retry:` hint + an initial
   `connected` event), or is the stream shape frozen by the browser E2E?** Tried: AGENTS.md
   pins island element ids and greppable strings, but says nothing about SSE framing; the
   E2E lives in the sibling repo, which is outside this session's "no unrelated research"
   scope. This gates adoption item 5.

---

**Handoff:** section (f) is the input for a `docs-health` **HARVEST** run (TODO_LIST.md /
ROADMAP.md) — not yet executed; the session was instructed to stop after reporting.
The session's deliverable report is untracked (`docs/research/`); per the harness
no-commit rule it was not committed manually — the auto-commit daemon picks it up.
