# Session Status Report — 2026-09-22 18:46 CEST

> CLOSED 2026-09-23 (docs-health): f1–f9 all landed — the ThemeScript
> adoption completed end-to-end (18:55), E2E ×2 green on the fixed
> driver (19:55), final gates + aarch64 + lychee + push-state verified
> (19:55), T26b/T26c shipped (`c1971c4`/`37aae5a`), and the v2.6.0
> train folded + tagged (`807ca0c`); erraudit re-measured (f16).
> Still open, routed: the owner console (TODO rows), daemon disposition
> (ROADMAP infra ask), the release TAIL — stack E2E ×2 on the v2.6.0
> chain + pbx-artmann relock #4 (TODO row), the REGISTRATIONS-0 watch
> (standing, AGENTS flake note).

Third mid-flight report (15:06 → 16:52 → 18:20 → now); covers
18:20–18:46. The standing "whole list" directive kept execution
running after the 18:20 report.

## a) Fully done (verified)

1. **E2E dest-clear fix PROVEN**: run 3 (fixed script) sailed through
   what had stalled for 300s — FS-OUTAGE-READY in 2.5s, and the blind
   transfer EXECUTED on sofia (partner leg re-routed to the 9196 echo
   app, transferer released with NORMAL_CLEARING). The concatenation
   diagnosis was right; the fix works.
2. **CHANGELOG + FEATURES fully synced** for every train this session
   (`4c4b36b`): failed-bubble story, delivered ✓ badge, thumbnails +
   sniffed types, drafts, call state chip, the affordance batch,
   thread-list dial row, retention/`/metrics`/timezone rows.
3. **T27c signed tags** (`4f067f8`): release.sh now cuts `git tag -s`
   and verifies locally before pushing. Bonus verification: the
   v2.5.0 tag ALREADY carries a good ED25519 signature (the daemon's
   key signs tags).
4. **T16d**: the 19-37 integration plan's seven closing-gate
   checkboxes ticked with a dated note — one of them (the logged-out
   dial smoke) is now the permanent LOGGED-OUT-DIAL-GUARDED E2E
   scenario.
5. **T27a nginx gzip** (`9f93537` + `44c0b8e`): `nginx.gzip.enable`
   flips nginx's recommended gzip settings (SSE is never gzipped by
   nginx itself); module-eval stand-in green.
6. **T26a wiring complete on my side**: module `/metrics` vhost
   location + the flake-check location list; the implementation
   itself was converged BY THE CONCURRENT SESSION (they renamed my
   name-colliding `Counts` function to `ReadCounts` and fixed the
   mem-store 500); my aggregates-only test (with the no-extension-
   strings leak pin) passes green against their version.
7. **TODO_LIST harvest** (`34e741d`): the redial-stall class recorded
   as RESOLVED (not a flake — the appended dial), send-failure train D
   marked shipped, the coturn REST-secret row parked for T26b's stack
   half.

## b) Partially done

1. **E2E ×2 green (T20f/g/h)**: run 3 died at the post-transfer
   `show channels` count wait (180s) — the DOCUMENTED transfer-step
   flake mode, re-run-once rule applies. Run 4's first invocation was
   a phantom (see d2); the real run 4 was relaunched foreground-
   captured and its verdict was still pending when this report was
   requested.
2. **Final gates**: BLOCKED — the tree does not compile (see d4).
3. **T26b (TURN) + T26c (export)**: parked; the server/config
   namespace belongs to the concurrent session's active train.

> Resolved 2026-09-22 evening: b2/b3 unblocked — the ThemeScript train
> landed end-to-end (18:55 session: v1.19.2 consumed, gates green) and
> the final gates ran green; T26b/T26c stay TODO rows. d4's unverified
> field was CONFIRMED real (PageProps.NoThemeScript in v1.19.2).

## c) Not started

- ~~T26b webphone half (TURN REST creds in /config.js), T26c
  per-extension export (zip).~~ done (`c1971c4`, `37aae5a`)
  ~~Post-gate ritual (aarch64, lychee,
  smoke, push-state verify).~~ done (19:55) ~~Next-train fold decision
  (v2.6.0).~~ done (folded + tagged `807ca0c`)
- Owner console items (unchanged). ← still open (TODO rows)

## d) Totally fucked up (this window; lessons kept)

1. **My gzip stand-in broke the whole flake**: commas where nix
   attrsets want semicolons — every nix command in the repo failed
   until fixed. AND I had misread the earlier module-check as green
   (`MODULE-RC=0` read from the wrong pipeline position — the pipe
   masked the failure). The same edit also forgot the documented
   module-check ritual: a new config key the module writes needs its
   stand-in option (`recommendedGzipSettings` was undeclared).
2. **Phantom E2E run**: run 4's background job reported "completed"
   with cleanup lines that were STALE output from run 3's derivation;
   I nearly recorded a verdict for a run that never built. Caught only
   by a `--dry-run` probe saying "will be built".
3. **Tooling artifacts cost real time**: a botched `nix log` redirect
   produced a one-line file that sent me digging; two commands wasted.
4. **Premature "invented field" verdict**: the concurrent session's
   `NoThemeScript` layout edit broke the build; I confirmed the field
   is absent from v1.18.0/1.18.1 and called it invented — but my
   v1.19.2 check never actually materialized in the module cache
   (GOMODCACHE confusion under nix develop), so whether v1.19.x ships
   the field is UNVERIFIED. Their edit may be a valid pending bump.
5. Minor: TODO_LIST exact-match anchors failed twice on table
   re-padding (line surgery used instead); a `cd /tmp` inside
   `nix develop` failed the flake lookup.

## e) Improvements

- Read `$?` IMMEDIATELY after the command it belongs to, never after
  a pipe chain; when a check "passes" suspiciously easily, re-verify
  by naming the artifact it should have produced.
- Backgrounded jobs: verify the derivation actually built (dry-run)
  before recording any verdict.
- The module-check stand-in ritual is a CHECKLIST item, not folklore:
  new module-written config key → stand-in option in the same commit.
- Never conclude "field does not exist" until the module-cache lookup
  is confirmed against the right GOMODCACHE.

## f) NEXT — up to 50, in order

1. ~~Verify whether templ-components v1.19.x ships `NoThemeScript`~~ done (v1.19.2 shipped the knob; adoption completed 18:55)
   ~~(right GOMODCACHE / `go doc`); then either complete the adoption —~~
   ~~go.mod bump + vendorHash + flake + DROP the CSP-hash exception and~~
   ~~the `!important` color-scheme rules (the standing watch says take~~
   ~~the knob) — or, if no release has the field, revert their~~
   ~~layout.templ hunk (with their session's awareness).~~
2. ~~E2E run 4 verdict; run 5 if the flake repeats; ×2 green closes~~ done (x2 green 19:55)
   ~~T20f/g/h.~~
3. ~~Final gates on the converged tree: full suite, erraudit,~~ done (green 18:55 + 19:55)
   ~~`BUILDFLOW_NO_RESULT_CACHE=1 buildflow`, `nix flake check`, smoke.~~
4. ~~aarch64 cross-build + ELF `b700` check (runbook step 8).~~ done (ELF b700 verified 19:55)
5. ~~lychee link check over the docs batch.~~ done (0 errors 19:55; 7503561 fixed the later break)
6. ~~Push-state verify all three repos (`git ls-remote`).~~ done (19:55 ls-remote)
7. ~~T26b webphone half: `turn_rest_secret` + TTL, `/config.js`~~ done (c1971c4)
   ~~emission, test (stack coturn row already parked).~~
8. ~~T26c per-extension export (zip: messages JSON, contacts vCard,~~ done (37aae5a)
   ~~fax list) — session-gated, owner-scoped.~~
9. ~~Decide + execute the v2.6.0 fold (the g2 theme + the CRM train's~~ done (tag 807ca0c pushed + gate-verified; stack E2E x2 + pbx relock #4 = release TAIL row)
   ~~bullet are both in [Unreleased]); release via the now~~
   ~~signed-tagging script; stack re-pin + E2E + pbx-artmann relock #4.~~
10. Owner console: deploy decision (v2.5.0 now vs fold-first).
11. Owner console: T4a rejection-banner live check (the NEXT deploy
    also shows the persisted reason in the bubble — T21a).
12. Owner console: T5 SMS-lane journalctl check.
13. Owner console: T11 14-decision batch (briefing doc).
14. Owner console: announcement drafts approval.
15. Daemon disposition (docs/status+planning exclusion ask).
16. ~~erraudit tier-1 re-check after the concurrent train fully lands;~~ done (19:55, tier-1 0 + tier-2 127/113)
    ~~tier-2 recount (bar: 102, must shrink).~~
17. The REGISTRATIONS-0 secondary E2E wedge — still unexplained; the
    dumps are the first stop if it recurs.
18. AGENTS/lessons candidates from today: the phantom-run trap, the
    stand-in checklist, the stale-job-output class, string-surgery
    tax (from the 18:20 report).
19. Monthly erraudit re-measure (2026-10-22); quarterly watches
    (2026-12-20 — the ThemeScript knob watch may close FIRST via f1).
20. ~~Final closing sweep: pgrep my booted processes (none should~~ done (per-session closing sweeps ran)
    ~~survive), one last `git ls-remote`, final report.~~

## g) Questions I cannot answer myself

1. **The ThemeScript adoption is half-landed by the other session**
   (layout.templ uses `NoThemeScript`; the tree does not compile on
   any pinned library version). If v1.19.x ships the knob: I complete
   the adoption and delete the CSP-hash + `!important` machinery —
   or is that their train to finish? If NO release has the field:
   may I revert their hunk to unbreak the tree, or do we wait for
   them?
2. **v2.6.0 fold timing**: [Unreleased] now holds two coherent trains
   (mine: every-surface-answers-back; theirs: Ledger CRM + the
   ThemeScript adoption). Fold both into v2.6.0 once the tree is
   green, or cut separately?
3. **Deploy cadence** (standing, sharpest form yet): prod still serves
   v2.4.0; the verified chain is one owner command away; main now has
   two unreleased trains. Deploy v2.5.0 now, or ship v2.6.0 first and
   deploy once?

— Session paused here per instruction. Waiting.
