import { readFileSync } from "node:fs";
import { runInThisContext } from "node:vm";
import test from "node:test";
import assert from "node:assert/strict";

// theme-preload.js is a plain IIFE (no imports/exports): running its
// source through runInThisContext evaluates it against these globals,
// exactly like a browser script tag — and unlike import(), it can be
// re-run per case without module-cache games.
const source = readFileSync(new URL("../theme-preload.js", import.meta.url), "utf8");
const preloadWith = (storage) => {
  const applied = [];
  globalThis.localStorage = storage;
  globalThis.document = {
    documentElement: {
      setAttribute: (name, value) => applied.push([name, value]),
    },
  };
  runInThisContext(source);
  return applied;
};

const localStorageReturning = (value) => ({ getItem: () => value });
const localStorageThrowing = () => ({
  getItem() {
    throw new Error("storage denied");
  },
});

test("forced themes apply before paint; anything else leaves the OS theme", () => {
  assert.deepEqual(preloadWith(localStorageReturning("light")), [["data-theme", "light"]]);
  assert.deepEqual(preloadWith(localStorageReturning("dark")), [["data-theme", "dark"]]);
});

test("auto, absent and denied storage never set data-theme", () => {
  assert.deepEqual(preloadWith(localStorageReturning("auto")), []);
  assert.deepEqual(preloadWith(localStorageReturning(null)), []);
  assert.deepEqual(preloadWith(localStorageThrowing()), []);
});
