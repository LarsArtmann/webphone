# Status — 2026-10-12 15:09 CEST — SUPERB harvest + hardening session (non-gated (f) cluster executed; both reports annotated; stack main RED — upstream pbx-lens vendorHash, NOT this session's code)

Session scope: resumed after the 10:52 report with the standing
"READ, UNDERSTAND, RESEARCH, REFLECT … execute one step at a time"
mandate. This session executed the 10-52 report's non-gated (f) items
end to end (owner-gated T13/T15 + the three §g questions untouched),
harvested what remains into TODO_LIST/ROADMAP, annotated BOTH prior
reports inline, fixed a formatter-drift red on webphone main, and
diagnosed the stack's fresh red to its true (upstream) root cause.
Supersedes: [2026-10-12_10-52_SUPERB-train-completion-status.md](2026-10-12_10-52_SUPERB-train-completion-status.md).

## (a) FULLY DONE (implemented, verified, committed, pushed)

1. **Item 21 (CI registration gap) — myth-busted.** `25cf08d` HAS CI run
   `38211100788` (push, success, 06:08 UTC) — the 10-52 report's claim
   was stale. `c2caead` had no run because it was never pushed (the
   daemon's push leg stalled — the item-20 class); once pushed it also
   registered and passed. Conclusion: the pipeline registers EVERY push;
   there is no registration gap. Recorded in both report annotations +
   the TODO_LIST closed-cluster note.
2. **Item 15 (VM concurrency cap) — `nix run .#check-vm`.** A sequential
   VM-suite driver in the stack flake: discovers every `vm-test-run-*`
   check at eval (list can't drift), runs them ONE at a time (the
   messaging broken-pipe was ~10 concurrent VM boots on one host),
   accepts subset names. Discovery expression proven; the app itself
   dogfooded — both changed suites ran through it green (`rc=0`,
   `/tmp/checkvm.log`, 4090 lines of evidence).
3. **Item 16 (value-pin) — config.js answers `true` under xml_curl.** The
   dynamic-directory suite now asserts
   `curl /config.js | grep '"sipCredential":true'` (module → settings →
   island login-order flag, end to end; the browser E2E's static-mode
   fixture covers the false arm). VM-verified green.
4. **Item 24 (`auth logout` audit line load-bearing).** The suite now
   greps the journal for `auth logout extension=1000 outcome=ok` right
   after the DELETE — a silent no-op logout (the fixture bug the
   five-red-run chain burned on) can no longer pass. VM-verified (run
   log line 4017: the audit line present, assert passed).
5. **Item 22 (static auth-register leg).** `telephony-webphone` gained
   `--auth-register 1001 'test-1001-u6t5s4' ok proxied` — a REAL 200
   digest REGISTER against the static directory (the `--assert` legs
   only ever proved 401 challenges). VM-verified green.
6. **Item 47 (FS module-mirror drift check) — with a research-backed
   design.** nixpkgs' `defaultModules` is an expression-local let (never
   exported — verified against the pinned rev), so a pure-eval diff is
   impossible. Landed instead: the mirror list hoisted to ONE home
   (`modules/telephony/freeswitch-modules.nix`, consumed by pbx.nix's
   override), and `telephony-eval` resolves every entry against nixpkgs'
   module universe (imported with `builtins.functionArgs` auto-fill —
   no hand-listed arg names). Stale entry ⇒ eval fails with
   `attribute '<name>' missing at freeswitch-modules.nix:<line>`; the
   check also greps the resolved list still carries `xml_int/mod_xml_curl`.
   BOTH arms proven: positive (`true` + check rc=0), negative (precise
   eval failure). Two Nix-semantics discoveries encoded in comments (see
   d/e): attr-missing ESCAPES tryEval on Nix 2.34 (verified empirically:
   tryEval catches `throw`, not attr-missing), and deep forcing
   (`deepSeq`/`all`) of module records pulls derivation inputs → stdenv
   bootstrap → pure-eval stack overflow — the probe stays WHNF-level
   (`map (m: m.path)`).
7. **Item 46 (bridge Cache-Control) — verified, no code change needed.**
   The not-found fall-through rides the SAME 200 response as the
   credential answer (`fs-directory.py` send path), so the single
   `no-store` header covers both shapes; comment pinned to state it.
8. **Items 49/28/29/30/48/50 (docs continuity).** wsprobe header gained
   the exit-code block (probe=0 observational, `--assert`/`--auth-register`
   exit 1 + verdict lines); webphone error-contract's internal-directory
   row now links the stack runbook § Dynamic SIP directory; the README
   "full contract: filled in the docs train" stub is FILLED (error-
   contract rows + stack runbook pointers); the mod_xml_curl
   silent-bearer trap + xml_int.curl absence got the long-form home in
   stack `docs/lessons/freeswitch.md`; the `--resolve webphone.test`
   smoke trick (WebAuthn rejects IP rp_id) landed in webphone AGENTS §
   Commands; AGENTS tri-repo rules gained the FS-module-mirror fact.
9. **Item 19 (stack CI ledger).** `docs/ci-posture.md` gained the
   "2026-10-12 red-run ledger": the lock-guard reds + notify-failure fix
   iterations (sibling lane, pointed at), the configjs drift red
   (`c3ec2d6`, real drift, fixed same day, HEAD green at `3b63222`), and
   the messaging broken-pipe (infra, green solo, mitigated by check-vm).
10. **Item 26 (HARVEST).** TODO_LIST: new top section "SUPERB
    unified-auth train tails" (owner-gated T13/T15 cluster; stack
    hardening tail; webphone/infra tail; the deliberately-NOT-minted
    closed cluster) + Last-sweep line. ROADMAP: three owner questions
    into § Open questions (cutover timing/staging, F49 scope, baseline
    regrowth) + "Raw ideas (2026-10-12 harvest)" (operator-API auth,
    lockout parity, TOTP/OAuth2, membership-remap/loginpage,
    passwordFile assertions, directory cap + /metrics count, nixpkgs
    auth-scheme warn + xml_int.curl proposal, bridge allowlist,
    -check-config, island REGISTER-fail hint). 06-55 residuals were
    CODE-VERIFIED first: `operatorWatchedUnits` lacks fs-directory
    (open), no fs-directory health leg (open), `directory.mode` options
    docs (done by T07/T14 — closed, not harvested).
11. **Item 27 (ANNOTATE).** Both prior reports resolved inline with the
    docs-health tooling (`emit-keys` → content-keyed spec → `--verify`
    → apply): 06-55 report 41 items (done struck with evidence; routed
    `**→ …**` markers; owner-gated rows stay BARE), 10-52 report 48
    items. Next harvest can no longer re-mint closed work.
12. **Stack CHANGELOG `[Unreleased]`** entries for every code-path change
    (check-vm, module-mirror check, the new suite asserts, wsprobe/
    fs-directory doc pins) per the changelog bar; pre-commit battery
    passed at commit (changelog-headings, gitleaks, lock-move-guard).
13. **Formatter-drift red on webphone main — caused, diagnosed, FIXED.**
    The daemon's nixpkgs bump updated prettier inside treefmt; 20 island
    files + shell.js failed CI's treefmt-check (runs `706e97d`,
    `6bb4272` red). Reflowed with `nix fmt` (content-inert — island
    node:test 0 fail; whitespace-drift flagged 4 non-`-w` lines, all
    proven re-indentation pairs), pushed `cd35b31` → **CI success** (all
    three jobs). Main is green; `ls-remote` verified `cd35b31` == HEAD.

## (b) PARTIALLY DONE

1. **Stack main is RED — root cause upstream, not this session's code.**
   Five runs failed (`d4e7bb7`…`6643bcf`, all carrying my harvest
   commits): `nix flake check` dies with **`hash mismatch in
   fixed-output derivation 'pbx-lens-088af23…-go-modules.drv'`** — the
   sibling lens lane's 23-file daemon commit `63b8563` moved pbx-lens
   dependencies without the re-pin, and pbx-lens is a FLAKE INPUT
   (`github:LarsArtmann/pbx-lens`, pinned `149e9f2`): the wrong hash
   lives in THAT repo's package recipe, upstream of the stack. NOT
   fixable properly from the stack repo; the lane's own notify-failure
   job fired (it posts to ALERT_WEBHOOK). aarch64 leg: success. My
   content is locally green (both VM suites via check-vm, telephony-eval
   incl. negative arm, `nix fmt` clean, pre-commit battery at commit).
2. **Stack has three unpushed daemon commits** (`7d6ec30` mixes my
   pbx.nix/eval.nix with the sibling's `forward-auth.nix`; `bc31a7b`,
   `e17e7dc` more sibling lane work). 60-min manual-push bar anchor:
   first unpushed commit 13:58:57 → bar met 14:58. I deliberately did
   not push mid-report: the tree also carries the sibling's UNCOMMITTED
   `FEATURES.md`, and pushing sweeps their half-done state into CI.
3. **The Nix 2.34 error-semantics discoveries** (attr-missing escapes
   tryEval; deep forcing overflows pure eval) live as comments in
   eval.nix but have no standalone lessons entry yet (candidate: stack
   `docs/lessons/` nix lesson).
4. **The 06-55→10-52 annotation verdicts** are inline but the two
   reports are NOT archived (correctly — owner-gated items remain bare;
   archive only when EVERY item resolves).

## (c) NOT STARTED

1. **T13** — pbx-artmann relock + staged prod cutover (owner-gated;
   pbx-artmann untouched, zero changes).
2. **T15** — prod verification suite (owner-gated).
3. **F49 in-VM rollback flip drill** (owner question g2).
4. **Baseline-regrowth audit** (+8604 lines; owner question g3).
5. **nixpkgs upstream items** (xml_int.curl default-set proposal;
   auth-scheme silent-ignore warn; owner call, item 14/37).
6. **Stack hardening tail items 17/18/20/23** — routed to TODO_LIST
   deliberately (shared test helpers, push-stall watchdog, suite-merge
   consideration): each needs VM-verified churn across many suites;
   wrong session to batch it.
7. **A dedicated stack ci-flake ledger FILE** — webphone has
   `docs/ci-flake-ledger.md`; the stack got ci-posture rows instead
   (same facts, different home; a file split would duplicate).

## (d) TOTALLY FUCKED UP (mine, honest)

1. **The module-mirror check took five build/eval iterations** (wrong
   relative import `../../` escaping the repo; hand-listed modules.nix
   args from a STALE FETCH — my raw.githubusercontent fetch silently
   returned content that did NOT match the pinned rev (14 args incl.
   `libopusenc` vs my 13); the overlaid perSystem pkgs stack-overflowed
   on `functionArgs` auto-fill; `deepSeq` overflowed pure eval by
   forcing derivation internals; then the tryEval-doesn't-catch-
   attr-missing discovery). Root cause, again: **verify-external-claims
   fired late** — one read of the store-path expression
   (`builtins.functionArgs`) up front replaces the whole fetch-and-pray
   step. The empirical method (nix-instantiate bisection) worked but
   was the expensive path.
2. **Two shell-quoting failures in the check-vm lambda**: `${n}` inside
   a Nix `''…''` string is NIX interpolation (undefined variable) —
   needs `''${n}`; then function-argument precedence needed parens. The
   repo's own AGENTS documents `''${}` escaping; I hit it anyway.
3. **I trusted LINE NUMBERS in a daemon-active file.** My 06-55 annotate
   map went stale when the daemon reflowed the report (+2 lines) between
   my grep and emit-keys — caught only because emit-keys refused the
   shifted lines and my content-snippet assert (0 hits) aborted. The
   save was content-keyed matching, not discipline. Line-number maps in
   this repo are DOA; content keys are the only sane spec format.
4. **`nix fmt` masked a broken flake**: my first fmt run "completed"
   while the flake failed to eval (the syntax error was fmt's own
   evaluation dying) — I read its output as success noise and moved on
   for one round. Gate output must be grep-verified (rc AND verdict),
   not skimmed.
5. **Edit bounces on sibling-formatter churn** (wsprobe.py,
   fs-directory.py reformatted minutes after I read them) — two wasted
   rounds. Handled per the rules (re-View + re-apply) but the lesson
   hardens: in this repo, re-View is mandatory even minutes later.
6. **A blind git-lock sleep** (~2 min on a held index.lock before
   checking `ps` for the holder). The second wait checked first. Small,
   but the pattern is "diagnose, then wait", not "wait, then diagnose".
7. **The 10-52 report's (b)5 / item-21 framing was wrong and I
   propagated it before busting it**: this session inherited "no CI run
   for 25cf08d" as fact; one `gh run list --commit` call falsified it.
   Cost: an investigation todo that was really a 30-second check.

## (e) WHAT WE SHOULD IMPROVE (larger > smaller)

1. **Self-describing eval beats fetched source**: `builtins.functionArgs`
   introspection + store-path imports (`pkgs.path`) are immune to stale
   raw-URL fetches and nixpkgs arg drift. Template for any future
   "mirror upstream structure" check.
2. **A Nix 2.34 semantics lesson for the stack** (`docs/lessons/`):
   attr-missing is uncatchable by tryEval (design consequence: fail the
   eval WITH the precise location instead of wrapping); deepSeq/all
   element-forcing escapes tryEval AND overflows pure eval on
   derivation-touching records; overlaid perSystem pkgs can recurse
   where pristine input pkgs don't. Encoded once in eval.nix comments;
   deserves the lesson home.
3. **Formatter-version bumps should be loud**: the daemon's nixpkgs bump
   silently changed prettier's canonical style → a 20-file reflow landed
   as a red→green CI cycle. A push-time gate diffing treefmt-check
   verdict across flake.lock bumps (or simply running the fast checks
   before the daemon's push) would catch it pre-push.
4. **whitespace-drift.sh mislabels formatter-bump reflows** as
   "hand-edited whitespace survived your formatter" — teach it the
   formatter-bump class (rc=1 with an all-indent/quote diff IS the
   reflow commit's expected shape) or document the exemption.
5. **check-vm could take `--jobs N`** (sequential today) for hosts with
   headroom, and the runbook/AGENTS should say when to prefer it over
   bare `nix flake check`.
6. **The stack's 1h-ceiling cancellations keep masking verdicts** (six
   runs burned the window today; the real red then surfaced only after
   completion). The recorded owner verdict predates this; splitting the
   workflow (fast checks / VM suites) would end the class.
7. **Content-keyed annotation is the only safe spec format** here — the
   emit-keys → snippet-spec → --verify flow proved out; if annotation
   passes recur, promote the session's key→verdict mapping helper into a
   small script instead of re-deriving the python each time.
8. **Attribute BEFORE acting on red mains in shared repos** — this
   session's five stack reds looked like they could be mine (my commits
   ride the same daemon sweeps) and one `gh run view --log-failed |
   grep hash` settled it as upstream. The check belongs BEFORE the
   emotional response, and next to the manual-push bar.

## (f) Up to 50 next items (impact-sorted; owner-gated marked)

**Stack main is red (upstream) — unblock first:**
1. ☐ **pbx-lens vendorHash**: the `149e9f2` input's go-modules hash is
   wrong INSIDE the pbx-lens repo (re-pin there: `nix build
   .#pbx-lens.goModules --rebuild` → `got:` → apply → bump the stack
   input) — LENS LANE's repo; owner routing call.
2. ☐ After the fix lands: stack push (incl. the three queued daemon
   commits) + CI-watch to green + verify `ls-remote`.
3. ☐ Re-run the messaging/other suites solo if the broken-pipe class
   recurs (check-vm makes this cheap now).
4. ☐ Watch webphone main (`cd35b31` green) for the next daemon drift;
   the reflow commit is the reference for the new prettier style.

**Owner-gated cutover cluster (T13/T15 — do not self-answer):**
5. ☐ **T13**: pbx-artmann relock to stack HEAD (clean tree; narHash
   covers all; release-runbook dance).
6. ☐ T13: stage prod with `telephony.directory.mode = "xml_curl"` +
   explicit `staticFallbackExtensions` (operator + ring groups; members
   that cannot webphone-login MUST be listed).
7. ☐ T13: verify the derivation secret file lands with correct
   owner/mode for BOTH webphone (DynamicUser) and the bridge
   LoadCredential.
8. ☐ T13: pre-cutover backup + the <2-min rollback drill rehearsed.
9. ☐ T15: prod derived-credential REGISTER (browser + one softphone).
10. ☐ T15: prod logout → next REGISTER denied (live revocation proof).
11. ☐ T15: prod static extensions unaffected during brokering.
12. ☐ T15: rollback drill executed ONCE on prod and reverted.
13. ☐ T13: pbx-artmann CHANGELOG entry written IN the cutover commit.
14. ☐ T13: post-cutover journal watch (`auth logout`, `directory query`
    lines prove the seam live).

**Owner questions (block items above / shape work):**
15. ☐ Cutover timing + staging shape (g1).
16. ☐ F49 scope: eval-golden enough or in-VM flip drill (g2).
17. ☐ Baseline-regrowth ruling (g3).
18. ☐ nixpkgs upstream items: xml_int.curl default-set + auth-scheme
    silent-ignore warn — pursue or keep stack-local (14/37).

**Stack hardening tail (TODO_LIST row):**
19. ☐ Port the answer-count journal + seam-assert pattern into the other
    webphone suites (shared helper in tests/common.nix).
20. ☐ Shared `secretFile` fixture helper (`user = <service>`) in
    tests/common.nix.
21. ☐ Wire `fs-directory.service` into `operatorWatchedUnits` under
    xml_curl (verified open: pbx.nix `operatorWatchedUnits` lacks it).
22. ☐ Consider fs-directory health joining the telephony-health view.
23. ☐ Pin the sipcred wire vector into the stack fixture (behavior is
    proven; the cross-repo HMAC vector is not asserted there).
24. ☐ xml_curl × fail2ban interplay check (digest-failure bans still
    apply?).
25. ☐ fs-directory journal noise under REGISTER storms (proxy-side rate
    limit?).
26. ☐ Merge dynamic-directory suite INTO telephony-webphone (halve CI
    cost; keep assertions readable) — pairs with item 6 in (e).
27. ☐ CI workflow split: fast checks vs VM suites (ends the 1h-ceiling
    masking class).
28. ☐ check-vm `--jobs N` flag + runbook guidance.
29. ☐ Nix 2.34 semantics lesson (stack docs/lessons/) — tryEval,
    deepSeq, pristine-vs-overlaid pkgs.
30. ☐ whitespace-drift.sh: formatter-bump class exemption.
31. ☐ Push-time formatter-drift gate (catch prettier bumps pre-push).

**Webphone/infra tail:**
32. ☐ Daemon push-stall watchdog (alert when origin/main lags HEAD
    >60 min — fired twice today).
33. ☐ Daemon-enforcement pre-commit hook (block go.mod/go.sum/flake.lock
    in daemon commits).
34. ☐ Recreate the curl-level passkey+brokering smoke under `scripts/`
    ONLY on demand (source lost with /tmp; VM suite + go tests carry the
    arms).
35. ☐ Smoke-harness guard rule: apply when a credential-equality leg
    lands (verified clean today; no code to write yet).

**Routed product/deferred (ROADMAP fuel — refine before touching):**
36. ☐ Operator-API session auth (kills `extension_password_files`).
37. ☐ Extension-login lockout parity.
38. ☐ TOTP/OAuth2 login modes.
39. ☐ Membership-remap ruling + loginpage ruling (written decisions).
40. ☐ `passwordFile` permission assertions in the webphone module.
41. ☐ `/internal/directory` listing cap.
42. ☐ Derived-credential count in /metrics.
43. ☐ Bridge extension allowlist (threat-model pass first).
44. ☐ `webphone -check-config` preflight flag.
45. ☐ Island registration-failure hint when `sipCredential` is on.
46. ☐ wsprobe: `--assert` legs fully parameterized (user still hardcoded
    in the unauthenticated probes; only `--auth-register` takes it).
47. ☐ Stack: dedicated ci-flake ledger file vs ci-posture rows (one home
    ruling).
48. ☐ Annotate-then-archive the two 2026-10-12 reports once the owner
    cluster (items 5-18) resolves.
49. ☐ Quarterly art-dupl baseline audit (due 2027-01-11; standing).
50. ☐ erraudit tier re-measure (due 2026-11-05; standing).

## (g) 3 questions for the owner

1. **The pbx-lens red (item 1):** the broken go-modules hash lives in
   the `pbx-lens` repo (the lens lane's 10-11 extraction), upstream of
   the stack — do you want the lens lane (or a fresh session) to re-pin
   THERE and bump the stack input, or should I attempt the cross-repo
   re-pin myself (two repos, one red-main fix, racing an active lane)?
2. **check-vm adoption (items 28/27):** is sequential check-vm the
   ratified LOCAL default (replacing bare `nix flake check` in AGENTS),
   and do you want the CI workflow split (fast checks vs VM suites) as
   the fix for the 1h-ceiling masking class?
3. **Nix 2.34 semantics (item 29):** file the attr-missing-escapes-
   tryEval + pure-eval-overflow discoveries as a stack lesson (and
   optionally a nixpkgs/nix docs note), or keep them as eval.nix
   comments only?

## Session inventory (what changed where, this session)

- **webphone** (pushed, CI GREEN at `cd35b31`; `ls-remote` verified):
  README stub, error-contract cross-link, AGENTS (+tri-repo FS line,
  +`--resolve` smoke line), TODO_LIST harvest section + sweep line,
  ROADMAP (3 questions + Raw ideas 2026-10-12), both status reports
  annotated (41+48 items), and the 20-file island formatter reflow
  (`cd35b31`, content-inert, island tests 0 fail). Docs-only otherwise —
  zero Go/JS logic changes.
- **nix-international-telephony** (pushed to `6643bcf` + 3 daemon
  commits queued): `apps.check-vm`; `modules/telephony/freeswitch-modules.nix`
  (new, ONE home); pbx.nix consumes it; `tests/eval.nix` module-mirror
  check (both arms verified); dynamic-directory suite +config.js pin
  +auth-logout assert; webphone suite +static auth leg; wsprobe exit
  codes; fs-directory Cache-Control comment; ci-posture ledger;
  lessons/freeswitch.md trap section; CHANGELOG [Unreleased]. All my
  gates locally green (check-vm run of both suites, eval check, fmt,
  pre-commit battery). Main RED from upstream pbx-lens (see b1) — the
  sibling's lane, their notify fired, unpushed commits don't fix it.
- **pbx-artmann**: untouched (owner-gated T13).
- **Key gates at report time:** webphone CI success (fast checks + build
  + go test -race) at `cd35b31`; stack local gates green / CI red
  upstream; both remotes verified with `ls-remote`; the sibling lane
  active throughout (their files untouched: rendering_audit.py,
  test_markers_check.py, forward-auth.nix, FEATURES.md, their CHANGELOG
  rows).
