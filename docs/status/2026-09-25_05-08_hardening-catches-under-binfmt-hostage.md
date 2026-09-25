# Status report — hardening-and-fixes run under the binfmt hostage: 4 release.sh catches, the shared-contact wire fix, and what I got wrong doing it

Generated 2026-09-25 05:08 CEST (Friday). Scope: THIS session —
04:32→05:08, resuming from the 04:04 report (whose three open
questions are now all answered; see the addendum there). Directive
was "get shit done" with a full nix-free queue. Webphone HEAD
`5ddd59e` == origin/main (verified); tree clean. The host-nix outage
(`/run/binfmt`) is STILL down at 05:08 — 10 hours — and remains the
single blocker on the v2.7.0 release tail. A detect-only watcher
(30s interval) runs until ~08:33.

---

## a) FULLY DONE (committed, pushed, verified)

1. **Shared-contact wire fix — the real product bug of the night**
   (`e43fea8`): `domain.SharedContact` carried no json tags, so
   `window.PBX_CONFIG` shipped Go-style `Name`/`Number` while the
   island typeahead reads lowercase `name`/`number` (README always
   documented lowercase) — every operator-configured shared contact
   was silently invisible in dial suggestions. Verified end-to-end
   before touching anything (island `config.js` passes the array
   through verbatim; `typeahead.js` reads `.name`/`.number`). Tags
   added; `TestConfigJSContactsWireKeys` pins the wire (quote-in-name
   escaping + capitalized-absent asserts); full `go test -count=1
   ./...` green on the exact tree via store go; CHANGELOG 2.7.0
   Fixed section folded; announcement drafts (all 3 variants) carry
   it. This re-does the fix the 22:35 reboot destroyed in the stack
   session's /tmp clone (their TODO row 36 handed it upstream).
2. **Stack side flipped in lockstep** (`dea945c`, pushed):
   `tests/webphone.nix` contacts asserts capitalized→lowercase +
   comment rewritten; TODO row 36 BLOCKED→IN_PROGRESS with the
   red-at-old-lock-by-design note (safe window: nobody can run stack
   gates without nix, which is down). The stack's untracked 04:05
   report swept in the same commit (its daemon was stalled since
   19:00). Stack has since advanced to `2be7601` (concurrent
   session's doc curation; 3 files dirty in their tree — theirs).
3. **release.sh hardened ×4** (all bash -n clean; preflight arms
   live-tested; each committed + pushed):
   - `fe09106` binfmt preflight in step 1 — verified live: a real
     invocation with a bogus version dies AT the check printing the
     exact root fix. Dry-run skips it (works on a broken host).
   - `c70ec15` the preflight/owner-sheet/lessons remediation
     CORRECTED: `systemctl restart systemd-binfmt` heals nothing
     (unit finished OK at the reboot; the generation's tmpfiles
     carries no binfmt rules). Verified fix derived from live
     forensics: kernel entry `aarch64-linux` names
     `/run/binfmt/aarch64-linux`; nix.conf pins store path
     `31rksc1…-qemu-aarch64-binfmt-P` whose binary is
     `bin/qemu-aarch64-binfmt-P` (the stack report's glob missed the
     trailing `-P` — I globbed and binary-checked it myself).
     Preflight now interpolates the live store path into its
     message.
   - `3d4af5c` step 4's bare `go test` → `nix develop -c go test`
     (same trap class as attempt 1; smoke + buildflow already
     self-heal).
   - `9f7badd` **the empty-release-notes catch**: the 2026-09-20
     "escaped brackets" awk fix only works on awks that PRESERVE
     unknown escapes — host gawk 5.4.1 warns and DROPS the backslash,
     collapsing the dynamic regex back into the one-char bracket
     class. Reproduced: the script's exact invocation extracted **0
     lines** for 2.7.0 on this host. v2.7.0 would have published
     EMPTY notes (the v2.3.0/v2.4.0 bug class). v2.6.0's published
     notes are fine (14954 chars — its release ran on an awk that
     preserved the escape). Fixed with a literal `index() == 1`
     prefix match (identical in every awk; verified 75/243 lines for
     2.7.0/2.6.0) + a belt-and-braces guard refusing empty notes.
     Audited every other awk in scripts/ — clean (no escapes).
4. **CHANGELOG cosmetics + links** (`37d6679`): stray double blank
   in 2.7.0 Changed removed; `[Unreleased]` compare base moved to
   v2.7.0; `[2.7.0]` release-tag footer link added.
5. **Store-binary outage fallback codified** (`37d6679`): full war
   story in lessons.md §Nix (outage mechanics + the
   GOTOOLCHAIN=local store-go pattern + the ldflags variable name)
   with an AGENTS.md commands-section pointer.
6. **Docs synced to new reality**: AGENTS stack-train line refreshed
   (7197f1c → 94ae28d forward-lock + pending v2.7.0 relock); owner
   sheet §1 rewritten — prod was deployed to **v2.6.0 early
   2026-09-25** (concurrent session probed + recorded), settling the
   deploy-ordering question: straight to 2.7.0; sheet §3 gained the
   one-line typeahead post-deploy check; status-report addendum
   (`5ddd59e`) answers all three g-questions from the 04:04 report.
7. **Remote-sync discipline restored**: both push daemons stalled
   (webphone ~04:07, stack ~19:00); I pushed manually 6× (daemon
   commits rode along), verifying every end state via ls-remote.
   Full dry-run of the hardened release.sh validated ALL 9 steps at
   `37d6679`-time (preconditions → gates → tag → lychee → stack →
   aarch64 → gh).

## b) PARTIALLY DONE

1. **M19 release tail — still hostage.** Everything before the tag is
   green/verified; the dry-run proves the script path; the resume
   command is unchanged:
   `nix develop -c scripts/release.sh 2.7.0 > /tmp/release-2.7.0-2.log 2>&1`
   (background, monitor by tailing the file). Blocked ONLY on the
   owner root fix (sheet §0). Load was 7.5–11 during 04:50–05:00
   (four concurrent crush sessions) — the script's own load gate
   handles that; it was 2.7 by 05:08.
2. **Dry-run completeness debt:** the last two release.sh edits
   (go-test wrap, notes extractor) got bash -n + targeted live tests
   (preflight arm, extractor against both versions) but the FULL
   dry-run never re-ran to completion afterward — the tree kept
   going dirty under concurrent sessions between my commit and the
   run. The `run` lines changed are print-only in dry-run; risk is
   minimal, but the pass is owed before attempt 2.
3. **M22 tail (E2E wall times into the watches row)** and the
   sheet's `<STACK_HASH>` fill — both structurally post-release.

## c) NOT STARTED (and why)

1. **Owner-terminal items** (not assistant-startable by AGENTS): the
   root binfmt fix itself, prod deploy of 2.7.0, pbx-artmann relock
   #5, live self-send + typeahead checks, telnyx-webhooks journal
   triage, announcement approvals, owner-calls batch, and the
   durable host fix (`boot.binfmt.emulatedSystems` — kills the
   failure class; sheet §0).
2. **Auto-launch of the release on heal** — deliberately NOT built:
   detect-only watcher chosen (unattended tag+E2E with four other
   sessions churning, and the load gate would fail-fast anyway;
   decide per g2 below).
3. **Smoke/island re-stamp on the `f70cfa4` tree** (the concurrent
   session's tw.css regen + all-view templ regeneration, 14.5→18.9KB):
   my 04:44 full go suite covers compile+unit on that tree, and the
   release runs smoke anyway; a pre-emptive belt is optional.
4. **Verdict-doc tw.css drift**: their regen updated AGENTS+CHANGELOG
   to 18.9KB, but the tailwind-coexistence verdict doc still says
   14.5KB. Their train, their call — not touched (concurrent-session
   rule).

## d) TOTALLY FUCKED UP (all caught in-flight, recorded for the pattern)

1. **I propagated stale remediation past evidence I had ALREADY
   read.** The stack session's 04:05 report (read by me ~04:39)
   documented that systemd-binfmt finished OK at the reboot and the
   generation lacks tmpfiles rules — yet my ~04:45 lessons.md entry
   and the release.sh preflight message still said `systemctl
   restart` (copied from the prior session's owner sheet). Caught by
   my own re-review ~10 min later; corrected in all three homes +
   the addendum. Root cause of the miss: transcribing from an earlier
   artifact instead of re-deriving from the freshest evidence.
2. **My original §0 shape claim was invented**: "/run/binfmt
   (symlink → /proc/sys/fs/binfmt_misc)" — wrong shape entirely (it
   is a directory holding interpreter symlinks). Written from
   assumption; fixed only during the correction pass.
3. **Pipe-exit trap twice**: `go test | grep -v …; echo EXIT=$?`
   (grep's exit), then `release.sh … | tail; echo exit=$?` (tail's
   exit, showing 0 for a script that exited 1). Both redone
   log-to-file + raw exit — which is EXACTLY what the release
   runbook already teaches ("log to files, tail them"). I violated a
   written rule twice in an hour.
4. **A verification label lied**: an echo printed "BOTH REPOS
   CLEAN+SYNCED" in the same output where webphone's tree was
   actually dirty (concurrent edits). Narrative corrected it, but the
   label was derived BESIDE the check, not FROM it.
5. **Edit-tool transcription typo** ("was"/"were") cost a retry on
   lessons.md — careless copy from the view output.
6. **The daemon raced 3 of my explicit commits** (content always
   preserved — it swept my staged files with its heuristic message;
   my descriptive message was lost twice). The briefing told me to
   expect daemon contention; I used the edit tool throughout (worked,
   but commits need staging IMMEDIATELY after edits to win the race).

## e) WHAT WE SHOULD IMPROVE

1. **Verify-before-trust extends to COMMENTS claiming a fix exists.**
   The awk extractor carried a comment saying the bracket bug was
   fixed in 2026-09-20; I trusted it for hours and only tested the
   seam while idling — 0 lines. Same failure class as M18
   (build-on-assumed-behavior), one night later, caught by luck of
   having idle poll time. Reflex: test the seam the moment it matters,
   not when convenient.
2. **Fresh forensics beat prior artifacts**: before writing any
   remediation, re-derive it from the newest evidence in-context —
   d1 above was 100% avoidable.
3. **`PIPESTATUS[0]` or log-to-file for every exit-code assertion**;
   the runbook rule exists — bind it outside release nights too.
4. **Commit immediately after edits** (stage while editing) to beat
   the daemon's sweep window; python-keyed inserts for contended
   files remain the briefing's advice and remain correct.
5. **Derive verification labels from check output**, never beside it.
6. **Test new shell/awk logic in the same breath as writing it** —
   the extractor test took 10 seconds; the near-miss cost an hour of
   lucky timing.
7. **Candidate for lessons.md** (not yet written): the
   escaped-bracket awk portability war story + the
   comments-claim-fixes trust trap (one combined Tooling-traps
   bullet).

## f) Next things (ranked; 1–10 unblock/complete the release)

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 1 | OWNER ROOT: heal /run/binfmt (sheet §0 — mkdir + interpreter symlink; restart heals NOTHING) | Critical | S |
| 2 | Re-run the full release.sh dry-run to completion (owed after the last 2 edits; needs quiet tree) | High | S |
| 3 | Launch attempt 2: `nix develop -c scripts/release.sh 2.7.0 > /tmp/release-2.7.0-2.log 2>&1`, tail the file, run NOTHING else during E2E | Critical | M |
| 4 | E2E ×2 green → wall times into the watches row (M22 tail, budget 445s) | High | S |
| 5 | aarch64 cross-build + ELF b700 byte verify (step 8) | High | S |
| 6 | gh release v2.7.0 — extractor now verified + empty-notes guard live; then verify notes + tag links resolve | High | S |
| 7 | Post-release closes: stack hash into sheet §1/§2; release TODO rows (webphone + stack IN_PROGRESS row); AGENTS train line to final state | High | S |
| 8 | Closing sweep per runbook step 9: one no-cache buildflow via scripts/buildflow.sh, ls-remote verify, dead-process proofs (incl. my watcher) | High | S |
| 9 | Re-grep "spike" at the tag (release-tree spike-free proof) | Medium | S |
| 10 | OWNER: deploy 2.7.0 straight over 2.6.0 + smoke `--expect-version 2.7.0` | High | M |
| 11 | OWNER: live checks — self-send 422 local refusal + typeahead shows shared contacts (sheet §3) | High | S |
| 12 | OWNER: pbx-artmann relock #5 + re-pin (sheet §2) | High | M |
| 13 | OWNER: telnyx-webhooks journal triage (row 40, prod SMS bridge) | High | S |
| 14 | OWNER: announcement approvals v2.1.0–v2.7.0 (drafts at docs/announcements/) | Medium | S |
| 15 | OWNER: owner-calls batch (unblocks the parked policy list) | High | M |
| 16 | OWNER durable: `boot.binfmt.emulatedSystems` in host config (or drop `/run/binfmt` from extra-sandbox-paths) — kills the class + the GC-rot risk of the hard store pin | High | S |
| 17 | lessons.md: the awk-portability + comments-claim-fixes bullet (e7 above) | Medium | S |
| 18 | Root-cause or alert on the stalled push daemons (webphone 04:07→, stack 19:00→; I pushed manually 6×) | Medium | S |
| 19 | Optional belt: smoke + island node tests re-stamp on the `f70cfa4` tree before the release does it | Low | S |
| 20 | Verdict-doc tw.css drift (14.5 vs 18.9KB) — coordinate with the concurrent session that regenerated it | Low | S |
| 21 | templ-components v1.19.3 patch ride at the next dep sweep (vendorHash same-breath rule) | Low | S |
| 22 | tw.css regen on every templ-components bump (recipe in the verdict doc) | Low | S |
| 23 | erraudit tier-2 re-measure due 2026-10-22 (baseline 127/113, must shrink) | Medium | S |
| 24 | Quarterly watches re-check due 2026-12-20 | Low | S |
| 25 | Post-2.7.0: restart [Unreleased] discipline in CHANGELOG | Low | S |
| 26 | M18 park watch: revisit if a go-health release makes StartupHandler Evaluate (parked with evidence, TODO row 48) | Low | — |
| 27 | Send-failure F: keep NOT building it unless self-sends recur post-C | Low | — |
| 28 | templ-components v1.20.x watch: RelativeTime only if it gains server-render mode | Low | M |
| 29 | Consider a tiny bash test harness for release.sh's pure functions (extractor, load_gate) — the guard exists but no automated test | Low | S |
| 30 | Watcher expiry ~08:33: re-arm or hand off if the session ends before the heal | Medium | S |

## g) Questions I cannot answer myself

1. **Will you run the root fix now?** Sheet §0, two lines +
   verification (`nix build nixpkgs#hello`). Everything release-shaped
   waits on it; the watcher catches the heal within 30s and logs to
   /tmp/binfmt-watch.log.
2. **Auto-launch or explicit go?** When binfmt heals AND load < 8
   sustained AND the tree is clean: should I launch release attempt 2
   immediately on my own, or wait for your explicit go? (I built
   detect-only tonight; both are one command from my side.)
3. **The push daemons stalled tonight** (webphone ~04:07 onward,
   stack since 19:00 — I pushed manually 6× to keep the release
   prerequisite honest). Known/intended behavior, or do you want the
   daemon's push loop root-caused? I cannot see its internals.

---

*Point-in-time snapshot at 05:08. Session continues polling; resume
state is one paste (f3) once g1 is answered.*
