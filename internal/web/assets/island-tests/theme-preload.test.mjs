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
  const classes = new Set();
  globalThis.localStorage = storage;
  globalThis.document = {
    documentElement: {
      setAttribute: (name, value) => applied.push([name, value]),
      classList: {
        add: (name) => classes.add(name),
        contains: (name) => classes.has(name),
      },
    },
  };
  runInThisContext(source);
  return { applied, classes: [...classes] };
};

const localStorageReturning = (value) => ({ getItem: () => value });
const localStorageThrowing = () => ({
  getItem() {
    throw new Error("storage denied");
  },
});

test("forced themes apply before paint; anything else leaves the OS theme", () => {
  assert.deepEqual(preloadWith(localStorageReturning("light")).applied, [["data-theme", "light"]]);
  assert.deepEqual(preloadWith(localStorageReturning("dark")).applied, [["data-theme", "dark"]]);
});

test("auto, absent and denied storage never set data-theme", () => {
  assert.deepEqual(preloadWith(localStorageReturning("auto")).applied, []);
  assert.deepEqual(preloadWith(localStorageReturning(null)).applied, []);
  assert.deepEqual(preloadWith(localStorageThrowing()).applied, []);
});

test("a dismissed welcome collapses before first paint", () => {
  assert.deepEqual(
    preloadWith({ getItem: (key) => (key === "wp-welcome-dismissed" ? "1" : null) }).classes,
    ["wp-welcome-dismissed"],
  );
});

test("an undismissed or garbage welcome flag never sets the class", () => {
  assert.deepEqual(
    preloadWith({ getItem: (key) => (key === "wp-welcome-dismissed" ? "0" : null) }).classes,
    [],
  );
  assert.deepEqual(preloadWith(localStorageReturning(null)).classes, []);
  assert.deepEqual(preloadWith(localStorageThrowing()).classes, []);
});
