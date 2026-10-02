# Status Report — 2026-09-30 12:59 CEST — `-t 2` Dedup Sweep (Zero Harmful Duplication)

**Session scope:** one task — owner-requested `-t 2` art-dupl dedup sweep, judged against
`docs/dedup-registry.md` (the ONE acceptance home). No other work was started or touched.
**Repo at session start:** `97d9c79`, clean tree. **Repo at report time:** `b1d494f`
(daemon commit; my registry edit + a concurrent session's work landed together).
**Flags override note:** the status-report skill's canonical output is a styled HTML
dashboard; the owner explicitly requested `.md` — honored, one-off, not propagated
into the skill.

**One-line verdict:** the codebase is CLEAN at `-t 2` — 15/15 actionable clone groups
attributed to standing rulings, 0 new, 0 harmful, 0 extractions needed.

> ARCHIVED 2026-10-03 (docs-health v6 sweep): the clean-at-`-t 2` verdict
> stands; the owner deltas (baseline ratification, settingsRow, registry
> ratification, suppression scope) live in the OWNER-calls TODO row +
> the registry's open-calls list, the standing-work table resolved with
> the v2.8.0 release and the 2026-10-01/02 trains, and the protocol
> text carried the sweep-log link rule. Per-item verdicts inline.


---

## a) FULLY DONE

~~1. **Protocol step 1 — attribution against MY OWN run, never the paste.**~~ done — this session (report of record)
   Re-ran `art-dupl --sort total-tokens -t 2 --suggest-generics --timing --rich-text`
   at HEAD `97d9c79` (dropped `--type-aware` per the tool's own precedence note).
   Result: **identical to the owner's pasted report** — same 15 shown groups, same
   windows (141 detected; 55 non-actionable, 71 filtered suppressed).
~~2. **Protocol step 2 — every occurrence site re-read at HEAD, not one site per group.**~~ done — this session (report of record)
   All 15 groups × all their windows read fresh: settings.templ (dt/dd rows,
   exportLink), layout.templ (signed-in span, both nav badges, panelHead-internal,
   welcome block ×2), messages.templ (thread row, segcount ×2, both send buttons,
   attachment lines), contacts.templ (export link, shared + personal avatar rows),
   history.templ (filter button, needsAPI, CDRRow), voicemail.templ (needsAPI,
   VoicemailRow), fax.templ (panelHead call, send button), phone.templ (data-i18n
   spans), idempotency.go (clock+lock prologue + its in-code rationale), vcard.go
   (escape branches), i18n_test.go (%% skip), actions.go + contacts_api.go
   (contactSaveFailed sites incl. the ErrListFull branch).
~~3. **Sibling-site greps — no unreported sites anywhere.** `wp-nav-badge` exactly ×3~~ done — this session (report of record)
   (layout ×2 + messages thread row), `wp-segcount` exactly ×2 (both messages
   composers), `data-i18n` confined to phone.templ (island DOM contract),
   `contactSaveFailed(w)` exactly ×2 (+ its contract test). Nothing the window
   clustering could have hidden.
~~4. **15/15 groups judged and attributed.** Highlights verified with evidence:~~ done — this session (report of record)
   the settings dt/dd rows still show 7 divergent value shapes
   (orUnset / plain / fmtInt / conditional / link); both nav badges and the thread
   badge carry **different conditions + count sources** (props.Unread /
   props.NewVoicemail / summary.Thread.Unread); the two needsAPI blocks use
   **different key families** (`history.needsAPI.*` vs `vm.disabled.*`) and markup
   shapes; shared vs personal avatar rows differ in domain rules (typed
   `contact.Phone` vs raw `contact.Number`, delete affordance, shared tag); the fax
   send button uniquely carries the htmx-indicator span. Zero groups warranted an
   extraction; zero standing rulings were overturned.
~~5. **Registry updated (one home per fact, widening not duplicating):**~~ done — this session (report of record)
   the "Coincidental closing-tag fragments" row now covers the contacts-export
   `</a>` vs history-needsAPI `</p>` pair (docs/dedup-registry.md:39); the
   "wp-segcount + submit button" row now covers the cross-tab fax-vs-messages send
   buttons (docs/dedup-registry.md:44); sweep-log line appended
   (docs/dedup-registry.md:68).
~~6. **Registry edit committed** by the auto-commit daemon in `b1d494f`~~ done — this session (report of record; b1d494f)
   (verified via `git show --stat`, 12:55:20).
~~7. **Concurrent-session discipline held.** A second session's~~ done — this session (report of record)
   `package/nixos-module.nix` rework (−97/+143 lines) and the untracked
   `docs/planning/2026-09-30_12-53_fax-paperless-integration-plan.md` were
   detected, reported, and **not touched**. Both rode along in the same daemon
   commit.
~~8. **Todo list maintained** throughout (5/5 steps completed).~~ done — this session (report of record)

## b) PARTIALLY DONE

~~1. **Protocol step 4 (sweep-log line "with a link to its status doc") — was~~ done — fixed in-session (the 12:59 line links this doc)
   incomplete.** My 12:55 sweep-log line had no status-doc link; prior entries cite
   their status docs (e.g. "§a.5"). Caught by this self-review; the link to THIS
   doc is being added right after it is written. _Update: fixed — see the 12:59
   line in docs/dedup-registry.md._
~~2. **End-state verification — local commit verified, remote NOT in sync.**~~ resolved by events — remote synced; ls-remote verified at the v2.8.0 ritual (2026-10-01)
   `git ls-remote origin main` returns `b0458d2` while local HEAD is `b1d494f`:
   **the remote is currently AHEAD of local.** Per AGENTS.md the daemon commits AND
   pushes continuously; either its push is in flight or the concurrent session
   pushed newer commits local hasn't pulled. I deliberately did not pull/rebase —
   that is a concurrent-session coordination call, not mine to make mid-flight.
~~3. **The `-t 2` sweep's coverage is the 15 SHOWN groups.** The 55 non-actionable~~ owner — routed to the OWNER-calls row by this sweep (suppression-scope ruling)
   and 71 suppressed groups were taken on the tool's classification plus sweep
   precedent (all five prior sweeps did the same) — they were NOT individually
   audited. Defensible, but it is convention, not a recorded ruling (see f/4).

## c) NOT STARTED (deliberately)

~~1. **No code refactors** — none were needed; zero harmful duplication means the~~ record stands — zero harmful duplication is the success state
   extraction machinery (helper + micro-test + old-vs-new binary render diff) was
   never exercised. This "not started" is the _success_ state, not a gap.
~~2. **No tests/gates re-run** — only a Markdown file changed (out of treefmt scope;~~ record stands — docs-only change, gates consciously skipped
   no code paths touched). A full `buildflow` run would have spent ~minutes
   proving a docs-only change; skipped consciously.
~~3. **Owner ratifications** — all three registry open owner calls remain open and~~ owner — all three live in the OWNER-calls TODO row
   are owner-only (see g).
~~4. **Suppressed-set audit** — not started; needs an owner ruling first (f/4, g/3).~~ owner — routed to the OWNER-calls row + the registry open-calls list by this sweep
~~5. **Anything about the concurrent session's nixos-module.nix change** — their~~ resolved by events — the nixos-module rework rode the 09-30 nix/family trains; its session owned it
   in-flight work; not mine to review mid-edit.

## d) TOTALLY FUCKED UP

**Nothing is fucked up.** No code was harmed, no tests broken, no false "done"
claims, no ghost systems created. The honest defect list is small and already
handled:

~~1. **One wasted edit round-trip.** My first multiedit batch failed on the~~ process record
   closing-tag registry row because I _re-typed_ the old_string from memory
   instead of copying it verbatim from the view output (whitespace mismatch).
   Retried with exact text — fixed. Violation of my own exact-match discipline,
   cost: one tool call.
~~2. **The sweep-log line initially violated the protocol it lives in** (missing~~ process record
   status-doc link, protocol step 4). Self-caught during this review, fixed same
   turn. Ironic: the session that enforced the registry didn't fully obey it on
   the first pass.
~~3. **Severity-1 count: zero.** Stating that plainly so the empty section is a~~ record stands
   verified claim, not an omission.

## e) WHAT WE SHOULD IMPROVE

~~1. **Paste-vs-rerun comparison was eyeballed.** The two outputs happened to match~~ not adopted — below the bar
   and I verified by reading; a scripted group-window diff (10 lines of shell) would
   make the protocol's "attribute against YOUR run" step mechanically provable.
   Candidate: a tiny `scripts/dedup-diff.sh` or a nix app.
~~2. **The suppression boundary is convention, not a recorded ruling.** Every sweep~~ owner — routed to the OWNER-calls row (suppression-scope ruling)
   to date (six now) judged only the SHOWN set. That is probably right — the tool's
   non-actionable/suppressed classes exist for a reason — but the registry protocol
   should SAY so, or mandate a periodic suppression audit. Today a future session
   must re-derive this from precedent.
~~3. **Canonical art-dupl flag set is unsettled.** This sweep ran~~ done in part — the protocol records the -t 3/-t 2 tiers (dedup-registry.md:11); the canonical FLAG set is still unset
   `--suggest-generics` (types erased — the strictly more sensitive mode); the
   owner's paste added `--type-aware`, which the tool itself warns to drop. The
   registry's protocol line says "run art-dupl YOURSELF (working baseline `-t 3`)"
   and nothing about flags. Record the canonical invocation.
~~4. **"Sweep-log line links its status doc" should be in the protocol text,** not~~ done — protocol step 4 carries the link rule (dedup-registry.md:21)
   just in the one prior entry's format. Cheap edit to docs/dedup-registry.md step 4.
~~5. **My edit hygiene slipped once** (see d/1). Rule re-confirmed: old_string is~~ process record
   always copied from view output, never reconstructed.
~~6. **Older sweep-log references are uncheckable shorthand.** "§a.5" and similar~~ not adopted — below the bar
   point into status docs whose names aren't given; whether they still resolve was
   NOT verified this session (out of scope per owner instruction). A light ANNOTATE
   pass could backfill explicit paths.
~~7. **art-dupl's availability is undocumented in AGENTS.md.** It resolved from the~~ not adopted — AGENTS sits at the 377-line cap; below the bar
   host PATH (`/run/current-system/sw/bin/art-dupl`), NOT just inside `nix develop`
   — a pleasant surprise nobody recorded. One line in the Commands block would save
   the next session the `which` dance.
~~8. **The registry table grows by widening rows** (this session) — fine at 14 rows,~~ record stands — watch item, no speculative refactor
   but if sweeps keep finding "family" attributions, consider one row per recurring
   clone _family_ with the sites list instead of per-pair descriptions. Keep an eye
   on it; do not refactor the table speculatively.

## f) Next tasks (up to 50 — 30 real ones, rest deliberately not padded)

**Dedup/process domain (this session's home ground):**

| #  | Task                                                                                                   | Who         |
| -- | ------------------------------------------------------------------------------------------------------ | ----------- |
~~| 1  | Ratify dedup baseline: `-t 3` ritual vs `-t 2` deep sweeps vs always-both                              | OWNER (g/1) |~~ owner — OWNER-calls row (art-dupl -t 3 baseline, FIFTH surfacing)
~~| 2  | `settingsRow`: build the component or accept the dt/dd rows permanently (surfaced in ALL six sweeps)   | OWNER (g/2) |~~ owner — OWNER-calls row (settingsRow build-or-retire, FOURTH surfacing)
~~| 3  | Ratify the registry as the ONE acceptance home (open since 09-23 03:01, re-raised twice)               | OWNER       |~~ owner — OWNER-calls row (registry ratification)
~~| 4  | Record a suppression ruling: is shown-only the ratified scope, or add a periodic suppressed-set audit? | OWNER (g/3) |~~ owner — OWNER-calls row + registry open-calls list (added by this sweep)
~~| 5  | Bake "every sweep-log line links its status doc" into the registry protocol step 4 text                | session     |~~ done — protocol step 4 carries the rule (dedup-registry.md:21)
~~| 6  | Record the canonical art-dupl invocation + flags in the registry protocol                              | session     |~~ done in part — tiers recorded (registry:11); the flag set still unset
~~| 7  | Script the paste-vs-rerun group diff (makes protocol step 1 mechanically provable)                     | session     |~~ not adopted — below the bar
~~| 8  | Decide if `-t 1` ultra-deep passes ever recur (precedent: 09-18, 09-23 01:12) or retire that mode      | OWNER       |~~ routed — ROADMAP art-dupl bullet records `-t 1` as forensic-only
~~| 9  | Light ANNOTATE pass: backfill explicit doc paths for old sweep-log refs ("§a.5" etc.)                  | session     |~~ not adopted — below the bar
~~| 10 | Add art-dupl (host-PATH availability + invocation) to AGENTS.md Commands block                         | session     |~~ not adopted — AGENTS line-capped; below the bar

**Session follow-ups (observed this run, small and concrete):**

| #  | Task                                                                                                                                              | Who                |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------ |
~~| 11 | Re-verify `git ls-remote` end state after the daemon's push settles (remote `b0458d2` vs local `b1d494f`)                                         | session            |~~ resolved by events — synced + verified at the v2.8.0 ritual (2026-10-01)
~~| 12 | Review the concurrent session's `nixos-module.nix` rework (−97/+143) AFTER its session declares it done                                           | next session       |~~ resolved by events — rode the 09-30 trains; gates green at the close-outs
~~| 13 | Confirm `module-check.nix` stand-ins still cover any NEW config keys that rework writes (flake-check gate)                                        | next session       |~~ done — the stand-ins gained freeformType (nix-review train, CHANGELOG Unreleased)
~~| 14 | Confirm the fax-paperless plan doc is claimed by its session (it rode the daemon commit unreviewed by me)                                         | next session       |~~ resolved by events — shipped as the T14 train (FEATURES + CHANGELOG)
~~| 15 | Standing trigger reminder: ANY future templ markup change → re-run the stack's browser E2E (AGENTS rule; this sweep read but didn't touch markup) | next markup change |~~ record stands — the E2E obligation lives in the island-honesty + cross-repo rows

**Known standing work from AGENTS.md context (NOT re-verified this session —
recorded here as-is, per the no-other-research instruction):**

| #  | Task                                                                                                                                                | Source              |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------- |
~~| 16 | v2.7.0 train: stack re-lock + re-pin (stack forward-locked at `94ae28d`, lowercase contacts assert flipped, awaiting the ride)                      | AGENTS.md           |~~ done — v2.8.0 tagged + stack lock bump `3afcf57` (TODO v2.8.0 deploy-tail row)
~~| 17 | erraudit tiers 1+2 monthly re-measure — due **2026-10-22**, tier-2 must stay 0                                                                      | AGENTS.md           |~~ done — re-measured early 2026-10-01 (tier-1 0, tier-2 0); next due 2026-10-22 (standing watches row)
~~| 18 | Setup-shell adoption: poll cqrs-htmx upstream for the two seams (`Config.DisableAuth`, service-optional `New()`)                                    | AGENTS.md           |~~ resolved by events — the adoption was REJECTED on footprint (+68%); the salvage rode httputil.NewServer, the seam poll is moot
~~| 19 | Pre-measure + record the pre-adoption stripped-binary baseline so the ≤ +8 MB / ≤ +20 % footprint gate is ready                                     | AGENTS.md           |~~ done — measured + recorded in the setup-shell NO-GO verdict (docs/planning/2026-09-30_10-37)
~~| 20 | After any dep bump the daemon sweeps: same-breath vendorHash roundtrip (recurred at `0a7a732`)                                                      | AGENTS.md           |~~ record stands — standing rule (the vendorHash roundtrip recurred + was handled)
~~| 21 | pbx-artmann relock ritual when the v2.7.0 train closes (docs/release-runbook.md dance)                                                              | AGENTS.md           |~~ routed — TODO v2.8.0 deploy tail (owner terminal: relock + re-pin)
~~| 22 | aarch64 cross-build verify by ELF bytes on next release train                                                                                       | AGENTS.md           |~~ done — ELF bytes verified at the v2.8.0 (b700) and 12:57 close-out (183) cross-builds
~~| 23 | `nix run .#vulnix` gate before next release                                                                                                         | AGENTS.md           |~~ done — green on the v2.8.0 release train (release.sh gates on it)
~~| 24 | Re-measure stack browser E2E budget if scenarios grow again (445 s baseline, FOUC +~15 s)                                                           | AGENTS.md           |~~ record stands — standing watch (budget 445s; greens 195s/184s)
~~| 25 | Harvest THIS report's section (f) into TODO_LIST/ROADMAP via docs-health HARVEST after owner review                                                 | status-report skill |~~ done — this v6 sweep harvested it (2026-10-03)
~~| 26 | Consider making the dedup sweep a recurring ritual (e.g. monthly, paired with the erraudit measure)                                                 | proposal            |~~ not adopted — proposal, no demand signal
~~| 27 | Decide whether the registry ruling table should track which sweeps saw each group (light provenance) — do NOT build speculatively                   | proposal            |~~ record stands — do NOT build speculatively (own text)
~~| 28 | i18n_test.go sits in dedup reports as clone-noise; check whether art-dupl's test-file suppression SHOULD cover Go test files (it currently doesn't) | proposal            |~~ not adopted — upstream tool policy (art-dupl); no ruling sought
~~| 29 | Sweep-log growth: at 7 entries consider archiving pre-registry-era entries (09-18) to a provenance note                                             | proposal            |~~ not adopted — below the bar
~~| 30 | When the setup-shell adoption lands, re-run this sweep — new middleware surface = new clone candidates                                              | proposal            |~~ resolved by events — the adoption was rejected; httputil.NewServer added no clone surface (sweeps since stayed clean)

Items 31–50: intentionally unused. The remaining ideas I could generate would be
padding (e.g. "document each ACCEPT row more") — the honest backlog is ~30 items.

## g) Questions I can NOT figure out myself

~~1. **Baseline ratification:** is `-t 3` the standing ritual (with `-t 2` deep sweeps~~ owner — OWNER-calls row (art-dupl -t 3 baseline, FIFTH surfacing)
   on your request), is `-t 2` the new baseline, or should future sessions run both?
   Six sweeps in, this is still the registry's oldest unresolved owner call.
~~2. **`settingsRow`:** build the shared dt/dd component or accept the settings rows~~ owner — OWNER-calls row (settingsRow build-or-retire, FOURTH surfacing)
   permanently? It has surfaced in EVERY sweep on record; my ACCEPT stands on the
   7-divergent-shapes evidence, but the standing ruling explicitly reserves this
   for you.
~~3. **Suppression scope:** should the dedup protocol add a periodic audit of the~~ owner — OWNER-calls row + registry open-calls list (added by this sweep)
   71 suppressed + 55 non-actionable groups (e.g. a quarterly `--include-generated`
   spot check), or is shown-only the permanently ratified scope? This changes the
   ritual's cost every run, so I won't pick unilaterally.

---

_Point-in-time snapshot — goes stale by design. Per the status-report skill, section
(f) is HARVEST fuel for TODO_LIST/ROADMAP after your review. WAITING FOR
INSTRUCTIONS._
