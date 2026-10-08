# SUPERB ASR — confidence decision + spike notes (M22–M25)

**2026-10-08 14:05** · closes M22 (decision), M23–M25 (spikes: NOTES,
not code) of the [SUPERB plan](2026-10-08_07-36_SUPERB-asr-provider-rollout.md).
Every claim below was verified against the live sources on this date
(links inline); nothing is remembered from hearsay.

---

## M22 — confidence / verbose-JSON: DECISION = REJECT the marker (F62 not adopted)

The plan offered a low-confidence UI marker "if the chosen provider
returns it". Verified reality on both live provider kinds:

- **Google STT V2** (`Alternative.confidence`, recognize reference):
  set ONLY for the top alternative of a non-streaming result; the docs
  say verbatim "This field is not guaranteed to be accurate and users
  should not rely on it to be always provided. The default of 0.0 is a
  sentinel value indicating `confidence` was not set." A literal 0.0
  and "no confidence provided" are indistinguishable on the wire.
- **Speaches** (OpenAI-compatible): `verbose_json` IS supported
  (`text|json|verbose_json|srt|vtt` in `routers/stt.py`) and segments
  carry `avg_logprob` / `no_speech_prob` — but there is NO calibrated
  confidence; any threshold over `exp(avg_logprob)` is a heuristic the
  operator would read as ground truth.

**Why reject:** the island honesty contract says the UI never shows a
state the network hasn't confirmed. A confidence badge built on a field
Google documents as unreliable, or on an invented logprob threshold,
would lie as often as it informs. The seam
(`asr.Provider.Transcribe -> text`) stays text-only.

**Revisit triggers** (either, in one verified train that adds the seam
method + UI marker together): Google populates reliable confidence
widely, or Speaches documents a confidence convention. Sources:
<https://cloud.google.com/speech-to-text/v2/docs/reference/rest/v2/projects.locations.recognizers/recognize>,
<https://github.com/speaches-ai/speaches/blob/master/src/speaches/routers/stt.py>.

---

## M23 — spike: streaming SSE/websocket transcription (replace 4 s chunking)

**Feasible, gated.** Verified against the Speaches repo (2026-10-08):

- `stream=true` on `POST /v1/audio/transcriptions` is streaming
  RESPONSE only (SSE partials of a fully-uploaded file) — useless for
  live mic capture; the request body still carries the whole file.
- The real thing exists: **`/v1/realtime` WebSocket** (OpenAI
  Realtime-compatible), with a Speaches extension
  `?intent=transcription` = live mic STT with no LLM turn. Events:
  `input_audio_buffer.speech_started/stopped/committed`,
  `conversation.item.input_audio_transcription.completed` (+ deltas).
  Auth via header or `api_key` query param.

**Sketch if adopted later:** the island already builds the mixed
AudioContext graph (`captureStream`); a WS variant would stream
`ScriptProcessor`/`AudioWorklet` chunks from that mix to
`/v1/realtime?intent=transcription`, replacing the MediaRecorder 4 s
cadence (latency drops from ~4-6 s to ~1 s). **Costs:** it is a
Speaches-only wire (Google has no such endpoint), so the island would
grow a provider-kind-specific path behind `PBX_CONFIG` — breaking the
"nothing outside `internal/asr` knows which wire answers" invariant
unless the server proxies the WS and re-exposes it generically. The
current segmented cadence (~15 req/min) is deliberately modest and the
silence guard now suppresses dead air.

**Verdict:** keep the 4 s chunking; revisit only if (a) local Speaches
becomes THE provider and (b) operators report the cadence as a product
problem. Source: <https://speaches.ai/usage/realtime-api/>.

---

## M24 — spike: native NixOS module for Speaches (replace the oci-container)

**Moderate effort, template exists.** Verified against nixpkgs
(2026-10-08):

- `speaches` itself: NOT in nixpkgs (no by-name, no python module, no
  NixOS module). Upstream flake remains devshell-only.
- The hard deps are DONE: `python3Packages.faster-whisper` (v1.2.1),
  `ctranslate2` (with `withCUDA`/`rocmSupport` options), `whisper-cpp`.
- The pattern exists: `wyoming-faster-whisper` ships an application
  package PLUS a full NixOS module
  (`services.wyoming.faster-whisper`) with model cache via
  `HF_HOME`/StateDirectory — the obvious template for
  `services.speaches`.

**Estimate:** 1–3 days for a nixpkgs-familiar packager: a
`buildPythonApplication` for speaches plus the missing Python deps
(likely gaps: `piper-tts`, `kokoro-onnx`; an STT-only build might
avoid the TTS deps if upstream's extras allow it). Until then the M2
oci-container recipe (plan F3–F9) stays the deployment story.
Sources: <https://github.com/NixOS/nixpkgs/blob/master/pkgs/by-name/wy/wyoming-faster-whisper/package.nix>,
<https://github.com/NixOS/nixpkgs/blob/master/nixos/modules/services/home-automation/wyoming/faster-whisper.nix>.

---

## M25 — spike: diarization (speaker labels)

**Shipped upstream, not stable enough to ride.** Verified against the
Speaches repo (2026-10-08):

- `POST /v1/audio/diarization` EXISTS (pyannote.audio pipeline;
  `json` segments with `SPEAKER_00` labels or `rttm`), plus a
  companion `/v1/audio/speech/embedding` (wespeaker ONNX) for speaker
  verification.
- The killer feature for a phone app: `known_speaker_names[]` +
  `known_speaker_references[]` (audio references) — diarized speakers
  map to KNOWN names by embedding cosine similarity. For a two-party
  call that means automatic "operator" vs "caller" line labels.
- Open issues at check time: pipeline memory leak on unload (#629),
  missing model TTL config (#631), over-segmentation with more
  speakers than present (#661), an unsafe-model-loading security
  report in the diarization executor (#679).

**Verdict:** do not build on it yet. Revisit when the open issues
close; the integration point would be a second optional call on the
server seam (post-call, over the stored transcript's audio), never a
live-path dependency. Source:
<https://github.com/speaches-ai/speaches/blob/master/src/speaches/routers/diarization.py>.
