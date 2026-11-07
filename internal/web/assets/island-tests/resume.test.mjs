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
const logTexts = () => doc.getElementById("log").children.map((li) => li.textContent);
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

// T19 M15/M21: the guards boot — a healthy connect — drives the login
// in-flight guard and the logout two-click arm.
const { sessions } = await import("../island/app/state.js");
const { t } = await import("../island/app/i18n.js");
const lastToast = () => doc.getElementById("toasts").children.at(-1);

test("a double-submitted login POSTs the session exactly once", async () => {
  scenario = "guards";
  await import("../island/app/main.js?case=guards");
  await flushes();

  const form = doc.getElementById("login-form");
  const submitsBefore = calls.fetches.filter((f) => f.startsWith("POST /api/session")).length;
  // Three boots registered submit listeners on the shared registry
  // element — fire only the guards instance's handler.
  const fire = () => form.listeners.submit.at(-1)({ preventDefault() {} });
  const first = fire();
  fire();
  await first;
  await flushes();
  const afterPair = calls.fetches.filter((f) => f.startsWith("POST /api/session")).length;
  assert.equal(
    afterPair - submitsBefore,
    1,
    "the concurrent second submit is dropped by the in-flight guard",
  );

  // The guard RELEASES: a later submit runs a fresh attempt.
  const second = fire();
  await second;
  await flushes();
  const afterRetry = calls.fetches.filter((f) => f.startsWith("POST /api/session")).length;
  assert.equal(afterRetry - afterPair, 1, "the guard is not a latch");
});

test("logout with a live call is a two-click move", async () => {
  scenario = "guards";
  sessions.set("live-1", { session: { bye: async () => {} } });
  doc.getElementById("phone-view").hidden = false;
  const logout = doc.getElementById("logout");
  const logoutClick = () => logout.listeners.click.at(-1)({});
  const fetchesBefore = calls.fetches.length;

  logoutClick();
  assert.equal(lastToast().textContent, t("logoutLiveCall"), "the first click warns");
  assert.equal(
    doc.getElementById("phone-view").hidden,
    false,
    "the first click does not tear down",
  );
  assert.equal(calls.fetches.length, fetchesBefore, "the armed click drops no server session");
  // The confirmed click and the idle one-click path run the ORIGINAL
  // teardown (destroySession → page reload), which the stub harness
  // bans by design — the arm is this train's new surface and the pin
  // above proves a misclick cannot hang up on the operator.
  sessions.clear();
});
