# Status: GitHub issue train — Reports tab, live-call context, CRM private CA (assistant session)

**Snapshot-rot header: claims verified as of 2026-11-06 10:45 CET.** Push/CI/doc-state
claims rot fastest; everything below is timestamped to this minute.

## TL;DR

All five open GitHub issues were worked. **Four are closed** (two were stale red-CI
flags with real verdicts, two were feature issues that shipped). **One stays open** —
`#9`, CRM switch-on — because the code half is done and the rest needs a root-only
secret and the PBX host. One dependabot PR that was failing CI is now green.

| # | Kind | Outcome |
| --- | --- | --- |
| 232 | red-main `41927384287` | **CLOSED** — not a flake: stale generated views + unbaselined clones; healed same night by `97a3656d` + `49aa34bd`. Ledger row added. |
| 233 | daemon-audit `9a30581` (go.mod/go.sum) | **CLOSED** — content was the intended 2026-11-05 templ wave; the landing shape was the defect (daemon carried it with no vendorHash re-pin, red 14 min). Ledger row added. |
| 12 | Reports surface over CDR | **CLOSED — shipped.** New Reports tab: pure projection over the same CDR page. |
| 10 | Live-call context + audible alert | **CLOSED — shipped**, with one premise correction (the ringtone already existed; it just ignored the output picker). |
| 9 | Switch on the Kith CRM | **OPEN** — `crm.ca_file` private-CA lane shipped + `crm` health leg re-verified live; acceptance criteria 1–2 need the CRM token (root-only) and the host. |
| PR 234 | dependabot ginkgo 2.33.1→2.34.0 | **GREEN** — re-pinned the vendored-module hash; all required checks pass on `a2ad2815`. |

## Scope — what this session VERIFIED vs did NOT verify

VERIFIED, with the artifact:

- `go test -count=1 ./...` exit 0 (10:44, final run after the last extraction and
  registry regen) · `tsc -p . --noEmit` exit 0 · island `node --test` 289/289 pass.
- `nix flake check` — **all checks passed** earlier in the session (10:20, including
  both VM tests); re-run after the final file move (see the run id in the follow-ups).
- `python3 scripts/webphone-smoke.py --bin /tmp/webphone-bin` — **51 + 4 + 8 + 9 + 9 + 8
  = 89 checks, 0 failed**, including two NEW probes for the Reports tab.
- Dependency bump PR: `nix build .#webphone.goModules` resolves, `go mod tidy` is a
  no-op beyond the pin, full test suite green on the PR branch; CI run 42188867847
  SUCCESS on `a2ad2815`.
- Live CRM probe against `https://crm.larsartmann.cloud` (real vhost, real private CA):
  with `crm.ca_file=/run/secrets/dnsblockd_ca_cert` and a deliberately wrong token,
  `/healthz` answers 503 with `crm: the bearer token was rejected (status 401)` —
  i.e. TLS through the private CA WORKS and a rejected token reddens the status
  surface naming `crm`. Counterfactual without `crm.ca_file`: also 401, which proves
  **this dev host already trusts the dnsblockd CA** (the PBX host, per the issue, does
  not — which is exactly why the scoped option exists).
- Findings-gate parity: `buildflow -s file-size-check` reports 45 findings on the
  pre-session head `36f0c6f4` (separate worktree) and 45 now — the one transient
  extra (my CRM additions pushing `client.go` 8 lines past the cap) was fixed by
  extracting `internal/crm/transport.go`.

NOT verified, and why:

- **Nothing is pushed.** Remote `main` is still `36f0c6f4` (the session-start head);
  14 commits sit local. See "Push state" — the manual-push bar is met and NOT
  executed, on the owner's explicit "stop now" instruction.
- CI has therefore never seen this train. Every gate above ran locally; the first
  real CI observation is the push (`.githooks`-guarded, so a bump-shaped commit
  cannot slip in unnoticed).
- Issue #9 acceptance criteria 1–2 (inbound call shows the Kith name; `call_logged`
  lands on the journal) — blocked on `/run/secrets/crm_api_token`, which is
  root/unit-scoped by design (`permission denied` to me), and on the PBX host.
- The consuming stack's typed front (`services.telephony.webphone.crm` in
  nix-international-telephony) still renders only `crm.url` + the env-file token, so
  it cannot pass `ca_file` through yet. Deliberately NOT touched (second repo, its own
  gates, and the tri-repo ritual wants webphone pushed first).
- `buildflow`'s final exit is ERROR on findings, not on steps: **71/71 steps passed**;
  the gate counts 45 file-size + 4 docs-gates + 1 cue-vet error-severity findings, all
  PRE-EXISTING (AGENTS tier-3, owner-only). No new finding remains.
- No human eyeball: everything here is machine-verified and self-reviewed only.

## What shipped

### Reports tab (issue #12)

`internal/reports` is a **pure projection** — no storage, no clock, no I/O: the caller
supplies records, zone, "now" and an optional budget. It folds the phone API's CDR
page (up to 500 rows) into totals (calls, talk time, direction in/out, outcome
answered/missed — two independent axes, so an outbound call that never connected is
neither), a per-day talk-time ladder, a 24-bucket hour ladder (busiest-first view), a
contact ranking ordered by talk → volume → number (a TOTAL order, so two runs can
never disagree), and the current ISO week.

New optional knob `reports.weekly_talk_budget` (duration, default off) renders ONE bar
under the week headline, flagged on overspend instead of silently clamped. Bars ride
native `<meter>` because CSP forbids inline `style` — a width-attribute bar would be a
rule we cannot keep.

Degradation: no phone API ⇒ the panel names the missing knob (never a quiet week of
zeros); an unparsable CDR start still counts in the totals and is reported as
"records without a readable timestamp", so the ladders never silently shrink the week.

Sites: `internal/reports/reports.go` (+ `reports_test.go`, 10 specs) ·
`internal/server/reports.go` (+ 5 handler specs) · `internal/web/views/reports.templ` ·
`internal/config` (`Reports` + `validateReports` + a Rejection family pin) ·
routes `/reports`, `/partials/reports` · nav tab · i18n EN+DE · CSS · README/AGENTS/
FEATURES/CHANGELOG/dom-contract · smoke probes.

### Live-call context + ring routing (issue #10)

**Context panel.** `#wp-live-context` in the island aside, rendered ONLY when the CRM
seam is on (zero config stays byte-identical). The shell (`shell.js` §2c2 — a tab→island
affordance, never an island module) reads the live card's new `data-remote` attribute,
asks `GET /partials/live-context?remote=…` once per remote via `htmx.ajax`, and empties
the slot when the last call ends. The panel renders `crmContextItems` — the SAME helper
as the conversation strip, so one contact never has two disagreeing renderings.

The degradation contract is the strictest in the app: seam off, unparsable/non-dialable
remote, no match, no published context, or an unreachable CRM all answer an **empty 200
body**. During a call there is no room for a banner, and the slot is fixed so nothing
shifts. The fragment is **session-gated** despite being "decoration": it resolves with
the server's CRM bearer, so an anonymous caller could otherwise turn a number into a
name, a company and a deal value.

**Ring routing.** The issue's premise that no ringtone exists was stale — `audio.js`
already generates a distinct incoming ring. What was missing is what the acceptance
criterion actually asks: it did not follow the audio output picker, which carried an
explicit *"ring tones stay room-alarms by decision"* ruling. That ruling is now
superseded (owner decision recorded at the site): the picker's device carries BOTH
halves of call audio via `AudioContext.setSinkId` (feature-detected, remembered across
the lazy context creation, so a pick made before the first ring still applies). A
browser without the WebAudio half keeps the room alarm. Found-and-fixed alongside:
choosing "Default" never moved the call audio back (the old code skipped the empty
value) — it now does.

Sites: `internal/server/livecontext.go` (+ 5 specs) · `views/livecontext.templ` ·
`views/crm.templ` (helper extracted) · `layout.templ` (slot + `LiveContext` prop) ·
`shell.js` · `island/app/{calls,audio,audioout}.js` · island specs (+3) · i18n · CSS ·
docs.

### CRM private CA (issue #9, in-repo half)

`crm.ca_file` — an OPTIONAL PEM bundle that extends the SYSTEM roots for the CRM client
ALONE, additively. The alternative the issue proposed (host-wide
`security.pki.trustedCertificateAuthorities`, or a process-wide `SSL_CERT_FILE`) would
vouch for every TLS connection the machine makes afterwards: the OIDC issuer, the
Paperless archive, the ASR provider. A missing or certificate-free bundle refuses the
boot with the `crm.ca_file` code, so a typo names itself instead of degrading
enrichment forever. The NixOS module gained `services.webphone.crm.caFile` (a plain
host path — a CA certificate is public material, so it must NOT ride a credential),
plus eval-time assertions and two module-check cases.

Sites: `internal/crm/{client,transport}.go` (+ `ca_test.go`, real in-test CA + TLS
server) · `internal/config` (+ validation, both test lanes) · `internal/app` ·
`package/crm.nix` · `nix/module-check-crm.nix` · README/AGENTS/CHANGELOG.

## Three findings worth carrying forward

1. **`.gitignore` swallowed a whole package.** The unanchored `reports/` rule (meant
   for the tooling's generated report output) silently ignored `internal/reports/`.
   Every local build and test passed; the sandbox build — which sees only tracked files
   — failed with "no required module provides package". The `nix flake check` caught it,
   which is the entire argument for that gate existing. Rule is now `/reports/`.
   **Open question this raises:** any future package named `docs`, `reports`, `work`,
   `ui-shots`… can vanish the same way. Worth one grep over `.gitignore` for unanchored
   directory rules next session.
2. **Two of the issues were written against a tree older than the code.** #10's "no
   ringtone" and #9's "no CRM health surface" were both already implemented; the real
   gaps were narrower (the picker; the CA/verification). Reading the code before
   building is what turned both from "write a feature" into "close a specific hole".
3. **The `crm` leg gates `/healthz` while the library treats it as non-critical.**
   A CRM outage therefore reds readiness even though the phone keeps working. Nothing
   routes traffic on that endpoint today (unit is `Type=simple`, Caddy fences the
   probes), so it is a pure alarm channel — which is what #9's criterion 3 wants. Flagged
   in the issue so the choice stays explicit if it ever becomes load-bearing.

## Push state — the one thing to do next

- 14 commits are local; `origin/main` = `36f0c6f4` (session start).
- Manual-push bar (ratified 2026-10-11): first unpushed commit `216ed28e` at
  **2026-11-06 09:30 CET**; now 10:45 → **~75 min > the 60-min bar**. The bar is MET.
- Not executed, because the owner's last instruction was "stop now, do not do anything
  else". Nothing was pushed, force-pushed, or reverted.

Exact sequence when you want it:

```console
cd ~/projects/webphone
git log --oneline origin/main..HEAD        # 14 commits, all local gates green
git push origin main
gh run watch $(gh run list --workflow=ci.yml --limit 1 --json databaseId --jq '.[0].databaseId') --exit-status
```

Then, and only then, the stack ritual (`nix-international-telephony` lock bump → tree
clean → pbx-artmann relock) per docs/release-runbook.md — nothing else in this train
requires a stack change, since `crm.caFile` is additive and unused by default.

## Open ends (each has a home, none is lost)

- **#9** stays open with a full status comment: what landed, what was verified live,
  and the exact host-side steps. Closing it would be dishonest while criteria 1–2 are
  unverified.
- **PR #234** is green and unmerged. Merging a dependency bump is an owner call; the
  mandatory tail (tidy → re-pin → watch) is already on the branch.
- **Stack `caFile` pass-through** — needs a change in
  `nix-international-telephony`'s typed CRM front; documented in the #9 comment, not
  attempted.
- **The daemon guard hole** (`core.hooksPath=/var/empty` in the pma unit, 2026-10-15)
  is untouched and still fleet-policy work in the nix-config repo. Today's #233 verdict
  is the second concrete instance of it costing a red main.
- **Findings gate** exits ERROR on 50 pre-existing tier-3 findings. Not this session's
  business, but it means a bare `buildflow` can never be green, which trains readers to
  ignore its exit code — the same failure mode as the retired baseline file.

## f) Up to 50 things we should get done next

_(brainstorm, not a commitment list — most belong in TODO_LIST/ROADMAP via docs-health
HARVEST, not here. Sorted by impact; source in brackets. Written 2026-11-06 11:45 during
the docs-health v8 sweep, which also verified this report's claims.)_

**Now / this week**

1. **Push the 14 local commits** — the manual-push bar is met (first unpushed commit
   `216ed28e` at 09:30; ~75 min at 10:45) and CI has never seen this train. Everything
   else on this list is cheaper once main is green remotely [§ Push state].
2. **Merge PR #234** (ginkgo 2.33.1→2.34.0, green on `a2ad2815`) — the bump is an owner
   call; the tidy → re-pin → watch tail is already on the branch [§ Open ends].
3. **Stack `crm.caFile` pass-through** — `nix-international-telephony`'s typed CRM front
   renders only `crm.url` + the env-file token, so the new option cannot reach the PBX
   host [§ Open ends].
4. **Issue #9 acceptance criteria 1–2** — needs `/run/secrets/crm_api_token` (root/unit
   scoped by design) and the PBX host: an inbound call must show the Kith name and
   `call_logged` must land on the journal [§ Open ends].
5. **Audit `.gitignore` for unanchored directory rules** — `reports/` silently ignored the
   new `internal/reports/` package and only the sandbox build caught it; `docs`, `work`,
   `ui-shots` are the same hazard class [§ Three findings 1].
6. **Decide the `crm` health leg's criticality** — a CRM outage reds `/healthz` readiness
   even though the phone keeps working, while the library treats crm as non-critical.
   Nothing routes traffic on that endpoint today, so this stays an alarm channel until it
   is made explicit [§ Three findings 3].
7. **Findings-gate posture** — `buildflow` cannot exit 0 while 50 pre-existing tier-3
   findings stand, which trains readers to ignore its exit code (the retired-baseline
   failure mode, one layer up) [§ Open ends].
8. **Re-run the hub fan-out benchmark on a quiet host** — the 2026-11-06 attempt (this
   sweep, `load ~22/32`) was inconclusive; its trigger (cqrs-htmx v4.16.0 + go-sse v0.7.0)
   has been fired for two trains [docs/reviews/2026-09-18_hub-fanout-baseline.md].
9. **Store test train** — `internal/store` sits at 64.8%, below the 75.3% the 2026-09-20
   owner-scoping suite recorded; the admin read seams, thread-flag columns and migration
   v5 grew faster than their suites [docs/reviews/2026-09-20_coverage-baseline.md].
10. **Wire the docs-gates battery into CI** — this sweep adopted it locally; the pre-canon
    archive residue is ratcheted, but nothing reds main when it GROWS [docs-health v8].

**Owner-gated, sitting and cheap once answered**

11. Answer the fresh 2026-10-29 owner-decision packet (unlocks 10 planned tasks) [TODO].
12. Ratify the fax carrier legs (`fax.send`/`fax.receive` enablement needs carrier action)
    [TODO Fax enablement].
13. Post the release announcements backlog (v2.1.0–v2.8.0 drafts) [TODO].

**Small, self-contained**

14. `/context` contract snapshot test pinning webphone's parse against Kith's published
    shape.
15. History nudge for the **notes** save path (`call_notes` PUT leaves a second tab stale
    — the F89 shape, different slot).
16. `scripts/templ-generate.sh` wrapper that pins the regen compiler to go.mod's version
    and asserts the generated header stamp (the 2026-11-05 wave's near-miss class).
