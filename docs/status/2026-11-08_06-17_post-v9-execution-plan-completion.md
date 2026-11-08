# Status — the post-v9 execution plan: every task executed, verified, pushed green (2026-11-08 06:17)

**Snapshot-rot header: every claim below was verified as of 2026-11-08
06:17 CET.** At HEAD (pushed): **CI run `42440660594` = completed success,
all seven jobs** (fast flake checks incl. the docs battery, `go test
-race`, nix build, live-boot smoke, three advisories). Local legs green:
full `go test -count=1 ./...` (rc=0), `scripts/docs-gates.sh --check`
5/5 PASS, `python3 scripts/status-index.py --check` exit 0, erraudit
tier-2 re-measure 0 violations. This report closes the 2026-11-07
18:46 SUPERB post-v9 execution plan (27 tasks): the P0/P1/P2 tiers were
largely complete at the halfway ledger; this session executed the
remainder — P2-6, P2-9, and all nine P3 tasks — plus one bug the
verification pass surfaced (paperless token parity) and the buildflow
findings-gate closure (P1-5b) that was RED at the halfway point.

## a) FULLY DONE

| #  | Work                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **P1-5b CLOSED: buildflow's findings gate is green and attributable.** The halfway verdict (RED, steps unattributed) attributed to three error-severity findings, all first-party: `cue-vet` — the workflow lacked the fleet `permissions:` block (`ci.yml` now declares `contents: read` with the policy comment); `erraudit` — `errors.As` at the gitignore canary migrated to the modern `errors.AsType` (the gateway tests' established shape); `go-humanize` — `formatThousands` re-implemented comma grouping; deleted for `humanize.Comma` (negative-safe, sign-first — the helper's own bug story is the library's default), its map-case test now pins the library against the same expectations.                                                                                                                                                                                               |
| 2  | **Paperless token parity: a real bug found by critical review and fixed end to end.** The question "is the pack superb?" triggered a re-audit that found `paperless.token_file` did not exist while its CRM twin did — an operator following the documented CRM pattern got the key silently dropped and a misleading "token is set without paperless.url" boot failure. Fixed: struct field + resolve-before-validate + both-sources rejection (`errPaperlessTokenBothSources`) + four-case test pin + README docpin row + two error-registry rows + CHANGELOG line. The docpin gate caught the missing documentation the moment the field existed — the instrument working as designed.                                                                                                                                                                                                                |
| 3  | **P2-9 branching-flow governance: verified, not re-saved.** Live scan 96 → 95 findings (the transport fix below removed one); 95/95 now resolve against the 157-entry baseline (63 already matched; 33 new sites from the landed trains — agent audit, dial gate, CRM context, live context, chat loop — added per-finding with `baseline add`, the only safe writer). **The footgun is on the record:** `--save-baseline` writes the UNACCEPTED remainder, not a snapshot — against a green gate it empties the file (observed live; restored from history `ed9ccb63`, 124 → 157 entries). One real fix absorbed from the scan: the CRM transport no longer type-asserts `http.DefaultTransport` unchecked (guarded clone; unreachable while net/http ships its own Transport, documented at the site).                                                                                                 |
| 4  | **P2-6 quiet-host benchmark: the verdict the busy host could not give.** Re-run at load 0.75/2.06/9.72: spreads 1.4–8.1 % (vs ~2× on 11-06), every shape at or below its 2026-09-24 median (1×1 33.2 vs 35.0; 100×10 104.3 vs 105.0) — no regression; the 2026-11-06 "verdict carried" entry is ANSWERED, not contradicted. Commit `e9c8836b`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| 5  | **P3-1 owner sitting pack, second edition** (`docs/planning/2026-11-07_21-30_owner-sitting-pack.md`): 58 rulings across A product/UX, B privacy/data, C infra, D tooling, E cross-repo, plus a 38-row F bulk table folding the entire 2026-09-22 briefing — each with a recommendation, the other branch's cost, and a payoff map. The first draft was audited before delivery and had real defects (a numbering hole filled by a placeholder row, ten ROADMAP rulings missing, the briefing unconsolidated); all corrected, then consistency-checked mechanically (58 headings = 58 template refs, 38 F-rows, zero broken links). TODO's OWNER-calls row now names it THE ONE answer surface.                                                                                                                                                                                                           |
| 6  | **P3-6 standing-watch calendar** (`docs/watches/2026-11-08_standing-watch-calendar.md`): every watch carries cadence, next due, the named trigger, evidence home, and the on-fire action. The lapsed erraudit monthly check was RE-MEASURED same-session (0 violations, `nolint-audit` 68/68 needed, 0 stale) so the calendar restarts from a measurement, not a moved promise; next erraudit due 2026-12-08.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| 7  | **P3-3 detection answers for the quiet classes**: the error contract now has a "Detection — how anyone KNOWS it happened" section covering twelve quiet rows (CRM down, archiver failing, push transports, receiver loops, dial-gate store, blob readiness, hooks fail-closed, ASR, flight-recorder off, passkey dismissal, transcription skips) — each naming the FIRST place to look — and records that the latency classes have no detection because nothing measures them yet (the design note below owns that). Two §f rows whose work verifiably landed were struck with evidence (the readmodels Contacts move: `store/contacts.go` deleted, `app.go` wires the engine; the empty-vs-filtered audit: the fleet-libs train's per-site rulings ARE that audit) — audit-all dropped 74 → 72, exactly at the pin, so this session's own corpus growth did not buy a ratchet raise.                    |
| 8  | **P3-4 measure-first latency design** (`docs/planning/2026-11-08_latency-measure-first-design.md`): dial-press → first-audio and ASR round-trip defined with t0/t1 at the USER-perceivable boundaries, four named legs per interval, the reporting home decided (island beacons, server-side histograms, opt-in flag default OFF = byte-identical), ~160 LOC cost estimate. No code, per the L3 ruling.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| 9  | **P3-5 four decision briefs** (`docs/planning/2026-11-08_decision-briefs.md`): paperless seam shape (Settings row recommended, per-fax markers' correlation cost named), F1 ink flip and F19 signed-out door (both SHIPPED 2026-11-07 per the review's recommendation — the briefs record the ratification consequence and the reversal cost), attended transfer (the M19 NO-GO stands; the "warn" option is a permanent lie at the worst moment; the GO-day trigger is the stack lab scenario).                                                                                                                                                                                                                                                                                                                                                                                                         |
| 10 | **P3-7 cross-repo handover pack** (`docs/planning/2026-11-08_cross-repo-handover-pack.md`): the four standing legs as verified step-by-step runbooks — stack `crm.caFile` pass-through (webphone's module side SHIPPED and pinned; the stack's option tree has no caFile today, read-verified), `/hooks/ring` dialplan leg (zero hooks references in the stack today, read-verified), `mod_enum` CI-ceiling repair (`applications.enum` at freeswitch-modules.nix:29), and the #9 owner sequence with the acceptance bar. Sequenced by blast radius per effort. No invented paths.                                                                                                                                                                                                                                                                                                                       |
| 11 | **P3-8 native-app preflight** (`docs/planning/2026-11-08_native-app-preflight.md`): the android repo is QUIET (clean tree, last commit 11-05, 2 commits/7d) — HALF the KMP gate met, so the spike STAYS deferred (the gate needs both); the first-device-train checklist carries the seven legs. Nothing mutated in the other repo (git status re-checked identical after the preflight).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| 12 | **P3-9 parked-decisions audit** (`docs/planning/2026-11-08_parked-decisions-audit.md`): all four parked items (scheduled sends NO-GO, session messenger, the ssetest `WithLastEventID` ride, #8 review dimensions) carry live triggers and single homes; the NO-GO premises re-verified at HEAD. One wording risk found (a ratification noted in two places) with the pack named the answer surface.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 13 | **P3-2 filings: one draft shipped, one filing verified to a NO.** The docs-gates four-defect draft (`docs/drafts/2026-11-08_docs-gates-four-sweep-defects.md`) bundles DF brittleness, standalone verbs ignoring the config, escaped-pipe cell splitting, and unvalidated verdict cells — one reproduction each, all first-hand from this session, voice-check `0 FAIL 0 WARN` after the checker caught two real defects (banner regex mismatch, missing attribution footer). The htmx `[object Event]` upstream filing is a **verified NO** (`docs/drafts/2026-11-08_htmx-object-event-verification-no.md`): the served extension is the cqrs-htmx fork, the current fork has no console logging, and upstream `htmx-extensions` sse.js has none either — the premise dissolved at source, which is exactly what the gate exists to catch. The two branching-flow drafts from 2026-10-11 already exist. |

## b) PARTIALLY DONE

1. **Filing the docs-gates issue**: the draft is committed and voice-checked;
   the POST is owner-gated (unsolicited filing, banner box unchecked, and the
   `--save-baseline` empty-write footgun discovered at P2-9 should be folded
   into the same issue as a fifth defect before posting).
2. **The benchowance between README reflow and doc tables**: the README
   config table gained the `paperless.token_file` row; dprint's markdown
   formatter re-aligns table columns, so the row's trailing padding will
   differ from hand-written rows — cosmetic, formatter-owned, no action.

## c) NOT STARTED

Nothing from the plan. Every task P0-1 … P3-9 is executed or verified to a
ruling (the P2 tier's nine tasks: eight done by the halfway ledger, P2-6 and
P2-9 this session; the P3 tier's nine: all this session).

## d) TOTALLY FUCKED UP

1. **`buildflow --save-baseline` emptied the baseline file** (0 findings
   written over 124 entries). Caught by the schema check immediately after
   (findings: 0), but the first `git restore` gave back an EMPTY committed
   version — the daemon had swept the emptied file between the save and the
   restore, so the real recovery came from history (`ed9ccb63`). Lesson
   recorded in the commit: the blanket save is a remainder-saver; per-finding
   `baseline add` is the only safe writer.
2. **The B4 token-parity flip-flop**: I first claimed paperless lacked the
   field, then "corrected" myself after misreading a grep (the TokenFile hit
   belonged to the CRM struct), wrote resolution code against a nonexistent
   field, and the COMPILE failure proved the original claim right. The
   review instinct was correct twice; the intermediate grep conclusion was
   wrong once. The compile gate caught what the grep miscue caused — and the
   end state (field added, tested, documented) is what the seam owed.
3. **The sitting pack's first edition shipped defective** (placeholder E-row,
   ten missing rulings, answer template lagging the IDs). Caught by a
   self-audit before delivery — but the audit should have been part of
   writing it, not a second pass after the owner's "is it superb" challenge.
4. **The nested-fence truncation in draft extraction** (the checker's
   `missing banner` FAIL was my extractor's bug, not the draft's): the
   first-cut regex stopped at the first inner fence — the exact failure mode
   the github-voice skill documents, hit anyway. The skill's `rsplit`
   guidance was applied only after the false signal.

## e) WHAT WE SHOULD IMPROVE

1. **Gate your own artifacts mechanically when a checker exists.** The
   sitting pack consistency check (58 headings = 58 template refs, zero
   broken links) and the voice checker (2 real FAILs) both earned their run;
   the pack's numbering hole would not have survived the mechanical check.
2. **Read struct fields with the struct in context.** A grep for `TokenFile`
   across config.go cannot tell which struct owns the hit — the flip-flop
   class. `go doc` or a scoped block read is the reliable unit.
3. **A ratchet dispute resolves by striking honest rows, not moving the
   number.** The 74-vs-72 trip ended at 72 with zero raise because two rows
   were genuinely landable — the sinkhole stays honest when the strikes are
   real.
4. **The `--save-baseline` semantics belong in AGENTS** (next docs touch):
   "blanket re-save writes the unaccepted remainder; against a green gate it
   empties the file — add per finding."
5. **CI `CHANGELOG coverage: skipped` is correct behavior** (PR-only job) but
   reads ambiguous in the run view; a skip-reason annotation would make the
   matrix self-explaining.

## f) NEXT (harvest inventory — rare tokens used deliberately)

| # | Task                                                                                                                                 | Why now                                                                                          |
| - | ------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ |
| 1 | Fold the `--save-baseline` remainder-write footgun into the docs-gates issue draft before posting (QWxyz fifth defect)               | The draft is committed; the footgun was discovered after it was written                          |
| 2 | Owner sitting: answer the pack (58 rulings + F bulk) — the ONE surface; every locked train (M05, M07–M11, M13) unblocks on it        | The pack is the session's central deliverable and the roadmap's largest unlock                   |
| 3 | The `annotate_max` 205 drain (zero headroom, the concurrent session's own §c3) — a deliberate ANNOTATE session, not a fix-train tail | Two sessions this week sat AT or re-derived pins; the drain removes the recurring re-measure tax |
| 4 | The paperless Settings row (brief 1's recommendation) — ~20 LOC view + i18n + pin once the sitting ratifies the shape                | The T31 invisible-seam gap is the one paperless harm still open                                  |
| 5 | The latency beacon train per the design note (~160 LOC, one tranche) once the sitting blesses the opt-in flag posture                | The design exists; the L3 ruling forbade coding ahead of the note                                |

## g) Questions for the owner (carried, unchanged, from the halfway report)

1. Archive-residue policy for the ~71 ratcheted items (pack D11 has the
   recommendation: per-row evidence live, transcribed verdict archived).
2. `file-size-check`'s 350-line cap: configurable-cap upstream ask
   (pack D10's recommendation) — the 46 accepts stay meanwhile.
3. Two sessions on one mandate: the disjoint-scope rule (pack C10) — this
   session overlapped a CI-heal session again without collision, but the
   narrative-commit tax repeated.

## h) Addendum — the verification battery, end state

`nix develop -c go test -count=1 ./...` rc=0 · `tsc` + island specs green in
the flake check (no JS touched since the last green gate) ·
`scripts/docs-gates.sh --check` 5/5 · `status-index.py --check` exit 0 ·
erraudit `check --max --no-suppress` 0 violations · buildflow findings gate
green with 157-entry baseline, 95/95 live findings resolved · push verified
via `gh run view` (NOT the watch exit code — the concurrent session's §c1
defect stands) · working tree clean, `origin/main == HEAD`.
