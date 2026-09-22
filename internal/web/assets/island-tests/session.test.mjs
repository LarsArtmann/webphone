// session.js under node:test — the server-session feedback map. A dead
// server session is deliberately non-fatal (the island keeps calling),
// but non-fatal must never mean invisible: every rejection class toasts
// in the user's language, and a genuinely dead SSE feed is reported once
// instead of after every reconnect attempt.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
const { t } = await import("../island/app/i18n.js");
const ui = await import("../island/app/ui.js");
const session = await import("../island/app/session.js");

const toastsHost = () => doc.getElementById("toasts");
const resetToasts = () => {
  const host = toastsHost();
  while (host.firstChild) host.firstChild.remove();
  return host;
};
const lastToast = () => {
  const el = toastsHost().children.at(-1);
  return el ? { className: el.className, text: el.textContent } : null;
};
const logTexts = () => doc.getElementById("log").children.map((li) => li.textContent);

// fetch stub: createSession and adoptFreshCsrfToken are the only fetchers
// on this path. handler(url) → {status, body} or throws.
function useFetch(handler) {
  globalThis.fetch = async (url, opts) => {
    const out = handler(url, opts);
    if (out instanceof Error) throw out;
    return {
      ok: out.status >= 200 && out.status < 300,
      status: out.status,
      json: async () => out.body ?? {},
    };
  };
}

test("a 401 from the server login toasts the credential-specific advice", async () => {
  resetToasts();
  useFetch(() => ({ status: 401 }));
  await session.createSession("1001", "pw");
  const toast = lastToast();
  assert.equal(toast.className, "toast toast-warn", "non-fatal: warn, not error");
  assert.equal(toast.text, t("sessionFailed401"));
  assert.ok(
    logTexts().some((line) => line.includes("HTTP 401")),
    "the operator-facing #log line must survive the toast addition",
  );
});

test("a 429 from the server login toasts the throttle advice", async () => {
  resetToasts();
  useFetch(() => ({ status: 429 }));
  await session.createSession("1001", "pw");
  assert.equal(lastToast().text, t("sessionThrottled"));
});

test("other HTTP failures toast the generic status copy", async () => {
  resetToasts();
  useFetch(() => ({ status: 503 }));
  await session.createSession("1001", "pw");
  assert.equal(lastToast().text, t("sessionFailed")(503));
});

test("an unreachable server toasts the network copy", async () => {
  resetToasts();
  useFetch(() => new Error("ECONNREFUSED"));
  await session.createSession("1001", "pw");
  assert.equal(lastToast().text, t("sessionNetFailed"));
});

test("a successful login still toasts nothing (feedback stays quiet)", async () => {
  resetToasts();
  useFetch((url) =>
    url === "/api/session"
      ? { status: 201, body: { did: "+441632960961" } }
      : { status: 200, body: { token: "fresh" } },
  );
  await session.createSession("1001", "pw");
  assert.equal(lastToast(), null);
});

test("the SSE pill is a labeled, localized state indicator (not aria-hidden)", () => {
  session.initSseLiveIndicator();
  const pill = doc.body.children.find((el) => el.id === "wp-sse-live");
  assert.ok(pill, "pill must exist");
  assert.equal(pill.getAttribute("aria-hidden"), null, "must not be hidden from AT");
  assert.equal(pill.getAttribute("role"), "img");
  assert.equal(pill.getAttribute("aria-label"), t("ssePillDown"));
  doc.dispatch("htmx:sseOpen", {});
  assert.equal(pill.getAttribute("aria-label"), t("ssePillLive"));
  doc.dispatch("htmx:sseError", {});
  assert.equal(pill.getAttribute("aria-label"), t("ssePillDown"));
});

test("the SSE feed toasts once after three consecutive failures", () => {
  resetToasts();
  session.initSseLiveIndicator();
  doc.dispatch("htmx:sseOpen", {}); // deterministic counter start

  doc.dispatch("htmx:sseError", {});
  doc.dispatch("htmx:sseError", {});
  assert.equal(lastToast(), null, "a flapping feed must stay quiet");

  doc.dispatch("htmx:sseError", {});
  assert.equal(lastToast().text, t("sseDropped"));
  assert.equal(lastToast().className, "toast toast-warn");

  doc.dispatch("htmx:sseError", {});
  const count = toastsHost().children.length;
  doc.dispatch("htmx:sseError", {});
  assert.equal(
    toastsHost().children.length,
    count,
    "a fourth failure must not stack another toast",
  );

  doc.dispatch("htmx:sseOpen", {});
  doc.dispatch("htmx:sseError", {});
  doc.dispatch("htmx:sseError", {});
  assert.equal(toastsHost().children.length, count, "a recovered feed resets the counter");
});

// --- session resume ---------------------------------------------------------

test("fetchLiveSession hands the resumed credentials back on 200", async () => {
  useFetch(() => ({
    status: 200,
    body: { extension: "1001", password: "pw", did: "+441632960961" },
  }));
  const data = await session.fetchLiveSession();
  assert.deepEqual(data, {
    extension: "1001",
    password: "pw",
    did: "+441632960961",
  });
});

test("fetchLiveSession maps every no-resume case to null", async () => {
  for (const status of [401, 429, 500]) {
    useFetch(() => ({ status }));
    assert.equal(await session.fetchLiveSession(), null, `HTTP ${status} must resume nothing`);
  }
  useFetch(() => ({ status: 200, body: {} }));
  assert.equal(
    await session.fetchLiveSession(),
    null,
    "a 200 without credentials must resume nothing",
  );
  useFetch(() => new Error("ECONNREFUSED"));
  assert.equal(await session.fetchLiveSession(), null, "an unreachable server must resume nothing");
});

test("signOutQuiet drops the server session without a reload", async () => {
  let seen = null;
  globalThis.fetch = async (url, opts) => {
    seen = { url, opts };
    return { ok: true, status: 204, json: async () => ({}) };
  };
  await session.signOutQuiet();
  assert.equal(seen.opts.method, "DELETE");
  assert.equal(seen.url, "/api/session");
  // The quiet path never reloads: the stub's location.reload THROWS, so
  // merely completing this test proves no reload happened.
});

// --- CSRF adoption retry ladder (SUPERB T14a) -------------------------------
// Login rotates the CSRF token, so the island must adopt the fresh one
// before any POST. A transient failure of GET /api/csrf retries with
// backoff (x3) and only the THIRD consecutive failure falls back to a
// page reload (the cookie session survives; the served page then
// carries a matching token again).

test("csrf adoption recovers on the second try without a reload", async () => {
  resetToasts();
  let csrfCalls = 0;
  useFetch((url) => {
    if (String(url).includes("/api/session")) {
      return { status: 201, body: { extension: "1001", password: "pw" } };
    }
    csrfCalls += 1;
    if (csrfCalls < 2) return { status: 503 };
    return { status: 200, body: { token: "fresh-token" } };
  });
  let reloads = 0;
  globalThis.window.location = {
    reload() {
      reloads += 1;
    },
  };
  try {
    await session.createSession("1001", "pw");
    assert.equal(csrfCalls, 2, "exactly one failed adoption then success");
    assert.equal(reloads, 0, "recovery on retry must never reload");
    assert.ok(
      logTexts().some((line) => line.includes("csrf token adopted")),
      "the adoption success must reach #log",
    );
  } finally {
    delete globalThis.window.location;
  }
});

test("csrf adoption falls back to reload only after three failures", async () => {
  resetToasts();
  let csrfCalls = 0;
  useFetch((url) => {
    if (String(url).includes("/api/session")) {
      return { status: 201, body: { extension: "1001", password: "pw" } };
    }
    csrfCalls += 1;
    return { status: 503 };
  });
  let reloads = 0;
  globalThis.window.location = {
    reload() {
      reloads += 1;
    },
  };
  try {
    await session.createSession("1001", "pw");
    assert.equal(csrfCalls, 3, "exactly three adoption attempts");
    assert.equal(reloads, 1, "the third failure falls back to one reload");
    assert.ok(
      logTexts().some((line) => line.includes("csrf adoption failed")),
      "the failure must reach #log before the reload",
    );
  } finally {
    delete globalThis.window.location;
  }
});
