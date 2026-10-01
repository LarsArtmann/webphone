// ice.js panel rendering under node:test: the live stats lines AND the
// setup-timing line composed from pcsetup's event marks plus the
// first-media timestamp the poll itself records.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();

const { sessions, state } = await import("../island/app/state.js");
const { startIcePanel, stopIcePanel } = await import("../island/app/ice.js");
const { instrumentSessionDescriptionHandler } = await import(
  "../island/app/pcsetup.js"
);

const flush = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

const makePC = (statsEntries) => {
  const listeners = {};
  return {
    iceConnectionState: "connected",
    iceGatheringState: "complete",
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    fire(type) {
      for (const fn of listeners[type] ?? []) fn();
    },
    async getStats() {
      return new Map(statsEntries);
    },
  };
};

const statsWith = (bytesReceived) => [
  ["t1", { type: "transport", selectedCandidatePairId: "p1" }],
  [
    "p1",
    {
      type: "candidate-pair",
      id: "p1",
      selected: true,
      nominated: true,
      state: "succeeded",
      currentRoundTripTime: 0.021,
      localCandidateId: "l1",
      remoteCandidateId: "r1",
    },
  ],
  ["l1", { type: "local-candidate", candidateType: "srflx" }],
  ["r1", { type: "remote-candidate", candidateType: "host" }],
  [
    "in",
    {
      type: "inbound-rtp",
      kind: "audio",
      packetsLost: 0,
      jitter: 0.002,
      bytesReceived,
    },
  ],
  ["c", { type: "codec", mimeType: "audio/opus" }],
];

const panelText = () =>
  doc
    .getElementById("ice-panel")
    .children.map((line) => line.textContent)
    .join("\n");

test("the panel renders live stats plus the setup story once media flows", async () => {
  const pc = makePC(statsWith(4096));
  instrumentSessionDescriptionHandler({ peerConnection: pc });
  pc.iceGatheringState = "gathering";
  pc.fire("icegatheringstatechange");
  pc.iceGatheringState = "complete";
  pc.fire("icegatheringstatechange");
  pc.iceConnectionState = "connected";
  pc.fire("iceconnectionstatechange");

  sessions.set("c1", {
    session: { sessionDescriptionHandler: { peerConnection: pc } },
    startedAt: Date.now(),
  });
  state.focusedId = "c1";

  startIcePanel();
  await flush();
  const text = panelText();
  stopIcePanel();

  assert.match(text, /ice: connected\s+path: srflx → host/);
  assert.match(text, /codec: opus\s+rtt: 21 ms/);
  assert.match(text, /setup: gather \d+ ms/);
  assert.match(text, /ice \+\d+\.\d s/);
  assert.match(text, /first media \+\d+\.\d s/);
  sessions.delete("c1");
});

test("an uninstrumented call shows the live lines but no setup line", async () => {
  const pc = makePC(statsWith(0));
  sessions.set("c2", {
    session: { sessionDescriptionHandler: { peerConnection: pc } },
    startedAt: Date.now(),
  });
  state.focusedId = "c2";

  startIcePanel();
  await flush();
  const text = panelText();
  stopIcePanel();

  assert.match(text, /path: srflx → host/, "the pair still resolves");
  assert.match(text, /received: 0 B/, "no media bytes yet");
  assert.doesNotMatch(text, /setup:/, "no event marks, no setup story");
  assert.doesNotMatch(text, /first media/, "no bytes, no first-media mark");
  sessions.delete("c2");
});
