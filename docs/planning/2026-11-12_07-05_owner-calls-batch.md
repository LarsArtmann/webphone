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

The 05-25 plan's T21 questions (Q1–Q3 CRM, cqrs-lint gating, C044, B5
authority, traces-in-backup, M1 first-live) land in this section when that
session runs — numbered continuing from Q6. Do NOT open a second owner doc.
