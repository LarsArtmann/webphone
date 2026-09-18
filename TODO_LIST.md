# TODO_LIST

Short- and mid-term improvement tasks. Done work is deleted, never struck
through (docs-health style: one home per fact, no decay). Long-shot ideas
live in FEATURES.md as WORTH_CONSIDERING.

| Task                                                                                                                                                                         | Status    | Priority | Effort | Evidence / notes                                                                                            |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | -------- | ------ | ----------------------------------------------------------------------------------------------------------- |
| Push the accept/reject fix (`00f13fe`) and bump the telephony stack's `webphone` input lock to it (currently pinned to the pre-fix `276c596`)                                | 🔴 `TODO` | High     | S      | Browser E2E green with the fix (local override run 2026-09-18); stack lock bump is one command after push   |
| Tag v2.0.0 on GitHub so CHANGELOG compare/release links resolve (lychee 404 × 2)                                                                                             | 🔴 `TODO` | Medium   | S      | `CHANGELOG.md`; buildflow lychee findings 2026-09-18                                                        |
| Track glibc CVE-2026-5450 (9.8) until nixpkgs ships a fix — nixpkgs bumped to 2026-09-17 and it is still unfixed; runtime closure is otherwise clean (8 derivations, vulnix) | 🔴 `TODO` | Medium   | S      | vulnix runtime-closure scan 2026-09-18; build-time-only advisories (binutils/gcc/bootstrap Go) never deploy |
