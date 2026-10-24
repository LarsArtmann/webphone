let warmStream = null;
let warmToken = 0;
let warmTimer = null;

function trackEnded(stream) {
  return () => {
    if (warmStream === stream) warmStream = null;
  };
}

function adopt(stream) {
  warmStream = stream;
  for (const track of stream.getTracks()) {
    track.addEventListener("ended", trackEnded(stream));
  }
}

function discard(stream) {
  for (const track of stream.getTracks()) track.stop();
}

function clearWarmTimer() {
  if (warmTimer) clearTimeout(warmTimer);
  warmTimer = null;
}

// scheduleWarmExpiry bounds a DECAYING warm (the dial-focus kind):
// dial intent goes stale, and an unbounded warm would keep the mic
// indicator lit long after the user abandoned dialing. The identity
// guard means an expired timer can never kill a LATER warm (e.g. the
// incoming-call one, which carries no TTL — a ring may outlast any
// sensible dial window).
function scheduleWarmExpiry(stream, ttlMs) {
  clearWarmTimer();
  warmTimer = setTimeout(() => {
    warmTimer = null;
    if (warmStream === stream) releaseWarmMic();
  }, ttlMs);
}

// warmMic starts the pre-acquisition. ttlMs > 0 arms an expiry: if the
// stream is neither consumed (takeWarmMic) nor released by then, it is
// returned to the device — the caller expresses how long its intent
// stays fresh (dial focus: seconds-to-a-minute; incoming ring: no TTL,
// the caller-gave-up/reject paths own the release).
export function warmMic(ttlMs = 0) {
  if (warmStream || warmToken) return;
  const mediaDevices = navigator.mediaDevices;
  if (!mediaDevices || !mediaDevices.getUserMedia) return;
  const token = ++warmToken;
  mediaDevices
    .getUserMedia({ audio: true, video: false })
    .then((stream) => {
      if (token !== warmToken) {
        discard(stream);
        return;
      }
      adopt(stream);
      if (ttlMs > 0) scheduleWarmExpiry(stream, ttlMs);
    })
    .catch(() => {})
    .finally(() => {
      if (token === warmToken) warmToken = 0;
    });
}

export function releaseWarmMic() {
  if (warmStream) discard(warmStream);
  warmStream = null;
  warmToken = 0;
  clearWarmTimer();
}

export function takeWarmMic() {
  const stream = warmStream;
  warmStream = null;
  warmToken = 0;
  clearWarmTimer();
  return stream;
}

export function micMediaStreamFactory(constraints) {
  if (constraints && !constraints.audio && !constraints.video) {
    return Promise.resolve(new MediaStream());
  }
  const warm =
    constraints && constraints.audio && !constraints.video
      ? takeWarmMic()
      : null;
  if (warm) return Promise.resolve(warm);
  const mediaDevices = navigator.mediaDevices;
  if (!mediaDevices || !mediaDevices.getUserMedia) {
    return Promise.reject(
      new Error("Media devices not available in insecure contexts."),
    );
  }
  return mediaDevices.getUserMedia(constraints);
}
