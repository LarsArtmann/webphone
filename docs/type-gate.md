# The island type gate — how to add a file without breaking it

`tsc --checkJs --strict` over every served JavaScript file. Green since
2026-10-27 (680 errors retired), enforced by the `island-typecheck` flake
check, and pinned by `internal/web/assets/typecheck_pin_test.go`.

## The shape

```console
TSC=$(nix build --no-link --print-out-paths nixpkgs#typescript)/bin/tsc
$TSC -p . --noEmit                                    # the gate: exit 0
$TSC -p . --noEmit --listFiles | grep -c internal/web/assets   # 26+ served files
nix build .#checks.x86_64-linux.island-typecheck      # the same gate in CI shape
```

`tsconfig.json` includes `types/**/*.d.ts` and
`internal/web/assets/**/*.js`, and excludes exactly one thing: the vendored
sip.js bundle (`internal/web/assets/vendor/**`). **`tsc` from the devshell is
the only authority** — an editor's language server bundles a different
compiler and reports phantom errors (vtsls on TS 5.x once reported ~1,600
beside a green gate).

## Why there are no `// @ts-check` pragmas

`internal/web/assets/assets.go` embeds the tree with `//go:embed all:island`
and serves the bytes verbatim. A pragma is a served-byte change, and three
independent pin sites grep served bytes (Go tests, `scripts/webphone-smoke.py`,
the consuming stack's `tests/browser-e2e.py`). `checkJs` + `include` checks
the same files with zero byte changes. The pin test fails on a bare
`@ts-ignore` or any `@ts-nocheck`.

## The three annotation patterns that carry the tree

1. **An identity function for a heterogeneous dictionary.** `i18n.js` uses
   `const defineMessages = (table) => table;` so each language map gets a
   contextual signature. The key union stays inferable (a typo in `t("key")`
   is an error) without flattening the values into `Record<string, string>`,
   which would kill the formatter functions that take an argument.
2. **`hasMethod(el, name)` for duck-typed nodes.** The `node:test` DOM stubs
   are plain objects — some answer `matches()` but not `closest()`. A real
   `HTMLElement` answers both, so `if (el.closest)` reads as dead code to the
   checker while remaining load-bearing at runtime. One documented helper
   beats four casts (and `instanceof Element` would throw on the stubs).
3. **A bound non-null local before a closure.** Narrowing does not survive a
   callback: a nullable `getElementById` read inside a `.then()` needs
   `var player = /** @type {HTMLAudioElement} */ (audio);` (or a typed
   lookup helper such as `audioElement`/`canvasElement`) before the closure.

Typed lookup helpers live at the top of `shell.js`: `audioElement`,
`canvasElement`, `canvasIn`, `closestForm`, `textareaIn`, `hasMethod`,
`errorMessage`.

## Forbidden escapes

- `// @ts-ignore` — a bare suppression with no reason; the pin test fails.
- `// @ts-nocheck` — one whole untyped file; the pin test fails.

Preferred, in order: narrow the type properly → a typed helper → a
`@ts-expect-error` **with a written reason** on the next line.

## Adding a file

1. Write it under `internal/web/assets/` (the include glob picks it up — there
   is nothing to register, and a new file that is NOT covered fails
   `TestTypeGateCoversEveryAssetSource`).
2. If it is an island module, add it to `main.js`'s import closure: the pin
   test refuses a module outside it (`TestIslandModulesMatchImportClosure`).
3. If it introduces a DOM id, declare it in `types/island.d.ts` **and** in
   `docs/dom-contract.md` — `TestIslandDomIdsMirrorTheDomContract` refuses one
   without the other.
4. Run `node --check <file>` while editing (the pre-commit hook does this on
   staged files, ~40 ms) and `$TSC -p . --noEmit` before committing.

## The two escape hatches this gate deliberately does NOT have

- **No excluded red file.** Excluding a failing file is how a gate rots; the
  exclude list is pinned to exactly one vendored glob, and that glob must reach
  the bundle and nothing else.
- **No narrowed include.** Listing individual files works until the next asset
  is added and joins the program as an unannotated import;
  `TestTypeGateIncludeStaysAGlob` refuses any non-`**` entry under the asset
  tree.

## History

`docs/planning/2026-10-27_19-20_SUPERB-island-type-gate-execution-pareto-plan.md`
carries the execution record (§9 Errata), and
`docs/status/2026-10-27_22-00_island-type-gate-complete-status.md` the
before/after numbers.
