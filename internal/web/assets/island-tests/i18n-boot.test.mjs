// The boot language order (wp-lang cookie → localStorage → navigator) is
// the split-brain fix: the server renders the shell, tabs and <html lang>
// from the SAME cookie (pages.go), so a cookie≠navigator user must boot
// the island in the cookie's language — one language everywhere, screen
// readers announcing with the right phonemes.
import test from "node:test";
import assert from "node:assert/strict";

import { installBrowserGlobals } from "./helpers.mjs";

installBrowserGlobals();

const loadI18n = (tag) => import(`../island/app/i18n.js?case=${tag}`);

test("a wp-lang cookie boots the island in its language", async () => {
  document.cookie = "wp-lang=de; path=/; samesite=strict";
  localStorage.store.delete("pbx-lang");
  const i18n = await loadI18n("cookie-de");
  assert.equal(i18n.getLang(), "de");
  assert.equal(i18n.t("registered"), "registriert");
});

test("the cookie beats localStorage when they disagree", async () => {
  document.cookie = "wp-lang=de; path=/; samesite=strict";
  localStorage.setItem("pbx-lang", "en");
  const i18n = await loadI18n("cookie-wins");
  assert.equal(i18n.getLang(), "de");
});

test("localStorage still leads when no cookie exists", async () => {
  document.cookie = "";
  localStorage.setItem("pbx-lang", "de");
  const i18n = await loadI18n("ls-de");
  assert.equal(i18n.getLang(), "de");
});

test("navigator is the last resort and applyI18n pins html lang to the boot lang", async () => {
  document.cookie = "";
  localStorage.store.delete("pbx-lang");
  const i18n = await loadI18n("navigator-en");
  assert.equal(i18n.getLang(), "en", "en-US navigator → en");
  i18n.applyI18n();
  assert.equal(document.documentElement.lang, "en");

  document.cookie = "wp-lang=de; path=/";
  const german = await loadI18n("html-lang-de");
  german.applyI18n();
  assert.equal(
    document.documentElement.lang,
    "de",
    "a cookie=de boot writes the same html lang the server rendered",
  );
});
