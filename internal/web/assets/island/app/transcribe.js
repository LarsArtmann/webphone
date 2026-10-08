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
import { log } from "./ui.js";

export const transcribeEnabled = asrEnabled;

// saveTranscriptSegment persists one live-call segment to the server's
// owner-scoped transcript store (the History tab's durable source).
// Fire-and-forget: a failed save costs the LATER re-read, not the live
// card — the caller logs/toasts, never blocks the capture loop.
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

// transcribeBlob posts one encoded audio blob and returns the trimmed
// text ("" when the provider heard nothing).
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
    throw new Error(`HTTP ${res.status}`);
  }
  const data = await res.json();
  return String(data.text || "").trim();
}

// transcribeUrl fetches a same-origin audio URL (voicemail playback,
// MMS attachment) into a blob and transcribes it.
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
const live = new Map();

export function isTranscribing(id) {
  return live.has(id);
}

// captureStream builds one audio MediaStream carrying both call legs: the
// remote party (peer-connection receivers) and the operator's own voice
// (the audio sender track), mixed through an AudioContext the CALLER now
// owns — the context is close()d on stop (browsers cap live AudioContexts,
// so a leaked one kills capture after ~6 calls in Chrome). Null when the
// browser cannot mix (no AudioContext) or the call has no peer connection
// yet — the caller renders an honest unsupported notice.
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
  return { stream: dest.stream, ctx };
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
function ensureRunning(ctx) {
  if (!ctx || typeof ctx.resume !== "function") return;
  const retry = () => {
    if (ctx.state === "closed") return;
    ctx.resume().catch(() => {});
  };
  if (ctx.state === "suspended") {
    log("live transcription waiting for a click or keypress to unlock audio", "warn");
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
  const { stream, ctx } = captured;
  ensureRunning(ctx);

  let recorder;
  try {
    recorder = new MediaRecorder(stream, recorderOptions());
  } catch (err) {
    log(`live transcription recorder failed: ${err.message}`, "warn");
    releaseCapture(ctx, stream);
    return false;
  }

  const controller = { recorder, ctx, stream };
  live.set(id, controller);
  recorder.addEventListener("dataavailable", async (event) => {
    if (!event.data || event.data.size === 0) return;
    try {
      const text = await transcribeBlob(event.data, {
        filename: "live.webm",
        language,
      });
      if (text) onText(text);
    } catch (err) {
      log(`live transcription segment failed: ${err.message}`, "warn");
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
function releaseCapture(ctx, stream) {
  if (ctx && typeof ctx.close === "function") {
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
export async function transcribeNow(url, target, { filename } = {}) {
  if (!transcribeEnabled) return;
  target.textContent = t("transcribeBusy");
  try {
    const text = await transcribeUrl(url, { filename });
    target.textContent = text || t("transcribeEmpty");
  } catch (err) {
    target.textContent = t("transcribeFailed")(err.message);
    log(`transcription failed: ${err.message}`, "error");
  }
}
