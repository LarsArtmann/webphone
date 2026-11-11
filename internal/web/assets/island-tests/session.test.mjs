// session.js under node:test — the server-session surface AFTER the
// auth-front-door train: the island no longer owns a login POST (the
// /login page's standalone module does that), so what remains here is
// the resume probe, the CSRF-rotation adoption ladder, the quiet
// sign-out, and the SSE feed's feedback map. A dead server session is
// deliberately non-fatal (the island keeps calling), but non-fatal must
// never mean invisible: the SSE feed reports one drop instead of a
// flap, and a genuinely dead feed is announced once.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
const { t } = await import("../island/app/i18n.js");
await import("../island/app/ui.js");
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

// fetch stub: adoptFreshCsrfToken and the probe are the only fetchers on
// this path. handler(url) → {status, body} or throws.
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

// The auth-front-door contract, pinned from the other side: the island
// exports NO login POST anymore — the extension+password submit lives on
// the /login page's standalone module. A regression that reintroduces an
// island-side login would split-brain the front door again.
test("the island no longer owns a login POST (the /login page does)", () => {
  assert.equal(session.createSession, undefined);
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

  // Recovery announces itself once (M17 J2), then the counter resets.
  doc.dispatch("htmx:sseOpen", {});
  assert.equal(lastToast().text, t("sseRestored"));
  doc.dispatch("htmx:sseError", {});
  doc.dispatch("htmx:sseError", {});
  assert.equal(toastsHost().children.length, count + 1, "a recovered feed resets the counter");
});

// M17 J2: the auto-recovery notice fires ONLY after an announced drop —
// a flap the user never heard about must not narrate its healing.
test("the SSE feed announces recovery only after an announced drop", () => {
  resetToasts();
  session.initSseLiveIndicator();
  doc.dispatch("htmx:sseOpen", {}); // deterministic counter start

  // Quiet flap: two failures (below the toast threshold), then recovery.
  doc.dispatch("htmx:sseError", {});
  doc.dispatch("htmx:sseError", {});
  doc.dispatch("htmx:sseOpen", {});
  assert.equal(lastToast(), null, "an unannounced drop must not announce recovery");

  // Announced drop: third failure toasts, and the recovery says so.
  doc.dispatch("htmx:sseError", {});
  doc.dispatch("htmx:sseError", {});
  doc.dispatch("htmx:sseError", {});
  assert.equal(lastToast().text, t("sseDropped"));
  doc.dispatch("htmx:sseOpen", {});
  assert.equal(lastToast().text, t("sseRestored"));
  assert.equal(lastToast().className, "toast toast-ok");

  // Recovery is one-shot: a later quiet cycle stays quiet.
  doc.dispatch("htmx:sseError", {});
  doc.dispatch("htmx:sseOpen", {});
  const count = toastsHost().children.length;
  assert.equal(lastToast().text, t("sseRestored"), "no second recovery toast");
  assert.equal(toastsHost().children.length, count);
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

test("fetchLiveSession resumes on a sip_credential-only answer", async () => {
  useFetch(() => ({
    status: 200,
    body: { extension: "1001", sip_credential: "derived-cred" },
  }));
  const data = await session.fetchLiveSession();
  assert.deepEqual(data, { extension: "1001", sip_credential: "derived-cred" });
  useFetch(() => ({ status: 200, body: { extension: "1001" } }));
  assert.equal(
    await session.fetchLiveSession(),
    null,
    "a 200 with NEITHER credential field must resume nothing",
  );
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
// Every login lane (the /login page's form POST, a passkey finish, the
// OIDC callback) rotates the CSRF token, so adoptServerSession must
// adopt the fresh one before anything else POSTs. A transient failure of
// GET /api/csrf retries with backoff (x3) and only the THIRD consecutive
// failure falls back to a page reload (the cookie session survives; the
// served page then carries a matching token again).

test("csrf adoption recovers on the second try without a reload", async () => {
  resetToasts();
  let csrfCalls = 0;
  useFetch((url) => {
    if (String(url).includes("/api/csrf")) {
      csrfCalls += 1;
      if (csrfCalls < 2) return { status: 503 };
      return { status: 200, body: { token: "fresh-token" } };
    }
    return { status: 200, body: {} };
  });
  let reloads = 0;
  globalThis.window.location = {
    reload() {
      reloads += 1;
    },
  };
  try {
    await session.adoptServerSession({});
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
  useFetch((url) => {
    if (String(url).includes("/api/csrf")) return { status: 503 };
    return { status: 200, body: {} };
  });
  let reloads = 0;
  globalThis.window.location = {
    reload() {
      reloads += 1;
    },
  };
  try {
    await session.adoptServerSession({});
    assert.equal(reloads, 1, "the third failure falls back to one reload");
    assert.ok(
      logTexts().some((line) => line.includes("csrf adoption failed")),
      "the failure must reach #log before the reload",
    );
  } finally {
    delete globalThis.window.location;
  }
});
