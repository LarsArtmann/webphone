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

test("the SSE feed toasts once after three consecutive failures", () => {
  resetToasts();
  session.initSseLiveIndicator();

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
  assert.equal(
    toastsHost().children.length,
    count,
    "a recovered feed resets the counter",
  );
});
