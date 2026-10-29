# Issue-review runbook

How an issue moves from "observed" to "adopted" in this repo, and what
review owes before work starts. Born from the 2026-10-27 triage: the
dialer adoptions (#6/#7/#8) cited external PRs as facts, and T19 had to
re-verify every claim at source before more weight rode on them.

## 1. Verify external claims AT SOURCE (before anything else)

Every claim about another project (numbers, behaviors, "they fixed it
this way") is read in the linked PR/commit/README — not taken from the
issue body citing it, however authoritative it reads.

- Fetch the PR (`gh pr view N --repo owner/repo`): does the body state
  the cited numbers/behavior? Is it merged?
- Fetch the README/docs section when the claim is product behavior.
- A claim that diverges gets CORRECTED in the issue body with a dated
  note — never silently inherited.

## 2. Locate the cross-repo counterpart

If the proposal touches a seam another repo owns (Kith/crm contract,
stack module option, pbx-artmann host wiring), find and link the
matching issue/option THERE (the "matching Kith issue" step of T19).
An adoption with a cross-repo leg gets the `cross-repo` label and its
non-webphone work is spelled out per repo.

## 3. Label

- `verified` — the external claims have been checked at source (step 1
  done; cite the verification comment).
- `cross-repo` — work spans another repo (tri-repo ritual applies:
  webphone pushes before the stack re-pins).
- `blocked` — an external actor gates progress (owner action such as
  minting a token, a live-host deploy, an upstream fix). Say WHAT and
  WHO in a comment so the unblock is greppable.

## 4. Shape the acceptance bar

Acceptance bullets are observable and testable BEFORE the first
commit: a bullet that no test, smoke scenario, or live probe could
ever verify is rewritten until one could.

## 5. Re-verify at close-out

The closing comment re-states the original claims and their final
state (shipped as proposed / adapted because X / dropped because Y).
The 2026-10-29 #6 verification comment is the worked example.
