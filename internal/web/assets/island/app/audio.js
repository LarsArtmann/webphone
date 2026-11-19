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
// rhythm, not by context. The gesture unlock below additionally arms a
// first-interaction resume so a never-gestured tab (cookie session
// restore) rings as soon as the user touches ANYTHING.

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

// The gesture unlock: Chrome leaves a context created outside a user
// gesture suspended FOREVER — and the first ring usually IS that
// moment (a cookie session restore boots the island without a click,
// so the 2026-11-19 missed call rang silently). The per-invite resume
// this replaces could only fail (a ring carries no gesture) and
// printed one autoplay warning per forked invitation. ONE capture-phase
// pointerdown/keydown pair owns the suspended→running transition: the
// first interaction anywhere on the page resumes the shared context
// and removes itself; a refused resume (no activation yet) stays
// armed for the next gesture.
/** @type {(() => void) | null} */
let disarmUnlock = null;

/** @param {AudioContext} audio */
function armGestureUnlock(audio) {
  if (disarmUnlock || audio.state !== "suspended" || !audio.resume) return;
  const disarm = () => {
    if (!disarmUnlock) return;
    const remove = disarmUnlock;
    disarmUnlock = null;
    remove();
  };
  const resume = () => {
    if (audio.state !== "suspended") {
      disarm();
      return;
    }
    audio.resume().then(disarm, () => {
      // That event carried no user activation; stay armed and let the
      // next interaction retry.
    });
  };
  window.addEventListener("pointerdown", resume, {
    capture: true,
    passive: true,
  });
  window.addEventListener("keydown", resume, {
    capture: true,
    passive: true,
  });
  disarmUnlock = () => {
    window.removeEventListener("pointerdown", resume, { capture: true });
    window.removeEventListener("keydown", resume, { capture: true });
  };
}

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
    armGestureUnlock(ctx);
  }
  return ctx;
}

/** @type {boolean} */
let resumeWarned = false;

// Bring the shared context from suspended to running. Call from gesture
// handlers (accept/reject/dial in calls.js); a no-op promise when the
// context already runs.
/** @returns {Promise<void>} */
export function resumeAudio() {
  const audio = audioCtx();
  if (audio.state === "suspended" && audio.resume)
    return audio.resume().catch((err) => {
      // Autoplay policy can refuse a resume without a fresh gesture;
      // the next user gesture retries (the gesture unlock above arms
      // that retry for the ring path). Operator-visible, not a toast —
      // and at most once per page, so a fork-storm cannot spam it.
      if (!resumeWarned) {
        resumeWarned = true;
        console.warn(`webphone: audio resume failed: ${errorMessage(err)}`);
      }
    });
  return Promise.resolve();
}

export function ringbackStart() {
  if (ringbackTimer) return;
  const audio = audioCtx();
  resumeAudio().catch(() => {
    // Deliberate discard: resumeAudio observes its own failure (the
    // console.warn above) and never rejects — this catch only keeps
    // the fire-and-forget contract explicit for the ring tone path.
  });
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
  // No gesture exists while a call rings: a fresh context stays
  // suspended (the gesture unlock above owns the first-interaction
  // transition — calling resume here could only fail and warn), and
  // the burst schedules into the suspended graph so the queued ring is
  // heard the moment ANY interaction unlocks the context. After the
  // first user gesture the context runs, so later incoming calls ring
  // audibly with no further unlock.
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
  /** @type {ReturnType<typeof setInterval> | null} */
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
