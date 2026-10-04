// enroll.js under node:test — the standalone enrollment flow. The
// contract under test: the one-time token is consumed via the verify
// endpoint (which burns it server-side), the registration ceremony
// rides the same coercion helpers as login, and every failure class
// answers the ONE anti-oracle message instead of splitting statuses.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
const { t } = await import("../island/app/i18n.js");

// The enroll module lives OUTSIDE the island tree (its page loads no
// island runtime), but shares the i18n + webauthn helpers by import.
await import("../enroll/enroll.js");

const { els } = await import("../island/app/ui.js");
const form = doc.getElementById("enroll-form");
const tokenInput = doc.getElementById("enroll-token");
const nameInput = doc.getElementById("enroll-credential-name");
const statusEl = doc.getElementById("enroll-status");
const errorEl = doc.getElementById("enroll-error");

const posted = [];
let verifyStatus = 200;
let finishStatus = 200;
let createScript = null;
let fetchThrows = false;

globalThis.navigator.credentials = {
  create: async () => {
    if (createScript instanceof Error) throw createScript;
    return createScript;
  },
};

globalThis.fetch = async (url, opts) => {
  if (fetchThrows) throw new Error("network down");
  posted.push({ url, opts });
  if (url === "/api/auth/passkey/enroll/verify") {
    if (verifyStatus !== 200) {
      return { ok: false, status: verifyStatus, json: async () => ({}) };
    }
    return {
      ok: true,
      status: 200,
      json: async () => ({ email: "lars@example.com", user_id: "user-7" }),
    };
  }
  if (url === "/api/auth/passkey/enroll/begin") {
    return {
      ok: true,
      status: 200,
      json: async () => ({
        options: {
          challenge: "AAAAAQ",
          rp: { id: "pbx.example.org", name: "WebPhone" },
          user: { id: "AQIDBA", name: "lars@example.com", displayName: "Lars" },
          pubKeyCredParams: [{ type: "public-key", alg: -7 }],
        },
        session_key: "ceremony-1",
      }),
    };
  }
  if (url.startsWith("/api/auth/passkey/enroll/finish")) {
    if (finishStatus !== 200) {
      return { ok: false, status: finishStatus, json: async () => ({}) };
    }
    return { ok: true, status: 200, json: async () => ({ status: "registered" }) };
  }
  return { ok: true, status: 200, json: async () => ({}) };
};

const attestationCredential = {
  id: "AAECAw",
  rawId: new Uint8Array([0, 1, 2, 3]).buffer,
  type: "public-key",
  response: {
    clientDataJSON: new Uint8Array([1]).buffer,
    attestationObject: new Uint8Array([7, 7, 7]).buffer,
    transports: ["internal"],
  },
};

const settle = () => new Promise((resolve) => setTimeout(resolve, 25));

function submit() {
  const list = form.listeners.submit;
  assert.ok(list && list.length === 1, "submit listener bound");
  list[0]({ preventDefault() {} });
  return settle();
}

function reset() {
  posted.length = 0;
  verifyStatus = 200;
  finishStatus = 200;
  createScript = attestationCredential;
  fetchThrows = false;
  form.hidden = false;
  statusEl.hidden = true;
  errorEl.hidden = true;
  tokenInput.value = "tok";
  nameInput.value = "laptop";
}

test("happy path: verify → begin → create → finish, success replaces the form", async () => {
  reset();
  await submit();

  const verify = posted.find((p) => p.url.endsWith("/verify"));
  assert.deepEqual(JSON.parse(verify.opts.body), { token: "tok" });
  assert.equal(posted.filter((p) => p.url.includes("/begin")).length, 1, "ceremony began for the verified user");

  const finish = posted.find((p) => p.url.startsWith("/api/auth/passkey/enroll/finish"));
  assert.ok(
    finish.url.includes("user_id=ceremony-1") && finish.url.includes("credential_name=laptop"),
    "ceremony key and device name ride the query",
  );
  const finishBody = JSON.parse(finish.opts.body);
  assert.equal(finishBody.response.attestationObject, "BwcH");
  assert.deepEqual(finishBody.response.transports, ["internal"]);

  assert.equal(form.hidden, true, "form replaced by the success status");
  assert.equal(statusEl.textContent, t("enrollSuccess"));
  assert.equal(errorEl.hidden, true);
});

test("a burned token answers the ONE failure message, ceremony never starts", async () => {
  reset();
  verifyStatus = 503;
  await submit();
  assert.equal(errorEl.hidden, false);
  assert.equal(errorEl.textContent, t("enrollFailed")(503));
  assert.equal(posted.filter((p) => p.url.includes("/begin")).length, 0);
});

test("a network failure says the network message, never 'HTTP 0'", async () => {
  reset();
  fetchThrows = true;
  await submit();
  assert.equal(errorEl.hidden, false);
  assert.equal(errorEl.textContent, t("enrollNetFailed"));
  assert.equal(posted.filter((p) => p.url.includes("/begin")).length, 0);
});

test("an empty token asks for the link's token, ceremony never starts", async () => {
  reset();
  tokenInput.value = "   ";
  await submit();
  assert.equal(errorEl.hidden, false);
  assert.equal(errorEl.textContent, t("enrollTokenMissing"));
  assert.equal(posted.length, 0, "nothing POSTs without a token");
});

test("a cancelled prompt says so — the token is spent, the copy is honest", async () => {
  reset();
  createScript = Object.assign(new Error("cancelled"), { name: "NotAllowedError" });
  await submit();
  assert.equal(errorEl.hidden, false);
  assert.equal(errorEl.textContent, t("enrollDismissed"));
  assert.equal(posted.filter((p) => p.url.includes("/finish")).length, 0);
});
