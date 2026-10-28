// The line switcher (D1) under node:test: it renders ONLY for a mapped
// multi-line user, POSTs the switch with the CSRF header, reloads the
// page on success (the boot probe owns every re-registration step),
// announces and resets on failure, and never switches under a live call
// (the reload would tear the call down).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
const { t } = await import("../island/app/i18n.js");
await import("../island/app/ui.js");
const session = await import("../island/app/session.js");
const { sessions } = await import("../island/app/state.js");

// The whoami container: the stub's document.querySelector answers null,
// so the test routes it at a fake .whoami element explicitly.
const whoami = doc.createElement("div");
whoami.className = "whoami";
doc.querySelector = (selector) => (selector === ".whoami" ? whoami : null);

// location.reload throws in the stub by design; the switcher's happy
// path calls it, so the test replaces it with a spy.
let reloads = 0;
globalThis.location.reload = () => {
  reloads += 1;
};
globalThis.window.location.reload = globalThis.location.reload;

const toastsHost = () => doc.getElementById("toasts");
const lastToast = () => {
  const el = toastsHost().children.at(-1);
  return el ? el.className : null;
};
const switcher = () => whoami.children.find((kid) => kid.id === "wp-ext-switch");

function useFetch(handler) {
  globalThis.fetch = async (url, opts) => handler(url, opts);
}

function reset() {
  for (const kid of [...whoami.children]) kid.remove();
  sessions.clear();
  reloads = 0;
  while (toastsHost().firstChild) toastsHost().firstChild.remove();
}

const multi = { extension: "1000", extensions: ["1000", "1001", "1002"] };

test("renders nothing without a multi-line mapped identity", () => {
  reset();
  session.renderLineSwitcher({ extension: "1000", extensions: ["1000"] });
  assert.equal(switcher(), undefined, "single line: no switcher");
  session.renderLineSwitcher({ extension: "1000" });
  assert.equal(switcher(), undefined, "no extensions list: no switcher");
});

test("renders one option per mapped line, current selected", () => {
  reset();
  session.renderLineSwitcher(multi);
  const select = switcher();
  assert.ok(select, "multi-line identity renders the switcher");
  assert.deepEqual(
    select.children.map((option) => option.value),
    ["1000", "1001", "1002"],
  );
  assert.equal(select.children[0].selected, true);
  assert.equal(select.attrs["aria-label"], t("lineSwitchLabel"));
});

test("switching POSTs the target with the CSRF header, then reloads", async () => {
  reset();
  session.renderLineSwitcher(multi);
  let seen = null;
  useFetch((url, opts) => {
    seen = { url, opts };
    return { ok: true, status: 200, json: async () => ({}) };
  });
  const select = switcher();
  select.value = "1001";
  for (const listener of select.listeners.change) listener();
  await new Promise((resolve) => setTimeout(resolve, 0));
  assert.equal(seen.url, "/api/session/switch");
  assert.ok("X-CSRF-Token" in seen.opts.headers, "the switch rides the CSRF header");
  assert.equal(JSON.parse(seen.opts.body).extension, "1001");
  assert.equal(reloads, 1, "success reloads so the boot probe re-registers");
});

test("a rejected switch announces, resets the select, never reloads", async () => {
  reset();
  session.renderLineSwitcher(multi);
  useFetch(() => ({ ok: false, status: 403 }));
  const select = switcher();
  select.value = "1002";
  for (const listener of select.listeners.change) listener();
  await new Promise((resolve) => setTimeout(resolve, 0));
  assert.equal(reloads, 0);
  assert.equal(select.value, "1000", "the select returns to the live line");
  assert.equal(lastToast(), "toast toast-warn");
});

test("a live call blocks the switch before any POST", async () => {
  reset();
  session.renderLineSwitcher(multi);
  let fetched = false;
  useFetch(() => {
    fetched = true;
    return { ok: true, status: 200, json: async () => ({}) };
  });
  sessions.set("call-1", { session: { state: "Established" }, held: false });
  const select = switcher();
  select.value = "1001";
  for (const listener of select.listeners.change) listener();
  await new Promise((resolve) => setTimeout(resolve, 0));
  assert.equal(fetched, false, "no switch POST under a live call");
  assert.equal(reloads, 0);
  assert.equal(select.value, "1000");
  assert.equal(lastToast(), "toast toast-warn");
});
