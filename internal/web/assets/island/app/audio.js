// Locally generated tones: US ringback for outgoing legs (2s on / 4s off
// at 440+480 Hz) and a distinct faster internal ring for incoming calls,
// so an incoming call is audible even while a ringback plays.

let ringbackCtx = null;
let ringbackTimer = null;

export function ringbackStart() {
  if (ringbackTimer) return;
  ringbackCtx = ringbackCtx || new AudioContext();
  const on = () => {
    const now = ringbackCtx.currentTime;
    [440, 480].forEach((freq) => {
      const osc = ringbackCtx.createOscillator();
      const gain = ringbackCtx.createGain();
      osc.frequency.value = freq;
      gain.gain.value = 0.06;
      osc.connect(gain).connect(ringbackCtx.destination);
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

let ringToneCtx = null;
let ringToneTimer = null;

export function ringToneStart() {
  if (ringToneTimer) return;
  ringToneCtx = ringToneCtx || new AudioContext();
  const burst = () => {
    const now = ringToneCtx.currentTime;
    [480, 960].forEach((freq) => {
      const osc = ringToneCtx.createOscillator();
      const gain = ringToneCtx.createGain();
      osc.frequency.value = freq;
      gain.gain.value = 0.05;
      osc.connect(gain).connect(ringToneCtx.destination);
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
