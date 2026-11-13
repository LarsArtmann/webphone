// M20: every panel load failure is honest. The voicemail refresh used to
// answer a NETWORK failure with the auth-failure copy (a lie about the
// cause), and the history/contacts loads swallowed !res.ok silently —
// the panel just stayed stale with no word. Each failure now names
// itself: voicemail keeps vmAuthFailed for the 401 arm and grows a
// distinct vmLoadFailed for transport/parse failures; history and
// contacts announce a warn toast on both the !res.ok arm and the
// thrown-fetch arm (the dedupe in announce() keeps SSE-nudge storms to
// one visible toast).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = { phoneApi: true, asr: false };

const { setCredentials } = await import("../island/app/auth.js");
setCredentials("1000", "pw");

const panels = await import("../island/app/panels.js?test=load-failures");

const toasts = () => [...doc.getElementById("toasts").children].map((el) => el.textContent);

// A failing response stub: every panel fetch answers 500 unless the test
// overrides. ok=false without ok being consulted is fine — panels branch
// on res.ok / status only.
const failFetch = (status) => async (url) => ({
  ok: false,
  status,
  headers: new Headers(),
  json: async () => ({}),
});

test("history load failure names itself instead of staying silent", async () => {
  globalThis.fetch = failFetch(500);
  await panels.refreshServerHistory();
  assert.ok(
    toasts().some((t) => t.includes("could not load call history: 500")),
    `expected the history failure toast, got ${JSON.stringify(toasts())}`,
  );
});

test("history load network error announces the reason", async () => {
  globalThis.fetch = async () => {
    throw new Error("socket gone");
  };
  await panels.refreshServerHistory();
  assert.ok(
    toasts().some((t) => t.includes("could not load call history: socket gone")),
    `expected the network-failure toast, got ${JSON.stringify(toasts())}`,
  );
});

test("contacts load failure names itself instead of staying silent", async () => {
  globalThis.fetch = failFetch(503);
  await panels.loadContacts();
  assert.ok(
    toasts().some((t) => t.includes("could not load contacts: 503")),
    `expected the contacts failure toast, got ${JSON.stringify(toasts())}`,
  );
});

test("contacts load network error announces the reason", async () => {
  globalThis.fetch = async () => {
    throw new Error("dns soup");
  };
  await panels.loadContacts();
  assert.ok(
    toasts().some((t) => t.includes("could not load contacts: dns soup")),
    `expected the network-failure toast, got ${JSON.stringify(toasts())}`,
  );
});

test("voicemail transport failure gets its own copy, not the auth line", async () => {
  globalThis.fetch = async () => {
    throw new Error("gateway timeout");
  };
  await panels.refreshVoicemail();
  const status = doc.getElementById("vm-status").textContent;
  assert.ok(
    status.includes("could not load voicemail") && status.includes("gateway timeout"),
    `expected distinct load-failure copy, got "${status}"`,
  );
  assert.ok(
    !status.includes("operator API"),
    "the auth-failure copy must not answer a transport failure",
  );
});

test("voicemail 401 keeps the auth-failure copy", async () => {
  globalThis.fetch = failFetch(401);
  await panels.refreshVoicemail();
  const status = doc.getElementById("vm-status").textContent;
  assert.ok(status.includes("operator API"), `expected the auth copy on 401, got "${status}"`);
});
