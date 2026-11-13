// login/login-page.js under node:test — the extension+password submit
// that MOVED to the /login page in the auth-front-door train (the island
// no longer owns a login POST). The feedback contract changed shape with
// it: server rejections answer INLINE where the user is looking (the
// library card's #login-error line), a 201 signs the browser home where
// the island's boot probe resumes and REGISTERs, and the remember
// checkbox stores the extension — never the password — under the same
// `pbx-extension` key the old island form used.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
// The script's meta reader guards with instanceof; node ships no
// HTMLMetaElement, so the guard needs the global to exist.
globalThis.HTMLMetaElement = globalThis.HTMLMetaElement || class HTMLMetaElement {};
await import("../login/login-page.js");

// fetch stub: one POST /api/session per case. handler(body) → {status,
// headers?, body?}.
let posts = [];
function useFetch(handler) {
  posts = [];
  globalThis.fetch = async (url, opts = {}) => {
    const out = handler(url, opts);
    if (String(url).includes("/api/session") && (opts.method ?? "GET") === "POST") {
      posts.push({ url: String(url), body: opts.body });
    }
    return {
      ok: out.status >= 200 && out.status < 300,
      status: out.status,
      headers: { get: (k) => out.headers?.[k] ?? null },
      json: async () => out.body ?? {},
    };
  };
}

const form = () => doc.getElementById("login-form");
const extEl = () => doc.getElementById("ext");
const passEl = () => doc.getElementById("pass");
const rememberEl = () => doc.getElementById("remember");
const errorEl = () => doc.getElementById("login-error");
const replaces = () => {
  const seen = [];
  globalThis.window.location = {
    replace(url) {
      seen.push(String(url));
    },
  };
  return seen;
};

const submit = async () => {
  const fire = form().listeners.submit.at(-1)({ preventDefault() {} });
  await fire;
  // The submit handler's catch-all reports through the inline error; let
  // the microtask queue drain so async failures surface as rejections.
  await new Promise((resolve) => setImmediate(resolve));
};

test("a 201 signs in and sends the browser home (the boot probe resumes)", async () => {
  useFetch(() => ({ status: 201, body: { did: "+441632960961" } }));
  const seen = replaces();
  extEl().value = " 1001 ";
  passEl().value = "pw";
  rememberEl().checked = true;
  let permissionAsked = false;
  globalThis.Notification = {
    permission: "default",
    requestPermission: async () => {
      permissionAsked = true;
      return "granted";
    },
  };
  await submit();
  assert.equal(posts.length, 1, "exactly one POST /api/session");
  assert.deepEqual(
    JSON.parse(posts[0].body),
    { extension: "1001", password: "pw" },
    "the extension is trimmed, the password is not",
  );
  assert.equal(
    permissionAsked,
    true,
    "the submit click is the notification-permission gesture (the old island login's behavior)",
  );
  assert.equal(
    localStorage.getItem("pbx-extension"),
    "1001",
    "the remembered extension survives (never the password)",
  );
  assert.equal(localStorage.getItem("pbx-extension-password"), null);
  assert.deepEqual(seen, ["/"], "the redirect home hands over to the island boot probe");
});

test("a 401 answers inline where the user is looking", async () => {
  useFetch(() => ({ status: 401 }));
  const seen = replaces();
  localStorage.setItem("pbx-extension", "kept");
  extEl().value = "1001";
  passEl().value = "wrong";
  await submit();
  assert.equal(errorEl().hidden, false, "the inline error line shows");
  assert.match(errorEl().textContent, /Wrong extension or password/);
  assert.deepEqual(seen, [], "a rejection must not navigate");
  assert.equal(
    localStorage.getItem("pbx-extension"),
    "kept",
    "a failed login never writes or clears the remembered extension",
  );
});

test("a 429 names the wait from Retry-After", async () => {
  useFetch(() => ({ status: 429, headers: { "Retry-After": "7" } }));
  extEl().value = "1001";
  passEl().value = "pw";
  await submit();
  assert.match(errorEl().textContent, /7s/);
});

test("an unchecked remember forgets the previously stored extension", async () => {
  localStorage.setItem("pbx-extension", "1000");
  useFetch(() => ({ status: 201, body: {} }));
  replaces();
  extEl().value = "1001";
  passEl().value = "pw";
  rememberEl().checked = false;
  await submit();
  assert.equal(
    localStorage.getItem("pbx-extension"),
    null,
    "signing in without remember clears the stored extension",
  );
});

test("an empty submit asks for the credentials instead of POSTing", async () => {
  useFetch(() => ({ status: 201, body: {} }));
  extEl().value = "";
  passEl().value = "";
  await submit();
  assert.equal(posts.length, 0, "no POST without both fields");
  assert.match(errorEl().textContent, /Enter your extension and password/);
});

test("blocked site data must not kill the sign-in (private-mode storage)", async () => {
  const good = globalThis.localStorage;
  // Browser privacy settings that block site data make the localStorage
  // property itself THROW — every storage access in the page script
  // must degrade instead of dying (the boot-time remembered read runs
  // under this stub via the fresh module instance below).
  globalThis.localStorage = {
    getItem() {
      throw new Error("SecurityError: blocked site data");
    },
    setItem() {
      throw new Error("SecurityError: blocked site data");
    },
    removeItem() {
      throw new Error("SecurityError: blocked site data");
    },
  };
  try {
    await import("../login/login-page.js?blocked-storage");
    useFetch(() => ({ status: 201, body: {} }));
    const seen = replaces();
    extEl().value = "1001";
    passEl().value = "pw";
    rememberEl().checked = true;
    await submit();
    assert.equal(posts.length, 1, "the sign-in still POSTs");
    assert.deepEqual(seen, ["/"], "the redirect home still happens");
    assert.equal(
      errorEl().hidden,
      true,
      "no error line — a storage failure is not a sign-in failure",
    );
  } finally {
    globalThis.localStorage = good;
  }
});
