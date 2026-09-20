# Island test harness (node:test + minimal DOM stubs)

Run: `nix run nixpkgs#nodejs -- --test --test-force-exit 'internal/web/assets/island-tests/*.test.mjs'`
(also the `island-js` flake check). Tests live BESIDE the served tree,
never inside it — `assets.go` embeds `island/` verbatim, and the stack's
browser E2E greps the served sources; test files must not leak into the
binary.

## Policy (T19, 2026-09-20)

- **Black-box behavior specs.** Each spec drives the same document-level
  listeners a browser event would hit, then asserts on the DOM (toasts
  in `#toasts`, `#log` entries). Grep-only assertions belong in the Go
  asset tripwires, not here.
- **`helpers.mjs` stubs grow only as the subjects demand.** They record
  attributes, children, and listeners; they are NOT a DOM implementation.
  If a spec needs a new browser API, first check the behavior is real
  (not an accident of the stub), then add the smallest honest stub.
- **`location.reload` throws** by design: the island never unloads
  (AGENTS invariant), so any test path reaching reload is a bug in the
  module or the spec — fail loudly.
- **Clock injection:** `shell.js`'s throttle reads
  `window.__wpClock.now()` when present, else `Date.now`. Specs install
  a fake through the seam instead of monkey-patching `Date.now`.
- **Module isolation is per FILE.** Each `.test.mjs` gets a fresh module
  registry per process; within a file, module state (e.g. the SSE
  failure counter in `session.js`) is shared — reset it explicitly at
  spec start when order could matter.

## main.js: deliberately NOT imported (decision recorded)

`main.js` is the composition root: its top level binds every UI element,
boots the SIP config, notifications, shortcuts, and the SSE indicator.
An import harness would need faithful stubs for `window.SIP`,
`Notification`, media elements, and every contract id — maintenance
weighing more than the wiring bugs it could catch, while the real
composition path is exercised end-to-end by the stack's browser E2E
(registration, call, transfer, DTMF, reconnect) and the served-DOM
contract test. Composition-root bugs surface there with a full browser;
unit-stubbing this file would mostly re-test the stubs. Revisit only if
`main.js` grows REAL logic (state machines, parsing) rather than wiring.
