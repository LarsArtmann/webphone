# Release Execution Session Status — v2.0.0 shipped, TODO_LIST emptied

**Date:** 2026-09-19 (session ~06:30–07:20 CEST, report 07:43) · **Trigger:** user
approved full execution of the pasted TODO_LIST ("Execute and Verify … until
everything works") · **Scope:** the 8-row TODO_LIST as of `ac56060`, executed
against the already-green P0–P7 state of
`docs/status/2026-09-19_00-05_cqrs-htmx-adoption-execution-status.md`.

**Headline:** webphone **v2.0.0 is tagged and on GitHub** (`d9d6d03`), all
lychee 404s resolved, the telephony stack consumes the release
(`fc6bc81`), and TODO_LIST went from 8 rows to 1. The glibc CVE row was
closed with evidence that contradicts what it claimed.

## a) FULLY DONE

1. **Baseline verified before touching anything.** Tree clean,
   `GOEXPERIMENT=jsonv2 go test ./...` green, fix `00f13fe` proven an
   ancestor of main, 28 unpushed commits counted at session start.
2. **Every "done" TODO row verified against code before deletion**
   (docs-health VERIFY, not trust-the-report): `eventsLimiter` wired at
   `GET /events` (server.go:112,157), hub reaper sweep with
   `hubIdleTTL = 10 * time.Minute` (sse.go:40-112), `hooksIdem`
   seen/record replay dedupe answering 202 (webhooks.go:227-275),
   `initSseLiveIndicator` + `#wp-sse-live` tripwires (server_test.go:298-299),
   `TestHookLimiterWrapsSecretGate` (ratelimit_test.go:101),
   `/openapi.json` route (server.go:171), deep-dive rubric score 62→92
   with appendix 09.
3. **glibc CVE-2026-5450 closed — the TODO row's premise was wrong.**
   nixpkgs PR #517918 ("glibc: 2.42-61 -> 2.42-67 … / CVE-2026-5928 /
   CVE-2026-5450") merged **2026-05-22**, backport #523468 2026-05-29;
   tracker issue #512409 closed as patched. Our locked nixpkgs
   (2026-09-17) ships glibc-2.42-84, and its `2.42-master.patch`
   explicitly carries the CVE fix (verified in the store path of the
   locked input). The 09-18 "still unfixed" reading was an **NVD
   range-match false positive** (vulnerable range 2.7–2.43 cannot see
   the distro patch suffix). vulnix re-run: the 8-derivation runtime
   closure carries **zero real advisories**; the zlib-1.3.2
   (CVE-2026-27820, 9.8) and unzip findings are build-closure-only and
   never deploy. AGENTS.md vulnix bullet rewritten with the mechanics
   (commit `7ba25cc`).
4. **Missing M39 deliverable added:** the webhook idempotency note
   (`hooksIdem`, successes-only recording, replay → `202 Accepted`) was
   implemented but never documented; now in AGENTS.md's Gateway seam
   bullet (`7ba25cc`).
5. **CHANGELOG release fold:** `## [Unreleased]` merged into
   `## [2.0.0] - 2026-09-19` (the 2.0.0 section had existed as prose
   since 09-18 but the **tag was never created** — that was the lychee
   404 × 2). 21 Added / 4 Changed / 9 Fixed bullets, one release
   (landed via daemon as `186c878`, see d.1).
6. **TODO_LIST refreshed:** 8 rows → 1 (new: `hookFaxStatus` error
   mapping). Completed rows deleted per docs-health, CVE row closed with
   the corrected story (`ee72831`). JS-test-runner standing gap routed
   to ROADMAP so it survives the report's expiry.
7. **Adoption plan annotated (docs-health ANNOTATE, inline):** 60 of 62
   M-rows struck with landing-commit hashes; open questions 2–3 resolved
   inline; §7 exit criteria marked **Met** (score 92); M54 (OOB badge
   push), M59 (sibling-report cross-link), open question 1 (XFF) left
   untouched because they are genuinely open. §4 micro table got a
   heading-level executed note (1:1 children of M-rows). `44db922`.
8. **Quality gates all green before push:** full test suite,
   `nix flake check` ("all checks passed"), buildflow **44 success /
   0 failed, exit 0**; the only lychee findings pre-push were the two
   known v2.0.0 bootstrap 404s.
9. **v2.0.0 released:** annotated tag created on `44db922`; pushed with
   main. Post-push lychee on CHANGELOG: **6/6 links OK, 0 errors**.
   `git ls-remote` confirms `refs/tags/v2.0.0` → `d9d6d03` and
   `refs/heads/main` → `44db9225`.
10. **Stack consumes the release:** nix-international-telephony
    `webphone` input `276c596e` → `44db9225`; `nix flake show` evaluates
    clean (default package `webphone-2.0.0`); the **stack's own**
    `.#packages.x86_64-linux.webphone` built green (exit 0, out-path
    `6q0nn0l7…`, in-sandbox test suite ran — different drv than
    webphone's pin because the stack's nixpkgs differs). Pushed as
    `fc6bc81` (message amended onto a daemon commit, see d.1).

## b) PARTIALLY DONE

1. **The push carried less than reported.** At push time origin/main was
   already at `6cdd75c` — a commit created mid-session by the daemon.
   My `git push` therefore moved only the last 5 commits
   (`6cdd75c..44db922`); the earlier 28 reached GitHub by some other
   channel (presumed daemon auto-push, mechanism unconfirmed). End state
   verified correct via `ls-remote`; the _timing_ claim "pushed at
   07:0x" is off by minutes for most of the release content. Effort to
   close: confirm the daemon's push behavior (S).
2. **GitHub Release object does not exist** — only the tag. The
   CHANGELOG link resolves to the tag page (lychee 0 errors), so the
   TODO's letter is met; release notes/visibility are missing. Blocked
   on: your preference (g.3). Effort: S.
3. **buildflow health: 9 tools unavailable, never itemized.** The gate
   passed (0 failed) but my `--verbose` grep failed to extract which 9
   tools were missing from PATH (go-licenses is a known one from the
   preflight log). A release gate with unidentified degraded tooling is
   a loose thread. Effort: S.
4. **Stack verification was package-scoped.** I built the webphone
   package in the stack's closure but did not run the stack's full
   `nix flake check` (FreeSWITCH/operator derivations unexercised with
   the new lock). The prior bump (`b775476`) appears to have been lock+
   eval only as well. Effort: M (mostly machine time).
5. **HARVEST of this report** — section (f) items are not yet routed
   into TODO_LIST/ROADMAP beyond what this session already wrote (the
   TODO row and the ROADMAP JS-gap line). Per docs-health, run HARVEST
   next session or on "continue". Effort: S.
6. **XFF open question (unchanged, external):** still unanswered;
   `remoteHostKey` stays the safe default. Not advanced this session.

## c) NOT STARTED

1. **`hookFaxStatus` error-mapping split** (every store error → 404;
   message counterpart distinguishes 404/500) — the one TODO_LIST row;
   waiting for a session with the provider-retry contract in mind.
2. **Release runbook** — the fold → tag → push → lychee re-check →
   stack bump → build → push dance now exists only in this session's
   history; not documented anywhere durable.
3. **aarch64 re-verification for the tag** — `nix flake check` warned
   "omitted these incompatible systems: aarch64-linux"; the release's
   cross-build claim still rests on the 09-18 verification.
4. **FEATURES.md freshness pass** — not examined this session; the
   release fold touched CHANGELOG/TODO/ROADMAP/AGENTS only.
5. **ROADMAP items untouched by design:** M54/UB1 OOB badge push (gated
   on stack E2E loop), JS test runner + tripwire port, MD1 root-tag
   bump checklist, `ClientIP`-trust note upstream, M59 structural-health
   cross-link (may deserve NOT-DO instead — no sibling report exists in
   this repo).
6. **Go module-path policy for v2+** — v2.0.0 tags on an unsuffixed
   module path are not `go get`-able; harmless for a flake-consumed app,
   but no decision record exists.

## d) TOTALLY FUCKED UP

1. **Auto-commit daemon raced me three times, and I repeated the
   previous session's documented mistake.** (a) AGENTS.md split across
   daemon commit `e5cdc6f` and my explicit `7ba25cc`; (b) the **release
   fold — the most consequential docs change of the session — is buried
   under the meaningless message "chore: auto-commit 2 changed file(s)"
   (`186c878`)**; I only patched the story into the follow-up TODO_LIST
   commit body; (c) the stack lock bump landed as `8dbccbb` heuristic
   before my explicit commit; that one I fixed by amending the unpushed
   commit (`fc6bc81`). Root cause: batching multiple doc files between
   explicit commits. The 00-05 report already listed this exact failure
   (§d.3) and I did it anyway. Severity: history noise, no content
   damage (verified per commit).
2. **The daemon appears to auto-push, and my mental model didn't
   survive contact.** Discovered via the push range (see b.1). If the
   daemon pushes, then "N unpushed commits" and "operator pushes when
   ready" assumptions are wrong, and future sessions could believe
   something is safely local when it is public (or vice versa).
   Mitigation: I verified the end state via `ls-remote` before
   declaring victory; mechanism still unconfirmed.
3. **vulnix fumbling nearly produced a wrong security claim.** Three
   unusable invocations (help text on stdin mode, `--print-out-path`
   flag that doesn't exist, args mode silently expanding **build**
   closures) before landing on: pass runtime paths as args **and
   cross-check findings against `nix-store -qR` output**. In between,
   zlib CVE-2026-27820 (9.8) looked like a new runtime exposure — it is
   build-closure-only. AGENTS.md documented this exact trap and I still
   burned four commands re-deriving it; the honest number came from a
   grep-intersection, not from the tool's default behavior.
4. **Two wasted write cycles on TODO_LIST.md** — the daemon/prettier
   reflowed the file between my read and write, twice ("modified since
   read"). Sloppy sequencing; should have been one read→write→fmt pass.
5. **`agentic_fetch` failed twice (API error) on the CVE research
   before I switched to `gh search`/`gh pr view`, which answered
   everything in one round each.** Wrong tool first; the fallback order
   should be: GitHub-hosted truth → gh CLI; web search last.
6. **False cross-reference shipped by the previous session, noticed,
   not fixed:** the 00-05 report §g.2 says "see the annotated status
   report of 21:38" — that report has **zero** annotations. I left it
   (declared out of session scope); it is now itemized in (f).
7. **`nix fmt` coverage ambiguity unresolved:** after my table edits it
   reported "formatted 0 files (0 changed)" — I did not verify whether
   the annotated plan file is actually inside treefmt's md scope or
   simply ignored. If ignored, the struck tables' alignment drifts
   forever.

## e) WHAT WE SHOULD IMPROVE

1. **Commit each task the moment it is green — one file-batch per
   commit, no exceptions.** Third consecutive session with daemon-race
   damage. Concrete rule: `git status --short` → `git add <the files of
   THIS task>` → explicit commit, then start the next file.
2. **Script the vulnix runtime-closure check** (flake app or devShell
   alias: `nix-store -qR <pkg> | vulnix`-equivalent with the
   findings-vs-closure cross-check built in) so the next CVE question is
   one command, not four rediscoveries. Include the NVD-range-FP caveat
   in its output.
3. **Use `docs-health`'s `annotate-rows.py` for table annotations**
   instead of hand-rolled regex scripts. My script worked (60/60, no
   cell mismatches), but it bypassed the tooling's atomicity/dry-run
   discipline that exists precisely for this.
4. **Write the release runbook down** (section c.2): fold-Unreleased →
   dated section → explicit commit → gates → annotated tag → push →
   lychee re-check → stack lock bump → stack build → push. Eight steps,
   all of which I derived live.
5. **Verify cross-references at write time** — "annotated" is a checkable
   state (`grep -c '~~'`), and the 00-05 report shipped one that was
   false.
6. **Gate releases on `nix flake check --all-systems`** so aarch64 is
   not silently omitted from the release evidence.
7. **Confirm the daemon's push behavior once** (watch a commit land and
   check `ls-remote` after a few minutes) and document it in AGENTS.md —
   it changes what "pushed" means for every future session.

## f) Up to 50 things we should get done next

Ranked by impact; routing in brackets. This is a brainstorm per the
skill — most non-top items are ROADMAP fuel, not commitments.

**Release/deployment follow-through**

1. Create the GitHub Release object for v2.0.0 with generated notes
   (link already resolves; this adds notes + notifies watchers). [S /
   High / Cleanup] — pending g.3
2. Confirm the auto-commit daemon's auto-push behavior; document in
   AGENTS.md what "pushed" means. [S / High / Documentation]
3. Run the stack's full `nix flake check` with the new webphone lock
   (FreeSWITCH/operator closures unexercised this session). [M / High /
   Quality]
4. Re-verify the aarch64-linux cross-build for the v2.0.0 tag
   (`nix build .#webphone --system aarch64-linux` or
   `nix flake check --all-systems`). [S / Medium / Quality]
5. Itemize buildflow's "9 tools unavailable" (go-licenses known
   missing); fix PATH so release gates run at full strength. [S /
   Medium / Quality]
6. Decide the stack's webphone input policy: pin to tag refs
   (`?ref=v2.x`) with explicit bumps vs track main. [S / High /
   Decision] — pending g.2
7. Write the release runbook (fold → tag → push → lychee → stack bump →
   build → push) into AGENTS.md or a flake app. [S / Medium /
   Documentation]
8. Consider a tag→release CI workflow so future tags cut releases
   automatically. [M / Low / Feature]
9. Stack-side switchover decision (AGENTS.md still says "remains open
   work"): import `nixosModules.default` vs keep reverse-proxying.
   [M / High / Decision]

**Code quality / bugs**

10. Split `hookFaxStatus` error mapping (404-on-every-store-error vs
    the message counterpart's 404/500), designed against the
    provider-retry contract. [S–M / Medium / Bug] — already the lone
    TODO_LIST row
11. Decide webhook idempotency store durability: memory TTL means a
    provider replay after restart re-applies; is that acceptable?
    Decision record + test if SQL is chosen. [M / Medium / Decision]
12. Session store is in-memory: restart drops all sessions. Decision
    record (acceptable for a PBX-fronted app?). [S / Low / Decision]
13. `/version` reports `(devel)` for non-module builds; inject the git
    tag via ldflags in the flake build. [S / Low / Feature]
14. Extend `/openapi.json` beyond `/api/session` (hooks? phone-api?) or
    record the boundary as deliberate. [M / Low / Documentation]
15. Go module-path policy for v2+ tags (`/v2` suffix vs NOT-DO record
    for a flake-consumed app). [S / Low / Decision]

**Docs integrity (docs-health)**

16. Fix the false cross-reference: 00-05 report §g.2 calls the 21:38
    report "annotated" — it has zero markers; annotate it or correct
    the reference. [S / Low / Documentation]
17. HARVEST this report's (f) into TODO_LIST/ROADMAP with evidence
    citations. [S / High / Documentation]
18. Verify the annotated plan file is inside treefmt's prettier scope
    (`nix fmt` said 0 changed — confirm that means covered-and-clean,
    not ignored). [S / Low / Cleanup]
19. FEATURES.md freshness pass against v2.0.0 (not examined this
    session). [S / Medium / Documentation]
20. Annotate the remaining 2026-09-18 status reports (18:50 sweep,
    22:08 CSP, 22:55 dedup) — all still have zero inline resolution
    markers; their "next tasks" fed work that is now done. [M / Low /
    Documentation]
21. Add the exact working vulnix runtime-closure invocation to
    AGENTS.md (including the findings-vs-closure cross-check). [S /
    Medium / Documentation]

**Upstream/library track**

22. XFF sanitization answer from the stack → flips `remoteHostKey` to
    `KeyExtractorFromClientIP` (standing open question). [S / High /
    Decision] — pending g.1
23. MD1: on the next cqrs-htmx root tag (SSE `retry:` hint), follow the
    ROADMAP bump checklist (AGENTS note, `BenchmarkHubFanOut`, upstream
    E2E). [M / Medium / Maintenance]
24. M54/UB1: OOB badge push — un-park only with the stack E2E loop
    available. [M / Medium / Feature]
25. M59: create the structural-health sibling report and cross-link it,
    or NOT-DO the row with a reason (no such report exists here). [S /
    Low / Decision]
26. CSP: watch templ-components for a ThemeScript opt-out knob; take it
    and drop the hash + `!important`s (standing AGENTS.md note). [S /
    Low / Maintenance]
27. Upstream `ClientIP`-trust note in cqrs-htmx httputil if XFF turns
    out sanitized. [S / Low / Documentation]

**Testing / robustness**

28. Island JS test runner decision; port the asset tripwires (toasts
    listener, live pill, 429 surfacing) into real DOM tests. [M / Low /
    Quality] — already in ROADMAP
29. Re-run the stack browser E2E once against the released binary as a
    deployment smoke (verdict said optional; a release is a natural
    trigger point). [M / Medium / Quality]
30. `remoteHostKey`/limiter: add a runbook line for the day login
    limiter keys need widening (NAT offices). [S / Low /
    Documentation]

**Hygiene**

31. Nixpkgs lock cadence: weekly `nix flake update` + vulnix rescan
    ritual — the CVE row proved lock state and doc claims drift. [S /
    Medium / Maintenance]
32. Investigate `signed` tags for future releases (`git tag -s`).
    [S / Low / Cleanup]
33. README deployment section: verify it references the NixOS module
    and a real tag (claim freshness, no research done this session).
    [S / Medium / Documentation]
34. Auto-commit daemon config: consider excluding `docs/status/` and
    `flake.lock` from heuristic commits so explicit messages win
    (infra change, upstream decision). [S / Medium / Cleanup]

_(34 items — stopping here rather than padding to 50; the remainder of
"what's next" is already represented by ROADMAP's existing bullets.)_

## g) Three questions I cannot figure out myself

1. **Does the telephony stack's reverse proxy sanitize
   `X-Forwarded-For` before requests reach webphone?** I tried nothing
   this session (it lives in the stack's nginx/freeSWITCH config, not
   this repo, and the standing open question says prior sessions
   couldn't resolve it either). The answer gates the flip from
   port-stripped peer-host keys (`remoteHostKey`) to
   `KeyExtractorFromClientIP` in the login/hook/events limiters — i.e.,
   whether rate limiting is per-IP or per-host behind the proxy.
2. **Should the stack pin `webphone` by tag ref (`github:LarsArtmann/webphone?ref=v2.0.0`)
   instead of tracking main?** Today's bump landed on `44db9225` (which
   happens to be the tagged commit), but the input still follows main —
   the next main commit silently changes the deployment input at the
   next `nix flake update`. I can't decide your upgrade cadence/risk
   posture for you.
3. **Do you want GitHub Release objects per tag (tag → release notes,
   potentially automated via workflow), or are tags + CHANGELOG
   enough?** The TODO row only required the tag (links resolve —
   verified), so I stopped at the tag. This decides follow-up #1 and #8
   in section (f).
