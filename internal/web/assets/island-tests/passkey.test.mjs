// passkey.js under node:test — the email-first login ceremony. The
// order is the contract under test: finish mints the server session
// BEFORE the island registers (so a failed SIP transport still leaves
// the tabs working), the whoami line shows the user's real identity
// (display name + numbers), and every rejection class lands in
// #passkey-login-error instead of the console alone.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
const { t } = await import("../island/app/i18n.js");
await import("../island/app/ui.js");
const { state } = await import("../island/app/state.js");
const { initPasskeyLogin } = await import("../island/app/passkey.js");

// --- minimal SIP stub (leaner twin of connection.test.mjs's) -----------

class StubEmitter {
  constructor() {
    this.listeners = [];
  }
  addListener(fn) {
    this.listeners.push(fn);
  }
}

class StubRegisterer {
  constructor() {
    this.state = "Initial";
    this.stateChange = new StubEmitter();
  }
  async register() {
    this.state = "Registered";
    for (const fn of this.stateChange.listeners) fn("Registered");
  }
}

const agents = [];

globalThis.SIP = {
  UserAgent: class {
    constructor(options) {
      this.options = options;
      agents.push(this);
    }
    static makeURI(uri) {
      return { toString: () => uri };
    }
    async start() {}
    async stop() {}
  },
  Registerer: StubRegisterer,
  Web: {
    defaultSessionDescriptionHandlerFactory: () => () => ({}),
  },
  RegistererState: { Registered: "Registered", Unregistered: "Unregistered" },
  SessionState: { Terminated: "Terminated" },
};

const silentNode = () => ({ connect: () => silentNode(), start() {}, stop() {} });
globalThis.AudioContext = class {
  constructor() {
    this.currentTime = 0;
    this.destination = {};
  }
  createOscillator() {
    return { frequency: {}, connect: silentNode, start() {}, stop() {} };
  }
  createGain() {
    return { gain: {}, connect: silentNode };
  }
};

// --- fetch + credentials stubs ------------------------------------------

const posted = [];
let beginStatus = 200;
let finishStatus = 201;
let credentialScript = null;
let finishCalls = 0;

globalThis.navigator.credentials = {
  get: async () => {
    if (credentialScript instanceof Error) throw credentialScript;
    return credentialScript;
  },
};

globalThis.fetch = async (url, opts) => {
  posted.push({ url, opts });
  if (url === "/api/auth/passkey/begin") {
    if (beginStatus !== 200) return { ok: false, status: beginStatus, json: async () => ({}) };
    return {
      ok: true,
      status: 200,
      json: async () => ({
        options: { challenge: "AAAAAQ", rpId: "pbx.example.org", allowCredentials: [{ id: "AAECAw", type: "public-key" }] },
        session_key: "user-1",
      }),
    };
  }
  if (url.startsWith("/api/auth/passkey/finish")) {
    finishCalls += 1;
    if (finishStatus !== 201) return { ok: false, status: finishStatus, json: async () => ({}) };
    return {
      ok: true,
      status: 201,
      json: async () => ({
        extension: "1000",
        password: "sip-secret",
        did: "+17287289311",
        display_name: "Lars",
        numbers: ["+17287289311"],
      }),
    };
  }
  if (url === "/api/csrf") {
    return { ok: true, status: 200, json: async () => ({ token: "fresh-token" }) };
  }
  return { ok: true, status: 200, json: async () => ({}) };
};

const assertiveCredential = {
  id: "AAECAw",
  rawId: new Uint8Array([0, 1, 2, 3]).buffer,
  type: "public-key",
  response: {
    clientDataJSON: new Uint8Array([1, 2]).buffer,
    authenticatorData: new Uint8Array([3]).buffer,
    signature: new Uint8Array([4, 5]).buffer,
    userHandle: null,
  },
};

const { els } = await import("../island/app/ui.js");

initPasskeyLogin();
const submitHandler = () => {
  const list = els.passkeyForm.listeners.submit;
  assert.ok(list && list.length === 1, "submit listener bound");
  list[0]({ preventDefault() {} });
  // The listener is deliberately fire-and-forget (a DOM listener must
  // not be awaited in production); the test settles the async chain
  // instead of racing it.
  return settle();
};

function reset() {
  posted.length = 0;
  finishCalls = 0;
  beginStatus = 200;
  finishStatus = 201;
  credentialScript = assertiveCredential;
  els.passkeyError.hidden = true;
  els.passkeyError.textContent = "";
  els.passkeyEmail.value = "";
}

const settle = () => new Promise((resolve) => setTimeout(resolve, 25));

const openedEvents = [];
doc.addEventListener("wp:session-opened", (event) => openedEvents.push(event.detail));

test("happy path: session minted, island registered, identity rendered", async () => {
  reset();
  credentialScript = assertiveCredential;
  els.passkeyEmail.value = "lars@example.com";
  await submitHandler();

  const begin = posted.find((p) => p.url === "/api/auth/passkey/begin");
  assert.deepEqual(JSON.parse(begin.opts.body), { email: "lars@example.com" });

  const finish = posted.find((p) => p.url.startsWith("/api/auth/passkey/finish"));
  assert.ok(finish.url.includes("user_id=user-1"), "session_key rides the query");
  const finishBody = JSON.parse(finish.opts.body);
  assert.equal(finishBody.rawId, "AAECAw");
  assert.equal(finishBody.response.clientDataJSON, "AQI");
  assert.equal("userHandle" in finishBody.response, false);

  assert.ok(agents.length >= 1, "SIP user agent built");
  assert.equal(agents.at(-1).options.authorizationUsername, "1000");
  assert.equal(agents.at(-1).options.authorizationPassword, "sip-secret");
  assert.equal(state.userAgent !== null, true);

  assert.equal(els.whoami.textContent, "Lars · +17287289311");
  assert.equal(els.loginView.hidden, true);
  assert.equal(els.phoneView.hidden, false);
  assert.equal(els.passkeyError.hidden, true);

  assert.equal(openedEvents.length, 1);
  assert.equal(openedEvents[0].did, "+17287289311");
  assert.equal(openedEvents[0].displayName, "Lars");
});

test("unknown email renders the localized anti-enumeration message", async () => {
  reset();
  beginStatus = 401;
  els.passkeyEmail.value = "nobody@example.com";
  await submitHandler();
  assert.equal(els.passkeyError.hidden, false);
  assert.equal(els.passkeyError.textContent, t("passkeyUnknown"));
  // The ceremony stopped at begin: no credential was requested, no
  // finish POST happened.
  assert.equal(finishCalls, 0);
  assert.equal(posted.length, 1);
});

test("a dismissed passkey prompt is quiet — no error, no finish call", async () => {
  reset();
  credentialScript = Object.assign(new Error("dismissed"), { name: "NotAllowedError" });
  els.passkeyEmail.value = "lars@example.com";
  await submitHandler();
  assert.equal(els.passkeyError.hidden, true);
  assert.equal(finishCalls, 0);
  assert.equal(posted.length, 1);
});

test("throttled begin answers the localized message", async () => {
  reset();
  beginStatus = 429;
  els.passkeyEmail.value = "lars@example.com";
  await submitHandler();
  assert.equal(els.passkeyError.hidden, false);
  assert.equal(els.passkeyError.textContent, t("passkeyThrottled"));
});

test("finish rejection surfaces the status, session stays unopened", async () => {
  reset();
  finishStatus = 503;
  els.passkeyEmail.value = "lars@example.com";
  const openedBefore = openedEvents.length;
  await submitHandler();
  assert.equal(els.passkeyError.hidden, false);
  assert.equal(els.passkeyError.textContent, t("passkeyFinishFailed")(503));
  assert.equal(openedEvents.length, openedBefore);
});

test("whoamiLine renders extension sessions identically to before", async () => {
  const { whoamiLine } = await import("../island/app/ui.js");
  assert.equal(
    whoamiLine({ extension: "1000" }),
    "1000@pbx.example.org",
  );
  assert.equal(
    whoamiLine({ extension: "1000", numbers: ["+17287289311"] }),
    "1000@pbx.example.org · +17287289311",
  );
  assert.equal(
    whoamiLine({ extension: "1000", display_name: "Lars" }),
    "Lars · 1000@pbx.example.org",
  );
});
