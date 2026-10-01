// pcsetup.js under node:test — the per-call ICE setup timing marks.
// The listeners must land at session-description-handler construction
// (before the first offer/answer) because gathering completes in well
// under a second; these tests drive the exact state transitions the
// browser fires and pin what the ice panel renders from them.
import test from "node:test";
import assert from "node:assert/strict";

const load = (tag) => import(`../island/app/pcsetup.js?case=${tag}`);

const makePC = () => {
  const listeners = {};
  const pc = {
    iceGatheringState: "new",
    iceConnectionState: "new",
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    fire(type) {
      for (const fn of listeners[type] ?? []) fn();
    },
  };
  return pc;
};

test("gathering start and complete are captured, duration reported", async () => {
  const { instrumentSessionDescriptionHandler, setupSummary } = await load("gather-duration");
  const pc = makePC();
  instrumentSessionDescriptionHandler({ peerConnection: pc });

  pc.iceGatheringState = "gathering";
  pc.fire("icegatheringstatechange");
  pc.iceGatheringState = "complete";
  pc.fire("icegatheringstatechange");

  const summary = setupSummary(pc);
  assert.notEqual(summary.gatherMs, null, "the observed span is reported");
  assert.ok(summary.gatherMs >= 0);
  assert.equal(summary.gatherCapped, false);
  assert.equal(summary.iceConnectedAt, null);
});

test("ICE connected is captured once, at the first connected/completed", async (t) => {
  const timers = t.mock.timers;
  timers.enable({ apis: ["Date"] });
  const { instrumentSessionDescriptionHandler, setupSummary } = await load("ice-connected");
  const pc = makePC();
  instrumentSessionDescriptionHandler({ peerConnection: pc });

  pc.iceGatheringState = "gathering";
  pc.fire("icegatheringstatechange");
  timers.tick(120);
  pc.iceConnectionState = "connected";
  pc.fire("iceconnectionstatechange");
  const firstMark = setupSummary(pc).iceConnectedAt;
  timers.tick(50);
  pc.iceConnectionState = "completed";
  pc.fire("iceconnectionstatechange");

  const summary = setupSummary(pc);
  assert.equal(summary.iceConnectedAt, firstMark, "later transitions do not move the mark");
});

test("a gather still running past the cap reads as capped", async (t) => {
  const timers = t.mock.timers;
  timers.enable({ apis: ["Date"] });
  const { instrumentSessionDescriptionHandler, setupSummary } = await load("capped");
  const pc = makePC();
  instrumentSessionDescriptionHandler({ peerConnection: pc });

  pc.iceGatheringState = "gathering";
  pc.fire("icegatheringstatechange");
  assert.equal(setupSummary(pc).gatherCapped, false, "within the cap window");

  timers.tick(1300);
  const stale = setupSummary(pc);
  assert.equal(stale.gatherMs, null, "no end mark ever arrived");
  assert.equal(stale.gatherCapped, true, "the stale gathering is flagged as capped");

  pc.iceGatheringState = "complete";
  pc.fire("icegatheringstatechange");
  const settled = setupSummary(pc);
  assert.equal(settled.gatherCapped, false, "a late complete un-flags the cap");
  assert.notEqual(settled.gatherMs, null);
});

test("uninstrumented and absent PCs summarize to null", async () => {
  const { setupSummary } = await load("null-summary");
  assert.equal(setupSummary(null), null);
  assert.equal(setupSummary(makePC()), null, "a PC nobody instrumented has no marks");
});

test("instrumenting twice attaches one set of listeners", async () => {
  const { instrumentSessionDescriptionHandler, setupSummary } = await load("idempotent");
  const pc = makePC();
  instrumentSessionDescriptionHandler({ peerConnection: pc });
  instrumentSessionDescriptionHandler({ peerConnection: pc });

  pc.iceGatheringState = "gathering";
  pc.fire("icegatheringstatechange");
  pc.iceGatheringState = "complete";
  pc.fire("icegatheringstatechange");

  const summary = setupSummary(pc);
  assert.notEqual(summary.gatherMs, null, "marks work; duplicate attach is a no-op");
});
