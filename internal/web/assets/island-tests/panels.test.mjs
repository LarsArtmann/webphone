// panels.js under node:test — the contacts live-update nudge. Every
// contacts mutation server-side publishes the payload-less "contacts"
// SSE event; the island re-fetches its dropdown list through its own
// session (mirroring the voicemail nudge pattern). The listener keys on
// detail.type because the htmx sse extension fires one bubbling
// htmx:sseMessage per delivered event with the raw MessageEvent as
// detail.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

const doc = installBrowserGlobals();
// phoneApi on (and one shared contact) so renderContacts passes its
// gate: without a config the dropdown legitimately stays hidden and
// only the fetch behavior is observable.
globalThis.window.PBX_CONFIG = {
  phoneApi: true,
  contacts: [{ name: "Shared Desk", number: "1099" }],
};

let fetchCalls = [];
globalThis.fetch = async (input) => {
  fetchCalls.push(String(input));
  return {
    ok: true,
    json: async () => ({
      personal: [{ id: "c1", name: "Nudged", number: "1002" }],
      shared: [],
    }),
  };
};

await import("../island/app/panels.js?test=contacts-nudge");

const contactsFetches = () =>
  fetchCalls.filter((url) => url.includes("/api/contacts")).length;

test("the contacts nudge re-fetches the island dropdown", async () => {
  fetchCalls = [];
  doc.dispatch("htmx:sseMessage", { detail: { type: "contacts" } });
  await new Promise((resolve) => setTimeout(resolve, 20));
  assert.equal(
    contactsFetches(),
    1,
    `expected exactly one /api/contacts fetch, got ${fetchCalls.join(", ")}`,
  );
});

test("other SSE deliveries leave the dropdown alone", async () => {
  fetchCalls = [];
  doc.dispatch("htmx:sseMessage", { detail: { type: "fax" } });
  doc.dispatch("htmx:sseMessage", { detail: {} });
  await new Promise((resolve) => setTimeout(resolve, 20));
  assert.equal(contactsFetches(), 0);
});

test("a nudged list renders the fetched rows", async () => {
  doc.dispatch("htmx:sseMessage", { detail: { type: "contacts" } });
  await new Promise((resolve) => setTimeout(resolve, 20));
  const wrap = doc.getElementById("contacts-wrap");
  assert.equal(wrap.hidden, false, "contacts wrap must be visible with rows");
  const list = doc.getElementById("contacts-list");
  const texts = list.children.flatMap((li) =>
    li.children.map((span) => span.textContent),
  );
  assert.ok(
    texts.includes("Nudged"),
    `fetched row must render, got: ${texts.join(", ")}`,
  );
});
