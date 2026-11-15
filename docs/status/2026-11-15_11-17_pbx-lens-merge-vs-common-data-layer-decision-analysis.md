# Status: pbx-lens merge vs common-data-layer decision analysis

**Snapshot:** claims verified 2026-11-15 11:17 CET at HEAD `42fc38e1`
(worktree clean at session start; no code, tests, builds, or commits were
made this session). Session scope: a PRO/CONTRA analysis of two options —
(a) merging `~/projects/pbx-lens` into `~/projects/webphone`, (b)
extracting a common data layer between them — plus a second pass after a
read-only overlap inventory. The deliverable was decision support, not
implementation. One unrelated finding surfaced (backup gap) and is
recorded as such.

> **Cadence warning:** the live `docs/status/` corpus was exactly 65
> reports before this file (the `.docs-gates.json` ratchet
> `max_unarchived: 65`). This report makes it 66, so the docs-cadence
> gate is RED until one live report is archived or drained. That is the
> first operational debt this session creates.

## What was actually done

The question was answered by reading both repositories, not from memory:
structure, `go.mod` graphs, domain types, storage doctrines, identity
wiring, retention/backup/search, deployment modules, and the existing
cross-references. A subagent independently inventoried duplication; its
report arrived after the first answer and materially sharpened three
points (recorded in §b and §d).

## a) FULLY DONE

1. **Repo inventory measured.** webphone: 30 `internal/` packages,
   88 759 total Go lines (48 319 non-test, 34 492 non-test and
   non-`_templ.go`, 231 test files). pbx-lens: 11 `internal/` packages,
   36 529 total (18 278 non-test, 12 525 non-test and non-generated,
   90 test files).
2. **Dependency divergence mapped.** Both on go 1.27.3 and the same
   fleet pins: templ v0.3.1100, templ-components v1.27.0, cqrs-htmx
   loginpage v4.18.0 / usermgmt v4.20.0 / webauthn v4.13.0,
   go-cqrs-lite system v4.18.0 / metaengine v4.23.0 /
   projectionhost v4.7.1, httputil v1.8.0, go-error-family v0.12.0,
   modernc sqlite v1.61.1. Drift is patch-level (catalog v4.14.0 vs
   v4.15.0; libc 1.80.1 vs 1.80.0; go-jose 4.1.4 vs 4.1.8). The only
   STRUCTURAL fork is `go-appkit` v0.12.0 + systemd v0.1.0 in lens
   versus webphone's documented appkit NO-GO (three filed upstream
   asks) and its samber/do + `httputil.NewServer` composition.
3. **Domain models compared.** webphone has NO call aggregate and no
   CDR persistence: `pbx.CDR` is an 8-field wire DTO
   (`internal/pbx/client.go:99`) folded by a pure projection
   (`internal/reports`). Lens owns a 19-field source-normalized
   `CallView` with stream id + ContentHash
   (`internal/calls/projection.go:20`) over a 1.2k-line call aggregate.
   The union of shared call fields is a number, a timestamp, a hangup
   cause and a duration; direction is derived differently
   (`in`/`out` from `Context == "public"` in webphone vs an explicit
   `inbound`/`outbound` in lens).
4. **Identity keys checked.** Lens keys streams on the CDR uuid; the
   webphone History surface has only the SIP Call-ID — its own code
   says CDR rows carry no call uuid (`internal/domain/callnote.go:7`).
   A deep link or join needs the phone API extended; that is the real
   integration cost.
5. **Acquisition paths proved disjoint.** Lens parses `Master.csv`
   (`internal/ingest/cdr`, 370 lines), Telnyx read APIs
   (`internal/ingest/telnyx`, 672), the receiver webhook JSONL and
   agent transcripts. webphone never parses CSV and has no Telnyx
   client; it calls `/phone-api/history` JSON under a session
   (`internal/pbx/client.go:166`). Zero acquisition code is currently
   duplicated.
6. **Retention semantics contrasted.** webphone deletes rows, journal
   bytes and blobs (`internal/retention`); lens tombstones via
   `call.retired`, keeps the event store append-only, and converges on
   rebuild. Same package name, opposite doctrine.
7. **Search capacity compared.** webphone has SQL `LIKE` transcript
   search (`internal/cqrs/transcriptsstore.go:129`); lens has FTS5
   with ranking and a 1.2k-line search package welded to its view
   schema. No FTS5/MATCH in webphone.
8. **Identity wiring compared.** Both embed usermgmt + loginpage +
   webauthn, but keep two different session truths (webphone
   deliberately ignores the usermgmt session and carries the SIP
   directory password; lens uses the usermgmt gate with `MaxUsers: 1`).
   The OIDC engines are near-duplicates: lens `internal/oidclogin`
   (752 non-test lines) explicitly implements "the webphone sso
   doctrine"; webphone `internal/sso` (361 non-test lines). This is
   the one genuine extraction candidate, and its right home is
   upstream, not a private two-repo module.
9. **Deployment coupling mapped.** `services.telephony.pbx-lens`
   (`modules/telephony/lens.nix`, 782 lines, ~30 options) and
   `services.telephony.webphone` are separate units, data dirs,
   backups and vhosts in nix-international-telephony. The stack's
   textagent already consumes lens read-only via
   `TEXTAGENT_LENS_URL` → `/api/export.jsonl`
   (`modules/telephony/text-agent.py:916`).
10. **Cross-references quoted.** The only artifact-level sharing is the
    EventCatalog domain id and owner card
    (`internal/domain/catalog.go:97,133`). Lens's own rule:
    "webphone gets a read-only deep link, never a data dependency"
    (`pbx-lens/AGENTS.md:7`).
11. **Merge PRO/CONTRA delivered** for both variants (one binary vs one
    repo, two binaries), with the framework wall, blast radius,
    binary-size precedent (+10.4 MB/+68% recorded for a much smaller
    adoption), AGENTS cap arithmetic (215 + 218 = 433 > 377), and
    doctrine collision.
12. **Common-data-layer PRO/CONTRA delivered** against the
    `go-modularize` Unix gate: the decisive failure is gate 2 (one
    consumer, not two — webphone has no call model to share).
13. **Middle path named:** one repo / two binaries / shared
    `internal/*`, which buys CI and auth sharing without runtime
    coupling — and still pays the gate/AGENTS merge.
14. **Verdict stated:** do not merge; integrate read-only; extract OIDC
    upstream; reopen only on named triggers.
15. **Unrelated finding recorded:** `package/backup-script.nix` backs
    up `webphone.db`, `system.db` and `files/` only — no `usermgmt.db`.
    Passkey identities and the auth-audit journal would be lost on
    restore if nothing else covers them (see §d3, unverified).

## b) PARTIALLY DONE

1. **Gate analysis was reasoned, not executed.** I read the
   `go-modularize` skill and applied its five gates, but never ran its
   detection commands (`go mod graph`, co-change statistics, consumer
   variance measurement). "Gate 2 fails" is a judgment from structure,
   not a measured artifact.
2. **LOC numbers were unnormalized in the first answer.** The quoted
   88.5k vs 36.5k included tests and generated templ code (and lens's
   `*_templ.go`, which is gitignored there, was counted from a local
   build). Corrected in §a1: 34 492 vs 12 525 non-test, non-generated.
   The verdict survives, but the "125k lines in one repo" framing was
   overstated.
3. **Binary-size argument is approximate.** Lens's on-disk binary is
   27 MB; webphone's current binary was never built this session, so
   the merged-size claim is extrapolation from a +10.4 MB precedent,
   not a measurement.
4. **Deployment review was sampled.** `lens.nix` was read only to ~line
   60 of 782; its 30 options, backup/purge timers and hardening were
   not enumerated. `package/backup-script.nix` was read for coverage,
   not audited.
5. **No effort estimate.** The answer gives qualitative costs
   ("weeks", "months-scale") without a work breakdown in engineer-days.
6. **UI collision only named.** Green brand vs lens teal theme,
   Tailwind artifacts, and nav/route merge were noticed, not designed.
7. **Public-product risk asserted from the README.** Lens is tagged
   v0.3.0 and `go install`-able; whether any external user depends on
   that was not checked (its AGENTS claims zero external Go
   consumers).
8. **The integration seam is a sketch.** No config schema, route,
   auth-topology or CSP analysis for a `lens.url` seam.

## c) NOT STARTED

1. No durable decision artifact other than this report.
2. No `TODO_LIST.md` row minted for the merge/extraction question.
3. No entry in the fleet reopen-trigger registry.
4. No `AGENTS.md` ruling recorded in either repo.
5. No archive/drain to restore the 65-report cadence ratchet.
6. No `docs/status/INDEX.md` regeneration for this file (owed).
7. No proof-of-concept for the read-only lens deep link.
8. No phone-API extension proposal for exposing the CDR uuid.
9. No OIDC extraction blueprint (API, consumers, versioning, upstream
   ask text).
10. No verification of the `usermgmt.db` backup gap (stack-side or
    host-side coverage).
11. No test, build, lint, or `buildflow` run in either repo.
12. No security/privacy review of a cross-repo read (lens holds
    recordings and transcripts; webphone holds sessions and directory
    credentials).
13. No revisit date or trigger owner assigned.

## d) TOTALLY FUCKED UP

1. **Verdict sequenced before the evidence was complete.** I answered
   before the overlap subagent reported, then had to append a
   correction. The first answer understated the strongest point:
   webphone has zero call model, which makes the shared-layer case
   weaker than "gate 2 is marginal".
2. **Measurement hygiene failure.** Summing every `*.go` file with
   `find ... -exec cat | wc -l` mixed tests and generated code and
   counted lens's locally generated `_templ.go` that is not committed.
   The published 88.5k/36.5k figures are not the code sizes; the
   normalized ones are 34.5k/12.5k. I stated the former as "Go LOC"
   without qualification. This is the session's clearest factual
   error.
3. **Backup gap asserted as fact before end-to-end verification.** The
   script genuinely omits `usermgmt.db`, but I did not check whether
   the stack, a host-level backup, or another unit covers it. It is a
   finding-shaped observation, not a confirmed data-loss bug.
4. **Gate language implied execution.** The 377-line AGENTS cap and
   the Unix gates were cited as constraints; neither gate was run.
   `buildflow` was never invoked.
5. **No artifact until asked.** The entire analysis lived in chat.
   A decision analysis whose output is only a conversation is not
   durable.
6. **No cross-check of claims against the other repo's tests.** Both
   repos have large suites (231 and 90 test files) that encode their
   doctrines; I read code but ran nothing.

## e) WHAT WE SHOULD IMPROVE

1. **Wait for fan-out before the verdict.** If a subagent is dispatched
   to gather evidence that the answer depends on, the answer waits.
   A correction appended to a published verdict costs more trust than
   a delayed first answer.
2. **Normalize every metric before comparing.** For Go size: exclude
   `_test.go`, exclude generated `*_templ.go`, and state whether
   vendored/generated code is counted. Cross-repo comparisons need a
   reproducible command in the report.
3. **Distinguish "verified" from "reasoned".** Gate claims, backup
   coverage, and binary sizes are all checkable; each should carry the
   command or be labelled an estimate.
4. **Make decision analyses produce an artifact by default** — an ADR
   or a status report with the verdict, the evidence table, and the
   reopen triggers. Chat-only analysis evaporates.
5. **Check the operational cost of writing docs.** A status report is
   not free: here it crosses the cadence ratchet. The report should
   name the debt it creates (it does, at the top).
6. **Prefer the smallest true sharing.** The evidence says the only
   real duplicate is the OIDC engine; the instinct to build a "common
   data layer" would have duplicated the fleet libs that already
   exist.
7. **Anchor API integration on join keys early.** The uuid/Call-ID
   mismatch is the kind of fact that kills an integration plan in
   week three; it should be the first thing checked, not a footnote.
8. **State the goal before the options.** "One operator surface",
   "less maintenance" and "one database" lead to different moves; the
   analysis had to assume the goal.

## f) UP TO 50 THINGS TO GET DONE NEXT

**Artifact and decision hygiene**
1. Regenerate `docs/status/INDEX.md` (`python3 scripts/status-index.py`) for this report.
2. Mint a `TODO_LIST.md` row: "pbx-lens merge/extraction — ruled keep-separate; reopen on named triggers".
3. Add that ruling to the fleet reopen-trigger registry with explicit triggers.
4. Record a one-line ruling + pointer in `AGENTS.md` once the owner ratifies.
5. Archive or drain one live status report to restore `max_unarchived: 65`.
6. Run the docs battery (`nix develop -c bash scripts/docs-gates.sh --check`) before pushing this file.
7. Link each repo's AGENTS to the other's ruling (webphone ↔ lens).
8. If the owner wants a formal record, write an ADR under `docs/planning/`.

**Measurement (replace estimates with numbers)**
9. Add a small `scripts/loc-normalized.sh` (non-test, non-generated, per package).
10. Build both binaries on the same toolchain; record size, startup RSS, idle RSS.
11. Diff `go mod graph` module sets; list the exact union and the appkit-only/do-only edges.
12. Run a co-change analysis (`git log --name-only` overlap) between the two repos.
13. Measure lens's FTS index size vs `system.db` size at production scale.
14. Enumerate lens's 30 NixOS options and map them against webphone's module surface.
15. Compare `tests/eval.nix` coverage for both modules in the stack.
16. Confirm whether `usermgmt.db` is covered by the stack/host backup.
17. If not, extend `package/backup-script.nix` and the backup VM test.
18. Add a restore-drill assertion for the identity DB (both repos).

**Integration design (read-only seam)**
19. Design `lens.url` config (default off, both-or-neither rules, unknown-key WARN).
20. Define the History row → lens call-detail deep link (new tab, no island unload).
21. Propose the phone-API extension that exposes the CDR uuid (or fix the join fingerprint).
22. Specify CSP/`frame-ancestors`/hostname rules if any embed is ever considered.
23. Decide auth topology: shared SSO gate vs separate lens login; loopback vs vhost.
24. Reuse the stack's existing `lensContextDays` semantics instead of inventing a second window.
25. Add a smoke/UI-capture surface for the deep link when enabled.
26. Write the degraded-mode rule (lens down = no link, never a broken page).
27. Decide whether the Reports tab stays a CDR fold or becomes lens-backed.
28. If lens-backed: define the cache/TTL and the no-store policy for its fragments.

**Extraction candidates (only if the duplication hurts)**
29. Draft the OIDC fleet-lib API from the two implementations (engine + callback seams).
30. File the upstream ask against cqrs-htmx (or a new fleet lib) with both consumers.
31. Inventory other fleet apps with hand-rolled OIDC (third/fourth consumers strengthen gate 2).
32. Decide on a `telephony-kinds` vocabulary module (SourceKind/Direction/E164, ~60 lines).
33. Decide whether `wire.FlexTime` (45 lines) should be shared.
34. Decide whether EventCatalog constants deserve a shared module (probably not).
35. Evaluate promoting lens `internal/backup` (VACUUM INTO + integrity check) to a fleet lib.
36. Document lens's FTS5 read-model pattern for reuse without code sharing.

**Doctrine and risk**
37. Write the doctrine contrast as a guard doc: disposable index vs authoritative store.
38. If any cross-store code ever lands, add an arch test that forbids pointing retention at the wrong DB.
39. Define the blast-radius rule for any future in-process embed (calls must survive lens faults).
40. Review data exposure for cross-repo reads (recordings, transcripts, credentials).
41. Pin lens's no-write-back invariant with a test if one does not exist.
42. Compare lens's `MaxUsers: 1` posture with webphone's multi-user mapping.
43. Evaluate sharing one identity DB across both apps (single sign-on) vs two.
44. Align retention/GDPR semantics: transcript truth vs bounded deletion.

**Process**
45. Adopt a session rule: fan-out evidence gates the final verdict.
46. Add decision-analysis checklist rows (measure, verify, artifact, triggers).
47. Run both repos' test suites before any future merge claim.
48. Run `buildflow` (webphone) and lens's `poll-gate.sh` if any code lands.
49. Set a revisit date for the decision (calendar row, not memory).
50. Teach the next session the three join-key/doctrine facts up front (§a3, §a4, §a6).

## g) QUESTIONS ONLY THE OWNER CAN ANSWER

1. **Is pbx-lens allowed to stop being a public product?** It is tagged
   v0.3.0 and `go install`-able; if an external user or a future one
   matters, a repo merge changes that contract. If it is permanently a
   private deployment component, the merge option becomes cheaper than
   this analysis assumed.
2. **What is the goal behind the question?** One operator surface (→
   read-only integration), less maintenance (→ shared repo or upstream
   extraction), or one database (→ genuine merge with doctrine work)?
   The recommended move differs for each.
3. **Which of the two concrete debts should be paid first:** the
   read-only lens deep link, or the `usermgmt.db` backup gap? And is
   the latter deliberately host-managed, or a bug to fix in
   `package/backup-script.nix`?

## Appendix — evidence index

| Claim | Where verified |
|---|---|
| No call aggregate in webphone; History is a per-request fold | `internal/reports/reports.go`, `internal/server/panels.go:351`, `internal/server/reports.go:37` |
| 8-field wire CDR | `internal/pbx/client.go:99` |
| 19-field source-normalized CallView | `pbx-lens/internal/calls/projection.go:20` |
| CDR rows carry no call uuid | `internal/domain/callnote.go:7` |
| Lens parses CSV; webphone does not | `pbx-lens/internal/ingest/cdr/cdr.go`, `internal/pbx/client.go:166` |
| Opposite retention doctrines | `internal/retention/retention.go` vs `pbx-lens/internal/retention/retention.go` |
| FTS5 only in lens | `pbx-lens/internal/search`, no FTS5/MATCH in webphone |
| OIDC duplication | `pbx-lens/internal/oidclogin/oidclogin.go` (header cites webphone doctrine) vs `internal/sso/sso.go` |
| Appkit-only structural fork | `pbx-lens/go.mod` vs `go.mod` + webphone AGENTS appkit NO-GO |
| Separate stack units and read-only lens use | `nix-international-telephony/modules/telephony/lens.nix`, `text-agent.py:916` |
| Read-only deep link rule | `pbx-lens/AGENTS.md:7` |
| Shared catalog identity only | `internal/domain/catalog.go:97,133` |
| Backup misses `usermgmt.db` | `package/backup-script.nix:12,23,25` (verified omission; coverage elsewhere unverified) |
