# Status — SUPERB pareto triage execution, part 2 (drain + bump + stack arms)

**Claims verified as of 17:35 CET.** The browser VM verdict LANDED (run 10,
exit 0); the webphone flake check passed (all checks); the fresh smoke read
54/54 + all scenarios green on the catalog-bumped binary; webphone main is
RED at a3e9c4d2 on the FOREIGN configjs quartet (tracker #252).

## Direct answers to the opening question

The handoff briefing's open list (§f top 5) is resolved: CI green at the
pre-report tip, docs-gates fully green with cadence headroom BANKED (63/65),
the catalog bump landed (its real blocker was an upstream tag-shape defect,
not the fleet-tags heuristic), and the fleet follow-through (F54/F55/F56) is
done. The stack-repo arms (T17/T18/T21) are executed to their in-session
boundary; the wall-clock and owner-gated halves stay scheduled.

## a) FULLY DONE

| Task                       | What landed                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Evidence                                                                         |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------- |
| Report fixes + gates       | The 15-14 report's table pipes (strike-safety REFUSED the flight-recorder metrics label) fixed; dprint realignment; docs-gates --check then isolated cadence as the only red                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |                                                                                  |
| T10/T11 cadence drain      | THREE 19-day-old reports (2026-10-25 03-54/15-16/21-52) archived gates-first: every §a row pass-struck, §b/§c adjudicated with fact-checked verdicts (CHANGELOG grep, FEATURES grep, module-check greps, stack pbx-prod read); the OIDC report's still-open asks (VM test, oauth2 golden, typed passkey options) ROUTED to a new TODO_LIST section so the pointers land somewhere real; annotool rebuilt from the skill assets; python strikes + dprint realignment; 66 → 63 unarchived, all gates green                                                                                                                                                                                                                                                                                                                       |                                                                                  |
| T16 catalog bump           | ROOT CAUSE, not symptom: `go get catalog/v4@v4.15.0` failed because the release was tagged `catalog/v4/v4.15.0` — a shape Go's tag mapping IGNORES for module `…/catalog/v4` (the valid form is `catalog/v4.15.0`; the @v/list proved the proxy never saw it). The release commit verified green in a detached worktree (build + tests), the correctly-shaped tag pushed on the SAME commit (dashboard-v0.13.0 unblock precedent), proxy pickup < 1 min, then the full bump tail: preflight (no removed symbols), tidy, contracts regen byte-identical, catalog render green, vendorHash re-pinned (the packages.nix edit FIRST hit the templ package's vendorHash — caught before writing), full Go suite green                                                                                                               | commits aaf81eef (daemon-swept go.mod/sum) + vendorHash 70040eae + docs f1a29421 |
| fleet-tags correction      | The bump exposed the script's false-clean: it checked `catalog/v4/v4.15.0` (wrong) and `catalog/v4/v4/v4.15.0` (wrong) and missed the resolvable `catalog/v4.15.0` — the wrong-shape upstream tag read as "tagged". Rewritten to derive the prefix by STRIPPING the go.mod major suffix; a wrong-shape-tagged release now FLAGS; self-test grown to 5 checks (incl. the incident class) and a C-locale sort fix (locale collation reordered `<root>` after `catalog`); live re-sweep clean                                                                                                                                                                                                                                                                                                                                     | 14115fd6 + f1a29421                                                              |
| F54 go-idempotency clock   | verify-before-filing executed (4 `time.Now()` sites store.go:153/174/194/218 + ticker; the library's own suite sleeps real time at 8 sites across 5 files — the sharpest evidence; dupe-search clean; checkout at latest tag); github-voice draft (docs/drafts/go-idempotency-clock-issue.md), checker 0 FAIL/0 WARN after the banner regex was matched to the checker's authority (the SKILL.md prose quote differs from `bannerBody`); FILED as go-idempotency#14, verified landed                                                                                                                                                                                                                                                                                                                                           |                                                                                  |
| F55 + F56                  | etag #5 + appkit #59/#60/#61 all OPEN upstream — no new movement to adopt; F56 (go-idempotency `# expect` anchor) verified ALREADY GREEN by audit-verify replay (5/5) — nothing to fix                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |                                                                                  |
| T22 docs-truth             | FEATURES: flight-recorder row ADDED (limiter, boot log, disk floor, module option, backup-tar fact), /metrics row gains the snapshot counters, /healthz row gains the crm + projections legs (the archived SUPERB report's routed gap), dashboard row rides v0.13.3's no-eval CSP, composition-root row names the pre-drain hub release; ROADMAP observability-depth entry; lessons.md: "plan says X, code pins Y — amend same turn" + "new dataDir artifacts owe the backup path an adjudication"; AGENTS: flight-recorder bullet caught up, templ pin names its one home, fleet-tags.sh joins Commands, S001 note corrected (the enroll site died in the ladder closeout), error-registry regen note gains the MOVE fact (the asr-split proof), currency line: catalog ADOPTED; go-appkit.md: go-flightrecorder now CONSUMED | fe0a7816 + follow-ups                                                            |
| T23 baseline + nolint      | `buildflow -s branching-flow --save-baseline` from this checkout with the result cache disabled: 220 findings re-recorded at current paths (COMPOSITION_mixin re-homed, stale old-path DUPLICATE_TYPE entries gone), all under the severity cap; `erraudit nolint-audit`: 69 directives, 69 needed, 0 stale                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | 03044b77                                                                         |
| T25 residue strikes        | The file-size train's three actionable rows struck with evidence (raceprobe decision c46b9666, baseline refresh, never-fired ledger condition); AGENTS regen-MOVE note                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | cdec4338 + 68c7e588                                                              |
| T19 cqrs-lint prep         | TODO rows (file the suppression-SM + D005 drafts, track the `/*` lexer on the next bump) verified in place; drafts verified in BuildFlow; S001: only server/wirefixtures.go:43 survives — AGENTS corrected; the FILINGS stay owner-tick-gated (the §g posture question stands)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |                                                                                  |
| T20 wire-fixture + android | Both deferral triggers re-checked UNFIRED (config.js keys stable, no second consumer, first device train operator-blocked); the handover pack drafts the line-picking contract (sessionIdentity `extensions`, `/api/session/switch`, no pre-login listing by anti-enumeration posture) + the CallStyle mapping (ring payload facts, the no-PII-through-push ruling, API 31-33 spec shape); android TODO_LIST carries the mirror row (commit 66d8371 there)                                                                                                                                                                                                                                                                                                                                                                     | 317fb3cd                                                                         |
| T21 hub runner             | SystemNix forgejo.nix: `UMask=002` + setgid 2770 cache root/state roots + a root ExecStartPre repair leg (walk bounded to the two small state roots — gomodcache et al are hash-keyed read-only hits); SystemNix `nix flake check --no-build` green; the hub TODO row records it (struck-on-proof); ringhook operator docs + the typed/validated forward + the dialplan E2E leg were already standing                                                                                                                                                                                                                                                                                                                                                                                                                          | SystemNix 8c108bdef (daemon), hub efab146                                        |
| T27 calendar               | Decay pass + baseline re-record annotated on their rows; fleet-tags joins as a TRIGGER-ONLY row (the catalog incident is its pinned class); NO dated row moved — none had fired                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | a908026a                                                                         |
| T18 F60 ring guard         | The inline re-shoot-on-ringing block existed TWICE (desktop + mobile passes) — extracted to `ring_guard_reshoot` with an overridable probe; three forced-state self-test cases (idle keeps the shot, forced ringing fires exactly one re-shoot, override path); a still-up banner after the settle wait now names itself                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | 87bc067e                                                                         |

## b) PARTIALLY DONE

1. **T17 voicemail arms** — verify-live.sh's authenticated audio-fetch leg +
   the browser.nix deposit leg + the E2E playback arm LANDED (stack commit
   7d89a06 + the seam fix: the first VM run failed because browser.nix never
   enabled `webphone.phoneApi`/`operator` — the panel honestly rendered the
   withheld-capability empty state; the eval also caught an uninterpolated
   `apiPasswordFile` derivation). The verdict LANDED: run 10 of the day is
   GREEN (telephony-browser exit 0), through an honest-observable ladder —
   the deposit's vmclient REGISTER (1001/TCP, 300 s expiry) forced the WSS
   assert to count `transport=ws;` contacts (sofia never prints
   `transport=wss`); the VM's headless chromium has NO audio output, so
   play() hangs between load and start (the server log shows the browser
   pulling 206 + 200, 67 KB through the vhost — the fetch chain is REAL),
   neither wp-playing nor readyState ever fire, and the arm's final
   observable set is playing | failed-toast | currentSrc resolving to the
   tokenized recording URL, with the ONE-TOAST-PER-EPISODE dedup asserted
   by a second play attempt on the toast branch. The voicemail fix ships
   with executable coverage; the audio-positive branch is the device-train
   item. F59 (operator.nix 206 pin) was ALREADY LANDED (lines 245-303).
2. **T18 night-run pair (F61)** — the scheduler pack `ui-night-pair.sh` landed
   (two transient systemd user units, run 1 now + run 2 +8h); the pair itself
   needs its quiet window (wall clock). First pair sets the diff noise floor.

## c) NOT STARTED

- T24/T26 and the other owner-gated rows — by design, not skipped.

## d) TOTALLY FUCKED UP

1. **`git add -A` swept a concurrent session's file into a commit (ae26ba18,
   earlier session half)** — recorded in the 15-14 report; this half used
   explicit paths everywhere.
2. **The daemon ate two edits** (index.html SAME-DAY CORRECTION + CHANGELOG
   bump entries): both were re-applied atomically and committed same-call.
   Lesson reinforced: edit → `git add <paths>` → commit in ONE tool call,
   never across.
3. **A placeholder edit in a batch script** (the T25 strike batch) aborted the
   script AFTER the real strikes but BEFORE the AGENTS note — the note was
   re-applied separately; the commit split was cosmetic (both landed).
4. **My own report's table carried literal pipes** in a code span — the same
   class strike-safety exists to catch; the tool caught MY bug the first time
   it ran. No false greens shipped.
5. **The fleet-tags two-form theory**: the T15 "done" verdict shipped with a
   wrong model (the suffixed tag form) that the bump then falsified. The
   script's own false-clean is exactly the class it was built to catch — the
   incident is now its pinned self-test case.
6. **Two multi-command runs joined by newlines instead of `&&`** (the T17 VM
   reruns): a failed python edit step did not stop the build step, so VM runs
   burned against STALE code. The runbook rule exists; violated twice under
   time pressure. Also: one edit anchor died on an em-dash/hyphen mismatch
   and one on `len("vm")` vs `len("vm-")` — anchor edits must be copied from
   the live bytes, never from memory.
7. **`git add -u` swept the concurrent session's four in-flight configjs
   files into a style commit** — caught before push, repaired by
   `git reset --soft` + restage of only my four files; their work returned
   to the working tree untouched. The daemon then swept and pushed their
   half-written quartet anyway (a3e9c4d2), reding CI
   (TestConfigJSCarriesNoSharedDirectory nil-pointer + island-js) — tracker
   #252 filed per the fix-forward protocol (bounded wait, surfaces named,
   both fix options).

## e) WHAT WE SHOULD IMPROVE

1. When a script's "done" rests on an INTERPRETATION of a failure message
   ("the bump's proxy fetch rejected the bare form" → two-form theory),
   re-derive the rule from the authority (Go's tag mapping, the proxy @v/list)
   before writing the verdict.
2. Upstream release tag shapes deserve a fleet-release-skill check: the
   tag-release.sh script produced a Go-invisible tag; a release checklist line
   ("the proxy's @v/list must serve the version") would have caught it at cut
   time.
3. VM-suite config drift: browser.nix needed the phone-api seam only when an
   arm DEPENDED on it — suite configs should carry a comment naming which arms
   need which seams (the fix commit carries it now).

## f) Up to 50 things to get done next (tiered, impact first)

<!-- check-rows:ledger -->

| # | Item                                                                                                    | Tier |
| - | ------------------------------------------------------------------------------------------------------- | ---- |
| 1 | Browser VM rerun verdict: fold into this report (green = T17 fully done; red = fix forward)             | 1    |
| 2 | CI watch on this session's push (name-verified; attribute any red)                                      | 1    |
| 3 | Night pair: run ui-night-pair.sh in the next quiet window; calibrate the diff threshold from the pair   | 2    |
| 4 | Owner: go-idempotency#14 review tick (the provenance banner awaits the manual review)                   | 2    |
| 5 | Owner: etag #5 + appkit #59-61 + the D005/suppression-SM drafts — the direct-fix-or-file posture ruling | 2    |
| 6 | Owner: ALERT_WEBHOOK_URL paste (hub A19) + T24 prod capture + T26 quiet-host hygiene                    | 2    |
| 7 | tag-release.sh tag-shape fix upstream (go-cqrs-lite) — the release script produced the inert shape      | 3    |
| 8 | The OIDC lane residuals (TODO_LIST section): VM test, oauth2 golden, typed passkey options              | 3    |
| 9 | gf-logs state root path: confirm the actual path on the host before the first switch exercises the leg  | 3    |

## g) Questions I can NOT figure out myself

1. **Filing posture (still open from the 15-14 report)**: direct-fix-or-file
   (like today's BuildFlow persist fix) vs human tick for ALL external
   filings — go-idempotency#14 is filed-and-awaiting-tick; is that the right
   default for OWN-repo findings?
2. **tag-release.sh**: the catalog release script cut a Go-invisible tag;
   want me to fix the script's shape rule upstream (one-line derivation
   change + a self-test) or leave it for the owner?
3. **gf-logs**: the hub row says "same treatment for gf-logs" but no config
   in the fleet defines that path — where does it live (a different host's
   unit?), so the provisioning leg covers it?

_Report: docs/status/2026-11-13_17-05_pareto-part2-drain-bump-stack-arms.md.
WAITING FOR INSTRUCTIONS._
