# Status — v2.13.0 RELEASED; B5 staged to the last byte; main red on a foreign treefmt drift

**Claims verified as of 00:40 CET, 2026-11-11.** Push/CI statements rot fastest; CI verdicts
were read from `gh run list` minutes ago, the prod probe ran ~35 minutes ago, the aarch64
toplevel build was still running at snapshot time.

## a) FULLY DONE

1. **v2.13.0 IS RELEASED.** Attempt 11 completed GREEN (`release-rc=0`, log
   `/tmp/release-2.13.0-11.log`): step 7 stack VM suites passed ("all checks passed!",
   including `vm-test-run-telephony-browser` — the browser E2E obligation for the C1/E2/C2
   served-markup changes is therefore satisfied against the pinned code), step 8 cross-built
   aarch64 with the ELF `b700` guard green, step 9 cut the GitHub release.
2. **The release artifact verified:** `gh release view v2.13.0` exists with 60 lines of
   notes (the CHANGELOG 2.13.0 section, non-empty, correct fold), tag object `d8e982da` on
   origin pointing at commit `55069a9d` (signed, verified). The release URL:
   `https://github.com/LarsArtmann/webphone/releases/tag/v2.13.0`.
3. **The seven-attempt war is over.** Attempts 5–11 ledger: lychee heal (5), lock-move-guard
   ×2 (6, 8), annotate regression (7), load guard ×2 (9, 10), green (11). The
   manual-relock-first pattern (marker + old→new + CHANGELOG mention BEFORE release.sh) is
   proven: attempt 11's step 7 found no lock diff and skipped cleanly.
4. **My CI red root-caused and fixed.** `de63b78e` (my 23:32 report push) failed the
   docs-health battery: annotate counted **205 findings vs the 203 pin**. Root cause: TWO
   rule-F glue sites in my own report (lines 14–15, code spans dprint had stripped spaces
   around at wrap points — CI's `--check` mode counts the RAW bytes, my local battery ran
   fmt first and counted repaired bytes). Fixed by rewording the spans away (the documented
   rule-F corollary). The G/uniform and A/evidence findings in the failure log are
   pre-existing archived-corpus items inside the ratchet — left alone (closed history).
5. **Cadence healed in the same push.** The other session's 23:44 report pushed the
   unarchived count to 66 > 65. Archived the oldest GATE-CLEAN report
   (`2026-11-08_22-46_superb-phase5-midflight-status.md`) gates-first (check-rows complete +
   bare-rows clean BEFORE the `git mv`; banner added; INDEX regenerated). The 18-day-old
   `2026-10-23_18-08` report was evaluated and is NOT archivable: 102 bare rows across 13
   tables, zero adjudication — it owes a real drain train.
6. **All docs gates green** after the fix: annotate exactly 203 (at pin), cadence 65/65,
   freshness/audit-all/routing/strike-safety/table-damage PASS, fmt canonical — verified in
   `--check` mode (CI parity) AND normal mode.
7. **C2 stale-prod watcher verified LIVE:** the `watch` workflow's third step
   (`stale-prod-watch (prod /version vs the newest release)`) ran green at 22:49 UTC and in
   subsequent watch runs. Correct semantics observed: before the release existed it reported
   NONE; now (release exists, prod still v2.12.0, drift ≪ 24 h) it correctly files nothing.
   No `stale-prod` issues exist.
8. **B5 preflight is complete and clean** (details in §b/§e): pin chain triple-verified,
   drift probe run, prod pre-state captured, x86_64 toplevel built with the ExecStart MOVED
   and the rendered-config secret audit clean.
9. **Stack CI green** on `278dc3d`; pbx-artmann has no Actions workflows (HTTP 404 —
   nothing to verify there; noted, not a failure).

## b) PARTIALLY DONE

1. **B5 — the prod switch itself NOT executed.** Everything before the switch is done:
   - Pin chain: pbx-artmann lock pins telephony `c0b7f509` (verified in `flake.lock`);
     telephony's lock pins webphone `bef56520`; **`bef56520` carries
     `webphoneVersion = "2.13.0"`** and the tag commit `55069a9d` IS its ancestor
     (`git merge-base --is-ancestor` YES); C1/E2/C2 code landed ≤ `bef56520` (lineage
     `…3ba8af1b → bef56520`), newer commits docs-only. Prod will run v2.13.0 code.
   - `nix run .#lock-drift-probe`: narHash matches the pinned rev; the pinned rev's CI is
     green; two ADVISORIES: stack main (`278dc3d`) is 2 commits past the pin (relock
     pending, "bump when ready"); our webphone pin `bef56520` differs from the sibling's
     `c5074bf8` — the expected release-relock shape, docs-only delta.
   - Prod pre-state captured: `ExecStart=/nix/store/lcwix5qb…-webphone-2.12.0/bin/webphone`,
     generation `26.11.20261107.f4a03c0` (matches the recorded pre-switch generation),
     arch x86_64, ssh key-auth live.
   - x86_64 toplevel BUILT (`/nix/store/2rpsrzq…-nixos-system-pbx-26.11.20261107.f4a03c0`):
     the closure carries `webphone-2.13.0` and the unit's ExecStart MOVED to
     `/nix/store/4nnvvrx5…-webphone-2.13.0/bin/webphone`; the rendered
     `webphone-config.json` was audited: 7 secret-shaped keys, **ALL file-lanes** under
     `/var/lib/telephony-secrets/`, zero inline values (the eval-time leak guard's
     contract holds in the artifact).
   - aarch64 toplevel: STILL BUILDING at snapshot time (background shell 1F9, no output yet).
   - The switch command (`nixos-rebuild switch --flake .#pbx --target-host
     root@pbx.artmann.tech`) awaits (i) the aarch64 build, (ii) owner authority — asked
     again in §g Q1.
2. **Main is green-for-my-work but RED overall.** My docs fix landed (pushed inside the
   daemon sweep `1c746689`, my push published tip `a52de727` at 23:01 UTC) and the docs
   gates pass on it — but main has since moved twice more (daemon sweeps of the OTHER
   session's in-flight refactor) and CI is RED on the last FOUR tips
   (`a3832e4a`, `a52de727`, `a0ac6024`, `346cad88`): **treefmt-check** fails because the
   daemon-swept refactor files carry non-canonical Go formatting (gofmt wants an extra
   space before a `//cqrs-lint:ignore(C023)` comment in readmodels' `Shutdown`, plus
   comment-alignment diffs in the fax/status files). The one-command heal is `nix fmt` over
   those files — NOT applied: the refactor is the other session's ACTIVE train and racing
   its formatter vs leaving main red is a judgment call (§g Q2). No red-main tracker issue
   visible in the issue search — the watcher's suppression needs a look.
3. **The post-release relocks are identified, not executed:** stack → webphone main tip
   (freeze lifted at release), pbx-artmann telephony → `278dc3d` (drift probe says bump
   when ready). Both are marker+CHANGELOG rituals away.

## c) NOT STARTED

1. The post-switch verification battery (§f items 2–9) — blocked on the switch itself.
2. The TODO_LIST sweep (§ "Trunk truth + v2.13.0 release train — open rows" rows 12–25,
   26–32 tooling, 39–40 cadence, 46–50 owner decisions) — the §f ledger of the 23:32
   report mirrors them.
3. The release-log archival (row 32: `/tmp/release-2.13.0-{5..11}.log` → evidence home) —
   /tmp is wiped between sessions; STILL NOT DONE.
4. The 18-08 report drain train (102 rows) and the 19:14 report's 27 bare rows.
5. Owner-dependent legs: A3.4 Telnyx reply send, C3.2 inbound mobile call, E3 budget
   number → e3.2, E5.2–E5.4 hikaru legs, E6 first real fax.
6. The monthly erraudit tier re-measure (AGENTS pin: next 2026-11-05 — now OVERDUE).
7. Post-release stack relock + pbx-artmann telephony bump (§b3).
8. Attribution pass on the other session's refactor when it lands (split files
   `threadsregister/threadsstate/threadsviews` — fleet-libs reopen-trigger check).

## d) TOTALLY FUCKED UP (honest damage report)

1. **I shipped a CI red myself.** My 23:32 report carried two rule-F glue sites; local
   battery green (fmt repaired before counting) vs CI check-mode red (raw bytes counted).
   Red window on main ~22:45–23:01 UTC for MY cause. Lesson: the `--check` battery is the
   ONLY pre-push docs verdict; I used the mutating battery. Fixed this session; the check
   now lives in my pre-push ritual.
2. **The daemon race got me on the push.** I verified commit `1c746689` (build + cqrs/
   domain/readmodels tests green in a worktree), but between verification and `git push`
   the daemon swept ANOTHER 28-file commit (`a52de727` — the other session's refactor
   spreading into domain/readmodels/server), and my push published it unverified locally.
   No data damage; CI adjudicated (and flagged the formatting drift). Lesson: capture the
   HEAD sha after verification and assert it unchanged immediately before push.
3. **Two wasted background builds (wrong cwd).** The first toplevel builds ran in the
   webphone checkout (background shells inherit my cwd, not the target repo's) and failed
   with a flake-attribute error; re-ran with explicit `working_dir`. Cost: minutes, nothing
   else.
4. **The edit tool fought the report file twice** (stale-read guard + literal
   backslash-backtick escapes in the glue sites); the byte-precise python replace with
   uniqueness assertions was the right tool from the start. Cost: two failed tool calls.
5. **Nothing was destroyed.** No reverts, no force, no lost work; prod untouched; the two
   unarchived dirty reports are intact; the other session's in-flight edits remain
   unstaged and untouched.

## e) WHAT WE SHOULD IMPROVE

1. **Pre-push docs verdict = `--check` mode, always.** The mutating battery certifies a
   tree CI never sees. (Candidate: fold a `--check`-mode annotate count into the pre-push
   ritual or a hook; the battery costs ~1 min.)
2. **Verify-then-push needs an atomic guard:** record `HEAD` after verification; re-read
   it in the same call as the push; abort if it moved. The daemon will keep racing.
3. **Filing a report means cadence accounting in the SAME push:** +1 report ⇒ archive one
   gate-clean report or the gate reds CI on your own push. This session did it AFTER the
   red taught it; make it the ritual.
4. **The archive candidate scan should be a script** (check-rows --strict + bare-rows over
   all live reports, oldest-first CLEAN list) — I hand-rolled it in bash this session;
   it is one `scripts/cadence-candidates.sh` away from being a ritual step.
5. **The 18-08 report must be drained before ~2026-12-07** (45-day cadence age threshold)
   or cadence reds on AGE, not count — a bigger, scheduled train, not a scare fix.
6. **Daemon-swept foreign files need a formatting gate at sweep time** — the treefmt red
   class keeps landing via `chore: auto-commit` sweeps of another session's in-flight
   files (the runbook documents it; the daemon bypasses hooks, so the guard stays
   advisory — CI is the only real gate, hence the red).
7. **Release-log preservation at attempt time:** each release.sh attempt should copy its
   log into the evidence home on completion, not "later" (row 32 still open; /tmp rots).
8. **Red-main watcher visibility:** four consecutive red tips and no tracker issue visible
   in the search — verify the LIST-API dedup is matching (the titles carry daemon commit
   subjects, which change every sweep, so the suppression may be missing).

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

| #  | Item                                                                                                                    | Owner leg?   |
| -- | ----------------------------------------------------------------------------------------------------------------------- | ------------ |
| 1  | Answer §g Q1–Q3                                                                                                         | yes          |
| 2  | Finish aarch64 toplevel build (running)                                                                                 | no           |
| 3  | Execute the B5 prod switch (`nixos-rebuild switch --flake .#pbx --target-host root@pbx.artmann.tech`)                   | gated by Q1  |
| 4  | Post-switch: prod `/version` = `v2.13.0`                                                                                | no           |
| 5  | Post-switch: ExecStart store path moved + generation advanced (record both)                                             | no           |
| 6  | Post-switch: journal shows NO unknown-config-key WARNs (the pin-predates-key healthy-quiet check)                       | no           |
| 7  | Post-switch: `/healthz` 200, `/metrics`+`/version` fencing unchanged, `/livez`+`/startupz` 200                          | no           |
| 8  | Post-switch: SIP registration + `/events` SSE + passkey lane live on prod                                               | no           |
| 9  | Post-switch: dialplan ring hook still 202s (stack side)                                                                 | no           |
| 10 | Post-switch: `ui-capture.py --base-url https://pbx.artmann.tech --track` — Reports NoAnswer-causes subline on real data | no           |
| 11 | Post-switch: voicemail click verdict (one honest toast per episode) on prod bytes                                       | no           |
| 12 | Post-switch: `compare.html` pixel diff vs the pre-switch run                                                            | no           |
| 13 | Post-switch: post-deploy smoke probe                                                                                    | no           |
| 14 | pbx-artmann deploy-ledger row (generation, rev, timestamp, verdict)                                                     | no           |
| 15 | Verify webphone CI green on the treefmt-healed tip                                                                      | no           |
| 16 | Heal or delegate the treefmt red (Q2)                                                                                   | gated by Q2  |
| 17 | Check red-main watcher suppression (no issues visible through 4 red tips)                                               | no           |
| 18 | Relock stack webphone input → main tip (freeze lifted), marker + CHANGELOG + push                                       | no           |
| 19 | Bump pbx-artmann telephony → `278dc3d` (drift probe: pending), `lock-drift-probe` re-run                                | no           |
| 20 | Archive the release attempt logs `/tmp/release-2.13.0-{5..11}.log` into the evidence home                               | no           |
| 21 | Drain-train the 18-08 report (102 bare rows, 13 tables) before the 45-day age threshold (~2026-12-07)                   | no           |
| 22 | Adjudicate the 19:14 report's 27 bare rows                                                                              | no           |
| 23 | Verify C2 watcher's next cycle post-switch (NONE verdict with v2.13.0 == v2.13.0)                                       | no           |
| 24 | Watch the other session's cqrs refactor land complete; attribute full-suite verdicts before acting                      | no           |
| 25 | Fleet-libs reopen-trigger sweep after the refactor lands (cqrs-htmx usage re-scored)                                    | no           |
| 26 | Monthly erraudit tier re-measure (OVERDUE since 2026-11-05; tier 1 exit 0, tier 2 = 0)                                  | no           |
| 27 | E3 budget number → then implement e3.2 against it                                                                       | yes (number) |
| 28 | A3.4 Telnyx reply send (end-to-end outbound SMS on prod)                                                                | yes          |
| 29 | C3.2 inbound mobile call (real phone rings the PBX)                                                                     | yes          |
| 30 | E5.2–E5.4 hikaru legs (external host work)                                                                              | yes          |
| 31 | E6 first real fax (capability declared; needs a real send)                                                              | yes          |
| 32 | Runbook: add the verify-HEAD-before-push atomic step                                                                    | no           |
| 33 | Runbook: add the `--check`-mode battery as the pre-push docs verdict                                                    | no           |
| 34 | Runbook: add "archive one gate-clean report when you file one" to the report ritual                                     | no           |
| 35 | Script the cadence-candidate scan (`scripts/cadence-candidates.sh`)                                                     | no           |
| 36 | Lower the annotate ratchet below 203 only via a real drain (drain-first, then re-measure)                               | no           |
| 37 | Verify the vendored island bundle untouched by the refactor (island tests in CI green)                                  | no           |
| 38 | Confirm the release notes render correctly on the release page (fold anchors)                                           | no           |
| 39 | Update TODO_LIST § open rows: strike the rows this session closed (release cut, docs gates)                             | no           |
| 40 | CHANGELOG: nothing owed — docs/test-only deltas stay changelog-silent (verify no code delta slipped)                    | no           |
| 41 | Check `.docs-gates.json` cadence pin still 65 (it is; re-measure at next sweep)                                         | no           |
| 42 | Post-refactor: re-run `art-dupl check` (new clones only) on the split cqrs files                                        | no           |
| 43 | Post-refactor: arch-test suite green (`internal/arch`) — module graph + error-code registry                             | no           |
| 44 | Android repo: confirm no wire-contract drift this train (C1/E2 changed no wire types)                                   | no           |
| 45 | Backup posture: confirm the switch generation still rides the backup set (system.db journal world)                      | no           |
| 46 | Verify `stale-prod-watch.sh --self-test` still 10/10 (no changes made — hygiene check)                                  | no           |
| 47 | Session hygiene: rebuild `/tmp/annotool/annotate` at next archive ritual (/tmp wiped)                                   | no           |
| 48 | Consider recording the two-session freeze protocol as an AGENTS.md rule (pending Q on binding)                          | yes          |
| 49 | Roadmap: next release train candidate list (what rides v2.14.0 — the refactor + any E-items)                            | no           |
| 50 | After the drain trains land: re-measure `audit_max`/`annotate_max` from the corpus (the pins are snapshots)             | no           |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **B5 authority + timing:** my blanket GO reading says "execute everything" — but this
   is the live phone system and the other session is mid-refactor on the same machine.
   **Do I execute the prod switch over ssh as soon as the aarch64 toplevel lands, or stage
   it and hand off?** (Pre-switch generation, rollback path, and the full verification
   battery are ready.)
2. **Main's treefmt red:** four consecutive red tips, the drift is the OTHER session's
   daemon-swept refactor files. **Do I heal main now (`nix fmt` over the swept files +
   push, racing their editor state), or leave main red until that session finishes its
   refactor and self-heals?**
3. **E3 budget number:** still unanswered from the last report — the Reports/latency
   budget figure only you know; e3.2 is blocked on it.

## Session evidence index

- Release: `/tmp/release-2.13.0-11.log` (`release-rc=0`), gh release v2.13.0 (60-line
  notes), tag `d8e982da` → `55069a9d`.
- CI: webphone run 42917466259 (de63b78e RED, annotate 205>203), 42918196739 (75195343
  RED, same cause), 42919746129+ (a52de727+ RED, treefmt-check), stack green on
  `278dc3d`; watch green at 22:49/23:19/23:35 UTC.
- B5: pbx-artmann lock telephony `c0b7f509`; webphone `bef56520` = v2.13.0 (tag ancestor);
  drift probe advisories ×2; prod ExecStart `lcwix5qb…-2.12.0`, gen `26.11.20261107.f4a03c0`;
  x86 toplevel `2rpsrzq…` with ExecStart `4nnvvrx5…-webphone-2.13.0`; config secret audit
  7/7 file-lanes; aarch64 toplevel building (shell 1F9).
- Docs: 23-32 report reworded (2 glue sites), 22-46 report archived gates-first, INDEX
  regen, all gates green (annotate 203 @ pin, cadence 65/65).
