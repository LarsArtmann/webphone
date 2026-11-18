# SUPERB open-queue closeout — T09/T17/T18 + the heal train

> STATUS: claims verified as of 2026-11-18 19:05. Push/CI statements rot
> fastest; the verification ledger at the bottom carries the shas.

The owner-GO'd pareto plan
([2026-11-14_11-41](../planning/2026-11-14_11-41_SUPERB-open-queue-pareto-plan.md))
is now **23 of 24 tasks done**. The only remainder is T18's prod window
(owner-gated: env credentials + a human caller). The two 14:28 report
§g questions about the VM test are SELF-ANSWERED by T09's landing (see
below); §g-1 (the vendorHash stomp provenance) stays an owner-side
postmortem — the repair itself shipped 2026-11-15 (`0ca42edd`).

## a) Landed this session

| Item                                        | What landed                                                                                                                                                                                                                                                                                                                  | Commit                                 |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| T09 — OIDC lane VM test                     | `webphone-oauth2` GREEN: real stub IdP (loopback https, system-trusted cert), live begin chain (client_id + PKCE S256 + nonce over TLS), cached-provider compose, fresh-process dead-IdP → `sso_error=unavailable`, fail-closed restart cycle (NRestarts + journal + healthz down)                                           | `65299eb9`                             |
| T09 honest findings                         | credential mode is **0440** (systemd LoadCredential; the 0400 guess was wrong — pinned empirically); discovery is boot-lazy but **begin-eager**; the 15-min positive discovery cache composes redirects with the IdP dead                                                                                                    | `65299eb9`                             |
| T17 — architecture diagrams refreshed       | `2026-11-17_10-03-webphone-architecture{,-improved}.d2/.svg` derived from the real `go list` graph (38 production packages, 146 internal edges); old generations kept; INDEX row; the ONE live delta left is the server/html carve                                                                                           | daemon-carried (`1e38ae4c`/`f3ad5396`) |
| T17 evidence: ghost code pin deleted        | `internal/server/api/family_test.go` pinned `push.draft_devices`/`push.draft_send` — codes NO production constructor emits (summary_api wraps `agent.summarize`; the fan is fire-and-forget riding `push.registry`/`push.transport`/`push.*_send`). Deleted + historical note in place; AGENTS § agent-chain prose corrected | daemon-carried (`fd54f189`)            |
| T18 — night-pair verdict rule               | The tracking plan's ledger now defines STABLE mechanically: same pass count × ±25% duration band × ring-clean × noise floor (bar = floor+1pp UP). Both prod-gated TODO rows updated; one owner window covers both                                                                                                            | `9dff7438`                             |
| statix red on origin (MINE, fixed same day) | The caddy-free VM rewrite opened a second `auth.` block — statix repeated-keys failed fast-flake-checks on `1a376563`; one-auth-attrset fix pushed                                                                                                                                                                           | `444b9cbd`                             |
| api-dial tripwire (14)                      | **FIRED**: revisit date passed with the trunk still screened AND the dev-box watcher degraded (pbx-artmann #45/#46 — ssh failing since 11-15). Recommendation flips DEFER → GO per the memo's own logic, but the decision is spend-bearing → routed to the OWNER+ASST sitting (pbx-artmann pareto-v2 N6)                     | `9dff7438`                             |
| Cross-session registry regen                | The concurrent authmap train's two new codes (`authmap.user.lines`, `authmap.user.duplicate_line`) carried the `TestErrorCodeRegistryIsFresh -update` regen per the same-push doctrine                                                                                                                                       | this push                              |

## b) The plan's final tally

T01–T08, T10–T16, T19–T24 **done** (earlier trains + the 14:28 report).
T09 done today. T17 done today. **T18**: local/support halves done
(verdict rule + staged command + self-tests); the wall-clock pair and
the ring-guard positive path are the plan's single remaining item —
owner prod window (the rule makes it judgment-free when it runs).

## c) Attributions during the battery

A concurrent train (authmap user-lines + admin + SSO actor + island
postcall) was mid-flight through most of this session. Attributed, not
touched: the Go build breaks that moved with their waves
(`undefined: usersByExtension` → `views.AdminUserLine` → healed), the
island `postcall-note.test.mjs` failures (341/0 on the suite minus that
file), and the treefmt red on `444b9cbd` (their committed
`internal/server/server.go` snapshot — self-heals as their train lands;
the current tree passes `nix build .#checks.format`). Per the
fix-forward protocol nothing of theirs was reverted or patched.

## d) §g status

- **§g-1 stomp provenance** (2026-11-15 vendorHash placeholder stomp):
  still open, owner-side. The repair + the "stage explicit paths"
  rule shipped; the provenance answer is the daemon-vs-session race
  question the owner may want dissected.
- **§g-2 VM "app reads it" bar**: ANSWERED by construction — the test
  proves more than the ask: the app reads the credential copy, parses
  the lane, discovers against a REAL IdP over TLS, and composes the
  redirect with the secret-bearing config.
- **§g-3 VM caddy posture**: RULED caddy-free inside the fixture (port
  443 belongs to the stub IdP; real ACME would be network noise). The
  caddy-keyed derivations stay covered byte-exactly by
  `nix/module-output-oauth2.golden`.

## e) Verification ledger

- `nix build .#checks.x86_64-linux.webphone-oauth2` — GREEN (fails=0,
  ~39s VM run; the earlier reds were iteration, each attributed in the
  TODO strike).
- `env -u GOEXPERIMENT go test -count=1 ./...` — **0 failures** on the
  final healed tree (`/tmp/go-final.log`, 19:0x).
- `nix build .#checks.x86_64-linux.format` (the ONLY format authority) —
  GREEN.
- `tsc -p . --noEmit` — exit 0.
- Island suite — 347 pass / 2 fail; BOTH failures are the concurrent
  train's LANDED `postcall-note.test.mjs` (announce count 4 vs the
  spec's once-per-episode) — filed as tracker **#253**; the suite minus
  that file is 341/0.
- `nix flake check` — blocked by exactly those two legs during this
  session's runs (island-js = #253; treefmt drifted on their in-flight
  `session_switch_test.go` and has since SELF-HEALED — the format check
  above is green). The oauth2 VM leg itself is green; a full flake
  re-run once #253 closes is the residual.
- docs-gates battery — PASS (all five doc gates + table-damage +
  strike-safety) before this report's push.
- CI: the statix axis on my pushes green (✅ in the `444b9cbd` log);
  the treefmt axis red on `444b9cbd` was their committed snapshot and
  self-healed; origin's current failure verdict = #253's island leg.
