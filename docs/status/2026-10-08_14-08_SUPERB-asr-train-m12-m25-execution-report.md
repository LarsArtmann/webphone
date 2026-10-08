# SUPERB ASR Train — M12–M25 Execution Report (resume session)

**2026-10-08 14:08** · session 13:37→14:08 (after the 13:33 pause; owner
said resume the whole plan). Scope executed: M12 completion through M14
(lessons half) + M19–M25. Prior-session context: status 13-33.

---

## a) FULLY DONE this session

1. **M12 closed** (was ~60% at pause): store per-call cap
   (`TranscriptSegmentsMaxPerCall` = 1500, same MIN-id bounded trim shape
   in `Append`, errorfamily code reused `store.transcript_trim`), fast
   bulk-seeded cap test (one tx of 1502 raw INSERTs + 1 Append → exactly
   1500 survive, oldest gone, other call untouched), island
   ACTIVE silence-probe spec (stub gained `nextRms`/`createAnalyser`/
   `connect()`; 0.001 RMS → no POST, 0.01 → POST; spec ordered BEFORE
   the 429 test so the module cooldown cannot swallow its segments),
   CHANGELOG entry. The owner-cap test was updated for the intended
   interaction (flood now spread over 4 calls — a single-call flood
   stops at the per-call cap first). `toast-ok` check CLOSED: the class
   exists in island/style.css:844 (shared CSS — shell toasts reuse it).
2. **Checkpoint gates green** (post-M12): `nix fmt` + `go test ./...`
   (21 pkgs) + island 209/209.
3. **M15 History transcript search**: store `Search` (owner-scoped LIKE,
   `likeEscape` reused from messages.go, matching lines only),
   `reassembleTranscripts` extracted as the ONE grouping home; server
   `recentTranscripts(r, sess, query)` branches Recent/Search; view
   hint "Transcript matches for: `<q>`" (i18n en/de both maps); spec
   covers match / no-match hides the section / owner scope both ways /
   `100%` literal (pre-encoded `%25`) / hint presence+absence. Bug found
   by the spec: the PhoneAPI-disabled branch dropped `Query` — fixed.
4. **M16 copy-to-clipboard, all three surfaces**: History rows
   (`data-copy-transcript` button, label localized), shell.js ONE
   delegated handler (clipboard API + honest toasts:
   "Nothing to copy yet." / "Copied to clipboard." / "Could not copy —
   select the text manually."), dynamic "Copy" chip planted next to
   real voicemail/attachment transcriptions (empty results plant
   nothing; planting wrapped in try/catch — a chip must never fail the
   render it accompanies), island call card "Copy transcript" button
   (`copyCardTranscript`, island i18n en/de). helpers.mjs stub gained
   `insertAdjacentElement("afterend")`. Specs: shell copy row spec,
   chip-plant+copy spec, island card copy spec, server served-markup
   assertion. Island suite 212/212 after.
5. **M17 export zip transcripts leg**: store `Export` (whole owner
   history, LIMIT = owner cap — NOT the 2000 render window), all three
   store queries consolidated behind private `fetchSegments` (one home
   for the SELECT), `transcripts.json` DTO + zip entry, read failure
   fails the whole export (a silently incomplete zip is not an honest
   export), nil store skips the leg (test compositions). testServer
   gained a `transcripts` field; export test asserts spoken order,
   identity, and that a neighbor extension's words never ride along.
6. **M18 settings provider display**: `Provider.Describe() string`
   added to the seam interface (both real implementations; no test
   fakes existed to break); Settings ASR row now renders
   "connected · `<kind>` · `<model>`" (e.g.
   `openai-compatible · whisper-1`, `google · telephony`). Tests in
   asr (default + named models) and server (seam on/off).
7. **M19 queued-hold spec** (test-only): the mid-flight queue is pinned
   by injecting `holdQueued` in the in-flight window (the hold button
   is disabled mid-flight — the queue is the defensive seam for
   programmatic toggles): queued opposite wish REPLAYS as a second
   re-INVITE; a matching wish never replays.
8. **M20 suite health** (test-only): cap test rewritten to bulk-seed
   the flood in ONE tx + a single Append (3.29 s → 0.04 s); NEW
   concurrency spec — 8 goroutines × 25 Appends across two calls: zero
   errors, zero lost rows, Recent reassembles everything (the
   single-connection pool serializes; the contract is no locks, no
   loss).
9. **M21 Recent single-pass**: lines append in fetch order + one
   reverse per group (`reverseLines`) instead of prepend-per-segment
   (was O(n²) on a marathon call). Behavior parity proven by the
   existing order-pinning tests; store suite ~6.7 s → ~0.11 s total.
   CHANGELOG "Changed" entry.
10. **M22 confidence decision — REJECT, verified live**: Google v2
    `Alternative.confidence` is top-alternative-only, "not guaranteed to
    be accurate… not rely on it", 0.0 = unset sentinel (docs quoted in
    the note); Speaches `verbose_json` carries avg_logprob/no_speech_prob
    but NO calibrated confidence. A marker built on either would violate
    the island honesty contract. Seam stays text-only; revisit triggers
    documented.
11. **M23–M25 spike notes — all verified against live sources**
    (`docs/planning/2026-10-08_14-05_SUPERB-asr-decisions-and-spikes.md`):
    M23: Speaches HAS a `/v1/realtime` WebSocket with
    `?intent=transcription` (true live-mic STT, no LLM) — feasible but
    provider-specific (would break the provider-agnostic seam unless
    proxyped server-side); keep 4 s chunking. M24: `speaches` NOT in
    nixpkgs but faster-whisper + ctranslate2 are, and
    `wyoming-faster-whisper` ships the application+module template —
    1–3 day estimate; oci-container stays the story. M25: Speaches
    diarization endpoint EXISTS (pyannote; even known-speaker voice
    references) but has open stability + a security issue — wait.
12. **M14 (half)**: `.buildflow.yml` truncation lesson appended to
    docs/lessons.md (Tooling traps) — the 00:46 daemon commit that cut
    69 lines of `skip_steps` rationale, ~16 h undetected, plus the two
    rules: buildflow at TRAIN START, and re-read policy files after
    daemon activity near them.

## b) PARTIALLY DONE

1. **M14**: lessons entry done; the AGENTS.md one-liner ("run buildflow
   at train start") is NOT yet added (file is 136/377 lines — room
   exists; next step).
2. **Final gates not yet run since the M17–M21 edits**: `nix fmt` ran
   after M12 and M16 but NOT after the later Go edits (export.go,
   transcripts.go, panels.go, google_test.go, server_test.go…) — a
   treefmt alignment red of the e62fe34 class is LIKELY without it (see
   d-2). Full `go test ./...` + island suite after M17–M21, buildflow,
   `nix flake check`: all pending.

## c) NOT STARTED (this session)

1. **M13 ui-capture ASR shots** (chromium harness, local-only budget).
2. **M7 TODO_LIST/ROADMAP harvest** (SUPERB plan + 16-54/18-15 §f + the
   M22–M25 revisit triggers → ROADMAP).
3. **M2 Speaches stack recipe** — owner-gated (stack-repo permission,
   question open since 09:10).
4. **M6 stack browser E2E re-run** — OWED and now more owed: served
   markup changed AGAIN this session (search hint, copy buttons,
   settings detail row). Runs in the stack repo.
5. Push verification / CI greening (see d-2 — this is now urgent).

## d) WHAT I FUCKED UP / WHAT ALMOST WENT WRONG

1. **The edit-tool "insert-before" clobber, THREE times**: inserting a
   new test before an existing one, my `new_string` omitted the original
   function header — transcripts_test.go (TestTranscriptsAppendTrimsPastTheCap
   body orphaned), calls-transcribe.test.mjs (bounded-queue spec header
   eaten), asr_test.go (dropped `var gotModel, gotFilename string`).
   All three caught by immediate re-view/re-run and repaired, but the
   pattern repeated. Rule I should internalize: when inserting before
   existing code, the old_string's FIRST LINE must reappear in
   new_string.
2. **Main has been RED since 09:54 and nobody looked** (the exact
   2026-10-05 failure mode AGENTS warns about): CI run 37746390621
   failed on (a) the M9 cap spec — the KNOWN morning-red that the M9
   test-stub fix (lazy json() → eager capture) already repairs, and
   (b) a treefmt struct-alignment red in google_test.go
   (`family   errorfamily.Family` → wider). Both fixes exist in the 16
   UNPUSHED local commits — the daemon's push-leg has been dead since
   ~09:54 (nothing pushed all afternoon; 16 commits pending, oldest
   13:15:12, i.e. 53 min at report time — 7 min from the 60-minute
   manual-push bar anchor). I only noticed when assembling THIS report;
   the morning session's "checkpoint green" claim was local-only.
3. `rg -rn` used TWICE despite the AGENTS gotcha (the `-r` is
   display-replace; mangled output, one wasted re-run; no file damage —
   ripgrep never writes).
4. M16 chip planting initially BROKE the 429 spec: the stub had no
   `insertAdjacentElement`, the throw fell into the render catch and
   flipped a healed transcription to "Transcription failed". Fixed by
   hardening the plant with its own try/catch (the right production
   behavior too) — but the first version shipped a convenience that
   could clobber its own payload.
5. A compound CSS selector (`[data-copy-transcript], [data-copy-target]`)
   in the delegated handler was untestable against the helpers stub's
   closest (exact-match only) — split into two closest calls. Minor, but
   I wrote it knowing the stub's limits.

## e) WHAT WE SHOULD IMPROVE

1. **CI-check after EVERY push must include the daemon's pushes** — the
   daemon pushed red at 09:54 and 4+ hours passed without anyone
   noticing. A hook/heartbeat that greps `gh run list` per session START
   (not just after own pushes) would have caught this at 13:37.
2. **Gate-first violated for this resume session too**: the new lessons
   rule ("buildflow at train START") was written BY this train but not
   PRACTICED by it — the session opened mid-train executing, never
   re-baselined. The next session must start with: `nix fmt` →
   `go test ./...` → island → buildflow → push (bar analysis below).
3. The edit-tool clobber pattern (d-1) deserves a lessons.md entry —
   three occurrences in one session across two languages is a pattern,
   not bad luck.
4. Store test seeding idiom (raw INSERT tx for volume, Append only for
   the path under test) is now used in three tests — could be a tiny
   seed helper next time someone touches the fourth.

## f) NEXT (ordered, ~30 items)

1. `nix fmt` (Go alignment drift is nearly certain — d-2b).
2. Full gates: `go test -count=1 ./...` + island all-files suite.
3. **Push decision**: 16 unpushed commits, oldest 13:15 — the
   60-minute bar from the FIRST UNPUSHED COMMIT triggers at ~14:15.
   With main red and the verified fix in hand (both red causes fixed
   locally), the bar says push manually after step 2 passes, then
   verify CI green (`gh run list` + `git ls-remote`).
4. `buildflow` full run (train-end gate; also proves the restored
   `.buildflow.yml` policy).
5. `nix flake check` (sandbox build + treefmt + island-lint + module
   check; the KVM backup test).
6. AGENTS.md one-liner: "run buildflow at train START" (M14 remainder;
   136/377 lines — room).
7. M13 ui-capture ASR shots (chromium+selenium harness; 14 shots +
   ASR-on set: settings row w/ provider detail, call card w/ transcript
   + copy, History transcripts w/ delete+copy+search hint; DOM asserts
   per shot; regen baselines; LOCAL-ONLY).
8. M7 harvest: TODO_LIST.md (open items from 16-54/18-15 §f + SUPERB
   plan remainders) + ROADMAP.md (M22 revisit triggers, M23 realtime
   WS, M24 native packaging, M25 diarization watch) + prune done items.
9. M6 stack browser E2E re-run (owner-gated; budget 445 s; re-run once
   on the transfer flake; markup changed twice today).
10. M2 Speaches stack recipe (owner-gated, F3–F9).
11. lessons.md: the edit-tool insert-before clobber entry (e-3).
12. dedup-registry sweep-log line for this session (fetchSegments
    consolidation pre-empted a clone; likeEscape reused).
13. README contract bullets if owed: transcripts.json export leg,
    History transcript search, provider display row (README documents
    contracts; check the export section wording).
14. FEATURES.md rows: transcript search / copy affordances / export leg
    / per-call cap + silence guard / provider display.
15. Verify daemon push-leg health after the manual push (does it resume
    pushing? if dead, every future session pushes manually).
16. Annotate the 13-33 report's next-list with what this session closed.
17. Re-check `TestErrorCodeRegistryIsFresh` (no new codes added — reused
    `store.transcript_trim` — but run to confirm the registry matches).
18. If owner answers Retry-After question: island cooldown honors it
    (small transcribe.js change + spec).
19. If owner answers silence-floor question: tune `SILENCE_RMS` from
    any production evidence.
20. If owner provides a GCP key / Speaches instance: the one
    live-reality transcription (M4's remaining doubt).
21. Status-report hygiene: this file's d-2 push state re-verified at
    resume.
22–30. (buffer: triage whatever CI says after the push; ui-capture DOM
    asserts may surface layout drift from the new buttons; consider
    `historyTranscriptGroups` comment refresh already done; smoke
    re-run optional; vulnix only at release time.)

## g) QUESTIONS FOR THE OWNER (cannot self-answer)

1. **M2 stack-repo permission** (open since 09:10): may I commit the
   Speaches oci-container recipe to `nix-international-telephony`, or
   should it stay a webphone-docs-only recipe until you deploy?
2. **Live-reality proof**: is fake-wire green sufficient for this train
   to count as verified, or do you want to hand me a throwaway GCP key
   (or stand up Speaches) for ONE real end-to-end transcription before
   M6/release?
3. **M6 timing**: run the stack browser E2E now (markup changed twice
   today; ~10 min + flake-retry budget) or batch it into the next
   release run?

*(Two older product questions also still open from 13-33 §g: island
429 cooldown Retry-After vs fixed 10 s; ship SILENCE_RMS 0.004 as-is.
Neither blocks M13/M7/gates.)*
