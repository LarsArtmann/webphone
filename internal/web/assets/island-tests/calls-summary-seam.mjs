// The summary-family call seam: each node:test file is a separate
// process that must install its globals BEFORE the island import, so
// the ended-card specs (calls-summary-off, calls-summary-noasr) share
// this one seam home instead of carrying byte-identical scaffolds.
// Family-local by design (the dedup registry's island-test ruling keeps
// the other spec families independently variable — nothing here is a
// helpers.mjs hoist). `config` becomes window.PBX_CONFIG when present;
// omit it for the strictest seam-off boot.
import { installBrowserGlobals } from "./helpers.mjs";

export async function installSummarySeam({ id, config } = {}) {
  installBrowserGlobals();
  if (config) {
    globalThis.window.PBX_CONFIG = config;
  }
  globalThis.SIP = {
    SessionState: {
      Establishing: "Establishing",
      Established: "Established",
      Terminating: "Terminating",
      Terminated: "Terminated",
    },
    Inviter: class {
      constructor() {
        this.id = id;
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

  const { placeCall } = await import("../island/app/calls.js");
  const { sessions, state } = await import("../island/app/state.js");
  const { els } = await import("../island/app/ui.js");

  els.remoteAudio.play = () => Promise.resolve();
  return { placeCall, sessions, state, els };
}

export const flush = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

export async function settleEnded(entry) {
  entry.session.state = globalThis.SIP.SessionState.Established;
  entry.session.stateChange.listeners.forEach((fn) => fn("Established"));
  entry.session.state = globalThis.SIP.SessionState.Terminated;
  entry.session.stateChange.listeners.forEach((fn) => fn("Terminated"));
  await flush();
}
