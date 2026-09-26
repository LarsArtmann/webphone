# Status: session 2026-09-26 — TODO execution + post-heal gate battery + self-review

Session scope: two work cycles on 2026-09-26 (~16:40–19:24 CEST), starting
from a clean tree at `d38e39e`. This report covers ONLY what this session
did and noticed (per instruction; no fresh research beyond session facts).
Time of writing: 2026-09-26 19:24 CEST.

Starting TODO_LIST state: 8 rows — 1 assistant-actionable, 5
owner-blocked-by-design, 1 done-parked, 1 not-due.

---

## Direct answers (brutal self-review)

**1. What did you forget?**
- Turn 1 treated TODO_LIST as the only work source and stopped there. The
  2026-09-25 sweep self-review (`docs/status/2026-09-25_05-10_docs-verify-sweep-self-review.md`)
  carried open items 7/8/9/13/20 — I only picked them up in turn 2. The
  post-heal battery (item 20) should have been obvious in turn 1: I had
  JUST verified host nix heals builds; the natural next question was
  "then prove main is green again".
- The FEATURES "Browser E2E" row staleness (says greens are 293s/322s from
  2026-09-22; the 2026-09-23 greens were 195s/184s). I edited the row
  DIRECTLY ABOVE it and didn't see it.
- The TODO_LIST header overclaim ("every claim checked" — self-review
  item 6). I read that header twice this session. Didn't act.
- The daemon push stall: turn 1's final message didn't run `git ls-remote`;
  the stall (see d1) was only discovered in turn 2. The runbook culture
  says verify end states that way — I skipped it in turn 1.

**2. What is stupid that we do anyway?**
- Trusting `result` symlinks after `--no-link` builds (caused d2).
- Relying on the auto-commit daemon's push without any lag detection —
  the release ritual DEPENDS on pushes, and the stall was silent.
- Docs accumulating drift while every sweep adds more docs to drift
  (AGENTS buildflow claims vs observed behavior, FEATURES E2E numbers,
  TODO header overclaim — three drift instances found in one session).

**3. What could you have done better?**
- Run `git ls-remote` as a turn-1 closer.
- Read the self-review doc as a work source immediately after
  classification (it is the VERIFY sweep's residue list).
- Understand `nix build --system` semantics on nix 2.34 instead of
  accepting a green artifact: the run printed "warning: ignoring the
  client-specified setting 'system', because it is a restricted setting
  and you are not a trusted user" and I did not fully resolve WHY the
  aarch64 drv still built. The ARTIFACT is byte-verified aarch64
  (e_machine=183) so the outcome stands, but the mechanism is murky in my
  head — that is an unresolved tool-semantics gap, honestly flagged.
- Tail the FULL buildflow output instead of the last 25 lines; I verified
  exit 0 + the skips list but never saw the per-step greens for go-test/
  treefmt/erraudit inside buildflow (I re-ran the go suite and erraudit
  tier 1 standalone to compensate — that part was done right).

**4. What could you still improve?** → see section (e).

**5. Did you lie?** No. Two imprecisions corrected here: (a) turn 1 said
"host nix fully healed" — true for builds via the hand-planted symlink,
but the DURABLE fix is still open (GC/reboot rot; sheet §0 says so, my
annotation kept that). (b) "aarch64 proven end-to-end" is true at the
artifact level, with the flag-semantics caveat above.

**6. Ghost systems?** None created; none found within session scope
(no new code was written — comments and docs only).

**7. Split brains?** None created knowingly. Two weak ones noted for the
watchlist: the command sheet's §0 STATUS + §6 drafts list now DUPLICATE
facts whose one-homes are AGENTS/lessons and TODO_LIST (owner-facing
convenience copies — drift risk, kept in sync today). And the
FEATURES-vs-AGENTS E2E-number drift (d4) is a real instance of the class.

**8. Scope creep?** None — stayed inside the session's TODO + post-heal
verification scope; explicitly did NOT research unrelated areas.

**9. Removed something useful?** No. The deleted TODO row's content is
preserved verbatim-enough in the closure narrative (docs-health style:
done work is deleted, never struck through).

**10. Tests?** No new tests needed (zero behavior changes: one comment,
doc rows). Full Go suite green (all packages), island JS green via flake
check's island-js, erraudit tier 1 green. Nothing to improve from this
session's changes specifically.

**Mistakes made and recovered (honesty ledger):**
- Sheet edit attempted without View first → tool refused → recovered.
- TODO_LIST edit raced a daemon commit (mod-time conflict) → re-read,
  content identical, re-applied. Benign but exactly why read-before-edit
  is a rule.
- d2 below: the near-miss that matters.

---

## a) FULLY DONE (evidence attached)

| # | Work | Evidence |
|---|------|----------|
| a1 | `classifyForUser` comment now cites `Family.HTTPStatus()` as the REJECTED alternative with the canonical mapping (Rejection→400, Transient/Infrastructure→503, Corruption/Orchestration→500) and the why: it would downgrade fixable refusals to 400 and scatter system-side families across 500/503, losing the pinned two-value 422/502 contract | `2be4061`, mapping verified against go-error-family **v0.10.2** `family.go` (not the AGENTS paraphrase); TODO row closed; the last open slice of the error-family row |
| a2 | TODO_LIST closure narrative (2026-09-26, extended with the battery evidence) + row deletion per docs-health convention | `2be4061` + `0230ead` |
| a3 | Command sheet: §0 STATUS note (healed 2026-09-25 05:32, `nix build nixpkgs#hello` green, durable fix still open — skip to §1); §4-§6 stale "TODO row 40/42/43" → stable task names (self-review item 7); §6 corrected — v2.7.0 drafts EXIST since `357ffec`, sheet claimed "to be added post-release" | `d2946b8` |
| a4 | FEATURES.md gained the missing `templ-components adoption` Platform row (layout.Base + EmptyState ×6 + `/assets/tw.css` 18.9KB + coexistence verdict + hand-roll dispositions) — self-review item 13 | `d2946b8` |
| a5 | Post-heal gate battery, ALL GREEN on main: `nix build .#webphone` (2.7.0, no vendorHash drift) · `nix flake check` all checks passed incl. the KVM backup VM · loopback smoke 41+4 · `nix run .#vulnix` zero real advisories (all distro-patched noise) · buildflow full exit 0 · erraudit tier 1 "no violations" · aarch64 cross-build ELF-verified **e_machine=183** (the healed `/run/binfmt` path proven at the artifact level) | session run log; main @ `0230ead` |
| a6 | Prod probed (read-only, foreign mode): 19 passed / 0 failed / 16 skipped, `--expect-version 2.6.0` green — TODO evidence refreshed to 2026-09-26 | session run log |

## b) PARTIALLY DONE

| # | Work | What works | What remains | Blocker | Effort |
|---|------|-----------|--------------|---------|--------|
| b1 | Getting main to origin | RESOLVED while this report was being written: the daemon pushed and `git ls-remote` = HEAD = `0230ead` (sync verified). Total stall was ~45 min, then self-healed | Residual: the stall was silent and its cause is unknown (timer? transient auth/network?) — g1 asks whether a lag threshold should be treated as broken | — |
| b2 | aarch64 verification story | Artifact byte-verified aarch64; the healed binfmt sandbox demonstrably works | `--system` flag semantics on nix 2.34 (restricted-setting warning vs successful aarch64 drv build) not understood | Needs a focused nix-docs pass (S) | S |
| b3 | TODO_LIST execution this session | 1 of 8 rows executed (the only assistant-actionable one); all others triaged with reasons | 6 rows remain — all owner-terminal or not-due by design (see c) | Owner decisions / ssh forbidden / calendar | — |

## c) NOT STARTED (deliberately — with reasons)

- **v2.7.0 release tail + deploy**: OWNER-terminal by explicit designation
  (pbx-artmann AGENTS forbids assistant ssh/deploy); executing the early
  steps would break the owner's copy-paste sheet. Precondition b1 must
  clear first.
- **Outbound SMS bridge root cause**: owner must journal
  `telnyx-webhooks` via ssh (forbidden to me); webphone side already
  fixed/classified (`1d53f44`, `6ac8962`).
- **OWNER-calls batch** (~27 queued decisions), **announcements posting**
  (drafts all exist; owner picks channel/wording), **release E2E for
  templ-components** (rides the v2.7.0 gate): owner decisions/ritual.
- **Standing watches**: not due (monthly erraudit 2026-10-22, quarterly
  2026-12-20) — but see f10: the tier-2 shrink needs PROJECT work before
  the deadline, not a re-measure.
- **erraudit tier-2 family-adoption conversions** (config.go 22,
  store/messages.go 20, pbx/client.go 8): known plan, not started; due
  date is the driver.

## d) TOTALLY FUCKED UP

1. **Remote `main` WAS 3 commits behind local — resolved while this
   report was written.** The daemon's push caught up (~45 min total
   lag) and `git ls-remote` now equals HEAD (`0230ead`); the v2.7.0
   ritual precondition ("changes must be PUSHED before the stack
   re-pins") is MET. Residual fuckup, downgraded: the stall was
   SILENT — nothing distinguishes "healthy daemon with slow push" from
   "broken push" from the outside, and the release ritual's first
   minutes would have been spent discovering this. Mitigation stands:
   f17 (release.sh remote-sync preflight) + f18 (lag watchdog).
2. **False-verification near-miss (self-caught, zero landed damage).**
   First aarch64 "verification" passed `--no-link` and then checked the
   STALE `result` symlink → read an x86_64 binary (e_machine=62). The
   byte check caught it (house rule: "verify by ELF bytes, never exit
   code alone" — it earned its keep); re-built with explicit `-o` and
   got e_machine=183. Had I trusted the exit code, I would have reported
   a false green on the exact leg the binfmt outage broke. Root cause:
   my own sloppy pairing of flags; lesson belongs in docs/lessons.md
   (f16).
3. **AGENTS lies about buildflow full mode.** AGENTS says gitleaks/
   codespell/markdown-lint RUN in build mode `full` (and that
   scripts/buildflow.sh appends the first two). Observed output today:
   all three "skipped by build mode 'full'". Severity: future sessions
   believe a stronger gate ran than actually ran; also makes
   self-review item 8 (3 markdown-lint warnings) unresolvable via the
   documented path. Root cause: unknown — `.buildflow.yml` or
   buildflow.sh behavior changed after AGENTS was written (not
   investigated per session scope). 
4. **FEATURES "Browser E2E" row is stale** (293s/322s @ 2026-09-22 vs
   the actual latest greens 195s/184s @ 2026-09-23 with the FOUC
   scenario aboard). I edited the adjacent row and missed it — doc
   drift I directly touched and skipped.
5. **TODO_LIST header overclaim still live** ("docs-health VERIFY:
   every claim checked against code/git, prod probed") — self-review
   item 6 asked to scope that phrasing honestly; evidently never
   landed. I read the header twice this session without acting.

## e) WHAT WE SHOULD IMPROVE

1. **Artifact checks use explicit out-links, never `result`** — pair
   `nix build … -o <path>` with the byte check, always. One-line AGENTS/
   lessons addition (f16).
2. **Push-lag detection**: release.sh already preflights host-nix +
   clean tree; add an unpushed-commits assert (`git ls-remote` vs HEAD)
   so the ritual fails LOUDLY at preflight instead of silently pinning
   unpushed state (f17). Consider the same assert at session phase
   boundaries.
3. **Self-review docs are a work source**: after a docs-health VERIFY
   sweep, its self-review open-items list should be folded into
   TODO_LIST (HARVEST) immediately — not left for a future session to
   rediscover (that's exactly what happened to items 7/13/20 this
   session).
4. **Understand the tool before accepting the green**: nix `--system`
   semantics (b2) — one short research pass, then a lesson line.
5. **Reconcile buildflow doc-vs-behavior** (d3): read
   scripts/buildflow.sh + .buildflow.yml, then fix AGENTS or the config
   — one of them is wrong today.
6. **Doc drift is compounding**: three drift instances found in one
   session (d3/d4/d5). Cheap fix batch: f11–f15 (all S).
7. **erraudit tier-2 needs scheduled project time before 2026-10-22**:
   a re-measure against a non-shrinking 113 baseline is a failed
   re-measure by definition. Convert the top seams first (f7–f9).
8. **Vulnix scope clarity**: runtime closure = triaged (clean);
   buildflow's vulnix also warns on BUILD inputs (bison CVE-2026-56389
   8.6 high, coreutils, gcc …). Document that build-input warnings are
   out of triage scope (or extend the triage CLI) so the noise has an
   owner (f23).

## f) Up to 50 things to get done next

Brainstorm ranked by impact (per skill: larger N = brainstorm, HARVEST
routes with rigor). Impact/Critical-…-Low, Effort S (<30m) / M / L (>2h).

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Push-lag policy: decide the threshold after which a silent daemon push lag is treated as broken (stall self-healed after ~45 min; see d1), then land f17/f18 so it can never bite mid-ritual | High | S | Ops |
| 2 | OWNER: cut signed tag v2.7.0, ride the runbook (stack lock bump → gates incl. browser E2E → aarch64 → relock #5 + re-pin → deploy `nixos-rebuild test` → switch) | Critical | L | Release |
| 3 | OWNER: post-deploy `webphone-smoke.py --base https://pbx.artmann.tech --expect-version 2.7.0` | High | S | Release |
| 4 | OWNER: journal `telnyx-webhooks` (today), grep sms/422/error, restart/fix creds, send test SMS, record root cause in TODO row + stack runbook | High | S-M | Ops |
| 5 | OWNER: post-deploy sign-in rejection-banner check (self-review item 17, g3) | Medium | S | Verify |
| 6 | OWNER-calls batch session — the ~27 queued decisions (f24–f50 are its agenda) | High | M | Decisions |
| 7 | erraudit tier-2: convert `internal/config.go` seam (22 stdlib_constructor findings) to go-error-family | High | M-L | Quality |
| 8 | erraudit tier-2: convert `store/messages.go` seam (20 findings) | High | M-L | Quality |
| 9 | erraudit tier-2: convert `pbx/client.go` seam (8 findings) | Medium | S-M | Quality |
| 10 | Re-measure erraudit tier-2 on 2026-10-22; update the AGENTS baseline (must shrink from 113) | High | S | Quality |
| 11 | Reconcile buildflow full-mode behavior vs AGENTS claim (gitleaks/codespell/markdown-lint skipped) — fix AGENTS or config | High | S | Docs/Tooling |
| 12 | Resolve the 3 markdown-lint warnings (self-review item 8) — after #11 unblocks the lint | Low | S | Docs |
| 13 | Fix TODO_LIST header overclaim ("every claim checked" → true scope; item 6) | Medium | S | Docs |
| 14 | Fix FEATURES Browser-E2E row (195s/184s greens with FOUC scenario) | Medium | S | Docs |
| 15 | Refresh FEATURES aarch64 row: "cross-builds cleanly (verified 2026-09-26, ELF-verified)" | Low | S | Docs |
| 16 | docs/lessons.md: the `--no-link` + `result` artifact trap; the `--system` restricted-setting note; the daemon-push-stall story | Medium | S | Docs |
| 17 | release.sh preflight: assert remote sync (fail loudly on unpushed commits) | High | S-M | Tooling |
| 18 | Daemon push watchdog (or owner-side check) so push stalls alert instead of rot | Medium | M | Tooling |
| 19 | OWNER (host root): durable binfmt fix — `boot.binfmt.emulatedSystems = [ "aarch64-linux" ]` or drop `/run/binfmt` from `extra-sandbox-paths`; kills the GC/reboot rot | High | M | Host |
| 20 | Understand nix 2.34 `--system` restricted-setting semantics; write the one-lesson note (b2) | Low | S | Tooling |
| 21 | `/version` enrichment: commit/dirty/commitDate via ldflags + ReadBuildInfo vcs settings; update smoke + error-contract docs (item 9) | Medium | M | Feature |
| 22 | Next dep sweep: ride templ-components v1.19.3 patch + regenerate `/assets/tw.css` per the verdict recipe | Medium | S | Deps |
| 23 | Vulnix scope note: build-input warnings (bison 8.6 etc.) are out of triage scope — document or extend `webphone-vulnix-triage` | Low | S | Docs |
| 24 | OWNER: ratify train-cut/cadence | Medium | S | Decision |
| 25 | OWNER: `/livez` consumer | Low | S | Decision |
| 26 | OWNER: HSTS maxAge | Low | S | Decision |
| 27 | OWNER: XFF sanitization (flip `KeyExtractorFromClientIP` once stack proves it) | Medium | S | Decision |
| 28 | OWNER: disclosure posture for announcements | Medium | S | Decision |
| 29 | OWNER: loopback `delivered` semantics | Low | S | Decision |
| 30 | OWNER: handler dual-layer ratification | Low | S | Decision |
| 31 | OWNER: gh-release habit | Low | S | Decision |
| 32 | OWNER: Go module v2 policy | Low | S | Decision |
| 33 | OWNER: recordings intent | Low | S | Decision |
| 34 | OWNER: TEMP-DIAG keep | Low | S | Decision |
| 35 | OWNER: oops ratification (g1 force-push) | Low | S | Decision |
| 36 | OWNER: `backup.retentionDays` | Medium | S | Decision |
| 37 | OWNER: ratify shipped stack `crm.{enable,url,tokenFile}` shape (`be876ae`) | Medium | S | Decision |
| 38 | OWNER: multi-contact "+N more" display | Low | S | Decision |
| 39 | OWNER: English-only journal bodies | Low | S | Decision |
| 40 | OWNER: self-send train-C semantics ratification | Medium | S | Decision |
| 41 | OWNER: templ-components history-blemish disposition | Low | S | Decision |
| 42 | OWNER: release.sh load-gate default ratification | Low | S | Decision |
| 43 | OWNER: g2 KVM-timeout policy | Low | S | Decision |
| 44 | OWNER: art-dupl `-t 3` baseline ratification | Low | S | Decision |
| 45 | OWNER: suppression-bucket doc | Low | S | Decision |
| 46 | OWNER: webhook 400 body-text dependents | Low | S | Decision |
| 47 | OWNER: `msg/`→`message/` idem-key rename | Low | S | Decision |
| 48 | OWNER: helper micro-test bar | Low | S | Decision |
| 49 | OWNER: missed-call REJECT semantics | Low | S | Decision |
| 50 | OWNER: search `?q=` URL semantics + store `Must*` panic-on-corrupt policy (pair) | Low | S | Decision |

HARVEST note (per skill): #1, #11-#18, #20-#23 are TODO_LIST material
(assistant-actionable); #2-#6, #19, #24-#50 are owner rows/agenda;
none are ROADMAP-only except possibly #18/#20 if judged tooling
long-shots. Deferred per "WAIT FOR INSTRUCTIONS".

## g) Three questions I cannot figure out myself

1. **The push stall**: it self-healed (~45 min, then remote caught up
   to `0230ead`), so nothing is blocked — but I cannot see the daemon's
   internals (it is outside this repo and outside Crush). Do you know
   why it stalled (push timer? transient network/auth?), and what lag
   threshold should future sessions treat as BROKEN rather than
   slow — 10 minutes? An hour? That number decides whether f17/f18 are
   nice-to-have or load-bearing.
2. **v2.7.0 ritual split**: the sheet is written as 100%
   owner-terminal. Do you want it to STAY that way, or should a future
   assistant session pre-cut the signed tag + prep the stack lock bump
   so your terminal time is only §1-§3? (I deliberately touched
   neither: running half the sheet's steps would break its copy-paste
   assumptions mid-ritual.)
3. **Has the outbound SMS path been human-tested on prod since the
   2.6.0 deploy went live (2026-09-25)?** The 422 root-cause row
   dates from 2026-09-19, before `1d53f44` was deployed; if you've
   sent a successful test SMS since, the owner-action narrows to a
   journal-only confirmation and the row can honestly shrink. I cannot
   answer this — sending requires your PBX session/device; my probes
   are unauthenticated.

---

*Point-in-time snapshot; goes stale by design. Section (f) is the
HARVEST input. Format override on record: user explicitly requested
`.md` at `docs/status/` (skill default is HTML); no manual commit made
(harness forbids unprompted commits — the auto-commit daemon picks this
file up).*
