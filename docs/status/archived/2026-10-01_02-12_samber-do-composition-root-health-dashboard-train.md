# Status: samber/do v2 composition root + go-health/go-health-dashboard train

**Date:** 2026-10-01 02:12 · **Session scope:** one session, executed the
owner's request: "more service oriented architecture with samber/do v2 and
/home/lars/projects/go-health and /home/lars/projects/go-health-dashboard".
This report covers ONLY this session's work and what was noticed along the
way. Point-in-time; re-verify before treating as current truth.

> ARCHIVED 2026-10-03 (docs-health v6 sweep): the train shipped green and its
> debts closed across the 2026-10-01/02 trains — health.css input + rebuild
> script + canary, erraudit re-measure 0/0, ELF-byte verify, vulnix gates,
> AGENTS compaction (404→377), family pins for the later codes; the
> store.ping/close/blob.probe pins fold into the error-code registry work,
> and the owner legs (v2.9.0 fold, /health exposure policy) ride the TODO
> dashboard row. Per-item verdicts inline.


**Overall verdict:** SHIPPED GREEN. The refactor landed end-to-end and every
gate passed (16/16 Go packages, buildflow green, `nix flake check` all
passed incl. the KVM backup VM, aarch64 cross-build OK, release-binary
smoke 48+4/0-failed with exact `/version`). But the session had real
self-inflicted friction (see d) and left genuine gaps (see b/c) — most
importantly: the health.css rebuild recipe is missing its dark-variant line
wherever it is recorded, the new error codes lack family pins, vulnix and
the aarch64 ELF check did not run, there is no narrative commit, and
**remote main is AHEAD of local main** (concurrent session or daemon push
desync — uninvestigated, deliberately untouched).

---

## a) FULLY DONE

~~1. **Research & design** — read the 2026-09-19 DI/health review (which had~~ done — this session (report of record)
   deliberately rejected samber/do), go-health + go-health-dashboard
   AGENTS/APIs, samber/do v2 vendored source, webphone's server/config/nix
   wiring; loaded the samber-do-best-practices and buildflow skills; measured
   before adopting.
~~2. **`internal/app` composition root on samber/do v2** (app.go, dashboard.go):~~ done — this session (report of record)
   one container owns object lifetime; `cmd/webphone` reduced to process
   concerns (config, logging, signals, httputil listener); shutdown order =
   HTTP drain → probe draining → container cascade; DO-1..DO-6 rules applied
   (no global injector, MustInvoke only in provider closures/New, no
   Override*, self-contained shutdowns, services framework-free with
   adapter-side conformance assertions in app.go).
~~3. **Named critical services**: `"sqlite"` = new `store.Database` adapter~~ done — this session (report of record)
   (HealthCheck = PingContext, idempotent Shutdown = Close, SQL() accessor),
   `"blob-dir"` = `blob.Store.HealthCheck`; check names are the
   `WithCriticalServices` contract; both eagerly invoked in `New` with named
   errors (go-health's lazy-service gotcha closed by design).
~~4. **Single-home blob write probe**: `blob.ProbeWrite` + `Store.HealthCheck`;~~ done — this session (report of record)
   server's `/healthz` blob check now delegates to it (the old local
   `probeBlobDir` deleted).
~~5. **`config.Dashboard{Enable,Title}`** with `WEBPHONE_DASHBOARD__*` nesting;~~ done — this session (report of record)
   zero value = disabled (opt-in operator surface); round-trip test added.
~~6. **Probe wiring with nil-fallback**: `server.Deps.Probe` (nil → the exact~~ done — this session (report of record)
   former `health.NewChecks` construction) — every existing test composition
   kept working untouched; `/healthz` byte-untouched; one probe instance
   threaded to server AND dashboard (no second readiness truth; probe
   deliberately NOT registered in the injector it reads — recursion class).
~~7. **Dashboard mounted at `/health`** (config-gated): both mux patterns~~ done — this session (report of record)
   (`/health` + `/health/`), probe aliases `/health/{livez,readyz,startupz}`,
   favicon route disabled (webphone owns it), rate limit 60/min, trend ring
   240, embedded Datastar SDK same-origin, `WithCSSPath("/assets/health.css")`.
~~8. **CSP boundary**: `httputil.Nonce` + `dashboard.RecommendedCSP(nonce)`~~ done — this session (report of record)
   override scoped to the dashboard subtree ONLY; per-request nonces;
   `TestAppServesMainPageUnderStrictCSP` pins that the app pages keep the
   strict policy; smoke checks the nonce'd CSP on /health.
~~9. **`/assets/health.css`**: scoped Tailwind v4.3.3 build (dashboard +~~ done — this session (report of record)
   templ-components layout/display/feedback/utils/datastar sources), prettier/
   treefmt-formatted, `tw.css` (app surface) untouched; **186/186 rendered
   classes covered** after the dark-variant fix; embedded + served with its
   own route.
~~10. **NixOS module**: Caddy vhost gained the unbuffered `/health/*` proxy~~ done — this session (report of record)
    handle (same class as `/events`); no new module options needed (dashboard
    rides freeform `settings`).
~~11. **Smoke suite**: boots WITH dashboard enabled (standing coverage) — 4 new~~ done — this session (report of record)
    checks (HTML+nonce, CSP shape, same-origin SDK, probe alias); failure
    messages enriched (inline-script excerpt); 48+4 green on the
    version-stamped nix release binary.
~~12. **Dependency work**: go-health-dashboard v0.10.1 + go-datastar added,~~ done — this session (report of record)
    samber/do promoted indirect→direct, go.sum/tidy settled, vendorHash
    ritual (fakeHash → got: → applied), x86_64 + aarch64-linux nix builds
    green; dependabot root `gomod /` entry already covers the new modules.
~~13. **Footprint gate**: GO — +1,132,918 bytes unstripped (+5.1%); stripped~~ done — this session (report of record)
    release binary ~15.4 MB (was 15.25) — far inside the ≤ +8 MB / ≤ +20 %
    gate that rejected the 2026-09-30 setup-shell adoption.
~~14. **Docs updated**: AGENTS.md (composition-root + dashboard-seam invariants,~~ done — this session (report of record)
    intro), CHANGELOG (Unreleased: Added ×2), README (dashboard section incl.
    fencing guidance), FEATURES.md (3 rows updated/added).
~~15. **Tests added**: app suite (probe triple, dashboard on/off, SSE stream,~~ done — this session (report of record)
    CSP boundary, shutdown idempotence), store.Database adapter test, blob
    probe tests ×3, config round-trip.

## b) PARTIALLY DONE

~~1. **AGENTS.md health.css recipe is INCOMPLETE**: it names the @source set~~ done — health.css.input + scripts/build-health-css.sh + the checks.health-css canary landed (nix-review batch 2)
   and the tailwindcss_4 pin, but does NOT record the
   `@custom-variant dark (&:where(.dark, .dark *));` input line — without it
   a rebuild silently loses every `dark:` utility (exactly the bug this
   session hit). The full input.css currently lives only in /tmp (evaporates
   with the session). ~15-minute fix; see f/1.
~~2. **Family pins for the NEW error codes**: `store.ping`, `store.close`,~~ done in part — the db adapter test covers shape (db_test.go:140); the family_test pins for store.ping/close/blob.probe were never added; the error-code registry table (TODO tooling row) will surface them
   `blob.probe` classify Infrastructure but got no per-seam assertions in the
   packages' family_test.go pattern (the db_test covers shape; the
   family_test pins were not extended). erraudit tier-1/tier-2 were NOT
   re-run explicitly per the AGENTS commands either (buildflow's erraudit
   step ran and flagged one blank-identifier issue, which was fixed).
~~3. **Working tree**: the dark-variant `health.css` rebuild is STAGED but not~~ process record — daemon-owned commits; the runbook's narrative-commit line is the standing mitigation
   yet committed (daemon lag); every other change landed via the daemon's
   heuristic auto-commits — there is NO narrative/phase-boundary commit for
   this train (runbook asks for one).
~~4. **Push state unverified-end**: local main (17684fc) is BEHIND remote main~~ resolved by events — the divergence was LIVE daemon behavior (observed twice; documented in the TODO dashboard row)
   (241b908) — remote contains commits this session never made and never
   pulled. Concurrent session or daemon desync; left untouched per the
   never-revert-others'-work rule.
~~5. **vulnix**: not run over the new runtime closure (govulncheck ran via~~ done — vulnix green on the release trains since (release.sh gates on it)
   buildflow and was green; the recorded release gate `nix run .#vulnix` was
   not exercised).
~~6. **aarch64 ELF verification**: cross-build succeeded and produced an output~~ done — aarch64 ELF bytes verified at the 12:57 close-out (183)
   path, but the AGENTS rule "verify by ELF bytes, never exit code alone"
   was not applied to the new binary.
~~7. **doanalyzerv2 DO-1..DO-6 audit**: done mentally + via the skill's~~ done in part — adapter-side conformance assertions landed (app.go, AGENTS); the private analyzer never ran
   checklist, but the private analyzer was not executed over internal/app.

## c) NOT STARTED

~~1. **Release train v2.9.0**: no flake version bump, no tag, no gh release;~~ routed — the v2.9.0 fold = owner §g2 (TODO dashboard row)
   CHANGELOG `Unreleased` holds the whole train.
~~2. **Consuming-stack work** (nix-international-telephony / pbx-artmann): no~~ routed — the stack E2E obligation (cross-repo/island-honesty rows)
   per-train lock bump, no browser E2E re-run there (no island/shell markup
   changed, so the E2E re-run trigger technically did not fire — but the new
   `/health` surface + vhost handle deserve a stack-side look).
~~3. **Stack vhost fencing decision** for `/health` (remote_ip matcher like~~ owner — stack /health exposure policy (TODO dashboard row)
   the probe triple, basic auth, or public+PublicMode) — owner decision,
   see g).
~~4. **TODO_LIST.md** sweep for this train (nothing added/updated).~~ done — the T21 + v5/v6 harvests updated TODO_LIST
~~5. **docs/error-contract.md**: dashboard surfaces (503 on drain,~~ not adopted — the deliberate exclusion stands (operator-facing); below the bar
   rate-limited 429, stale-pusher warn) not added to the failure→feedback
   table (arguably operator-facing, not user-facing — undecided).
~~6. **openapi boundary note**: decision made (dashboard stays OUT of the~~ not adopted — the boundary decision stands (2026-09-22); a doc line below the bar
   product-API spec per the 2026-09-22 boundary decision) but recorded only
   in this session's reasoning, not in a doc.

## d) TOTALLY FUCKED UP (self-inflicted, all recovered)

~~1. **Useless footprint spike ×2**: two "delta: 0" builds before realizing the~~ process record
   Go linker dead-code-eliminates unreachable references (`var _ =
   dashboard.New` doesn't link the package). Correct move would have been:
   measure with reachable wiring or accept measuring at the end (which is
   what happened anyway).
~~2. **`go mod vendor` in a proxyVendor repo**: I created a `vendor/` working~~ process record
   tree that this repo deliberately does not carry (nix builds vendor from
   the FOD). It broke the buildflow gate with 53 vendor-consistency findings
   and cost a debugging round before `trash vendor` fixed it. Should have
   checked `.gitignore`/`proxyVendor = true` BEFORE vendoring.
~~3. **Mux subtree mistake**: first mount registered only `"/health"` (exact~~ process record
   match in Go 1.22+ mux semantics) — `/health/sse` and friends 404'd; app
   tests caught it; fixed with both patterns. Should have known upfront.
~~4. **Smoke variable shadowing**: my own dashboard check reused the `page`~~ process record
   variable, so the later "zero inline scripts" check inspected the
   DASHBOARD page and failed. I then chased the phantom across two more
   debugging rounds —
~~5. **— including a port collision blind**: my debug boot on :18099 failed to~~ process record
   bind (dnsblockd owns that port, per AGENTS' own quick-start example port!)
   and I fetched and analyzed DNSBLOCKD's page as if it were webphone's
   (inline theme script confusion), plus several `kill` attempts silently
   failed (`kill: unsupported builtin` in that shell) leaving stray servers
   around during the session (all gone at close, verified).
~~6. **Three broken CSS-coverage regexes in a row** (selector-boundary regex,~~ process record
   missing import, CSS-escape blindness) before settling on
   backslash-stripped substring comparison. Verification tooling was sloppier
   than the code it verified.
~~7. **health.css built WITHOUT the dark custom variant on the first try** —~~ process record
   caught by the class-coverage check (the check was worth it), but a
   moment's thought about templ-components' class-based dark mode would have
   prevented it.

## e) WHAT WE SHOULD IMPROVE (durable lessons this train earned)

~~1. **Record build inputs next to build outputs**: an asset generated from an~~ done — health.css.input committed (nix-review batch 2)
   input.css should have that input COMMITTED (e.g.
   `internal/web/assets/health.css.input` or a script), not reconstructed
   from prose in AGENTS.md. Same applies to tw.css.
~~2. **Port hygiene for live debugging**: pick a random high port, never the~~ process record
   AGENTS example port (18099 is dnsblockd's); verify the boot log says
   "listening" before trusting any response.
~~3. **Trust failing tests over theories**: the mux 404s and the shadowed~~ process record
   `page` both gave the answer in one read; I theorized past them.
~~4. **Verify the verifier**: a coverage check with a broken regex is worse~~ process record
   than no check (false alarms AND false confidence). Test the checker on
   one known-present and one known-absent class before trusting a sweep.
~~5. **Check repo vendoring posture before any `go mod vendor`** — proxyVendor~~ process record
   repos (this one) never carry vendor/.
~~6. **Narrative commits at phase boundaries** even with the daemon running —~~ process record
   history is the recovery path when a train spans gates this heavy.
~~7. **AGENTS.md is over its size budget (582/377 lines)** — this train ADDED~~ done — compacted 404→377 at the 12:57 close-out
   ~60. The next doc pass should push dashboard/composition detail into
   docs/ and keep AGENTS.md to rules-only.
~~8. **New error paths need their family pins in the same train**, not as~~ done in part — later trains landed family pins same-train (T22); the store.ping/close/blob.probe pins remain open (b2)
   follow-up — the convention exists precisely so it is not forgotten.

## f) NEXT — up to 50 things, rough priority order

~~1. Commit the staged dark-variant health.css (daemon may have done it by~~ done — committed + the 2026-10-02 rebuild fixed a genuinely stale artifact (zombie classes, --blur-xs)
   read time — verify).
~~2. Record the exact health.css input (with `@custom-variant dark`) — commit~~ done — health.css.input + scripts/build-health-css.sh + canary (CHANGELOG Unreleased Changed)
   the input file or a `scripts/build-health-css.sh`; update AGENTS recipe.
~~3. Investigate the local-behind-remote main divergence (241b908 vs 17684fc):~~ resolved by events — LIVE daemon behavior, documented (TODO dashboard row)
   fetch, READ the incoming commits, attribute (concurrent session vs
   daemon), then integrate — never blind-pull.
~~4. Extend store/family_test.go + blob/family_test.go with pins for~~ done in part — db adapter shape covered (db_test.go:140); family pins not added (folds into the error-code registry work)
   `store.ping`, `store.close`, `blob.probe`.
~~5. Run erraudit tier-1 + tier-2 explicitly (`--type-aware~~ done — re-measured early 2026-10-01: tier-1 0, tier-2 0 (T16)
   --disable-extensions`, `--enforce-go-error-family`) and confirm 0.
~~6. Run doanalyzerv2 over internal/app (DO-1..DO-6, the private analyzer).~~ done in part — adapter-side conformance assertions (app.go); the analyzer never ran
~~7. Run `nix run .#vulnix` (triage CLI: `webphone-vulnix-triage`) over the~~ done — vulnix gates green on the release trains since
   new runtime closure.
~~8. Verify the aarch64 binary by ELF bytes (readelf machine = AArch64).~~ done — ELF bytes verified at the 12:57 close-out (183)
~~9. Decide + implement stack-side `/health` fencing (see g/1).~~ owner — stack /health exposure policy (TODO dashboard row)
~~10. Write the narrative commit for the train (or a~~ process record — daemon-owned commits; the runbook mitigation stands
    `docs/status`-referencing commit message if history is already daemon-
    committed — an empty commit with the story is acceptable per owner call).
~~11. Update TODO_LIST.md (remove/mark anything this train obsoated; add the~~ done — the T21 + v5/v6 harvests
    stack-side follow-ups).
~~12. Add dashboard surfaces to docs/error-contract.md (or record the~~ not adopted — the deliberate exclusion stands; below the bar
    deliberate exclusion with rationale).
~~13. Record the openapi-boundary decision (dashboard out) in~~ not adopted — the boundary decision stands; below the bar
    docs/planning/2026-09-22_12-02 area or the spec's rationale comment.
~~14. Add an app test pinning probe refresh cadence when the dashboard is on~~ not adopted — below the bar
    (CachedResponse advances; startupz latches; dashboard warn before
    Start).
~~15. Add an app test for `/health` 429 behavior (rate limit) and~~ not adopted — below the bar
    `/health/trend` + `/health/export` (Trend enabled but untested here).
~~16. Consider `WithMaxSSEConnections` + `WithShutdownDrain` defaults for the~~ not adopted — below the bar
    mounted dashboard (SSE budget parity with /events' limiter posture).
~~17. Consider surfacing MORE container checks now that a UI exists: crm~~ not adopted — below the bar (health-washing risk avoided)
    reachability (non-critical, cheap HEAD), retention sweep last-run age,
    gateway webhook target reachability (non-critical) — each must be able
    to FAIL honestly (health-washing rule).
~~18. Rebuild health.css deterministically in CI (nix derivation or flake~~ done in part — the checks.health-css canary catches desync; a nix rebuild derivation not built
    check) so library bumps cannot desync the committed CSS.
~~19. Add a flake check that greps health.css for `:where(.dark` (dark-variant~~ done — the canary guards the dark-variant fingerprint; prettier via treefmt
    canary) and for `/assets/health.css` being prettier-clean.
~~20. Update scripts/release-hygiene.sh if it enumerates assets/routes.~~ not adopted — below the bar
~~21. Re-run the consuming stack's browser E2E once (cheap confidence, even~~ routed — the stack E2E obligation (cross-repo row)
    with no markup change) before the v2.9.0 relock.
~~22. Cut the v2.9.0 release train: fold CHANGELOG, bump webphoneVersion,~~ routed — the v2.9.0 fold = owner §g2 (TODO dashboard row)
    full gates, tag, stack re-pin, pbx-artmann relock (runbook).
~~23. In the stack: expose the dashboard behind its chosen fence; add a~~ owner — rides the exposure-policy call (TODO dashboard row)
    monitor365/Uptime check against `/health` JSON (Accept header).
~~24. Consider dashboard `PublicMode` for any internet-exposed variant.~~ owner — rides the exposure-policy call (PublicMode)
~~25. Document the dashboard in the stack runbook § "Webphone error contract"~~ routed — the stack-runbook cross-ref rides the TODO boot-contract tail row (prepared patch)
    cross-ref (both sides in sync rule).
~~26. Trim AGENTS.md toward the 377-line budget (move detail to docs/).~~ done — compacted 404→377 at the 12:57 close-out
~~27. Sweep docs/lessons.md with this session's portable war stories (port~~ not adopted — the portable stories live in this report; below the bar
    collision, linker DCE spike, vendor posture).
~~28. Add `internal/app` to arch_test's coverage if the service-package list~~ not adopted — app is not a service package (own text: currently fine)
    ever grows rules for composition roots (currently fine — app is not a
    service package).
~~29. Consider a test container variant of internal/app (override gateway to~~ not adopted — below the bar
    loopback etc.) if more app-level tests appear.
~~30. Rename `probeRefreshIf` → fold into a documented option table if more~~ not adopted — below the bar
    probe knobs appear.
~~31. Check go-licenses output for go-datastar/go-health-dashboard (buildflow~~ not adopted — below the bar (buildflow's license step ran green)
    ran; confirm the report includes the new modules).
~~32. Add the dashboard page to the DOM-contract docs as an explicitly~~ not adopted — below the bar
    OUT-of-contract surface (one line) so nobody adds its ids to the
    contract later.
~~33. Consider CSP `report-to`/console assertion in the stack E2E for /health~~ not adopted — below the bar
    (zero console errors under the subtree policy).
~~34. Probe alias duplication note: consider dropping `/health/readyz` alias~~ not adopted — below the bar
    if it proves noise (RegisterRoutes requires non-empty patterns —
    documented workaround stays).
~~35. Island: nothing needed — verify once in a manual browser session that~~ not adopted — below the bar
    /health does not interfere (different subtree; expected no-op).
~~36. Consider exposing `dashboard.title` through the NixOS module docs table~~ not adopted — below the bar
    (settings is freeform — README table only).
~~37. Add a `--dashboard` flag-equivalent to the quick-start block in AGENTS~~ not adopted — AGENTS line-capped
    commands (env line) so future sessions boot it by default in smoke-like
    fashion.
~~38. Metrics: optionally enable dashboard `/health/metrics` later if ops~~ record stands — metrics OFF (recorded decision)
    wants per-check series (currently OFF; webphone /metrics stays the one
    scrape surface — recorded decision).
~~39. Federation: the stack could later run one health-hub (cmd/health-hub in~~ routed — ROADMAP health-surface long shots (stack-side federation)
    the dashboard repo) federating webphone + other services instead of
    webphone's embedded page — evaluate at the next architecture review.
~~40. Re-measure erraudit tiers on the 2026-10-22 date AGENTS already pins.~~ done — re-measured 2026-10-01; next due 2026-10-22 (standing watches row)
~~41. Confirm the daemon eventually committed+pushed the staged health.css~~ done — the artifact is committed + canary-guarded (the 2026-10-02 rebuild proves it)
    (`git ls-remote` vs local) before any release fold.
~~42. Consider `session.Attach`-level logging filter for /health/sse request~~ not adopted — below the bar
    spam (pusher reconnects) in the request log — decide after observing.
~~43. Add a golden render test for the dashboard page (pin nonce-attr~~ not adopted — below the bar
    presence structurally, not bytes) if library bumps start churning markup.
~~44. Evaluate `do.InvokeStruct[server.Deps]`-style wiring if Deps grows more~~ not adopted — below the bar
    container-backed fields (currently manual MustInvoke list — fine at 13
    fields).
~~45. Write the deferred "why the probe is not in the injector" rationale into~~ other repo — go-health upstream docs; not filed
    go-health's upstream docs (consumer-side note; PR-worthy).

## g) Questions for the owner (cannot be figured out from the repo)

~~1. **Stack-side `/health` exposure policy**: when pbx-artmann adopts this,~~ owner — stack /health exposure policy (TODO dashboard row)
   should `/health` be (a) fenced by remote_ip exactly like the probe triple
   (operator-VPN only), (b) public with `dashboard.PublicMode` (masked names,
   unauthenticated status page), or (c) behind basic auth at the vhost? This
   decides the stack's vhost block and whether I should flip PublicMode on
   in webphone's mount options.
~~2. **Git history**: remote main (241b908) is ahead of local main (17684fc)~~ resolved by events — LIVE daemon behavior, documented (TODO dashboard row: never verify via push logs, only ls-remote)
   with commits this session did not author. May I fetch and READ them to
   attribute (concurrent session vs daemon) and integrate — and do you want
   a narrative commit added for this train on top, given the daemon already
   heuristic-committed the work?
~~3. **Release timing**: fold v2.9.0 (composition root + dashboard) now while~~ routed — the v2.9.0 fold decision = owner §g2 (TODO dashboard row)
   the train is green, or hold `Unreleased` and bundle with the next feature
   (e.g. the fax-paperless plan) into one release?

---

_Report closes the session's work; no further changes made after the final
gate battery (flake check all-passed, smoke 48+4, 16/16 packages)._
