# Status Report — Master-TODO session: T21/T22 deliverables + full verification battery (2026-10-02 11:43)

Session scope: the owner's blanket **"NOW GET SHIT DONE! The WHOLE TODO LIST!"**
issued on top of the docs-health AUDIT v5 sweep
(`docs/status/2026-10-01_17-27_*`) and the master Pareto plan
(`docs/planning/2026-10-01_17-32_SUPERB-master-todo-pareto-execution-plan.md`,
tasks T01–T27). This session's mandate was to EXECUTE the plan and verify it.

**Headline finding:** by the time this session started, a **concurrent Crush
session (sessions 2–7, ~18 h, ~40 auto-commit trains) had already shipped the
overwhelming majority of the assistant-actionable backlog** — T05 (island boot
language), T06 (ring-silence AudioContext), T07 (mic pre-warm code), T08 (owed
pins), T10 (ETag/gzip), T11 (modulepreload), the T12–T19 UI/UX trains (messages
organization, snippet replies, trust feedback, fax depth, history filters,
voicemail), T20 (hygiene), T22 (health.css input + canary), T23
(`scripts/ui-capture.py`), and T26 (AGENTS compaction). This session therefore
delivered the **non-overlapping T21 nix-review batch-2 items**, fixed a
regression the concurrent train introduced, then pivoted to the **single most
valuable remaining action: running the full verification battery** the plan
flagged as "the whole served surface is unverified" (T03).

> ARCHIVED 2026-10-03 (docs-health v6 sweep): the T21 batch-2 wins + the
> verification battery stand (module golden, release guard, devshell dedupe,
> actionlint, family pins, format repair); the KVM/quiet-window legs ran green
> at 20:12; the bootreport nolints landed; owner legs ride their rows.
> Per-item verdicts inline.

---

## a) FULLY DONE (implemented AND verified this session)

~~1. **T21a — module-output golden (`checks.webphone-module`).**~~ done — this session (report of record)
   The generated module output was only covered by substring asserts, which miss
   reordering/whitespace. Added a full-text golden:
   - `nix/module-output.nix` renders the Caddy vhost `extraConfig`, the
     `retentionDays = 7` backup script, and the settings JSON in one document.
   - `nix/module-output.golden` is the committed fixture (1865 B).
   - `nix/module-check.nix` gained a `module-output-golden` case that `diff`s
     the rendered text against the fixture and fails on drift.
     Verified: `nix build .#checks.x86_64-linux.webphone-module` green (the
     linkFarm now carries **16** entries incl. `module-output-golden`, which
     prints _"module output matches the committed golden"_).

~~2. **T21b — release-time version guard (`scripts/release.sh`).**~~ done — this session (report of record)
   The flake's `webphoneVersion` is a manual single point of failure; nothing
   checked it against tag history. Added a step-1 guard: before the bump, the
   binding MUST equal `git describe --tags --abbrev=0` (normalised `v`), else
   exit 1. False-positive edge (a deliberate bump-before-tag commit) documented
   in-code; override `WEBPHONE_RELEASE_SKIP_VERSION_GUARD=1`. Verified: values
   `flake=2.8.0 tag=2.8.0` → passes; `bash -n` clean.

~~3. **T21c — devShell Go pin/env dedupe (`nix/devshell.nix`).**~~ done — this session (report of record)
   `devShells.default` and `devShells.ci` duplicated the `go_1_27` pin and the
   `GOTOOLCHAIN=local` env; factored to shared `goTools`/`goEnv` bindings so the
   CI and interactive shells cannot drift. Verified: both shells evaluate and
   report `go1.27.1`, `GOTOOLCHAIN=local`; `.#ci` carries templ +
   golangci-lint; the default shell keeps all interactive tools.

~~4. **T21d — `actionlint` over `.github/workflows/ci.yml`.** Ran the real linter:~~ done — this session (report of record)
   **exit 0, zero findings.**

~~5. **T21e — accepted exceptions recorded (`CHANGELOG.md`).**~~ done — this session (report of record)
   Added a "Nix-review batch 2 (2026-10-02)" entry recording the module golden,
   the release guard, the devshell dedupe, the health.css input/canary, and the
   snippet error-family pins, plus the explicit note that the `go-standard`
   migration stays declined and the hardcoded `webphoneVersion` / module-check
   stand-in permissiveness are accepted exceptions.

~~6. **T22 — reply-snippet error-family pins (`internal/store/family_test.go`).**~~ done — this session (report of record)
   The new `store.snippet_save` / `store.snippet_delete` codes and the
   `ErrSnippetListFull` sentinel were unpinned. Added pins (family +
   `errorfamilytest.AssertCode`) and a closed-DB persistence case. Verified:
   `go test -count=1 ./internal/store/` green.

~~7. **Regression repair — `format` gate restored.**~~ done — this session (report of record)
   The concurrent train committed a rebuilt `internal/web/assets/health.css`
   that was **not prettier-clean**, so `checks.x86_64-linux.format` (treefmt owns
   `*.css`) was RED. Ran `nix fmt internal/web/assets/health.css`; the format
   check is green again. (Root cause recorded in `f)`.)

~~8. **erraudit `context_loss` false positives suppressed (`internal/server/panels.go`).**~~ done — this session (report of record)
   The upgraded erraudit flagged two `archivedCount` OUT-param error paths
   (the documented garbage-by-definition FP class, already suppressed at
   `store/messages.go:268`). Added reasoned
   `//nolint:erraudit // context_loss FP: ... is an OUT param …` on both returns.
   Verified: erraudit now reports **3** findings, down from 5 (the 2 context_loss
   entries are gone).

**Verification ledger (all green unless noted):**

| Gate                                                                                            | Result                                                                                                                                                   |
| ----------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `go test -count=1 ./...`                                                                        | **all packages `ok`** (cmd, app, arch, blob, config, crm, domain, fax, gateway, messaging, paperless, pbx, server, session, store, vcard, assets, views) |
| island `node:test` (all `*.test.mjs`)                                                           | **164 pass / 0 fail**                                                                                                                                    |
| `checks.{format,island-js,island-lint,statix,deadnix,vulnix-triage,webphone-module,health-css}` | **all build green**                                                                                                                                      |
| `nix build .#webphone`                                                                          | green (2.8.0)                                                                                                                                            |
| live smoke (`scripts/webphone-smoke.py`)                                                        | **47 passed + 4 restart + 8 boot-failure, 0 failed**                                                                                                     |
| `actionlint .github/workflows/ci.yml`                                                           | exit 0                                                                                                                                                   |
| `buildflow` full (no result cache)                                                              | ran; findings gate tripped by **gomod-check (54, known FP)** + **erraudit (3 remaining)**                                                                |

---

## b) PARTIALLY DONE

~~1. **T21f — KVM backup VM test + backup-drill NOT run.** BLOCKED by the host:~~ done — ran green later (the 20:12 battery: flake check incl. the KVM VM, all checks passed)
   `loadavg` sat at **120–156** for the whole session (32 cores), far above the
   project's own release `load_gate` default of 8. Running the KVM VM test under
   that contention would be slow and flaky (the known timeout flake mode) and
   would mask/resemble a real regression — deferred by the same policy the
   release script enforces.

~~2. **erraudit — 3 of 5 findings remain.** `cmd/webphone/bootreport.go:235,253,254`~~ done — all three reasoned nolints landed (bootreport.go:235/253/254)
   flag `_, _ = fmt.Fprintf(...)` as _"Error may be ignored using blank
   identifier"_. These are the concurrent session's boot-contract code; the
   intentional-ignore (writing the contract block immediately before `exit`) is
   legitimate and needs a reasoned `//nolint:erraudit // …`, but I left them to
   the session that owns the file (see `g)`).

~~3. **buildflow "green" is unreachable by design here.** Even with the erraudit~~ process record — the gomod-check FP is documented AGENTS noise
   items fixed, the findings gate still trips on **gomod-check (54)** — the
   verified FALSE POSITIVE documented in AGENTS ("`go mod vendor` regenerates
   `modules.txt` byte-identically; `go build -mod=vendor` green"). No assistant
   action can clear it.

~~4. **T27 docs harvest — only the CHANGELOG arm is mine.** The concurrent session~~ done — the living docs settled; this v6 sweep closes the history arms
   is actively rewriting `TODO_LIST.md` / `README.md` / `FEATURES.md`; I added
   the CHANGELOG history entry and did not race their living-doc edits.

---

## c) NOT STARTED (owner-gated or owned by the concurrent session)

~~- **Owner legs:** T01 (v2.8.0 deploy tail — ssh/deploy), T02 (prod SMS bridge~~ routed — TODO owner rows (the handover legs)
  journal), T04 (stack browser E2E), T09 (owner-calls sitting), T24 (cross-repo
  obligations), T26 (any further owner decision), T27 (4-plan `g1` call).
~~- **Concurrent-session-owned:** T11 perf extras, T12–T19 UI/UX M9–M26 remainder,~~ done — the concurrent train shipped them all (T11 extras, T12–T19, T23, T25.1)
  T23 (their `scripts/ui-capture.py` just landed, `ui-shots/` produced), T25
  (server carve / schema_version gate).
~~- **Not touched:** `check-rows.py` fix-or-migrate; the 37 pre-existing archived~~ done — superseded: the v6 sweep runs check-rows + the completeness grep over the corpus at its close
  files using the old marker-cell format.

---

## d) TOTALLY FUCKED UP (and how)

**Nothing shipped broken by this session.** One derive-then-discard cycle cost
time:

~~1. **I rebuilt twice for a task a concurrent agent had already done.**~~ process record
   I independently built `internal/web/assets/health.css.input` +
   `scripts/build-health-css.sh` (T22) before discovering the concurrent session
   had committed its own, more mature versions (staged build, locked-nixpkgs
   rev, self-perpetuation guard). My first rebuild also **regressed the dark
   variant** (canary 140 → 22) because my `@source` set was incomplete; I
   detected it by grepping `:where(.dark` and **restored the committed
   artifact** before committing anything. Lesson: check for a concurrent
   session's work BEFORE building shared tooling.

~~2. **My `checks.nix` health-css canary edit was rejected** ("file modified since~~ process record
   last read") because the concurrent session added its own canary in the same
   window. Correct outcome — I did not force it.

~~3. **A first full `go test ./...` failed** on `internal/server/messages_test.go`~~ process record
   - `views/i18n_test.go`. I attributed it correctly: those were the concurrent
     session's **in-flight, uncommitted** `messages.templ` edits (mtimes
     advancing live). I did not "fix" their files; the later full run was green.

---

## e) WHAT WE SHOULD IMPROVE

~~1. **Two agents on one repo is the dominant hazard, not the work.** The daemon~~ process record — the claims-lockfile idea wasn't adopted; the AGENTS concurrent rule + small commits stand
   committed MY work-in-progress (health.css input/script) mid-edit, and the
   concurrent session overwrote it. The AGENTS "concurrent sessions" section
   warns about reverts but not about **same-task duplication**. A cheap fix: a
   lockfile / claim file (e.g. `.crush-claims`) that a session writes before
   starting a plan task.
~~2. **Derived artifacts must be formatted by their generator.** The~~ routed — TODO tooling row (the nix fmt append, named there)
   `build-health-css.sh` rebuild shipped an un-prettified `health.css` and went
   RED on `format` until this session reformatted it; the script should end with
   `nix fmt`. (The concurrent session's version does not.)
~~3. **Gate "green" needs a defined meaning under the known gomod-check FP.**~~ process record — the noise semantics live in the AGENTS buildflow-health warning
   `buildflow` can never exit 0 here; the team should record "green = 0 findings
   except gomod-check" so a real regression is not lost in the noise.
~~4. **Run the KVM tests when the host is quiet.** T21f is genuinely undone; a~~ done — the KVM battery ran green in the later quiet window (20:12)
   quiet window (load < 8) is required.
~~5. **Verify my own claims before the report** — the erraudit re-run confirmed the~~ process record
   panels.go fix (5 → 3); do that for every "fixed" claim.

---

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Ordered by value/effort; `[owner]` = human-only; `[session]` = concurrent agent.

~~1. **Add the missing `nix fmt` to `scripts/build-health-css.sh`** so a rebuild~~ routed — TODO tooling row (the nix fmt append, named there as t21-t22 f1)
   cannot re-break the `format` gate. `[session]`
~~2. **Fix the 3 `bootreport.go` erraudit `ignored` findings** with reasoned~~ done — all three reasoned nolints landed (bootreport.go:235/253/254)
   nolints (or thread a writer error). `[session]`
~~3. **Run `checks.x86_64-linux.webphone-backup` (KVM) + the backup drill** in a~~ done — ran green later (the 20:12 battery)
   quiet window. `[owner, load-gated]`
~~4. Run `nix flake check --all-systems` / aarch64 cross-build + ELF verify.~~ done — aarch64 exit-green at the T22 close-out (ELF verify re-owed at each final gate)
~~5. Run `nix run .#vulnix` and triage the runtime closure (34 raw advisories).~~ done — vulnix zero-real-advisories at the T22 close-out (TODO island row)
~~6. T01 — deploy the v2.8.0 tail on pbx.artmann.tech. `[owner]`~~ routed — TODO deploy-tail row (owner terminal)
~~7. T02 — journal the telnyx-webhooks unit; fix + test SMS. `[owner]`~~ routed — TODO owner row (the SMS-bridge journal leg)
~~8. T04 — stack browser E2E (445 s budget) for the served-markup delta.~~ routed — TODO island-honesty row (the owed stack browser E2E)
~~9. Fix `golangci-lint`: `internal/store/db_test.go:268` unchecked `rows.Close`.~~ resolved by events — the later full suites + lint are green
~~10. Fix `oxfmt`: `internal/web/assets/island-tests/selftest.test.mjs` unformatted.~~ resolved by events — the format gates are green
~~11. Fix `ruff-format`/`ruff-check`/`mypy`/`vulture` on `scripts/perf-baseline.py`~~ routed — TODO tooling row (the scripts-lint batch; mypy named there)
    - `scripts/ui-capture.py` (38 + 42 + 16 + 1 findings). `[session]`
~~12. Address `jscpd` duplication in `calls.test.mjs` (2 clones). `[session]`~~ not adopted — below the bar (test-file clones; the dedup registry scopes product code)
~~13. `nix-checker`: consider extracting `vendorHash` to `nix/vendorHash.nix`.~~ not adopted — below the bar
~~14. `nix-flake-check`: add `meta.description` to `apps.x86_64-linux.vulnix`.~~ not adopted — below the bar
~~15. Decide `go-mod-ignore-check` (vendor not ignored in go.mod) posture.~~ not adopted — below the bar
~~16. `go-version-auto-configure`: go directive has a patch component (1.27.1).~~ not adopted — below the bar
~~17. Confirm the AGENTS.md size fix (405 lines > 377 cap) or compact again. `[owner permission]`~~ done — compacted 2026-10-02 (404→377)
~~18. T11 — outgoing-call mic warm + ICE gathering-time panel + `iceServers` eval.~~ done — T11 shipped (dial-focus warm + ICE setup line; the trim eval not adopted)
~~19. T12–T19 — remaining UI/UX M9–M26 workstreams. `[session]`~~ done — the T12–T19 trains shipped (CHANGELOG Unreleased)
~~20. T23 — review the 12-shot `ui-shots/` matrix; persist an exit criterion.~~ done — the matrix reviewed at the T23 visual pass (14 shots)
~~21. T25 — trigger-gated `internal/server` carve; `schema_version` on first ALTER.~~ resolved by events — T25.1 shipped with T18; the carve trigger stayed unfired
~~22. `check-rows.py` fix-or-migrate + the 37 old-format archived files.~~ done — superseded: the v6 sweep runs check-rows + the completeness grep over the whole corpus at its close
~~23. Owner-calls sitting (28 rows) — unblocks the next product cycle. `[owner]`~~ routed — TODO owner row (the sitting, 28 rows)
~~24. Post the v2.8.0 release announcements. `[owner]`~~ routed — TODO owner row (the announcements)
~~25. Standing watches re-check (sip.js 0.22, templ-components, oxlint globals).~~ routed — standing watches row
~~26. erraudit tier-2 monthly re-measure (next due 2026-10-22; early 2026-10-01: 0/0/0).~~ routed — standing watches row (next due 2026-10-22)
~~27. Commit the 4 owner-gated plans or keep them live (`g1`). `[owner]`~~ owner — the 4 plans stay live pending the sitting (this sweep's standing default)
~~28. Add a `.crush-claims` lockfile convention to stop same-task duplication.~~ not adopted — below the bar (the AGENTS concurrent rule + small commits stand instead)
~~29. Record the "buildflow green = all-but-gomod-check" definition in AGENTS.~~ process record — the noise semantics live in the AGENTS buildflow-health warning
~~30. Re-run the full battery after the next concurrent-session commit to keep it honest.~~ process record — the 12:57 closeout battery did exactly this

---

_Evidence: `nix/module-output.{nix,golden}`, `nix/module-check.nix`,
`nix/devshell.nix`, `scripts/release.sh`, `internal/store/family_test.go`,
`internal/server/panels.go`, `internal/web/assets/health.css`, `CHANGELOG.md`.
Verification commands: `nix develop -c go test -count=1 ./...`,
`nix run nixpkgs#nodejs -- --test --test-force-exit
internal/web/assets/island-tests/*.test.mjs`,
`nix build .#checks.x86_64-linux.{format,island-js,island-lint,statix,deadnix,vulnix-triage,webphone-module,health-css}`,
`python3 scripts/webphone-smoke.py`, `BUILDFLOW_NO_RESULT_CACHE=1
scripts/buildflow.sh`, `nix run nixpkgs#actionlint`._
