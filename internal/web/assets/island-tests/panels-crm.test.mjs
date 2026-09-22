// recordCrmCall under node:test — the post-call journal report. Own file
// because config.js reads PBX_CONFIG.crm once per module registry: this
// file boots with crm on, while panels.test.mjs pins the crm-off no-op on
// its own instance. The body shape must match the server's /api/calls
// contract, including the per-report idempotency key the server dedupes on.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
globalThis.window.PBX_CONFIG = {
  phoneApi: true,
  crm: true,
  contacts: [],
};

let callPosts = [];
const contactsShape = () => ({
  ok: true,
  status: 200,
  headers: new Headers(),
  json: async () => ({ personal: [], shared: [] }),
});
globalThis.fetch = async (input, init) => {
  if (String(input).includes("/api/calls")) {
    callPosts.push({ url: String(input), body: init?.body ? JSON.parse(init.body) : null });
    return { ok: true, status: 204, headers: new Headers() };
  }
  return contactsShape();
};

const panels = await import("../island/app/panels.js?test=crm-journal");

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

test("the journal report carries the call facts and a fresh idempotency key", async () => {
  callPosts = [];
  await panels.recordCrmCall({ dir: "out", target: "+493012345678", dur: 42, established: true });
  await panels.recordCrmCall({ dir: "in", target: "+493012345678", dur: 0, established: false });

  assert.equal(callPosts.length, 2, "each ended call reports exactly once");
  const [first, second] = callPosts.map((p) => p.body);
  assert.deepEqual(first, {
    number: "+493012345678",
    direction: "out",
    seconds: 42,
    outcome: "answered",
    key: first.key,
  });
  assert.match(first.key, UUID_RE, "the key must be a UUID");
  assert.deepEqual(second, {
    number: "+493012345678",
    direction: "in",
    seconds: 0,
    outcome: "missed",
    key: second.key,
  });
  assert.notEqual(first.key, second.key, "every report must carry a fresh key");
});

test("a failed journal report toasts instead of vanishing", async () => {
  globalThis.fetch = async (input) => {
    if (String(input).includes("/api/calls")) {
      return { ok: false, status: 502, headers: new Headers() };
    }
    return contactsShape();
  };

  await panels.recordCrmCall({ dir: "in", target: "+493012345678", dur: 5, established: true });

  const toast = doc.getElementById("toasts").children.at(-1);
  assert.ok(toast, "a CRM failure must surface a toast");
  assert.match(toast.textContent, /HTTP 502/, `toast must carry the detail, got: ${toast.textContent}`);
});
