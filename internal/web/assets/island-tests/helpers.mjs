// Browser-global stubs so the island modules import under node:test.
// ui.js resolves every DOM-contract id at import time; i18n.js reads
// localStorage + navigator.language at import time. Keep the stubs as
// small as the tests demand — they are not a DOM implementation.
export function installBrowserGlobals() {
  const registry = new Map();
  const docListeners = new Map();
  // matchesSelector is the stub's selector interpreter: an exact
  // .selector convention hit, or a className match against a class
  // selector (".wp-chip" matches className "wp-chip", including one
  // class of a space-separated list).
  const matchesSelector = (node, selector) =>
    node.selector === selector ||
    node.className === selector ||
    (selector.startsWith(".") && node.className.split(" ").includes(selector.slice(1)));
  const makeEl = (id) => ({
    id,
    textContent: "",
    className: "",
    hidden: true,
    placeholder: "",
    title: "",
    lang: "",
    value: "",
    type: "",
    children: [],
    listeners: {},
    parent: null,
    // selector is the stub's stand-in for a CSS selector match: closest
    // and querySelector compare it (or an exact className) instead of
    // interpreting selectors. Tests set .selector on the fakes they
    // route through delegated listeners.
    selector: null,
    closest(selector) {
      let node = this;
      while (node) {
        if (matchesSelector(node, selector)) return node;
        node = node.parent;
      }
      return null;
    },
    querySelector(selector) {
      const walk = (node) => {
        for (const kid of node.children) {
          if (matchesSelector(kid, selector)) return kid;
          const found = walk(kid);
          if (found) return found;
        }
        return null;
      };
      return walk(this);
    },
    get firstChild() {
      return this.children[0] ?? null;
    },
    append(...kids) {
      for (const kid of kids) {
        kid.parent = this;
        this.children.push(kid);
      }
    },
    prepend(child) {
      child.parent = this;
      this.children.unshift(child);
    },
    replaceChildren(...kids) {
      for (const kid of kids) kid.parent = this;
      this.children = [...kids];
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
    attrs: {},
    setAttribute(k, v) {
      this.attrs[k] = String(v);
    },
    removeAttribute(k) {
      delete this.attrs[k];
    },
    getAttribute(k) {
      return k in this.attrs ? this.attrs[k] : null;
    },
    hasAttribute(k) {
      return k in this.attrs;
    },
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
    body: makeEl("body"),
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
    // dispatchEvent is the REAL DOM API the island code calls (session.js
    // dispatches CustomEvents); it forwards the event object itself.
    dispatchEvent(event) {
      for (const fn of docListeners.get(event.type) ?? []) fn(event);
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
  // Older node lacks CustomEvent; the island dispatches wp:* events with it.
  globalThis.CustomEvent =
    globalThis.CustomEvent ||
    class CustomEvent {
      constructor(type, opts) {
        this.type = type;
        this.detail = opts?.detail;
      }
    };
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
  // Attachment-chip removal rebuilds FileLists through DataTransfer;
  // node ships none, so here is the minimum surface shell.js touches.
  globalThis.DataTransfer =
    globalThis.DataTransfer ||
    class DataTransfer {
      constructor() {
        this.files = [];
        this.items = { add: (file) => this.files.push(file) };
      }
    };
  return globalThis.document;
}
