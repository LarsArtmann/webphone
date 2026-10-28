# Status report — the island type gate is now ENFORCED (T01–T18, partially)

> **SUPERSEDED 2026-10-27 21:30.** This snapshot was taken mid-campaign and its numbers are the
> 20:30 ones. The campaign finished the same evening: **0** type errors over **26 of 26** served
> files, `tsc -p . --noEmit` green, `nix flake check` green (all 32 checks, VM tests included),
> `tsconfig.all.json` deleted, and the debt list replaced by coverage pins. Read the completion
> record in
> [docs/planning/2026-10-27_19-20_SUPERB-island-type-gate-execution-pareto-plan.md](../planning/2026-10-27_19-20_SUPERB-island-type-gate-execution-pareto-plan.md)
> §9 — everything below stays as the point-in-time record, including its "not run" list.

**Snapshot: 2026-10-27 20:30 CET.** Working tree: `internal/web/assets/shell.js` mid-edit (this
session's), everything else committed. Local `HEAD` is `4fe8421`; **12 commits ahead of
`origin/main`, none pushed**, so CI has not seen any of this train (age at report time: 20 min,
below the repo's 60-min manual-push bar — the daemon's push leg is simply stalled).

Per the repo's status convention this report makes **no** deployment claims, states what was
measured locally, and names what was **not** run (§8).

---

## 0. The four numbers

| Metric                                                 | Before this session  | Now                      |
| ------------------------------------------------------ | -------------------- | ------------------------ |
| Type errors over the whole asset tree                  | **680**              | **79** (−601, −88 %)     |
| Asset files covered by a **green, enforced** type gate | **0** (red, unwired) | **25 of 26**             |
| Type-gate nix checks + pin assertions                  | 0 + 0                | **1 check, 5 pin tests** |
| Island behaviour tests                                 | 260 pass             | **260 pass** (unchanged) |
| `go test ./...`                                        | not run since edits  | **green**                |

The one remaining file is `shell.js` (79 errors, down from 280). It is the **single declared
debt** in `typecheckDebt`, and the pin test refuses to let that list grow.

---

## 1. What the gate actually is now

```
tsc -p .             → the GATE. Green. 25 asset files + types/**. Wired into nix.
tsc -p tsconfig.all.json → the MEASURING project. Red by design (79). Never gates.
```

Three artifacts, each verified by running it:

1. **`nix/checks.nix` → `island-typecheck`** — sandboxed, network-free, runs
   `tsc -p tsconfig.json --noEmit`, and writes `$out` listing **every file it actually covered**
   (25 lines today). A gate that silently stopped covering a file shows up in the build log
   instead of being trusted. **Proven to bite**: a one-character identifier typo in a checked
   module failed the check with 26 errors; reverting made it green.
2. **`internal/web/assets/typecheck_pin_test.go`** — five assertions, each **proven to bite**:
   | Assertion                                              | How it was proven                     |
   | ------------------------------------------------------ | ------------------------------------- |
   | `checkJs`/`strict`/`noEmit`/`allowJs` still true       | flipped `checkJs` to false → FAIL     |
   | every `include` entry resolves to ≥1 file              | ran before the include matched → FAIL |
   | the include's **import closure is closed**             | new assertion; see §4                 |
   | `typecheckDebt` == (asset tree − include)              | dropped a name → FAIL                 |
   | no bare `@ts-ignore`/`@ts-nocheck`                     | appended one → FAIL                   |
   | the nix check reads `tsconfig.json`, never `.all.json` | static check                          |
3. **`nix/devshell.nix`** — `typescript` + `nodejs` in the default shell, so the gate is one
   command for the next person.

---

## 2. Per-file ledger

| File                                        | Started |    Now | Gate                                                         |
| ------------------------------------------- | ------: | -----: | ------------------------------------------------------------ |
| `enroll/webauthn.js`                        |       4 |      0 | ✅ (first fix: my own `@type {const}` bug — not valid JSDoc) |
| `enroll/enroll.js`                          |      28 |      0 | ✅                                                           |
| `island/app/i18n.js`                        |      79 |      0 | ✅                                                           |
| `island/app/typeahead.js`                   |      30 |      0 | ✅                                                           |
| `island/app/transcribe.js`                  |      41 |      0 | ✅                                                           |
| `island/app/panels.js`                      |      46 |      0 | ✅                                                           |
| `island/app/calls.js`                       |     126 |      0 | ✅                                                           |
| `island/app/connection.js`                  |      34 |      0 | ✅                                                           |
| `island/app/main.js`                        |      12 |      0 | ✅                                                           |
| `island/app/shortcuts.js`                   |       0 |      0 | ✅ (joined once its closure went clean)                      |
| the 15 files already clean at session start |       0 |      0 | ✅                                                           |
| **`shell.js`**                              | **280** | **79** | ⛔ only debt                                                 |

Error mix left (all in `shell.js`): TS2339 ×36 · TS18047 ×14 · TS2740 ×10 · TS2454 ×5 ·
TS2345 ×5 · TS2774 ×4 · TS7006 ×2 · tail. Every remaining one is a **DOM element subtype**
problem (a `<canvas>`/`<audio>`/`<input>` read through an `HTMLElement`) or a `var` narrowing
that a closure drops — no logic left to discover.

---

## 3. Ten real defects the gate found (not cosmetics)

Each was a compile error _because the declared type was true and the code was not_:

1. **`rejectCause` was declared `string`, assigned a SIP status code number.** The failure-reason
   switch (`486`/`404`/`403`/`480`/`487`/`408`) could therefore never match, and the operator's
   ended card fell back to the generic sentence. `calls.js`.
2. **A declined WebAuthn prompt crashed the enrol page.** `navigator.credentials.create()`
   returns `Credential | null`; the code dereferenced it unconditionally. Now both the throw and
   the silent null land on the same honest "cancelled — ask for a new link" copy.
3. **History rows could print the word "undefined".** The CDR fields the `/phone-api` proxy
   forwards are all optional (`billsec`, `caller_id_number`, `destination_number`); each now has
   a fallback the operator can read (`—`, `0:00`). `panels.js`.
4. **`serializeCredential`'s `kind` parameter was documented as `"create" | "get"`** while every
   caller passes `"login"` / `"registration"` — my own doc defect from the previous session, now
   naming the real vocabulary.
5. **The SSO error card built message keys by string concatenation** (`"ssoError" + code`). The
   server's eight `?sso_error` codes are now a static map: a typo is a compile error, an unknown
   code still falls back to the generic sentence.
6. **`document.activeElement.isContentEditable`** was read off a bare `Element` (which has no
   such property) in the incoming-call focus guard — a `TypeError` for any non-HTMLElement
   active element. `connection.js`.
7. **`event.target.closest(...)` on a bare `EventTarget`**, 62 sites in `shell.js`. The old
   `!element ||` guards were reaching for a narrowing that did not exist; a document-dispatched
   event would have thrown instead of falling through. Now one typed `eventElement()` helper.
8. **Nullable session ids passed straight into `sessions.get()`** and into `encodeURIComponent`
   (`state.focusedId`, CDR call ids) — every one now guards or falls back.
9. **A dead disposition table lookup**: the operator-confirm chips looked message keys up through
   a bare `string`, one typo away from rendering a raw key into the UI.
10. **`t()`'s key union was `string`.** Typing it surfaced three genuine misuse sites the moment
    it landed (a dynamic tuple array, a callback passed as `t`, a raw `string` from a `<select>`).

---

## 4. The plan's §0 was wrong — measured, corrected, documented

The 19:20 plan chose "checked-subset `include` + a pin naming the debt", justified by:
_"TypeScript still infers types through an unchecked module into a checked one"_ — i.e. the
subset would not be blind.

**That premise is false in the direction that matters.** TypeScript **fully checks** every module
an included module imports, and reports _that module's_ errors. Measured proof:

```
tsconfig include = types/** + island/app/ui.js
tsc -p … --noEmit  →  79 errors, ALL in i18n.js
```

`ui.js` imports `i18n.js`, so a red `i18n.js` made **every consumer of it un-includable** — which
was nearly the whole island. A subset include is green only when the subset's **import closure**
is green.

**Two consequences, both applied:**

- **`i18n.js` moved from task T11 to the first annotation task.** It is the universal dependency;
  typing it unblocked 17 files at once. (This is why the gate went from 0 to 18 checked files in
  one step rather than one file at a time.)
- **The remaining ordering must follow the import closure, not the error count.** The pin test now
  asserts the closure property explicitly, so the next person cannot list a file and be surprised.

This is recorded as an erratum for the executing plan; the plan's §9 errata table is the right
home for it and T23 will carry the line.

---

## 5. Process findings (these cost more than the annotations did)

### 5.1 The daemon reverted my working tree **twice** by committing to a detached HEAD

Reflog evidence, both occurrences:

```
055030c HEAD@{0}: checkout: moving from 64dc66a… to main      ← my work is in 64dc66a
64dc66a HEAD@{1}: commit: chore: auto-commit 1 changed file(s) (heuristic)
3b2391e HEAD@{2}: commit: chore: auto-commit 4 changed file(s) (heuristic)
055030c HEAD@{3}: checkout: moving from main to HEAD           ← detached here
```

`checkout main → HEAD`, commit, `checkout <sha> → main`: my uncommitted edits were committed onto
a detached HEAD and then **silently reverted out of the working tree** by the return to `main`.

**Recovery recipe** (worked twice, no data lost):

```console
git reflog -8                                   # find the detached pair
git show <sha>:<path> > <path>                  # pull the file back out
git log -1 --format=%H -- <path>                # which commit last touched it
```

**Mitigation adopted:** stage + commit **per task**, in the same tool call as the edit's
verification. Everything after the second incident survived.

### 5.2 The daemon committed a **broken intermediate** state

A blanket regex I used to convert `event.target` → `eventElement(event)` also rewrote the helper's
own body, producing **infinite self-recursion**:

```js
function eventElement(event) { var target = eventElement(event); … }
```

The daemon committed that state twice (`0f3d5e1`, `d7cbb64` — verified with
`git show <sha>:internal/web/assets/shell.js | grep -c`). The island suite caught it within
minutes (`ReferenceError` / 45 failures), but **nothing in the commit path gates `shell.js`**:
`island-js` and the type gate run in `nix flake check`, not at commit time. A pre-commit hook that
ran `node --test island-tests/` would have caught it before it entered history.

### 5.3 The served-byte pins fired for real, and were right to

`TestShellJSHandlesReloadButtons` and `TestShellJSHandlesTranscribeButtons` failed after the
helper refactor **although the handlers still work** — they grep served bytes for
`closest("[data-x]")`. This is the third pin site AGENTS warns about, doing its job.

The pin update had to solve a second-order problem: **prettier wraps these calls across lines**, so
a literal-string pin breaks on formatting, and a pin that breaks on formatting is a pin the next
person deletes. Fix: the pins now (a) name the whole helper call, not a bare selector (which would
also match a comment), and (b) compare **whitespace-free** text via a new `withoutSpace` helper.

### 5.4 The editor LSP is not an authority (reconfirmed)

`vtsls` (TypeScript 5.x) reported ~1,577 phantom project errors and stale per-file line numbers
throughout; `tsc` 7.0.2 from the devshell was the only source of truth. Documenting this in
AGENTS is T21 and is now overdue — it cost attention every single file.

---

## 6. What is left in `shell.js` (79 errors, grouped by fix shape)

| Cluster                                                                    | ~Count | Fix                                                                                              |
| -------------------------------------------------------------------------- | -----: | ------------------------------------------------------------------------------------------------ |
| voicemail media/canvas (`vm*`, `audio`, `canvas`, `peaks`)                 |     30 | type the id lookups as `HTMLAudioElement`/`HTMLCanvasElement`; `vmBuildWave(uuid, audio)` params |
| command palette (`overlay`, `overlayInput`, `overlayList`, `overlayItems`) |     13 | bind locals after `buildOverlay()`, one `run` field on the item shape                            |
| composer drafts + segcount (`.value`, `.requestSubmit`, `.files`, `.type`) |     20 | `HTMLInputElement`/`HTMLTextAreaElement`/`HTMLFormElement` casts at the lookup                   |
| leftover vacuous `!x.closest` guards (TS2774)                              |      4 | delete — the helper made them impossible                                                         |
| agent pill (`IslandAgentModeState` typedef)                                |      3 | add the typedef, drop the `.type` on a created button                                            |
| `window.__wpClock` test seam                                               |      3 | declare it in `types/globals.d.ts`                                                               |
| `TS2454` "used before assigned" on 5 test-seam exports                     |      5 | declare them `                                                                                   |
| misc (`errorMessage`, `.hidden`, `.click`)                                 |      4 | the established one-home patterns                                                                |

None of these is a design question. It is one focused sitting.

---

## 7. Plan tasks not started

| Task                                                                                     | State                                                                                             |
| ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| **T19** flip include to the whole tree, delete the debt list + `tsconfig.all.json`       | blocked on `shell.js` reaching 0 — nothing else                                                   |
| **T20** `errorMessage` sweep + micro-test                                                | partially done inline (transcribe/panels/calls/connection/main adopt it); the sweep + test remain |
| **T21** AGENTS.md stance (377-line cap)                                                  | not started — and now the single most valuable doc task (§5.4)                                    |
| **T22** README dev section + CHANGELOG line + `.buildflow.yml` note                      | not started                                                                                       |
| **T23** ROADMAP triggers + the 16:04 plan's three defects **+ this report's §4 erratum** | not started                                                                                       |
| **T24** typed-rejection design                                                           | not started                                                                                       |
| **T25** island-tests typing verdict                                                      | not started                                                                                       |
| **T26** full verification run                                                            | **partially** run, see §8                                                                         |
| **T27** 30-day retrofit report                                                           | not started (nothing to retrofit yet)                                                             |
| **T28–T29** android KMP                                                                  | Tier 0, untouched, by design                                                                      |

---

## 8. Unverified — stated plainly

| Claim                                              | Status                                                 |
| -------------------------------------------------- | ------------------------------------------------------ |
| `tsc -p .` (the gate) green                        | **verified**, exit 0, repeatedly                       |
| `nix build .#checks.x86_64-linux.island-typecheck` | **verified** green, 25-file artefact                   |
| `nix flake check` (the whole suite)                | **NOT RUN** this session                               |
| `buildflow`                                        | **NOT RUN** this session                               |
| `python3 scripts/webphone-smoke.py`                | **NOT RUN**                                            |
| `scripts/ui-capture.py` (24-shot visual harness)   | **NOT RUN**                                            |
| `go test -count=1 ./...`                           | **verified** green (after fixing the two pins in §5.3) |
| island `node --test` (260 specs)                   | **verified** green after every file                    |
| push / CI verdict                                  | **none** — 12 commits unpushed (§0)                    |

`nix flake check` and `buildflow` are the two that could still surprise: treefmt has already
reformatted `shell.js` once mid-edit, and BuildFlow's oxfmt owns everything outside
`island/**` + `shell.js` — including `tsconfig.json`, `tsconfig.all.json` and `types/*.d.ts`,
which have only been oxfmt-checked by hand.

---

## 9. Next actions, in order

1. **Finish `shell.js`** with the §6 clusters (one sitting), then add it to `tsconfig.json` and
   delete the last debt line. That is **T19** in a single move: flip the include to the whole
   tree, delete `tsconfig.all.json`, update its two doc references, re-run the nix check.
2. **`nix flake check`** end-to-end (T26) — the first time the new check runs inside the real
   suite.
3. **`buildflow`** (T26) — expect formatter findings on the new JSON/`.d.ts` files.
4. **T20** `errorMessage` sweep + micro-test (the last `err.message` sites are in `shell.js`).
5. **T21** AGENTS stance: the `tsc`-is-the-authority rule, the served-byte pin whitelist, the
   subset/import-closure rule, and "commit per task, the daemon detaches HEAD".
6. **T22/T23** docs + the erratum line from §4.
7. Push once the 60-min anchor passes (or immediately if the owner wants CI on this train).

---

## 10. Honest self-assessment

**Good:** the gate is real, not theatre — it is green, it is sandboxed, it is pinned against its
own weakening, and it was proven to fail on a typo before anyone trusted it. 601 errors retired in
~70 minutes with zero behaviour regressions (260 island specs + the full Go suite green). The
defects it surfaced (a failure-reason switch that could never match, a null credential, three
"undefined" renders) are exactly the class the owner asked for when he said an error without a
visible surface is a bug.

**Not good enough yet:** `shell.js` is 79 errors from done, the gate is not yet in
`nix flake check`, and two of the three whole-repo gates (`buildflow`, the smoke script) have not
run — so "enforced" is proved locally but not yet proved by the repo's own suite. Nothing is
pushed, so none of this has been seen by CI.

**The one habit worth changing:** the daemon detaches HEAD and reverts the tree. Commit per task,
always, in the same breath as the verification command.

---

_Written 2026-10-27 20:30 CET at `HEAD 4fe8421` + one uncommitted file (`shell.js`, mid-annotation).
Every number above was produced by a command whose output is quoted or reproducible from §8._
