# Status — T17 typeahead ruling executed, Kith token staged + live-verified, #9 unblocked to deploy

**Claims verified as of 2026-10-29 17:48 CET.** Push/CI statements re-checked at that
stamp: main is `[ahead 4]` with commits that are NOT this session's lane (the signal
attachments train, oldest 16:58 — the owning session's push); every commit of THIS
session is pushed, and the typeahead commit `1d2a331` is an ancestor of the tip that
ran CI green at 16:46 CET (run head `96a393b`). The 16:33 failure run belongs to the
layout-train docs commit and was healed by 16:46.

This report covers the segment since
[14:09](2026-10-29_14-09_superb-closeout-dialgate-t19-t27-and-stack-e2e.md):
three owner answers were given via the question prompt and all three were executed,
plus the Kith-token follow-up the push-secrets WARN prompted. Concurrent sessions
remained active (dedup rail-row re-key, a layout train, the signal inbound-attachments
lane — theirs, not mine).

---

## a) FULLY DONE

1. **Owner answer 1 — T17 placement ruling EXECUTED as code** ("typeahead", not
   just recorded): the cooldown state now renders in the dial typeahead rows.
   - Server: `GET /api/dials` (new Query surface, session-gated, contacts read
     budget) lists the signed-in extension's ACTIVE cooldown windows — newest
     first, capped at 50 (`DialSlotsMaxPerResponse`), expired windows skipped
     lazily; the read never writes. Handler `DialSlots` with the same gate shape
     as the POST: 404 off, 503 nil store, 502 on store failure (the island hides
     the decoration, never blocks a dial). Apidocs entry `calls.dialSlots` +
     drift row; route mounted beside the POST.
   - Store: `DialWindows.Active(ctx, owner, cooldown)` (scan owner filter, sort
     at desc, in-window filter).
   - Island: `dialgate.js` gained the slots cache — SERVER TRUTH, fetched
     debounced (30 s TTL) plus force-refreshed after every gate consult; a
     failed fetch keeps the previous cache (degrade to hidden, never a lie).
     `cooldownFor(raw)` matches display forms to stored remotes through the
     dialable-alphabet normalizer; `cooldownRemainder` phrases the wait.
     `typeahead.js` decorates matching rows ("cooling down · 43m left" /
     "Abklingzeit · noch 43m", i18n ×2 maps); the remaining-window phrasing
     moved to ONE home in dialgate.js (calls.js's private copy deleted).
   - Tests: store pin (`TestDialWindowsActive`: listing, newest-first, window
     echo, owner scoping, lazy expiry), API pin (`TestDialGateSlotsListing`:
     empty array, ordering, scoping, off-404, dead-502), island pins (slots
     cache with normalized matching + degraded fetch, in
     `dialgate.test.mjs`; row decoration incl. "clean contacts decorate
     nothing", in `typeahead.test.mjs` — 9/9 + 7/0).
   - Gates: full Go suite green, island suite 281/281, `tsc --noEmit` green,
     art-dupl 0 new clones, `nix fmt`, CHANGELOG row, plan errata updated to
     the ruled state. Commits `1d2a331` + `bf9436f` — pushed, CI-verified green.
2. **Owner answer 2 — pbx-artmann push**: resolved before I touched it — its
   daemon had already pushed the 15-commit train (`master == origin/master`
   verified at execution time). The #9 host wiring is now on origin for deploys.
3. **Owner answer 3 — Kith token path: found, staged, and LIVE-VERIFIED.**
   - Path (SystemNix `modules/nixos/services/crm.nix`): the CRM's machine API
     runs on `-api-token ${API_TOKEN}` from the sops template; the value is the
     sops secret `crm_api_token` in `platforms/nixos/secrets/crm.yaml`
     (encrypted to the evo-x2 host age key — this box's decrypt verifiably
     fails), and that comment names the pre-module location
     `~/.local/share/crm/api-token` as the SAME token.
   - That legacy file exists locally (65 bytes) → staged verbatim as
     `~/.pbx-prod-secrets/webphone_crm_token` (mode 600, never printed; the
     push-secrets WARN is gone for the next run).
   - **Live probe: HTTP 200** with that bearer against
     `https://crm.home.lan/api/contacts/by-phone` (reserved never-matched
     number, the health-probe pattern; python urllib, default TLS
     verification, no `-k`). One probe closed three open questions: the token
     authenticates against the real evo-x2 deployment, the vhost routes
     `/api/*` (assumed-but-unverified in the 14:09 report), and the
     crm.home.lan cert chain verifies under a stock trust path (supporting the
     dnsblockd-CA assumption behind the PBX probe's `-k` drop).
   - Documented on issue #9 (comment 6498649976): the exact mint/read path,
     why this box cannot rotate the sops value (no age key), and the staged
     state.
4. **Issue #9 is now unblocked to a pure deploy + verify loop**: host wiring on
   origin, token staged and verified, stack module plumbing pre-existing. The
   remaining steps are exactly: `scripts/push-secrets.sh` → deploy the
   pbx-artmann train → the live acceptance bar (owner-side; ssh is banned in
   this shell).

## b) PARTIALLY DONE

1. **The #9 acceptance bar** (inbound caller name, `call_logged` on the
   journal, revoked token → Settings tri-state red): deploy-gated exactly as
   before, but the gate is now ONLY the deploy — every prerequisite is green.
2. **The dialgate smoke scenario** covers the POST contract (9 checks) but was
   NOT extended with the new GET /api/dials slots probe — the surface has API
   tests, no smoke reachability pin. Small, deliberate scope cut at commit
   time; listed in (f).
3. **Typeahead decoration freshness**: the cache refreshes on render (debounced)
   and after every consult, but a fetch that lands mid-render does not
   re-render the open list — the decoration appears on the next keystroke.
   Documented in-code as the tradeoff; an immediate re-render hook would close
   it (listed in (f)).
4. **The final local commits of OTHER lanes** (signal attachments, 4 commits,
   oldest 16:58) are unpushed at snapshot time — 50 min old, inside the 60-min
   bar, not this session's train to push.

## c) NOT STARTED

1. **FEATURES.md rows** — carried from the 14:09 report and STILL not written;
   the feature count missing there is now three (review artifact, dial gate,
   typeahead cooldown state). No test gates this file; it only moves when
   someone remembers it.
2. **`buildflow` full run** — also carried; CI + flake check + targeted tools
   all green, but the tier-2 errorfamily gates have still not explicitly run
   this session.
3. **Dedup-registry row for the dial-gate test pair** (`dialgate_test.go` +
   `dialgate_api_test.go` in two files, one handler): art-dupl-clean, human
   ruling still absent.
4. **M62 (95s budget)** — the last carried §g owner question; still unanswered,
   default still standing.
5. **A separate webphone CRM token identity** vs the shared machine token —
   posture question, not started (see g).

## d) TOTALLY FUCKED UP

1. **The token answer was in my own reading the whole time.** The 14:09
   investigation stopped at "sops decrypt fails → host-side" and I asked the
   owner how to mint one — but the SystemNix comment I had ALREADY READ says
   the sops value IS the token "the pre-module era stored at
   ~/.local/share/crm/api-token", and that file sits on this box, 600-perm,
   never moved. One `ls` I never ran would have closed #9's token leg an hour
   earlier without consuming an owner question. The question I asked ("how is
   it minted") was the wrong question — the right one was "does a copy already
   exist here".
2. **Island-test append without reading the file tail**: I appended the
   slots-cache test AFTER the file's explicitly-commented "runs LAST" 404-latch
   test, got bitten by the module-state latch (`gateOff=true` swallowed my
   fetch, `slotsFetched=0`), and burned a debug cycle reordering. The latch was
   documented fifteen lines above where I appended.
3. **A type-gate cycle I could have skipped**: `cooldownRemainder` was typed
   with required `{last_at, window}` while its real caller passes the verdict's
   optional shape — tsc correctly failed it (the gate WORKING as designed, to
   its credit); I then ran the island suite before tsc in my gate order and
   re-ran both after the fix. Cheap lesson re-learned: tsc is fast; run it
   before the slower suites when the edit is typed JS.
4. **Report/errata sequencing**: the 14:09 report froze the T17 placement as
   an OPEN owner question; the ruling and its implementation landed ~30 min
   later, so the report is already stale on that point by design
   (point-in-time), but I edited the plan errata before re-annotating the
   report — correct order per docs-health doctrine, yet I did not leave the
   report a pointer; anyone reading only the 14:09 file gets the stale
   question with no forward reference.

## e) WHAT WE SHOULD IMPROVE

1. **"Does a copy already exist?" before "how do I mint one?"** — for any
   secret whose provenance is documented in a comment we have read: enumerate
   the named locations (legacy paths, dev fallbacks, sibling checkouts) BEFORE
   declaring extraction impossible. This session's biggest self-inflicted
   round-trip.
2. **Read the tail of a test file before appending to it** — module-state
   ordering constraints live in comments; appending blind re-learns them at
   the cost of a red cycle.
3. **Gate order for typed-JS edits: tsc first**, then the slower island suite.
4. **Smoke scenarios should grow WITH their surface** — adding the GET without
   its smoke check leaves the cheapest reachability pin unclaimed; next route,
   add both in one train.
5. **FEATURES.md needs a forcing function** — three consecutive sessions
   forgot it; either a checklist line in the landing ritual or fold its
   update into the CHANGELOG commit.
6. **Status reports could carry a "superseded-by" forward pointer** when a
   same-session event resolves an open question the report poses (cheap,
   keeps the corpus honest without rewriting point-in-time snapshots).

## f) Up to 50 things we should get done next

1. **Owner: `scripts/push-secrets.sh`** — the staged token ships (WARN gone,
   guard already validated the file shape).
2. **Owner: deploy the pbx-artmann train** (origin has it) — `sudo
   scripts/deploy.sh`-equivalent on the PBX host.
3. **Live #9 acceptance bar**: inbound caller name from Kith; `call_logged`
   on the contact journal; revoked token → Settings tri-state red in one
   cycle.
4. Close issue #9 with the re-verification comment (runbook step 5) once (3)
   passes.
5. Decide the CRM token IDENTITY posture: keep the shared machine token
   (webphone + CV + CLI) or mint a webphone-only one — see (g).
6. FEATURES.md rows: automatic review artifact, dial-gate seam, typeahead
   cooldown state (three, one commit).
7. Run a full `buildflow` (tier-2 errorfamily gates explicitly) at the next
   train start.
8. Extend the dialgate smoke scenario with a GET /api/dials probe (empty
   before dials, listed after one).
9. Typeahead: re-render the open list when a slots fetch lands (close the
   next-keystroke decoration lag).
10. Dedup-registry row (or one-file consolidation) for the dial-gate test
    pair.
11. M62 95s budget: the last unanswered §g item — decide or explicitly drop.
12. `readmodels.DialWindows.Active`: consider pruning expired rows opportunistically
    on the Attempt path (the read is lazy; rows live until trim) — cost/benefit
    call.
13. Reverse-direction apidocs drift assertion: every mounted session-gated
    route must appear in the drift `want` list (closes the silent-missing-entry
    class structurally, not per-route).
14. `GET /api/dials` cap knob: 50 is hardcoded; if a power user ever owns more
    active windows than that, the typeahead silently under-reports (edge case,
    documented, cheap to leave).
15. Island: the `gateOff` 404 latch never resets on SSE reconnect — document
    as accepted or reset on `wp:session-opened` (ruling).
16. `.typeahead-cooldown` could carry `title` with the last-call age for
    pointer-hover parity with the inline chip's state line.
17. The cooldown chip and the row badge share phrasing pieces but not code
    paths for the "last call X ago" half — one shared formatter candidate for
    the dedup registry's next sweep.
18. pbx-artmann: verify the crm-reachability probe goes green post-deploy
    (the `-k` drop now verifies identity — the first real run is the deploy).
19. SystemNix: consider documenting the API-token lifecycle (where it lives,
    who shares it, how to rotate) in the crm module's README — the comment
    trail was findable but scattered.
20. Webphone Settings: the CRM tri-state's poll semantics — confirm a revoked
    token flips it red without a reload (part of acceptance item 3).
21. Kith (crm repo): `/api/contacts/by-phone` on a never-matched number
    answered 200 — consider 404-vs-200 semantics for the health probe pattern
    (document which is the contract; webphone treats either as auth-OK).
22. The live probe pattern (python urllib + Bearer from a file, never
    printing) deserves a one-liner in docs/issue-review-runbook.md's
    verification section.
23. CHANGELOG: the apidocs contract additions (`calls.dialGate` earlier,
    `calls.dialSlots` now) still owe their line (contract-surface changes).
24. Next release runbook pass: version bump + `nix run .#vulnix` + stack
    re-pin — the release carries review artifact + dial gate + typeahead.
25. Stack browser E2E after THIS markup-free island change is optional (no
    served bytes changed) — but the next served-markup train owes one.
26. docs/status: add forward pointers ("superseded-by") to stale open
    questions in the 14:09 report (docs-health ANNOTATE, not rewrite).
27. `wiring_parity_test.go`: grow the omission allowlist doc if the next
    deliberately-unwired Deps seam appears.
28. Second wiring-parity pin: `server.Deps` → `api.Deps` hand-copy
    (server.go's literal) — same bug class, one layer down.
29. `writeJSONStatus(w, code, v)` helper to kill the hand-rolled
    WriteHeader-before-writeJSONBody footgun class.
30. island type gate: my `@param {{last_at?: number, window?: number}}`
    widening worked — consider the same optional-shape treatment for future
    verdict-shaped helpers at authoring time.
31. The slots fetch rides the same `X-CSRF-Token` header on a GET — harmless
    (server ignores it on GET) but worth one comment clarifying why it is
    sent (csrfToken() reuse, not a requirement).
32. `typeahead-cooldown` i18n: the bare-"left" case (expired-at-render) is
    unreachable today (Active filters) but tested only via the "" remainder —
    a one-line comment pins the invariant.
33. Engine `WithLimit` ceiling note: `Active`'s scan relies on the documented
    100-row unlimited ceiling — the cap 50 sits safely under it; keep the
    constant pair in sync if either moves.
34. The signal-lane unpushed commits (not mine) — the owning session should
    watch the 60-min bar; flag if it fires.
35. webphone-android: mirror the dial-gate contract (404-means-off,
    `dialCooldownSeconds`, slots GET) when that lane resumes.
36. `crm-reachability.nix`: split one-unit-two-legs into two alerts (CRM vs
    IdP) to localize outages.
37. Kith issue crm#4 (disposition vocabulary) — the call_logged disposition
    payload is shipped webphone-side; the Kith-side vocabulary row can note
    the producer exists now.
38. Errata hygiene: the plan's §349 "UNVERIFIED external claims" line is
    stale (T19 verified everything) — annotate per docs-health.
39. Consider a `docs/secrets-map.md` (owner-private? repo-public placeholder
   names only) mapping each consumer → secret name → where it lives; today's
   archaeology took three repos.
40. The token staging used `install -m 600` (no umask juggling after the
   shell rejected it) — note the mvdan/sh `umask` builtin gap in the
   concurrent-session habit row next time AGENTS.md is touched.
41. `refreshDialSlots` debounce constant (30 s) is untested — pin it with a
   fake clock if the typeahead tests ever grow timing assertions.
42. If the shared-token posture stays: a rotation drill (new token → sops →
   restart → push-secrets → tri-state green) belongs in the release runbook
   as a once-a-year rehearsal.
43. If a webphone-only token is chosen: the Kith server takes ONE -api-token
   today — per-consumer tokens would need a Kith change (list of tokens or
   per-scope tokens); file it against the crm repo if ruled.
44. The layout train's 16:33 CI failure (healed 16:46) — owning session
    should confirm the heal was their fix, not a flake.
45. `dialgate_api_test.go`'s 3.4 s expiry test: consider `t.Parallel()` for
    the suite wall-clock if it ever matters.
46. The slots cache is per-tab (module state); two tabs of the same session
    each hold their own — acceptable (server truth on next consult), document
    if multi-tab use ever reports it.
47. Post-deploy: grep the PBX journal for `dial-gate:` WARN lines once — the
   dead-store path has never run in production; one clean deploy proves the
   logging shape.
48. `TestDialGateSlotsListing` asserts `"window":3600` — if the default
    cooldown ever changes in tests, this is the line to find.
49. Status-corpus: this is the second same-day report; if the cadence holds,
   the index-file question (declined earlier as YAGNI) reopens at ~30 files.
50. Personal checklist (from 14:09 §e, still open): grep-before-write for
   test names, eval-not-parse for Nix, fetch/python probe before assumption
   — ratify into AGENTS.md hard-won rules or drop.

## g) Questions I cannot answer myself

1. **CRM token identity posture**: the staged token is the SHARED machine
   identity (webphone + the CV syncer + the local `crm` CLI all authenticate
   with the same bearer). Keep that (one identity to rotate, simplest), or do
   you want a webphone-ONLY token so a webphone-side revocation can't take
   the CV syncer down with it? The latter needs a Kith-side change (today the
   server takes exactly one `-api-token`), so it would be an crm-repo issue +
   rotation train, not a webphone change.
2. **Deploy sequencing**: with the token staged + verified and the host wiring
   on origin, do you want the #9 deploy (push-secrets → pbx-artmann deploy →
   acceptance bar) to happen TODAY as its own train, or ride the next release
   cut (the runbook's version-bump + vulnix + stack re-pin dance)? I cannot
   tell whether you want #9 live before the release carries the dial-gate and
   review seams.
3. **M62 / 95s budget** (carried, twice-unanswered): what was the owner
   decision this item awaits — the plan references a 95 s budget I can find
   no ruling for? If it is obsolete, say so and I will mark it dropped in the
   errata instead of carrying it a third time.
