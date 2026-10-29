# SUPERB — owner decision packet (2026-10-29 15:53)

Six rulings, one sitting. Each card is self-contained: decision, evidence,
options, recommendation. Answer inline, e.g.:

> A1: GO · E2EE: GO, store=own-db · pma#348: authorize · tw: keep+release-gate
> config: GO file-strict · triage: pull forward

Nothing here is pre-decided; every card gates a train in
[the Pareto plan](2026-10-29_15-02_SUPERB-pareto-execution-plan.md)
(M04, M05, M07–M11, M19–M20 park behind this packet).

---

## 1. Group rooms — ruling A1 (unlocks M07–M09, M13)

**Decision:** adopt A1 from the
[model spike](2026-10-29_12-20_matrix-group-rooms-model-spike.md):
matrix threads stay peer-keyed for DMs and become room-keyed for groups —
`domain.Address` accepts both `@…` and `!…`; the thread key
`(owner, network, remote)` is unchanged byte-for-byte; groups are
born-inbound only. The one schema change is a nullable, network-neutral
`Message.Sender` (migration v6) that signal groups reuse verbatim later.

**The spike already settled the four sub-decisions** (sender column, receiver
room-name cache over `m.room.name`/`canonical_alias`, outbound sends straight
to the room id, receipts stay DM-only in v1). A2 (all-room-keyed) fragments DM
history; B (participants list) is YAGNI; C (per-peer shadow threads) is
recorded as an anti-pattern.

**Ask:** GO / NO-GO on A1. Effort if GO: three 100-minute tranches (M07–M09),
then signal groups (M13) ride the same Sender design.

## 2. E2EE — ruling P2 (unlocks M10–M11)

**Decision:** adopt P2 from the
[research doc](../research/2026-10-29_12-35_matrix-e2ee-go-story.md):
keep our hand-rolled sync loop as the driver; bridge mautrix-go's
`OlmMachine` for Megolm **decrypt-only** (inbound DMs readable in v1; no
encrypted send).

**Correction now verified at source (F46, 2026-10-29):** the doc's
`(verified)` flag on "NewOlmMachine takes the concrete client type" was an
overclaim at write time; it is TRUE —
`func NewOlmMachine(client *mautrix.Client, log *zerolog.Logger, cryptoStore Store, stateStore StateStore) *OlmMachine`
(`crypto/machine.go`, maunium.net/go/mautrix/crypto). P2's minimal-client shim
is viable; P1 (adopt mautrix.Client wholesale) stays rejected.

**Sub-decisions bundled:**
- **Crypto store home:** own sqlite db (`e2ee.db`, readmodels-adjacent, own
  DDL ladder — mirrors the usermgmt isolation precedent) vs mautrix's sqlstore
  module (one more dep, schema owned upstream). Recommendation: own db.
- **Trust posture:** TOFU with an "unverified device" presence marker; no
  cross-signing UI in v1.
- **Backup:** crypto store joins the backup set + restore drill (F54) —
  losing Olm sessions means losing history.

**Ask:** GO / NO-GO on P2; if GO, confirm store home (own-db recommended).

## 3. pma#348 — who implements the daemon push-leg fix (M04)

**Decision:** the root cause is verified and filed
([#348](https://github.com/LarsArtmann/projects-management-automation/issues/348),
draft at `docs/drafts/pma-push-leg-durability-issue-2026-10-28.md`):
`Stop()` drops pending batches with no drain, `Start()` has no unpushed sweep.
The ask: SIGTERM flush + startup sweep + a shutdown-completion line.

**Stall evidence (the tax):**
| date | event | cost |
|------|-------|------|
| 2026-10-27 | fleet deploy killed the push process | oldest unpushed waited 30m37s; gaps 54m/70m |
| 2026-10-29 | webphone push stall | 62 min (bar invoked 13:05) |
| 2026-10-29 | ~3 stalls across repos | bar invoked again ~15:10 |
| 2026-10-29 | crush-config lessons | sat unpushed ~2 h (pushed itself by 15:53) |

**Ask:** who lands it? (a) owner implements in pma, (b) **authorize me to
implement + PR it in pma** (recommended; ~100 min, fix is fully specified),
(c) leave the watch script + bar ritual as the mitigation. Note the
fleet-wide angle: `core.hooksPath` override in the pma unit is a SEPARATE
known gap — this decision covers only the push leg.

## 4. tw-freshness — flip to blocking? (post-M02)

**Evidence:** the advisory ("rebuild both Tailwind artifacts, diff against
committed bytes") fired on every CI run for ~24 h while the UI train edited
templates (verified on run 40890346682 and predecessors). M02 (commit
`f8c6343`) rebuilt the artifacts, so it is green at HEAD **right now** — but
any concurrent template edit re-drifts it until the next rebuild.

**Options:**
1. **Keep advisory in CI + gate `release.sh` on freshness** (recommended) —
   releases can never ship stale bytes; daily concurrent-train noise stays
   non-gating.
2. Flip the CI job to blocking — every concurrent template edit reds main
   until someone rebuilds; with two+ live sessions this is a tax, not a gate.
3. Keep as-is (advisory everywhere) — accepts that a release could ship stale
   artifacts.

**Ask:** pick 1 / 2 / 3.

## 5. Config typo strictness (L7) — posture

**Evidence:** `config.Load`'s koanf `Unmarshal` silently drops unknown keys
from BOTH the JSON file and `WEBPHONE_*` env — a typo turns a seam off
instead of failing the boot (the outage class that motivated the docpin key
tree; the same reflection helper can feed the allowed set, F22).

**Design (already constrained):** strict (boot-failing, named-key error) for
the FILE; warn-only for ENV (a dozen `WEBPHONE_*` operational vars are not
config keys; a false reject is an outage on every deployment).

**The genuine owner call:** the NixOS module writes a freeform `settings`
map, so a deployment carrying an extra passthrough key today would start
failing its boot. Options:
1. **GO file-strict + env-warn** (recommended) — fail-loud doctrine; the
   module's own keys all pass; a foreign key is a real misconfiguration the
   operator wants named.
2. Env-warn only — zero boot-semantics change, file typos stay silent.
3. Keep silent — accept the outage class.

**Ask:** pick 1 / 2 / 3 (option 1 = M05 unblocked).

## 6. Branching-flow triage — pull the cadence forward?

**Evidence:** 164 live branching-flow findings vs 78 governed baseline
entries (the 10-28 train-start measurement; ~7 new findings per train day —
D1's `session_api.go`, the Deps-grown `server.go`, the chat trains' matrix
sites). Findings are non-gating warnings under the SDK severity cap, but the
gap is the baseline rot the 10-28 retirement of `.buildflow-baseline.json`
was supposed to end: suppression claims nothing and looks reviewed.

**Options:**
1. **Pull forward to this week** (recommended; M19–M20 are already planned
   as two 100-min tranches) — the standing cadence says next re-measure
   2026-11-05; six more trains of drift first.
2. Honor 2026-11-05 — accept ~40 more ungoverned findings by then.

**Ask:** pull forward / honor the date.

---

*Packet assembled by the 2026-10-29 SUPERB execution session; evidence
verified same-day (CI runs, source fetches, repo states).*
