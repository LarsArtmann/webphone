# SUPERB ASR Train — Resume Execution Status (M8–M12)

**2026-10-08 13:33** · session resumed 09:18 after the "WAIT HERE" pause
(owner instruction: break down, execute, verify, repeat until done). This
report covers ONLY this session's run. Train context: plan
`docs/planning/2026-10-08_07-36_SUPERB-asr-provider-rollout.md`; prior
status `docs/status/2026-10-08_09-10_SUPERB-asr-train-execution-status.md`.

Tree is **clean** — the auto-commit daemon already committed everything
this session produced (no narrative commits made; expected per AGENTS).

---

## a) FULLY DONE (this session, all verified green)

1. **M9 save-retry queue — the RED cap spec is FIXED.** Root cause was
   the TEST stub, not the implementation: `segment += 1` ran at
   fetch-invocation (all 52 synchronous) but the stub's
   `json: async () => ({text: \`seg-${segment}\`})`read the counter
   LAZILY at`.json()`time — every segment resolved to`seg-52`, making
   the drop-oldest assertions meaningless. Fix: eager capture
   (`const n = (segment += 1)`) inside`fetchImpl`. An instrumented
   debug spec (written, run, trashed) proved the queue logic was already
   correct: cap 50, exactly 2 drops from 52 failing emits, order
   preserved, full drain.`calls-transcribe.test.mjs` 4/4 green.
2. **M8 island cooldown spec verified green** (the `headers: new Headers()`
   fix from the prior session was correct; full island suite confirmed).
3. **M8 F31 shell specs WRITTEN and green** (2 new specs in
   `shell.test.mjs`, 44/44 total):
   - auto-runs serialize: clip B's first fetch only starts after clip
     A's run fully settles (gated promise proves the chain);
   - 429 + Retry-After → exactly ONE retry (success path renders healed
     text; a second 429 fails honestly, never loops).
4. **Checkpoint gates green**: `nix fmt` (4 drift files fixed),
   `go test -count=1 ./...` — all 21 packages ok; island suite 205/205.
5. **M10 retention sweep for call_transcripts — DONE.**
   `SweepResult.Transcripts` + `DELETE FROM call_transcripts WHERE
   created_at < cutoff` in `store.Sweep`; retention log line
   (`transcripts` count); `Counts.Transcripts` + `ReadCounts` subquery;
   `webphone_transcripts_total` gauge in `/metrics`. Tests: sweep test
   extended (fresh survives via `Recent`, aged row via raw SQL goes,
   second sweep no-op) + metrics test pins the family. Green.
6. **M11 end-user transcript delete — DONE.**
   - Store: `Transcripts.DeleteCall(owner, callID)` (idempotent,
     `store.transcript_delete` family code; error-contract regenerated).
   - API: `DELETE /api/transcripts?call=` (session-gated, ASR-seam 404,
     nil-store 503, missing param 400, idempotent 204; same limiter as
     the POST).
   - View: `TranscriptGroup.CallID` threaded through `panels.go`;
     each History transcript row renders a 🗑 button
     (`data-delete-transcript` + server-localized `data-confirm`).
   - shell.js: confirm-gated delegated handler — declined confirm never
     fetches; failed delete KEEPS the row + warns; success removes the
     row and the emptied section.
   - i18n: `history.transcript.delete` + `.deleteConfirm` in BOTH maps.
   - Tests: `TestDeleteTranscript*` (round-trip, owner scoping —
     another extension's DELETE is a 204 no-op that cannot erase the
     owner's rows, idempotency, 404 seam-off, 400 no-param) + shell
     spec (confirm/keep/remove/failure paths). templ regenerated. All
     green (server 4.5s ok, views ok, shell 44/44).
7. **CHANGELOG**: [Unreleased] entries landed for M8+M9 (resilience,
   save queue) and M10+M11 (erasure, retention).
8. **Todo list repaired** to true state (25 train items + gates).

## b) PARTIALLY DONE

- **M12 per-call cap + silence guard — island half implemented,
  spec-less; store half NOT started.**
  - DONE (in tree, committed by daemon): `captureStream` now returns
    `dest`; `attachSilenceProbe(ctx, dest)` taps the mix with an
    AnalyserNode (fftSize 32768 ≈ 0.74 s tail window, RMS < 0.004 ≈
    −48 dBFS ⇒ skip + log, fail-OPEN when analyser/connect missing).
    Verified fail-open: transcribe + calls-transcribe specs still green
    (17/17) because the stubs lack `createAnalyser` — the guard is
    inert there BY DESIGN.
  - MISSING: the active-probe spec (stub `createAnalyser` with a
    controllable RMS: silent ⇒ no POST + log line, loud ⇒ POST), the
    store-side per-call cap (`TranscriptSegmentsMaxPerCall` ≈ 1500,
    second bounded trim in `Append`) + its fast bulk-seeded test, the
    M12 CHANGELOG line.

## c) NOT STARTED (from the plan's execution graph, in order)

M10→M11→M12 done/partial as above; then **M15** (History transcript
search), **M16** (copy-to-clipboard), **M17** (export zip), **M18**
(settings provider/model row), **M19** (queued-hold spec), **M20**
(concurrent-append spec + 5.4 s cap-test speedup), **M21** (`Recent()`
single-pass), **M22** (confidence marker decision), **M2** (Speaches
stack recipe), **M13** (ui-capture ASR shots), **M6** (stack browser
E2E re-run — owed: served markup changed again with M11), **M14**
(lessons + AGENTS buildflow one-liner), **M23–M25** (spike notes),
**M7** (TODO_LIST/ROADMAP harvest), final gates (fmt → buildflow →
`nix flake check` → CI verified via `gh run list` + `git ls-remote`).

## d) TOTALLY FUCKED UP

Nothing this session. No red suite left in the tree, no reverted work,
no gate failures. Two near-misses, honestly noted:

- An edit against `transcribe.js` was REJECTED mid-flight ("modified
  since last read" — daemon/formatter rewrap). Recovered by re-View;
  no damage. I let the read go stale across a long gate run — my
  mistake, the collision guard saved me.
- Wasted gate invocation with a wrong package path
  (`./internal/views/...` vs `./internal/web/views/`) and one mangled
  `rg -rn` (replace flag) output. Trivial, self-caught.

## e) WHAT WE SHOULD IMPROVE (self-critique, this session)

1. **Instrument FIRST when a timing/queue spec is red.** The prior
   session burned significant time on off-by-one theories about the
   cap spec; a 10-line debug spec solved it in one run. Rule: when a
   red spec involves promise interleaving, write the trace harness
   before the third re-read of the implementation.
2. **TDD order slipped on M12**: island silence-guard code landed
   before its spec exists. If the daemon pushes now, main carries an
   untested client path (fail-open softens the risk, not the debt).
   Write the spec before resuming anything else.
3. **`nix fmt` not yet run after the M10–M12 Go edits** — the new
   struct fields (`Transcripts int64`) likely want gofmt column
   realignment in `sweep.go`. Must precede the next gate run.
4. **`shellToast("Transcript deleted.", "ok")`** — I never verified
   that a `toast-ok` CSS class exists in app.css; if not, the success
   toast renders unstyled. One-line check owed.
5. **Daemon committed half-M12** — nothing red, but the "small
   committed units" rule argues for landing spec+code together;
   sequencing risk is real with an adversarial auto-committer.
6. `SweepResult`/`Counts` field additions ripple into any struct
   literal with positional init — none existed (checked via green
   build), but a compile elsewhere (stack consumers) is impossible by
   design (internal/), so this is fine — noted only to close the loop.

## f) NEXT — up to 50 items, in execution order

> **CLOSED 2026-10-08 ~15:50** — the 14:08–15:50 continuation executed
> this list to completion: M12 (items 1–7) plus M15–M25 landed and
> verified (see the 14-08 report + the TODO_LIST "ASR/SUPERB train
> follow-ups" harvest row for what remains — all of it owner-gated or
> grouped polish). CI green at `7c87501`+; buildflow EXIT 0.

1. Write the M12 silence-guard island spec (analyser stub: silent ⇒ no
   POST + log; loud ⇒ POST; fail-open already proven).
2. Store: `TranscriptSegmentsMaxPerCall = 1500` + per-call bounded
   trim in `Append` (second statement, same MIN-id shape).
3. Store test: bulk-seed 1502 rows via one tx + single `Append` →
   count 1500, oldest two gone, other calls untouched.
4. M12 CHANGELOG entry.
5. `nix fmt` (Go realignment + island/shell formatting).
6. Verify `toast-ok`/`toast-warn` classes exist in app.css; add if
   missing.
7. Go test + island suite checkpoint after M12 completion.
8. M15: extend the History CDR query path — SQL `LIKE` over
   `call_transcripts` (owner-scoped) merged with CDR hits.
9. M15: view hint when a transcript matched (not just a CDR row).
10. M15: tests — match, no-match, owner scope, seam-off.
11. M16: copy-to-clipboard on History transcript sections
    (clipboard API + toast; CSP note: no inline handlers).
12. M16: copy affordance on the live call card transcript.
13. M16: copy on voicemail/attachment transcript targets (shell.js).
14. M16: island + shell copy specs (clipboard stub).
15. M17: export zip gains a transcripts leg (JSON per call) + test.
16. M18: settings panel row — provider kind + model name (props from
    config; `asrOn()` seam already provider-agnostic).
17. M18: settings view test.
18. M19: island spec — pause/resume on the QUEUED hold path
    (`holdQueued` mid-flight).
19. M20: store concurrent-`Append` spec (goroutines; per-owner AND
    per-call caps hold under race).
20. M20: speed up the 5.4 s per-owner cap test (stride or smaller
    fixture) — suite back under 2 s.
21. M21: `Recent()` single-pass assembly (reverse pass, no O(n²)
    prepend).
22. M22: probe what the live provider(s) return for confidence;
    decide + document (Google v2 returns confidence 0.0 = unset —
    likely REJECT for marker plumbing; write the note).
23. M2: stack module option schema `asr.{enable,image,model,apiKey,ram}`
    (webphone-module check stand-ins).
24. M2: oci-container service (cpu image, int8, PRELOAD_MODELS,
    loopback-only, hardened overrides) on the consuming stack repo —
    NEEDS owner permission (open question, prior report §g).
25. M2: stack README section + image digest pin.
26. M13: ui-capture ASR-on shots (settings row, call card with
    transcript, voicemail button) + DOM asserts + baselines.
27. M6: stack browser E2E re-run (budget 445 s; known transfer flake —
    re-run once) — OWED, markup changed twice since last run.
28. M6: dom-contract test re-run (`TestServedPageHoldsTheDomContract`).
29. M14: docs/lessons entry — `.buildflow.yml` truncation incident +
    "run buildflow at train START" rule.
30. M14: AGENTS.md one-liner for the buildflow-at-start rule (watch
    the 377-line cap).
31. M23: spike note — streaming SSE transcription via Speaches
    (feasibility only).
32. M24: spike note — native Speaches NixOS module packaging effort.
33. M25: spike note — diarization options (pyannote via Speaches?).
34. M7: docs-health HARVEST — fold status 16-54/18-15 §f + this plan
    into TODO_LIST.md / ROADMAP.md; prune done items.
35. Final gates: `nix fmt` → buildflow (inside nix develop) →
    `nix flake check`.
36. Verify CI green (`gh run list`) + `git ls-remote` end state after
    the daemon's pushes.
37. Re-run the full smoke (`webphone-smoke.py`) once M12+M15 land —
    it proves wiring end-to-end after each server-touching train.
38. Consider: smoke leg asserting the delete endpoint (round-trip via
    HTTP) — cheap addition to `asr_scenario`.
39. Consider: teach the island cooldown to honor the seam's
    Retry-After (currently fixed 10 s) — only if the owner wants it
    (question g-2).
40. Check `docs/error-contract.md` "Boot surface" needs no update
    (M10–M12 added no boot paths — confirm, one grep).
41. Update `FEATURES.md` transcript feature rows when the train closes
    (erasure + retention + caps are user-visible statuses).
42. Stack runbook § "Webphone error contract" cross-sync if the
    delete endpoint changed the table (it did — 204/400/404/503 rows).
43. Re-verify `scripts/whitespace-drift.sh` passes before any manual
    commit (no manual commits planned; daemon owns them).
44. Sweep-log line verification in a live boot (the smoke boots with
    retention off — a one-off `retention_days=1` boot proves the log
    line; optional but cheap).
45. Confirm the 429 shell retry's 10 s bound cannot stack with the
    island cooldown into a 20 s dead window on one flood (reasoned
    safe — shell handles vm/MMS, island handles live capture; write
    the reasoning into the M8 lessons line).
46. Look at `TranscriptSegmentsMaxPerExtension` (5000) vs new per-call
    1500: document the interaction (3+ long calls start eating each
    other's owner budget — intended).
47. Post-train: re-measure erraudit tiers 1+2 (monthly due 2026-11-05;
    early check cheap after new error paths).
48. Post-train: `buildflow -s gitleaks` + `-s codespell` on-demand
    (they gate releases only).
49. Consider a `wp-transcript-call` stable-id requirement for future
    morph-swap of the History section (currently outerHTML re-render;
    harmless, note for D-contract hygiene).
50. Close the train: mark todo items completed, final CI check,
    status report.

## g) Questions for the owner (cannot self-answer)

1. **Live-reality verification credentials**: M4 proved both provider
   wires with in-process fakes; a REALITY run needs a GCP project +
   API key (or a Speaches host). Do you want to hand me a throwaway
   Google key for one verified live transcription before this train
   closes, or is fake-wire green + the stack E2E sufficient?
2. **Island cooldown tuning**: the live loop pauses a fixed 10 s on a
   429 while the shell honors the server's Retry-After (bounded 10 s).
   Should the island cooldown also read Retry-After (one more moving
   part) or stay fixed (simpler, always ≤ budget recovery)?
3. **Silence-guard floor**: 0.004 RMS (≈ −48 dBFS) over a 0.74 s tail
   window is a reasoned guess — I cannot synthesize honest far-end
   hold audio locally. Ship as-is and tune from production skip logs,
   or do you want a different default before M12 closes?

**Execution is PAUSED here awaiting instructions** (per the standing
"WAIT" posture: report, then wait).
