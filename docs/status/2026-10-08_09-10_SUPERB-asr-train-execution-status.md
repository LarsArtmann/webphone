# Status Report — SUPERB ASR Train Execution (mid-train snapshot)

**2026-10-08 09:10** · scope: THIS SESSION ONLY — the execution of
[`docs/planning/2026-10-08_07-36_SUPERB-asr-provider-rollout.md`](../planning/2026-10-08_07-36_SUPERB-asr-provider-rollout.md)
(M1–M25 / F1–F69), started ~08:00 from clean `f9f8d5e`, interrupted for
this report ~70 minutes in. Not a project-wide audit.

---

## a) FULLY DONE (verified green this session)

1. **Baseline gates at train START** (the `.buildflow.yml` lesson, now
   honored): `go build ./...` ✓, full `go test -count=1 ./...` 21/21
   packages ✓, `buildflow` executed and completed with the restored
   `skip_steps` config intact (job ran in background — verdict tail
   showed no findings-gate red; see d)3 for the sloppiness).
2. **M1 — decision record** (§9 appended to the plan): both provider
   paths chosen (Google adapter in-repo + Speaches recipe on the stack),
   retention = age-sweep tied to `retention_days` + end-user delete,
   autoplay = never-silent (`ctx.resume()` + first-gesture retry). Each
   documented with rationale and reversibility.
3. **M3 — Google Cloud STT v2 provider, complete and green**:
   - Verified the REAL wire shape first (docs research): regional host
     `https://speech.{location}.rep.googleapis.com`, ad-hoc recognizer
     path `…/recognizers/_:recognize`, base64 `content` in JSON,
     `x-goog-api-key` header auth, `autoDecodingConfig` sniffing, no
     per-request billing minimum.
   - `internal/asr/google.go`: `GoogleClient` (EU-resident defaults:
     `europe-west3`, model `telephony`, language floor `de-DE` with
     per-request override), 429 → Transient, 401/403 → shared
     `ErrUnauthorized`.
   - `asr.Provider` seam interface (`Enabled` + `Transcribe`); the
     OpenAI-compatible `Client` and `GoogleClient` both satisfy it
     (pinned by a compile-time spec).
   - Config: `asr.provider`/`asr.project`/`asr.location`/`asr.language`
     + validation (unknown kind, google-without-project, path-shaped
     ids rejected; google token-without-url now LEGITIMATE) +
     `Enabled()` semantics per kind. 5 new config/family test cases.
   - Composition-root branch; `server.Deps.ASR` is now the interface.
   - **Found + fixed the nil-interface trap** the type swap created:
     `configjs.go`/`panels.go` called `.Enabled()` on a nil interface
     (the old concrete nil-receiver was safe) → routed both through the
     one-home `asrOn()` helper.
   - Error-code registry regenerated (`config.asr.provider`,
     `config.asr.project`; `asr.http` now runtime-split across two
     files).
   - Tests: `google_test.go` (request shape incl. recognizer path +
     api-key header + base64, language fallback chain, error family
     table, response join) + `google_internal_test.go` (regional host
     derivation — the EU-residency guarantee).
   - Docs: README config-table rows for all new keys + a full Google
     recipe (IAM/API-key/DPA one-time setup, cost math); CHANGELOG
     `[Unreleased]` entry.
4. **M4 — end-to-end smoke, the M4 core (server half)**: new
   `asr_scenario` in `scripts/webphone-smoke.py` with in-process fake
   providers for BOTH wires. 9/9 checks green: `/config.js` carries
   `"asr":true`, session login, `/api/transcribe` round-trip per wire,
   honest multipart on the OpenAI wire, `POST /api/transcripts` → 204,
   History partial renders the stored text, Google wire hits the exact
   recognizer path with the API key. Full smoke verdict: 47 + 4 + 8 + 9
   passed, 0 failed.
5. **M5 — autoplay/suspended-context fix**: `ensureRunning(ctx)` in
   `transcribe.js` — immediate `resume()`, one-shot `pointerdown` +
   `keydown` retries armed when the context came up suspended (Chrome
   keeps `resume()` pending until a gesture), an English `#log` warn
   line (never silent), inert after close. Island specs: suspended →
   resumed + both retries armed; running → nothing armed; closed →
   gesture is a no-op. `helpers.mjs` window stub gained `addEventListener`
   (was missing entirely). F23 (real-browser reload-mid-call) is NOT
   locally executable — no SIP registrar on this host; owed to the
   stack browser E2E (documented).
6. **M8/M9 island + shell implementation code** (see b) for the spec
   gaps): shell `postTranscribe` retries ONCE on 429 honoring
   Retry-After (bounded to 10 s, hostile headers can't freeze the tab);
   `runTranscribe` now returns its promise; auto-runs serialize through
   one promise chain (manual clicks stay immediate). Island live loop:
   10 s cooldown after a 429 (segments skipped, capture alive, one warn
   log). `calls.js`: per-call save queue — failed segments queue
   in order, cap 50 drop-oldest, one-at-a-time flush on next success +
   a final flush attempt in `teardownSession`; once-per-call warn toast
   preserved.

## b) PARTIALLY DONE

1. **M8 specs (F31 etc.)**: the island 429-cooldown spec is written but
   its first run tripped on a stub-fidelity bug (response stub without
   `headers` crashed `authedFetch`'s `noteThrottled`) — stub fixed with
   `new Headers()`, **not yet re-run green**. The shell-side spec
   (serialization: one auto-run in flight; retry-once on 429) is **not
   written at all** yet.
2. **M9 specs**: order-preserving flush spec GREEN. The **cap spec is
   RED and untriaged** (see d)1) — either my expectation of which
   segments survive the cap is wrong (microtask interleaving of the
   first in-flight save vs. the 52 pushes) or the drop-oldest arithmetic
   genuinely drops one segment too many. Code and the two test files
   were the 3 uncommitted files at report time (daemon will pick them
   up).
3. **M3 residual**: `nix fmt` not yet run over the train's Go/JS edits
   (deferred to train-end gate per repo habit — but island/shell edits
   specifically owe `nix fmt` BEFORE the next gate run per AGENTS.md).

## c) NOT STARTED (plan order)

M2 (Speaches stack service), M6 (stack browser E2E + dom-contract
re-run), M7 (docs-health harvest), M10 (retention sweep — the
`internal/retention` package + `store.Sweep` extension point already
identified), M11 (DELETE endpoint + History button), M12 (per-call cap
+ RMS silence guard), M13 (ui-capture ASR shots), M14 (lessons + AGENTS
rule), M15 (History transcript search), M16 (copy buttons), M17 (export
zip leg), M18 (Settings provider/model row), M19 (queued-hold spec),
M20 (concurrency spec + cap-test speedup), M21 (Recent() single-pass),
M22 (confidence decision), M23–M25 (spike notes), final gates
(buildflow + `nix flake check` + CHANGELOG sweep + CI/ls-remote
verify).

## d) TOTALLY FUCKED UP (this session's honest list)

1. **The M9 cap spec is red and I paused mid-debug** — the session
   interrupt landed with a failing test in the tree. Root cause not yet
   isolated (test-expectation arithmetic vs. real off-by-one).
2. **Background gate jobs started and not watched**: baseline buildflow
   (job ran to completion — I only checked it AFTER being asked for
   this report) and a stray hung test run I had to kill. The
   "gates-first" rule was honored in letter, sloppy in spirit.
3. **The placeholder-test near-miss**: while splitting an unexported
   helper test I briefly wrote a `t.Skip("placeholder…")` spec — the
   EXACT anti-pattern the 10-07 status report d) called out. Caught it
   in the same minute and replaced it with the real internal test file,
   but the reflex existed.
4. **M4 smoke scenario first run 8/9 red** because I forgot login needs
   CSRF adoption from `GET /` — a pattern `restart_scenario` already
   implements verbatim. Re-learned instead of copied.
5. **Daemon interference cost 3 edit-failure round trips** (mtime bumps
   between read and edit) — the known ritual; each recovered by
   re-View, but I keep getting surprised instead of writing
   atomically from the start.
6. **Narrative commits not made**: the daemon swept M3/M5/M8 work into
   `chore: auto-commit N changed file(s)` commits; the release-runbook
   "narrative commit per phase boundary" obligation is unmet for this
   train so far (history readability suffers; the plan/decision record
   carries the narrative instead).

## e) WHAT WE SHOULD IMPROVE (process, observed this session)

- **Watch every background gate immediately** or run it foreground;
  "started" is not "green".
- **Copy the working pattern first** (CSRF adoption, scenario
  skeletons) before inventing a variant — the repo already encodes
  most rituals as code.
- **Island test stubs should mirror the real `Response` shape** (headers
  object on EVERY stub) — a `makeJsonResponse()` helper in
  `helpers.mjs` would kill this class of false failure; `noteThrottled`
  is not the last reader of `res.headers`.
- **Commit-sized verification units**: run the specific spec file
  (seconds) after every island edit instead of batching the suite —
  three of my failures were one-edit-deep.
- **Interface swaps ripple**: replacing a concrete type with an
  interface silently changes nil-semantics; grep for unguarded method
  calls on the field BEFORE running the suite next time (the suite
  caught it here — loudly, via panic).

## f) Next up to 50 (execution order; M-numbers map to the plan)

1. Triage + fix the M9 cap spec (red line in the tree — first action).
2. Re-run the 429-cooldown spec green; finish F31 (shell serialization
   + retry-once spec in `shell.test.mjs`).
3. `nix fmt` over the train so far, then `go test ./...` + island suite
   full green checkpoint.
4. M10: extend `store.Sweep` + `SweepResult` with `call_transcripts`
   (shared cutoff), wire the count into `retention.Start` logging; sweep
   tests + `ReadCounts` inclusion.
5. M11: `Transcripts.DeleteCall(owner, callID)` store method + test.
6. M11: `DELETE /api/transcripts?call=` handler (session-gated,
   seam-gated 404, owner-scoped) + server spec (round-trip, cross-owner
   404, seam-off).
7. M11: History "delete transcript" button + confirm + i18n en/de +
   views spec; dom-contract check if ids change.
8. M12: per-call segment cap in `Transcripts.Append` + store test.
9. M12: island RMS silence guard before POST (+ spec; decode via the
   mixing context).
10. M15: History search — SQL `LIKE` over `call_transcripts.text`
    merged with CDR hits.
11. M15: view hint when a transcript matched (i18n both maps).
12. M15: search tests (match/no-match/owner scope).
13. M16: copy-to-clipboard on History transcript sections (+ spec).
14. M16: copy affordance in the call card transcript (+ spec).
15. M16: copy for voicemail/MMS transcript targets in shell.js (+ spec).
16. M17: transcripts leg in the settings export zip (+ test).
17. M18: Settings panel provider kind + model row (props from config).
18. M19: queued-hold (`holdQueued` mid-flight) pause/resume spec.
19. M20: concurrent `Append` spec (goroutines; cap holds).
20. M20: cap-test speedup (stride var / smaller fixture, < 2 s).
21. M21: `Recent()` single-pass assembly (no O(n²) prepend).
22. M22: confidence decision note (Google `confidence` 0.0-sentinel
    argument documented; no plumbing unless owner wants it).
23. M2: Speaches `virtualisation.oci-containers` service + stack module
    options + stand-ins (stack repo, AFTER webphone push; see g)3).
24. M2: loopback-only network + systemd hardening overrides.
25. M2: HF model cache volume pre-provision + offline-boot note.
26. M2: pin image digest + stack README section.
27. M13: ui-capture ASR-on boot (selenium harness) + new shots
    (settings row, call card with transcript, voicemail button).
28. M13: DOM asserts per new shot; baseline regen.
29. M6: re-run `TestServedPageHoldsTheDomContract` after all markup
    changes land (M11/M15/M16 first — E2E after markup settles).
30. M6: stack browser E2E re-run (budget 445 s; one retry on the known
    transfer flake) — includes the F23 reload-mid-call ASR check.
31. M14: docs/lessons `.buildflow.yml` truncation incident entry.
32. M14: AGENTS.md "run buildflow at train START" one-liner (377 cap).
33. M23: streaming-SSE transcription spike NOTE (docs/planning).
34. M24: native Speaches NixOS module packaging spike NOTE.
35. M25: diarization options spike NOTE.
36. M7: docs-health HARVEST — this report's §f + the plan + the two
    10-07 status files → `TODO_LIST.md` / `ROADMAP.md`.
37. CHANGELOG sweep for every code-touching medium task still missing
    one (M5/M8/M9 done? verify at train end).
38. FEATURES.md: transcription row gains persistence/delete/search
    status updates.
39. AGENTS.md ASR bullet: google provider kind + queue/cooldown facts.
40. error-contract.md regen after M10/M11 add codes.
41. Final gates: `nix fmt` → buildflow → `nix flake check` →
    `nix build .#webphone`.
42. Final smoke re-run (full suite incl. asr_scenario) on the built
    binary.
43. Push discipline: verify CI green + `git ls-remote` after daemon
    pushes (2026-10-05 rule).
44. vulnix gate (`nix run .#vulnix`) — release gate rides the train.
45. Consider `helpers.mjs` `makeJsonResponse()` stub helper (e item).
46. Consider narrative `git commit` for the remaining phase boundaries
    (daemon races it; try committing at each green checkpoint).
47. M4 browser half (F17/F19): real-browser call transcript — fold into
    the stack E2E ASR leg rather than a local harness (no local PBX).
48. Check `docs/dedup-registry.md` sweep-log line for the train
    (repo ritual) if any similarity rulings emerged.
49. Re-measure nothing tier-3; monthly error-family tier-1+2 check is
    due 2026-11-05 (early is fine if touching error paths).
50. Post-train: status appendix marking M-tasks done (docs-health
    ANNOTATE convention).

## g) Questions I can NOT answer myself

1. **Which provider goes live FIRST (real verification target)?** Both
   code paths now exist and are fake-tested. Speaches = ~3–4 GB RAM on
   the PBX host, €0 marginal, needs the stack service (g)3); Google
   Frankfurt = $0.96/audio-hr, needs a GCP project + restricted API key
   + DPA acceptance on your side. Which do I wire for the first
   real-audio verification?
2. **Retention detail**: transcripts riding the GLOBAL
   `retention_days` (same window as messages/faxes) — confirmed OK, or
   do you want a separate, SHORTER transcript cutoff (they are call
   content, arguably more sensitive than CDR metadata)?
3. **May I commit to the stack repo** (`~/projects/nix-international-
   telephony`, present on this host) for M2/M6, following the ritual
   (webphone pushed first, stack tree clean, then stack commit)? Or
   should stack changes stay operator-authored and I only deliver the
   module draft + recipe as documents?
