# Status — SUPERB catalog-contracts plan execution (openapi + asyncapi lanes)

**Claims verified as of 19:15 CEST, 2026-10-24.** Push/CI statements rot
fastest; this snapshot names the exact commits and run IDs it saw.

Session: execution of the gated Pareto plan
[docs/planning/2026-10-24_16-50_SUPERB-catalog-contracts-plan.md](../planning/2026-10-24_16-50_SUPERB-catalog-contracts-plan.md)
(M1–M18). A **concurrent session** ran the eventcatalog lane in parallel
(`internal/domain/catalog.go`, `webphone catalog` CLI, go.mod vendorHash —
plan item M4 de facto done by it); this session took the non-colliding
lanes: `internal/server`, `internal/server/api`, `internal/server/hooks`,
`internal/gateway`, `cmd/webphone` dispatch, `docs/contracts/`.

## Direct answers to the three opening questions

**What did you forget?**

1. **The error-code registry.** Three new `errorfamily` codes
   (`webphone.server.openapi_render`, `webphone.server.wire_render`,
   `webphone.cli.contracts`) shipped without registering them —
   `TestErrorCodeRegistryIsFresh` (internal/arch) is red, and it is red
   **on pushed main** (CI run 40113240902, failed 16:50Z). I never ran
   `go test ./...` before the daemon pushed; the full-suite debt came due
   publicly.
2. **CI watch for 3+ hours.** The ratified bar says check CI after every
   push; I first looked at 19:10. The daemon pushes silently, so the red
   sat unowned.
3. **The smoke pin.** `scripts/webphone-smoke.py:774` asserts
   `"openapi": "3.1.0"` — but the catalog exporter emits **3.0.3**. The
   served contract silently downgraded its spec version in the swap
   (discovered at report time; no CI leg runs the smoke, so nothing
   caught it).
4. **Train-start buildflow** (AGENTS rule) — never run this session.
5. **AGENTS.md memory updates** — the catalog API gotchas learned today
   are in this report, not yet in the memory file.

**What could you have done better?**

1. **Commit + verify as ONE unit.** The auto-commit daemon snapshots and
   pushes every few minutes; working half-verified in watched dirs means
   _it_ decides what ships. The red-test ghost route and the unregistered
   codes both landed pushed that way.
2. **Red-test first, then trust.** My drift test's first version gave a
   FALSE GREEN — the ghost passed because lowercased OpenAPI method keys
   get canonicalized past the CSRF gate (403 masks the catch-all 404).
   Only verifying the red path exposed it; the test is trustworthy now
   (uppercase verbs), but the order should always be red-test → green.
3. **Register error codes in the same commit as the code that mints
   them** — the arch test exists precisely for this; make it part of the
   "new error path" checklist, not a full-suite discovery.
4. **Reconcile with concurrent sessions at session START** — I found the
   other session's lane ~20 minutes in, by luck of reading the log.

**What could you still improve?**

1. The `catalog` API knowledge (security-scheme via `Registry()`, no
   response headers, 204 schema rule, kafka default) belongs in
   AGENTS.md now that two repos depend on it.
2. Upstream candidates (headers, 204, multipart, spec version) should be
   FILED in go-cqrs-lite with verified evidence, not just listed here.
3. The generated asyncapi.yaml's formatter stability (treefmt scope) is
   unverified — the freshness gate design depends on it.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                                                                                                                                                                                                                                         | Evidence                                                                                                                                     |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **M3 DTO hoists** (behavior-parity): `loginRequest`, `sessionIdentity` + description tags, `ContactListResponse`, `ContactSaveRequest`, `CallLogRequest`, `CSRFTokenResponse`                                                                                                                                                                                                                                                                | internal/server/session_api.go, internal/server/api/{contacts,calls,csrf}_api.go; suites green                                               |
| 2  | **M5 catalog registry** — 8 operations over the 4 core paths, `cookieAuth` scheme (`Registry().AddSecurityScheme`), `Validate()=0` pinned                                                                                                                                                                                                                                                                                                    | internal/server/apidocs.go; `TestAPIContractIsValid`                                                                                         |
| 3  | **M6 `/openapi.json` swap** — the 227-line hand-written `openapiSpec` const DELETED; handler serves deterministic generated bytes (same route/content-type/no-cache); 3 pin tests adapted (`Retry-After` contract now asserted via 429 description text — the exporter cannot express response headers)                                                                                                                                      | server.go; middleware_test.go; `TestOpenAPIEndpoint` green UNTOUCHED                                                                         |
| 4  | **Honesty deltas of the swap** — DELETE /api/session documented 200/401 (a LIE; handler always answers 204) → now truthful 204+403; added the hidden 403-CSRF / 429-limiter / 500-csrf codes; field-by-field diff old→new verified (generated ⊇ old everywhere except the corrected lie)                                                                                                                                                     | F26 comparison, this session's log                                                                                                           |
| 5  | **M7 route-vs-registry drift test** + ghost-path red-test (which caught the lowercase-method CSRF masking trap; probes now use canonical verbs)                                                                                                                                                                                                                                                                                              | internal/server/apidocs_test.go `TestAPIContractPathsAreMounted`                                                                             |
| 6  | **M8 full-machine-API extension** — all 11 session-gated surfaces (transcribe, transcripts ×2, push-devices ×3 added) with request/response schemas derived from the real DTOs; seam pin test (404-when-off + Retry-After)                                                                                                                                                                                                                   | commit **a7be533**; `TestOpenAPISeamSurfacesMatchHandlers`                                                                                   |
| 7  | **M9 wire contracts** — 5 `/hooks/*` lanes + gateway receipt + 5 SSE kinds declared from the REAL wire types (`RingRequest`, `HookMessageRequest`, `HookFaxRequest`+`FaxStatusPages`, `HookFaxStatusRequest`, `HookMessageStatusRequest`, `ReceiptEnvelope`); receipt direction corrected Sends→Receives after source check (the bridge answers webphone's POST with it); embedded `FaxStatusPages` schema promotion verified field-by-field | internal/server/wirecontracts.go + hooks hoists + gateway/webhook.go; `TestWireContractsIsValid`, `TestWireContractSchemasMatchHandlerTypes` |
| 8  | **M10 `webphone contracts` CLI** — AsyncAPI 3.0 YAML one-shot (no config/db/network); fabricated kafka default server suppressed; `docs/contracts/asyncapi.yaml` (495 lines) committed                                                                                                                                                                                                                                                       | cmd/webphone/contracts.go, main.go dispatch (re-read before edit — no collision)                                                             |
| 9  | **M15 split rule + passkey conditional note** — one home in docs/contracts/README.md (browser JSON→OpenAPI; /hooks/*+SSE→AsyncAPI, never both; passkey/SSO surfaces deliberately absent from both)                                                                                                                                                                                                                                           | docs/contracts/README.md                                                                                                                     |
| 10 | **Pin-site check** — smoke script's only openapi pins are status+version (see the 3.1.0 finding below); stack browser-e2e greps NO old-spec strings                                                                                                                                                                                                                                                                                          | scripts/webphone-smoke.py:770-778, pbx-artmann/tests clean                                                                                   |

## b) PARTIALLY DONE

1. **RED MAIN (top item).** CI run **40113240902 FAILED** (16:50Z) on
   `TestErrorCodeRegistryIsFresh`: 3 codes missing from the doc registry.
   The fix is one command, verified identified:
   `go test ./internal/arch -run TestErrorCodeRegistryIsFresh -update`,
   then review the generated error-contract rows, `go test ./...`, fmt,
   commit. NOT applied — you said report then wait; the fix is ready to
   execute on your GO.
2. **M11 freshness gate** — design decided (regen + `git diff --exit-code`
   via the `webphone contracts` binary inside a flake check leg); NOT
   implemented. Unresolved sub-question: is the generated YAML stable
   under treefmt (formatter-vs-generated fight), and does a binary-run
   leg belong in every `nix flake check`.
3. **M1 decision record** — adopt is DE FACTO (the surface shipped), but
   TODO_LIST.md:68 is untouched, deliberately: the concurrent session's
   eventcatalog work rewrites the context of that 2026-10-19 ruling and
   the record should normalize BOTH lanes in one edit.
4. **CHANGELOG** — my machine-API entry (generated /openapi.json,
   `webphone contracts`, 11 surfaces) is still owed; the concurrent
   session added theirs (eventcatalog, c85aefb, unpushed).
5. **Suite state** — 22/23 packages green (`go test -count=1 ./...` at
   19:12); internal/arch red per above. Gateway package green after the
   `ReceiptEnvelope` hoist.
6. **Commit narrative** — the daemon snapshotted the wire-contracts work
   into 46cc6db + 78017f5 (pushed, chore messages); my narrative commits:
   70bf5a9 (drift-test fix), a7be533 (M8). Unpushed: c85aefb (daemon,
   AGENTS+CHANGELOG eventcatalog lines).

## c) NOT STARTED

- **M2 honesty batch** — F4 report addendum (inline-DTO prereq, now
  executed — annotate as DONE), F5 "routine bump"→418-commits erratum,
  F6 render check, F7 re-run report verify commands (NOTE: the report's
  "zero catalog usage" verification NOW returns differently — go.mod HAS
  catalog; the addenda must state the snapshot predates today's adoption).
- **M12 HARVEST** — this report's (f) list into TODO_LIST.
- **M13 metaengine v4.19→v4.20 train prep** (418 commits, codec/typedfixture).
- **M14 docs/research/README.md index.**
- **M16 hub cross-link** (eventcatalog-hub → asyncapi.yaml).
- **M17 webphone-android spec consumption** (F62–65).
- **AGENTS.md memory updates** (catalog gotchas, `webphone contracts`
  command, drift-test lesson), **docs/lessons.md** incident narratives,
  **dedup-registry sweep-log line**.
- **Gates**: buildflow full (overdue), `nix flake check`, `nix fmt` after
  the latest Go edits, `nix run .#vulnix`, live smoke
  (`scripts/webphone-smoke.py` — will FAIL today on the 3.1.0 pin, see d3).

## d) TOTALLY FUCKED UP

1. **Red pushed main, unowned ~2.5 h.** Registry-drift failure pushed by
   the daemon before any full-suite run; CI run 40113240902 red since
   16:50Z; first looked at 19:10. Root cause: train-start gates skipped +
   daemon pushes faster than verification.
2. **The daemon pushed my red-test ghost** (zz_ghost_probe_test.go +
   ghost route inside apidocs.go, commit e9635bf) — a deliberate
   temporarily-red state, committed and pushed mid-flight. Fixed forward
   in 70bf5a9. Root cause: temp/red-test files must never live in watched
   package dirs while the daemon runs.
3. **Silent OpenAPI version downgrade** — hand spec said 3.1.0, generator
   emits 3.0.3; the smoke script still pins 3.1.0 and will fail on its
   next live run. Undetected by every gate (smoke is not a CI leg).
   Resolution options: pin smoke to 3.0.3 (generator truth) + upstream
   candidate "emit 3.1.0"; or fork the version string client-side
   (dishonest — rejected).
4. **False-green drift test (briefly).** The first version passed WITH the
   ghost present (lowercase methods canonicalize past CSRF, masking the
   catch-all 404). Corrected same-session; lesson recorded above.

## e) WHAT WE SHOULD IMPROVE

1. **Make `go test ./...` + registry freshness part of the pre-daemon-push
   window**: commit early, verify, THEN let the daemon push — or run the
   full suite before every walk-away.
2. **Keep generated artifacts and their pins in one commit**: spec swap +
   smoke pin + registry rows were three separate discovers; one checklist
   item "grep the pins for every served-bytes change" (the AGENTS rule
   exists — I applied it to the DOM pins but not the smoke's version pin).
3. **Status inventories as data**: the seam pin test duplicates handler
   knowledge in code; acceptable today, but a shared
   route→expected-codes table would serve registry, pins, and drift test.
4. **Concurrency protocol**: start-of-session `git log --stat -5` +
   mtimes of shared dirs (cmd/, go.mod, TODO_LIST) BEFORE choosing lanes;
   I did it late.
5. **Skill-format divergence (flagged per skill)**: status-report is
   HTML-canonical; you explicitly demanded `.md` at `docs/status/` —
   honored here, not propagated into the skill.

## f) NEXT — up to 50, tiered (impact first)

**Red-main recovery (do first):**

1. Register the 3 error codes (`-update`), review generated rows.
2. `nix develop -c go test -count=1 ./...` green.
3. `nix fmt`; `bash scripts/whitespace-drift.sh` after staging.
4. One-call commit; push (manual bar applies: red with verified fix);
   watch CI to green (state the anchor).
5. Fix the smoke's OpenAPI version pin (3.1.0 → 3.0.3) + its comment.
6. Run the live smoke against a fresh binary (`/openapi.json` leg).
7. CHANGELOG `[Unreleased]`: generated /openapi.json (11 surfaces),
   `webphone contracts` + committed asyncapi.yaml.

**Plan remainder (M-items):**
8. M1: record the adopt verdict in TODO_LIST.md:68 — reconciling the
2026-10-19 ruling with BOTH shipped lanes (eventcatalog + openapi/asyncapi).
9. M2/F4: report finding-1 addendum (hoisting prereq — mark EXECUTED).
10. M2/F5: "routine bump" → 418-commits erratum in the report.
11. M2/F6: headless render check of the HTML report.
12. M2/F7: re-run the report's verify commands with the addendum note.
13. M11/F48: freshness script (regen + `git diff --exit-code`).
14. M11/F49: flake check leg — resolve treefmt-vs-generated-YAML first.
15. M12/F50–51: HARVEST (f)-list into TODO_LIST (docs-health mode).
16. M13/F52–56: metaengine v4.20.0 delta read + bump-train note.
17. M14/F57–58: docs/research/README.md index (8 reports).
18. M16/F61: hub inventory page → link asyncapi.yaml.
19. M17/F62–65: webphone-android consumes /openapi.json + asyncapi.yaml.
20. AGENTS.md Commands block: add `webphone contracts` one-liner.
21. AGENTS.md: catalog API gotchas (Registry().AddSecurityScheme;
Deterministic marshals; `doc:`/`description:` struct tags; kafka
default server; no response headers; 204 schema rule).
22. AGENTS.md: lowercase-method CSRF masking trap (drift-test lesson).
23. docs/lessons.md: daemon-vs-red-test incident (e9635bf/70bf5a9).
24. docs/lessons.md: registry-drift red-main incident + root cause.
25. error-contract.md: review the 3 generated code rows for wording.
26. dedup-registry: sweep-log line (apidocs/wirecontracts structural
similarity — intentional sibling, ruling recorded).
27. Verify a content-type pin still exists for /openapi.json post-swap.
28. Re-run the full island+flake gates (`nix flake check`).
29. buildflow full (train gate — twice overdue).
30. `nix run .#vulnix` before any release claim.
31. Decide: catalog v4.9.0 (pinned) vs v4.10.0 (audit's latest tag).
32. Consider folding the wire-contracts catalog INTO
internal/domain.BuildCatalog (one catalog, two exports) or keep the
two-catalog split — one-home review with the concurrent session's
design in hand.
33. Stack integration: expose the contracts file to the consuming stack's
telephony-eval (optional).
34. Stack browser E2E re-run at next relock (served bytes changed on
/openapi.json only; DOM untouched — verify the obligation scope).

**Upstream filings (go-cqrs-lite catalog, verify-before-filing + github-voice):**
35. openapi: response-header support (Retry-After forced into prose).
36. openapi: exempt 204 from the 2xx-body-schema validation rule.
37. openapi: `WithParam` carries no type (schema-less query params).
38. asyncapi: don't default to a fabricated kafka server (empty-host hack).
39. openapi: multipart request-body support (transcribe modeled body-less).
40. openapi: operationId Pascal leaves dots (`postSession.create`).
41. openapi: emit OpenAPI 3.1.0 (content-schema dialect is 2020-12-shaped).

**Smaller polish:**
42. ETag on /openapi.json (cheap client caching; owner call).
43. `/tmp/generated-openapi.json` + `/tmp/wire.json` render dumps —
consider a tiny `-render -o` debug flag instead of temp tests.
44. wirecontracts_test.go `jsonContainsField` recursion — fine for a
test; note for future schema assertions (go-snaps candidate).
45. Consider pinning the generated /openapi.json bytes as a golden
fixture (drift test covers routes; bytes cover wording).
46. Push c85aefb (daemon's AGENTS+CHANGELOG) — verify end state.
47. Record the concurrent-session handoff facts (eventcatalog lane) in
TODO_LIST so the next session doesn't re-audit it.
48. metaengine M13 outcome may gate the catalog bump (31) — sequence them.
49. Retire `/tmp/old-openapi.json` scratch files (they are outside the
repo; nothing to do in-tree — noted for hygiene only).
50. After green CI: re-verify `git ls-remote` end states (ratified bar).

## g) Questions I can NOT figure out myself

1. **The 2026-10-19 eventcatalog ruling vs today's shipped eventcatalog
   export** (concurrent session): was that ruling formally revised by
   you? Should my M1 record in TODO_LIST normalize BOTH lanes (eventcatalog
   federation + openapi/asyncapi wire contracts) as adopted, or does the
   eventcatalog lane need its own decision record?
2. **catalog v4.9.0 (pinned by the concurrent session) vs v4.10.0 (the
   audit's latest tag)**: deliberate known-good pin, or should the bump
   ride the metaengine v4.20.0 train (M13)?
3. **M11 freshness home**: a binary-run leg inside `nix flake check`
   (every check invocation pays a build+run) vs CI-go-tests-only — which
   cost posture do you want?

_Report: docs/status/2026-10-24_19-15_catalog-contracts-execution-status.md.
Format `.md` per your explicit instruction (skill default is HTML — divergence
flagged, not propagated). WAITING FOR INSTRUCTIONS._

---

## Close-out (20:55 CEST, same day) — the completion train executed

The SUPERB completion plan
([docs/planning/2026-10-24_19-27_SUPERB-contracts-completion.md](../planning/2026-10-24_19-27_SUPERB-contracts-completion.md))
ran its tiers. Claims verified as of 20:55 CEST:

- **1% DONE**: red main RECOVERED — the three error codes registered
  (`docs/error-contract.md` regenerated block, commit 7b7bcc6, CI run
  40153856104 GREEN); the smoke pin trued to the generator's 3.0.3 and
  the live smoke runs 48/48 incl. `/openapi.json`; CHANGELOG Added
  entry landed.
- **4% DONE**: TODO_LIST carries the wire-lane record + the hub
  push-pair update; the audit report annotated (addenda + erratum +
  verify re-run, render-checked); `scripts/contracts-fresh.sh` + the
  `contracts-fresh` flake check gate the committed asyncapi.yaml
  (green path AND red path verified); AGENTS/lessons/dedup carry the
  gotchas and the three war stories; the remainder is HARVESTED into
  TODO_LIST rows.
- **20% DONE (as records)**: both dependency trains priced
  (metaengine v4.19→v4.20 = 418 commits, codec contract — the
  readmodels capability assertions are the risk surface; catalog
  v4.9→v4.10 = 213 commits, no fix for the seven upstream candidates
  — owner-gated ride recommendation);
  `docs/research/README.md` indexes all seven reports; the seven
  upstream exporter candidates carry a TODO row gated on pinned-version
  repro + dupe search (deliberately NOT filed unverified).
- **rest DONE or HONESTLY DEFERRED**: the three-catalog one-home ruling
  is recorded in docs/contracts/README.md; the /openapi.json content
  type is now pinned by test; the hub cross-link turned out to need a
  PIPELINE home (services/ is generated — verified by append+restore,
  zero drift left) and rides a TODO row; android consumption rides a
  TODO row (other repo's gates).
- **Gates, all green**: `go test -count=1 ./...` 31/31 · buildflow
  exit 0 (warnings = the documented 9-unavailable noise + advisory
  baseline/timing notes) · `nix flake check` all checks passed ·
  `nix run .#vulnix` zero real advisories (both findings
  distro-patched range-match noise) · live smoke 48/48.
- **Owner calls left open** (rows, not surprises): the catalog bump
  timing, the ETag/golden-fixture pair, upstream filings, android.
