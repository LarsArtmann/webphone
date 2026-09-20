// Browser-global stubs so the island modules import under node:test.
// ui.js resolves every DOM-contract id at import time; i18n.js reads
// localStorage + navigator.language at import time. Keep the stubs as
// small as the tests demand — they are not a DOM implementation.
export function installBrowserGlobals() {
  const registry = new Map();
  const docListeners = new Map();
  const makeEl = (id) => ({
    id,
    textContent: "",
    className: "",
    hidden: true,
    placeholder: "",
    title: "",
    lang: "",
    value: "",
    children: [],
    listeners: {},
    parent: null,
    get firstChild() {
      return this.children[0] ?? null;
    },
    append(child) {
      child.parent = this;
      this.children.push(child);
    },
    remove() {
      if (this.parent) {
        this.parent.children = this.parent.children.filter((c) => c !== this);
        this.parent = null;
      }
    },
    addEventListener(type, fn) {
      (this.listeners[type] ??= []).push(fn);
    },
    classList: {
      add() {},
      remove() {},
      toggle() {},
      contains() {
        return false;
      },
    },
    style: {},
    dataset: {},
    setAttribute() {},
    removeAttribute() {},
    getAttribute: () => null,
  });
  globalThis.document = {
    getElementById: (id) => {
      if (!registry.has(id)) registry.set(id, makeEl(id));
      return registry.get(id);
    },
    createElement: () => makeEl(""),
    querySelectorAll: () => [],
    querySelector: () => null,
    documentElement: makeEl("html"),
    addEventListener(type, fn) {
      if (!docListeners.has(type)) docListeners.set(type, []);
      docListeners.get(type).push(fn);
    },
    // dispatch hands a plain event object to every listener registered
    // for the type; tests pass target/detail through props.
    dispatch(type, props = {}) {
      const event = { target: null, detail: {}, preventDefault() {}, ...props };
      for (const fn of docListeners.get(type) ?? []) fn(event);
      return event;
    },
    cookieSet: "",
    set cookie(v) {
      this.cookieSet = v;
    },
    get cookie() {
      return this.cookieSet;
    },
  };
  globalThis.window = {};
  globalThis.location = {
    hostname: "pbx.example.org",
    host: "pbx.example.org",
    reload() {
      throw new Error("location.reload is banned in tests; assert the toast instead");
    },
  };
  // node >= 21 ships a getter-only global navigator — defineProperty, not
  // assignment.
  Object.defineProperty(globalThis, "navigator", {
    value: { language: "en-US" },
    configurable: true,
  });
  globalThis.localStorage = {
    store: new Map(),
    getItem(k) {
      return this.store.has(k) ? this.store.get(k) : null;
    },
    setItem(k, v) {
      this.store.set(k, String(v));
    },
    removeItem(k) {
      this.store.delete(k);
    },
  };
  return globalThis.document;
}
