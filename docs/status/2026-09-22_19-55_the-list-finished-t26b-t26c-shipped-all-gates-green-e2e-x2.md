# Status 2026-09-22 ~19:55 — the list is finished: T26b/T26c shipped, every gate green, E2E ×2 green

The finishing session of the SUPERB 20-year durability plan. Every
actionable row of the TODO list that does not require the owner's
terminal or an owner decision is DONE and VERIFIED. This report closes
the train.

## a) Shipped this session (webphone `main`)

- **T26b TURN REST credentials** (`c1971c4`): `turn_rest.secret` +
  `turn_rest.ttl` (48h default). With the secret set, `/config.js`
  derives a coturn REST pair per response (username = unix expiry,
  credential = base64(HMAC-SHA1(secret, username))) for every
  `turn:`/`turns:` entry — one shared expiry, STUN entries stay
  credential-free, the unset-secret path ships static creds verbatim.
  Boot validation rejects a secret without a TURN URL (dead config)
  and non-positive TTLs. Tests verify the actual HMAC, not a grep
  (`configjs_test.go`), plus four Load/validation cases.
- **T26c per-extension data export** (`37aae5a`): `GET /api/export`
  (session-gated, Settings-tab download link): zip with messages.json
  (every thread + messages + attachment manifest), faxes.json,
  contacts.vcf. Owner scoping rides the extension-scoped store reads;
  blobs stay in the blob store (manifest only). 401 pin for anonymous.
- **ThemeScript verification gaps closed** (`9cb1c37`): island-lint now
  fail-covers `theme-preload.js`; a node:test spec pins the preload
  (light/dark apply; auto/absent/denied storage do nothing) via
  `vm.runInThisContext` (import() cannot re-run an IIFE — node's
  module cache ignores query-busting, verified empirically); smoke
  checks 4a assert the asset serves + zero inline `<script>` on the
  shell.
- **Honest-tool noise silenced** (`708ef0d`): .codespellrc documents the
  German UI words + the deliberate historical-typo quotes; shellcheck's
  `cd || exit` guards in two scripts; AGENTS trimmed to exactly 377.
- **Docs**: README (TURN + export rows, dead CRM link dropped — the
  repo is private, lychee's only 404), CHANGELOG, FEATURES, TODO
  harvest (`46c0064`), plan execution log, erraudit tier-2 early
  re-measure into AGENTS (127 total / 113 outside the crm seam; grew
  from 102 via the CRM train + turn_rest idioms; top unconverted seams
  recorded: config.go 22, store/messages.go 20, pbx/client.go 8).

## b) Shipped this session (stack `main`, pushed)

- `7e4658f` — the logged-out-dial-guard scenario was UNREACHABLE: a
  real logout RELOADS to the login-card shell (session.js by design),
  which renders no data-dial buttons at all (run 4 died there with
  NoSuchElementException). The scenario now seeds its own contact,
  reproduces the in-place signed-out island DOM (`#phone-view` hidden +
  `#login-view` shown — the session-expiry-mid-tab hazard the guard
  actually exists for), pins the toast + `#ext` focus, AND pins the
  safe-by-construction reload (zero data-dial buttons survive logout).
- `a5be17c` — run 5 died at the contacts roundtrip on
  ElementClickInterceptedException: a transient transfer-verdict toast
  overlaid the nav at the click point (run 4 passed on timing luck).
  `click_tab` now retries until the click lands.

## c) Gate scoreboard (all on the final tree, in order)

| Gate | Verdict |
|---|---|
| `go test -count=1 ./...` | all packages ok (run twice) |
| erraudit tier-1 (`--type-aware --disable-extensions`) | 0 violations |
| erraudit tier-2 (`--enforce-go-error-family`) | 127 total / 113 outside crm seam — re-measured + recorded (see a) |
| buildflow full (`BUILDFLOW_NO_RESULT_CACHE=1` via wrapper) | RC 0 (warnings only; the earlier "69" was a host-go invocation artifact — the pipeline-exit-code lesson AGAIN) |
| `nix run .#vulnix` (release gate + triage) | zero real advisories (all distro-patched in locked nixpkgs); the 18:15 NVD-404 was transient |
| `nix flake check` | all 18 checks — twice (before + after the island-lint scope edit; includes KVM backup VM + island-js with the new spec) |
| smoke | 40 passed + restart scenario 4/4 |
| island node tests | 56/56 (54 + 2 new) |
| templ regen + oxfmt/prettier/shfmt verify | clean |
| lychee | 0 errors (87 links) |
| aarch64 cross-build | ELF bytes verified: `7f 45 4c 46` + `02 00 b7 00` (EM_AARCH64) |
| stack browser E2E | **GREEN ×2**: 198.11s + forced rerun 236.93s (budget 445s; distinct runs proven by differing marker timings + cleanup timestamps) |
| codespell / shellcheck / ruff | all clean after the fixes |

## d) Mistakes this session (kept as lessons)

- Misread `buildflow` exit 69 as a tree failure: it was my own
  invocation OUTSIDE nix develop (host go 1.26.7) + reading RC after a
  masking pipe. The wrapper run was RC 0 all along.
- Started E2E "run 7" that was actually a cache HIT (empty build
  output = nothing ran); detected via identical run-6 timings, forced
  a real rerun with `--rebuild`.
- First theme-preload spec used import() query-busting — node's module
  cache ignores it (second case silently reused the first module);
  rewrote on `vm.runInThisContext`.
- Two table-row edits in TODO_LIST failed on exact-match anchors;
  finished with line surgery (the standing lesson).
- Drafted an export handler with an unused-import placeholder guard —
  caught by go vet before commit, removed.

## e) Push state (ls-remote, 19:52)

- stack `a5be17c` = origin ✓ (both E2E fixes public)
- pbx-artmann `24cb90d` = origin ✓
- webphone: origin at `0a34611`, local `46c0064` — the daemon is
  mid-catchup (it has pushed everything up to 0a34611); re-verify with
  `git ls-remote` before any tri-repo step.

## f) What remains (all owner-gated or owner-terminal — by design)

1. v2.6.0 fold decision + release train (TODO row 1): [Unreleased]
   holds three coherent themes; every precondition I could satisfy is
   green (E2E ×2, aarch64, gates). Then stack re-pin + pbx-artmann
   relock #4 per docs/release-runbook.md.
2. Deploy + post-deploy smoke (owner terminal; command in TODO row 2).
3. T26b STACK half: coturn `static-auth-secret` sharing webphone's
   `turn_rest.secret` (parked row).
4. Owner-calls batch session (~15 decisions, briefing doc ready).
5. Release announcements (drafts ready; owner picks channels).

## g) Owner questions (carried from 18:46, still open)

1. **v2.6.0 fold timing** — fold the three [Unreleased] themes and cut
   now (everything is green), or hold for more? The plan is complete;
   every day unreleased is a day prod serves v2.4.0.
2. **Deploy cadence** — deploy the v2.5.0 chain that pbx-artmann
   already locked, or wait for v2.6.0 and deploy once?
3. (Resolved without asking: ThemeScript adoption ownership — the CRM
   session completed it cleanly and the standing watch was taken: CSP
   is now `script-src 'self'` with zero inline scripts, hash-pin
   machinery deleted, verified by the modernized strict-CSP test.)

The list is finished. Everything above is verified from actual runs,
not intent.
