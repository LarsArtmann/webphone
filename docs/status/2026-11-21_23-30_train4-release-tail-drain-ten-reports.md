# Status: train 4 + release.sh tail mode + the 18-08 drain executing — ten reports archived gates-first (2026-11-21 evening, whole-list continuation)

**Created:** 2026-11-21 ~23:30 CET · author: Crush/GLM-5.3 (whole-list mandate continuation)
**Snapshot:** claims verified as of writing; CI verdicts named per push. Live status corpus 55 → 45 reports.

The owner's instruction resumed the mandate ("THE WHOLE TODO LIST"). This
session executed the three executable items and drove the 18-08 drain ten
reports deep. Owner-gated rows (prod deploy, §D10 ratify, capture creds,
pma canary, ALERT_WEBHOOK_URL) stay present-to-owner.

## a) FULLY DONE

| # | Item                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Evidence                                                                              |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------- |
| 1 | **Dependency bump train 4** — httputil v1.9.0 (additive `Capabilities.Pusher`), go-paperless v0.5.3, go-idempotency v0.3.2→v0.6.0, nanoid v1.67.0 (+prng-chacha/aes-ctr-drbg), golang.org/x/oauth2 v0.38.0; full suite green, the 13-invariant contract suite 3× green under `-race`, vendorHash re-pinned (`sha256-0xL+…`), package build green                                                                                                                                     | commit `8b4a45e8` (+ daemon splits `b7aa600b`/`d44118ff`); CI GREEN (ci-verdict rc=0) |
| 2 | **go-idempotency#15 filed** — the contract-suite clock ask, bannered, voice-checked 0 FAIL 0 WARN. The sweep doc's WithClock migration was INEXECUTABLE as written (v0.6.0's `WithClock` is a MemoryStoreOption; `contract.Options` still carries only TimingScale) — premise failure recorded in the sweep doc erratum, the TODO strike, the ci-flake ledger row, and the test's mechanism comment; TimingScale 5 stays as the band-aid with #15 the reopen trigger                 | github.com/LarsArtmann/go-idempotency/issues/15; commit `8b4a45e8`                    |
| 3 | **release.sh §e-1/§e-2** — `--resume-tail` (skips steps 1-5 when the tag is on origin; refuses unless HEAD is the tag commit, so the tail verifies the shipped tree) + the step-7 load gate now WAITS for drain (30s polls, 30-min cap `WEBPHONE_RELEASE_LOAD_WAIT_SECS`) instead of sampling once; runbook carries both. Behavior-tested: HEAD-mismatch refusal, unknown-flag rejection, drain heartbeats + cap exit (observed live across a real 40→264 load storm), override pass | commit `1f84a0d0`; CI GREEN                                                           |
| 4 | **18-08 drain: TEN reports archived gates-first** — the whole 2026-11-08 cluster (08-15, 12-55, 15-48, 16-25, 17-19, 17-22, 21-48) + 15-57 android-wire, 17-03 voicemail-closeout, 19-14 reports-tab; ~510 rows adjudicated (done-at hashes / routed arrows / NOT-DO + won't-implement rulings with reopen triggers); check-rows --strict + bare-rows green rc-gated before every `git mv`; INDEX regenerated each tranche                                                           | commits `a9433075`, `cd94c242`, `08318596` (+ this report's tail)                     |
| 5 | **One stale-routing caught and corrected in-sweep** — the 12-55 report's P3-6 pin-test residue was routed to a fresh TODO row, then the 17-19 report (five hours later same day) revealed `TestCRMCreateLinkArms` had landed via `69e0acfb`; the row is struck STALE with the honest provenance and the archived arrows resolve to it                                                                                                                                                | TODO_LIST row + the 17-19 strike                                                      |

## b) VERIFICATION EVIDENCE

- Full suite `env -u GOEXPERIMENT nix develop -c go test -count=1 ./...` green on train 4; `nix build .#webphone` green on the re-pinned hash.
- CI name-verified GREEN: `8b4a45e8` (train 4), `1f84a0d0` (release.sh). The drain-tranche pushes (`cd94c242`, `08318596`, this report) watched at push.
- Every archive: `strike-safety.py rows` → `status-items --verify` (full-output read) → write → `check-rows --strict` + `bare-rows` rc-captured to files → `git mv` ONLY on double-zero (one gate-before-move slip on 15-48 caught and reversed within a minute — the documented ritual held after).

## c) REMAINS (the executable band, next session)

1. **Drain continuation**: the two big 2026-11-10 reports (16-03 trunk-screened, 38 rows; 19-14 trunk-wire, 27) + the ~40 newer; the TODO row carries the proven recipe.
2. **go-structure-linter#13** (skip `*_test.go` in isTestDataFile) — owner's tool, untouched this session (the drain took the context budget).
3. Watcher fire-drill canary; Android v0.2.3 decision — owner-adjacent.

## d) TOTALLY FUCKED UP (honest)

1. **One gate-before-move violation**: on 15-48 I chained the gates with `;` so `git mv` ran on a FAILED verify (unmatched spec) — the exact 2026-10-15 class. Caught by bare-rows' output, reversed immediately, re-done rc-gated. All later moves used the `if [ $CR -eq 0 ] && [ $BR -eq 0 ]` guard.
2. **The stale routing (§a-5)**: I routed the 12-55 P3-6 residue before reading the SAME DAY's later report. Correct posture for future sweeps: read the ±1-day neighbor reports before minting residue rows.
3. **Two accidental 30-minute load_gate polls** (harness env-prefix bug on my side, not the script) — the second burned a half hour of wall clock. The script's own scrub-pattern is the proven shape.

## e) WHAT WE SHOULD IMPROVE

1. The annotate writer's `any:`/`@substr` keys break on backticked and en-dashed row text — pick substrings away from markup (this session: ~15 fix-and-reverify cycles).
2. Rows containing literal `~~` (inline-code tildes) are structurally unmatchable by the writer (resolve skips them) — leave them bare; they are prose, not gated tables.
3. The drain recipe (predrain → per-report adjudication → rc-gated archive) is now written INTO the TODO row — the next session should not re-derive it.

## f) NEXT

1. Drain the remaining 45 (start: 16-03, 19-14 trunk-wire).
2. go-structure-linter#13 fix.
3. Owner-gated: deploy switch + pbx-artmann relock lane, §D10 ratify, capture creds window, pma canary + SD verdict.

## g) QUESTIONS (standing, from the 19:50 report §g)

Unchanged: (1) deploy now + whose lane is the pbx-artmann relock;
(2) pma push-heat canary arm + sentinel-denylist SD verdict;
(3) §D10 ratify `max_file_size: 1500`. Plus NEW: (4) the 17-03 §g live
observation (the voicemail discriminating click) dies with the deploy —
worth bundling with the post-deploy probes.
