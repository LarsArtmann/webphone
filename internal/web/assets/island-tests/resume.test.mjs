// main.js boot resume under node:test — the two failure classes of a
// resumed session's SIP connect. A REGISTER the server refuses means
// the stored credentials are stale: the session is dropped and the
// login form returns. A transport failure (PBX/WebSocket down at load)
// is NOT a stale session: the cookie is KEPT, the phone shows its
// honest offline state, and the reconnect backoff owns recovery — the
// tabs stay usable. Found live by the T23 visual harness: the
// bare-boot island deleted its injected session within a second of
// every page load.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
await import("../island/app/i18n.js");
await import("../island/app/ui.js");

// main.js registers browser-network listeners on window (offline/
// online + the HX toast channel); helpers ship a bare window object.
const winListeners = {};
globalThis.window.addEventListener = (type, fn) => {
  (winListeners[type] ??= []).push(fn);
};

let scenario = "transport";
const calls = { fetches: [] };

class StubRegisterer {
  constructor() {
    this.stateChange = {
      listeners: [],
      addListener(fn) {
        this.listeners.push(fn);
      },
      fire(state) {
        for (const fn of this.listeners) fn(state);
      },
    };
  }

  async register() {
    if (scenario === "rejected") {
      this.stateChange.fire("Unregistered");
      throw new Error("403 Forbidden");
    }
  }
}

class StubUserAgent {
  constructor(options) {
    this.options = options;
    this.connected = false;
  }

  static makeURI(uri) {
    return { toString: () => uri };
  }

  isConnected() {
    return this.connected;
  }

  async start() {
    if (scenario === "transport") throw new Error("connection failed");
  }

  async stop() {}

  async reconnect() {}
}

globalThis.SIP = {
  UserAgent: StubUserAgent,
  Registerer: StubRegisterer,
  Web: {
    defaultSessionDescriptionHandlerFactory: () => () => ({}),
  },
  RegistererState: {
    Initial: "Initial",
    Registered: "Registered",
    Unregistered: "Unregistered",
    Terminated: "Terminated",
  },
  SessionState: { Terminated: "Terminated" },
};

// fetch: the live-session probe answers a session; DELETEs are
// recorded — the stale-credentials path must be the ONLY deleter.
globalThis.fetch = async (url, opts = {}) => {
  const method = opts.method ?? "GET";
  calls.fetches.push(`${method} ${url}`);
  if (String(url).endsWith("/api/session") && method === "GET") {
    return {
      ok: true,
      status: 200,
      json: async () => ({
        extension: "1001",
        password: "pw",
        did: "+441632960900",
      }),
    };
  }
  return { ok: true, status: 204, json: async () => ({}) };
};

const phoneHidden = () => doc.getElementById("phone-view").hidden;
const loginHidden = () => doc.getElementById("login-view").hidden;
const bannerHidden = () => doc.getElementById("offline-banner").hidden;
const logTexts = () =>
  doc.getElementById("log").children.map((li) => li.textContent);
const flushes = async (rounds = 12) => {
  while (rounds--) await new Promise((resolve) => setImmediate(resolve));
};

test("resume over a dead transport KEEPS the session and shows the offline phone", async () => {
  scenario = "transport";
  await import("../island/app/main.js?case=transport");
  await flushes();
  assert.equal(phoneHidden(), false, "phone view visible");
  assert.equal(loginHidden(), true, "login stays hidden");
  assert.equal(bannerHidden(), false, "offline banner up (honest state)");
  assert.ok(
    !calls.fetches.some((c) => c.startsWith("DELETE")),
    "a transport failure must not delete the server session",
  );
  assert.ok(
    logTexts().some((line) => line.includes("SIP unreachable")),
    "the kept-session reason is logged",
  );
});

test("resume with a refused REGISTER drops the stale session and returns the login", async () => {
  scenario = "rejected";
  await import("../island/app/main.js?case=rejected");
  await flushes();
  assert.equal(loginHidden(), false, "login visible");
  assert.ok(
    calls.fetches.some((c) => c.startsWith("DELETE /api/session")),
    "stale credentials drop the session",
  );
  assert.equal(
    doc.getElementById("login-error").hidden,
    false,
    "the rejection reason is shown where the user looks",
  );
});
