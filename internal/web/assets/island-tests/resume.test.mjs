// main.js boot resume under node:test — the two failure classes of a
// resumed session's SIP connect, AFTER the auth-front-door train. A
// REGISTER the server refuses means the stored credentials are stale:
// the session is dropped QUIETLY and the browser goes back to the /login
// front door (the ?error= query carries the reason; the island has no
// login form to show anymore). A transport failure (PBX/WebSocket down
// at load) is NOT a stale session: the cookie is KEPT, the phone shows
// its honest offline state, and the reconnect backoff owns recovery —
// the tabs stay usable. Found live by the T23 visual harness: the
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
const calls = { fetches: [], replaces: [] };

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

// window.location.replace is the island's ONLY navigation primitive now
// (the front door owns the login surface); record where it sends the
// browser instead of following.
globalThis.window.location = {
  replace(url) {
    calls.replaces.push(String(url));
  },
};

// The login page's submit requests notification permission under its own
// gesture; the boot states the RESULT in #log (the stub answers
// "granted" — the assertion pins the line's presence, not the verdict).
globalThis.Notification = { permission: "granted" };

const phoneHidden = () => doc.getElementById("phone-view").hidden;
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
  assert.equal(bannerHidden(), false, "offline banner up (honest state)");
  assert.ok(
    !calls.fetches.some((c) => c.startsWith("DELETE")),
    "a transport failure must not delete the server session",
  );
  assert.ok(
    logTexts().some((line) => line.includes("SIP unreachable")),
    "the kept-session reason is logged",
  );
  assert.ok(
    logTexts().some((line) => line.includes("notifications: granted")),
    "the boot states the notification-permission result in #log",
  );
  assert.equal(
    calls.replaces.length,
    0,
    "a transport failure must NOT bounce the browser to the login page — the session is good",
  );
});

test("resume with a refused REGISTER drops the stale session and returns to the front door", async () => {
  scenario = "rejected";
  await import("../island/app/main.js?case=rejected");
  await flushes();
  assert.ok(
    calls.fetches.some((c) => c.startsWith("DELETE /api/session")),
    "stale credentials drop the session",
  );
  assert.ok(
    calls.replaces.some((url) => url.startsWith("/login?error=")),
    "the browser goes back to /login, carrying the reason for its alert",
  );
});

// T19 logout arm: a live call forces the two-click move — the first
// click warns, the armed click tears down nothing. The confirmed click
// and the idle one-click path run destroySession → page reload, which
// the stub harness bans by design.
const { sessions } = await import("../island/app/state.js");
const { t } = await import("../island/app/i18n.js");
const lastToast = () => doc.getElementById("toasts").children.at(-1);

test("logout with a live call is a two-click move", async () => {
  scenario = "guards";
  await import("../island/app/main.js?case=guards");
  await flushes();
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
  sessions.clear();
});
