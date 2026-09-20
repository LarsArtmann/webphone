// i18n.js under node:test — the en/de dictionary parity the Go-side sync
// test enforces for the SERVER dictionaries, now enforced island-side too
// (AGENTS: add new keys to BOTH maps).
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();
const i18n = await import("../island/app/i18n.js");

test("island i18n dictionaries keep en/de keys in sync", () => {
  const en = Object.keys(i18n.I18N.en).sort();
  const de = Object.keys(i18n.I18N.de).sort();
  assert.deepEqual(de, en);
});

test("t() resolves en strings and falls back through en for de", () => {
  assert.equal(i18n.getLang(), "en");
  assert.equal(i18n.t("registered"), "registered");
  assert.equal(i18n.t("definitely-not-a-key"), undefined);
});

test("t() serves the German table after setLang", () => {
  i18n.setLang("de");
  assert.equal(i18n.getLang(), "de");
  assert.equal(i18n.t("registered"), "registriert");
  assert.equal(document.cookieSet, "wp-lang=de; path=/; samesite=strict; max-age=31536000");
  i18n.setLang("en");
});
