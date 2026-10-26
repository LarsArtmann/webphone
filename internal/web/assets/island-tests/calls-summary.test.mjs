// The ended card's post-call row (M12): a normally terminated call
// leaves a settled card with Close — plus "Summarize → draft SMS" when
// the agent seam is on AND the call actually carried media. The agent
// gate is forced ON before any island import (config.js reads
// window.PBX_CONFIG at import time).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = { agent: true, sipDomain: "sip.test" };

globalThis.SIP = {
  SessionState: {
    Establishing: "Establishing",
    Established: "Established",
    Terminating: "Terminating",
    Terminated: "Terminated",
  },
  Inviter: class {
    constructor() {
      this.id = "call-s1";
      this.state = globalThis.SIP.SessionState.Establishing;
      this.stateChange = {
        listeners: [],
        addListener(fn) {
          this.listeners.push(fn);
        },
      };
    }
    invite() {
      return Promise.resolve();
    }
  },
};
globalThis.SIP.UserAgent = { makeURI: (raw) => ({ toString: () => raw }) };

const node = () => ({
  connect() {
    return node();
  },
  start() {},
  stop() {},
});
globalThis.AudioContext = class {
  constructor() {
    this.currentTime = 0;
    this.destination = {};
  }
  createOscillator() {
    return { frequency: {}, connect: node, start() {}, stop() {} };
  }
  createGain() {
    return { gain: {}, connect: node };
  }
  createMediaStreamDestination() {
    return { stream: { id: "mixed", getTracks: () => [] } };
  }
  createMediaStreamSource() {
    return { connect() {} };
  }
  close() {
    return Promise.resolve();
  }
};
globalThis.window.AudioContext = globalThis.AudioContext;
globalThis.MediaStream = class {
  constructor() {}
  addTrack() {}
};

let fetchCalls = [];
let fetchImpl = async () => ({ ok: true, status: 204 });
globalThis.fetch = async (url, options = {}) => {
  fetchCalls.push({ url, options });
  return fetchImpl(url, options);
};

const { placeCall } = await import("../island/app/calls.js");
const { sessions, state } = await import("../island/app/state.js");
const { els } = await import("../island/app/ui.js");

els.remoteAudio.play = () => Promise.resolve();

const flush = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

// One established call, ended normally: the settled card stays with the
// post-call row (summarize + close), and the session table is clean.
async function endedEstablishedCall(id) {
  state.userAgent = {};
  const inviter = new globalThis.SIP.Inviter();
  inviter.id = id;
  // placeCall builds its own Inviter; drive it through the real path.
  await placeCall("+4930123456");
  const entry = sessions.get(id);
  assert.ok(entry, "the call is live");
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
  return entry;
}

test("a terminated established call leaves a settled ended card with summarize and close", async () => {
  const entry = await endedEstablishedCall("call-s1");
  assert.equal(sessions.has("call-s1"), false, "the session table is clean");
  const card = entry.dom;
  assert.ok(card.parent, "the ended card stays visible");
  assert.ok(card.classList.contains("wp-ended-card"), "the ended hook is set");
  assert.equal(card.dataset.callId, "call-s1", "the summarize ask can name the call");
  assert.match(
    card.querySelector(".call-state-text").textContent,
    /Ended/,
    "the state line is the settled ended label",
  );
  const summarize = card.querySelector(".summarize-btn");
  assert.ok(summarize, "the summarize affordance exists (seam on, media flowed)");
  assert.equal(summarize.textContent, "Summarize → draft SMS");
  assert.ok(card.querySelector(".close-ended-btn"), "Close is always there");
});

test("a call that never carried media offers Close only", async () => {
  state.userAgent = {};
  await placeCall("+4930123457");
  const entry = sessions.get("call-s1");
  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
  assert.equal(sessions.has("call-s1"), false);
  assert.ok(entry.dom.parent, "the ended card stays visible");
  assert.equal(
    entry.dom.querySelector(".summarize-btn"),
    null,
    "no transcript exists for a call without media — no summarize affordance",
  );
  assert.ok(entry.dom.querySelector(".close-ended-btn"));
});

test("summarize: busy state, success toast, nav nudge, settled button", async () => {
  const entry = await endedEstablishedCall("call-s1");
  const btn = entry.dom.querySelector(".summarize-btn");
  fetchCalls = [];
  let draftsNudges = 0;
  doc.addEventListener("wp:drafts-changed", () => {
    draftsNudges += 1;
  });
  fetchImpl = async () => ({
    ok: true,
    status: 200,
    json: async () => ({ id: "d-1", contact: "+4930123456", text: "Danke!" }),
  });
  btn.click();
  assert.equal(btn.disabled, true, "the row is busy immediately");
  assert.equal(btn.textContent, "Summarizing…");
  await flush();
  const ask = fetchCalls.find((c) =>
    String(c.url).includes("/summary-draft"),
  );
  assert.ok(ask, "the island asked the summary endpoint");
  assert.equal(ask.options.method, "POST");
  assert.ok(
    String(ask.url).includes(encodeURIComponent("call-s1")),
    "the call id rides the path encoded",
  );
  assert.equal(
    btn.textContent,
    "Draft in Messages",
    "the settled state names where the draft went",
  );
  assert.equal(btn.disabled, true, "a drafted call does not re-ask");
  assert.ok(draftsNudges >= 1, "the nav nudge fired for Messages");
  assert.ok(
    [...doc.getElementById("toasts").children].some((el) =>
      /draft ready in Messages/i.test(el.textContent),
    ),
    "the success toast names the destination",
  );
});

test("summarize: a server duplicate still settles honestly and nudges", async () => {
  const entry = await endedEstablishedCall("call-s1");
  const btn = entry.dom.querySelector(".summarize-btn");
  fetchImpl = async () => ({
    ok: true,
    status: 200,
    json: async () => ({ duplicate: true }),
  });
  btn.click();
  await flush();
  assert.equal(btn.textContent, "Draft in Messages");
  assert.ok(
    [...doc.getElementById("toasts").children].some((el) =>
      /Already drafted/i.test(el.textContent),
    ),
    "the duplicate verdict is its own toast",
  );
});

test("summarize: an outage returns the retryable settled state", async () => {
  const entry = await endedEstablishedCall("call-s1");
  const btn = entry.dom.querySelector(".summarize-btn");
  fetchImpl = async () => ({ ok: false, status: 502 });
  btn.click();
  await flush();
  assert.equal(btn.disabled, false, "a failed ask stays retryable");
  assert.equal(
    btn.textContent,
    "Summarize → draft SMS",
    "the SETTLED state is the original action, never the hoped-for one",
  );
  assert.ok(
    [...doc.getElementById("toasts").children].some((el) =>
      /agent is unreachable/i.test(el.textContent),
    ),
    "the outage toast explains the failure",
  );
});

test("summarize: a missing transcript warns and stays retryable", async () => {
  const entry = await endedEstablishedCall("call-s1");
  const btn = entry.dom.querySelector(".summarize-btn");
  fetchImpl = async () => ({ ok: false, status: 404 });
  btn.click();
  await flush();
  assert.equal(btn.disabled, false);
  assert.ok(
    [...doc.getElementById("toasts").children].some((el) =>
      /No transcript/i.test(el.textContent),
    ),
    "the 404 lands as its own honest toast",
  );
});

test("close removes the ended card", async () => {
  const entry = await endedEstablishedCall("call-s1");
  const card = entry.dom;
  card.querySelector(".close-ended-btn").click();
  assert.equal(card.parent, null, "the card is gone on Close");
});
