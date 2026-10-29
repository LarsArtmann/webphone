// Live + on-demand transcription. This is the ONE client home for the
// /api/transcribe seam: every audio surface (live call capture, voicemail,
// MMS attachments) funnels through transcribeBlob/transcribeUrl, so the
// wire contract lives in exactly one place. The whole feature is gated on
// PBX_CONFIG.asr (the server seam's Enabled() state) — with no provider
// configured the island never POSTs and hides its affordances.
//
// Live call capture records short MediaRecorder segments of the call's
// audio (remote party + the operator's own voice) and appends each
// segment's transcript. Segmented "near-live" is deliberate: the
// OpenAI-compatible transcription endpoint takes whole files, so streaming
// one long file is not an option without a stateful streaming protocol the
// self-hosted providers do not share.

import { authedFetch } from "./auth.js";
import { asrEnabled } from "./config.js";
import { t } from "./i18n.js";
import { announce, errorMessage, log } from "./ui.js";

export const transcribeEnabled = asrEnabled;

// saveTranscriptSegment persists one live-call segment to the server's
// owner-scoped transcript store (the History tab's durable source).
// Fire-and-forget: a failed save costs the LATER re-read, not the live
// card — the caller logs/toasts, never blocks the capture loop.
/** @param {IslandTranscriptSegment} segment */
export async function saveTranscriptSegment(segment) {
  const res = await authedFetch("/api/transcripts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(segment),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
}

// Segment length: short enough to feel live, long enough that the provider
// gets usable context and the request rate stays modest (~15/min).
const CHUNK_MS = 4000;

// Near-silence floor for the segment gate (RMS over the analyser's
// tail window, ≈ -48 dBFS): below any speech, above the codec noise
// floor. Whisper-family models hallucinate confident text on pure
// silence — the guard spends nothing on segments that carry no voice.
const SILENCE_RMS = 0.004;

// How long EVERY surface goes quiet after the seam answered 429 (a couple
// of segment cycles — enough for the flood budget to recover). The deadline
// is shared across the island's surfaces (#5): a 429 on any island POST
// (live segment, voicemail button, any transcribeUrl caller) arms it, and
// the live loop — the only perpetual source of POSTs — honors it no matter
// which surface tripped it. The shell's classic-script auto path keeps its
// own serialized queue by design (the shell must work when island scripts
// fail); the server's per-IP bucket remains the cross-script guard.
const COOLDOWN_MS = 10_000;

// Live-loop 429 cooldown deadline (armed on any surface's 429).
let cooldownUntil = 0;

// Transient-failure backoff (#43): after a failed segment the loop skips
// the next ones for a doubling window (2s → 60s cap) instead of retrying
// at full cadence against a provider that is clearly struggling.
const BACKOFF_BASE_MS = 2000;
const BACKOFF_CAP_MS = 60_000;
// Repeated-failure visibility (#44): after this many consecutive
// non-429 failures, tell the operator ONCE per failure streak that the
// transcription is effectively down — the per-segment log lines are too
// quiet for a call the operator is actively on.
const FAILURE_TOAST_AT = 3;
let liveFailures = 0;
let backoffUntil = 0;
let failureToasted = false;

// transcribeBlob posts one encoded audio blob and returns the trimmed
// text ("" when the provider heard nothing).
/**
 * @param {Blob} blob
 * @param {{filename?: string, language?: string}} [options]
 * @returns {Promise<string>}
 */
export async function transcribeBlob(
  blob,
  { filename = "audio.webm", language } = {},
) {
  const query = new URLSearchParams();
  if (filename) query.set("filename", filename);
  if (language) query.set("lang", language);
  const res = await authedFetch(`/api/transcribe?${query}`, {
    method: "POST",
    headers: { "Content-Type": blob.type || "application/octet-stream" },
    body: blob,
  });
  if (!res.ok) {
    // A 429 arms the SHARED cooldown (#5): the server's flood budget is
    // per-IP, so a limit hit from any surface limits every surface.
    if (res.status === 429) cooldownUntil = Date.now() + COOLDOWN_MS;
    throw new Error(`HTTP ${res.status}`);
  }
  const data = await res.json();
  return String(data.text || "").trim();
}

// transcribeUrl fetches a same-origin audio URL (voicemail playback,
// MMS attachment) into a blob and transcribes it.
/**
 * @param {string} url
 * @param {{filename?: string, language?: string}} [options]
 * @returns {Promise<string>}
 */
export async function transcribeUrl(url, { filename, language } = {}) {
  const res = await authedFetch(url);
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  const blob = await res.blob();
  return transcribeBlob(blob, {
    filename: filename || url.split("/").pop() || "audio",
    language,
  });
}

// --- live call capture ------------------------------------------------------

// live maps a call id to its running controller (recorder + metadata).
/** @type {Map<string, IslandCaptureController>} */
const live = new Map();

/** @param {string} id @returns {boolean} */
export function isTranscribing(id) {
  return live.has(id);
}

// captureStream builds one audio MediaStream carrying both call legs: the
// remote party (peer-connection receivers) and the operator's own voice
// (the audio sender track), mixed through an AudioContext the CALLER now
// owns — the context is close()d on stop (browsers cap live AudioContexts,
// so a leaked one kills capture after ~6 calls in Chrome). Null when the
// browser cannot mix (no AudioContext) or the call has no peer connection
// yet — the caller renders an honest unsupported notice. The destination
// node rides along so the silence probe can tap the same mix.
/**
 * @param {IslandCallEntry | null | undefined} entry
 * @returns {IslandCapture | null}
 */
export function captureStream(entry) {
  const sdh = entry && entry.session && entry.session.sessionDescriptionHandler;
  const pc = sdh && sdh.peerConnection;
  if (!pc) return null;
  const AudioCtx = window.AudioContext || window.webkitAudioContext;
  if (typeof AudioCtx !== "function") return null;
  const ctx = new AudioCtx();
  const dest = ctx.createMediaStreamDestination();

  const receivers = pc.getReceivers
    ? pc.getReceivers().filter((r) => r.track)
    : [];
  if (receivers.length) {
    const remote = new MediaStream();
    receivers.forEach((r) => remote.addTrack(r.track));
    ctx.createMediaStreamSource(remote).connect(dest);
  }
  const sender =
    pc.getSenders &&
    pc.getSenders().find((s) => s.track && s.track.kind === "audio");
  if (sender && sender.track) {
    const local = new MediaStream([sender.track]);
    ctx.createMediaStreamSource(local).connect(dest);
  }
  return { stream: dest.stream, ctx, dest };
}

// attachSilenceProbe taps the mix destination with an AnalyserNode and
// answers one question per segment boundary: "is the tail of the mix
// near-silent?" (the analyser's rolling window, fftSize 32768 ≈ 0.74 s —
// long enough that trailing pauses inside speech do not trip it). Null
// when the browser or the test stub lacks the pieces — the guard then
// FAILS OPEN (segments always POST); a mis-built probe must never
// silently kill capture.
/**
 * @param {AudioContext} ctx
 * @param {MediaStreamAudioDestinationNode} dest
 * @returns {(() => boolean) | null}
 */
function attachSilenceProbe(ctx, dest) {
  if (!ctx || typeof ctx.createAnalyser !== "function") return null;
  if (!dest || typeof dest.connect !== "function") return null;
  try {
    const analyser = ctx.createAnalyser();
    analyser.fftSize = 32768;
    dest.connect(analyser);
    const window = new Float32Array(analyser.fftSize);
    return () => {
      analyser.getFloatTimeDomainData(window);
      let sum = 0;
      for (let i = 0; i < window.length; i += 1) sum += window[i] * window[i];
      return Math.sqrt(sum / window.length) < SILENCE_RMS;
    };
  } catch {
    // Analyser setup only feeds the silence heuristic; capture itself
    // still starts when this returns null.
    return null;
  }
}

function recorderOptions() {
  if (typeof MediaRecorder === "undefined") return undefined;
  if (
    typeof MediaRecorder.isTypeSupported === "function" &&
    MediaRecorder.isTypeSupported("audio/webm;codecs=opus")
  ) {
    return { mimeType: "audio/webm;codecs=opus" };
  }
  return undefined;
}

// ensureRunning drives a capture AudioContext out of the suspended
// state. Browsers gate audio behind a user gesture; a mid-call page
// reload means capture starts WITHOUT one, and a suspended context
// records silence — the "transcription silently produces nothing" trap.
// One immediate resume covers the warm-gesture case; when the context
// came up suspended, one-shot first-gesture retries are armed too
// (Chrome holds resume() unresolved until the gesture fires). Never
// silent: the wait is logged where the operator reads it (#log stays
// English by contract).
/** @param {AudioContext} ctx */
function ensureRunning(ctx) {
  if (!ctx || typeof ctx.resume !== "function") return;
  const retry = () => {
    if (ctx.state === "closed") return;
    // A refused resume waits for the next pointerdown/keydown retry —
    // the "waiting for a click" warn above is the visible surface.
    ctx.resume().catch(() => {});
  };
  if (ctx.state === "suspended") {
    log(
      "live transcription waiting for a click or keypress to unlock audio",
      "warn",
    );
    window.addEventListener("pointerdown", retry, { once: true });
    window.addEventListener("keydown", retry, { once: true });
  }
  retry();
}

// startLiveTranscription begins capturing a live call and reports each
// segment's text through onText. Idempotent per call id. The language
// hint rides every segment (auto-detect misfires DE↔EN on 4 s phone
// chunks — the operator's UI language is the honest default). Returns
// false when the browser or the seam cannot support it (the caller
// shows the notice).
/**
 * @param {string} id
 * @param {IslandCallEntry} entry
 * @param {(text: string) => void} onText
 * @param {{language?: string}} [options]
 * @returns {boolean}
 */
export function startLiveTranscription(id, entry, onText, { language } = {}) {
  if (!transcribeEnabled || live.has(id)) return false;
  if (typeof MediaRecorder === "undefined") {
    log("live transcription unsupported: no MediaRecorder", "warn");
    return false;
  }
  const captured = captureStream(entry);
  if (!captured) {
    log("live transcription unsupported: no call audio to capture", "warn");
    return false;
  }
  const { stream, ctx, dest } = captured;
  ensureRunning(ctx);
  const isSilent = attachSilenceProbe(ctx, dest);

  /** @type {MediaRecorder} */
  let recorder;
  try {
    recorder = new MediaRecorder(stream, recorderOptions());
  } catch (err) {
    log(`live transcription recorder failed: ${errorMessage(err)}`, "warn");
    releaseCapture(ctx, stream);
    return false;
  }

  const controller = { recorder, ctx, stream };
  live.set(id, controller);
  recorder.addEventListener("dataavailable", async (event) => {
    if (!event.data || event.data.size === 0) return;
    // 429 cooldown (any surface can arm it) and transient-failure backoff
    // share this gate: skip the segment cycles until the window passes.
    // Capture keeps running; the budget recovers; no retry storm. One warn
    // covers the whole window (the cadence is ~15 segments/min).
    if (Date.now() < cooldownUntil) return;
    if (Date.now() < backoffUntil) return;
    // Silence guard: a near-silent tail window means the segment carries
    // no voice (mute, far-end hold without SIP hold, ringing) — posting
    // it would buy hallucinated text at full price. Skips are logged so
    // the operator can tell "nothing was said" from "capture broke".
    if (isSilent && isSilent()) {
      log("live transcription skipped a near-silent segment", "info");
      return;
    }
    try {
      const text = await transcribeBlob(event.data, {
        filename: "live.webm",
        language,
      });
      // A success is proof the seam recovered: clear the streak, the
      // backoff, and the once-per-streak toast flag together.
      liveFailures = 0;
      backoffUntil = 0;
      failureToasted = false;
      if (text) {
        // A stop can race an in-flight segment: once capture is torn
        // down, a late result must not repaint a call the island has
        // already settled. The text still reaches onText, because the
        // caller's save queue keeps it (attribution is by call id, and
        // the audio really was heard); only this page's live log drops
        // it, which is also what stops a dead call's entry from being
        // recreated forever in a long-lived tab.
        const stale = live.get(id) !== controller;
        onText(text);
        if (!stale) recordSegment(id, text);
      }
    } catch (err) {
      if (/\b429\b/.test(errorMessage(err))) {
        cooldownUntil = Date.now() + COOLDOWN_MS;
        log("transcription rate-limited; pausing segments briefly", "warn");
      } else {
        liveFailures += 1;
        backoffUntil =
          Date.now() +
          Math.min(BACKOFF_CAP_MS, BACKOFF_BASE_MS * 2 ** (liveFailures - 1));
        log(`live transcription segment failed: ${errorMessage(err)}`, "warn");
        if (liveFailures >= FAILURE_TOAST_AT && !failureToasted) {
          failureToasted = true;
          announce(t("transcribeUnavailable"), "warn");
        }
      }
    }
  });
  recorder.addEventListener("stop", () => {
    releaseCapture(ctx, stream);
    live.delete(id);
  });
  recorder.start(CHUNK_MS);
  return true;
}

// releaseCapture tears the mixing graph down: the AudioContext is the
// scarce resource (browsers cap concurrent instances), the mixed
// stream's tracks die with it. Defensive typeof guards keep the node
// test stubs honest.
/**
 * @param {AudioContext} ctx
 * @param {MediaStream} stream
 */
function releaseCapture(ctx, stream) {
  if (ctx && typeof ctx.close === "function") {
    // Teardown: a refused close dies with the page anyway.
    ctx.close().catch(() => {});
  }
  if (stream && stream.getTracks) {
    stream.getTracks().forEach((track) => {
      if (typeof track.stop === "function") track.stop();
    });
  }
}

// pauseLiveTranscription reflects the call's hold state onto the
// capture: a held call carries no audio in either direction (both
// track directions are disabled), so recording on would ship silence
// to the provider — the hallucination magnet. Pause/resume is a no-op
// for a call that is not being transcribed (safe on every hold flip).
/** @param {string} id @param {boolean} paused */
export function pauseLiveTranscription(id, paused) {
  const controller = live.get(id);
  if (!controller || !controller.recorder) return;
  try {
    if (paused && controller.recorder.state === "recording") {
      controller.recorder.pause();
    } else if (!paused && controller.recorder.state === "paused") {
      controller.recorder.resume();
    }
  } catch {
    // A mid-teardown toggle: the recorder is gone either way.
  }
}

// stopLiveTranscription ends capture for one call. A no-op when the call
// was never being transcribed (safe to call on every hangup).
/** @param {string} id */
export function stopLiveTranscription(id) {
  const controller = live.get(id);
  if (!controller) return;
  live.delete(id);
  releaseCapture(controller.ctx, controller.stream);
  try {
    controller.recorder.stop();
  } catch {
    // Already stopped / torn down with the call — the recorder raises a
    // DOMException on a double stop; capture is over either way.
  }
}

// transcribeNow is the on-demand helper shared by the island's voicemail
// panel: fetch a URL and render the text into target (or an error).
/**
 * @param {string} url
 * @param {HTMLElement} target
 * @param {{filename?: string}} [options]
 * @returns {Promise<void>}
 */
export async function transcribeNow(url, target, { filename } = {}) {
  if (!transcribeEnabled) return;
  target.textContent = t("transcribeBusy");
  try {
    const text = await transcribeUrl(url, { filename });
    target.textContent = text || t("transcribeEmpty");
  } catch (err) {
    target.textContent = t("transcribeFailed")(errorMessage(err));
    log(`transcription failed: ${errorMessage(err)}`, "error");
  }
}

// --- live caption log (multi-call train) ------------------------------------
//
// Every accepted segment lands in a per-call bounded log (the caption
// strip's scrollback — the durable whole stays server-side in History)
// and fans out to subscribers (calls.js renders the strip). The log is
// the copy surface too: reading the words from data, never from DOM
// text nodes.

// CAPTION_LINE_CAP bounds one call's scrollback, scalar-bounded like
// the server's store cap — a marathon call cannot grow the strip.
export const CAPTION_LINE_CAP = 30;

/** @type {Map<string, string[]>} */
const segmentLogs = new Map();
/** @type {Map<string, Set<(text: string) => void>>} */
const segmentSubs = new Map();

/** @param {string} id @param {string} text */
function recordSegment(id, text) {
  let entries = segmentLogs.get(id);
  if (!entries) {
    entries = [];
    segmentLogs.set(id, entries);
  }
  entries.push(text);
  if (entries.length > CAPTION_LINE_CAP) entries.shift();
  const subs = segmentSubs.get(id);
  if (subs) for (const fn of subs) fn(text);
}

// callSegments snapshots one call's logged segments (oldest first).
/** @param {string} id @returns {string[]} */
export function callSegments(id) {
  return [...(segmentLogs.get(id) ?? [])];
}

// clearCallSegments drops one call's log (teardown, explicit clear).
/** @param {string} id */
export function clearCallSegments(id) {
  segmentLogs.delete(id);
}

// subscribeCallSegments wires one listener to a call's future segments
// and returns the unsubscribe handle. The listener receives ONLY new
// segments (the strip replays callSegments() for its initial paint).
/**
 * @param {string} id
 * @param {(text: string) => void} fn
 * @returns {() => void}
 */
export function subscribeCallSegments(id, fn) {
  let subs = segmentSubs.get(id);
  if (!subs) {
    subs = new Set();
    segmentSubs.set(id, subs);
  }
  subs.add(fn);
  return () => {
    const current = segmentSubs.get(id);
    if (current) current.delete(fn);
  };
}
