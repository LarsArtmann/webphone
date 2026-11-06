# Status: Admin visibility surface (read-only, all lines) — 2026-11-05

Claims verified as of 06:47 (push/CI statements rot fastest; the CI
verdict line names its run).

## The train in one paragraph

The owner asked: "As an Admin I want access to ALL messages, fax, phone
calls, users, phone numbers." Shipped as a read-only Admin tab for
accounts carrying `admin: true` on the shared `auth.users` mapping:
cross-line conversations (with a read-only transcript deep link), all
fax jobs (documents stream via admin-gated fallbacks), merged per-line
CDR pages + all transcribed calls, the account inventory with live
session counts, and the numbers inventory (identities + mapped lines).
Admin is a property of the identity, resolved PER REQUEST from the live
mapping — nothing is minted into the session, so a config demotion
bites on the admin's next request, and a deployment with no admin
configured is byte-identical to one without the feature (every admin
route answers the styled 404 for everyone). Owner-scoping stays the
rule everywhere else: the only cross-owner reads are the deliberate
`*Any`/`ListAll*` seams behind the admin gate, pinned by name.

## Verification state

- `go test -count=1 ./...` green at train end (~00:15) AND re-verified
  green 06:47 on the post-daemon tree (server/admin/views/config/app/
  store/readmodels all `ok`; daemon landed later commits: the art-dupl
  #14 sweep and a templ regen with the PINNED compiler repairing my
  older-compiler output — commit 97a3656d).
- Island tests: 286/286 green (no island changes; closure untouched).
- `scripts/format-check.sh` (the sandboxed treefmt authority): green.
- tw.css + login.css rebuilt after the app.css admin block
  (`scripts/build-tw-css.sh`, committed artifacts).
- Live boot verification: real binary + real config file — anonymous
  `/admin*` 401s, zero nav difference; a directly-minted admin session
  row (what any ceremony mint produces) renders the panel + nav link on
  `/admin` and `/partials/admin` (users/numbers sections confirmed).
- CI on origin/main: the last full `CI` run (2026-11-05T02:25:38Z) is
  SUCCESS — after every commit in this train. Three earlier mid-train
  failures (00:18–01:48Z) were in-flight daemon pushes; the daemon's
  own repair commits ("baseline the admin train's gate prologues and
  templ shapes", 49aa34bd) settled them.

## What is NOT verified (addendum 18:15: both former gates CLOSED)

- ~~A completed `buildflow` run for this train~~ done at `3dab5c6c` — run `20261105-055334-271d2092` (started 06:53, full mode,
  `BUILDFLOW_NO_RESULT_CACHE=1`, tree at 3dab5c6c): **68/68 steps green,
  0 fail, 100%**. Verdict source: `buildflow history` (the streamed log
  tail was lost to output buffering; the state DB is the authority).
  Later same-day runs against other sessions' in-flight trees are not
  this train's verdict.
- No real browser ceremony (passkey/SSO) → admin tab flow was run; the
  gate reads the session row, so the mint path is covered by
  construction, but the honest label is "ceremony-untested".
- The consuming stack's browser E2E ~~has not been re-run~~ **CLOSED
  2026-11-05 02:06, evidence-verified**: the stack pinned webphone
  9cb5c64d (admin train included; unattributed daemon commit abbe2de at
  02:00:57) and the E2E output
  `/nix/store/c5a2b0pmnj4z03npfivp6pm212kswszr-vm-test-run-telephony-browser`
  was registered 02:06:23; the log shows the full suite (registration,
  WebRTC call, DTMF, reconnect-recovery watchdog, fs_cli drain)
  finishing in 245.08s with no failure. No-admin config = the E2E's
  default, so this also proves the byte-parity claim end-to-end.

## Gaps deliberately left (scope rulings this train)

- Write-shaped admin actions stay REJECTED (acting happens on each
  line's own login) — the D4 delegation item in TODO_LIST owns that.
- Admin voicemail visibility (per-line voicemail boxes via each line's
  credentials) was NOT built: the ask named messages/fax/calls/users/
  numbers, and voicemail is its own surface. Candidate follow-up.
- The Admin tab is static (no SSE live updates) — cross-owner SSE
  fragments would need admin-scoped rendering; documented as future.
- The styled-404 shell does not render the Admin nav link for an admin
  (the 404 page builds its own ShellProps); cosmetic, self-recovering
  on the next navigation.

## Self-review (what this session did badly)

1. Ran `templ generate` with the dev-shell CLI v0.3.1020 while go.mod
   pins v0.3.1070 — the warning was printed and ignored; the daemon had
   to repair the generated shapes (97a3656d). Lesson: the pinned-
   compiler doctrine applies to templ too, not just art-dupl baselines.
2. Wrote Go `continue` inside a templ loop — templ renders it as literal
   text ("continue" shipped into served bytes in a test run). Caught by
   my own nav-conditional test; fixed as an if-guard. templ's control
   flow is a SUBSET of Go's.
3. Referenced `icons.Message` from a grep that included my own
   generated file (self-confirming evidence). Should have listed the
   icons package's exports first.
4. Left a broken placeholder (`base64.StringsDecode??`) in a test file
   mid-edit — the auto-commit daemon could have committed it broken.
   Write tests in one pass in daemon-active repos.
5. Three compile-fix cycles (`sess` unused, if-scoped variables, a
   nonexistent `Seq` field) that a slower check-before-write pass would
   have caught.

## Key seams (for the next session)

- `internal/admin` — the service (config-free; IsAdmin/Users/Numbers/
  Extensions/SIPPassword, nil-safe).
- `internal/app` `AdminService` — the config→runtime adapter (shared
  map + identities; broken mapping without an admin degrades to nil,
  with an admin it fails boot loudly).
- `internal/server/admin.go` — gate (`requireAdmin`, `isAdminSession`),
  routes (`/admin`, `/partials/admin`, `/admin/threads/{id}`,
  `/partials/admin/threads/{id}`), panel builders, per-line CDR merge.
- Cross-owner seams: `store.ListAllThreads/GetThreadAny/
  AttachmentByIDAny`, `messaging.AllThreads/ThreadAny/AttachmentByIDAny`,
  `readmodels.Faxes.ListAll/GetAny`, `Transcripts.RecentAll`.
- Pins: `server/admin_test.go`, `store/admin_reads_test.go`,
  `readmodels/admin_reads_test.go`, `config/admin_test.go`,
  `app/admin_test.go`, `admin/admin_test.go`.
- Docs: dom-contract.md § "Conditional ids (admin surface)", AGENTS.md
  invariant bullet, README config row, FEATURES row, CHANGELOG entry,
  TODO_LIST D4 update.
