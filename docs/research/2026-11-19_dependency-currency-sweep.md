# Dependency currency sweep — 2026-11-19 (read-only, upstream half)

> **Date:** 2026-11-19 · **Claims verified as of 08:22 CET / 07:22 UTC** (remote tags, proxy updates, CI verdicts — all rot fast; re-verify before riding any train).
> **Mode:** READ-ONLY currency sweep per the bump discipline (AGENTS.md § Conventions: bumps do NOT ride the auto-commit daemon; they land via the mandatory tail — `go get` → `go mod tidy` → vendorHash re-pin → push → CI-watch). No go.mod/go.sum/flake.nix was touched. This report is the enumerate/baseline/classify half plus ready-to-run train plans.

## Verdict vs the 2026-11-13 CURRENCY note

**The 2026-11-13 "every scored family at upstream latest" state has DRIFTED in 6 days: 10 direct dependencies are behind their latest remote tags, every one with a consumed surface.** All 10 preflights read as ADDITIVE (no breaking migration required — two apparent breakages dissolved under verification, see § False-positive removals). The identity family (cqrs-htmx root + usermgmt + webauthn + loginpage + identity-model), templ-components, go-health, go-health-dashboard, go-sse, go-error-family, go-etag, go-retry, go-branded-id, go-codec, go-atomic-write, go-flightrecorder remain AT latest — the drift is concentrated in go-cqrs-lite's engine wave, templ, and four small SDKs.

Baseline at sweep time: `go test -race ./...`, `nix build .#webphone`, smoke, workflow-e2e all GREEN at sha `25301790`; CI is RED on the Docs-health battery only (dprint markdown on the docs(status) commit — the concurrent docs session's lane, not a Go surface). 10 unpushed daemon commits pending. Trains below inherit this pre-existing-failures record.

## Sweep table (direct deps behind, all consumed-surface → actionable)

| Module                                  | Pin → Latest          | Consumed surface (files)                                                              | bump-preflight verdict                                                                                                                                                                                                                                            |
| --------------------------------------- | --------------------- | ------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| go-cqrs-lite/system/v4                  | v4.18.0 → v4.19.0     | 33 (internal/cqrs, readmodels)                                                        | additive; **fix-class: RawQuery+ProjectionSpec mixed domains no longer crash-loop the projection worker** (`no decoder registered` class) — webphone has 10 RawQuery files                                                                                        |
| go-cqrs-lite/metaengine/v4              | v4.23.0 → v4.26.0     | 29 (internal/cqrs, readmodels)                                                        | additive; Typed fold family, batch-ingestion spine, deterministic persisted bytes, `SSEDroppedEvents` counter, slice-of-unit fold construction panic; engine family promoted Experimental→Stable. The 60–600s cap-trim deadlock fix is ALREADY in the v4.23.0 pin |
| go-cqrs-lite/metaengine/sqliteengine/v4 | v4.8.0 → v4.9.1       | 2 (threadsregister.go driver registration + dual test)                                | additive (removed symbol is the library's own test); batch-ingestion engine leg                                                                                                                                                                                   |
| a-h/templ                               | v0.3.1100 → v0.3.1118 | all views + 10 generated `*_templ.go` (runtime)                                       | additive, 59 files; **requires `templ generate` regen with CLI pinned to the new go.mod version** (the CLI≠library pairing rule, AGENTS.md Commands)                                                                                                              |
| httputil                                | v1.8.0 → v1.9.0       | 13 (middleware chain, Metrics, SecurityHeaders, Nonce, NewServer)                     | additive; `Capabilities.Pusher` field (issue #47); removed symbols are the library's own tests                                                                                                                                                                    |
| modernc.org/sqlite                      | v1.61.1 → v1.62.0     | 7 (store drivers: system.db + usermgmt.db)                                            | 79 files; removed exports are libc-internal shim symbols on non-linux ports — not consumed. Persistence substrate → the heaviest gates                                                                                                                            |
| go-idempotency                          | v0.3.2 → v0.6.0       | 1 TEST-only (`internal/server/hooks/idempotency_contract_test.go`; prod NO-GO stands) | additive; **ships `idempotency.WithClock` — the 2026-11-08 upstream ask (injectable clock vs wall sleeps) LANDED**                                                                                                                                                |
| go-paperless                            | v0.5.2 → v0.5.3       | 3 (internal/paperless adapter; `sdk.CustomFieldValue` at paperless.go:111)            | additive — preflight's "removed" Upload/CustomFieldValue symbols MOVED client.go→upload.go verbatim (verified in the v0.5.3 tag tree); the real fix is FK-clearing PATCH goldens (bank-sync's surface, not webphone's)                                            |
| sixafter/nanoid                         | v1.66.0 → v1.67.0     | 4 (domain ids, blob store)                                                            | NO-GO-FILE-CHANGES: zero .go deltas — trivial ride-along                                                                                                                                                                                                          |
| golang.org/x/oauth2                     | v0.37.0 → v0.38.0     | 2 (internal/sso)                                                                      | additive; 2 files (deviceauth) — webphone rides the OIDC code flow                                                                                                                                                                                                |

### False-positive removals (why the preflight scan alone is not a verdict)

- **go-paperless**: bump-preflight listed `Capabilities, CustomFieldValue, UploadRequest, …` as removed from `client.go`. `git grep` at tag `v0.5.3` shows `upload.go:21 UploadRequest`, `upload.go:49 CustomFieldValue`, `upload.go:57 (*Client).Upload` — a file move, same symbols (failure mode F6: changelogs/scans omit context; verify yourself).
- **go-idempotency / httputil / sqliteengine / templ**: every "removed exported symbol" lives in the library's own `*_test.go`/bench files — not consumer surface.

### No-consumed-surface / indirect drift (note only — no action)

- **go-datastar v0.6.2 → v0.7.0 and go-datastar/static v0.6.1 → v0.8.0**: unchanged stance since 2026-11-13 (indirect-riding consumers). go-health-dashboard v0.13.3 (latest) still pins datastar v0.6.2/v0.6.1 → the registry trigger "dashboard library swaps its embedded SDK" has NOT fired.
- **go-credentials-store v0.4.3** (absent family; was v0.4.2 at the 2026-11-13 note) — NO-GO premises unchanged. **go-filewatcher v2.5.0** — unchanged.
- **Indirect riders** (bitset, go-jose v4.1.9, gofrs/flock v0.13.2, go-webauthn/x v0.4.1, sixafter aes-ctr-drbg/prng-chacha, x/crypto v0.59, x/net v0.63, x/sys v0.50, x/mod, x/sync, x/tools, prometheus family, …): MVS lifts them with the direct bumps' `go mod tidy`; they are never bumped separately here.

## Reopen-trigger registry: NONE FIRED

Checked against TODO_LIST.md § Fleet-libs reopen-trigger registry (2026-11-04, T25):

- **metaengine NESTED group addressing** (→ delete `trimCall`): v4.24–v4.26 tags are batch-ingestion spine / Typed folds / determinism — no nested per-call floor inside a per-owner group. Not fired.
- **appkit tag beyond v0.12.0 ships a new battery/module**: root tops at v0.12.0; `flightrecorderhealth` stays v0.1.7 (the version already verified 2026-11-15); httpx v0.2.0, preset v0.1.1. Not fired.
- **dashboard swaps its embedded SDK** (go-datastar): dashboard still pins v0.6.x (above). Not fired.
- **templ-components new status/toast component**: no release beyond v1.27.0. Not fired.
- **usermgmt major bump / second identity consumer**: usermgmt at v4.20.0 latest, no major. Not fired.
- **go-branded-id parse-with-rules hook / new linter rule**: v0.7.1 is latest (unchanged). Not fired.
- All remaining triggers (second Go consumer, Idempotency-Key public contract, Type=notify, drop-folder, wedge incident, CLI structured output, ~10k-row read models, second readiness consumer, second retry loop, dashboard without Caddy, non-systemd targets, new file-persisting feature, latency localization, missed-update reports, second binary): no firing event observed in this sweep.

**Adjacent (not a registry trigger, but retires open work):** go-idempotency v0.6.0's `WithClock` is the durable fix the 2026-11-08 draft filing asked for. If the train lands green: (1) do NOT post `docs/drafts/2026-11-08_go-idempotency_contract-clock-race.md` (moot — say so in the draft's status line instead), (2) retire the TimingScale 5 band-aid in `idempotency_contract_test.go` by pinning the store clock (flake ledger row 42407471289 closes).

## Fleet release-hygiene findings (owner-facing, note-only for webphone)

1. **go-cqrs-lite ROOT release v4.14.1 is UNTAGGED** (root CHANGELOG newest heading, no `refs/tags/v4.14.1` on remote; `scripts/fleet-tags.sh` flagged it). Webphone does not consume the root module (0 importers) — but this is the dashboard-v0.13.0 class (changelog claims a tag that does not exist).
2. **sqliteengine dual-shape tags point at DIFFERENT commits**: `metaengine/sqliteengine/v4.9.1` → 9cb7654b vs `metaengine/sqliteengine/v4/v4.9.1` → 9586802c. The proxy serves the bare (resolvable) form, so consumers are fine — but which commit is the release? Worth an owner look (metaengine/v4.25.0 and the alias tags share this pattern).
3. **go-cqrs-lite changelog waves stop at the 2026-11-08 v4.23.0 entry while tags run to v4.26.0** (v4.24 "batch ingestion spine", v4.25, v4.26 "Typed fold family" carry release-tag subjects but no wave heading). Tag-ahead-of-changelog drift.

## Train plans (execute in this order; each is one daemon-free push + CI-watch)

The mandatory tail for EVERY train: `nix develop -c go get <mods>` → `nix develop -c go mod tidy` → **verify go.mod by grepping it** (never trust `go get` output) → check the `go` directive did not move past the devshell toolchain (currently `go 1.27.3`; a dep floor above it reds every Nix build — if it moves, decide floor-vs-toolchain before proceeding) → `nix build .#webphone.goModules --rebuild`, read `got:`, apply the vendorHash → gates → commit (human commit; the `.githooks/commit-msg` guard binds humans only — daemon commits touching go.mod are a defect class, report if seen) → push → `scripts/ci-verdict.sh <sha>` until not-pending. **Rollback (all trains): `git revert` the bump commit, re-pin vendorHash back (`nix build .#webphone.goModules --rebuild` again — the hash is content-addressed, the old value re-applies), push, CI-watch.**

### Train 1 — go-cqrs-lite engine wave (fix-value first)

`system/v4 v4.19.0` + `metaengine/v4 v4.26.0` + `metaengine/sqliteengine/v4 v4.9.1` (the trio releases together; the RawQuery decoder crash-loop fix is the motivator — webphone's 10 RawQuery files are the blast-radius class).

```console
nix develop -c go get github.com/larsartmann/go-cqrs-lite/system/v4@v4.19.0 github.com/larsartmann/go-cqrs-lite/metaengine/v4@v4.26.0 github.com/larsartmann/go-cqrs-lite/metaengine/sqliteengine/v4@v4.9.1
nix develop -c go mod tidy
# verify: grep the three pins in go.mod; go directive still 1.27.3
nix develop -c env -u GOEXPERIMENT go test -count=1 ./internal/cqrs/... ./internal/readmodels/... ./internal/server/...
nix develop -c env -u GOEXPERIMENT go test -count=1 -race ./...
nix build .#webphone.goModules --rebuild   # read got:, apply vendorHash in nix/packages.nix path
nix flake check                             # or at minimum .#checks.x86_64-linux.webphone-module
buildflow                                   # full gate per AGENTS.md
```

Extra gate (skill Phase 4.7 — the bump touches persistence): copy the production `system.db`, `PRAGMA integrity_check`, boot the NEW build against the copy, diff a read-model dump vs the OLD build (byte-identical modulo timestamps = pass); the retention/trim tests (`transcripts_trim_test.go`, `TrimStream` pins) must stay green. Then the live smoke: `python3 scripts/webphone-smoke.py --bin /tmp/webphone-bin` (all scenarios incl. msglifecycle/restart).

### Train 2 — templ v0.3.1100 → v0.3.1118 (codegen pairing)

```console
nix develop -c go get github.com/a-h/templ@v0.3.1118
nix develop -c go mod tidy
nix develop -c go run github.com/a-h/templ/cmd/templ@v0.3.1118 generate ./internal/web/views/
# the CLI version MUST equal the go.mod library version — both directions red main (AGENTS.md Commands)
nix fmt   # then the authority: nix build .#checks.x86_64-linux.format (treefmt cache trap)
nix develop -c env -u GOEXPERIMENT go test -count=1 ./...
nix build .#webphone.goModules --rebuild   # vendorHash re-pin
nix flake check
```

Review the regen diff (expected: runtime-version bump across `*_templ.go`; ANY codegen-era shape change beyond that = stop and re-read the 2026-11-05 wave precedent). loginpage v4.18.0 accepts templ ≥ v0.3.1100 — 1118 satisfies it; templ-components v1.27.0's own generated code stays (its era-1100 output compiles against the 1118 runtime within the 0.3 series — the `go test` gate proves it).

### Train 3 — modernc.org/sqlite v1.61.1 → v1.62.0 (substrate)

Same tail as Train 1 plus: `nix build .#webphone --system aarch64-linux` (verify by ELF bytes, not exit code), `nix run .#vulnix`, and the **webphone-backup-drill** check (`nix build .#checks.x86_64-linux.webphone-backup-drill`) — snapshot AND restore over the new sqlite build. Full prod-data compat gate as Train 1 (both `system.db` and `usermgmt.db`).

### Train 4 — small SDKs (all additive; one train, five bumps)

`httputil v1.9.0` + `go-paperless v0.5.3` + `go-idempotency v0.6.0` + `sixafter/nanoid v1.67.0` + `golang.org/x/oauth2 v0.38.0`.

```console
nix develop -c go get github.com/larsartmann/httputil@v1.9.0 github.com/larsartmann/go-paperless@v0.5.3 github.com/larsartmann/go-idempotency@v0.6.0 github.com/sixafter/nanoid@v1.67.0 golang.org/x/oauth2@v0.38.0
nix develop -c go mod tidy
```

- Migration (only code change in the sweep): `idempotency_contract_test.go` — construct the store with `idempotency.WithClock` (fake/pinned clock), delete the TimingScale 5 stall budget; run the 13-invariant suite under `-race` repeatedly (the flake it papers over was scheduler-dependent).
- Delivering-layer gates: httputil middleware behavior (CSP/SecurityHeaders/Nonce pins already in server tests — they run in `go test ./internal/server/...`), the paperless seam stays optional/OFF (nil-safe adapter; `sdk.CustomFieldValue` call site unchanged — symbol moved, not changed), SSO tests (`internal/sso`).
- Then the standard tail + vendorHash re-pin + `nix flake check` + smoke.

## Method (reproducibility)

`nix develop -c go list -m -u all` (network, 2026-11-19 ~08:15 CET) intersected with go.mod's direct require block; `scripts/fleet-tags.sh --repos` over all 22 local fleet library clones (local CHANGELOGs vs remote tags; clones verified current with origin); `scripts/bump-preflight.sh <module> <old> <new>` for all 10 behinds; tag-tree `git grep` verification for the two suspicious removal scans; `scripts/ci-verdict.sh 25301790` for the baseline. Vendor hygiene: no `vendor/`, no `go.work`, no `replace` directives in go.mod.
