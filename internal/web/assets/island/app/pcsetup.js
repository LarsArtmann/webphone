// Per-call peer-connection SETUP timings, captured from events at
// session-description-handler construction (connection.js wraps the
// factory): ICE gathering duration and the moment the connection went
// connected. The listeners must land before the first offer/answer —
// gathering completes in well under a second under the 1000 ms cap, so
// any polling approach (the ice.js 2 s tick) would always miss it.
//
// Leaf module by design: connection.js (the factory wrap) and ice.js
// (the rendering) both reach it without importing each other — the
// calls/ice/connection independence the arch test enforces.

const timings = new WeakMap();

function marksFor(pc) {
  let marks = timings.get(pc);
  if (!marks) {
    marks = {
      gatheringStartedAt: null,
      gatheringEndedAt: null,
      iceConnectedAt: null,
    };
    timings.set(pc, marks);
  }
  return marks;
}

// instrumentSessionDescriptionHandler attaches the setup-timing
// listeners to the handler's peer connection. Call it the moment the
// factory returns the handler — the PC exists from construction.
export function instrumentSessionDescriptionHandler(sdh) {
  const pc = sdh && sdh.peerConnection;
  if (!pc || !pc.addEventListener) return;
  if (timings.has(pc)) return;
  const marks = marksFor(pc);
  // Constructed mid-gathering (possible for re-hydrated handlers):
  // record what is already true instead of missing it.
  if (pc.iceGatheringState === "gathering") {
    marks.gatheringStartedAt ??= Date.now();
  }
  pc.addEventListener("icegatheringstatechange", () => {
    const state = pc.iceGatheringState;
    if (state === "gathering") marks.gatheringStartedAt ??= Date.now();
    if (state === "complete") marks.gatheringEndedAt ??= Date.now();
  });
  pc.addEventListener("iceconnectionstatechange", () => {
    const state = pc.iceConnectionState;
    if (
      (state === "connected" || state === "completed") &&
      !marks.iceConnectedAt
    ) {
      marks.iceConnectedAt = Date.now();
    }
  });
}

// GATHER_CAP_MS mirrors the connection.js iceGatheringTimeout: sip.js
// stops WAITING for gathering at the cap and proceeds with whatever
// candidates it has, leaving the PC in state "gathering" forever — that
// stale state is exactly how a capped gather is detected.
const GATHER_CAP_MS = 1000;

// setupSummary reports the captured setup timings for a peer
// connection (null when the PC was never instrumented):
//   - gatherMs: gathering duration when start AND end were seen
//   - gatherCapped: still gathering past the cap (sip.js gave up
//     waiting; the call rides partial candidates)
//   - gatheringStartedAt: the baseline other phases are measured from
//   - iceConnectedAt: when ICE first reported connected/completed
export function setupSummary(pc) {
  const marks = pc ? timings.get(pc) : null;
  if (!marks) return null;
  const gatherMs =
    marks.gatheringStartedAt && marks.gatheringEndedAt
      ? marks.gatheringEndedAt - marks.gatheringStartedAt
      : null;
  const gatherCapped =
    marks.gatheringStartedAt !== null &&
    marks.gatheringEndedAt === null &&
    Date.now() - marks.gatheringStartedAt > GATHER_CAP_MS + 200;
  return {
    gatherMs,
    gatherCapped,
    gatheringStartedAt: marks.gatheringStartedAt,
    iceConnectedAt: marks.iceConnectedAt,
  };
}
