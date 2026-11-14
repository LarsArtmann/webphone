> ARCHIVED 2026-11-14 (docs-health v12 sweep): §f table is ledger-exempt and gates green (check-rows --strict + bare-rows at archive time); archived to bring the unarchived count back under the cadence ratchet (66 → 65, threshold 65)
> Live remainder: nothing exclusive — §f row 1 (dashboard v0.13.3 + unsafe-eval drop) landed 2026-11-13 per the AGENTS currency block; rows 2–10 stay harvest inventory in the ledger table below; §c's "headroom spent" warning is the reason this archive exists.

# Fleet-libs currency sweep (2026-11-13) — self-review: what I forgot, what nearly went wrong

_Snapshot: claims verified as of 09:49 CET, 2026-11-13. Push/CI statements rot fastest; the shas and run ids are inline. Scope: the paste's 24-family utilization-audit refresh (docs/review/libs/) plus the gates it tripped. No code paths touched — docs, index, one archive, one push._

---

## a) FULLY DONE

1. **Currency ground truth for all 24 family reports + index** — remote tags per repo, INCLUDING per-submodule refspec queries for the multi-module repos (cqrs-htmx root/loginpage/usermgmt/identity-model/webauthn; go-cqrs-lite system/event/command/decider/projectionhost/snapshot/metaengine/sqliteengine/catalog/record/id; httputil endpoints/server_timing; templ-components; go-etag; go-sse). Adjacent four (go-branded-id, go-flightrecorder, go-paperless, go-sqlitestore) swept too.
2. **Baseline replay before editing:** `scripts/audit-verify.sh` 105/105 green on the pre-train tree; 106/106 after (the +1 is this train's new verify line).
3. **Four reports refreshed with real drift:** go-idempotency.md (the T18 test-side contract pin recorded — the pin landed 23:33 the same day the report was written; status "indirect-only" was stale in go.mod too: direct require), go-health-dashboard.md (v0.13.3 upstream: the `data-nonce` root bridge, floor = webphone's exact pins), go-filewatcher.md (v2.5.0 `Healthy()`/circuit breaker — premise unchanged), go-credentials-store.md (v0.4.x adminui hardening — premise unchanged). Plus 4 drifted `go.mod:NN` line refs corrected (error-family 24→33, retry 32→41, httputil 31→44, server_timing 32→45).
4. **index.html currency pass** — title stamp, summary paragraph (2026-11-13 block), 4 scorecard rows, the dep-tree line, headline counts re-read (error-family import sites 124→127), footer.
5. **AGENTS.md CURRENCY sentence** updated — and its own staleness fixed (it still listed go-paperless v0.5.2 as "behind"; the 2026-11-12 fleet train had bumped it to latest).
6. **TODO_LIST nonce-lane struck row** got the v0.13.3 fact: the bridge is the SDK-side half of the adjudicated reopen trigger; re-check `csp.go` dropping `'unsafe-eval'` when the bump rides.
7. **Self-inflicted dprint drift caught and fixed** (2 files re-padded; `dprint check` clean after).
8. **Pre-existing cadence red cured by the pinned ritual:** archived 2026-11-12_22-05 (the only report passing `check-rows --strict` + `bare-rows` unmodified — scanned all 66), banner + `git mv` + `scripts/status-index.py` regen, gates green.
9. **Manual push under the bar's first arm** (red main + verified fix in hand, anchor stated): CI **GREEN on 8c2e4f51** via `scripts/ci-verdict.sh` (workflow-name-verified).

## b) PARTIALLY DONE

1. **The untagged-release-commit class was NOT swept.** Currency = tag equality only; no per-repo `git log <tag>..origin/main` check for release commits pushed without tags — the exact class that burned the fleet once (dashboard v0.13.0, tagged by a webphone train because upstream's tag was missing).
2. **Adjacent-family utilization claims not re-read** — go-branded-id/go-flightrecorder/go-paperless/go-sqlitestore got tag verification + audit-verify replay, not a fresh deep read.
3. **Push tail:** the TODO_LIST fact-append (daemon commit e147f435) was still local at snapshot time (origin 8c2e4f51); its CI verdict rides the daemon's push.

## c) NOT STARTED

1. **dashboard v0.13.3 bump execution** — deliberately deferred: dependency bumps ride their own train with the tidy → vendorHash re-pin → CI-watch tail, not a docs sweep. The report + TODO row arm it as zero-collision.
2. **The cadence headroom is spent:** my report consumed the exact one-report headroom the morning's archive bought (count now 65/65, matrix-verified ARCHIVE-HEALTHY) — the NEXT report birth re-reds the gate unless a drain lands first. Candidates analyzed: 2026-11-02_05-29 is owner-answer-gated ("do not entomb"), 2026-11-02_05-19 is trigger-gated — a real per-row evidence pass is owed, not a rushed strike.

## d) TOTALLY FUCKED UP

No user-visible damage, CI ended green — but three honesty defects, all caught before or by gates:

1. **The first tag sweep was structurally wrong.** `awk -F/ '{print $NF}'` over path-scoped submodule tags: cqrs-htmx's "v4.19.0 / v4.20.0" tail was USERMGMT tags, and I briefly concluded the root pin was two releases behind — nearly manufacturing false "behind latest" claims into cqrs-htmx/templ-components/go-output/go-appkit reports. Caught only because I re-queried with explicit refspecs before writing. Same defect family as the 2026-11-02 "grep count as fact" (comments matched tokens): a pipeline shape trusted without looking at its matches.
2. **Self-inflicted gate red:** edited markdown tables without running dprint (rule F is documented in AGENTS — the drift class is known).
3. **Archive attempt in the wrong order:** targeted the gate-named OLDEST report first and only then discovered 13 unstruck rows; the ritual's own wording says gates BEFORE `git mv` — I wasted a cycle going in unprepped (cheap, but sloppy).

## e) WHAT WE SHOULD IMPROVE

1. **Multi-module tag discipline:** per-submodule refspec queries from the first command; better, encode it in a script so the next sweep cannot repeat the mangle (f-item 4).
2. **Baseline the docs battery BEFORE editing** — I attributed format/cadence failures from first principles after the fact; a pre-edit `docs-gates --check` would have made attribution a diff.
3. **dprint fmt immediately after every markdown table edit** — not "when the gate complains".
4. **Stronger verify anchors:** this train's new verify line is RUN-OK class (passes even if the symbol is renamed); `# expect` anchors are the load-bearing shape.
5. **Cadence headroom economics:** the gate counts unarchived reports; an archive sweep that drains exactly to threshold buys ONE report of headroom — the next session's status report re-reds it (observed this session: archive to 64, my report lands at 65/65). Either drains should bank ≥2 reports of headroom or the ratchet structure keeps every sweep ending on the edge. Owner-ruling territory (g-question 2).

## f) Up to 50 things we should get done next

<!-- check-rows:ledger -->

| #  | Task                                                                                                                                                                                   | Impact                                 | Effort | Category      |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------- | ------ | ------------- |
| 1  | go-health-dashboard v0.13.3 zero-collision bump train, and live-test whether `csp.go` can drop `'unsafe-eval'` on a nonce'd host (the `data-nonce` bridge is the SDK-side half)        | Medium (CSP hardening lane)            | S-M    | Dependency    |
| 2  | Honest cadence drain to bank headroom (target ≤63 unarchived): per-row evidence pass over the 2026-11-06/2026-11-11 report backlog (strike canon + strike-safety before each `git mv`) | Medium (next report birth stays green) | M      | Docs-health   |
| 3  | Untagged-release-commit sweep: `git log <tag>..origin/main` across the 20 fleet repos — the "v0.13.0 tag was missing" class tag equality cannot see                                    | Medium (stale-pin alarm class)         | S      | Quality       |
| 4  | Encode per-submodule refspec tag listing as `scripts/fleet-tags.sh` (kills the awk last-component mangle that nearly wrote 4 false "behind latest" claims)                             | Low-Medium (next sweep's honesty)      | S      | Tooling       |
| 5  | catalog/v4 v4.14.0→v4.15.0 bump (documented rider on the next go-cqrs-lite train; dir restructure + hub npm pin)                                                                       | Low                                    | S      | Dependency    |
| 6  | File the go-idempotency injectable-clock upstream ask (the contract suite wall-sleeps; TimingScale 5 is the local band-aid, ci-flake-ledger row run 42407471289)                       | Low                                    | S      | Upstream      |
| 7  | Check go-etag issue #5 (v0.7.0 CHANGELOG section still uncut for entitytag?) and the appkit asks #59/#60/#61 filed 2026-11-12                                                          | Low                                    | S      | Upstream      |
| 8  | Tier 1+2 error-family re-measure (calendar due 2026-12-08; `--max --no-suppress` + both enforce flags at 0)                                                                            | Low                                    | S      | Quality       |
| 9  | Re-measure the DF-brittle pins (annotate_max 203 / audit_max 72) at the next sweep — this session's corpus growth moved the living-doc line count again                                | Low (gate honesty)                     | S      | Docs-health   |
| 10 | Strengthen this train's RUN-OK verify line to an `# expect` anchor shape in go-idempotency.md                                                                                          | Low                                    | S      | Documentation |

## g) Questions I cannot answer myself

1. **Bump priority:** should the v0.13.3 dashboard bump + `'unsafe-eval'`-drop experiment ride NOW as its own small train, or wait for the next coordinated wave / the nonce-lane work? (Both orders are defensible; it is a priority call.)
2. **Ratchet policy:** should the cadence gate exclude fresh reports (e.g. < 7 days old) from the unarchived count, or should drains bank ≥ 2 reports of headroom per sweep? Today's economics: one archive = one report of headroom, so the next status report re-reds main (the corpus was at 66 this morning for exactly that reason). Ratchets are evidence-gated owner territory — I will not move 65 without a ruling.
3. **Archive-selection policy:** I archived the NEWEST gates-passing report (22-05) because it was the only one that passed unmodified; oldest-first would have required striking 13 rows in the 2026-10-25 report first. Which preference should the ritual encode?

---

_HARVEST: section f rows 1-10 are session-grounded candidates; rows 3, 4, 6, 7, 9 extend standing registry entries (TODO_LIST reopen-trigger registry / docs-gates pins). Do not entomb._
