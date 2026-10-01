let warmStream = null;
let warmToken = 0;

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

export function warmMic() {
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
}

export function takeWarmMic() {
  const stream = warmStream;
  warmStream = null;
  warmToken = 0;
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
