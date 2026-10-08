# Live transcription (all audio streams) — session status

- **Date:** 2026-10-07 16:54 CEST
- **Scope:** implement live transcription for every audio stream (calls,
  voicemail, MMS audio), server-side ASR seam, feature complete + tested.
- **Verdict:** backend seam + endpoint + island/shell client are DONE and
  green at the unit/integration level. The feature has **never run against
  a real ASR provider or a real browser**, so end-to-end behavior is
  UNPROVEN. One real defect (AudioContext leak) and several rough edges
  remain. `buildflow` and `nix flake check` were NOT run.

---

## a) FULLY DONE

### Backend

1. **Config seam** (`internal/config`): `asr.{url,token,model}`.
   Asymmetric enable rule (URL alone enables; token without URL fails
   closed). Absolute-http(s) URL validation. Positive + negative + family
   tests. (`internal/config/config.go`, `config_test.go`,
   `family_test.go` — all green.)
2. **ASR client** (`internal/asr/asr.go`): provider-agnostic
   OpenAI-compatible `/v1/audio/transcriptions`; multipart with the HONEST
   Content-Type; optional Bearer; `ErrDisabled`/`ErrUnauthorized` sentinels
   classified; every path errorfamily-coded (`asr.empty|request|transport|
   http|read|decode|encode|url`); `requestTimeout=90s`, `defaultModel=
   whisper-1`, 1 MiB response cap. Unit + family tests green.
3. **Endpoint** (`internal/server/transcribe.go`): `POST /api/transcribe`,
   session-gated (`requireSession`), 25 MiB `MaxBytesReader`, `?filename=`
   - `?lang=`, returns `{"text":…}`; disabled seam → 404; upstream failures
     → 502 (never a 401). Tests: session gate, disabled-404, round-trip,
     empty-400, provider-401→502, `config.js` flag.
4. **Composition root** (`internal/app/app.go`): `*asr.Client` provider,
   injected as `Deps.ASR`. `Deps.ASR` + `handlers.transcribeLimiter`
   (hook-grade bucket) + the route in `server.go`.
5. **`window.PBX_CONFIG.asr`** flag (`configjs.go`) + test. Safe on a nil
   client (`Enabled()` nil-receiver false).
6. **Error-code registry** regenerated for the 8 `asr.*` + `config.asr.url`
   codes; `TestErrorCodeRegistryIsFresh` green.

### Client

7. **`island/app/transcribe.js`**: one home for the seam —
   `transcribeBlob`, `transcribeUrl`, `transcribeNow`, and the live-call
   controller (`startLiveTranscription`/`stopLiveTranscription`/
   `isTranscribing`, `captureStream`, 4 s `MediaRecorder` segments). 6 new
   node:test specs green.
8. **Live call wiring** (`island/app/calls.js`): `asrEnabled`-gated
   transcribe toggle + transcript panel on every call card; button label
   reflects state; `teardownSession` stops capture.
9. **Island voicemail panel** (`island/app/panels.js`): `asrEnabled`-gated
   per-row transcribe button + transcript target.
10. **Server tabs** (`shell.js`): ONE delegated handler on
    `[data-transcribe-src]` (fetch same-origin audio → POST → render);
    buttons in `voicemail.templ` and `messages.templ` (audio attachments).
11. **i18n**: island en/de (6 keys) + server en/de (`vm.transcribe`,
    `media.transcribe`, `settings.asr`); parity tests green.
12. **Settings tab** reports the seam's state.
13. **CSS**: `.call-transcript`, `.vm-transcript`,
    `.wp-vm-transcript`, `.wp-attachment-transcript` (logical properties —
    the RTL gate passes).
14. **Docs**: CHANGELOG `[Unreleased]`, README (config rows + a Live
    transcription contract section), FEATURES row, AGENTS.md seam rule.

### Gates actually run

- `go build ./...` ✅
- `go test -count=1 ./...` ✅ (all packages, incl. arch/contract/CSP)
- island `node:test` — 192 pass ✅
- `oxlint` no-undef over island + shell.js ✅
- `go vet ./...` ✅
- `golangci-lint` (default linters) on new packages — 0 errcheck, 0 new ✅
- `nix fmt` (treefmt) ✅
- `scripts/whitespace-drift.sh` ✅

All work is committed (auto-commit daemon); working tree clean.

---

## b) PARTIALLY DONE / WEAK

1. **Live calls are NOT truly streaming.** 4 s `MediaRecorder` segments,
   one HTTP request each. Words are cut at chunk boundaries; no overlap
   between segments; latency = segment length + round-trip.
2. **No speaker separation.** Remote + operator audio is mixed into one
   mono stream, so the transcript cannot say who said what (a common need).
3. **Capture does not react to hold/mute.** `holdSession` disables receiver
   tracks; the recorder keeps producing (likely silent) segments. Unhandled.
4. **Live capture never passes a language hint.** `transcribeBlob` supports
   `language`; the live path always omits it (auto-detect per segment).
5. **No transcript persistence** anywhere (by design, but unconfirmed as a
   requirement).
6. **No SSE / cross-tab share.** A live transcript lives only in the tab
   that started it.
7. **On-demand shell.js feedback is English** while the button label is
   localized — a small inconsistency (justified by the D3 shell-English
   rule, but jarring next to a German button).
8. **Only Go-side greps guard shell.js.** No node:test behavior spec for the
   new `[data-transcribe-src]` handler (contrast: the existing
   `data-reload` handler IS covered by `TestShellJSHandlesReloadButtons`).
9. **No browser/E2E/live-smoke test.** The stack `browser-e2e.py`
   obligation is untouched; no loopback smoke exercising `/api/transcribe`.
10. **MMS attachment DOM ids carry a brand prefix** (`att-transcript-
    Attachment:<id>`) — valid but ugly; works only because
    `getElementById` tolerates `:`.
11. **`asr.NewClient` does not validate the scheme**; it trusts config
    validation. A direct call with a bad URL returns a client that fails at
    request time, not construction.

---

## c) NOT STARTED

1. `buildflow` (the repo's real quality gate) — **NOT RUN**.
2. `nix flake check` (package + sandbox tests + treefmt + island-lint +
   island-js + module check) — **NOT RUN**.
3. `nix build .#webphone` (the full Nix package build) — **NOT RUN**.
4. `scripts/webphone-smoke.py` live smoke with transcription enabled —
   **NOT RUN** (no provider available).
5. `nix run .#vulnix` — **NOT RUN**.
6. Stack browser E2E re-run — **NOT RUN** (owed after any served-markup
   change).
7. `TODO_LIST.md` / `ROADMAP.md` updates for the seam's follow-ups —
   **NOT DONE** (CHANGELOG/FEATURES/README/AGENTS only).
8. NixOS module documentation of `asr.*` (`settings` is freeform, so it
   works, but there is no explicit mention) — **NOT DONE**.
9. `docs/error-contract.md` prose section for the ASR seam (only the
   generated registry block was updated) — **NOT DONE**.

---

## d) TOTALLY FUCKED UP

1. **AudioContext leak (REAL DEFECT).** `captureStream()` creates a new
   `AudioContext` per `startLiveTranscription` and never `close()`s it.
   Browsers cap AudioContexts (~6 in Chrome); repeated start/stop across
   calls will exhaust them and silently break capture. **Must fix:** hold
   the context on the controller and `close()` it in
   `stopLiveTranscription`.
2. **"Verified" is over-claimed.** The feature is unit/integration-green but
   has **zero** end-to-end evidence against a real provider or a browser.
   The segmented-capture path in particular is only exercised by a stub
   `MediaRecorder`; real `MediaRecorder`/WebAudio behavior is unproven.
3. **One failing gate slipped in and was only caught by luck.** The RTL
   logical-property test (`TestStylesUseLogicalProperties`) failed on my
   `margin-left`, then passed after a fix — but I ran the asset tests late;
   a stricter "run the full suite after every edit" loop would have caught
   it sooner.

---

## e) WHAT WE SHOULD IMPROVE

1. **Fix the AudioContext leak** (d1) before anything ships.
2. **Run `buildflow` + `nix flake check`** — the two gates that actually
   decide red/green here.
3. **Add a shell.js behavior spec** for `[data-transcribe-src]` (mirror the
   `TestShellJSHandlesReloadButtons` precedent: a Go grep + a node:test).
4. **Make live capture honest about hold/mute** — pause/resume the recorder
   with the call state.
5. **Add a microphone/language config knob** and pass an explicit language
   to the provider for calls.
6. **Decide transcript persistence** (store with the call? show in History?)
   — currently deliberately dropped.
7. **Streaming/overlap**: consider overlapping segments or a provider with a
   true streaming endpoint to remove the word-boundary cuts.
8. **Accessibility**: announce new transcript text via the polite
   `#wp-live` region; give the transcript panel `aria-live="polite"`.
9. **Provider schema flexibility**: support the whisper.cpp native
   `/inference` shape as an alternative to OpenAI `/v1/audio/transcriptions`.
10. **Budget test for `/api/transcribe`** (429 path), mirroring the other
    limited surfaces.

---

## f) UP TO 50 THINGS TO DO NEXT

1. Fix the AudioContext leak in `transcribe.js` `captureStream`/`stop`.
2. Run `buildflow`; fix every finding.
3. Run `nix flake check`; fix every finding (module check, island-js).
4. Run `nix build .#webphone`.
5. Add a shell.js node:test spec for the transcribe delegation.
6. Add a Go served-asset grep test for `data-transcribe-src`.
7. Add an island test that a disabled `asr` flag hides the affordances.
8. Add a Go test that `messages.templ`/`voicemail.templ` emit the button
   only for audio attachments.
9. Pause the live recorder on hold; resume on hold-release.
10. Stop the recorder on mute of the remote (or keep — decide).
11. Pass a language hint for live calls (from `wp-lang`).
12. Add `asr.language` config option.
13. Close/cleanup the mixed `MediaStream` tracks on stop.
14. Guard against browser AudioContext cap (shared/one context).
15. Surface a "listening" visual state on the call card.
16. Announce transcript deltas via `#wp-live` (aria-live polite).
17. Add `aria-live="polite"` to `.call-transcript`.
18. Decide + implement transcript persistence with call history.
19. Add transcripts to the CDR/History tab if persisted.
20. Consider overlapping audio segments to reduce word cuts.
21. Add `/inference` (native whisper.cpp) support as provider variant.
22. Add a health/readiness leg for the ASR provider (optional, non-critical).
23. Log provider latency as a metric (`webphone_asr_seconds`).
24. Add a config-validation test for a non-http(s) `asr.url` at `NewClient`.
25. Add `asr.model` to the Settings panel for operator visibility.
26. Document `asr.*` in the NixOS module README/options comment.
27. Add the ASR seam to `docs/error-contract.md` prose.
28. Harvest follow-ups into `TODO_LIST.md`.
29. Add a `webphone-smoke.py --asr` mode with a fake provider.
30. Re-run the stack browser E2E after the markup change.
31. Add a rate-limit test for `/api/transcribe` (429).
32. Test oversized-body (413/400) on `/api/transcribe`.
33. Test the `lang`/`filename` query pass-through end to end.
34. Make MMS attachment DOM ids prefix-free (use the raw id value).
35. Unify shell.js feedback copy with the localized button (or accept D3).
36. Add a `transcribe.js` test for `transcribeNow` error rendering.
37. Add a `panels.js` test for the voicemail transcribe button.
38. Add a `calls.js` test for the transcribe toggle label transitions.
39. Consider a per-call transcript scroll/clear affordance.
40. Consider copying/saving a transcript (download).
41. Cap transcript length client-side (avoid unbounded DOM).
42. Handle provider empty-text mid-call gracefully (already skips; test it).
43. Add retry/backoff for a transient segment failure.
44. Add a "transcription unavailable" toast on repeated failures.
45. Confirm hold semantics: does `holdSession` keep receiver tracks
    live? (verify against capture).
46. Verify `hangup` path always reaches `teardownSession` → stop capture.
47. Add a memory-ceiling budget for concurrent live transcriptions.
48. Consider transcribing only the remote leg by default (privacy/quality).
49. Add a feature flag doc note: transcription sends audio off-box.
50. Add a CHANGELOG/README note that live transcription is near-live.

---

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Which ASR provider + model will ship** (self-hosted whisper.cpp URL? a
   hosted OpenAI-compatible API?), and is there a running instance I may
   point a smoke test at — or should I add a fake-provider harness instead?
2. **Should live-call transcripts be PERSISTED** (stored on the call record
   and shown in History), or stay ephemeral to the call card by design?
3. **Auto or manual?** Should transcription start automatically for every
   call/voicemail (privacy + cost implications), or stay strictly
   user-initiated as implemented?

---

## Reference — files touched this session

- Config: `internal/config/config.go`, `config_test.go`, `family_test.go`
- Seam: `internal/asr/asr.go`, `asr_test.go`, `family_test.go`
- Server: `internal/server/transcribe.go`, `transcribe_test.go`,
  `server.go`, `pages.go`, `configjs.go`, `panels.go`
- Composition: `internal/app/app.go`
- Island: `transcribe.js` (new), `calls.js`, `panels.js`, `config.js`,
  `i18n.js`; `shell.js`
- Views: `voicemail.templ`, `messages.templ`, `settings.templ`, `i18n.go`
- Styles: `island/style.css`, `app.css`
- Tests: `island-tests/transcribe.test.mjs` (new)
- Docs: `CHANGELOG.md`, `README.md`, `FEATURES.md`, `AGENTS.md`,
  `docs/error-contract.md` (generated registry)

---

## UPDATE 2026-10-07 18:00 — follow-up train (all three questions answered)

Owner decisions: **(1)** research 2026's best local-vs-API → done,
recommendation LOCAL; **(2)** transcripts PERSIST to History; **(3)**
auto-start EVERYWHERE. All follow-ups landed and every gate is green.

### Landed

- **Research verdict (README "Choosing a provider")**: LOCAL ASR — caller
  audio is GDPR personal data and the US-inference APIs (OpenAI, Groq,
  AssemblyAI) have no EU residency. Primary: Speaches (ex
  faster-whisper-server) + `large-v3-turbo` (OpenAI-compatible endpoint,
  own flake.nix, CPU-viable, MIT). Fallback: whisper.cpp `whisper-server`
  (`/inference` path needs a proxy rewrite). Cloud-if-ever: Deepgram EU.
  Also: always pass an explicit `language` (auto-detect misfires DE↔EN on
  4 s chunks) — the island now sends the session language with every
  live segment.
- **Persistence**: schema v3 `call_transcripts` table (versioned
  migration), `store.Transcripts` (Append + Recent grouped-by-call,
  5000-segment cap with a scalar-bounded trim), `POST /api/transcripts`
  (session-gated, validated, seam-gated), History tab renders the newest
  transcribed calls as their OWN section — deliberately NOT joined onto
  CDR rows (the phone API's CDRs carry no uuid; joining would be
  guesswork).
- **Auto-start everywhere**: calls auto-capture on Established
  (`ensureTranscription`); voicemail + MMS audio self-transcribe once per
  src+target per page life (shell.js `autoTranscribe`, PBX_CONFIG.asr
  gated, morph-safe). Buttons remain as manual re-runs.
- **AudioContext leak FIXED**: the mixing context is owned by the
  controller and `close()`d on stop (plus mixed-track teardown) — pinned
  by spec.
- **Hold honesty**: capture pauses on hold, resumes on release (no
  silence shipped to the provider). Mute stays honest by nature (the
  disabled sender track simply leaves the mix).
- **a11y**: start/stop announce politely; each segment delta rides the
  polite live region (clipped to 140 chars).
- **Gating gap FIXED**: the server-rendered voicemail/MMS transcribe
  buttons rendered unconditionally (dead buttons with the seam off) —
  now gated by `VoicemailPanelProps.ASR` / `ThreadViewProps.ASR`, and
  the Notifier carries the boot-time flag so SSE pushes match full
  renders.
- **shell.js spec** (owed from the first pass): behavioral specs for the
  delegated handler + auto-start + once-guard (island shell.test.mjs)
  and a served-asset tripwire (TestShellJSHandlesTranscribeButtons).

### Gates

`go test ./...` (20 pkgs) · 200/200 island node:test · oxlint · `nix
fmt` · `templ generate` · **buildflow** (findings gate passes) ·
**`nix flake check`** (all checks incl. the KVM backup VM test) ·
`nix build .#webphone`. CI green on the daemon's pushes.

### Incidental fixes (not mine, fixed to unblock the gate)

- `.buildflow.yml` had LOST its documented `skip_steps` section
  (truncated in the 00:46 daemon commit; AGENTS.md still documents the
  policy) — restored the 69 rationale lines + keys from git history.
  The findings gate was failing on go-structure-linter noise that was
  triaged 2026-09-18.
- erraudit `legacy_as` x2 in `gateway/family_test.go` /
  `store/db_test.go` (pre-existing): migrated `errors.As` →
  `errors.AsType` per the go-error-modernization decision tree (both
  were true type-extraction sites reading fields afterward).

### Still open

- No run against a REAL provider or browser (fake-provider tests only).
  When a Speaches instance exists: boot with `asr.url`, smoke it.
- The stack's browser E2E owes a re-run (served markup changed:
  ASR-gated buttons are now ABSENT by default — the E2E must not assume
  them; dom-contract ids unchanged).
