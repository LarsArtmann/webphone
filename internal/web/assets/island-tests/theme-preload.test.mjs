import { test } from "node:test";
import assert from "node:assert/strict";

// theme-preload.js is an IIFE applying the stored theme before first
// paint. Node caches modules by URL, so each case imports with a fresh
// query: the side effect reruns against that case's storage stub.
const preloadWith = async (storage) => {
  const applied = [];
  globalThis.localStorage = storage;
  globalThis.document = {
    documentElement: {
      setAttribute: (name, value) => applied.push([name, value]),
    },
  };
  await import(`../theme-preload.js?case=${crypto.randomUUID()}`);
  return applied;
};

const localStorageReturning = (value) => ({ getItem: () => value });
const localStorageThrowing = () => ({
  getItem() {
    throw new Error("storage denied");
  },
});

test("forced themes apply before paint; anything else leaves the OS theme", async () => {
  assert.deepEqual(await preloadWith(localStorageReturning("light")), [
    ["data-theme", "light"],
  ]);
  assert.deepEqual(await preloadWith(localStorageReturning("dark")), [
    ["data-theme", "dark"],
  ]);
});

test("auto, absent and denied storage never set data-theme", async () => {
  assert.deepEqual(await preloadWith(localStorageReturning("auto")), []);
  assert.deepEqual(await preloadWith(localStorageReturning(null)), []);
  assert.deepEqual(await preloadWith(localStorageThrowing()), []);
});
