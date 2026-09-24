# Status Report — SUPERB Stack-Adoption Execution Run (session 2026-09-24 ~13:40–18:03)

**Report time:** 2026-09-24 18:03 CEST
**Session scope:** Full execution mode of the pareto plan
`docs/planning/2026-09-24_13-33_SUPERB-stack-adoption-pareto.md` (22 medium tasks
M1–M22). This report covers exactly what this session did, broke, and left open.
**Local head:** `f5aa3c6`; **remote at report time:** `94ae28d` — the last ~8
commits (M7-verdict-doc through M14) are LOCAL-ONLY; the daemon pushes
periodically but had not caught up when this was written. Verify with
`git ls-remote` before building anything off main elsewhere.

---

## a) FULLY DONE (verified green at time of landing)

| Task | What landed | Verification | Commit(s) |
| --- | --- | --- | --- |
| Baseline repair | `go 1.27.1` directive restored (daemon had reverted it) | build green | daemon-swept |
| **M2 — upload body bound** | `http.MaxBytesReader` (61MB envelope const `uploadBodyLimit`) in `requireSessionMultipart`; `TestUploadBodyIsBounded` (oversize 400 "too large", boundary-at-limit still parses); smoke probe "oversized upload rejected" (early-response-safe chunked sender); error-contract row | server suite + live smoke 41 checks + 4 restart, 0 failed | `55cb19f` (+daemon `4c8e379`) |
| **M5 — dependency train sweep** | cqrs-htmx v4.12.0, httputil v1.3.0, go-error-family v0.10.2, go-sse v0.6.1 + 7 indirect go-cqrs-lite modules to v4.11.x; go-health HELD at v0.3.0 (alpha policy); vendorHash roundtrip same-breath (`Z9dCSY…`) | full `-count=1 ./...` green; `nix build .#webphone` green (doCheck inside); buildflow re-run failed ONLY on treefmt (my new CSS) which was then formatted | `f04b51a`+`94ae28d` (daemon-swept) |
| **M6 — SafeDetail redaction consistency** | One-home 500 writers `internalError` (panels.go) + `safeDetail` (pages.go): full detail → operator log, family default → client; wired at 4 sites (2×"load conversation", 2×degraded-page panel); sub-500 writers deliberately untouched (byte-stable webhook texts, safe 4xx detail); `TestInternalErrorRedactsDetail`; error-contract row | server suite green | `a74f8e6` |
| **M7 — Tailwind coexistence spike** | Session-gated throwaway route `/dev/spike/tailwind` (spike.templ/spike.go) + `/assets/spike/probe.js` + Tailwind v4 build; A/B computed-style probe via headless Chromium 153 through a cookie-injecting loopback proxy. **VERDICT GREEN**: all 5 probed webphone surfaces byte-identical under the Tailwind load, all library components styled, zero inline scripts/styles. Verdict doc with data + build recipe: `docs/planning/2026-09-24_16-38_tailwind-coexistence-verdict.md` | CSP + DOM-contract tests green; chromium dump-dom diff | `681cf6a` (+daemon) |
| **M8 — Retry-After on 503s** | `retryAfterHint` ("1" = cqrs-htmx DefaultRetryAfter) sent by the fail-closed hooks gate + the unconfigured phone-api proxy; startupz deliberately excluded (probe consumers ignore headers); `TestFailClosedHooksCarryRetryAfter` | server suite green | `09a5afa` + `1e1dc20` |
| **M9 — templ wave 1 (EmptyState)** | 6 true empty-state sites → `display.EmptyState` (Search/Inbox/Document/Phone/Clock/Users icons; existing i18n keys reused as titles); permanent `/assets/tw.css` (14.5KB scoped Tailwind build from the adopted components' v1.19.2 sources) linked in shell head; 3 wp-empty occurrences deliberately kept (`<code>` config guidance ×2, settings theme note — not empty states) | full suite green; smoke 41+4 green; render check (all swapped panels carry EmptyState markup; tw.css serves 200) | `f1a2853` (+daemon `f70cfa4`) |
| **M10/M11 — waves 2+3 dispositions** | Both REJECTED with written rationale (in the verdict doc + TODO row): RelativeTime (browser-locale strings break the per-extension language invariant; byte-stable timestamp pins deliberate; its live-update script dead under CSP), CountBadge (wrong shape for the standalone `wp-nav-badge` pill), errorpage module (404 is shell-integrated, `.wp-error` is an htmx responseHandling wire contract, family copy already lives server-side). This is deliberate-non-use treatment, not drift | documented; no dep added | `357cd60` |
| **M12 — errorfamily.LogErrorContext** | 3 sites adopted (internalError, safeDetail, gateway-failure arm): message carries op prefix via `%w` wrap, attrs from the library (family/code/retryable/exit_code), transient→Warn per library semantics; provider-rejection arm kept hand-rolled (extra `status` attr the helper can't carry); `TestInternalErrorLogsFamilyFields` (new file — avoided the contested helpers_contract_test.go) | server suite green | `88cf0ef` |
| **M13 — errorfamilytest asserts** | fax/messaging/gateway family tests → `AssertFamily`/`AssertCode`; wrap-preservation + type-visibility checks kept hand-rolled (compare two errors); classify_test.go untouched (pins webphone's own mapper); test-only dep added; **vendorHash roundtrip #2** same-breath (`pP91z8…`) | 3 packages green; nix build green | `0bd6576` (+daemon `4d801cf`) |
| **M14 — httpspec conformance** | `TestHTTPSpectChainConformance` runs httputil's 19-spec suite against the full New() chain; `testServer` gained a `handler` field. **19/19 PASS on first run, zero skips** — no divergences to document; middleware_test asserts kept (they pin the panic lane httpspec doesn't cover) | `-run` verbose all-PASS + package green | `e9f6310` |
| TODO_LIST upkeep (docs-health style) | 3 done rows deleted (never struck through), Tailwind row rewritten twice to final done-state with remaining cleanup named | — | with each wave commit |

Housekeeping facts worth keeping: two vendorHash roundtrips performed this
session; `nixpkgs#tailwindcss` is v3 (DOES NOT WORK for the library — v4 syntax);
`nixpkgs#tailwindcss_4` (4.3.3) is the build pin; chromium
`agamzss…-153.0.8010.52` in the store works headless with `--no-sandbox`.

---

## b) PARTIALLY DONE

- **M15 — send-failure remainder (C + fax-lane guard + F): RESEARCHED AND
  DESIGNED, ZERO CODE WRITTEN.** Design reached (see §e/§f for the shape): a
  pre-emptive provider-rejection fast path INSIDE both services — persist the
  queued row, then mark it failed locally with kind `rejected` and return
  `*gateway.ErrProviderRejected{Status:400, Detail:"…your own number…"}` so the
  EXISTING train-E 422 arm renders it and the failed-row evidence survives.
  This resolves the plan's owner dichotomy ("instant refusal vs
  evidence-preserving row") by delivering BOTH. Needs: identities map injection
  into `messaging.New`/`fax.New` (or a resolver param), guard between
  append-queued and gateway-call in both services, English service-level detail
  text (no i18n churn — the 422 arm wraps it), service tests (row FAILED +
  Rejection family + gateway not called), server test (422 + reason + failed
  bubble, config `identities` via `newTestServerWithConfig`). **F stays
  conditionally deferred by the plan's own rule** (only if self-sends recur
  post-B+C).
- **Gates: not yet re-run as a FULL set after M12–M14.** Package-level green
  everywhere, but the cycle needs one final: full `go test -count=1 ./...`,
  `buildflow` (full, no result cache — the last full run predates waves 1–M14),
  `nix flake check` (island-lint/oxlint over the new probe.js? it lives outside
  island/, but the check records the scanned list — verify), live smoke, and
  `nix build` — all scheduled for the M19 release gate.
- **Push state:** everything through `94ae28d` is on origin; the session's last
  ~8 commits were local-only at report time (daemon lag). Not verified end-state.

## c) NOT STARTED

- **M18** — go-health `WithEvaluationHook` → /metrics counters. UNVERIFIED
  ASSUMPTION IN THE PLAN: the hook may only exist in go-health v0.4.0, which is
  HELD as alpha. First step is checking the vendored v0.3.0 API; if absent, M18
  parks with the hold.
- **M19** — v2.7.0 release train (fold → bump → gates → tag → stack bump →
  browser E2E ×2 → aarch64 ELF verify → pbx-artmann relock #5). Includes a
  PRE-REQ this session created: remove the throwaway spike route + probe.js
  before the fold (tracked in TODO row).
- **M20** — docs/registry sync: AGENTS.md templ-components adoption table +
  dedup registry rows for internalError/safeDetail, error-contract is current
  but FEATURES.md/CHANGELOG.md need the cycle's entries (EmptyState wave,
  Retry-After, upload cap, SafeDetail redaction, dep train, httpspec).
- **M21** — branded-id Valuer/Scanner seam comment (5 min).
- **M22** — standing watches upkeep (calendar note; rows re-checked 09-23, next
  erraudit 2026-10-22 — nothing due).
- **Owner-terminal tasks (cannot execute, pbx-artmann AGENTS forbids assistant
  ssh/deploy):** M1 (deploy v2.6.0 chain to prod — STILL serves v2.5.0), M3
  (post-deploy verify + rejection-banner self-send), M4 (SMS bridge root cause
  via journalctl — prod outbound SMS broken since 2026-09-19), M16 (owner-calls
  batch sitting ~20 decisions), M17 (post 5 announcements). A consolidated
  owner command sheet was planned but not yet written.

## d) TOTALLY FUCKED UP (self-inflicted, all caught + fixed, recorded honestly)

1. **Smoke-script method orphaning (M2):** my multiedit inserted
   `oversized_upload` between `sse_events` and `wait_for` inside the Smoke
   class; `wait_for` became a nested (dead) def → AttributeError mid-smoke.
   Fixed by relocating. Root cause: anchoring an insert on a small
   `finally: conn.close()` snippet that appeared in a method I didn't intend to
   split.
2. **Committed with a failing suite (M8):** `go test … | tail -2 && git commit`
   — the pipe masked go's exit code; commit `09a5afa` landed while
   `TestFailClosedHooksCarryRetryAfter` failed (header missing — see next).
   Caught immediately after, fixed in `1e1dc20`. Lesson: never gate a commit
   behind a piped test; use the raw exit or `set -o pipefail`.
3. **Lost edit in a redo (M8):** my first secretGate edit was rejected
   ("read the file first"); the re-apply multiedit carried only the const and
   silently dropped the header-setting line. The test caught it (this is why
   the micro-test existed — the pin did its job).
4. **tw.css 404 in the running binary (M9):** moved the file to
   `internal/web/assets/tw.css` but the `//go:embed` pattern still listed only
   the old explicit files — the shell linked a 404. Caught by my own render
   check (good), but the first cut shipped the link broken (bad).
5. **Commit-message erosion from daemon races:** repeatedly the auto-commit
   daemon swept the bulk of a unit as "chore: auto-commit N files" seconds
   before my explicit commit, leaving my detailed message pinned to a 1-file
   remnant (M2, M5, M8-bulk, M9, M13). Content is all on main and narrative
   lives in the verdict/plan docs, but several commits lost their story. Also
   one daemon sweep (`f04b51a`) bundled a CONCURRENT session's in-flight
   templ/panels_test.go work together with my dep bumps.
6. Minor: used `rg -r` (replace flag) accidentally and misread garbled output;
   `PIPESTATUS` quirks in this shell printed misleading exit codes twice —
   resolved by re-running raw both times.

## e) WHAT WE SHOULD IMPROVE (process, from this run)

- **Never gate commits behind piped commands** — pipe-masking bit me once
  materially (d.2). Consider `set -o pipefail` habitually.
- **Multiedit anchors:** stop anchoring inserts on short repeated snippets
  (`finally:` blocks, closing braces); anchor on unique function signatures or
  distinctive comment lines. Two of the four fumbles (d.1, d.3) share this root
  cause.
- **Daemon race protocol:** `git add` + commit IMMEDIATELY after the last edit
   of a unit (minimize the window), or accept daemon commits and put the
  narrative in the plan/verdict docs (current de-facto mitigation). A
  pre-commit narration file (e.g. `docs/commit-notes/`) could preserve stories
  the daemon eats.
- **go:embed discipline:** any new asset file needs the embed pattern, the
  route, AND the link — three places, one checklist (the spike caught it; a
  permanent asset test asserting every `<link>` in the shell resolves 200 would
  pin it forever — worth adding).
- **Exit-code hygiene in this shell (mvdan/sh):** PIPESTATUS is unreliable;
   prefer separate commands or `if ! go test …; then`.
- **Spike cleanup debt:** the spike route + probe.js are still in the tree by
  design (needed if waves continued) — do not let them ride a release (tracked
  in TODO + M19 pre-req).

## f) NEXT — up to 50 things, ordered

**Finish the current cycle (this session's lane):**
1. M15: verify `ErrProviderRejected` renders via the 422 arm for a
   service-returned instance (read actions.go arm + gateway/webhook.go:36-60).
2. M15: inject the identities map (or extension→DID resolver) into
   `messaging.New` + `fax.New`; update wiring + all constructor callers/tests.
3. M15: implement the self-send guard in `messaging.Service.Send` (between
   AppendMessage and gateway call: mark failed kind `rejected`, notify, return
   `&gateway.ErrProviderRejected{Status:400, Detail:"messages to your own
   number cannot be sent"}`).
4. M15: same in `fax.Service.Send` (fax wording).
5. M15: service tests ×2 (row FAILED, Rejection family via errorfamilytest,
   gateway untouched).
6. M15: server test: identities config → POST /messages/send to own DID → 422
   with the reason + failed row in the thread; same for /fax/send.
7. M15: error-contract row + send-failure plan doc status update; delete TODO
   row or mark C done.
8. Full-suite run: `nix develop -c go test -count=1 ./...` (raw exit).
9. `templ generate` check (no .templ touched in M15 — confirm clean).
10. Live smoke re-run (self-send 422 could deserve a probe if cheap).
11. Final `buildflow` full (BUILDFLOW_NO_RESULT_CACHE=1) — the cycle gate.
12. `nix flake check` (KVM-gated backup VM included) — first run this session.
13. `nix run nixpkgs#nodejs -- --test … island-tests/*.test.mjs` (island
    untouched this session, but the gate is cheap and the cycle needs it).
14. Push verification: `git push` if daemon lagged; `git ls-remote` end-state.

**M18 lane:**
15. Check go-health v0.3.0 (vendored source in module cache) for
    `WithEvaluationHook` existence.
16. If absent → park M18 with a TODO note tied to the alpha hold (v0.4.0).
17. If present → wire hook counters into /metrics mirroring
    `webphone_crm_lookups_total` rendering + test.

**M21/M22 quick wins:**
18. M21: seam comment at store ID call sites ("adopt id.Valuer/Scanner on next
    storage-format touch").
19. M22: verify watches rows fresh; add calendar note for 2026-10-22 erraudit
    re-measure (the tier-2 count may have MOVED with this session's new error
    paths — worth an early re-measure note).
20. erraudit tier-1 quick check over the session's new files (actions.go,
    panels.go, pages.go changes) — bar is `--type-aware` exit 0.

**M20 docs sync:**
21. AGENTS.md: templ-components adoption table row for EmptyState + the
    rejected waves' rationale pointer.
22. AGENTS.md: dedup registry rows for internalError/safeDetail + the
    retryAfterHint const.
23. AGENTS.md: commands section — tailwind build recipe pointer + the
    tailwindcss_4-vs-v3 trap.
24. AGENTS.md: CSP section — the tw.css layering fact (unlayered app.css beats
    all layers; verified empirically).
25. FEATURES.md: wave-1 EmptyState, upload cap, Retry-After, 500 redaction,
    httpspec, dep train — status updates.
26. CHANGELOG.md: unreleased section for the whole cycle.
27. docs/dom-contract.md: verify no id drift from wave 1 (EmptyState has no
    pinned ids — confirm the contract test still parses).
28. Stack runbook § webphone error contract: sync the new rows (upload 400,
    500 redaction, Retry-After) — cross-repo one-home rule.

**M19 release train (after the above):**
29. Remove spike route + `internal/server/spike.go` + `views/spike.templ(._templ.go)`
    + `assets/spike/` + embed `all:spike` + asset routes (keep `/assets/tw.css`!).
30. Re-verify suite + smoke after spike removal.
31. Release fold: daemon sweep check, clean tree.
32. CHANGELOG cut + version bump to 2.7.0 (flake.nix `webphoneVersion` +
    any version refs).
33. Gates: buildflow, vulnix (`nix run .#vulnix`), `nix flake check`.
34. Tag + push + gh release (git ls-remote verify; ls-remote tag check).
35. Stack train bump (`nix-international-telephony`): input lock to the new
    webphone train, `nix flake check` there.
36. Stack browser E2E ×2 (mandatory: markup changed this cycle — EmptyState
    wave; budget 445s, known ~90s transfer flake — re-run once before digging).
37. aarch64 cross-build + ELF byte verification.
38. pbx-artmann relock #5 + re-pin (stack tree MUST be clean first; narHash
    covers the whole tree).
39. Post-release: probe + smoke `--expect-version 2.7.0`.

**Owner lane (prepare, cannot execute):**
40. Consolidated owner command sheet: deploy v2.6.0 chain (M1), post-deploy
    verify (M3), SMS bridge journal ritual (M4).
41. Owner-calls briefing refresh (M16): add THIS session's new decisions to
    ratify (wave 2+3 rejections, upload cap 61MB, Retry-After: 1, httpspec
    adoption, M15 pre-emptive-rejection design if green-lit).
42. Announcements check (M17): drafts exist for v2.1.0–v2.6.0; a v2.7.0 draft
    will be needed after M19.

**Hardening / debt noticed en route (optional, park if unwanted):**
43. Permanent-asset link test: every `<link href>` in the served shell returns
    200 (would have caught d.4 automatically).
44. Consider asserting in TestServedPageSatisfiesStrictCSP that /assets/tw.css
    carries no `@import`/url() to off-origin resources (it's a build artifact —
    cheap pin).
45. tw.css regeneration is MANUAL — add the regeneration trigger to the
    templ-components watch row explicitly (done in verdict doc; mirror in AGENTS
    watches).
46. The smoke count narrative: main suite is now 41 checks INCLUDING the new
    upload probe (the historical "41+4" was pre-probe with one consolidated
    check elsewhere) — update AGENTS' historical note at next edit.
47. errorfamilytest now a direct dep — watch it on the next error-family bump
    (same train discipline).
48. Concurrent-session coordination: helpers_contract_test.go and views/
    panels_test.go were being edited by another session mid-flight; I routed
    around them (new test files). Their in-flight work rode daemon commits —
    nothing broken, but their session should verify its own landing.
49. Consider `gopls` diagnostics distrust note is already in AGENTS — this
    session re-confirmed it (~150 stale errors while builds were green).
50. Spike leftovers hygiene: after removal, grep for "spike" in routes/assets
    to ensure zero references (probe.js, tw.css old path).

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **M15 design green-light?** The plan marks C's semantics as an OWNER call
   ("instant refusal vs evidence-preserving failed row"). My design delivers
   BOTH (persist the failed row locally, then answer 422 with the reason via
   the existing rejection arm — no provider roundtrip). Implement as designed,
   or hold M15 for the owner-calls sitting?
2. **M18 if the hook is alpha-only:** if go-health v0.3.0 lacks
   `WithEvaluationHook` and only v0.4.0 (held as alpha) has it — park M18 with
   the hold, or do you accept taking the alpha for this one feature?
3. **M19 ordering vs M1 deploy:** should v2.7.0 (containing this whole cycle)
   go through the full release train NOW and queue BEHIND the still-pending
   v2.6.0 prod deploy, or would you rather deploy v2.6.0 first (M1/M3/M4 are
   yours) and cut v2.7.0 after — i.e., is there any pressure to fold v2.7.0
   before the prod deploy happens?

---

**Session stopped here per instruction. Waiting for directions.**
