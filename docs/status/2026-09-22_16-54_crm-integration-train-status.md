# CRM integration train — status (2026-09-22 16:54 CEST)

Session scope: design + full implementation of the **optional webphone ↔
Ledger CRM integration** across BOTH repos (`~/projects/webphone` and
`~/projects/crm`): caller-name enrichment (CRM → phone) and call-activity
journaling (phone → CRM), config-gated OFF by default on both sides.

Concurrent session observed: a composer-drafts train (T21d) ran in the
same webphone checkout. Its `TestFormatClockAndStampFollowLanguage`
(view helpers) fails on this host — TZ-sensitive (asserts UTC wall times,
host is UTC+2). Not introduced or touched by this train.

---

## a) FULLY done

### Design (settled before code)

- Contract: number ↔ contact identity. Webphone resolves numbers to CRM
  names (read) and journals finished calls (append). Matching lives in
  ONE place — the CRM: digits-only normalize, suffix ≥8 digits,
  prefix-delta ≤4, trunk-0 variants. Webphone sends numbers verbatim.
- Opt-in both sides: webphone `crm.url` + `crm.token` (validated
  both-or-neither, absolute http(s) URL, env `WEBPHONE_CRM__URL/_TOKEN`);
  CRM `-api-token` (surface unmounted without it).
- Never-mint rule: unknown callers never become CRM contacts; the
  call-log endpoint answers 204 for them.
- Enrichment fails open: dead CRM ⇒ raw numbers + one debug log; never
  a page error, never a call-path dependency.

### CRM repo (`~/projects/crm`)

- Contact domain: `Phones []string` on State, Created/Updated payloads
  (`phones` json, omitempty — old events decode fine), Create/Update
  commands; decide signatures extended; fold sets on create and
  wholesale-replaces on update; app registration, `ContactView.Phones`,
  contact projection carry it.
- HTTP: `phones` field on create + edit forms (kept on validation
  failure), list column, detail display, search haystack, CSV import
  `phones` column (aliases: phone, phone number(s), mobile, tel).
- `internal/httpapi/api.go`: machine API —
  `GET /api/contacts/by-phone?number=` (always 200; empty list = miss;
  400 unusable number) and `POST /api/contacts/{id}/calls` (204; 400
  direction/number/seconds/JSON; 404 missing/deleted/not_active; 409
  other conflicts; 500 logged). Constant-time bearer compare, mounted
  OUTSIDE the passkey gate in BOTH auth modes (`mountAPI` in wire) with
  a boot log line.
- Journal composition: "Incoming call from +49 … (2m 03s) — answered".
  API dispatches WITHOUT an actor (machine entries stay anonymous —
  decision recorded; do not fabricate a user actor).
- Tests (green in the pinned sandbox): `api_test.go` (token 401s;
  exact / trunk / miss / too-short / unrelated lookups; call log 204 +
  journal visible on the detail page; 404 unknown contact; 400s; JSON
  wire shape) + `helpers_test.go` tables (parsePhones: whitespace is
  formatting, not a separator; phoneMatches trunk/refusal matrix;
  findContactsByPhone) + contact domain tests (payload + fold).
- Docs: README machine-API section (with curl), CHANGELOG entries,
  FEATURES (contacts row + new Machine API section), AGENTS.md
  (matching contract, parsePhones trap, no-actor decision, 404 mapping).
- Hygiene: templ generate clean, gofmt clean, targeted sandbox suite
  (httpapi + domain + app) GREEN.

### Webphone repo (this repo)

- `internal/config`: `CRM{URL,Token}` + fail-closed validation.
- `internal/crm` (new package): `Client` (3 s timeout, bearer, lookup +
  log call; ErrDisabled/ErrUnauthorized/ErrNotFound) and `Resolver`
  (TTL cache 6 h positive / 5 min negative, cap 1024, transport
  failures never cached, nil-safe). Stdlib-only — arch test happy.
- Server: `Deps.CRM`; all four number surfaces (History, Messages
  thread list + transcript header, Fax, Voicemail) resolve page numbers
  → `crmNames` → `Names` prop → `displayName` (fallback raw number;
  `data-dial` keeps the number); SSE thread/fax fragments enriched too
  (notifier carries the resolver; newNotifier callers updated).
- `POST /api/calls` (calls_api.go): requireSession + CSRF + contacts
  budget limiter; 204 disabled / unknown-number / success; 400
  malformed; 502 CRM outage (island toasts via i18n `crmLogFailed`
  en+de).
- Island: `config.js crmEnabled` (PBX_CONFIG carries `"crm":bool`
  always — feature flag, not existence probe); `recordCrmCall`
  fire-and-forget in panels.js; hooked in the Terminated branch of
  calls.js beside recordHistory.
- `cmd/webphone/main.go`: client + resolver wiring + boot log when
  enabled.
- Tests: `internal/crm` unit suite; server `/api/calls` contract suite
  (anon 401/403, matched journal forwarding asserted at the stub, auth
  unknown → silent 204 no upstream call, outage → 502, disabled → 204,
  malformed → 400); history-enrichment render test (name shown, number
  on the dial button, `"crm":true` in config.js, `"crm":false` when
  off); config validation table.
- Gates, run SEQUENCED on a quiet machine: `go build ./...` ✓, go vet ✓,
  full `go test -count=1 ./...` GREEN except the concurrent session's
  TZ-sensitive views test (attributed above), island node:test **49/49**
  (incl. en/de parity for the new key), live smoke **38/38 + restart
  4/4**, `nix build -L .#webphone` **EXIT=0**.
- Docs: README (config keys + "Ledger CRM (optional)" section),
  CHANGELOG Unreleased, FEATURES row, AGENTS.md integration-seam block.

## b) PARTIALLY done

- CRM full-suite verdict: green except the PRE-EXISTING load-dependent
  flake `TestRegistrationRateLimitContract` (internal/identity —
  untouched by this train; passes 3/3 in isolation; failed twice under
  full-suite load while my webphone gates ran concurrently). A
  quiet-machine full-suite rerun has NOT happened yet.
- Git record: the auto-commit daemon swept the webphone train into
  chore commits mixed with the concurrent composer train (narrative
  lives in the docs, not a dedicated commit). CRM working tree held my
  changes, but I did not verify the CRM daemon's commit/push state.
- Failure-feedback map: the `crmLogFailed` toast exists in code and
  i18n but was NOT added to the failure-map table in webphone AGENTS.md.

## c) NOT started

- Webphone `buildflow` (THE quality gate).
- CRM `buildflow` (only go build + race tests ran in the pinned sandbox).
- `nix flake check` (webphone) — only `nix build .#webphone` ran.
- Stack browser E2E — island JS changed; runbook says re-run
  (`nix build -L .#telephony-browser` in the stack repo).
- OpenAPI spec: `/api/calls` missing from `openapi.json` (contacts API
  has spec-vs-handler pinning; this endpoint does not).
- TODO_LIST sweeps in both repos.
- Plan doc in `docs/planning/` (this report stands in).
- NixOS module typed `crm.{url,token}` options (freeform documented
  instead — decision to record).
- Stack input bump / aarch64 cross-build (release territory).
- Island unit test for `recordCrmCall` (coverage is server-side
  contract + i18n parity only).
- CSV import E2E for the phones column (unit-level only).

## d) Totally fucked up

- First `parsePhones` split on WHITESPACE — shredded "+49 30 12345678"
  into four garbage "numbers". The sandbox gate caught it; worse, my
  first test assertion ("12345678 must not match") was ALSO wrong —
  under the ≤4 prefix-delta rule it legitimately matches the same line.
  Two bugs canceling; both were re-derived from the contract.
- Trunk-prefix matching ("030 …" vs "+49 …") missing in the first cut —
  masked by the garbage bug (tests passed for the wrong reason). Fixed
  with trunk-strip variants.
- Ran webphone and CRM race suites CONCURRENTLY (twice) → starved the
  CRM identity test → chased a "pre-existing flake" that was my own
  scheduling.
- Smaller gate-caught bugs: `io_Copy` typo; nonexistent
  `domain.ErrNotFound` referenced; dead `app.System` in a test helper;
  stub helper clobbering fields set before it; lowercase call to a
  renamed-exported views function; missing `context` import; stale-file
  edit refusals after the concurrent session touched shared files.

## e) Improvements

1. Sequence gate runs; never overlap two `-race` suites on this host.
2. Read each repo's toolchain traps BEFORE the first build (hit the CRM
   sibling-drift trap on command one; used verify-pinned.sh after).
3. Derive test expectations from the CONTRACT, not from what the stub
   currently returns.
4. Grep for helper-name collisions before writing test files.
5. Use view (not bash sed) before editing files a concurrent session
   may have touched.
6. When two changes cancel out, stop and re-derive both from spec.
7. Write the plan doc first — the whitespace decision would have been
   forced before code.
8. Add one HTTP-level integration test spanning island → server → CRM
   stub (current coverage is per-layer).
9. Record daemon-swept trains with a narrative doc commit boundary.
10. Cross-check user-facing failure surfaces against the AGENTS
    failure-map table and the stack runbook at implementation time.

## f) Next things (ordered, 24)

1. Re-run the CRM full suite in the pinned sandbox, quiet machine.
2. Webphone `buildflow` (inside `nix develop`).
3. CRM `buildflow` (`scripts/buildflow.sh` modes for gitleaks/codespell).
4. `nix flake check` (webphone).
5. Stack browser E2E (`nix build -L .#telephony-browser`).
6. `/api/calls` → `openapi.json` + spec-vs-handler test.
7. Verify CRM repo daemon commit/push state.
8. Add `crmLogFailed` row to the failure-map table (webphone AGENTS.md).
9. TODO_LIST sweeps in both repos.
10. Island unit test for `recordCrmCall` (helpers.mjs fetch stubs).
11. Decide typed NixOS `crm.{url,token}` options vs documented freeform.
12. Idempotency key on `/api/calls` (a retry could double-journal).
13. Single-flight resolver lookups (30-row history cache miss burst).
14. CSV import E2E for the `phones` column.
15. Stack-side wiring story for CRM URL/token (secrets dir) when
    co-located with the stack.
16. Backoff/Retry-After handling when the CRM rate limits (future).
17. Cross-doc the CRM surfaces in the stack runbook's webphone error
    contract section.
18. Restore-drill: prove call_logged events survive a journal restore.
19. Lookup hit/miss/timeout counters (slog debug exists; counters
    better).
20. Multi-contact number policy: webphone takes the first match;
    consider surfacing "+N more".
21. Confirm English-only journal bodies are wanted for CRM lines.
22. Island local recent-calls enrichment via a lookup endpoint.
23. Include `internal/crm` in the next erraudit tier-2 re-measure
    (due 2026-10-22).
24. Plan v1.1: CRM + webphone on different hosts (TLS, token rotation).

## g) Questions

1. The CRM's `TestRegistrationRateLimitContract` is load-dependent (the
   per-connection limiter under parallel -race load). Fix the test in
   the CRM repo (control connection reuse), treat it as go-cqrs-lite
   drift to re-pin, or accept flakiness under load?
2. Should webphone's NixOS module grow typed `crm.{url,token}` options
   now (token via environmentFile), or stay freeform until the CRM
   actually deploys next to the stack?
3. Is incoming-call-card enrichment (island fetches and renders the
   name on the ringing card) in scope for v1.1, or is tab-level
   enrichment the product boundary?
