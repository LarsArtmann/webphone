# Status — 2026-10-12 10:52 CEST — SUPERB train COMPLETION session (T08 green; T11/T12/T14 closed)

Session scope: resumed after the 06:55 report (which ended with "WAIT FOR
INSTRUCTIONS"); this session executed the standing instruction to finish
everything short of the owner-gated boundary. **T08 went from red (run 3) to
green (run 10) across nine VM runs, each isolating exactly one real defect**;
T11 (webphone train-end gates), T12 (stack gates + browser E2E + push/CI),
and T14 (tri-repo docs) are closed. T13/T15 remain untouched (owner-gated).
Supersedes: [2026-10-12_06-55_SUPERB-unified-auth-train-status.md](2026-10-12_06-55_SUPERB-unified-auth-train-status.md).

## (a) FULLY DONE (implemented, verified, committed, pushed)

1. **T08 — dynamic-directory VM suite GREEN (run 10).** The full seam is
   proven in one NixOS VM: bridge boots + serves the spliced secret, static
   directory scoped (1000 absent / 1001 present), login answers
   `sip_credential` (never `password`), F46 derived-credential REGISTER=200,
   F47 wrong credential denied AND logout revocation visible at the seam
   (`credentials: []`) AND at the next REGISTER (denied, with the
   `auth logout` audit line), F48 static-fallback extension registers with
   its file password. The nine-run defect chain (each red run = one real
   bug, isolated and fixed):
   1. stack pinned an OLD webphone (no `sip_credential`) → flake.lock bump
      to the train HEAD (T12-part, prior session).
   2. root-owned secret file → webphone honestly refused boot → fixture
      `user = "webphone"` (prior session).
   3. **nixpkgs' FreeSWITCH omits `xml_int.curl`** — mod_xml_curl
      CRIT-failed to load SILENTLY (403 on every dynamic REGISTER) →
      `pbx.nix` now mirrors the default module list + `xml_int.curl`
      (documented inline; a dropped module fails loudly in the suites).
   4. **mod_xml_curl v1.11.1 has NO bearer scheme** — its auth-scheme
      parser knows basic/digest/NTLM/GSS/any and SILENTLY IGNORES unknown
      values (verified in FS source) → the bridge accepts HTTP Basic
      (`webphone:<secret>`, constant-time compare of the password half)
      plus Bearer for probes; the binding renders `basic`.
   5–7. post-logout register still 200'd → **instrumentation bisected it**:
      webphone was INNOCENT — the TEST fixture logged in without
      `curl -c` (session cookie never saved → DELETE was a silent no-op
      204 with no audit line) and held the PRE-login CSRF token (login
      rotates the pair) → fixture fixed (cookie save + post-login
      `/api/csrf` re-adoption, the island's own flow).
   8. **wsprobe's REGISTER template hardcoded user 1000** in
      From/To/Contact (the `user` arg only fed the digest) → parameterized
      across all six call sites.
   9. GREEN.
   Kept as defense-in-depth: sofia `accept-blind-reg=false` +
   `accept-blind-auth=false` rendered under xml_curl only (static mode
   stays byte-identical golden) — an unknown user must be DENIED, never
   blind-accepted, or logout revocation would be a lie.
2. **T11 — webphone train-end gates as ONE unit, all green:** `nix fmt` →
   buildflow (**67 success / 0 failed / 0 skipped**) → `go test -count=1
   ./...` → `nix flake check` (**all checks passed**) → new curl-level
   passkey+brokering smoke **13/13** (boots with `auth.passkey.*` +
   `sip.credential_secret`; proves healthz userauth leg, passkey surfaces,
   `sip_credential`-only wire, internal-directory equality with the login
   answer, 401 gates, uniform passkey-finish refusal) → push under the
   ratified manual-push bar (anchor: first unpushed commit 06:53:35, bar
   met ~07:53) → **CI green** on `6fc1d79`.
3. **T12 — stack gates + browser E2E + push/CI, all green:** `nix fmt`
   clean; **full `nix flake check` EXIT=0** (every VM suite, treefmt,
   island lint); the browser E2E (`telephony-webphone`) green after
   teaching `tests/configjs_check.py` the new `PBX_CONFIG.sipCredential`
   key (type-checked — exactly the drift the runbook's "served-markup
   changes owe a fresh stack browser E2E" rule exists to catch);
   `telephony-messaging`'s broken-pipe flake attributed (10-way concurrent
   VM load; green on solo rerun); stack pushed (daemon recovered) at
   `3b63222`; **HEAD CI = success**; the older failing CI run diagnosed as
   EXACTLY the pre-fix configjs assert — root cause confirmed closed.
4. **T14 — tri-repo docs closed** (F75–F77): `docs/error-contract.md`
   rows (passkey revoke failure, unmapped-revoke 404, internal-directory
   seam contract incl. the no-`sipcred.*`-codes rationale); AGENTS.md
   "Owner decisions (2026-10-12)" (brokering contract + passkey
   mgmt/audit, 144/377 lines); webphone CHANGELOG (brokering, passkey
   management, audit trail under Added; identity-model import under
   Changed); stack CHANGELOG Added entry; stack AGENTS dynamic-directory
   bullet (incl. the silent-bearer trap); stack runbook
   **§ Dynamic SIP directory** with the <2-min ROLLBACK drill (F72).
   **pbx-artmann deliberately untouched** — its changelog entry belongs
   to the owner-gated cutover commit.
5. **Plan execution record** updated with honest T06/T07/T08 rows —
   including the correction of my own premature "run 4 green" claim (see d1).

## (b) PARTIALLY DONE

1. **`configjs_check.py` type-checks `sipCredential` but does not
   value-pin it** — an xml_curl-mode E2E fixture asserting `true` would
   pin the module wiring end-to-end; current fixture answers `false`
   (passkey mode off there).
2. **VM concurrency flake hardening** — the messaging broken-pipe was
   attributed and cleared, but `nix flake check` still runs ~10 VM suites
   concurrently on one host; no jobs-cap/serial-VM-phase added.
3. **HARVEST from section (f) into `TODO_LIST.md`** not run — the owner's
   "THEN WAIT FOR INSTRUCTIONS" gates it; the report holds the list.
4. **The 06:55 report is not annotated** with what this session closed
   (docs-health ANNOTATE pass pending — old items there are stale now).
5. **Stack CI history** — the two failing runs before HEAD's are
   explained (one = the configjs drift, fixed; the rest predate this
   session), but no post-mortem note was left in the stack's
   ci-flake-ledger equivalent.

## (c) NOT STARTED

1. **T13** — pbx-artmann relock + staged prod cutover. Owner-gated;
   pbx-artmann has ZERO changes from this train.
2. **T15** — prod verification (derived-credential REGISTER live, logout
   revocation live, static extensions unaffected, rollback drill ONCE on
   prod). Owner-gated.
3. **F49 in-VM rollback flip drill** — eval-golden only so far (owner
   question open).
4. **nixpkgs upstream PR** — proposing `xml_int.curl` for the default FS
   module set (or at least documenting the silent-ignore trap) — decided
   against unilaterally; upstreaming is an owner call.
5. **Plan's deferred list (unchanged)** — membership-remap, loginpage
   ruling, TOTP/OAuth2 login modes, operator-API session auth,
   extension-login lockout parity, passwordFile permission assertions.

## (d) TOTALLY FUCKED UP (mine, honest)

1. **I pre-wrote "run 4 — green" into the plan's execution record BEFORE
   the verdict.** Run 4 was RED (the bearer discovery). Claim-ahead-of-
   verification is the exact anti-pattern the verify skills exist for. I
   caught it, rewrote the row honestly, and the final record states the
   full run history — but it happened.
2. **Three VM runs (≈25 min) burned on MY OWN test-fixture bugs after I
   had the evidence to instrument.** Run 5's failure (post-logout
   register 200) was met with TWO theory-driven fixes (blind-reg render,
   then a wrong-cache theory) instead of the one instrumented run that
   actually localized everything (bridge answer-count journal + seam
   assert). The PAIR1-DIAG "instrument first" lesson — which the stack's
   own runbook already records — was re-learned the expensive way.
3. **Run 6's `accept-blind-reg` fix was built on an unverified mental
   model** of sofia's flag semantics (turned out inverted-ish and a no-op
   for the observed flow). Kept because the fail-closed posture is right,
   but the run itself was waste caused by not reading sofia_reg.c first.
4. **The bearer wire-shape assumption rode unverified from the plan's
   research straight into T07's config.** One source read of
   mod_xml_curl.c BEFORE wiring would have saved runs 4–5 entirely. The
   verify-external-claims gate fired three runs too late.
5. **Vacuous-pass asserts in my smoke harness** — "internal directory
   answers the SAME credential" passed with two empty strings. Fixed
   with an explicit non-empty guard + early abort, but it shipped once.
6. **Near-miss repeat of (1):** the T08 row nearly claimed green at run 8;
   the honest-history rewrite landed only because the verdict was polled
   before delivery. Record-after-verify is now the personal rule.

## (e) WHAT WE SHOULD IMPROVE (larger > smaller)

1. **Instrument-first, always** — one debug-instrumented run beats two
   theory runs; make the bridge's answer-count journal and seam-content
   asserts the TEMPLATE for future seam suites, not a one-off.
2. **Verify consumer-source claims BEFORE wiring** — any "FS supports X"
   / "library accepts Y" fact in a plan gets a source-check before it
   becomes config (would have saved 2 VM runs + a package rebuild here).
3. **Negative-arm tests must vary ALL identity axes** — the hardcoded
   user 1000 in wsprobe masked the static-fallback path for nine runs;
   parameterize probes completely when writing them.
4. **Assert the observable seam between steps** — "post-logout directory
   must be empty" turned a mystery into a one-line bisection; intermediate
   state asserts belong in every multi-party test.
5. **VM-suite concurrency cap** — `nix flake check` should serialize or
   cap VM jobs (NIX_BUILD_MAX_JOBS-style); today's broken-pipe flake was
   host RAM/CPU pressure, and CI has the same shape.
6. **Daemon-commit ordering discipline** — with the daemon interleaving,
   any "record then verify" ordering is dangerous; the rule is now
   record-AFTER-verify, and verification claims cite the exact run/log.
7. **Smoke harness vacuous-pass guards** — never compare extracted values
   without an emptiness gate; early-abort beats a green lie.
8. **Daemon push-stall visibility** — both daemons' push legs stalled
   today (>60 min); the manual-push bar worked but needed manual
   attention; a watchdog alert (push-lag metric) would close the gap.
9. **CI-run registration gap** — webphone `25cf08d` (docs-only, 08:08) has
   NO CI run registered at 10:52; harmless for docs but the pipeline
   should register (and skip-with-verdict) every push.
10. **Stack ci-flake ledger** — the stack keeps flake history in prose;
    today's messaging broken-pipe + the pre-fix CI failures deserve
    ledger rows so the next session doesn't re-dig.

## (f) 50 next items (impact-sorted; ☐ open; owner-gated marked)

**Cutover + prod (owner-gated cluster):**
1. ☐ **T13**: pbx-artmann relock to stack `3b63222` (ritual: clean tree,
   narHash covers all; release-runbook dance) — OWNER-GATED.
2. ☐ T13: stage the prod deploy with `telephony.directory.mode = "xml_curl"`
   + explicit `staticFallbackExtensions` (operator, ring groups) — OWNER-GATED.
3. ☐ T13: verify the derivation secret file lands on prod with correct
   owner/mode for BOTH webphone (DynamicUser read) and the bridge
   LoadCredential — OWNER-GATED.
4. ☐ T13: pre-cutover backup + the <2-min rollback rehearsed on staging
   (mode flip → switch → static REGISTER) — OWNER-GATED.
5. ☐ **T15**: prod derived-credential REGISTER (browser + one softphone).
6. ☐ T15: prod logout → next REGISTER denied (live revocation proof).
7. ☐ T15: prod static extensions unaffected during brokering.
8. ☐ T15: prod rollback drill executed ONCE and reverted (F81).
9. ☐ T13: pbx-artmann CHANGELOG entry written IN the cutover commit.
10. ☐ T13: post-cutover CI + journal watch (`auth logout`, `directory
    query` lines prove the seam is live).

**Answer the owner's open questions (blocking items 1–4):**
11. ☐ Owner: cutover timing + staging shape (question g1).
12. ☐ Owner: F49 scope ruling — eval-golden enough or in-VM flip drill?
13. ☐ Owner: baseline-regrowth ruling from the 06:55 report (still open).
14. ☐ Owner: pursue the nixpkgs `xml_int.curl` upstream PR or keep
    stack-local (with the silent-ignore trap documented upstream either way)?

**Test/infra hardening:**
15. ☐ VM jobs cap / serial VM phase for `nix flake check` (local + CI).
16. ☐ Value-pin `sipCredential: true` in an xml_curl-mode E2E fixture.
17. ☐ Port the bridge answer-count journal + seam assert pattern into the
    OTHER webphone suites (shared helper in tests/common.nix).
18. ☐ Shared `secretFile` fixture helper (`user = <service>` pattern) in
    tests/common.nix (the run-2 lesson, still copy-pasted).
19. ☐ Stack ci-flake ledger rows: messaging broken-pipe (host load), the
    configjs drift, the pre-fix CI failures.
20. ☐ Daemon push-stall watchdog (alert when origin/main lags HEAD >60 min).
21. ☐ CI run registration for every push (or explicit skip verdict) —
    investigate why 25cf08d has no run.
22. ☐ wsprobe: keep the parameterized user as the template default and
    add a `--assert user=1001 static` leg to the main webphone suite.
23. ☐ Consider merging the dynamic-directory suite INTO telephony-webphone
    (one machine, halved CI cost) if assertions stay readable (06:55 e5).
24. ☐ Add the `auth logout` audit line to the dynamic-directory suite's
    asserts (it silently proved the no-op logout bug — make it load-bearing).
25. ☐ Smoke-harness guard pattern (non-empty before compare) upstreamed
    into webphone-smoke.py's helper style.

**Docs continuity:**
26. ☐ docs-health HARVEST: move (f) items into TODO_LIST.md/ROADMAP.md.
27. ☐ docs-health ANNOTATE: mark the 06:55 report's resolved items inline.
28. ☐ Stack runbook: cross-link § Dynamic SIP directory from the webphone
    failure playbook's index.
29. ☐ Webphone README § Derived SIP credentials: fill the "full contract:
    filled in the docs train" stub with the error-contract link.
30. ☐ Add the mod_xml_curl silent-bearer trap to the stack's
    docs/lessons/freeswitch.md (long-form home for the AGENTS bullet).

**Product/deferred-plan items (ROADMAP fuel):**
31. ☐ Operator-API session auth (shrinks the file-outage class the plan
    names as the residual risk).
32. ☐ Extension-login lockout parity (passkey vs password flood buckets).
33. ☐ TOTP/OAuth2 login modes (explicitly deferred this train).
34. ☐ Membership-remap ruling.
35. ☐ loginpage ruling.
36. ☐ passwordFile permission assertions in the module (dataDir-style).
37. ☐ Upstreamable nixpkgs improvement: assert-or-warn on unknown
    auth-scheme values in mod_xml_curl (silent-ignore is the trap).
38. ☐ Bridge: consider an optional allowlist of extensions it will proxy
    (defense-in-depth vs. webphone compromise; needs a threat-model pass).
39. ☐ `webphone -check-config` flag that validates the brokering pair
    (secret file readable + directory endpoint self-probe) for preflight.
40. ☐ Island: show a registration-state hint when `sipCredential` is on
    and REGISTER fails (today the island retries; operator sees journal).

**Small cleanups:**
41. ☐ Delete the run-by-run scratch files (/tmp/t08-*) or archive the
    final green log next to the suite.
42. ☐ Name the smoke harness: promote /tmp/wp-passkey-smoke.sh into
    scripts/ (it is the only curl-level passkey+brokering E2E).
43. ☐ `.buildflow-baseline.json` audit (the +8604-line regrowth, still
    unruled — 06:55 question 1).
44. ☐ Re-measure error-family tiers 1+2 (due 2026-11-05 per AGENTS).
45. ☐ Quarterly art-dupl baseline audit (due per the 2026-10-11 ruling).
46. ☐ Stack: consider `Cache-Control` on the bridge's not-found (already
    no-store on answers; verify not-found answers carry it too).
47. ☐ freeswitch.nix: assert at eval time that the mirrored module list
    matches nixpkgs' default set (diff the two lists, fail on drift).
48. ☐ Document the `--resolve webphone.test` smoke trick (IP rp_id is
    rejected by WebAuthn — hostname + resolve is the pattern).
49. ☐ wsprobe: exit-code documentation block (ok/denied/drop semantics).
50. ☐ Update AGENTS.md Tri-repo rules with the FS-module-mirror fact (one
    line; the pbx-artmann consumer inherits the FS package change).

## (g) 3 questions for the owner (unchanged from 06:55 — still blocking)

1. **Prod cutover (T13):** T08/T11/T12 are green and pushed — do you want
   the pbx-artmann relock + staged cutover queued as the NEXT session,
   and should prod's `staticFallbackExtensions` list your operator/ring-
   group extensions explicitly before the flip? (Items 1–4 above.)
2. **F49 scope:** is the eval-level rollback proof (F41 golden,
   byte-identical static rendering) acceptable, or do you want the full
   in-VM static↔xml_curl flip drill (adds ~10 min CI + a second toplevel)?
3. **Baseline regrowth:** `.buildflow-baseline.json` grew +8604 lines in
   a concurrent session (after a staleness warning said 19 entries
   matched nothing). Audit/trim, or deliberate SDK-side re-save?

## Session inventory (what changed where, this session)

- **webphone** (clean, synced; code CI-green at `6fc1d79`; HEAD
  `25cf08d` is docs-only): plan-record rows T06–T08 (honest run
  history), CHANGELOG entries, AGENTS 2026-10-12 decisions,
  error-contract rows. No Go/JS changes this session — the train's code
  was already green; this session's code changes were all stack-side.
- **nix-international-telephony** (clean, pushed, HEAD CI success at
  `3b63222`): pbx.nix FS module-list mirror (+`xml_int.curl`);
  freeswitch.nix Basic binding + blind-reg guard; fs-directory.py
  Basic-auth + answer-count journal; wsprobe.py user parameterization;
  webphone-dynamic-directory.nix seam/render asserts + fixture fixes;
  configjs_check.py `sipCredential` key; ops-runbook § Dynamic SIP
  directory; AGENTS + CHANGELOG rows. One infra flake (messaging)
  attributed + green on rerun.
- **pbx-artmann:** untouched (owner-gated T13).
- **Key gates at report time:** webphone buildflow 67/0, flake check ✓,
  smoke 13/13; stack flake check EXIT=0; both HEADs pushed; stack HEAD
  CI success; webphone last code-bearing run success (HEAD's docs-only
  commit has no run registered — item 21).
