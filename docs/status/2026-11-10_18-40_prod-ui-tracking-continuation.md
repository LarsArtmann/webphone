# Prod UI-tracking continuation — the loop is trusted, the gates are green

> Claims verified as of 18:40. Continuation of the 14:42 screenshot-tracking
> train, executing the Pareto plan (docs/planning/2026-11-10_15-05_SUPERB-
> prod-ui-tracking-pareto-plan.md). The plan's Ledger section carries the
> per-task evidence rows; this report is the snapshot.

## Done (T01–T04, T09–T30 minus owner-gated)

- **Gates green (T01–T03)**: the island-js red dissolved — it was the
  other session's in-flight spec, finished by their own follow-up
  commits (328/328 at HEAD, `island-js` sandbox green, CI success on
  run 42832784499). `nix-hash-fix` decoded: its lock-drift warning
  fires on a GREEN build with a factually-backwards premise (go.mod
  11-08 is OLDER than the vendorHash pin 11-10); mechanism + verify
  recipe recorded in `.buildflow.yml`. Upstream finding = the owner's
  own tool, noted not filed.
- **TLS-front smoke (T04)**: caddy `tls internal` (18443) → binary
  (18097); the remote lane ran `--dry-run` plus three full tracked
  runs through it. `add_cookie` under the Secure mint proven by real
  session surfaces. Two gotchas are now harness features: chromedriver
  honors the `acceptInsecureCerts` CAPABILITY only (the switch form is
  silently ignored), and the python HTTP lane shares `--host-map` via a
  `socket.getaddrinfo` patch.
- **Instrument v2 (T09–T14, T19–T20, T26–T28)**: `--diff-threshold` /
  `--badge-bar` (defaults preserved), `--diff-clip` (verified on real
  drifted pairs — 10.372%/7.849% clipped where pairs used to skip),
  `--self-test` (29 checks, green in both envs), `--dry-run` (verified
  over TLS), the ring guard (negative-tested on an idle boot),
  per-shot URLs + deployment commit in summary.json, wire-facts.json,
  byte-size deltas + gallery cross-links in compare.html, `--prune N`
  (verified live: removed exactly the two empty crashed-attempt dirs),
  partial-run compare tolerance (self-test-pinned), `nix run
  .#ui-track` flake app. Discovered live and fixed: a crashed capture
  left EMPTY run dirs that hijacked the next compare's "previous"
  (`discard_empty_run`).
- **Conditional surfaces (T15–T18)**: `/enroll` probed anonymously and
  joined on the fixture boot; the dense `history-transcript` q= word is
  scraped from the first live transcript; CRM-strip detection verified
  LIVE against a real fake-Kith boot (`crm-context` joined the matrix);
  admin detection implemented + self-tested, then hit a PRODUCT WALL:
  `isAdminSession` needs `sess.LoginEmail`, which the extension-password
  capture lane never has — a credential-lane session can never be
  admin. Documented, detection stays dormant-but-correct.
- **Docs (T21, T22, T29, T30)**: .gitignore truth-up (track history
  persists; `--prune` is the only retention), mypy.ini PIL mirror,
  AGENTS "24 shots" → deployment-following matrix + the instrument
  honesty one-liner, the 2026-10-29 plan's stale "harness not re-run"
  row struck with evidence, lessons.md § tracking-train process lessons
  (4 entries: re-run foreign reds at HEAD, the daemon narrative race,
  capture browsers are real SIP endpoints, the mixed-tree push ritual),
  the harness docstring carries the TLS recipe, the bare-boot fixture
  recipe and the 429-stays-manual ruling. CHANGELOG `[Unreleased]` line
  added.

## Rulings recorded (T23, T24)

- **a11y emulation axes: DEFERRED.** Technically trivial (same CDP
  command the runs already use), but each axis doubles the matrix and
  the ugliness inventory (T05) is still owner-pending. Opt-in design
  noted in the plan ledger.
- **lighthouse-style scores: NO-GO recommendation** (evaluation only
  per the task's scope): npm/chrome-launcher brittleness, nix name
  collision, and the layout signal already covered better by full-page
  shots + DOM pins + the layout proof. OG-2 stays with the owner.

## Blocked (owner-gated)

- **T05–T08 (OG-1)**: the authenticated prod capture needs
  `WEBPHONE_CAPTURE_EXTENSION` / `WEBPHONE_CAPTURE_PASSWORD` — asked
  twice, still unanswered. Everything the credentials unlock (the
  ranked ugliness inventory, the first UI fix, the mobile pass, the
  first real prod diff) is staged behind them.
- **OG-3 (shot storage/retention)**: `--prune` exists; the WHERE ruling
  is the owner's.

## Noise facts for calibration (T07/T09/T25 input)

Two local runs 3 minutes apart over identical seeds: 22/26 pairs above
zero, MAX 2.5% (thread surfaces — relative timestamps are the noise
source). TLS runs 90s apart: 28/28 above the 0.5% bar, same cause. The
0.5% badge bar is optimistic on this product; expect 2–3% after real
calibration.

## Next

1. Owner answers (creds; fix-foreign-spec permission is moot — T01
   resolved itself; shot storage).
2. T05 authenticated prod run → ranked inventory → T06 first UI fix.
3. T07 second prod run → first real prod diff → T09/T10 calibration
   values on REAL data.

## Harvest

- [ ] §f — T05 credentials (OG-1) remain the single blocker for the
      authenticated prod matrix; ask again with the one-liner ready.
- [ ] §f — T25 night-run calibration needs a quiet-window run pair;
      schedule it the next time the laptop is on overnight.
