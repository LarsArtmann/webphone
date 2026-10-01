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

let ctx = null;
let ringbackTimer = null;
let ringToneTimer = null;

function audioCtx() {
  ctx = ctx || new AudioContext();
  return ctx;
}

// Bring the shared context from suspended to running. Call from gesture
// handlers (accept/reject/dial in calls.js); a no-op promise when the
// context already runs.
export function resumeAudio() {
  const audio = audioCtx();
  if (audio.state === "suspended" && audio.resume) return audio.resume();
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
