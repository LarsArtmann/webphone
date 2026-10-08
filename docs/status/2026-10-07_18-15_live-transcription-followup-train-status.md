# Status Report — Live Transcription Follow-Up Train

**2026-10-07 18:15 CEST** · session scope: the ASR follow-up train (provider
research, persistence, auto-start, defect fixes, gates) on top of the
16-54 live-transcription build. Point-in-time snapshot; prior report:
`2026-10-07_16-54_live-transcription-all-audio-streams-status.md`.

---

## a) FULLY DONE

1. **Provider research (owner question 1) — decided and documented.**
   Two-source web research (self-hosted landscape + hosted APIs). Verdict:
   LOCAL ASR. Caller audio is GDPR personal data; OpenAI/Groq/AssemblyAI
   are US-inference-only (no EU residency). Recommendation: **Speaches**
   (ex faster-whisper-server) + `large-v3-turbo` (first-class
   OpenAI-compatible `/v1/audio/transcriptions`, ships its own flake.nix,
   CPU-viable for 4 s segments, MIT); whisper.cpp `whisper-server` as the
   lightweight fallback (`/inference` path needs a proxy rewrite);
   Deepgram EU only if cloud is ever mandated. Documented in README §
   "Live transcription" with reasoning + the explicit-language rule
   (auto-detect misfires DE↔EN on short phone chunks).
2. **Persistence (owner decision 2).** Schema v3 `call_transcripts`
   (versioned migration in `store/db.go`; never ad-hoc), `store.Transcripts`
   (Append + Recent grouped-by-call, newest call first, spoken order
   inside; 5000-segment cap per extension via a scalar-bounded range
   trim), `POST /api/transcripts` (session-gated, seam-gated, validated,
   nil-store 503), wired through samber/do (`Deps.Transcripts`). History
   tab renders a "Call transcripts" section (own CSS, i18n en/de,
   CRM-name enrichment) — **deliberately NOT joined onto CDR rows** (the
   phone API's CDR struct carries no uuid; joining would be guesswork,
   not correlation).
3. **Auto-start everywhere (owner decision 3).** Calls: capture begins on
   Established (`ensureTranscription`), announced politely; voicemail +
   MMS audio: shell.js `autoTranscribe` self-transcribes once per
   src+target per page life (PBX_CONFIG.asr gated, debounced, morph-safe
   — filled targets never re-run). Buttons remain as manual re-runs.
4. **AudioContext leak FIXED.** The mixing context is owned by the
   capture controller and `close()`d on stop; mixed tracks stopped.
   Spec-pinned (`stopLiveTranscription closes the AudioContext…`).
5. **Hold honesty.** Capture pauses on hold, resumes on release
   (`pauseLiveTranscription` wired into `holdSession` after the re-INVITE
   settles) — the provider never receives (and hallucinates on) hold
   silence. Mute stays honest by nature: the disabled sender track
   simply leaves the mix.
6. **Language hint.** Every live segment carries the session language
   (`getLang()` → `?lang=`), per the research finding.
7. **a11y.** Start/stop announced; each segment delta rides the polite
   live region (clipped to 140 chars); failed persistence warns ONCE per
   call and never blocks capture.
8. **Gating gap FIXED (found this session).** The server-rendered
   voicemail/MMS transcribe buttons rendered UNCONDITIONALLY — dead
   buttons with the seam off (production default). Now gated:
   `VoicemailPanelProps.ASR`, `ThreadViewProps.ASR` → `Bubble(asr)`;
   the Notifier carries the boot-time flag so SSE pushes match full
   renders. View specs pin both directions.
9. **shell.js behavior spec (owed from the first pass).** Delegated
   handler + auto-start + once-guard specs in `shell.test.mjs`, plus a
   served-asset tripwire `TestShellJSHandlesTranscribeButtons` (mirrors
   the reload-buttons precedent).
10. **Gates — all green, first time this feature has been under them:**
    `go test ./...` (20 pkgs) · 200/200 island node:test · oxlint ·
    `nix fmt` · `templ generate` · **buildflow** (findings gate passes) ·
    **`nix flake check`** (all checks incl. the KVM backup VM test) ·
    `nix build .#webphone` (no vendorHash change — no new deps) · CI
    green on the daemon's pushes.
11. **Incidental unblocking (not my code, fixed because the gate was
    red):** restored `.buildflow.yml`'s documented `skip_steps` section
    (69 lines of rationale + keys, truncated by a daemon commit at
    00:46; AGENTS.md still documented the policy — pure drift), and
    migrated two pre-existing erraudit `legacy_as` findings
    (`gateway/family_test.go`, `store/db_test.go`) `errors.As` →
    `errors.AsType` per the go-error-modernization tree (both genuine
    type-extraction sites).
12. **Docs:** README (provider recommendation + new contract bullets),
    CHANGELOG (auto-start/persistence entry), FEATURES row updated,
    AGENTS.md ASR-seam bullet rewritten, 16-54 status report appended
    with an 18:00 update.

## b) PARTIALLY DONE

1. **The feature overall:** unit/integration-green under ALL repo gates,
   but **zero real-world verification** — no Speaches instance exists
   yet, no browser run with audio flowing. Auto-start's browser-side
   behavior is untested (see d-3 for the concrete autoplay risk).
2. **Provider deployment:** a recommendation on paper. Nothing in the
   consuming stack (`nix-international-telephony`) provisions a
   Speaches/whisper service or wires `asr.url` — the operator must
   deploy one before the seam can be enabled anywhere.
3. **Docs corpus:** updated README/CHANGELOG/FEATURES/AGENTS, but
   **TODO_LIST.md was NOT harvested** with this session's follow-ups
   (docs-health HARVEST debt — see f-1).
4. **Retention story:** transcripts are count-capped only (5000/ext).
   No age-based sweep (`Sweep` doesn't touch `call_transcripts`), no
   delete affordance (GDPR erasure). Deliberate first cut, unfinished
   policy.

## c) NOT STARTED

1. Stack-side work entirely: Speaches NixOS service/module, `asr.*`
   settings plumbed through the consuming stack's module defaults.
2. Real-provider smoke (boot with `asr.url` → live call → transcript).
3. Stack browser E2E re-run — OWED per AGENTS (served markup changed:
   ASR-gated buttons are now absent by default; the E2E must not assume
   them). dom-contract ids unchanged; contract test green.
4. Transcript deletion (per-call or bulk) and retention sweep wiring.
5. Client-side resilience: retry/backoff for dropped segments and
   failed saves (currently log + once-toast, then gone).
6. `ctx.resume()` on the capture AudioContext (autoplay policy — see
   d-3).
7. TODO_LIST harvest + ROADMAP fuel from this report's section f.

## d) TOTALLY FUCKED UP!

Nothing shipped broken — every defect below was caught by a gate or a
spec I wrote, before landing. Honest inventory of the session's own
fuckups and near-misses:

1. **First-pass trim SQL was O(cap) per append.** I wrote a
   `NOT IN (SELECT … LIMIT 5000)` keep-list that materializes 5000 ids
   on EVERY segment append; my own cap test took 9.6 s and failed
   through the read window. Rewrote as a scalar-bounded range delete.
   Lesson: per-row triggers deserve a cost thought at write time, not
   after a red test.
2. **I initially wrote a placeholder `t.Skip` test with dead code**
   (`server.deps`, unused imports) into `transcripts_api_test.go`,
   then deleted it. Never should have existed — noise in a pinned
   suite.
3. **Unverified browser assumption baked into auto-start:** Chrome
   starts AudioContexts "suspended" without a user gesture. Outbound
   calls have the dial click; inbound have the answer click; but a page
   RELOAD mid-call re-auto-starts capture with NO gesture → potentially
   a suspended context recording silence. I never call `ctx.resume()`.
   Might be fine (gesture propagation is generous), might silently
   transcribe nothing after a reload. Needs a browser test or a
   defensive `ctx.resume()`.
4. **Edit discipline vs the daemon:** SEVEN edit attempts failed on
   "file modified since last read" (daemon commits touch mtimes;
   AGENTS literally warns to re-View before each Edit). Wasted round
   trips; on the first voicemail.templ retry I also burned a cycle
   because I miscounted tab indentation in an old_string.
5. **templ/Go slips that the compiler caught:** a ternary in templ
   markup (`{ a ? b : c }` — not valid Go/templ), and the `asr` flag
   edit targeted `transcriptBody` while the audio block actually lived
   in `Bubble`. Also `window.PBX_CONFIG && PBX_CONFIG.asr` (bare
   identifier — ReferenceError under node; works in a browser by
   accident). All caught by generator/specs — but each was a
   preventable authoring error.
6. **Gates were run only at the END of the train.** The previous
   session shipped a whole feature without buildflow; I repeated the
   shape (build first, gate last). Consequence: the `.buildflow.yml`
   truncation sat undetected for ~16 h and I discovered it only when
   the findings gate went red — after all implementation. Running
   buildflow FIRST would have surfaced the broken config immediately.

## e) WHAT WE SHOULD IMPROVE!

1. **Run buildflow at train START, not just at the end** — it verifies
   the gate CONFIG, not just the code. This session's biggest process
   miss (see d-6).
2. **AudioContext lifecycle helper:** one shared "unlocked context"
   util (`resume()` guarded) for capture + ringback + waveform, instead
   of three independent assumptions about autoplay policy.
3. **Client resilience policy:** live segments and persistence saves
   are fire-and-forget with no retry. A tiny bounded retry (or 429
   backoff in shell auto-start) would make auto-start-everywhere safe
   under bursts (transcribe limiter = 60/min per IP; a long call at
   15/min + a voicemail page auto-transcribing can collide).
4. **Retention/erasure policy for transcripts** — they're call CONTENT
   at rest: decide sweep integration + a delete affordance.
5. **TODO_LIST discipline:** land follow-ups into TODO_LIST.md the
   moment they're identified, not "next session".
6. **Test-suite runtime:** the transcript cap test appends 5010 rows
   (~5 s, now the slowest store spec). A stride var or smaller test cap
   would keep the suite snappy.
7. **`Recent` reassembly prepends lines** (O(n²) on long calls) — fine
   at current caps, sloppy in principle.
8. **AGENTS.md line budget:** the ASR bullet is getting long; next
   addition should move evidence to docs/ (the cap is 377 lines).

## f) Up to 50 things we should get done next

**Verify what exists (highest value first):**

1. Deploy a Speaches instance (nixosModules candidate in the consuming
   stack) with `large-v3-turbo`; point `asr.url` at it.
2. Real-provider smoke: boot the binary with `asr.url`, place a
   loopback call, confirm segments + History rendering end to end.
3. Browser test of auto-start after mid-call page RELOAD (autoplay/
   suspended-context risk, d-3); add `ctx.resume()` if needed.
4. Re-run the stack browser E2E (owed: served markup changed).
5. Verify 429 behavior under combined load (live call + voicemail page
   auto-start); add backoff/serialization if it red-flags.

**Harden:**
6. Defensive `ctx.resume()` in `captureStream`.
7. Bounded retry for failed `/api/transcripts` saves (per-call queue,
flushed on stop).
8. 429-aware backoff in shell.js `autoTranscribe` (serialize + retry
once after Retry-After).
9. Transcript delete affordance (per-call `DELETE /api/transcripts?…`

- History UI) — GDPR erasure.

10. Age-based retention for `call_transcripts` (Sweep integration,
    config-gated cutoff like messages).
11. Silence/VAD guard client-side: skip POSTing segments that are
    near-silent (cheap RMS check) to spare the provider hallucinations.
12. Cap per-CALL segments too (a 6 h call alone could eat the whole
    per-extension budget).
13. Show a "saved to History" affordance (or transcript count) so
    persistence is discoverable, not silent.

**Docs/process debt:**
14. docs-health HARVEST this report's section f into TODO_LIST.md/
ROADMAP.md.
15. Add the `.buildflow.yml` truncation incident + "gate first" lesson
to docs/lessons.md.
16. AGENTS.md: consider a one-line "run buildflow before AND after big
trains" rule.
17. Re-measure error-family tiers if the next monthly window lands
(due 2026-11-05).

**Nice-to-have / polish:**
18. Transcript search in History (query box already exists for CDRs —
extend to transcript text).
19. Copy-to-clipboard on transcript sections.
20. Diarization/speaker labels (upstream: Speaches + pyannote) — only
if the provider supports it cheaply.
21. Streaming SSE transcription (Speaches supports it) instead of 4 s
chunks — bigger client change, defer.
22. Export transcripts with the settings zip.
23. Settings panel: show asr model name alongside the on/off state.
24. Island spec: pause/resume on the QUEUED hold path (holdQueued
mid-flight) — currently only the settled path is spec'd.
25. Go spec: transcript store under CONCURRENT appends (the trim's
worst case).
26. Reduce the cap-test runtime (stride var or smaller fixture).
27. `Recent` line assembly without prepend.
28. Consider `no_speech_prob`/confidence handling if the provider
returns verbose JSON (currently plain text only).
29. `.buildflow.yml`: add a comment warning that skip_steps is
load-bearing policy (another silent truncation would quietly
re-fail the gate — or worse, quietly PASS noise).
30. ui-capture harness: add an ASR-on shot set (T23 extension) so the
visual regression covers the gated buttons.

## g) Questions I cannot answer myself

1. **Is there a Speaches (or other OpenAI-compatible ASR) instance I may
   point at — or should the NEXT train be "deploy Speaches into the
   consuming stack as a NixOS module" (and if so: GPU host available,
   or CPU-only)?** This decides whether we smoke first or build
   deployment first.
2. **Transcript retention & erasure policy:** is the 5000-segment count
   cap alone acceptable, or do you want age-based sweeping (tied to the
   messages retention config) and a user-facing delete? These are call
   RECORDS of real people — I don't want to guess the compliance bar.
3. **If browser testing confirms the suspended-AudioContext risk after
   reloads:** demote call auto-start to "first user gesture" (a one-time
   click banner), or always fire and let it silently produce nothing
   until interaction? (Privacy/cost vs. simplicity tradeoff only you
   can call.)

---

_Report format note: written as `.md` per the explicit prompt instruction
(canonical status-report format is styled HTML; override flagged)._
