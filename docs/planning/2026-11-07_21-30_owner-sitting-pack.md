# Owner sitting pack — every open ruling in one dated agenda (2026-11-07 21:30)

**Status:** PACK (answers owed) · **Owner time to clear it:** ~45–60 min ·
**Evidence verified at** HEAD `605d2289` (2026-11-07 21:30): every `Evidence:`
path in §5 was `-e`-checked this session, and every "unlocks" pointer names a
section that exists in `TODO_LIST.md`.

Why this file exists, and what it replaces: open owner questions accumulated in
**three** homes that no session can join by hand — `ROADMAP.md` §Open questions
(47 bullets, several of which are already resolved in code and only need
ratification), the six cards of the
[2026-10-29 packet](2026-10-29_15-53_SUPERB-owner-decision-packet.md) (delivered
as a file, never answered), and the three posture calls the
[post-v9 execution plan](2026-11-07_18-46_SUPERB-post-v9-execution-plan.md)
raised mid-flight. This pack is the ONE sitting. The packet stays as the
evidence appendix it already is; **answers go here** and this file becomes the
record.

The three questions below are ordered by blast radius, not by age: A and B are
the ones where a wrong answer costs users, C and D cost maintenance, E is
outside this repository. Rows tagged **[ratify]** are already implemented — the
answer only blesses what ships today, and "yes" is the do-nothing option.

---

## How to answer

Answer inline in the block below, one line per ID. `GO` / `NO-GO` for binaries;
the option number for multi-way calls; `defer` is always legal and always means
"the status quo continues to ship".

```text
A1:   A2:   A3:   A4:   A5:   A6:
B1:   B2:   B3:   B4:   B5:   B6:   B7:   B8:
C1:   C2:   C3:   C4:   C5:   C6:   C7:   C8:   C9:   C10:
D1:   D2:   D3:   D4:   D5:   D6:   D7:   D8:   D9:   D10:  D11:  D12:  D13:  D14:  D15:
E1:   E2:   E3:   E4:   E5:   E6:   E7:   E8:   E9:   E10:  E11:
```

---

## A. Product & UX — what the user sees (6 rulings)

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
  transport cannot confirm, which is the exact class the island-honesty contract
  forbids.
- **Evidence:** AGENTS.md § Gateway seam; `internal/gateway/loopback.go`

### A3. Provider-failure retryability

Today only `transient` failures render a Retry affordance; a provider verdict
(bad number, blocked, carrier reject) is terminal and the user re-composes. Fax
`Resend` is the counter-example: it reuses the stored document.

- **Recommend:** keep provider verdicts terminal, and state the rule in
  `docs/error-contract.md` so the typed `FailureKind` machine is not read as an
  oversight.
- **Cost of the other branch:** a Retry button on a hard reject trains users to
  press it, and it will fail identically — worse UX than no button.
- **Unlocks:** closes §Open tasks "Provider-failure retryability".

### A4. Missed-call semantics + badge clearing

Two sub-questions bundled. (a) A deliberately REJECTED incoming call does not
count as missed today (`wp:call-missed` never fires). (b) The missed badge
clears on History tab click, not when the island's _Recent calls_ panel opens.

- **Recommend:** (a) keep REJECT out of "missed" — the user was present and
  chose; (b) also clear on the island panel opening, since the user has then
  seen the list. (b) is ~10 min.
- **Cost of the other branch:** (a) counting a reject as missed inflates the
  badge the user must then clear manually; (b) leaving it means a badge survives
  a panel the user already read.
- **Unlocks:** the two `wt`/badge rows in §Open tasks.

### A5. Search `?q=` URL semantics

A typed message search is ephemeral today. Deep-linking it would push `?q=`
into history so back-button unwinds filters.

- **Recommend:** adopt `?q=` (deep-linkable). It is the platform-native
  behaviour for a search box and costs one `pushState`.
- **Cost of the other branch:** a user cannot share or bookmark a search, and
  back-button leaves the tab entirely instead of unwinding the filter.
- **Unlocks:** the search row in §Open tasks.

### A6. Recordings in the webphone UI

The stack records calls server-side (stereo WAV, operator basic auth); this repo
has NO recording capability, only CDR rows. Putting playback in the UI is a
product decision with a consent/jurisdiction dimension.

- **Recommend:** NO-GO for now, recorded with its trigger ("the stack exposes a
  per-extension recording list with a consent posture"). The operator already
  has the recordings; the missing piece is a permission model, not a player.
- **Cost of the other branch:** shipping playback without per-extension consent
  scoping would let any session with a CDR row reach another party's audio.
- **Evidence:** AGENTS.md § Conventions (recording is two-level)

---

## B. Privacy, identity & data — what we store and send (8 rulings)

### B1. Group rooms — ruling A1 (packet card 1)

Matrix threads stay peer-keyed for DMs and become room-keyed for groups;
`domain.Address` accepts both `@…` and `!…`; the thread key is unchanged; groups
are born-inbound only. One schema change: nullable network-neutral
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
**decrypt-only** (inbound DMs readable, no encrypted send). The `NewOlmMachine`
signature claim was re-verified at source (F46): P2's minimal-client shim is
viable, P1 (adopt `mautrix.Client` wholesale) stays rejected. Bundled
sub-decisions: crypto store in its own sqlite db (`e2ee.db`, own DDL ladder,
mirroring the usermgmt isolation precedent) vs mautrix's sqlstore module; TOFU
trust with an "unverified device" marker and no cross-signing UI in v1; the
crypto store joins the backup set and the restore drill.

- **Recommend:** GO on P2, crypto store in its own db, TOFU, backups included.
- **Cost of the other branch:** encrypted rooms stay unreadable, which is most
  of the rooms a real account is in.
- **Unlocks:** M10–M11.
- **Evidence:** `docs/research/2026-10-29_12-35_matrix-e2ee-go-story.md`

### B3. Config typo strictness (L7) — packet card 5

`config.Load`'s koanf `Unmarshal` silently drops unknown keys from both the JSON
file and `WEBPHONE_*` env, so a typo turns a seam off instead of failing the
boot. Design already constrained: strict (boot-failing, named-key) for the FILE,
warn-only for ENV.

- **Recommend:** option 1 — GO file-strict + env-warn. The genuine risk it
  carries: the NixOS module writes a freeform `settings` map, so a deployment
  carrying a passthrough key would start failing its boot. Mitigation already in
  the codebase: unknown keys WARN at boot today (`internal/config/unknown.go`),
  so the strictness flip has a measured blast radius to check first.
- **Cost of the other branch:** the outage class stays — a misspelled
  `gateway.webhook_secre` silently disables webhook auth.
- **Unlocks:** M05.

### B4. Token-hygiene parity for the newer secrets

`gateway.webhook_secret` has a `*_file` variant; `paperless.token` and
`crm.token` do by `crm.token_file` but paperless has none explicitly.

- **Recommend:** the standing answer (`environmentFile` for the rest) plus one
  explicit line in AGENTS saying so — the asymmetry is a gap in the DOC, not in
  the code.
- **Cost of the other branch:** adding four more `*_file` knobs multiplies the
  both-or-neither validation surface for zero new capability.

### B5. Session creation is a cross-owner IDOR if a handler forgets. Should it be structurally impossible? **[ratify]**

Every store query is extension-scoped; the ONLY cross-owner reads are the
deliberate `*Any`/`ListAll*` seams, admin-gated and pinned. Those pinned pins now
include the store's own regression suite (P1-1 of the plan, this session).

- **Recommend:** ratify the posture (owner-scoping is the rule, `*Any` is the
  audited exception) so future reviews cite a decision instead of re-deriving
  it from three test files.
- **Cost of the other branch:** none today; the cost is the recurring review
  time spent re-proving the same invariant each audit.

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
- **Unlocks:** the UMask row in §Open questions.

### B8. `/healthz` CRM criticality — the plan's own posture call (decided, needs blessing)

Measured this session: an unreachable CRM yields `503` +
`{"status":"degraded",…,"crm":{"status":"fail"}}`. Kept CRITICAL deliberately,
because it is the only surface reporting a revoked CRM token when the dashboard
is off, nothing routes on readiness (unit `Type=simple`, Caddy `remote_ip`
fences), and the go-health probe still grades the leg non-critical (WARN) — two
audiences, both true.

- **Recommend:** ratify "criticality is per surface, and this is the strict one".
- **Cost of the other branch:** making it non-critical removes the only alarm
  that fires when the token dies.
- **Evidence:** `internal/app/health_crm_test.go`;
  AGENTS.md § Architecture invariants (`/healthz` bullet)

---

## C. Infrastructure & deployment — what runs where (10 rulings)

### C1. Continue pushing the pma push-leg fix, or live with the mitigation? (packet card 3)

Root cause verified and filed upstream: `Stop()` drops pending batches with no
drain, `Start()` has no unpushed sweep. Measured tax: 30m37s + 54m + 70m +
62 min stalls across four repos in three days.

- **Recommend:** authorize the fix + PR in pma (~100 min). The watch script and
  the 60-min bar are a mitigation that costs a session minute each time it fires.
- **Cost of the other branch:** the stalls continue; every one of them costs
  more than the fix.
- **Evidence:** `docs/drafts/pma-push-leg-durability-issue-2026-10-28.md`

### C2. `tw-freshness`: advisory forever, or gate the release? (packet card 4)

The advisory fired on every CI run for ~24 h while templates were edited. It is
green at HEAD right now.

- **Recommend:** option 1 — keep the CI job advisory, and gate `release.sh` on
  freshness. A release can then never ship stale Tailwind bytes, while a
  concurrent template edit does not red main.
- **Cost of the other branch:** option 2 (blocking) turns every concurrent
  session into a main-red incident; option 3 accepts shipping stale CSS.
- **Unlocks:** the `release.sh` change and the tw row.

### C3. HSTS on `pbx.artmann.tech`

The `caddy.hsts` option ships opt-in.

- **Recommend:** GO once https-only is proven on that host; enable with
  `maxAge` starting at 1 day and climbing. HSTS is hard to walk back, so the
  ramp is the point.
- **Cost of the other branch:** staying off leaves a downgrade window on every
  first visit.

### C4. Where may CI skip? The docs/test-only subset policy (routed from a

2026-10-06 report; it lived only in a timestamped file)

Which CI legs may be skipped for a docs-only or test-only commit, and which must
never be.

- **Recommend:** the formatter gate is NEVER skippable (tonight's reds were
  formatter reds); `test-race` and the browser E2E may be skipped on a docs-only
  diff; everything else runs.
- **Cost of the other branch:** an unprincipled skip list reintroduces precisely
  the red-main class the 10-06 incident produced.

### C5. Should coverage have teeth?

The BuildFlow gate passes at ANY coverage level.

- **Recommend:** a floor on the union of hand-written packages only (not
  `-coverpkg` unions, which count generated `_templ.go`), set at today's
  measured value minus a small margin, so it ratchets down but never
  silently. This is fleet-wide (the step is upstream), so the answer also
  decides whether webphone files an upstream ask.
- **Cost of the other branch:** coverage stays informational and erodes.
- **Unlocks:** the coverage row in §Open questions.

### C6. `--border-strong` vs WCAG 1.4.11

The token measures 1.86:1 (light) / 1.53:1 (dark) against its surface; 1.4.11
wants 3:1 for non-text UI.

- **Recommend:** ACCEPT AND RECORD — the border is decorative separation and the
  surrounding affordances carry the contrast. A retune touches every input and
  panel edge, which is a visual-identity call, not a colour tweak.
- **Cost of the other branch:** a full visual-identity pass on every edge in the
  product, for a criterion whose intent (distinguishability) the form already
  meets.

### C7. `backup.destDir` / `dataDir` nesting legality

systemd `StateDirectory` splits on spaces; nesting is currently neither
forbidden nor documented.

- **Recommend:** forbid by assertion (the module already asserts `dataDir` lives
  under `/var/lib/`), because a silent nesting produces a backup that excludes
  its own source or recurses into it.
- **Cost of the other branch:** an operator configures it, and the failure
  appears at restore time — the worst possible moment.

### C8. Stack verification timing for the hardened module

Ride the per-train relock, or relock + browser E2E now?

- **Recommend:** ride the per-train relock (the default). The unit environment
  changed but the packaging did not, and a standalone E2E run costs a full
  stack-VM cycle for one boolean.
- **Cost of the other branch:** ~one session of stack-lab time per module change.

### C9. One-time force-push ratification

During the v2.6.0 fold a `--force-with-lease` was used once on a self-authored
commit ~1 minute after pushing it (the daemon raced `git add`).

- **Recommend:** ratify the narrow exception (own commit, seconds old,
  lease-protected) and write it into the runbook as a shape, not a permission.
- **Cost of the other branch:** the runbook says "never", an exception happened,
  and the record is ambiguous — which is how the next exception gets invented.

### C10. Two sessions on one mandate — who stands down? (the plan's third posture call)

Two Crush sessions worked this repository concurrently this session. Measured
cost: three narrative commit messages eaten by `chore: auto-commit`, one
duplicated status report, and repeated re-verification of shared files.

- **Recommend:** one session owns the mandate; the second takes a disjoint scope
  (a different repo, or a non-overlapping read-only audit). Zero-cost rule if a
  second session is unavoidable: it never edits `TODO_LIST.md`, `CHANGELOG.md`,
  `AGENTS.md` or `docs/status/` while the owner session is live.
- **Cost of the other branch:** the tax above repeats on every train.

---

## D. Tooling & process — how we work (15 rulings)

### D1. Art-dupl ritual baseline: `-t 1` or `-t 3`? (FIFTH surfacing)

`-t 1` is the ratified CI baseline (deepest sweep, zero harmful clones);
`-t 3 --type-aware` is the ritual baseline (43–45 shown groups, mostly idiom
noise, triage 1 extract / 3 accept).

- **Recommend:** keep BOTH and stop asking: `-t 1` is the gate (CI), `-t 3` is
  the ritual reading. They answer different questions, and the confusion is that
  they were ever described as alternatives.
- **Cost of the other branch:** picking one loses either the deepest gate or the
  readable ritual.
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

- **Recommend:** degrade to an error for the ROW (log + skip), keep `Must*` for
  write-path invariants. A corrupt row should cost one row, not the process —
  and the readmodels legacy import already proves the skip-and-demote pattern
  works (`legacy_import_test.go`, this session).
- **Cost of the other branch:** one bad row is a denial of service on the whole
  extension forever.
- **Unlocks:** the `Must*` row in §Open questions.

### D4. Concurrent-session breakage policy

When another session's committed-but-broken code blocks the shared gate: fix and
commit immediately, or keep hands off and report?

- **Recommend:** unblock the gate immediately IF the fix is mechanical and
  obviously local (format, generated artifact, one-line import), otherwise
  report and leave. "Hands off" leaves main red for hours; "always fix" collides
  with in-flight edits.
- **Cost of the other branch:** uncodified, each session improvises and each
  improvisation is reviewed as if it were policy.

### D5. Daemon hardening: exclude `docs/status/` and `docs/planning/`, and gate

before sweeping

The heuristic daemon has twice swept half-written reports and has three times
committed a non-compiling tree.

- **Recommend:** both asks, upstream in pma. This session added a local
  mitigation already (`whitespace-drift.sh` as GATE 1, `node --check` as GATE 2
  in `.githooks/pre-commit`) — but that hook is bypassed by the daemon's own
  `core.hooksPath` override, which is the fleet-wide gap below.
- **Cost of the other branch:** the class repeats, and each instance costs a
  red-main attribution.

### D6. `core.hooksPath` override in the pma unit

The daemon unit exports `GIT_CONFIG_{COUNT,KEY_0,VALUE_0}=core.hooksPath=
/var/empty`, so every daemon commit runs NO hook — enabling hooks re-binds the
daemon fleet-wide.

- **Recommend:** YES, enable hooks in the daemon, scoped to the two cheap gates
  (fmt-drift, syntax). The alternative is that every guard this repo builds is
  advisory against its most frequent committer.
- **Cost of the other branch:** guards that do not guard.
- **Evidence:** AGENTS.md § Conventions (the 2026-10-15 guard-hole note)

### D7. Push-lag policy: at what lag is the pusher BROKEN, not slow? **[ratify]**

The ratified bar is 60 minutes from the oldest unpushed commit
(`scripts/push-stall-watch.sh`), with a warn-only 30-min line available.

- **Recommend:** ratify 60 min as BROKEN and add the 30-min WARN line — the
  2026-10-14 stall recovered at ~40 min and was invisible to the bar.
- **Cost of the other branch:** a 40-minute stall stays silent, as it did.

### D8. Session phase-boundary hand-pushes **[ratify]**

Three hand-pushes of daemon-swept content followed runbook precedent but
overrode a hard never-push rule.

- **Recommend:** ratify as standing session duty, bounded by: main must be red
  with a verified fix in hand, or the 60-min bar met. That is already how it is
  practised and already written in AGENTS.
- **Cost of the other branch:** an undocumented practice that the runbook
  contradicts.

### D9. `.buildflow-baseline.json` growth — audit, trim, or re-save? **[resolved this session, ratify the shape]**

The file went 78 → 124 findings with 46 `file-size-check` accepts, chosen
per-finding (so a NEW oversize file still gates) rather than per-rule (a silent
suppression). `buildflow baseline add` is the only writer.

- **Recommend:** ratify per-finding accepts as the shape, and the file-size cap
  question below as the separate real question.

### D10. `file-size-check`'s hardcoded 350-line cap — keep, skip, or push upstream? (the plan's own second posture call)

46 files exceed it. Accepting them per-finding keeps the gate meaningful but
freezes a number nobody chose.

- **Recommend:** push upstream for a configurable cap, and keep the 46 accepts
  meanwhile. A repo-wide cap on Go files is a reasonable default but an
  unreasonable constant; webphone already carries a legitimate exception class
  (`views/_templ.go` neighbours, the store ladder).
- **Cost of the other branch:** skipping the step (even visibly) removes a gate
  that caught nothing today but costs nothing either.
- **Unlocks:** the 46-entry baseline becomes derivable instead of historical.

### D11. Archive-residue policy — per-row evidence or transcribed verdict? (the plan's own first posture call)

~71 ratcheted findings remain in the pre-canon ARCHIVED corpus, plus ~1600
un-inventoried §a/§b rows in the live reports. The v9 recipe (per-row evidence,
~15 min/file) is expensive and was proven.

- **Recommend:** per-row evidence for the LIVE reports (the 1600), and
  transcribed-verdict-per-cell for the ARCHIVED corpus, which is CLOSED HISTORY
  (its 71 findings are ratcheted, never re-litigated). This splits the cost
  where it buys something.
- **Cost of the other branch:** a blanket "transcribe a verdict" for live
  reports would mark evidence nobody produced — the exact vacuous-PASS class
  this train closed.
- **Unlocks:** the cadence ratchet and the archive rows.

### D12. The three docs-gates defects — file upstream, or vendor a local shim?

Found this session: DF brittleness (a token flips "generic" at a living-doc line
count, so the same row flips verdict as docs grow), standalone verbs ignore
`.docs-gates.json`, an escaped pipe inside a table cell mangles a strike, and an
unvalidated verdict cell accepts garbage.

- **Recommend:** file upstream (drafts prepared, P3-2) AND keep the local guards
  this session shipped (`strike-safety.py`, the `generic_df` pin, the wrapper
  passing `--audit-max` explicitly in `--check`). Both, because the local guards
  are cheap and the upstream fix has no date.
- **Cost of the other branch:** vendoring alone forks the tool; filing alone
  leaves the sweep unsafe until upstream ships.

### D13. Daemon-format coupling note: which home?

The auto-commit daemon's reflow and the gates' formatter are coupled; the note
has three candidate homes (a test comment, the error-contract preamble, AGENTS).

- **Recommend:** AGENTS — it is the only home every session reads before
  touching a formatter, and the coupling has bitten three times.
- **Cost of the other branch:** a test comment is read by whoever runs that
  test; the error contract is read by whoever debugs an error.

### D14. AGENTS.md restructure permission (498 lines vs the 377 BuildFlow cap)

Compaction moves war stories to `docs/lessons.md` and rewrites sections every
concurrent session depends on.

- **Recommend:** grant it, and do it in a session that owns the file (see C10).
  The cap is already enforced, so today the file either shrinks or the gate
  drifts.
- **Cost of the other branch:** the cap is a preflight that will eventually
  block a train at an inconvenient moment.

### D15. Domain-language glossary: ratify or delete

`docs/DOMAIN_LANGUAGE.md` exists with a memo appended in 2026-10-14; the
declarative trains made its vocabulary load-bearing in code.

- **Recommend:** ratify. Deleting a glossary whose terms are now type names
  would make the code harder to read, not the docs shorter.
- **Cost of the other branch:** the vocabulary is re-derived from code by every
  new session, at ~30 min each.

---

## E. Cross-repo & carrier — outside this repository (11 rulings)

### E1. Carrier-leg verdicts: inbound-MMS isolation test, Google Voice, and

`+14084757593` ownership

The test was never recorded anywhere. MMS ships and is webhook-tested today, so
this is a verdict-recording call, not a build gap.

- **Recommend:** run the isolation test once on a real carrier line, then record
  the GV verdict and number ownership in FEATURES + AGENTS. Needs a carrier
  phone.
- **Cost of the other branch:** an unanswerable question in a doc that a future
  reader will trust.

### E2. Release numbering + the bridge's "webphone >= 2.8" claims

- **Recommend:** reword the bridge claims to commit/date form. Webphone's own
  AGENTS line is already commit-pinned; making both forms match removes a
  version claim nobody remembers to bump.
- **Cost of the other branch:** cutting 2.8.0 to satisfy a doc string.

### E3. GitHub Release objects per tag

v2.0.0 has a tag, no Release object; later tags do.

- **Recommend:** Release objects for user-visible trains only, generated from
  the CHANGELOG's `[Unreleased]` fold. Tags remain the machine contract.
- **Cost of the other branch:** either every tag gets an empty Release (noise on
  the repo page) or the changelog has no public face.

### E4. Go module-path policy for v2+ tags

`/v2` suffix vs a NOT-DO record.

- **Recommend:** NOT-DO record — a flake-consumed application never runs
  `go get` on itself, so a `/v2` path buys nothing and breaks every flake input
  URL that names the module.
- **Cost of the other branch:** a module-path migration with zero consumers.

### E5. Handler gating: keep the deliberate dual layer? **[ratify]**

`requireSession` helpers plus `Sessions.Require` on `/events` and `/phone-api/`.

- **Recommend:** keep. The middleware wiring covers the routes whose handlers
  cannot call a helper (SSE, the byte proxy), so the redundancy is by shape, not
  by accident. The contract test that allowlists exactly four 401 writers is the
  proof it stays honest.

### E6. Rate-limit keys: does the stack sanitize `X-Forwarded-For`?

Gates the flip from port-stripped peer-host keys to `KeyExtractorFromClientIP`.

- **Recommend:** measure on the stack before deciding; until then keep the safe
  default. The answer is one `curl` with a spoofed header against the vhost.
- **Cost of the other branch:** flipping blind means one client can poison a
  rate-limit bucket by spoofing.

### E7. Stack `/health` dashboard exposure

- **RESOLVED 2026-11-04** — the fence is the proxy contract (`forwardAuth` under
  operator SSO, loopback-only otherwise) plus a boot WARN naming it;
  `WithMiddleware` was declined because the app is proxy-blind.
- **Recommend:** ratify the recorded decision so the row can close.
- **Evidence:** `docs/planning/2026-11-04_23-45_dashboard-fence-decision.md`

### E8. HSTS posture remains in C3 — no separate ruling; listed so the count in

the ROADMAP closes cleanly.

### E9. SUPERB prod cutover timing + `staticFallbackExtensions`

Should prod's first static list name operator/ring-group extensions explicitly?

- **Recommend:** yes, name them. A fallback list that omits the operator's own
  extension is a fallback that cannot route the one call that matters.
- **Cost of the other branch:** a degraded-mode call to the operator fails.

### E10. F49 scope: eval-level rollback proof or a full in-VM flip drill?

- **Recommend:** eval-level (byte-identical static golden) as the gate, with the
  in-VM drill as a one-time manual ritual recorded in the runbook. A ~10-minute
  CI tax on every run for a static/xml_curl flip that changes monthly is not
  priced right.
- **Cost of the other branch:** ~10 min CI + a second toplevel, forever.

### E11. Stack-side TEMP-DIAG answer-phase dump in the browser E2E (commit

`b96d4c2` there) — keep or revert?

- **Recommend:** revert. It was diagnostic scaffolding, and a dump that stays
  becomes an unowned output nobody reads.
- **Cost of the other branch:** E2E logs grow with a payload no assertion
  consumes.

### E12. `gateway.attachment_limit` and the sniff-fallback lifespan

Two small ones bundled: implement a local attachment cap, or ratify the bridge's
honest-422? And does the bridge's octet-stream sniff fallback live forever?

- **Recommend:** ratify the honest 422 (the bridge knows the real limit and
  names it); keep the sniff fallback forever with a comment naming it as
  permanent seam robustness rather than a deploy shim — a fallback with a
  deadline gets deleted by someone who does not know why it exists.
- **Cost of the other branch:** a local cap duplicates a limit the bridge owns
  and the two will drift.

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
| E1 GO       | the carrier verdict rows close                         |

Everything else in the pack closes a ROADMAP bullet with a bounded answer and
unlocks no train — that is why they are at the bottom of their groups.

---

## What this pack deliberately does NOT contain

- **Anything already ratified in AGENTS.** Fleet-libs reopen triggers, the
  DECLINED/NO-GO lists, the island-type-gate verdicts, the language stance, the
  admin read-seam posture: all final, all with written triggers. Re-litigating
  them is out of scope by the plan's guardrail 4.
- **Owner-gated production actions.** The passkey deploy proof, the mic ritual,
  the `/health` deploy, the ident-manager rebuild: each is a terminal ritual,
  not a ruling.
- **Anything a measurement can answer.** E6 (`XFF` sanitization) is written as
  "measure", not "decide", because the answer is one curl.

---

## Source index (every link verified to exist at HEAD `605d2289`)

| Home                                                                                                | What it owns                             |
| --------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| [2026-10-29 owner decision packet](2026-10-29_15-53_SUPERB-owner-decision-packet.md)                | the six cards' evidence one-pagers       |
| [2026-10-29 group-rooms model spike](2026-10-29_12-20_matrix-group-rooms-model-spike.md)            | A1's four sub-decisions                  |
| [2026-10-29 matrix E2EE go-story](../research/2026-10-29_12-35_matrix-e2ee-go-story.md)             | P2's feasibility + the F46 correction    |
| [pma push-leg durability draft](../drafts/pma-push-leg-durability-issue-2026-10-28.md)              | C1's root cause + the stall table        |
| [2026-09-22 owner-calls briefing](2026-09-22_13-50_owner-calls-briefing.md)                         | the 38 rows this pack folds in           |
| [sessions credentials-at-rest review](../reviews/2026-09-30_sessions-credentials-at-rest-review.md) | B6's baseline                            |
| [modularization proposal](../modularization/2026-10-17_PROPOSAL.html)                               | the dormant carve triggers               |
| [native-app platform decision](2026-10-23_16-13_native-app-platform-decision.md)                    | the P0/P1/P2 split + the Flutter HARD NO |
| [fax thumbnail decline note](2026-10-02_09-05_fax-thumbnail-decline-note.md)                        | A1                                       |
| [fax capability decision](2026-10-30_04-15_fax-decision.md)                                         | the declared-capability doctrine         |
| [dashboard fence decision](2026-11-04_23-45_dashboard-fence-decision.md)                            | E7                                       |
| [post-v9 execution plan](2026-11-07_18-46_SUPERB-post-v9-execution-plan.md)                         | where B8, D10, D11 came from             |
| [release runbook](../release-runbook.md)                                                            | C9, E2                                   |
| [error contract](../error-contract.md)                                                              | A3's ruling home                         |
| [dedup registry](../dedup-registry.md)                                                              | D1                                       |
