# Owner sitting pack — every open ruling in one dated agenda (2026-11-07)

> **REVISED 2026-11-08 02:50 CET** (second edition): the first draft was
> audited before delivery and had real defects — an E-group numbering hole, an
> answer template that lagged the IDs, ten ROADMAP rulings missing, and the
> 2026-09-22 briefing's 38 rows unconsolidated. All corrected here. One item
> moved from "question" to "incident": B4 was executed as a fix this session
> because it was a bug, not a choice.

**Status:** PACK (answers owed) · **Owner time to clear it:** ~60 min ·
**Evidence verified at** the paperless-parity train (2026-11-08 02:50): every
`Evidence:` path in the source index was `-e`-checked this session, the config
keys cited exist in `internal/config/config.go`, and the buildflow posture
calls carry their measured numbers (124-entry baseline, 3 error findings found
and FIXED this session, findings gate now green).

Why this file exists, and what it replaces: open owner questions accumulated in
**four** homes that no session can join by hand — `ROADMAP.md` §Open questions
(~50 bullets, several already resolved in code and only needing ratification),
the six cards of the
[2026-10-29 packet](2026-10-29_15-53_SUPERB-owner-decision-packet.md)
(delivered as a file, never answered), the three posture calls the
[post-v9 execution plan](2026-11-07_18-46_SUPERB-post-v9-execution-plan.md)
raised mid-flight, and the ~38 briefing rows of the
[2026-09-22 owner-calls briefing](2026-09-22_13-50_owner-calls-briefing.md)
that were never folded into a decision surface. This pack is the ONE sitting.
The packet stays as the evidence appendix it already is; **answers go here**
and this file becomes the record.

Rows are ordered by blast radius, not by age: A and B are where a wrong answer
costs users or data, C and D cost maintenance, E is outside this repository.
Rows tagged **[ratify]** are already implemented — the answer only blesses what
ships today, and "yes" is the do-nothing option. Section §F is the bulk table:
38 one-line ratifications answerable with a single "all yes" if you trust the
recommendations.

## How to answer

Answer inline in the block below, one line per ID. `GO` / `NO-GO` for binaries;
the option number for multi-way calls; `defer` is always legal and always means
"the status quo continues to ship". For §F either `F: all yes` or pick rows.

```text
A1:   A2:   A3:   A4:   A5:   A6:   A7:
B1:   B2:   B3:   B4:   B5:   B6:   B7:   B8:
C1:   C2:   C3:   C4:   C5:   C6:   C7:   C8:   C9:   C10:  C11:
D1:   D2:   D3:   D4:   D5:   D6:   D7:   D8:   D9:   D10:  D11:  D12:  D13:  D14:  D15:  D16:  D17:  D18:
E1:   E2:   E3:   E4:   E5:   E6:   E7:   E8:   E9:   E10:  E11:  E12:  E13:  E14:
F:    (all yes, or list exceptions)
```

---

## A. Product & UX — what the user sees (7 rulings)

### A1. Fax thumbnails: ratify the DECLINE **[ratify]**

The tab shows a filename and a download link, never a preview. Previewing would
mean pdf.js (+1.2 MB always-loaded, against the lean-serving premise) or a Go
rasterizer that does not exist honestly.

- **Recommend:** ratify the DECLINE, with the re-open trigger already written
  ("a pure-Go PDF rasterizer becomes a real dependency").
- **Cost of the other branch:** +1.2 MB on every page load for every user,
  including the ones who never open Fax.
- **Evidence:** `docs/planning/2026-10-02_09-05_fax-thumbnail-decline-note.md`

### A2. Loopback gateway: should a dev send resolve to `delivered`? **[ratify]**

Loopback (no PBX) marks outbound SMS `sent` and outbound fax `transmitted`.
`sent` is the honest terminal state — nothing downstream will ever confirm
delivery — while fax's `transmitted` has a real egress event behind it.

- **Recommend:** keep `sent`. Ratifying makes the honesty rule explicit so a
  future contributor does not "fix" the asymmetry.
- **Cost of the other branch:** the loopback dev flow would display a state the
  transport cannot confirm, which is the exact class the island-honesty
  contract forbids.
- **Evidence:** AGENTS.md § Gateway seam; `internal/gateway`

### A3. Provider-failure retryability

Today only `transient` failures render a Retry affordance; a provider verdict
(bad number, blocked, carrier reject) is terminal and the user re-composes.
Fax `Resend` is the counter-example: it reuses the stored document.

- **Recommend:** keep provider verdicts terminal, and state the rule in
  `docs/error-contract.md` so the typed `FailureKind` machine is not read as an
  oversight.
- **Cost of the other branch:** a Retry button on a hard reject trains users to
  press it, and it will fail identically — worse UX than no button.

### A4. Missed-call semantics + badge clearing

Two sub-questions, one recommendation. (a) A deliberately REJECTED incoming
call does not count as missed today (`wp:call-missed` never fires). (b) The
missed badge clears on History tab click, not when the island's _Recent calls_
panel opens.

- **Recommend:** (a) keep REJECT out of "missed" — the user was present and
  chose; (b) also clear on the island panel opening, since the user has then
  seen the list. (b) is ~10 min of island work.
- **Cost of the other branch:** (a) counting a reject as missed inflates the
  badge the user must then clear manually; (b) leaving it means a badge
  survives a panel the user already read.

### A5. Search `?q=` URL semantics

A typed message search is ephemeral today. Deep-linking it would push `?q=`
into history so back-button unwinds filters.

- **Recommend:** adopt `?q=` (deep-linkable). It is the platform-native
  behaviour for a search box and costs one `pushState`.
- **Cost of the other branch:** a user cannot share or bookmark a search, and
  back-button leaves the tab entirely instead of unwinding the filter.

### A6. Recordings in the webphone UI

The stack records calls server-side (stereo WAV, operator basic auth); this
repo has NO recording capability, only CDR rows. Putting playback in the UI is
a product decision with a consent/jurisdiction dimension.

- **Recommend:** NO-GO for now, recorded with its trigger ("the stack exposes
  a per-extension recording list with a consent posture"). The operator
  already has the recordings; the missing piece is a permission model, not a
  player.
- **Cost of the other branch:** shipping playback without per-extension
  consent scoping would let any session with a CDR row reach another party's
  audio.

### A7. UI/UX execution shape (the 120-idea catalogue)

Does the catalogue and its Pareto plan graduate into ROADMAP/TODO clusters, is
mobile first-class or graceful-degradation, and which SEAM workstreams
(messaging richness M21, pin/archive/mute M22) need a server-side design
before UI work starts?

- **Recommend:** graduate the clusters into ROADMAP (they are parked in a
  timestamped plan file today, which no sweep owns); mobile stays
  graceful-degradation (the PWA-LITE verdict + the container-query layout
  already deliver it); M21/M22 design first only where a server seam is
  missing (scheduled sends is the known NO-GO, already recorded).
- **Cost of the other branch:** the catalogue stays a dead file and mobile
  regressions get adjudicated per-bug instead of by policy.
- **Evidence:** `docs/planning/2026-10-01_03-53_SUPERB-ui-ux-pareto-plan.md`

---

## B. Privacy, identity & data — what we store and send (8 rulings)

### B1. Group rooms — ruling A1 (packet card 1)

Matrix threads stay peer-keyed for DMs and become room-keyed for groups;
`domain.Address` accepts both `@…` and `!…`; the thread key is unchanged;
groups are born-inbound only. One schema change: nullable network-neutral
`Message.Sender` (migration v6), which signal groups later reuse verbatim.

- **Recommend:** GO. The spike already settled the four sub-decisions (sender
  column, room-name cache over `m.room.name`, outbound straight to room id,
  receipts DM-only in v1).
- **Cost of the other branch:** group conversations stay impossible on matrix
  and signal — the chat seam ships DM-only indefinitely.
- **Unlocks:** M07–M09 (3 × 100 min), then M13 (signal groups).
- **Evidence:** `docs/planning/2026-10-29_12-20_matrix-group-rooms-model-spike.md`

### B2. E2EE — ruling P2 + the crypto-store home (packet card 2)

Keep the hand-rolled sync loop; bridge mautrix-go's `OlmMachine` for Megolm
**decrypt-only** (inbound DMs readable, no encrypted send). The
`NewOlmMachine` signature claim was re-verified at source (F46): P2's
minimal-client shim is viable, P1 (adopt `mautrix.Client` wholesale) stays
rejected. Bundled sub-decisions: crypto store in its own sqlite db (`e2ee.db`,
own DDL ladder, mirroring the usermgmt isolation precedent) vs mautrix's
sqlstore module; TOFU trust with an "unverified device" marker and no
cross-signing UI in v1; the crypto store joins the backup set and the restore
drill.

- **Recommend:** GO on P2, crypto store in its own db, TOFU, backups included.
- **Cost of the other branch:** encrypted rooms stay unreadable, which is most
  of the rooms a real account is in.
- **Unlocks:** M10–M11.
- **Evidence:** `docs/research/2026-10-29_12-35_matrix-e2ee-go-story.md`

### B3. Config typo strictness (L7) — packet card 5

`config.Load`'s koanf `Unmarshal` silently drops unknown keys from both the
JSON file and `WEBPHONE_*` env, so a typo turns a seam off instead of failing
the boot. Design already constrained: strict (boot-failing, named-key) for the
FILE, warn-only for ENV. The unknown-key WARN already exists
(`internal/config/unknown.go`), so the flip has a measured blast radius.

- **Recommend:** option 1 — GO file-strict + env-warn. One known consequence:
  the NixOS module writes a freeform `settings` map, so a deployment carrying
  a passthrough key would start failing its boot; grep the deployment's
  `settings` for keys absent from README before flipping.
- **Cost of the other branch:** the outage class stays — a misspelled
  `gateway.webhook_secre` silently disables webhook auth. This session's
  paperless bug is the class in miniature: a dropped key failed with a
  misleading message instead of naming the unknown key.
- **Unlocks:** M05.

### B4. Token-hygiene parity — CLOSED as a bug, recorded as the standing rule **[ratify]**

The ROADMAP asked whether the newer secrets need `*_file` variants. The
session that wrote this pack found the answer was already "yes" and
UNFINISHED: `crm.token_file` shipped with the NixOS module train, but its seam
twin `paperless.token_file` did not exist — the koanf decode silently dropped
the key and validation failed with a misleading message. Fixed this session
(struct field, resolve-before-validate, both-sources rejection, four-case test
pin, docpin row, CHANGELOG line).

- **Recommend:** ratify the standing rule the fix encodes: **every
  bearer-token seam accepts the systemd credential form, and a new token seam
  lands with its `_file` twin in the same commit.**
- **Cost of the other branch:** the next token seam repeats the asymmetry.
- **Evidence:** `internal/config/config.go` (the Paperless struct + the
  resolve block), `internal/config/config_test.go`
  (`TestLoadResolvesPaperlessTokenFile`), README config table

### B5. Owner-scoping posture **[ratify]**

Every store query is extension-scoped; the ONLY cross-owner reads are the
deliberate `*Any`/`ListAll*` seams, admin-gated and pinned (store + readmodels
regression suites now cover them, this week's P1-1).

- **Recommend:** ratify the posture (owner-scoping is the rule, `*Any` is the
  audited exception) so future reviews cite a decision instead of re-deriving
  it from three test files.
- **Cost of the other branch:** none today; the cost is recurring review time
  re-proving the same invariant each audit.

### B6. Sessions credentials at rest — cadence

The session row carries the extension's directory password (the `/phone-api`
proxy needs it), swept on read and on Create. Baseline review exists; next
scheduled re-review 2026-12-30 or on the next session-storage change.

- **Recommend:** keep the cadence, and add the trigger half explicitly (any
  session-store schema change re-opens the review the same day).
- **Evidence:** `docs/reviews/2026-09-30_sessions-credentials-at-rest-review.md`

### B7. Existing prod data vs the `UMask=0077` tightening

Files written before the hardening carry group/other bits.

- **Recommend:** one-time `chmod -R go-rwx` in the next ops window, followed by
  a fresh backup. New-files-only leaves a two-tier data directory whose older
  half is readable by the `users` group.
- **Cost of the other branch:** a permanent split posture nobody can reason
  about later.

### B8. `/healthz` CRM criticality — decided this week, needs blessing **[ratify]**

Measured: an unreachable CRM yields `503` +
`{"status":"degraded",…,"crm":{"status":"fail"}}`. Kept CRITICAL deliberately,
because it is the only surface reporting a revoked CRM token when the
dashboard is off, nothing routes on readiness (unit `Type=simple`, Caddy
`remote_ip` fences), and the go-health probe still grades the leg
non-critical (WARN) — two audiences, both true.

- **Recommend:** ratify "criticality is per surface, and this is the strict
  one".
- **Cost of the other branch:** making it non-critical removes the only alarm
  that fires when the token dies.
- **Evidence:** `internal/app/health_crm_test.go`; AGENTS.md § Architecture
  invariants (`/healthz` bullet)

---

## C. Infrastructure & deployment — what runs where (11 rulings)

### C1. The pma push-leg fix: authorize the PR? (packet card 3)

Root cause verified and filed upstream: `Stop()` drops pending batches with no
drain, `Start()` has no unpushed sweep. Measured tax: 30m37s + 54m + 70m +
62 min stalls across four repos in three days.

- **Recommend:** authorize the fix + PR in pma (~100 min). The watch script
  and the 60-min bar are a mitigation that costs a session minute each time it
  fires.
- **Cost of the other branch:** the stalls continue; every one of them costs
  more than the fix.
- **Evidence:** `docs/drafts/pma-push-leg-durability-issue-2026-10-28.md`

### C2. `tw-freshness`: advisory forever, or gate the release? (packet card 4)

The advisory fired on every CI run for ~24 h while templates were edited. It
is green at HEAD right now.

- **Recommend:** option 1 — keep the CI job advisory, and gate `release.sh` on
  freshness. A release can then never ship stale Tailwind bytes, while a
  concurrent template edit does not red main.
- **Cost of the other branch:** option 2 (blocking) turns every concurrent
  session into a main-red incident; option 3 accepts shipping stale CSS.

### C3. HSTS on `pbx.artmann.tech`

The `caddy.hsts` option ships opt-in.

- **Recommend:** GO once https-only is proven on that host; enable with
  `maxAge` starting at 1 day and climbing. HSTS is hard to walk back, so the
  ramp is the point.
- **Cost of the other branch:** staying off leaves a downgrade window on every
  first visit.

### C4. Where may CI skip? The docs/test-only subset policy

Which CI legs may be skipped for a docs-only or test-only commit, and which
must never be. (Routed from a 2026-10-06 report; it lived only in a
timestamped file.)

- **Recommend:** the formatter gate is NEVER skippable (the reds that
  motivated the row were formatter reds); `test-race` and the browser E2E may
  be skipped on a docs-only diff; everything else runs.
- **Cost of the other branch:** an unprincipled skip list reintroduces
  precisely the red-main class the 10-06 incident produced.

### C5. Should coverage have teeth?

The BuildFlow gate passes at ANY coverage level.

- **Recommend:** a floor on the union of hand-written packages only (not
  `-coverpkg` unions, which count generated `_templ.go`), set at today's
  measured value minus a small margin, so it ratchets down but never silently.
  This is fleet-wide (the step lives upstream), so the answer also decides
  whether to file the upstream ask.
- **Cost of the other branch:** coverage stays informational and erodes.

### C6. `--border-strong` vs WCAG 1.4.11

The token measures 1.86:1 (light) / 1.53:1 (dark) against its surface; 1.4.11
wants 3:1 for non-text UI.

- **Recommend:** ACCEPT AND RECORD — the border is decorative separation and
  the surrounding affordances carry the contrast. A retune touches every input
  and panel edge, which is a visual-identity call, not a colour tweak.
- **Cost of the other branch:** a full visual-identity pass on every edge in
  the product, for a criterion whose intent (distinguishability) the form
  already meets.

### C7. `backup.destDir` / `dataDir` nesting legality

systemd `StateDirectory` splits on spaces; nesting is currently neither
forbidden nor documented.

- **Recommend:** forbid by assertion (the module already asserts `dataDir`
  lives under `/var/lib/`), because a silent nesting produces a backup that
  excludes its own source or recurses into it.
- **Cost of the other branch:** an operator configures it, and the failure
  appears at restore time — the worst possible moment.

### C8. Stack verification timing for the hardened module

Ride the per-train relock, or relock + browser E2E now?

- **Recommend:** ride the per-train relock (the default). The unit environment
  changed but the packaging did not, and a standalone E2E run costs a full
  stack-VM cycle for one boolean.
- **Cost of the other branch:** ~one session of stack-lab time per module
  change.

### C9. One-time force-push ratification

During the v2.6.0 fold a `--force-with-lease` was used once on a self-authored
commit ~1 minute after pushing it (the daemon raced `git add`).

- **Recommend:** ratify the narrow exception (own commit, seconds old,
  lease-protected) and write it into the runbook as a shape, not a permission.
- **Cost of the other branch:** the runbook says "never", an exception
  happened, and the record is ambiguous — which is how the next exception gets
  invented.

### C10. Two sessions on one mandate — the protocol (plan posture call 3)

Two Crush sessions worked this repository concurrently this week. Measured
cost: three narrative commit messages eaten by `chore: auto-commit`, one
duplicated status report, and repeated re-verification of shared files.

- **Recommend:** one session owns the mandate; the second takes a disjoint
  scope (a different repo, or a non-overlapping read-only audit). If a second
  session is unavoidable: it never edits `TODO_LIST.md`, `CHANGELOG.md`,
  `AGENTS.md` or `docs/status/` while the owner session is live, and gate
  collisions resolve under the flock convention of D18.
- **Cost of the other branch:** the tax above repeats on every train.

### C11. IO-storm source investigation (owner-level host access)

The box tripped the no-heavy-builds gate 6× in 60 min once (freeze class); a
`/proc/pressure/io` watch during the next window would name the source.

- **Recommend:** park a 5-line PSI watcher in the next ops window (assistant
  has no host access; this is an owner-terminal ritual like the mic check).
- **Cost of the other branch:** every future "load ate the benchmark" verdict
  stays an inference.

---

## D. Tooling & process — how we work (18 rulings)

### D1. Art-dupl ritual baseline: `-t 1` or `-t 3`? (FIFTH surfacing — close it)

`-t 1` is the ratified CI baseline (deepest sweep, zero harmful clones);
`-t 3 --type-aware` is the ritual baseline (43–45 shown groups, mostly idiom
noise, triage 1 extract / 3 accept).

- **Recommend:** keep BOTH and stop asking: `-t 1` is the gate (CI), `-t 3` is
  the ritual reading. They answer different questions, and the confusion is
  that they were ever described as alternatives. Ratifying retires the FIFTH
  surfacing of the same question.
- **Cost of the other branch:** picking one loses either the deepest gate or
  the readable ritual.
- **Evidence:** `docs/dedup-registry.md`

### D2. Helper micro-test bar

Must a one-home helper land ONLY with a micro-test pinning its contract, or is
suite-level coverage acceptable with micro-tests batched later?

- **Recommend:** micro-test required for helpers that are (a) byte-pinned in a
  view test or (b) named in AGENTS' one-home list; batched for the rest. The
  bar becomes a rule with a boundary instead of an unenforceable absolute.
- **Cost of the other branch:** "always" blocks trivial helpers; "never" loses
  the contract pins AGENTS already relies on.

### D3. Store `Must*` panic-on-corrupt policy

Scans call `MustParsePhone`/`MustContactID`/`MustParseExtension` on raw DB
strings; a hand-edited row panics the server at query time.

- **Recommend:** degrade to an error for the ROW (log + skip), keep `Must*`
  for write-path invariants. A corrupt row should cost one row, not the
  process — and the readmodels legacy import already proves the
  skip-and-demote pattern works (`legacy_import_test.go`, pinned this week).
- **Cost of the other branch:** one bad row is a denial of service on the
  whole extension forever.

### D4. Concurrent-session breakage policy

When another session's committed-but-broken code blocks the shared gate: fix
and commit immediately, or keep hands off and report?

- **Recommend:** unblock the gate immediately IF the fix is mechanical and
  obviously local (format, generated artifact, one-line import), otherwise
  report and leave. "Hands off" leaves main red for hours; "always fix"
  collides with in-flight edits.
- **Cost of the other branch:** uncodified, each session improvises and each
  improvisation is reviewed as if it were policy.

### D5. Daemon hardening: exclude `docs/status/` + `docs/planning/`, gate before sweeping

The heuristic daemon has twice swept half-written reports and has three times
committed a non-compiling tree.

- **Recommend:** both asks, upstream in pma. This repo added a local
  mitigation already (`whitespace-drift.sh` GATE 1 + `node --check` GATE 2 in
  `.githooks/pre-commit`) — but that hook is bypassed by the daemon's own
  `core.hooksPath` override, which is D6.
- **Cost of the other branch:** the class repeats, and each instance costs a
  red-main attribution.

### D6. `core.hooksPath` override in the pma unit

The daemon unit exports `core.hooksPath=/var/empty` via env-level git config,
so every daemon commit runs NO hook — enabling hooks re-binds the daemon
fleet-wide.

- **Recommend:** YES, enable hooks in the daemon, scoped to the two cheap
  gates (fmt-drift, syntax). The alternative is that every guard this repo
  builds is advisory against its most frequent committer.
- **Cost of the other branch:** guards that do not guard.

### D7. Push-lag policy: at what lag is the pusher BROKEN, not slow? **[ratify + extend]**

The ratified bar is 60 minutes from the oldest unpushed commit
(`scripts/push-stall-watch.sh`), with a warn-only 30-min line available.

- **Recommend:** ratify 60 min as BROKEN and add the 30-min WARN line — the
  2026-10-14 stall recovered at ~40 min and was invisible to the bar.
- **Cost of the other branch:** a 40-minute stall stays silent, as it did.

### D8. Session phase-boundary hand-pushes **[ratify]**

Three hand-pushes of daemon-swept content followed runbook precedent but
overrode a hard never-push rule.

- **Recommend:** ratify as standing session duty, bounded by: main must be red
  with a verified fix in hand, or the 60-min bar met. That is already how it
  is practised and already written in AGENTS.
- **Cost of the other branch:** an undocumented practice that the runbook
  contradicts.

### D9. `.buildflow-baseline.json` — per-finding accepts **[ratify]**

The file went 78 → 124 findings this week with 46 `file-size-check` accepts,
chosen per-finding (so a NEW oversize file still gates) rather than per-rule
(a silent suppression). `buildflow baseline add` is the only writer; the file
is regenerated only from THIS checkout (host-absolute ids, upstream #13).

- **Recommend:** ratify per-finding accepts as the shape. D10 is the separate
  real question.

### D10. `file-size-check`'s hardcoded 350-line cap (plan posture call 2)

46 files exceed it. Accepting them per-finding keeps the gate meaningful but
freezes a number nobody chose.

- **Recommend:** push upstream for a configurable cap, and keep the 46 accepts
  meanwhile. A repo-wide cap on Go files is a reasonable default but an
  unreasonable constant.
- **Cost of the other branch:** skipping the step (even visibly) removes a
  gate that costs nothing and might catch something.

### D11. Archive-residue policy — per-row evidence or transcribed verdict? (plan posture call 1)

~71 ratcheted findings remain in the pre-canon ARCHIVED corpus, plus ~1600
un-inventoried §a/§b rows in the live reports. The v9 recipe (per-row
evidence, ~15 min/file) is expensive and was proven on ten files.

- **Recommend:** per-row evidence for the LIVE reports (the 1600), and
  transcribed-verdict-per-cell for the ARCHIVED corpus, which is CLOSED
  HISTORY (its 71 findings are ratcheted, never re-litigated). This splits the
  cost where it buys something.
- **Cost of the other branch:** a blanket "transcribe a verdict" for live
  reports would mark evidence nobody produced — the exact vacuous-PASS class
  this train closed.

### D12. The docs-gates defects — file upstream, or vendor a local shim?

Four found this week: DF brittleness (a token flips "generic" at a living-doc
line count, so the same row flips verdict as docs grow), standalone verbs
ignore `.docs-gates.json`, an escaped pipe inside a table cell mangles a
strike, and an unvalidated verdict cell accepts garbage.

- **Recommend:** file upstream (drafts ready) AND keep the local guards
  already shipped (`strike-safety.py`, the `generic_df` pin, the wrapper
  passing `--audit-max` explicitly in `--check`). Both, because the local
  guards are cheap and the upstream fix has no date.
- **Cost of the other branch:** vendoring alone forks the tool; filing alone
  leaves the sweep unsafe until upstream ships.

### D13. Daemon-format coupling note: which home?

The auto-commit daemon's reflow and the gates' formatter are coupled; the note
has three candidate homes (a test comment, the error-contract preamble,
AGENTS).

- **Recommend:** AGENTS — it is the only home every session reads before
  touching a formatter, and the coupling has bitten three times.
- **Cost of the other branch:** a test comment is read by whoever runs that
  test; the error contract is read by whoever debugs an error.

### D14. AGENTS.md restructure permission (over the 377-line BuildFlow cap)

Compaction moves war stories to `docs/lessons.md` and rewrites sections every
concurrent session depends on.

- **Recommend:** grant it, and do it in the session that owns the file (see
  C10). The cap is enforced, so today the file either shrinks or the gate
  drifts.
- **Cost of the other branch:** the cap is a preflight that will eventually
  block a train at an inconvenient moment.

### D15. Domain-language glossary: ratify or delete

`docs/DOMAIN_LANGUAGE.md` exists with a memo appended 2026-10-14; the
declarative trains made its vocabulary load-bearing in code.

- **Recommend:** ratify. Deleting a glossary whose terms are now type names
  would make the code harder to read, not the docs shorter.
- **Cost of the other branch:** the vocabulary is re-derived from code by
  every new session, at ~30 min each.

### D16. Tier-2 family-adoption intent

`go-error-family` landed in go.mod via daemon sweep with no narrative — is the
migration project starting, or is the dep swept-but-unused until the owner
calls the family shape?

- **Recommend:** swept-and-used is the actual state (every new error path
  already uses `errorfamily.New*`/`Wrap*` per the tier-2 bar), so ratify "the
  family is the standing error shape" and close the ambiguity — no migration
  project exists because the codebase already migrated.
- **Cost of the other branch:** a future train re-asks the question or
  "completes" a migration that already happened.

### D17. erraudit `tree` same-name-dedupe display defect — file upstream?

It hid 3 of 7 sentinels once (pick site flips between runs); erraudit is the
owner's own tool.

- **Recommend:** yes, file it upstream in the erraudit repo with the
  two-runs-two-pick-sites reproduction. Owning the tool makes the cost of the
  fix near zero and the display defect actively misleads triage.
- **Cost of the other branch:** every erraudit triage re-derives which
  findings are real.

### D18. Cross-session gate protocol: flock, staggered schedules, or sustained quiet?

From the v2.6.0 release night (load 60–174 from parallel agent sessions): a
flock convention (`/tmp/webphone-gates.lock` holding pid + scope) so
concurrent sessions yield instead of colliding, staggered schedules, or
"sustained quiet" as the accepted de-facto rule.

- **Recommend:** the flock convention, owned by a 10-line wrapper around
  buildflow + `nix flake check` invocations. Staggered schedules need a
  scheduler nobody operates; sustained quiet is what already fails under
  contention.
- **Cost of the other branch:** the next release night under parallel sessions
  repeats the load-174 incident.

---

## E. Cross-repo & carrier — outside this repository (14 rulings)

### E1. Carrier-leg verdicts: inbound-MMS isolation test, Google Voice, and `+14084757593` ownership

The test was never recorded anywhere. MMS ships and is webhook-tested today,
so this is a verdict-recording call, not a build gap.

- **Recommend:** run the isolation test once on a real carrier line, then
  record the GV verdict and number ownership in FEATURES + AGENTS. Needs a
  carrier phone.
- **Cost of the other branch:** an unanswerable question in a doc that a
  future reader will trust.

### E2. Release numbering + the bridge's "webphone >= 2.8" claims

- **Recommend:** reword the bridge claims to commit/date form. Webphone's own
  AGENTS line is already commit-pinned; making both forms match removes a
  version claim nobody remembers to bump.
- **Cost of the other branch:** cutting a release purely to satisfy a doc
  string.

### E3. GitHub Release objects per tag

v2.0.0 has a tag, no Release object; later tags do.

- **Recommend:** Release objects for user-visible trains only, generated from
  the CHANGELOG's `[Unreleased]` fold. Tags remain the machine contract.
- **Cost of the other branch:** either every tag gets an empty Release (noise
  on the repo page) or the changelog has no public face.

### E4. Go module-path policy for v2+ tags

`/v2` suffix vs a NOT-DO record.

- **Recommend:** NOT-DO record — a flake-consumed application never runs
  `go get` on itself, so a `/v2` path buys nothing and breaks every flake
  input URL that names the module.
- **Cost of the other branch:** a module-path migration with zero consumers.

### E5. Handler gating: keep the deliberate dual layer? **[ratify]**

`requireSession` helpers plus `Sessions.Require` on `/events` and
`/phone-api/`.

- **Recommend:** keep. The middleware wiring covers the routes whose handlers
  cannot call a helper (SSE, the byte proxy), so the redundancy is by shape,
  not by accident. The contract test that allowlists exactly four 401 writers
  is the proof it stays honest.

### E6. Rate-limit keys: does the stack sanitize `X-Forwarded-For`?

Gates the flip from port-stripped peer-host keys to
`KeyExtractorFromClientIP`.

- **Recommend:** measure on the stack before deciding; until then keep the
  safe default. The answer is one `curl` with a spoofed header against the
  vhost.
- **Cost of the other branch:** flipping blind means one client can poison a
  rate-limit bucket by spoofing.

### E7. Stack `/health` dashboard exposure **[ratify resolved]**

- **RESOLVED 2026-11-04** — the fence is the proxy contract (`forwardAuth`
  under operator SSO, loopback-only otherwise) plus a boot WARN naming it;
  `WithMiddleware` was declined because the app is proxy-blind.
- **Recommend:** ratify the recorded decision so the ROADMAP row can close.
- **Evidence:** `docs/planning/2026-11-04_23-45_dashboard-fence-decision.md`

### E8. SUPERB prod cutover timing + `staticFallbackExtensions`

Should prod's first static list name operator/ring-group extensions
explicitly?

- **Recommend:** yes, name them. A fallback list that omits the operator's own
  extension is a fallback that cannot route the one call that matters.

### E9. F49 scope: eval-level rollback proof or a full in-VM flip drill?

- **Recommend:** eval-level (byte-identical static golden) as the gate, with
  the in-VM drill as a one-time manual ritual recorded in the runbook. A
  ~10-minute CI tax on every run for a static/xml_curl flip that changes
  monthly is not priced right.
- **Cost of the other branch:** ~10 min CI + a second toplevel, forever.

### E10. Stack-side TEMP-DIAG answer-phase dump in the browser E2E — keep or revert?

- **Recommend:** revert. It was diagnostic scaffolding, and a dump that stays
  becomes an unowned output nobody reads.

### E11. `gateway.attachment_limit`: local cap knob or the bridge's honest 422?

- **Recommend:** ratify the honest 422 — the bridge knows the real limit and
  names it. A local cap duplicates a limit the bridge owns, and the two will
  drift.

### E12. Sniff-fallback lifespan in the bridge: permanent or delete-after-deploy?

- **Recommend:** keep forever, with a comment naming it as permanent seam
  robustness rather than a deploy shim — a fallback with a deadline gets
  deleted by someone who does not know why it exists.

### E13. Stack pin vs tag on release trains

The v2.6.0-era stack lock pinned post-tag `main` while the tag pinned the
release commit; that matches the documented "ride main" decision and the
v2.5.0 precedent (lock ≠ tag), but re-pinning the stack to the tag itself for
release trains is the alternative.

- **Recommend:** keep ride-main. A release train's fixes land on main within
  hours anyway; a tag-pin step adds a manual action to every release for
  reproducibility the flake lock already delivers.
- **Cost of the other branch:** a manual re-pin step per release that nobody
  remembers at 2 a.m.

### E14. Release cadence / pin policy **[ratify]**

Cut a train when `[Unreleased]` accumulates a user-visible theme (the v2.4.0
fold pattern), deploy once per train via the single owner command, keep the
stack riding `main` until a hotfix cadence emerges.

- **Recommend:** ratify as written; the recorded revisit trigger ("a security
  fix ever needs to ship inside an hour") is the only thing that re-opens it.

---

## F. Bulk ratifications — the 2026-09-22 briefing's open rows, one line each

Answer `F: all yes` (the recommendation in every row is the status quo or a
no-op) or list exceptions. Rows carrying the same ruling as a lettered ID are
marked "same as" — one answer covers both.

| #   | Ruling                                                                        | Recommendation                                                                                                                 |
| --- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| F1  | Registry ratification: `docs/dedup-registry.md` is THE home for clone rulings | yes                                                                                                                            |
| F2  | `settingsRow` build-or-retire                                                 | retire (the row keeps outliving the helper — FOURTH surfacing)                                                                 |
| F3  | Suppression-bucket doc: what `art-dupl` "filtered suppressed" hides           | document in dedup-registry, not upstream                                                                                       |
| F4  | Dedup suppression-SCOPE: shown-only vs periodic suppressed-set audit          | periodic quarterly audit (already the AGENTS rule)                                                                             |
| F5  | Webhook 400 body-text dependents                                              | none exist; ratify the current validation precedence as final                                                                  |
| F6  | `msg/`→`message/` idem-key rename                                             | ratify (in-memory only, no persistence, no external reader)                                                                    |
| F7  | `auth:`-prefix × daemon-sweep interface                                       | manual `auth:` commit before the sweep window; a daemon `chore:` is acceptable only when the pinned session suite provably ran |
| F8  | devShell tool additions (dprint/prettier/ruff/lychee)                         | yes — silences 4 buildflow warnings                                                                                            |
| F9  | lychee authenticate-vs-exclude policy                                         | exclude archived dirs (the buildflow warn names the exact config)                                                              |
| F10 | Docs-only express-push ratification                                           | yes, bounded by the 60-min bar                                                                                                 |
| F11 | Mixin/dupe detector-policy home                                               | dedup-registry sweep-log only (no AGENTS promotion)                                                                            |
| F12 | `config.CRM`/`Paperless` URL+Token twin standing ruling                       | keep per-seam both-or-neither (baselined; detector noise routed upstream)                                                      |
| F13 | `vcard.Card`/`SharedContact`/`apiSharedContact` triad rejection               | ratify                                                                                                                         |
| F14 | `errorfamily.Join` gap                                                        | keep the 4 reasoned nolint aggregate sites; no upstream constructor                                                            |
| F15 | Routed-as-resolved → archive for owner-gated plans                            | ratify (four plans already archived this way)                                                                                  |
| F16 | Train-level `v` verdict markers as the annotation hash bar                    | ratify (six sweeps running)                                                                                                    |
| F17 | Check-rows marker-cell baseline for the 37 pre-canon archived files           | accept as-is (closed history)                                                                                                  |
| F18 | Push-lag threshold policy                                                     | same as D7                                                                                                                     |
| F19 | Release cadence / pin policy                                                  | same as E14                                                                                                                    |
| F20 | AGENTS restructure permission                                                 | same as D14                                                                                                                    |
| F21 | Existing prod data chmod / re-backup                                          | same as B7                                                                                                                     |
| F22 | `destDir`/`dataDir` nesting legality                                          | same as C7                                                                                                                     |
| F23 | Stack verification timing                                                     | same as C8                                                                                                                     |
| F24 | One-time force-push exception shape                                           | same as C9                                                                                                                     |
| F25 | Release.sh load-gate default (`load >= 8` refuses)                            | ratify the conservative default; E2E auto-retry rejected (masks regressions)                                                   |
| F26 | g2 KVM-timeout                                                                | keep the current VM-test timeout; raise only on two consecutive infra-flake runs                                               |
| F27 | Templ-components history-blemish disclosure                                   | keep the acknowledged-and-fixed disclosure posture                                                                             |
| F28 | Self-send train-C semantics                                                   | ratify current (owner-DID self-sends refused locally after persist)                                                            |
| F29 | CRM policy trio (names/deep-links/context surfaces)                           | ratify as shipped (data-mesh, read-only, context on single-contact surfaces)                                                   |
| F30 | Missed-call REJECT semantics                                                  | same as A4                                                                                                                     |
| F31 | Search `?q=` URL semantics                                                    | same as A5                                                                                                                     |
| F32 | Store `Must*` panic policy                                                    | same as D3                                                                                                                     |
| F33 | Helper micro-test bar                                                         | same as D2                                                                                                                     |
| F34 | Coverage teeth                                                                | same as C5                                                                                                                     |
| F35 | CI-bar subset policy                                                          | same as C4                                                                                                                     |
| F36 | Daemon-format coupling note home                                              | same as D13                                                                                                                    |
| F37 | Cross-session gate protocol                                                   | same as D18                                                                                                                    |
| F38 | erraudit `tree` dedupe upstream ask                                           | same as D17                                                                                                                    |

---

## The payoff map — which answers unlock which work

| Answer      | Unlocks (TODO home)                                    |
| ----------- | ------------------------------------------------------ |
| B1 GO       | M07–M09 (group rooms, 3 × 100 min), then M13           |
| B2 GO       | M10–M11 (E2EE decrypt-only) + the backup-set row       |
| B3 GO       | M05 (config strictness)                                |
| C1 GO       | the pma#348 PR (~100 min, cross-repo)                  |
| C2 option 1 | the `release.sh` freshness gate                        |
| C5 GO       | the coverage-floor policy + one upstream ask           |
| C7 GO       | the `backup.destDir` assertion row                     |
| D1 GO       | closes the FIFTH-surfacing dedup row                   |
| D3 GO       | the `Must*` row + a store hardening train              |
| D5/D6 GO    | two pma asks (guard hooks + daemon pre-sweep gate)     |
| D10 GO      | the 46-entry baseline becomes derivable                |
| D11 GO      | the archive-residue recipe splits live/archived        |
| D12 both    | the sweep becomes safe locally AND upstream fixes ship |
| D14 GO      | AGENTS compaction (~120 lines moved to lessons)        |
| D18 GO      | the gate-collision class closes fleet-wide             |
| E1 GO       | the carrier verdict rows close                         |
| F all yes   | 38 briefing rows retire into one recorded sitting      |

Everything else in the pack closes a ROADMAP bullet with a bounded answer and
unlocks no train — that is why they sit at the bottom of their groups.

---

## Resolved or obsolete since the ROADMAP bullets were written (no answer owed)

Recorded so nothing is silently lost, and so the ROADMAP section can be pruned
in the same sitting:

- **v2.9.0 fold timing** — obsolete: v2.9.0 shipped (the build renders
  `webphone-2.9.0.drv`); the fold question predates it.
- **Ring-silence autoplay** — RESOLVED 2026-10-01 (owner ruled KEEP; the
  gesture-scoped shared `AudioContext` shipped + pinned).
- **Contacts ownership** — CONFIRMED 2026-10-01 (Ledger's domain; recorded as
  the boundary, no longer a call).
- **Stack input policy** — DECIDED 2026-09-20 (ride `main`, per-train bump).
- **`nix flake check --all-systems`** — RESOLVED as NOT-DO (runbook §8).
- **Release.sh E2E load handling** — SHIPPED 2026-09-23 (`load_gate` +
  clean-tree assert); only the ratification remains (F25).
- **`errorfamily.Join`** — the 4 aggregate sites carry reasoned nolints; the
  upstream-constructor option stays declined unless F14 flips.
- **Dashboard fence** — shipped and pinned (E7); cutover + HSTS remain E8/C3
  as genuine calls.

---

## Source index (every link verified to exist this session)

| Home                                                                                                | What it owns                             |
| --------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| [2026-10-29 owner decision packet](2026-10-29_15-53_SUPERB-owner-decision-packet.md)                | the six cards' evidence one-pagers       |
| [2026-10-29 group-rooms model spike](2026-10-29_12-20_matrix-group-rooms-model-spike.md)            | B1's four sub-decisions                  |
| [2026-10-29 matrix E2EE go-story](../research/2026-10-29_12-35_matrix-e2ee-go-story.md)             | B2's feasibility + the F46 correction    |
| [pma push-leg durability draft](../drafts/pma-push-leg-durability-issue-2026-10-28.md)              | C1's root cause + the stall table        |
| [2026-09-22 owner-calls briefing](2026-09-22_13-50_owner-calls-briefing.md)                         | the §F rows' original wording            |
| [sessions credentials-at-rest review](../reviews/2026-09-30_sessions-credentials-at-rest-review.md) | B6's baseline                            |
| [modularization proposal](../modularization/2026-10-17_PROPOSAL.html)                               | the dormant carve triggers               |
| [native-app platform decision](2026-10-23_16-13_native-app-platform-decision.md)                    | the P0/P1/P2 split + the Flutter HARD NO |
| [fax thumbnail decline note](2026-10-02_09-05_fax-thumbnail-decline-note.md)                        | A1                                       |
| [fax capability decision](2026-10-30_04-15_fax-decision.md)                                         | the declared-capability doctrine         |
| [dashboard fence decision](2026-11-04_23-45_dashboard-fence-decision.md)                            | E7                                       |
| [post-v9 execution plan](2026-11-07_18-46_SUPERB-post-v9-execution-plan.md)                         | where B8, D10, D11 came from             |
| [release runbook](../release-runbook.md)                                                            | C9, E13, F25                             |
| [error contract](../error-contract.md)                                                              | A3's ruling home                         |
| [dedup registry](../dedup-registry.md)                                                              | D1, F1–F4                                |
| [UI/UX pareto plan](2026-10-01_03-53_SUPERB-ui-ux-pareto-plan.md)                                   | A7                                       |
