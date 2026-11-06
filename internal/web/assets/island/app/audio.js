import { errorMessage } from "./ui.js";

// Locally generated tones: US ringback for outgoing legs (2s on / 4s off
// at 440+480 Hz) and a distinct faster internal ring for incoming calls,
// so an incoming call is audible even while a ringback plays.
//
// ONE shared AudioContext: Chrome leaves a context created outside a
// user gesture `suspended` — an incoming ring (no gesture yet) would be
// SILENT. Sharing one context means a single resumeAudio() at the first
// user gesture (accept, reject, dial) brings every tone to life; the
// ringback and ring patterns stay audibly distinct by frequency and
// rhythm, not by context.

/** @type {AudioContext | null} */
let ctx = null;
/** @type {ReturnType<typeof setInterval> | null} */
let ringbackTimer = null;
/** @type {ReturnType<typeof setInterval> | null} */
let ringToneTimer = null;

// The operator's chosen output, when the audio-out picker offers one. A
// RING must land on the device the operator is wearing: on a
// multi-output desk the USB headset is the phone, and a ring on the
// motherboard speaker is a missed call. WebAudio's AudioContext.setSinkId
// (Chrome 110+) is the half that routing needs; remembered here so a
// picker choice made BEFORE the first tone still applies when the
// context is finally created by a gesture.
/** @type {string} */
let preferredSink = "";

// applySink routes the context's OUTPUT to a device. Feature-detected
// structurally (the DOM lib and older browsers disagree about this
// method's existence), and never fatal: where the sink cannot be set the
// ring stays a room alarm, which is exactly the pre-picker product.
/** @param {AudioContext} audio @param {string} deviceId @returns {void} */
function applySink(audio, deviceId) {
  const sinkable =
    /** @type {AudioContext & { setSinkId?: (id: string) => Promise<void> | void }} */ (
      audio
    );
  if (typeof sinkable.setSinkId !== "function") return;
  try {
    const verdict = sinkable.setSinkId(deviceId);
    if (verdict && typeof verdict.then === "function") {
      verdict.catch((err) => {
        console.warn(
          `webphone: ring tone output switch failed: ${errorMessage(err)}`,
        );
      });
    }
  } catch (err) {
    console.warn(
      `webphone: ring tone output switch failed: ${errorMessage(err)}`,
    );
  }
}

// setToneSink is the audio-out picker's other half: the same chosen device
// carries the remote audio (a media element) AND every locally generated
// tone. Called with "" to return to the browser default.
/** @param {string} deviceId */
export function setToneSink(deviceId) {
  preferredSink = deviceId || "";
  if (ctx) applySink(ctx, preferredSink);
}

/** @returns {AudioContext} */
function audioCtx() {
  if (!ctx) {
    ctx = new AudioContext();
    if (preferredSink) applySink(ctx, preferredSink);
  }
  return ctx;
}

// Bring the shared context from suspended to running. Call from gesture
// handlers (accept/reject/dial in calls.js); a no-op promise when the
// context already runs.
/** @returns {Promise<void>} */
export function resumeAudio() {
  const audio = audioCtx();
  if (audio.state === "suspended" && audio.resume)
    return audio.resume().catch((err) => {
      // Autoplay policy can refuse a resume without a fresh gesture;
      // the next user gesture retries. Operator-visible, not a toast.
      console.warn(`webphone: audio resume failed: ${errorMessage(err)}`);
    });
  return Promise.resolve();
}

export function ringbackStart() {
  if (ringbackTimer) return;
  const audio = audioCtx();
  resumeAudio().catch(() => {});
  const on = () => {
    const now = audio.currentTime;
    [440, 480].forEach((freq) => {
      const osc = audio.createOscillator();
      const gain = audio.createGain();
      osc.frequency.value = freq;
      gain.gain.value = 0.06;
      osc.connect(gain).connect(audio.destination);
      osc.start(now);
      osc.stop(now + 2);
    });
  };
  on();
  ringbackTimer = setInterval(on, 6000);
}

export function ringbackStop() {
  if (ringbackTimer) clearInterval(ringbackTimer);
  ringbackTimer = null;
}

export function ringToneStart() {
  if (ringToneTimer) return;
  const audio = audioCtx();
  // No gesture exists while a call rings — this resume usually no-ops,
  // but after the FIRST user gesture the context stays running, so a
  // second incoming call rings audibly without any further unlock.
  resumeAudio().catch(() => {});
  const burst = () => {
    const now = audio.currentTime;
    [480, 960].forEach((freq) => {
      const osc = audio.createOscillator();
      const gain = audio.createGain();
      osc.frequency.value = freq;
      gain.gain.value = 0.05;
      osc.connect(gain).connect(audio.destination);
      osc.start(now);
      osc.stop(now + 0.4);
    });
  };
  burst();
  ringToneTimer = setInterval(burst, 2000);
}

export function ringToneStop() {
  if (ringToneTimer) clearInterval(ringToneTimer);
  ringToneTimer = null;
}
