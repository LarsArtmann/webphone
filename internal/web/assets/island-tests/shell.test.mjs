// shell.js under node:test — the error surfacing the Go asset tripwires
// can only grep for. htmx swaps NOTHING on error responses, so before the
// 3c handler a dead tab session made every tab click and form submit fail
// silently. These specs drive the real document-level listeners the way
// htmx fires them: responseError carries detail.xhr, sendError carries
// none.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
await import("../shell.js");

const toasts = () => doc.getElementById("toasts");
const fakeXhr = (status, headers = {}) => ({
  status,
  getResponseHeader: (name) => headers[name] ?? null,
});

test("a 401 from a dead tab session toasts what happened, without reloading", () => {
  doc.dispatch("htmx:responseError", { detail: { xhr: fakeXhr(401) } });
  const toast = toasts().children.at(-1);
  assert.equal(toast.className, "toast toast-error");
  assert.match(toast.textContent, /session ended/i);
  assert.match(toast.textContent, /calls keep working/i);
  assert.match(toast.textContent, /reload/i);
});

test("server-authored feedback (HX-Trigger) is never double-toasted", () => {
  const before = toasts().children.length;
  doc.dispatch("htmx:responseError", {
    detail: { xhr: fakeXhr(422, { "HX-Trigger": '{"showMessage":{}}' }) },
  });
  assert.equal(toasts().children.length, before);
});

test("the throttle collapses an error storm into one toast", () => {
  const before = toasts().children.length;
  for (let i = 0; i < 5; i++) {
    doc.dispatch("htmx:responseError", { detail: { xhr: fakeXhr(401) } });
  }
  doc.dispatch("htmx:responseError", { detail: { xhr: fakeXhr(502) } });
  doc.dispatch("htmx:sendError", {});
  assert.equal(toasts().children.length, before);
});

test("after the throttle window the next failure toasts again", () => {
  const realNow = Date.now;
  const base = realNow();
  try {
    Date.now = () => base + 10_000;
    doc.dispatch("htmx:sendError", {});
    const toast = toasts().children.at(-1);
    assert.equal(toast.className, "toast toast-error");
    assert.match(toast.textContent, /network request failed/i);
  } finally {
    Date.now = realNow;
  }
});

test("statuses without server feedback get an honest generic toast", () => {
  const realNow = Date.now;
  const base = realNow();
  try {
    Date.now = () => base + 30_000;
    doc.dispatch("htmx:responseError", { detail: { xhr: fakeXhr(429) } });
    assert.match(
      toasts().children.at(-1).textContent,
      /HTTP 429/,
    );
  } finally {
    Date.now = realNow;
  }
});
