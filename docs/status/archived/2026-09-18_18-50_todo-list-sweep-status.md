# Status Report: WebPhone TODO_LIST Sweep (23 items, High → Low)

**Date:** 2026-09-18 18:50
**Scope:** This session only. One mandate: work the TODO_LIST top-down —
READ, UNDERSTAND, RESEARCH, REFLECT, then execute and verify one task at
a time. 15 of the 23 TODO_LIST items are DONE and verified; the
remaining 8 are scoped, started, or deliberately deferred (details in
(b)/(c)). The open "deployment ownership" question (old report g1) is
RESOLVED: this repo now ships `nixosModules.default`.

> Format note: the status-report skill's canonical format is HTML; the
> user explicitly requested Markdown at `docs/status/<ts>.md` — honored.

**Verification state at report time:** `go build ./...` ✅ ·
`go test ./...` 9/9 packages ✅ (arch, config, domain, gateway, pbx,
server, store, vcard, views) · island JS `node --check` ✅ · templ
regenerated ✅. NOT yet run: `buildflow` full gate, `nix flake check`,
`go mod tidy` (see b1), upstream browser E2E (required — this session
changed markup).

---

## a) FULLY DONE ✅

### Batch 1 — High priority

1. **Island invariant drift FIXED (was TODO row 5, High/S):** the live
   `SIP.UserAgent` now lives in `state.js` (`state.userAgent`), written
   only by `connection.js`; `calls.js` and `main.js` read it from state —
   `calls.js` no longer imports `connection.js`. Verified later by the
   new arch test (see #15) — which then caught a SECOND violation
   (calls→ice, see d5) that my grep-driven fix had missed; fixed via a
   `wp:calls-changed` CustomEvent + listener in `ice.js`, so calls/ice/
   connection are now pairwise independent.
2. **Config loader unit-tested (High/S):** `internal/config/config_test.go`
   — defaults, env `__` nesting (`WEBPHONE_GATEWAY__WEBHOOK_URL` proves
   the leaf-with-underscore case), JSON file incl. lists + `"12h30m"`
   duration, env-beats-file, three invalid-config cases, and an
   `envKeyToPath` table.
3. **Gateway webhook-mode outbound tested (High/M):**
   `internal/gateway/webhook_test.go` — multipart fields, attachment and
   fax-PDF parts, `Bearer` secret header (and its absence without a
   configured secret), `provider_ref` receipt decode, 5xx and malformed-
   receipt errors, missing-PDF error. **Contract fix included:** the code
   now accepts the bare-token receipt its own doc comment always
   promised (any small non-JSON, non-`{` answer ≤ 256 bytes).
4. **Login rate limiting (Medium/S):** `internal/server/ratelimit.go` —
   per-client keyed token buckets (`golang.org/x/time/rate`, pruning
   sweep). `POST /api/session`: 1 token/2 s, burst 5. `/hooks/*`: 1/s,
   burst 60, wired OUTSIDE `secretGate` so secret brute-force is also
   throttled. 429 + `Retry-After`. Client key is the direct peer
   (documented: behind the stack proxy the shared secret remains the
   real boundary). Tests: login and hooks both trip, fresh server per
   test isolates buckets.
5. **Deployment ownership DECIDED + NixOS module SHIPPED (High/M,
   resolves old report g1):** decision — **this repo ships the module**
   (`package/nixos-module.nix`, `flake.nixosModules.default`) so binary
   and deployment shape stay in sync; the consuming telephony stack
   keeps BOTH options (import the module, or keep its own reverse proxy
   = the README default). Reversible by design. Module: hardened systemd
   unit (dedicated user, StateDirectory, Protect*/Restrict* set, no
   capabilities, restart-on-failure), freeform settings →
   `webphone-config.json` via `WEBPHONE_CONFIG`, `environmentFile` for
   secrets (`WEBPHONE_GATEWAY__WEBHOOK_SECRET`), optional nginx vhost
   with WSS upgrade on `websocket_path`. Flake check `webphone-module`
   evaluates the module and builds its generated config + derived listen
   port. **Stale v1 `package/default.nix` (static site from the
   nonexistent `../src`) removed** via `git rm`.

### Batch 2 — Medium/Low code items

6. **countUnread cached (Low/S):** `internal/server/unread.go` — 5 s
   TTL per extension, invalidated at every unread mutation (inbound
   webhook, open-thread mark-read), so badges stay correct without
   rescanning all threads per shell render. Regression test proves the
   webhook path surfaces on the NEXT render after a cached 0.
7. **Fax page-count parsing (Low/S):** `flexPages` (json/v2 custom
   unmarshaler) accepts number, numeric string, or null; aliases
   `pages`/`page_count`/`num_pages` (first non-zero wins) on BOTH
   `/hooks/fax` and `/hooks/fax/status`. Garbage counts are a clean 400,
   never a panic. Table-tested with an outbound job seeded end-to-end.
8. **pbx.Client error paths tested (Medium/M) — caught a REAL
   production bug:** `JoinPath` percent-encoded the query separator, so
   the history call went upstream as `/phone-api/history%3Flimit=30`.
   Fixed in `client.do` (query split off and re-attached). Test coverage:
   disabled client (`ErrDisabled` everywhere), invalid base URL, exact
   Basic-auth header, all voicemail/history paths, 401/5xx/malformed
   responses, context-timeout wrap, `ResolvePath`.
9. **Manual theme toggle (Low/S):** header button cycles
   auto → light → dark; persisted in `localStorage("wp-theme")`; applied
   as `data-theme` on `<html>`. CSS override selectors
   (`:root[data-theme=…]`) out-specify BOTH stylesheets' media queries,
   so the island tokens follow too. Pre-login usable.
10. **History search/filter (Low/M):** `?q=` substring over numbers and
    names (case-insensitive) + `?dir=in|out`; upstream window widens to
    100 when filtering, display still capped at 30; filter form swaps
    the panel via HTMX and echoes active values. Stub-upstream test
    covers all filter axes.
11. **Keyboard shortcuts + media keys (Low/M):** new island module
    `shortcuts.js`: A answer, H hangup, M mute, P hold, Esc
    reject/hangup, MediaPlayPause (answer-or-mute), MediaStop (hangup);
    inert while typing in any field, ignores modifier combos. The
    underlying actions were extracted as exported functions in
    `calls.js`, and main.js's accept/reject handlers now reuse them
    (deduplication instead of parallel logic).
12. **vCard import/export (Low/M):** new `internal/vcard` package —
    tolerant decoder (RFC folded lines, quoted params, N-field fallback,
    escapes, `tel:` URI scheme) + vCard 3.0 encoder with round-trip
    test. `POST /contacts/import` (multipart, upsert-by-number so
    re-import renames instead of duplicating, digit-bearing guard so
    letter-only garbage can never dial, invalid rows skipped with an
    honest message), `GET /contacts/export` (`text/vcard` attachment,
    session-gated). UI: import form + export link on the Contacts tab.
13. **Message pagination (Low/M):** `store.ListMessagesPage`
    (page/offset, `hasMore` via LIMIT+1), `messaging.ThreadWindow`, and
    a `?older=N` param on the thread partial with a "Load older
    messages" button ABOVE the sse-swap region (a live push resets to
    the newest window — the tradeoff is documented in the view comment).
    Tested with a 205-message thread across three page states.
14. **German translations for the server tabs (Medium/M):**
    `views/i18n.go` — ~90-key en/de dictionary, `T()` with English
    fallback and key-surfacing on typos, plus a dictionary-sync test.
    `Lang` flows through every props struct; **SSE fragments follow the
    per-extension language** remembered on `ExtensionHubs` (the notifier
    renders without a request — hub language is refreshed on every shell
    render and SSE connect). The island's EN/DE switch now also writes
    the `wp-lang` cookie and re-fetches the active tab partial via
    `htmx.ajax` — no reload, the island (and any call) survives.
    Accept-Language `de*` is the no-cookie fallback; English stays the
    default. Handler-composed panel errors are translated too.
    Documented limitation: service-layer validation reasons
    (`ErrInvalidSend`/`ErrInvalidFax`) remain English by design.
15. **Import direction machine-enforced (Medium/M):**
    `internal/arch/arch_test.go` — (i) `internal/domain` imports nothing
    internal; (ii) blob/config/fax/gateway/messaging/pbx/session/store/
    vcard never import server or web; (iii) the island module graph:
    calls/ice/connection pairwise independent (parses the real `import`
    statements of the served JS). These run in the ordinary test suite,
    so every gate now fails on violations — no linter-config fight.

---

## b) PARTIALLY DONE ⚠️

1. ~~**`go.mod` tidiness:** `golang.org/x/time` is required and builds,
   but gopls still flags "should be direct" — `go mod tidy` /
   `buildflow -s gomod-check --fix` has not been run yet. One command
   away.~~ done (06:42 #1: already direct at go.mod:18; the gopls warning was stale)
2. ~~**server_test.go split (Medium/M):** achieved organically for NEW
   tests (webhooks, history, contacts, rate-limit, pagination, i18n
   live in their own files), but the original monolith still holds the
   sse/proxy/session/fax/contract tests, and the jscpd-flagged 14-line
   helper clone is NOT deduped.~~ done (06:42 #5: split into session/messages/fax/sse/proxy test files; twin helpers deduped)
3. ~~**NixOS module verification:** flake check exists but `nix flake
   check` itself has not been run this session; the module is
   eval-checked, not VM-tested (systemd unit never ran under a real
   NixOS activation); new Nix files not yet passed through nixfmt.~~ done (flake check green 06:42 #11; stack NixOS VM test green 06:42 #15)
4. ~~**Docs for this session:** README/FEATURES/CHANGELOG/AGENTS/
   TODO_LIST do NOT yet reflect this session's work~~ done (06:42 #4 docs overhaul)
5. ~~**Session persistence (Low/M):** decision MADE (keep in-memory by
   design — persistence trades in the password-at-rest posture; route
   to FEATURES as WORTH_CONSIDERING), but the TODO_LIST/FEATURES edit
   is not written yet.~~ done (FEATURES WORTH_CONSIDERING row)
6. ~~**Upstream browser E2E:** this session changed markup (header
   actions + theme button, nav labels de, history filter form, contacts
   import/export UI, older-messages button, transcript fragments) — the
   E2E lives in the telephony stack and has NOT been re-run (see c1).~~ done at `00f13fe` (green 06:42 #14); island changed again after — re-run pending (TODO_LIST)

## c) NOT STARTED ❌

Remaining TODO_LIST items, untouched by design or blocked on (b):

1. ~~**Upstream switchover in nix-international-telephony** (High/M):
   repo exists locally at `~/projects/nix-international-telephony`, not
   yet assessed this session; needs input swap, WSS proxy config,
   config.js → server config migration, and the browser E2E re-run.~~ done (executed by a parallel session; E2E green after `00f13fe` — ROADMAP)
2. ~~**Tag v2.0.0** (Medium/S): tag not created; needs the right commit
   picked and (for the CHANGELOG links to resolve) a push.~~ done at `d9d6d03` (pushed; lychee 0 errors)
3. ~~**`-coverpkg` union coverage** in the buildflow test-coverage step
   (Medium/S) and **cqrs-lint documented skip** (Low/S): both are
   `.buildflow.yml` edits; not started.~~ cqrs-lint skip: done (06:42 #2); -coverpkg: **Won't implement —** upstream-blocked → ROADMAP
4. ~~**Nixpkgs channel bump** to clear the 14 build-chain CVEs
   (Medium/S): `nix flake update` + rebuild not run.~~ done (06:42 #7); glibc story corrected `7ba25cc`
5. ~~**sip.js 0.22 evaluation** (Low/M): research + report not started;
   no bump attempted (E2E re-run is a hard precondition).~~ done (evaluation report — stay on 0.21.2)
6. ~~**Final verification pass:** `BUILDFLOW_NO_RESULT_CACHE=1 buildflow`,
   `nix flake check`, aarch64 cross-build.~~ done (06:42 #11: all green)

## d) TOTALLY FUCKED UP! 💥

Nothing shipped is broken (build + all 9 test packages green at close).
The honest mishap list:

1. **Wrong-file multiedit:** I pasted the TEST file's import block over
   `actions.go`'s imports (both files were mid-edit in the same batch).
   Caught immediately (the replacement content named `store`/`testing`
   in a non-test file), restored in the next step. One wasted cycle;
   lesson: never batch import-block edits across two files without
   re-reading both.
2. **Rate limiter wired inside the wrong layer:** first wiring put the
   hooks limiter INSIDE `secretGate`, so unauthenticated requests were
   401'd before ever consuming a token — my own test showed 65 requests
   and zero 429s. Reordered limiter outermost (which is also the better
   design: it throttles secret brute-force) and made the test send the
   real secret. The test caught the wiring, exactly its job.
3. **Wrong json/v2 method name:** first `flexPages` implemented
   `UnmarshalerFrom` instead of `UnmarshalJSONFrom` — a valid method the
   decoder simply ignores, so numeric pages passed and string pages
   400'd. The compiler could not catch this; only the table tests did.
4. **Assertions written from the fixture, not the view:** the history
   filter test asserted caller NAMES the row template never renders
   (`cdrTarget` shows numbers). Same failure class again with the vCard
   duplicate test (`data-dial` attribute counted as a second row). Both
   fixed by deriving expectations from the rendered HTML after a debug
   log of the full panel.
5. **My invariant fix was incomplete — the enforcement test caught MY
   OWN bug:** I resolved the TODO's `calls.js→connection.js` import by
   grepping for `getUserAgent` only, missing that `calls.js` ALSO
   imported `ice.js`. The new arch test failed on its first run and
   forced the proper fix (event decoupling). Lesson re-earned: a
   grep-scoped fix for a structural rule is incomplete by construction;
   write the machine check FIRST.
6. **Placeholder junk in a test draft (again):** the first i18n server
   test draft contained undefined helpers and nonsense placeholder
   lines — caught on read-back and fully rewritten. Same failure class
   as the previous session's report; not yet fully learned.
7. **Ignored a dangling import edge of my own making:** when moving
   answer/reject out of main.js I initially left unused imports
   (`bindSession`, `teardownSession`, `announce`) behind; the
   `node --check` + rg sweep caught them before any gate did.

## e) WHAT WE SHOULD IMPROVE! 🛠️

1. ~~Nav labels don't switch language until a full page load~~ → TODO_LIST (nav-language row)
2. ~~An SSE `thread` push resets an open transcript to page 0~~ → TODO_LIST (live-polish row)
3. ~~NixOS module has no runtime test~~ done (stack `telephony-webphone` NixOS VM test, 06:42 #15)
4. Hooks rate limit is per-IP — improved: now `httputil.KeyedRateLimiter` (same per-host semantics, computed Retry-After); per-secret buckets remain the documented upgrade path (ROADMAP)
5. countUnread TTL (5 s) is a deliberate staleness window — documented contract; unchanged.
6. ~~Shortcut discoverability: A/H/M/P/Esc have no visible affordance~~ → ROADMAP (UX polish)
7. ~~German coverage boundary is undocumented for users~~ done (AGENTS.md language policy + FEATURES i18n row say validation reasons stay English)
8. ~~`server_test.go` remainder (sse/proxy/session) should complete
   the split + dedupe the twin helpers~~ done (06:42 #5)
9. ~~The smoke suite still lives in `/tmp`~~ → TODO_LIST (recreate in-repo; /tmp copy lost)
10. ~~`RateLimit` defaults are code constants~~ **Won't implement for now —** no operator demand; revisit on request

## f) Top things to get done next (impact-sorted)

1. ~~Docs overhaul for this session~~ done (06:42 #4)
2. ~~Run `buildflow -s gomod-check --fix` (x/time → direct) + formatting
   over the new files (nixfmt for the module, prettier for island JS).~~ done (06:42 #1 verified direct; island prettier `718cbe7`)
3. ~~`BUILDFLOW_NO_RESULT_CACHE=1 buildflow` full gate + `nix flake
   check` (incl. the new module check) + aarch64 cross-build.~~ done (06:42 #11: all green)
4. ~~Re-run the upstream browser E2E (markup changed: header, history
   form, contacts UI, older-messages button, i18n).~~ done at `00f13fe` (green; later island changes → TODO_LIST re-run row)
5. ~~Assess + plan the upstream switchover in
   `~/projects/nix-international-telephony`~~ done (executed; see ROADMAP)
6. ~~Tag v2.0.0 (pick the commit, push) so CHANGELOG links resolve.~~ done at `d9d6d03`
7. ~~`.buildflow.yml`: `-coverpkg` union coverage + documented cqrs-lint
   skip.~~ cqrs-lint: done; -coverpkg: **Won't** (upstream-blocked) → ROADMAP
8. ~~`nix flake update` (nixpkgs) to clear the 14 build-chain CVEs.~~ done (06:42 #7)
9. ~~Complete the server_test.go split + dedupe the twin helpers.~~ done (06:42 #5)
10. ~~sip.js 0.22 evaluation report (research only; bump gated on E2E).~~ done (evaluation report)
11. ~~NixOS VM test for the systemd unit (runtime proof for the module).~~ done (stack `telephony-webphone` VM test, 06:42 #15)
12. Nav label language refresh (decision + fix, see e1). → TODO_LIST
13. Transcript paging survives SSE pushes (see e2). → TODO_LIST
14. Shortcut help overlay (see e6). → ROADMAP
15. ~~Nginx WSS proxy example tested against the module's generated vhost.~~ done (module ships the optional vhost; the stack's E2E nginx proxies WSS green, 06:42 #14)
16. ~~Session-persistence decision recorded in FEATURES (from b5).~~ done
17. Rate-limit tuning knobs in config (only if an operator asks). **Won't for now**
18. ~~German strings for service validation reasons (if g3 answers
    "keys").~~ **Won't implement —** policy: English, operator-facing (AGENTS.md)
19. Voicemail nudge content-hash guard (prior finding, cheap). → ROADMAP
20. ~~Smoke suite moved into the repo (prior finding).~~ → TODO_LIST
21. `/events` endpoint Go test (prior finding). — partially closed (stream + rate-limit tests); 401/heartbeat → ROADMAP
22. ~~Nav badges live-update (prior finding, now easier with the unread
    cache + hubs lang plumbing as the pattern).~~ → PARKED (OOB verdict)
23. aarch64 runtime note for the module (StateDirectory/hardening on
    aarch64 — build-only verified today). — still build-only → TODO_LIST (aarch64 re-verify row)

## g) Top questions I cannot figure out myself (max 3)

1. ~~**Which commit should carry the v2.0.0 tag?**~~ Resolved: single `v2.0.0` tagged at the release commit `d9d6d03` (07:43 report); no 2.1.0 was cut.
2. ~~**May I touch `nix-international-telephony` in a follow-up session,
   and does its browser E2E run somewhere I can reach from here?**~~ Moot: the switchover was executed and the E2E run green (06:42 #12–15).
3. ~~**Should service-layer validation reasons become i18n keys?**~~ **Won't implement —** they stay English by policy (operator-facing, runbook-greppable; AGENTS.md language bullet).

---

**Bottom line:** 15 of 23 TODO_LIST items done and test-verified,
including both open High questions (island invariant drift is now
machine-enforced, and the deployment ownership question is answered
with a shipped NixOS module); two real bugs were found and fixed by new
tests (pbx query-encoding, hooks limiter placement). Remaining: go.mod
tidy, the full buildflow/nix gates, the browser E2E after this markup
work, the docs overhaul, and the 8 scoped items in (c).
