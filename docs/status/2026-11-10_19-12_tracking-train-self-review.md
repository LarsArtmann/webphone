# Prod UI-tracking continuation — self-review: what was forgotten, what was weak, what is next

> Claims verified as of 19:12. This session executed the Pareto plan
> (docs/planning/2026-11-10_15-05_SUPERB-prod-ui-tracking-pareto-plan.md)
> end-to-end minus the owner-gated tasks, then turned the plan's own
> questions on itself. Companion evidence: the plan's Ledger section (17
> rows), the continuation report
> (docs/status/2026-11-10_18-40_prod-ui-tracking-continuation.md), and the
> struck §f rows in the 14:42 report. Main tip `0f067c7e`; every push ran
> green CI (`fix(harness)` 42869749284 success, the §f strikes 42869974985
> success, the daemon bulk 42868918061 success).

## a) FULLY DONE (verified, not claimed)

- **T01–T03 — the red-gate triad.** The island-js red was the other
  session's in-flight spec, finished by their own commits before I
  touched anything: 328/328 specs at HEAD, `island-js` sandbox check
  green, CI success. `nix-hash-fix` decoded (its lock-drift warning
  fires on a GREEN build with a factually-backwards premise: go.mod
  64c1c637 11-08 is OLDER than the vendorHash pin 5c00b153 11-10);
  mechanism + verify recipe recorded in `.buildflow.yml`. Full
  buildflow EXIT 0 at train end.
- **T04 — the TLS-front smoke.** caddy `tls internal` on 18443 →
  binary on 18097; dry-run + full tracked runs through the remote
  lane; Secure-cookie `add_cookie` proven by real session surfaces.
  Two real gotchas are now harness features (the `acceptInsecureCerts`
  CAPABILITY vs the ignored switch; the `socket.getaddrinfo` patch so
  the python HTTP lane shares `--host-map`).
- **T09–T14 — the instrument v2.** `--diff-threshold`/`--badge-bar`
  (defaults preserved), `--diff-clip` verified on REAL drifted pairs
  (settings-light 10.372% / media-transcribe-dark 7.849% clipped where
  pairs previously skipped), `--self-test` (29 checks, green in BOTH
  envs), `--dry-run` (verified over TLS), the ring guard (negative-
  tested: 0 matches on an idle boot), per-shot URLs + deployment
  commit in summary.json, wire-facts.json, byte-size delta columns +
  gallery cross-links, threads-shortfall notes.
- **T15–T18 — conditional surfaces.** `/enroll` probed (not guessed —
  the product reaches it via distributed token links, so a login-page
  href probe would have been dead code), dense `history-transcript`
  q= word scraped from the first live transcript, CRM-strip detection
  verified LIVE against a real fake-Kith boot, admin detection
  implemented + self-tested — then correctly identified as
  product-blocked (`isAdminSession` needs `LoginEmail`; the
  extension-password lane never sets it).
- **T19/T20 — retention + one command.** `--prune N` verified live
  (removed exactly the two empty crashed-attempt dirs); `nix run
  .#ui-track` green with all flags.
- **T21/T22/T29/T30 — docs truth.** .gitignore truth-up, mypy.ini PIL
  mirror, AGENTS "24 shots" → deployment-following matrix + instrument
  honesty, the 2026-10-29 plan's stale row struck with evidence, four
  lessons entries, the harness docstring carries the TLS recipe +
  fixture recipe + the 429-stays-manual ruling, CHANGELOG
  `[Unreleased]` entry, ten satisfied §f rows struck in the 14:42
  report (docs-gates green after).
- **T23/T24 — rulings recorded** in the plan ledger: a11y emulation
  DEFERRED (cheap CDP, but matrix doubling before the ugliness
  inventory exists); lighthouse NO-GO recommendation (npm/chrome-
  launcher brittleness, nix name collision, signal already covered by
  shots + DOM pins + layout proof).

## b) PARTIALLY DONE

- **T09/T10 calibration VALUES** — the flags landed and the clip path
  is proven on real (local) data, but the owner-facing calibration on
  PROD data does not exist: it needs T05's authenticated runs. The one
  honest number I have: identical-seed local runs 3 minutes apart peak
  at 2.5% changed (relative timestamps), so the 0.5% badge bar is
  optimistic — recorded, not solved.
- **T13 ring guard** — negative-tested only. The positive path (a real
  inbound call mid-run triggering the re-shoot) has never executed; it
  needs a live deployment and a caller. The code is 8 lines and reads
  correctly, but "reads correctly" is not "verified".
- **T15 admin surface** — detection landed and self-tested, but the
  surface can never fire in a credential-lane capture (the product
  wall). A real admin shot needs an identity-carrying capture lane
  (passkey ceremony cannot be automated) — this is a dead end by
  design until that lane exists.
- **T25 night calibration** — mechanism understood, not executed: a
  quiet-window run pair on real data requires the laptop on overnight;
  recorded in §f, not done.
- **T26 `--open`** — implemented (4-line guard, help-text verified)
  but never exercised live: the flag was not passed in any actual run.

## c) NOT STARTED (all owner-gated, all staged)

- **T05** — the authenticated prod capture (OG-1 creds). The single
  command is ready: `nix run .#ui-track -- --base-url
  https://pbx.artmann.tech --track`.
- **T06–T08** — first UI fix with before/after proof, the first real
  prod diff chain, the mobile prod pass. All depend on T05.
- **T19's OG-3 half** — the retention ruling (laptop-only vs synced
  store). `--prune` exists; the WHERE is the owner's.
- **OG-2's build half** — lighthouse scores (my recommendation is NO-GO;
  the ruling is the owner's).

## d) TOTALLY FUCKED UP (honest list)

- **The fake-CRM dead-server mistake.** My first T16 verification
  started the fake CRM inside a short-lived python snippet — it died
  with the snippet, the app booted against a dead port, and I burned
  two boots + a debugging round on a harness-setup bug of my own
  making. The lesson (harness-adjacent servers must outlive the
  process that configured them) is only in the plan ledger, not
  lessons.md — it should be.
- **The edit-tool staleness dance with the daemon.** Three multiedit
  calls were refused or landed stale because the auto-commit daemon
  touched files between my read and write; one plan-ledger edit was
  silently refused and I nearly didn't notice (the tail looked
  plausible until I grepped for the row count). I re-verified AFTER,
  which caught it — but the workflow should be read-immediately-before-
  every-edit on shared files, not after the first refusal.
- **The options-block edit that dropped its own payload.** My capture()
  chrome-options rewrite landed without the insecure-capability and
  host-map wiring it was supposed to add — the TLS run then failed on
  DNS, and I "fixed" it twice (resolver patch in python, capability in
  a debug snippet) before noticing the host-map chrome args were
  missing from the file entirely. Cost: two wasted TLS runs. The
  multiedit new_string simply omitted the planned lines — reviewing
  the diff against the INTENT, not the file, would have caught it.
- **The `sed`-line-number trap in the nix rebuild.** `nix build ... |
  tail -4; echo EXIT=$?` captured tail's exit code, not nix's —
  PIPESTATUS is unavailable in this shell and I reached for the
  pattern anyway (AGENTS warns about exactly this). The FOD verdict I
  then relied on came from buildflow's own output instead, so no harm
  landed, but the verification was invalid as written.

## e) WHAT WE SHOULD IMPROVE

1. **Instrument-first for verification claims**: three of today's four
   mistakes were verification loops where the instrument (a probe, a
   config, a pipeline exit code) was weaker than the claim. The repo
   rule "never trust a local-vs-CI diverging verdict before ruling the
   leak out" generalizes: never trust a verification whose plumbing you
   did not re-read this session.
2. **Daemon-race discipline**: stage + commit in ONE tool call worked,
   but edit-vs-daemon races still cost three refusals. Consider doing
   large multi-edit batches inside a single `nix develop -c` python
   rewrite (atomic write) when the target file is daemon-hot.
3. **`--open` and the positive ring path deserve real tests** in the
   first authenticated prod session — both are the only landed-
   but-never-executed code paths in the harness.
4. **The plan ledger is now the real status home** (17 evidence rows)
   — future sessions should sweep THAT table, not the reports, when
   resuming this work; consider harvesting it into TODO_LIST rows.
5. **mypy/ruff zeroing should happen at landing time, not at gate
   time**: I wrote ~450 new harness lines, then spent a dedicated
   cleanup pass zeroing 8 mypy + 2 ruff findings. Running the two
   steps scoped to the changed file (`buildflow -s ruff-check`,
   `-s mypy-check`) after each chunk would have folded that cost into
   the work.

## f) NEXT (ranked, roughly Pareto-ordered)

1. **[OG-1] T05** — the authenticated prod capture; the command is
   staged, credentials are the only missing input.
2. **T05 act 2** — the owner gallery walkthrough → the ranked top-5
   ugliness inventory into docs/planning/.
3. **T06** — the first UI fix from the inventory, proven by a
   before/after tracked run (the whole point of the train).
4. **T07** — the second authenticated run → the first REAL prod
   compare chain; judge badges/byte-deltas/clipped marks in anger.
5. **T09/T10 values** — calibrate threshold + badge bar on real prod
   data (0.5% is optimistic; expect 2–3%).
6. **T08** — the mobile prod pass (`--viewport 390x844`) + the default-
   or-on-demand ruling.
7. **Live-verify the two dormant paths**: `--open` and the ring guard's
   positive path, in T05's session.
8. **[OG-3]** — the shot-storage/retention ruling; `--prune` awaits it.
9. **T25** — night-run stability pair on the next overnight session.
10. **Harvest the plan ledger** into TODO_LIST rows so the next session
    starts from the dispatch queue, not the plan doc.
11. **lessons.md**: add the fake-CRM dead-server lesson (it lives only
    in the plan ledger today).
12. **Upstream finding** — buildflow's nix-hash-fix lock-drift warning
    fires on a green build with a backwards premise; an issue on
    LarsArtmann/buildflow when the owner wants it.
13. **T15's honest closure** — decide whether admin tracking waits for
    an identity-carrying capture lane or dies (it cannot fire from the
    credential lane; that is a product fact, not a bug).
14. **T16 on real prod** — the CRM-strip surface joins only when prod's
    real Kith publishes context; verify in T05's run.
15. **T28's second half** — partial-run tolerance is self-test-pinned;
    pin the EMPTY-run discard guard with a self-test case too (it has
    a live observation but no automated pin).
16. **`--emulate-feature`** — the one-line opt-in a11y axis if T06's
    fixes surface motion/contrast uglies.
17. **T20 polish** — `ui-track` passes no `--out-dir`; prod runs land
    in the repo's `ui-shots/` by default — confirm the owner is happy
    with that default once real-data runs start.
18. **Noise-floor documentation** — after T07, write the observed
    per-surface noise table into the harness docstring so the badge
    bar is a measured number, not a guess.
19. **CHANGELOG on release** — the `[Unreleased]` block should fold
    into the next version cut (2.13.0 shipped today; this lands in
    2.13.1/2.14.0 whenever the owner cuts).
20. **T30's 429 ruling** — verified written into the docstring; keep it
    there when the AGENTS command block is next edited (it is not in
    AGENTS itself).

## g) QUESTIONS (cannot answer myself)

1. **May I file the buildflow nix-hash-fix issue upstream** (LarsArtmann/
   buildflow) with today's evidence — commit-order proof (go.mod older
   than the vendorHash pin) plus the step's own "build green but N
   lock-drift warnings" diagnose lines — or do you want to carry it
   yourself since it is your tool?
2. **Which lane should admin capture ride, if any**: accept that
   credential-lane captures can never see the Admin tab (current
   state), or is a semi-automated identity lane (session cookie exported
   from YOUR passkey login into an env var, used by the harness only,
   never logged) acceptable to shoot the admin surface on prod?
3. **Default retention number for `--prune`**: when you rule OG-3's
   storage home, what N should the one-command recipe default to (runs
   carry real caller data on disk; my default is NO pruning unless you
   pass --prune, on the theory that silent deletion of evidence is
   worse than a full disk)?
