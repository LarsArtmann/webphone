# Master TODO Pareto train — Session 2 status (2026-10-01 23:11)

Point-in-time snapshot of the second execution session against
`docs/planning/2026-10-01_17-32_SUPERB-master-todo-pareto-execution-plan.md`
(tasks T01–T27), issued under the owner's blanket "NOW GET SHIT DONE! The WHOLE
TODO LIST!". `TODO_LIST.md` remains the living source. Prior session:
`2026-10-01_20-12_master-todo-execution-train-status.md` (its Session-2
addendum covers the earlier half of today).

Owner answers carried into this session: (Q1) T06 ring fix **KEEP**; (Q2)
AGENTS.md compaction **authorised**; (Q3) stack E2E status **"i do not know"** →
T04 stays **BLOCKED / handover-only**.

> ARCHIVED 2026-10-03 (docs-health v6 sweep): this session's pins stood (T08/T05
> owed pins, T10 ETag+gzip, T20 hygiene, T26 compaction 705→364); the remaining
> program (T10.6 baseline, T11–T19, T21–T23) executed by the follow-on sessions;
> owner legs stay on their rows; T04 blocked cross-repo. Per-item verdicts
> inline.

---

## a) FULLY DONE (implemented AND verified this session)

~~1. **Baseline suite confirmation.** Island `node:test` **120/120**; full~~ done — this session (report of record)
`go test -count=1 ./...` green on every package. (The "suite red" state I
inherited was my own test bug — see §d.)
~~2. **T08 owed pins.**~~ done — this session (report of record)

- Shell spec: the swap-time `aria-current` mirror (`page`/`false`), the tab
  skeleton reveal on a navigating swap + hide after, and the "settled send
  is never rolled back by a later error" morph edge (the genuinely-unpinned
  edge; the append/failure/new-composer cases were already pinned).
- Server: `TestServedPageHoldsTheDomContract` now asserts
  `aria-current="page"` on the active tab and `"false"` on the rest.
- Corrected the over-promising `#wp-live` comment (`layout.templ`): it
  claimed "connection recovery" announcements that are unbuilt (M17/J2).
- `docs/release-runbook.md`: added the "served-markup changes owe the stack
  browser E2E" rule (explicit, with the internals-only exemption).
  ~~3. **T05 owed server pin.** `TestShellHtmlLangFollowsSessionLang` — 4 cases:~~ done — this session (report of record)
  `wp-lang` cookie wins over `Accept-Language`, German `Accept-Language`
  without a cookie, EN default.
  ~~4. **T10 perf (the ETag/gzip half).**~~ done — this session (report of record)
- `/assets/*` now returns a strong content ETag (sha256 over the embedded
  bytes) and answers `If-None-Match` with a **bodyless 304**.
- **Scoped gzip** over the whole static subtree only (never `/events`),
  with `Vary: Accept-Encoding`, 304 short-circuit before any
  `Content-Encoding`, and `Content-Length` stripping.
- Rewrote the stale "caching buys nothing" comment (now documents
  revalidate-always + content ETag).
- New tests `TestAssetsCarryContentETag` (4 asset paths incl. a
  `FileServerFS` path) and `TestAssetsGzipWhenAccepted`.
  ~~5. **T20 hygiene (the actionable clusters).**~~ done — this session (report of record)
- `chmod +x scripts/render-diff.py` (ruff EXE001).
- 3 errcheck `defer Close()` findings fixed (`fax/service.go`,
  `paperless/paperless_test.go`, `server/version_test.go`).
- 8 oxlint warnings cleared (`theme-preload.js` catch binding + 6 unused
  test vars + a useless spread in `typeahead.test.mjs`).
- 2 samber-linter HW-4 infos on `internal/app/app.go` suppressed **with a
  reason** at the `ProvideNamed` sites (the pair IS eagerly resolved via
  `MustInvokeNamed`; the linter can't see transitive resolution).
- Fixed 4 broken `file://…dedup-registry.md` links in the archived art-dupl
  report (path was `../` from `docs/status/archived/`; correct is `../../`).
  ~~6. **T26 AGENTS.md compaction.** 705 → **364 lines** (≤377 BuildFlow cap).~~ done — this session (report of record)
  All train chronology/evidence moved to `docs/lessons.md` under a new
  "Provenance moved out of AGENTS.md (2026-10-01 compaction)" section; rules
  kept, evidence relocated.
  ~~7. **Formatting drift repair.** After the earlier T06 import edit,~~ done — this session (report of record)
  `internal/web/assets/island/app/calls.js` was unformatted for the flake's
  treefmt gate; ran `nix fmt` (the flake formatter owns island files;
  BuildFlow's formatter excludes them) — `checks.x86_64-linux.format` now
  green.

**Verification ledger:**

| Gate                                     | Result                                                                                              |
| ---------------------------------------- | --------------------------------------------------------------------------------------------------- |
| island `node:test`                       | 120 pass / 0 fail                                                                                   |
| `go test -count=1 ./...`                 | all packages ok                                                                                     |
| fresh-binary smoke                       | 47 passed + 4 restart, 0 failed                                                                     |
| `buildflow` (full, no cache)             | all steps green; findings gate tripped by **gomod-check (54) ONLY** (known false positive)          |
| `nix flake check`                        | "all checks passed!"                                                                                |
| `nix build .#checks.x86_64-linux.format` | green                                                                                               |
| golangci-lint                            | 0 issues                                                                                            |
| lychee                                   | 1 finding = **external** `https://pbx.artmann.tech/` HTTP 502 (owner's PBX down; not a repo defect) |

---

## b) PARTIALLY DONE

~~1. **T20** — the three actionable finding clusters are fixed, but the~~ process record — the advisories were left deliberately; the FP class is AGENTS-documented
_preflight_ advisories remain: `go-line-flipflop` (go.mod `go` line changed
10×/20 commits), `vendor/vendor-freshness` ("go.mod newer than
vendor/modules.txt"), and the low-disk warning on `/mnt/buildcache`. These
are advisory, not gate failures, and were deliberately left (touching vendor
risks the vendorHash roundtrip).
~~2. **T10** — code + tests done; **10.6 curl timing baseline was NOT produced**~~ done — the Python timing baseline landed later (scripts/perf-baseline.py, T10.6)
(`curl` is a banned tool in this harness; I did not substitute a
Python/`time` baseline). The perf _claim_ is currently backed by the
ETag/gzip tests, not a before/after number.
~~3. **T05** — the island boot-language fix (earlier session) + the new server~~ routed — the owed stack E2E rides the TODO island row (T04 blocked cross-repo)
pin are done; a **stack browser E2E** to confirm the `html lang` change
cross-repo is still owed/blocked (T04).
~~4. **T08** — pins done; but the **stack browser E2E obligation** for the~~ routed — same row (the markup delta's E2E obligation)
served-markup delta (the `#wp-live` comment is comment-only, but the nav
spec/tests touch served behavior) is not run (T04 blocked).

---

## c) NOT STARTED

~~- **T11** — perf extras: `modulepreload` for the island ESM graph, outgoing-call~~ done — T11 shipped (modulepreload, dial-focus warm, ICE setup line); the trim eval was not adopted
mic warm (mirror incoming `mic.js`), ICE gathering-time panel, `iceServers`
trim eval.
~~- **T12–T19** — the entire UI/UX M9–M26 train: T12 dial affordances + segment~~ done — the whole UI/UX train shipped (CHANGELOG Unreleased; per-module verdicts in the archived 06-59 report)
countdown; T13 history filters + URL state; T14 voicemail playback + fax
depth; T15 feedback/trust (reconnect banner, undo, retry-in-banner, confirm
consistency, spinner, success pulse); T16 visual tokens/theming/shell sizing;
T17 onboarding + mobile extras; T18 messaging richness + pin/archive/mute
(server design first); T19 i18n/RTL + call depth.
~~- **T21** — nix-review batch 2 (module golden, tag guard, VM/drill, actionlint,~~ done — golden + tag guard + post-split VM; actionlint/dedupe/exception rows not adopted
devshell dedupe, exceptions).
~~- **T22** — samber/do + dashboard follow-ups (health.css commit+CI, family~~ done — health.css script + canary, family pins, divergence documented; the /health policy rides the TODO health row (owner)
pins, divergence, `/health` policy).
~~- **T23** — visual verification gate harness (`scripts/ui-capture.py` + 12-shot~~ done — scripts/ui-capture.py + the 14-shot matrix + the AGENTS note
matrix + AGENTS note).
~~- **T25** — `internal/server` carve (trigger-gated) + `schema_version` gate +~~ resolved by events — schema_version shipped with T18; the carve stays gated; the gateway bits ride the cross-repo row
gateway stack-side bits.
~~- Owner legs **T01/T02/T04/T09/T24/T27** — handover-only; no assistant ssh/E2E.~~ routed — TODO owner rows (handover legs stay owner-terminal)
~~- Docs-health HARVEST of new TODO items into `TODO_LIST.md`, and moving shipped~~ done — the one-home moves + this v6 sweep
rows to CHANGELOG/FEATURES.

---

## d) TOTALLY FUCKED UP (and how)

~~1. **I shipped a broken test and called the suite "hung".** My~~ process record
`calls.test.mjs` "failed hold announces the hold direction" test's
`HoldInviter.invite()` returned an _ungated_ promise, but `placeCall` awaits
`inviter.invite()` on the **initial** call — so `await placeCall("1003")`
never resolved and node:test cancelled the file after 60s. Two background
shells were consumed chasing this before I read the code. **Lesson:** read
the production function's await graph before stubbing its collaborator.
~~2. **Then I got the toast assertion wrong.** I asserted the failure toast was~~ process record
`lastToast()`, but a failed HOLD re-renders the settled state and
re-announces "connected", so the failure toast is _not_ last. Fixed by
scanning all toasts. **Lesson:** an assertion about "last" is a claim about
ordering; verify nothing else appends.
~~3. **I never ran the formatter after the earlier T06 import edit.**~~ process record
`buildflow` full **failed on the treefmt check** for `calls.js` — formatting
drift had been sitting since the earlier session (the previous session's
"buildflow green" claim predated it). Cost a failed full gate run.
**Lesson:** after ANY island `.js` edit, `nix fmt` (BuildFlow's formatter
excludes island files — the flake treefmt owns them).
~~4. **I wrote a duplicate test block first.** My first optimistic-bubble tests~~ process record
re-covered already-pinned behavior; I only noticed by grepping after the
run, then replaced them with the one genuinely-missing edge. **Lesson:**
grep for existing coverage before adding a test with a similar name.
~~5. **My AGENTS.md compaction took three passes** (501 → 490 → 392 → 364) because~~ process record
I preserved too much elaboration on the first two. **Lesson:** compact by
_moving_ paragraphs wholesale, not by trimming words within them.
~~6. **Wasted tool calls on stale LSP diagnostics.** `vtsls` and~~ process record
`golangci_lint_ls` kept reporting errors at old line numbers after edits
(missed-call syntax "errors", the 3 errcheck warnings) that `node --check`,
`go test`, and golangci-lint all proved false. I re-litigated them by
restarting the LSP twice. **Lesson:** after an LSP diagnostic contradicts a
real command result, trust the command; don't re-verify.

---

## e) WHAT WE SHOULD IMPROVE

~~1. **Formatter ownership is a trap.** BuildFlow's oxfmt _excludes_ island files~~ done — the AGENTS formatting rule exists (nix fmt BEFORE the gates after island edits)
but the flake's treefmt _owns_ them. Nothing in the loop reminded me, so
drift hid for a whole session. Add a memory line: "island `.js`/`shell.js`/
`*.css` → `nix fmt` after edits; BuildFlow does not format them."
~~2. **The test harness's silent-hang failure mode** (`await` on an~~ done in part — AGENTS documents the island-test stub pitfalls (reload/DOM); the always-pending-hang line wasn't separately recorded
always-pending stub) should be a documented gotcha next to the existing
"reload must be stubbed" rule — it cost the most time this session.
~~3. **`lastToast()` is a footgun** in `calls.test.mjs`; a helper that asserts a~~ not adopted — below the bar; the helper wasn't added
message is _present among_ toasts (not necessarily last) would prevent the
regression I wrote.
~~4. **Perf claims without numbers.** T10's value is a load-time improvement; the~~ done — perf-baseline.py (Python urllib + time) is the standing approach
plan asks for a curl baseline. Without it the change is contract-tested but
not measured. Find a non-curl timing approach (Python `urllib` + `time`) for
the next perf task.
~~5. **LSP freshness.** Consider `BUILDFLOW_NO_RESULT_CACHE=1` in more Go loops,~~ process record
and treat stale LSP diagnostics as known-noisy rather than re-checking.
~~6. **Gate honesty.** "buildflow green" must mean the _current_ tree; the~~ process record
inherited claim was stale. Always re-run the cheap format check after an
island edit before trusting a prior green.

---

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Ordered roughly by the plan's Pareto ranking; [owner] = handover-only.

**Perf / verification**

~~1. T10.6 — produce a real before/after timing baseline (Python, not curl).~~ done — scripts/perf-baseline.py (cold vs gzip vs revalidate)
~~2. T11 — `modulepreload` for the island ESM graph (+ fresh stack E2E).~~ done — modulepreload shipped (T11); the fresh E2E rides the TODO island row
~~3. T11 — outgoing-call mic warm, mirroring `mic.js` (node:test).~~ done — the dial-focus mic warm shipped (calls.js:556)
~~4. T11 — ICE panel: gathering duration + time-to-first-media.~~ done — ice.js setupLine (gather/ice/first media)
~~5. T11 — evaluate `iceServers` trimming (note + numbers).~~ not adopted — below the bar; no row taken

**UI/UX M9–M26 (T12–T19)**
~~6. T12.1 — domain-ownership + "already exists?" audit for dial affordances (M10 lesson).~~ done — the ownership audit is the standing protocol (the M10 lesson is AGENTS doctrine)
~~7. T12 — A4 name-on-type + A5 normalization hint.~~ not adopted — below the bar; name-on-type wasn't taken
~~8. T12 — A8 DTMF animation/tones, A9 re-dial, K5 disclosure.~~ done in part — T12 affordances + countdown; the DTMF animation/tones weren't taken
~~9. T12 — M12 over-limit segment countdown (i18n en/de).~~ done — T12 segment countdown
~~10. T13 — M11 D8–D10 history filters.~~ done in part — the URL-addressable filter landed (history.templ); the D8–D10 chip shape varies
~~11. T13 — M16 E2/E3 deep-linkable URL state.~~ done in part — deep-linkable filter state landed; last-active restore wasn't taken
~~12. T13 — E7/E8 filter chips.~~ not adopted — below the bar (chips as such)
~~13. T14 — M13 voicemail playback (C1–C3, C9/C10).~~ done — T14 voicemail player
~~14. T14 — M14 fax depth (C4–C6).~~ done — T14 fax timeline + resend
~~15. T15 — J2 reconnect banner (uses `#wp-live`; complements the honesty contract).~~ done — the honesty train's offline banner + T15's SSE recovery announce
~~16. T15 — J3 undo + J4 retry-in-banner.~~ done in part — the retry surface is the failed bubble's Retry/Dismiss (T15); undo wasn't taken
~~17. T15 — J7 confirm consistency + J8 spinner + J9 success pulse.~~ done in part — the T15 surfaces landed; spinner/pulse weren't taken
~~18. T16 — F3/F4/F7/F9 visual tokens.~~ done — T16 mirrored tokens
~~19. T16 — M20 theming depth.~~ done in part — T16 theme system; the depth extras weren't taken
~~20. T16 — M26 shell sizing E5/E6.~~ done — T16 shell sizing
~~21. T17 — M18 onboarding/demo K1–K4.~~ done — T17 onboarding
~~22. T17 — M19 mobile extras I3/I6/I7/I9.~~ done in part — T17 mobile; the I3/I6/I7/I9 extras weren't all taken
~~23. T18.1 — server design pass: M21 snippets/schedule (SEAM note first).~~ done — T18 seam design (D1–D14) + snippets (schedule cut NO-GO)
~~24. T18.2 — server design pass: M22 pin/archive/mute (SEAM note first).~~ done — T18 thread flags
~~25. T18.3 — implement per the design + tests.~~ done — T18 implementation + tests
~~26. T19 — M24 locale switch UI + RTL + status dots.~~ done in part — T19 RTL groundwork; the locale-switch UI/status dots weren't taken
~~27. T19 — M25 call depth A6/A10.~~ done — T19 focus mode + device self-test

**Infra / nix / tooling**
~~28. T21 — module golden + tag guard.~~ done — nix/module-output.golden + the release.sh tag guard
~~29. T21 — VM/drill hardening + actionlint.~~ done in part — VM+drill ran green post-split (the 20:12 battery); actionlint never run
~~30. T21 — devshell dedupe + documented exceptions.~~ not adopted — the devshell dedupe + exception rows weren't taken
~~31. T22 — `health.css` commit + CI regeneration check.~~ done — TODO health row (build-health-css.sh + checks.health-css canary)
~~32. T22 — per-seam family pins + divergence check.~~ done — T22 family pins (store.thread_flag, store.count_archived, snippets)
~~33. T22 — `/health` access policy decision.~~ routed — TODO health row (the /health exposure policy is owner)
~~34. T23 — persist `scripts/ui-capture.py` + the 12-shot matrix + AGENTS note.~~ done — scripts/ui-capture.py + the 14-shot matrix + the AGENTS note
~~35. T25 — `internal/server` carve (trigger-gated).~~ record stands — the carve stays trigger-gated
~~36. T25 — `schema_version` gate.~~ resolved by events — T18 shipped versioned migrations (schema_version v2)
~~37. T25 — gateway stack-side bits.~~ routed — TODO cross-repo row (the gateway stack-side bits)
~~38. Investigate the `vendor-freshness` preflight (run `GOEXPERIMENT=jsonv2 go mod vendor` in a scratch tree; confirm no diff → document as another known FP).~~ not adopted — below the bar; the vendor FP class is AGENTS-documented
~~39. Investigate `go-line-flipflop` (align go-version-auto-configure vs go-mod-update dispositions).~~ not adopted — below the bar
~~40. `/mnt/buildcache` disk-pressure housekeeping.~~ housekeeping — host disk pressure; no repo action recorded

**Docs / hygiene**
~~41. HARVEST this report's (f) into `TODO_LIST.md`; move shipped T05/T06/T07/T08/T10/T26 rows to CHANGELOG/FEATURES (one home per fact).~~ done — the one-home moves landed + this v6 sweep harvested
~~42. Add the "island files → `nix fmt`" memory line to AGENTS.md.~~ done — the AGENTS formatting rule exists
~~43. Add the "always-pending stub hangs node:test" gotcha to docs/lessons.md.~~ done in part — AGENTS documents island-test stub pitfalls; the hang line wasn't separately recorded
~~44. Decide on a `assertToastPresent` helper to replace `lastToast()` ambiguity.~~ not adopted — below the bar; the helper wasn't added
~~45. Re-check the 4 fixed links resolve in a rendered doc (lychee on the archived file alone).~~ process record — lychee ran green at the v2.8.0 release tail
~~46. Monthly erraudit tier-1+2 re-measure (calendar: 2026-10-22).~~ routed — standing watch (TODO watches row; next due 2026-10-22)
~~47. Confirm the daemon pushed local HEAD (`git ls-remote` showed remote `1ac0d4d` vs local `75a5397` at report time).~~ done — the daemon caught up; the ls-remote ritual is standing

**Owner-gated (do NOT do as assistant)**
~~48. [owner] T04 — run the stack `.#telephony-browser` E2E (blocked; you answered "i do not know").~~ routed — TODO cross-repo row (T04, owner terminal)
~~49. [owner] T01/T02/T09/T24/T27 — handover legs (ssh/deploy/PBX).~~ routed — TODO owner rows (deploy tail, sitting, handover legs)
~~50. [owner] Ratify the `-t 3` dedup baseline + the pending release-load E2E-retry choice.~~ routed — TODO OWNER-calls row (the -t 3 baseline pending ratification)

---

## g) Questions I CANNOT answer myself (max 3)

~~1. **T10 measurement.** `curl` is banned in my tool harness, so I skipped the~~ resolved by events — the Python baseline landed (perf-baseline.py)
curl timing baseline the plan asks for. Do you want me to produce a
`python3 urllib` + `time` before/after number instead, or is the ETag/gzip
test contract sufficient evidence for you?
~~2. **Train scope for the next session.** T12–T19 is ~100 sub-tasks across the~~ resolved by events — the follow-on sessions drove the whole T12–T19 train
whole UI/UX surface. Do you want it driven strictly in the plan's tier order
(one module at a time, each with its own verification), or should I pick the
single highest-value module (e.g. T15 feedback/trust, which completes the
honesty-contract story) and land it fully first?
~~3. **The external lychee 502.** `https://pbx.artmann.tech/` resolves to a 502~~ process record — the 502 was the PBX down (external; lychee green at the release tail)
from this host. Is that expected (your PBX deliberately down / not reachable
here), so I should leave the finding as an accepted external condition, or do
you want it fenced out of the link check?

---

_Point-in-time snapshot. `TODO_LIST.md` remains the living source; (f) is
HARVEST material for docs-health._
