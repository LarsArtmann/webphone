# Status: daemon-damage repair + push-train dedup (2026-10-24)

**Snapshot-rot header:** all claims verified as of **2026-10-24 07:28 CEST (+0200)**,
HEAD `b7e6efb`, working tree clean. Push and CI statements rot fastest; they are
timestamped, not presented as timeless.

**Session scope (what this report covers):** one session that started from "fix it
all properly" with no named defect. The work found and repaired damage the
auto-commit daemon had committed, then made the 2026-10-23 push train pass the
clone gate. Nothing outside this session's observations is researched here.

**Final state in one paragraph:** HEAD `b7e6efb` contains the repair and the dedup
train; remote `origin/main` is still `ff185ae` (3 commits unpushed, push leg idle
for about 1 h 55 min); local equivalents of every CI job are green, including the
CI-pinned `art-dupl` check at **0 new clone groups**; CI at `ff185ae` is green
except the clone gate, which is expected because that commit predates the dedup
fix.

---

## a) FULLY DONE

### 1. Restored the deleted `.buildflow.yml`

- The daemon commit `49e9495` deleted the entire 150-line quality-gate policy
  file (baseline pointer, tool paths, oxfmt/treefmt ownership split, skip
  rationales).
- Restored byte-identically from `4f9a14b` (`git show 4f9a14b:.buildflow.yml`);
  verified `git diff 4f9a14b HEAD -- .buildflow.yml` is empty and the file is
  150 lines at HEAD.
- `scripts/buildflow.sh` runs successfully against the restored config
  (passed with warnings only; see d).

### 2. Reverted the daemon's formatter drift on prettier-owned files

- `49e9495` reflowed 20 `internal/web/assets/island/**` + `shell.js` files to
  oxfmt's ~100-column style (joined lines that prettier splits at 80).
- Ran `nix fmt` (treefmt, prettier): 20 files changed back; verified
  `git diff 4f9a14b HEAD -- internal/web/assets` is empty, so the restoration is
  byte-identical to the pre-daemon state.
- `checks.x86_64-linux.format` passes; `island-js` passes 228/228.

### 3. Fixed the daemon-broken matrix test

- `058b488` rewrote `_, _ = fmt.Fprint(...)` to `_ = fmt.Fprint(...)` in three
  places in `internal/matrix/matrix_test.go`, which does not compile
  (`assignment mismatch: 1 variable but fmt.Fprint returns 2 values`).
- Restored the two-value assignments; `go test ./internal/matrix/` green.

### 4. Made the clone gate green without touching the baseline

- CI-pinned detector: `art-dupl` **v0.7.2**, `-t 1`, same flags as the workflow.
- Start state: **9 new clone groups** from the push train (session-gated JSON
  prologues, JSON response tail, cap-refusal branches, loopback guard, render
  tail, 2xx arms).
- Extracted one homes:
  - `api.writeJSONBody`, `api.storeAtCapacity` (`internal/server/api/responses.go`)
  - `api.pushSession`, `api.asciiIdentifier` (`internal/server/api/push_api.go`)
  - `server.requireThread`, `server.renderHTML` (`internal/server/actions.go`,
    `pages.go`)
  - `server.machineSurface` (`internal/server/internal_directory.go`)
  - `server.textAgentSession`, `server.textAgentGate`,
    `server.textAgentPanelAction` (`internal/server/textagent_api.go`)
  - `push.statusOK` (`internal/push/push.go`)
- End state: **0 new groups** vs the 75-group baseline; the baseline file is
  untouched.

### 5. Kept the error-code registry authoritative

- First attempt put the push transport status ladder into a shared
  `transportAnswer` helper with `op` as a parameter. `TestErrorCodeRegistryIsFresh`
  rejected it: the registry parser requires the `push.*_send` codes as literals
  at the call site, and the helper's first string (a format string) was parsed as
  a code.
- Reworked to: `push.statusOK` dedups the 2xx expression, each transport keeps
  an early `if statusOK(...) { return nil }`, and each failure switch keeps its
  literal codes and messages in its own file. Registry test green.

### 6. Added micro-tests for the new pure helpers

- `internal/server/api/responses_test.go`: `storeAtCapacity` (cap vs unrelated
  error) and `writeJSONBody` (content type + body).
- `internal/push/status_internal_test.go`: `statusOK` table.
- The remaining helpers are exercised through handler-level tests
  (`push_api_test.go`, `contacts_api_test.go`, `internal_directory_test.go`,
  `trace_test.go`, server suite), but do not have dedicated unit tests yet
  (see b).

### 7. Documentation and memory updates

- `CHANGELOG.md` `[Unreleased]` → `### Changed`: the dedup train plus the
  `.buildflow.yml` / formatter restoration.
- `docs/dedup-registry.md`: 9 new EXTRACTED rows plus a 2026-10-24 sweep-log
  entry (0 new groups; no baseline write; includes the transport-ladder ruling).
- `AGENTS.md`: new helpers appended to the One-home helpers bullet.
- This status report.

### 8. Verification battery (all local)

| Gate                                                   | Result                                                                               |
| ------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| `go build ./...`                                       | pass                                                                                 |
| `go test -count=1 ./...`                               | all packages pass                                                                    |
| `nix develop .#ci -c go test -race ./...`              | all packages pass                                                                    |
| `nix fmt`                                              | 0 changes after final run                                                            |
| `checks.format`                                        | pass                                                                                 |
| `checks.island-js`                                     | 228 tests, 0 fail                                                                    |
| `checks.island-lint`                                   | pass                                                                                 |
| `checks.statix`, `checks.deadnix`                      | pass                                                                                 |
| `nix build .#webphone`                                 | pass                                                                                 |
| `checks.vulnix-triage`                                 | pass                                                                                 |
| `checks.webphone-module`                               | pass                                                                                 |
| `scripts/buildflow.sh` (`BUILDFLOW_NO_RESULT_CACHE=1`) | passed with warnings                                                                 |
| CI-pinned `art-dupl check . -t 1`                      | 0 new groups (baseline 75)                                                           |
| `go mod tidy -diff`                                    | empty (go.mod stays `go 1.27.2`, which cqrs-htmx v4.14.1 / usermgmt v4.14.3 require) |

---

## b) PARTIALLY DONE

1. **Push and CI verification.** Everything is committed locally across three
   daemon commits (`7f596b8`, `cc48a17`, `b7e6efb`), but none is pushed. CI has
   therefore not been observed green at HEAD. Local runs replicate every job,
   but a push is the only end-to-end proof.
2. **Micro-test coverage of extracted helpers.** Only `statusOK`,
   `storeAtCapacity`, `writeJSONBody` have direct unit tests.
   `machineSurface`, `requireThread`, `pushSession`, `asciiIdentifier`,
   `textAgentSession`/`textAgentGate`/`textAgentPanelAction` and `renderHTML`
   rely on handler-level tests.
3. **Stale-baseline protocol note.** The sweep-log line I appended to
   `docs/dedup-registry.md` does not yet link to this status report (the protocol
   says "append one line ... with a link to its status doc").
4. **Lessons capture.** The daemon-damage incident is recorded in the CHANGELOG
   and dedup registry, but not in `docs/lessons.md`, which is the declared home
   for war stories.
5. **Consistency of the new render/JSON helpers.** `renderHTML` is used by the
   two generic paths (`pages.partial`, `panels.partialThread`) and the text-agent
   message tail; other render tails (`actions.go` toast path, textagent
   thread-scoped path, shell/nav/passkey pages) still set the header inline.
   `writeJSONBody` is used by contacts + push, not yet by CSRF or transcription.
6. **Root cause of the daemon damage.** Not investigated. The deletion and the
   oxfmt pass were repaired but can recur on the next daemon cycle.
7. **`TODO_LIST.md`.** No rows were added for the follow-ups this session
   surfaced (daemon recurrence, push stall, helper test debt).
8. **Push-stall watchdog.** `scripts/push-stall-watch.sh` was not run; the
   60-minute bar is now exceeded (>1 h 55 min from the oldest unpushed commit),
   and the status is established by direct `git ls-remote` polling instead.

---

## c) NOT STARTED

1. Full `nix flake check` (includes the KVM-gated `webphone-backup` and
   `webphone-backup-drill` VM tests). Only the fast checks and the package build
   were run.
2. Live smoke test (`scripts/webphone-smoke.py --bin ...`) after the repair.
3. Visual harness (`scripts/ui-capture.py`, 20 shots) after the repair.
4. Stack browser E2E on the next relock (nothing in the DOM contract changed,
   but the release-runbook obligation stands).
5. `nix run .#vulnix` (runtime-closure vulnix triage; `checks.vulnix-triage`
   passed, the release gate itself was not run).
6. Incident entry in `docs/lessons.md` and/or `docs/ci-flake-ledger.md`.
7. Root-cause analysis of the pma unit's formatter/config behavior.
8. A guard that fails when `.buildflow.yml` disappears.
9. Refreshing `.buildflow-baseline.json` (47 accepted entries matched nothing in
   each of the last 3 runs, per the buildflow warning).
10. Cleanup of pre-existing quality warnings noticed this session (codespell 88,
    file-size, jscpd, preflight reimplementation warning).
11. Cross-build sanity (`aarch64-linux`) — not run; no packaging input changed.
12. `scripts/whitespace-drift.sh` — not run because nothing was staged (daemon
    commits land without staging by me).

---

## d) TOTALLY FUCKED UP (honest list)

1. **The daemon's commits were worse than one bug at a time.** The same
   auto-commit chain deleted a policy file, rewrote 20 formatter-owned JS files,
   and broke a Go test. Any one of those would have red-ed a different gate;
   together they obscured which gate was "the" failure.
2. **My first dedup extraction was the wrong shape.** `transportAnswer` was a
   clean-looking shared ladder, but it conflicted with the project's literal-code
   registry gate and cost one extraction/revert cycle. The lesson is cheap and
   recorded: run `go test ./internal/arch` immediately after any errorfamily
   refactor.
3. **My refactor created new clone groups while removing old ones.**
   Extracting `pushSession`, `requireLoopback`, and `statusOK` produced three new
   idiomatic groups (3-line session prologue ×3, loopback guard ×2, `case
   statusOK` ×2). They were then re-extracted into `machineSurface`,
   `textAgentPanelAction` and success-first early returns. The net result is
   good, but the path was two iterations longer than necessary.
4. **I did not run buildflow at the start.** The AGENTS rule says run a full
   buildflow at train start. Running it late meant the daemon damage and the
   pre-existing warnings surfaced only after the code changes.
5. **I did not catch the matrix compile break from the diff alone.** The LSP
   diagnostics did (once the project loaded), but a `go build ./...`/`go vet`
   sweep should have preceded any edit.
6. **I let the daemon commit in mid-flight chunks.** The resulting history is
   three `chore: auto-commit` commits and several red CI runs at intermediate
   SHAs. No content is wrong at HEAD, but the history is noisy and the
   "narrative commit per phase" convention was not followed.
7. **The registry sweep-log entry lacks the required status-doc link.** Minor,
   but it violates the written protocol (see b.3).
8. **I did not push.** Deliberate (critical rule), but it means the user-visible
   proof (CI green on main) is absent, and main is red right now.

---

## e) WHAT WE SHOULD IMPROVE

### Process (this session's lessons)

1. **Read the damage before fixing the symptom.** The first investigation move
   should be `git log --stat` + `git diff <last-known-good>` after any daemon
   activity. That is what exposed the deleted config and the formatter reflow.
2. **Run buildflow at train start.** A late full run lets daemon damage age.
3. **Check the project's own guard tests before designing a shared helper.**
   The error-code registry is a _literal-at-call-site_ contract; a shared error
   ladder is structurally not allowed. Check `internal/arch` tests first.
4. **Expect the clone detector to flag new idioms created by extraction.**
   Remove a clone in a way that does not create a new 3-line twin; plan the
   helper's call-site shape before writing it.
5. **Use the push-stall watchdog mechanically.** `scripts/push-stall-watch.sh`
   would have surfaced the stalled push leg without 5-minute polling rounds.

### Engineering / durability

6. **Daemon guard hole.** The pma unit exports
   `GIT_CONFIG_COUNT/KEY_0/VALUE_0=core.hooksPath=/var/empty`, so daemon commits
   bypass the commit-msg guard. The 2026-10-15 AGENTS note already documents
   this; 2026-10-24 is evidence that it also lets destructive formatter/config
   changes through.
7. **No invariant protects policy files.** `.buildflow.yml` is load-bearing for
   every gate; nothing fails if it is deleted (buildflow simply runs with
   defaults). A CI check should assert its presence and key contents.
8. **Formatter ownership needs a hard fence.** treefmt owns
   island/shell/theme-preload; oxfmt owns island-tests and Go. The daemon ran a
   formatter across the wrong side of that fence. `checks.format` catches it
   only _after_ a commit; a pre-commit/preflight assertion would catch it before.
9. **The buildflow baseline is stale.** 47 accepted entries matched nothing for
   three consecutive runs; refresh it deliberately (CI-pinned version) or the
   accepted-noise mechanism slowly loses meaning.
10. **Quality-gate noise is high.** 88 codespell findings, file-size warnings
    (i18n.go at 480 lines), jscpd duplicates in island tests, and one
    canonical-helper warning are all pre-existing and detect-only. They are
    tolerated, but they train readers to ignore the gate.

### Housekeeping

11. Link this report from the dedup registry sweep log.
12. Add the daemon-damage class to `docs/lessons.md`.
13. Add TODO_LIST rows for the open follow-ups.
14. Decide whether the intermediate red CI runs on `ff185ae` and earlier should
    be left as history (no rewrite) or avoided in the future by manual narrative
    commits.
15. Keep the daemon from touching dependency/config/formatter ownership zones
    (the `.githooks/commit-msg` guard already refuses daemon commits touching
    `go.mod`/`go.sum`/`flake.lock`/`.buildflow.yml` — but only for humans).

---

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

1. Push `b7e6efb` (manual-push bar is now exceeded: main red + verified fix +
   oldest unpushed commit > 60 min) and watch the CI run to green.
2. Confirm all four CI jobs green at the pushed SHA (`nix build`, `go test
   -race`, fast flake checks, art-dupl).
3. Add the status-report link to the 2026-10-24 dedup sweep-log line.
4. Write the `docs/lessons.md` entry for the daemon damage (config deletion +
   oxfmt reflow + `_ = fmt.Fprint` break; detection commands).
5. Root-cause how the daemon deleted `.buildflow.yml` (pma unit formatter step?
   heuristic commit? hook bypass?) and document the finding.
6. Add a CI/flake check asserting `.buildflow.yml` exists and still contains the
   `baseline:` and `exclude:` keys.
7. Add a treefmt/pre-commit assertion that island/shell files match prettier
   (not just after-commit `checks.format`).
8. Decide the fleet policy for the pma unit's `core.hooksPath=/var/empty`
   override (owner call; affects daemon commits fleet-wide).
9. Run full `nix flake check` including the KVM-gated backup tests.
10. Run `scripts/webphone-smoke.py --bin <fresh binary>` and record the verdict
    lines (`smoke: N passed`, restart, boot-failure).
11. Run the 20-shot `scripts/ui-capture.py` harness after the repair.
12. Run `nix run .#vulnix` and triage the runtime closure.
13. Add a direct unit test for `server.machineSurface` (disabled seam + non-
    loopback 403 + loopback pass).
14. Add a direct unit test for `server.requireThread` (no session, bad id, good
    id; both `http.NotFound` and `h.notFound` callbacks).
15. Add a direct unit test for `api.asciiIdentifier` (letters/digits/extras,
    empty, unicode, disallowed characters).
16. Add a direct unit test for `api.pushSession` (401 path, seam-off 404, happy
    path) with the existing stub seams.
17. Add a direct unit test for `server.textAgentGate` (disabled seam writes the
    styled 404).
18. Add a direct unit test for `server.textAgentPanelAction` (error mapping,
    empty-toast vs toast, panel refresh).
19. Add a render-failure unit test for `server.renderHTML` (500 after headers).
20. Migrate `actions.go`'s toast render tail onto `renderHTML`.
21. Migrate textagent's thread-scoped render tail onto `renderHTML`.
22. Migrate `pages.go` shell/nav tails and `passkey_api.go` enroll tail onto
    `renderHTML` if the helper's header semantics stay identical.
23. Migrate `csrf_api.go` onto `writeJSONBody` (keeping `Cache-Control`/`Vary`
    before it).
24. Decide whether the baselined `transcribe.go` JSON tail should also use
    `writeJSONBody` (would delete a baseline group; harmless either way).
25. Re-run the CI-pinned `art-dupl check` after items 20-24 and keep it at 0 new.
26. Refresh `.buildflow-baseline.json` with the CI-pinned v0.7.2 and the gate's
    exact flags if the stale-entry warning is to be cleared.
27. Add TODO_LIST rows for the daemon-recurrence guard, push-stall watchdog use,
    and the helper test debt.
28. Add a check that `go mod tidy -diff` stays empty in CI (keep `go 1.27.2`;
    cqrs-htmx v4.14.1/usermgmt v4.14.3 demand the patch floor).
29. Verify `release.sh` still greps `webphoneVersion` correctly (untouched this
    session; cheap proof).
30. Re-run the vendorHash `--rebuild` recipe only if a future `go get` lands (no
    dependency changed this session).
31. Address the preflight warning about `internal/pbx/client.go ResolvePath`
    re-implementing the canonical helper.
32. Add `docs/architecture-understanding/archived` and `docs/research/archived`
    to the lychee excludes (preflight warning).
33. Decide whether the 88 codespell hits (`Speaches` product name, intentional
    German words, `onText`) get dictionary entries or stay detect-only.
34. Split `internal/web/views/i18n.go` (480 lines vs the 350 warning) or
    baseline it deliberately.
35. Review the island-test jscpd duplicates (9 findings) for genuine copy-paste
    worth extracting.
36. Rebuild/reinstall the local buildflow binary (preflight: built at `3509b30`,
    repo HEAD `08e9cae`) so future preflights reflect current BuildFlow.
37. Run `scripts/push-stall-watch.sh` at the start of the next session and
    record the anchor it prints.
38. Add the manual-push-bar anchor command to the status-report template.
39. Verify the daemon did not touch `.art-dupl-baseline.json` (currently 0
    diffs) and keep baseline writes manual-only.
40. Confirm the dprint reformat of the 2026-10-23 observability plan is
    acceptable (verified formatting-only) and decide whether docs should be
    dprint-owned by policy.
41. Run `scripts/whitespace-drift.sh` on the next staged change before commit.
42. Add a handler-level test for the real contacts cap path if one is missing
    (500/extension).
43. Add a handler-level test for the 8-device push cap path if one is missing.
44. Document the `machineSurface` asymmetry (`http.NotFound` for the SIP
    directory vs `h.notFound` for the flight recorder) at the call sites.
45. Re-read the text-agent action refactor against the original approve/discard/
    set-mode semantics (tests green, but a manual parity read is cheap).
46. Verify the `statusOK` early-return change keeps the FCM
    `200-with-unregistered-marker` behavior (the marker is currently only
    checked on non-2xx lanes; confirm this matches the previous semantics).
47. Confirm no served-markup strings changed (the DOM contract was untouched;
    the release-runbook E2E obligation still applies on the next relock).
48. Add a short "daemon damage" section to the release runbook or AGENTS pointing
    at the detection commands and the restored invariants.
49. Decide whether an auto-commit daemon is allowed to commit formatter/config
    changes at all, or should be restricted to content under its owned paths.
50. Re-run the whole verification battery once after the push so the status
    report's final claim ("CI green at HEAD") becomes true.

---

## g) QUESTIONS I CANNOT ANSWER MYSELF (max 3)

1. **Push authorization.** The manual-push bar is now exceeded (main red,
   verified fix in hand, oldest unpushed commit > 1 h 55 min), but my standing
   rule says never push unless asked. May I push `b7e6efb` and watch CI to green?
2. **Daemon root cause.** Do you want the pma unit / fleet hook policy
   (`core.hooksPath=/var/empty`, plus whatever ran the formatter) investigated
   and fixed now, or documented as a recurring risk and left to the daemon
   owner?
3. **Commit ownership.** Should I take over narrative commits at phase
   boundaries going forward (avoiding the daemon's partial red commits), or keep
   letting the daemon commit and just fix the end state?
