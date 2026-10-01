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

---

## a) FULLY DONE (implemented AND verified this session)

1. **Baseline suite confirmation.** Island `node:test` **120/120**; full
   `go test -count=1 ./...` green on every package. (The "suite red" state I
   inherited was my own test bug — see §d.)
2. **T08 owed pins.**
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
3. **T05 owed server pin.** `TestShellHtmlLangFollowsSessionLang` — 4 cases:
   `wp-lang` cookie wins over `Accept-Language`, German `Accept-Language`
   without a cookie, EN default.
4. **T10 perf (the ETag/gzip half).**
   - `/assets/*` now returns a strong content ETag (sha256 over the embedded
     bytes) and answers `If-None-Match` with a **bodyless 304**.
   - **Scoped gzip** over the whole static subtree only (never `/events`),
     with `Vary: Accept-Encoding`, 304 short-circuit before any
     `Content-Encoding`, and `Content-Length` stripping.
   - Rewrote the stale "caching buys nothing" comment (now documents
     revalidate-always + content ETag).
   - New tests `TestAssetsCarryContentETag` (4 asset paths incl. a
     `FileServerFS` path) and `TestAssetsGzipWhenAccepted`.
5. **T20 hygiene (the actionable clusters).**
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
6. **T26 AGENTS.md compaction.** 705 → **364 lines** (≤377 BuildFlow cap).
   All train chronology/evidence moved to `docs/lessons.md` under a new
   "Provenance moved out of AGENTS.md (2026-10-01 compaction)" section; rules
   kept, evidence relocated.
7. **Formatting drift repair.** After the earlier T06 import edit,
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

1. **T20** — the three actionable finding clusters are fixed, but the
   _preflight_ advisories remain: `go-line-flipflop` (go.mod `go` line changed
   10×/20 commits), `vendor/vendor-freshness` ("go.mod newer than
   vendor/modules.txt"), and the low-disk warning on `/mnt/buildcache`. These
   are advisory, not gate failures, and were deliberately left (touching vendor
   risks the vendorHash roundtrip).
2. **T10** — code + tests done; **10.6 curl timing baseline was NOT produced**
   (`curl` is a banned tool in this harness; I did not substitute a
   Python/`time` baseline). The perf _claim_ is currently backed by the
   ETag/gzip tests, not a before/after number.
3. **T05** — the island boot-language fix (earlier session) + the new server
   pin are done; a **stack browser E2E** to confirm the `html lang` change
   cross-repo is still owed/blocked (T04).
4. **T08** — pins done; but the **stack browser E2E obligation** for the
   served-markup delta (the `#wp-live` comment is comment-only, but the nav
   spec/tests touch served behavior) is not run (T04 blocked).

---

## c) NOT STARTED

- **T11** — perf extras: `modulepreload` for the island ESM graph, outgoing-call
  mic warm (mirror incoming `mic.js`), ICE gathering-time panel, `iceServers`
  trim eval.
- **T12–T19** — the entire UI/UX M9–M26 train: T12 dial affordances + segment
  countdown; T13 history filters + URL state; T14 voicemail playback + fax
  depth; T15 feedback/trust (reconnect banner, undo, retry-in-banner, confirm
  consistency, spinner, success pulse); T16 visual tokens/theming/shell sizing;
  T17 onboarding + mobile extras; T18 messaging richness + pin/archive/mute
  (server design first); T19 i18n/RTL + call depth.
- **T21** — nix-review batch 2 (module golden, tag guard, VM/drill, actionlint,
  devshell dedupe, exceptions).
- **T22** — samber/do + dashboard follow-ups (health.css commit+CI, family
  pins, divergence, `/health` policy).
- **T23** — visual verification gate harness (`scripts/ui-capture.py` + 12-shot
  matrix + AGENTS note).
- **T25** — `internal/server` carve (trigger-gated) + `schema_version` gate +
  gateway stack-side bits.
- Owner legs **T01/T02/T04/T09/T24/T27** — handover-only; no assistant ssh/E2E.
- Docs-health HARVEST of new TODO items into `TODO_LIST.md`, and moving shipped
  rows to CHANGELOG/FEATURES.

---

## d) TOTALLY FUCKED UP (and how)

1. **I shipped a broken test and called the suite "hung".** My
   `calls.test.mjs` "failed hold announces the hold direction" test's
   `HoldInviter.invite()` returned an _ungated_ promise, but `placeCall` awaits
   `inviter.invite()` on the **initial** call — so `await placeCall("1003")`
   never resolved and node:test cancelled the file after 60s. Two background
   shells were consumed chasing this before I read the code. **Lesson:** read
   the production function's await graph before stubbing its collaborator.
2. **Then I got the toast assertion wrong.** I asserted the failure toast was
   `lastToast()`, but a failed HOLD re-renders the settled state and
   re-announces "connected", so the failure toast is _not_ last. Fixed by
   scanning all toasts. **Lesson:** an assertion about "last" is a claim about
   ordering; verify nothing else appends.
3. **I never ran the formatter after the earlier T06 import edit.**
   `buildflow` full **failed on the treefmt check** for `calls.js` — formatting
   drift had been sitting since the earlier session (the previous session's
   "buildflow green" claim predated it). Cost a failed full gate run.
   **Lesson:** after ANY island `.js` edit, `nix fmt` (BuildFlow's formatter
   excludes island files — the flake treefmt owns them).
4. **I wrote a duplicate test block first.** My first optimistic-bubble tests
   re-covered already-pinned behavior; I only noticed by grepping after the
   run, then replaced them with the one genuinely-missing edge. **Lesson:**
   grep for existing coverage before adding a test with a similar name.
5. **My AGENTS.md compaction took three passes** (501 → 490 → 392 → 364) because
   I preserved too much elaboration on the first two. **Lesson:** compact by
   _moving_ paragraphs wholesale, not by trimming words within them.
6. **Wasted tool calls on stale LSP diagnostics.** `vtsls` and
   `golangci_lint_ls` kept reporting errors at old line numbers after edits
   (missed-call syntax "errors", the 3 errcheck warnings) that `node --check`,
   `go test`, and golangci-lint all proved false. I re-litigated them by
   restarting the LSP twice. **Lesson:** after an LSP diagnostic contradicts a
   real command result, trust the command; don't re-verify.

---

## e) WHAT WE SHOULD IMPROVE

1. **Formatter ownership is a trap.** BuildFlow's oxfmt _excludes_ island files
   but the flake's treefmt _owns_ them. Nothing in the loop reminded me, so
   drift hid for a whole session. Add a memory line: "island `.js`/`shell.js`/
   `*.css` → `nix fmt` after edits; BuildFlow does not format them."
2. **The test harness's silent-hang failure mode** (`await` on an
   always-pending stub) should be a documented gotcha next to the existing
   "reload must be stubbed" rule — it cost the most time this session.
3. **`lastToast()` is a footgun** in `calls.test.mjs`; a helper that asserts a
   message is _present among_ toasts (not necessarily last) would prevent the
   regression I wrote.
4. **Perf claims without numbers.** T10's value is a load-time improvement; the
   plan asks for a curl baseline. Without it the change is contract-tested but
   not measured. Find a non-curl timing approach (Python `urllib` + `time`) for
   the next perf task.
5. **LSP freshness.** Consider `BUILDFLOW_NO_RESULT_CACHE=1` in more Go loops,
   and treat stale LSP diagnostics as known-noisy rather than re-checking.
6. **Gate honesty.** "buildflow green" must mean the _current_ tree; the
   inherited claim was stale. Always re-run the cheap format check after an
   island edit before trusting a prior green.

---

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Ordered roughly by the plan's Pareto ranking; [owner] = handover-only.

**Perf / verification**

1. T10.6 — produce a real before/after timing baseline (Python, not curl).
2. T11 — `modulepreload` for the island ESM graph (+ fresh stack E2E).
3. T11 — outgoing-call mic warm, mirroring `mic.js` (node:test).
4. T11 — ICE panel: gathering duration + time-to-first-media.
5. T11 — evaluate `iceServers` trimming (note + numbers).

**UI/UX M9–M26 (T12–T19)**
6. T12.1 — domain-ownership + "already exists?" audit for dial affordances (M10 lesson).
7. T12 — A4 name-on-type + A5 normalization hint.
8. T12 — A8 DTMF animation/tones, A9 re-dial, K5 disclosure.
9. T12 — M12 over-limit segment countdown (i18n en/de).
10. T13 — M11 D8–D10 history filters.
11. T13 — M16 E2/E3 deep-linkable URL state.
12. T13 — E7/E8 filter chips.
13. T14 — M13 voicemail playback (C1–C3, C9/C10).
14. T14 — M14 fax depth (C4–C6).
15. T15 — J2 reconnect banner (uses `#wp-live`; complements the honesty contract).
16. T15 — J3 undo + J4 retry-in-banner.
17. T15 — J7 confirm consistency + J8 spinner + J9 success pulse.
18. T16 — F3/F4/F7/F9 visual tokens.
19. T16 — M20 theming depth.
20. T16 — M26 shell sizing E5/E6.
21. T17 — M18 onboarding/demo K1–K4.
22. T17 — M19 mobile extras I3/I6/I7/I9.
23. T18.1 — server design pass: M21 snippets/schedule (SEAM note first).
24. T18.2 — server design pass: M22 pin/archive/mute (SEAM note first).
25. T18.3 — implement per the design + tests.
26. T19 — M24 locale switch UI + RTL + status dots.
27. T19 — M25 call depth A6/A10.

**Infra / nix / tooling**
28. T21 — module golden + tag guard.
29. T21 — VM/drill hardening + actionlint.
30. T21 — devshell dedupe + documented exceptions.
31. T22 — `health.css` commit + CI regeneration check.
32. T22 — per-seam family pins + divergence check.
33. T22 — `/health` access policy decision.
34. T23 — persist `scripts/ui-capture.py` + the 12-shot matrix + AGENTS note.
35. T25 — `internal/server` carve (trigger-gated).
36. T25 — `schema_version` gate.
37. T25 — gateway stack-side bits.
38. Investigate the `vendor-freshness` preflight (run `GOEXPERIMENT=jsonv2 go mod vendor` in a scratch tree; confirm no diff → document as another known FP).
39. Investigate `go-line-flipflop` (align go-version-auto-configure vs go-mod-update dispositions).
40. `/mnt/buildcache` disk-pressure housekeeping.

**Docs / hygiene**
41. HARVEST this report's (f) into `TODO_LIST.md`; move shipped T05/T06/T07/T08/T10/T26 rows to CHANGELOG/FEATURES (one home per fact).
42. Add the "island files → `nix fmt`" memory line to AGENTS.md.
43. Add the "always-pending stub hangs node:test" gotcha to docs/lessons.md.
44. Decide on a `assertToastPresent` helper to replace `lastToast()` ambiguity.
45. Re-check the 4 fixed links resolve in a rendered doc (lychee on the archived file alone).
46. Monthly erraudit tier-1+2 re-measure (calendar: 2026-10-22).
47. Confirm the daemon pushed local HEAD (`git ls-remote` showed remote `1ac0d4d` vs local `75a5397` at report time).

**Owner-gated (do NOT do as assistant)**
48. [owner] T04 — run the stack `.#telephony-browser` E2E (blocked; you answered "i do not know").
49. [owner] T01/T02/T09/T24/T27 — handover legs (ssh/deploy/PBX).
50. [owner] Ratify the `-t 3` dedup baseline + the pending release-load E2E-retry choice.

---

## g) Questions I CANNOT answer myself (max 3)

1. **T10 measurement.** `curl` is banned in my tool harness, so I skipped the
   curl timing baseline the plan asks for. Do you want me to produce a
   `python3 urllib` + `time` before/after number instead, or is the ETag/gzip
   test contract sufficient evidence for you?
2. **Train scope for the next session.** T12–T19 is ~100 sub-tasks across the
   whole UI/UX surface. Do you want it driven strictly in the plan's tier order
   (one module at a time, each with its own verification), or should I pick the
   single highest-value module (e.g. T15 feedback/trust, which completes the
   honesty-contract story) and land it fully first?
3. **The external lychee 502.** `https://pbx.artmann.tech/` resolves to a 502
   from this host. Is that expected (your PBX deliberately down / not reachable
   here), so I should leave the finding as an accepted external condition, or do
   you want it fenced out of the link check?

---

_Point-in-time snapshot. `TODO_LIST.md` remains the living source; (f) is
HARVEST material for docs-health._
