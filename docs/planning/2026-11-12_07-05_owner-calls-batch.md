# Owner calls — the open batch (2026-11-12)

> ONE home for the questions the 2026-11-12 06-17 SUPERB plan batches for the
> owner. The sibling plan (05-25, task T21) appends ITS pending questions here
> instead of forking a second doc — the stub section at the bottom is theirs.
> Default answers already executed under the owner's blanket GO are marked
> EXECUTED-BY-DEFAULT so the sitting can ratify or reverse them in one pass.

**Status:** OPEN (answers owed) · **Clearing time:** ~10 min for Q1–Q5 ·
**Created by:** the 06-17 plan's M04 (F009) · Sources cited per question.

## Questions

**Q1 — Phone-tab label i18n** (§f-15 of the 04-09 report; TODO_LIST row):
the mobile bottom-bar "Phone" label is English-only by the D3 shell-copy rule
while the six tab labels beside it localize. Keep English (D3 stands — island
copy is the localized surface, shell copy is English) or carve a one-word
exception for the nav label?
_Default if unanswered:_ D3 stands, label stays English; the row records the
ruling instead of a change.

**Q2 — Empty-state action buttons** (§f-23, owner eyes): empty-state rows say
"add one above" but carry no button; the templ-components EmptyState library
supports ActionText. The 04-09 session ruled it a duplicate affordance (the
composer is directly above every empty list). Owner may disagree — want real
buttons that scroll/focus the composer?
_Default if unanswered:_ the DECLINED ruling stands (no duplicate affordance).

**Q3 — index-render.jpg refresh** (04-10 report, question 1): the audit
index's rendered screenshot is the 2026-11-04 edition; the footer now captions
it honestly. Refresh the render (a capture run + crop) or keep the caption?
_Default if unanswered:_ caption stays; refresh happens at the next corpus
sweep that changes the headline numbers.

**Q4 — templ-components taste re-litigation** (04-10 report, question 3):
does the 2026-11-11 owner-taste climate (GREEN is the brand; the bar is the
owner's eyes) re-open the DECLINED templ-components component rulings
(OutboundStatusBadge, feedback.Alert, panelHead, the 9 ruled-out wp-empty
sites)? Each decline carries byte-pin or wire-contract evidence; the standing
rule is refresh-trigger-only re-litigation.
_Default if unanswered:_ declines stand; the climate alone is not a named
trigger — say "re-open <component>" and the train gets dispatched.

**Q5 — BuildFlow bandit draft banner** (the 06-17 plan M25/F065): an upstream
issue draft to BuildFlow (bandit findings' owner banner) sits with its owner
banner unchecked. Provide the banner text (one line) or say "archive" and the
draft dies.
_Default if unanswered:_ archived at the next sweep as a zombie draft.

## Executed-by-default (ratify or reverse)

**E1 — Bump timing** (04-10 report, question 2): go-paperless v0.5.2 was
EXECUTED 2026-11-12 (the 06-17 plan's M14, under the blanket GO; landed
585f630a, vendorHash re-pinned, suites green) — ratify the precedent or
revert to bump-only-on-trains. The catalog half of E1 DISSOLVED: F045's
resolvability check found NO v4.15.0 published (latest = v4.14.0 = our
pin); there was nothing to bump and the stale v4.9.0→v4.10.0 owner-gate
row is struck as superseded.

## Gated rows — staged recipes (run when un-gated)

**M18 — stack redeploy + voicemail verdict** (owner-gated; ~60m):

1. In the stack repo: bump the webphone pin to current main, `nix flake check`, deploy (`nixos-rebuild switch`).
2. On prod, click play on a voicemail with a DEAD recording (the standing live defect).
3. RECORD: the toast text shown (the live trigger's name), whether it repeats per attempt, and — if accessible — `journalctl -u webphone | grep -i voicemail` around the click.
4. Verdict line lands in the TODO_LIST voicemail row; the toast contract was fixed 2026-11-10 (one honest toast per episode) — prod confirms the fix rode the deploy.

**M19 — authenticated prod capture** (owner-gated on CREDS; ~60m):

- Needs `WEBPHONE_CAPTURE_EXTENSION` + `WEBPHONE_CAPTURE_PASSWORD` for pbx.artmann.tech (third ask — see Q6 below), OR the stack-redeploy-first order (M18 before M19).
- Command: `WEBPHONE_CAPTURE_EXTENSION=1001 WEBPHONE_CAPTURE_PASSWORD=<sip-dir-password> nix shell nixpkgs#chromium nixpkgs#chromedriver nixpkgs#python312.withPackages(ps: [ ps.selenium ps.pillow ]) --command python3 scripts/ui-capture.py --base-url https://pbx.artmann.tech --track`
- Then `--open` the gallery for the owner walkthrough. Run dir: `ui-shots/pbx.artmann.tech/<UTC-stamp>/`.

**M20 — night-run stability pair** (owner-gated on WINDOW; calendar note):

- Two tracked runs ≥8h apart in the quiet window (e.g. 02:00 + 11:00 CET) against the SAME deployment; diff the noise floor, calibrate the badge threshold (currently `--badge-bar` default), record values in the tracking plan. First leg can piggyback on M19's first tracked run.

**M21 — ring-guard positive path** (owner-gated on a REAL inbound call):

- During a tracked run, place a real inbound call to the captured extension; the harness re-shoots `#incoming-call:not([hidden])`; the re-shot frame lands in the run dir. Pair with M19/M20 runs to avoid a dedicated session.

**Q6 — M19 credentials (third ask)**: provide the prod capture creds (or order M18 first so the session mints its own).

## Stub: the sibling plan's batch (05-25 T21 appends here)

The 05-25 deepening plan's T21 landed 2026-11-12 (the deepening trains
session): every sitting question it owed this batch, numbered continuing
from Q6, each with its evidence link and a default if unanswered.

**Q7 — CRM timeline facts (the CRM trio, 1/3)** (TODO_LIST "issue #9 +
SUPERB execution tails" § Owner questions): email/SMS/fax message
BODIES into the CRM timeline? Today only call metadata crosses; sending
bodies makes the phone a producer of the CRM's relationship history.
_Default:_ bodies stay local; only call metadata crosses (the data-mesh
ruling 2026-10-29 — AGENTS § CRM seam).

**Q8 — Actionable CRM strip (2/3)** (same TODO row): completing a task
or moving a deal FROM the phone flips the read-only ownership rule the
strip is built on. Want write-back, or does read-only stand?
_Default:_ read-only stands (acting happens in the CRM; issue #10's
product keeps the phone a reader).

**Q9 — Kith push authority (3/3)** (same TODO row): whole branch as-is /
cherry-pick the CRM-context commits onto a clean branch / leave it to
that session?
_Default:_ leave it to the owning session (the Kith repo's train owns
its landing order).

**Q10 — cqrs-lint gating posture** (TODO_LIST "Owner decisions pending";
AGENTS § erraudit): the 186→12 triage left 12 deliberate findings
keep-visible and the tool NON-GATING until the error_rules-bearing
BuildFlow release. Promote cqrs-lint (and the other tier-2 tools) to
GATING at that release, or keep advisory?
_Default:_ promote at the release (the elevation key is documented in
`.buildflow.yml`); the 12 keeps stay visible either way.

**Q11 — C044 prefixed-key permanence** (`.cqrs-lint.json` rules.disable
carries the machine reason; AGENTS § erraudit): `id.ID.String()`'s
`Brand:value` prefixed map keys are DECLINED as an architecture ruling —
confirm permanent, or price a TODO row for a wire-shape migration?
_Default:_ permanent (the registry + consumers read the prefixed form;
no migration path exists).

**Q12 — B5 prod-switch authority + timing** (TODO_LIST "Release-green +
B5-staged" open rows, from the 00-45 report §g Q1): who executes the
pbx-artmann relock + prod cutover (B5) and when — the staged recipe
sits ready.
_Default:_ the next release train's tail (release-runbook dance), one
session owning the whole sequence.

**Q13 — traces-in-backup** (TODO_LIST "Traces-in-backup ruling (owner)",
routed from f22/f23/d4/g1): should `<dataDir>/traces` ride the backup
set, or be excluded from the module's backup unit + documented? The
2026-11-09 journal swap moved the db set to system.db and left this
open. Evidence: `internal/app` traceRecorder writes `<dataDir>/traces`
(flight-recorder seam, default OFF).
_Default:_ excluded + documented (traces are diagnostic exhaust, not
state; 20 snapshots gzip best-speed).

**Q14 — ASR M1 first-live + retention/autoplay** (TODO_LIST ASR row
"OWNER-gated (14-08 report §g)"): the live provider for the first real
transcription (Speaches deployment vs throwaway key), retention and
autoplay ratification. Still owner-gated as of 2026-11-12 (verified —
nothing self-answered).
_Default:_ unchanged gates; the row waits for the answer.

**Q15 — crash-window blob orphan sweep** (the deepening plan T03's
standing owner question; TODO_LIST full-code-review row): the
compensating deletes cover the persist-failure path; the CRASH window
(file durable, row lost, process dies before compensation) still
orphans blobs. Add a boot-time orphan sweep, or accept the residue?
_Default:_ accept (documented residue; a sweep risks deleting
not-yet-referenced files written in the same second).

**Q16 — storage-ladder §5 DROP date** (TODO_LIST §5 owner row, the
one-query emptiness proof): the seven frozen hand-ladder families keep
their tables until a dated DROP. Proposal: the first 2.15.x release
window after 2027-01-01. Ratify the date or name another.
_Default:_ ratify (16 months of backup-window evidence is generous).

**Q17 — race-probe transient** (the deepening plan T02's follow-up):
the concurrent-first-sight race probe has failed twice in FULL-suite
runs and passes alone every time (assertions are deterministic — a
scheduling artifact of -race under load). Policy: keep it in the suite
with a documented re-run protocol (capture full `-v` output before any
re-run), or move it behind `testing.Short()`?
_Default:_ keep + re-run protocol (two flakes, both never reproduced in
isolation; the probe guards a real data-correctness question).

## F62 posture verification (2026-11-12, no owner action needed)

- **brandid-lint BD003**: 75 directives swept for the placement contract
  (single physical line, last line of its comment group): 1 real
  displacement found and FIXED (matrix ClientConfig — the T14
  wiring-struct comment insertion had pushed the directive out of the
  target's group, covering nothing); the 2 `settings_templ.go` flags
  are GENERATOR REFLOW (the authored `settings.templ` groups terminate
  correctly — templ's emission reflows the comment). The full tool scan
  needs the v0.5.1 binary (not in go.mod — opt-in); posture otherwise
  unchanged since the 2026-10-25 clean verdict.
- **#48 (remote-leg-only transcription default)**: still owner-gated
  (TODO_LIST ASR polish row) — privacy call, unchanged.
- **ASR deploy-proof rows**: still owner-gated (M1/M2/M4/M6 block, Q14
  above) — unchanged.
