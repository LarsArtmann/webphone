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

// watchRemoteAudio is the audio-gap leg of the measure-first latency
// design: it watches one call's remote stream through an AnalyserNode on
// the shared graph and fires ONCE when the first non-silent sample
// reaches the render path — the honest "the user hears audio" stamp,
// distinct from "signaling established" (ringback vs instant-audio
// deployments differ exactly there).
//
// Analysis-only: the source node is NOT connected to the destination, so
// the watcher cannot double the element's output. Feature-detected and
// never fatal — a browser without the WebAudio graph pieces (or a stub
// in the specs) returns a no-op disposer and the leg simply never
// stamps.
/**
 * @param {MediaStream} stream the call's remote stream
 * @param {() => void} onFirstAudio called once, from a poll tick
 * @returns {() => void} disposer — stop polling and detach the graph
 */
export function watchRemoteAudio(stream, onFirstAudio) {
  let stopped = false;
  let timer = null;
  /** @type {AnalyserNode | null} */
  let analyser = null;
  /** @type {MediaStreamAudioSourceNode | null} */
  let source = null;
  try {
    const audio = audioCtx();
    if (typeof audio.createMediaStreamSource !== "function") return () => {};
    source = audio.createMediaStreamSource(stream);
    analyser = audio.createAnalyser();
    // Small window, coarse resolution: we only need "is any sample away
    // from the digital-silence midpoint", not spectral fidelity.
    analyser.fftSize = 512;
    source.connect(analyser);
  } catch (err) {
    console.warn(
      `webphone: remote audio watcher unavailable: ${errorMessage(err)}`,
    );
    return () => {};
  }
  const buf = new Uint8Array(analyser.fftSize);
  // 50 ms poll: the leg's resolution is bounded by the poll, and the
  // noise floor (any byte deviating from the 128 midpoint by >4, i.e.
  // roughly -31 dBFS) keeps silence from stamping.
  timer = setInterval(() => {
    if (stopped) return;
    analyser.getByteTimeDomainData(buf);
    for (let i = 0; i < buf.length; i++) {
      const deviation = Math.abs(buf[i] - 128);
      if (deviation > 4) {
        stopped = true;
        if (timer) clearInterval(timer);
        timer = null;
        onFirstAudio();
        return;
      }
    }
  }, 50);
  return () => {
    stopped = true;
    if (timer) clearInterval(timer);
    timer = null;
    try {
      source.disconnect();
      analyser.disconnect();
    } catch {
      // A context that closed under us has already detached the graph.
    }
  };
}
