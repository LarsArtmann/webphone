# Parked: stack-runbook boot-contract patch (NOT applied here)

**Date:** 2026-10-02
**Applies to:** `nix-international-telephony/docs/ops-runbook.md` § "Webphone error contract"
**Why parked:** the stack repo is a separate dispatch (tri-repo rules);
apply under the ritual, webphone first, CLEAN stack tree, then relock
(`docs/release-runbook.md`). No stack edits from this train.

## Patch text (append to § "Webphone error contract")

```markdown
### Boot surface (webphone 2026-10-02)

The operator is a user: every webphone boot failure renders the
five-part contract (WHAT / REASSURE / WHY / FIX / ESCAPE) plus the
underlying error to the journal, version-stamped and class-tagged,
English-only. Exit codes: 1 = designed boot failure (config,
data-dir, timezone, paperless, listen, generic), 2 = panic (Go's
default, deliberate). With `Restart=on-failure` + `RestartSec=5`
(the module default) each attempt renders once per 5s until fixed;
`systemctl stop webphone` ends the loop. Classes and their fix hints
live one-home in webphone's `docs/error-contract.md` § "Boot
surface" — read that table before journal triage; the journal lines
to grep are `webphone boot failed (class=…` (or `webphone boot
panicked`) and the five `WHAT:/REASSURE:/WHY:/FIX:/ESCAPE:` markers.
```

## Applying session's steps

1. Land/pull this webphone commit first; verify `git ls-remote`.
2. Stack tree must be CLEAN before the runbook commit (narHash covers
   the whole tree only at relock, but keep the tree clean regardless).
3. Apply the patch text above to ops-runbook.md § "Webphone error
   contract", commit, push.
4. If a relock rides the same train: webphone push → stack re-lock →
   pbx-artmann re-pin (full ritual in webphone docs/release-runbook.md).

## Records (T13, 2026-10-02 session)

- `do.InvokeAs` CONSIDERED-REJECTED for webphone: every dependency in
  the container is concrete (no interface registrations), so `Invoke`
  / `InvokeNamed` cover all needs; adopting `InvokeAs` would add a
  second resolution vocabulary for zero call sites. Re-litigate only
  if an interface dep ever lands in the container.
- MustInvoke call-site count corrected: ~33 REAL calls (a raw grep
  for "MustInvoke" shows 37 matches, 4 of which are comment lines).
  Any doc quoting "37 call sites" should read ~33 calls / 37 matches.
- Browser E2E NOT owed by this train: no served markup, DOM ids, or
  island assets changed (boot reporting is process-level, `cmd/` +
  docs only).
- Boot-contract re-audit rides the standing erraudit re-measure date
  **2026-10-22** (tiers 1+2 must stay 0; re-grade the boot surfaces
  against this contract in the same pass).
