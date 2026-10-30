# Status — Fax capability honesty (the Fax tab stopped lying)

**Snapshot:** 2026-10-30 02:25 CET. Claims below were verified at that
moment; push/CI statements rot fastest, re-check with `git ls-remote`
and `gh run list` before quoting them anywhere.

**Session shape:** started as a research question ("what is a Telnyx Fax
Application?"), became an audit of `~/projects/pbx-artmann`, and ended as
a cross-repo honesty train in `webphone` + `nix-international-telephony`.

---

## 0. What the session was actually about

The owner's charge, verbatim in spirit: **fax is disabled, so why does the
UI not say so?** The finding is worth stating plainly because it is the
root of everything below:

- The Fax tab rendered a fully working compose form and the empty state
  *"No faxes yet — send a PDF above."* on a deployment that has **one DID
  (voice, on the SIP FQDN connection) and no outbound fax path** — the
  stack's bridge answers `/gateway/fax` with an intentional `503`.
- `FEATURES.md` declared every fax feature `🟢 FULLY_FUNCTIONAL` with no
  deployment qualifier, and `AGENTS.md` said nothing about the gap.
- Nothing crashed. The lie surfaced only **after** a user did everything
  the UI asked, as a failed job carrying the provider's honest text.
- Two hosts disagree in plain sight: the **stack's own example host**
  `hosts/pbx-prod/default.nix` sets `services.telephony.fax.enable = true`,
  while the **real deployment** (`pbx-artmann/hosts/pbx/default.nix`) sets
  no fax at all. Reading the stack made fax look shipped.

So: not a missing check, a **missing axis**. Fax has two independent
directions (send / receive) and the product modelled neither.

---

## a) FULLY DONE

| # | Work | Evidence |
|---|------|----------|
| 1 | **Config declares fax lanes.** `config.Fax{Send, Receive *bool}` with `SendEnabled()/ReceiveEnabled()/AnyEnabled()`; pointer bools so an unset field = **on** and no forgotten literal can strip the UI. Env `WEBPHONE_FAX__SEND=false`, file `{"fax":{"send":false}}`. | `internal/config/config.go:526`; `TestLoadFaxCapability` (default both on → env turns ONE off → file turns both off) |
| 2 | **The Fax panel never offers what it cannot honour.** Form renders only when `Send`; a `.wp-empty wp-fax-off` notice names the missing half (with the owner ask); capability-aware subtitle and empty state; both-on markup **byte-identical** to pre-change. | `internal/web/views/fax.templ:18,90,121`; `TestFaxPanelDeclaresTheDeploymentCapability` (4 capability states × 2 languages), `TestFaxPanelDefaultMarkupIsUnchanged`, `TestFaxListEmptyStateFollowsTheCapability` |
| 3 | **One conversion, three consumers.** `handlers.faxCapability()` is the only config→view mapping; the panel, the Settings row and the **SSE payload** all read it, so a push can never promise an affordance the render withholds. | `internal/server/panels.go:175`; `internal/server/notifier.go` (constructor carries `views.FaxCapability`, mirroring the `asrOn` precedent); `TestFaxPanelDeclaresTheDeploymentLane`, `TestFaxPanelKeepsSendingOnADeploymentWithoutInbound` |
| 4 | **Settings tells on the deployment.** A `Fax` row appears **only when a direction is missing** and names it (“receive only — no outbound fax provider”, “send only — no fax number is routed here”, “not configured — nothing can be sent or received”). Zero-config panel stays byte-identical. | `internal/web/views/settings.templ:251`; asserted in both the view and the server test |
| 5 | **Stored history survives.** Rows, document links and the timeline still render on a deployment with no lane (a capability change must not eat data). | `TestFaxRowRendersWithoutCapability` |
| 6 | **i18n parity, both languages.** 11 new keys (4 subtitles, 4 empty states, 3 notices) + 4 Settings keys, EN and DE. | `internal/web/views/i18n.go`; the repo's parity + unused-key tests stayed green |
| 7 | **A live, permanent proof.** New smoke scenario boots the real binary twice: env-off (no form, notice, Settings row) and default (form, no notice). | `scripts/webphone-smoke.py:775`; ran clean: **faxcap 7 passed, 0 failed** |
| 8 | **Harness de-lied.** `ui-capture.py` demanded `.wp-compose-fax` “always present”; now accepts the form **or** the notice. (Verified the assertion path is `find_elements("css selector", …)`, so the comma selector is valid.) | `scripts/ui-capture.py:133` |
| 9 | **Stack declares the truth** instead of leaving webphone to guess: `settings.fax = { receive = mkDefault cfg.fax.feed.enable; send = mkDefault false; }` — receive follows the fax **feed** (the only path into the Fax tab), send is off because nothing transmits faxes. | `nix-international-telephony/modules/telephony/web.nix:345` |
| 10 | **Stack pins its own derivation eval-only** (feed on ⇒ receive-only; default fixture ⇒ neither), wired into the existing check loop. | `nix-international-telephony/tests/eval.nix` `faxCheck`; `nix build .#checks.x86_64-linux.telephony-eval` → **exit 0** |
| 11 | **Deployment truth confirmed at eval:** `pbx-prod` now evaluates to `{"receive":false,"send":false}`. | `nix eval --json …services.webphone.settings.fax` |
| 12 | **Documentation tells the same story in four places, each in its own job:** README (config table + a "Fax capabilities (deployment facts)" section incl. the one-connection-per-number fact), FEATURES (code status ≠ able-to-fax + a capability row), CHANGELOG, TODO_LIST (the enablement path), AGENTS (a rule), lessons.md (the war story). | `README.md`, `FEATURES.md`, `CHANGELOG.md`, `TODO_LIST.md:62`, `AGENTS.md:125`, `docs/lessons.md` § Honest UI surfaces |
| 13 | **Gates run and green:** `go test -count=1 ./...`, `tsc -p . --noEmit`, `scripts/format-check.sh` (= `.#checks…format`, exit 0), `templ generate` (0 updates), smoke 49/4/8/9/9 + faxcap 7/7, stack `telephony-eval` exit 0. | command output in session |
| 14 | **JS assumptions checked before claiming safety:** `shell.js` guards its composer lookups with `form && …`, so an absent fax form cannot throw. | `internal/web/assets/shell.js:2361,2374` |

Commits: webphone daemon-committed 6 commits in the window (01:57–02:13),
stack 3+ (latest `87331b2`, 02:14). **Both repos are clean but UNPUSHED.**

---

## b) PARTIALLY DONE

| # | Work | What is missing |
|---|------|-----------------|
| 1 | **Deployment effect.** | The stack pins `github:LarsArtmann/webphone` via `flake.lock`; the pinned revision **predates** this change and `koanf` silently ignores the unknown `fax` key. So `pbx-artmann` will keep serving the old, lying Fax tab until: push webphone → stack lock bump → pbx-artmann relock → switch. None of those happened. |
| 2 | **Visual proof.** | I updated the capture harness's marker but never ran `scripts/ui-capture.py`. Nobody has *seen* the off-state panel; the dashed `.wp-empty` box, its width in the tab column, and the German copy are unverified visually. |
| 3 | **`buildflow`.** | Not run at train start **or** end (AGENTS calls for both). Lint tiers, duplication, blog-gate, docs gates (AGENTS size cap OK at 188 lines, but `docs-numbers`-class checks are the stack's), vulnerability lane: unknown state for this train. |
| 4 | **`nix flake check`.** | Only `format`, `island-typecheck` (via `tsc`) and the Go/JS suites were exercised. **Not** run: `island-lint` (oxlint), `island-js` (node tests in the sandbox), `webphone-module` (module eval with stand-ins), the KVM-gated backup VM test, `vulnix`. |
| 5 | **Stack's own flake check.** | Only `telephony-eval` was built. `webphone-module`, VM suites, treefmt of the stack repo (I formatted just the two files I touched) are unverified. |
| 6 | **DC check of the derived value.** | I made `fax.receive` follow `fax.feed.enable`, but did not prove what a *fully* wired fax deployment looks like end-to-end (rxfax → TIFF → feed → `/hooks/fax` → tab). That wiring predates me and remains unexercised. |
| 7 | **Email/DID side.** | The one-connection-per-number constraint is asserted in docs from vendor reading; I did **not** verify it against the live Telnyx account (i.e. that `+1727289311` is bound to the FQDN connection and cannot carry a fax app). |

---

## c) NOT STARTED

1. Push of either repo (deliberate: I do not push unprompted).
2. Stack lock bump + pbx-artmann relock/deploy (the release-runbook ritual).
3. Buying/provisioning a dedicated fax DID; setting `gateways.<gw>.faxDid`.
4. Implementing outbound fax anywhere (bridge `/fax`, or a Telnyx Programmable Fax app + fax-enabled number).
5. Hiding/removing the Fax tab from navigation when both lanes are off (I chose *visible with history*; not confirmed with the owner).
6. Retention interaction: if `both off`, do old fax documents still need `retention_days` coverage? (They do get swept today — unchanged, but unreviewed in this new light.)
7. Visual acceptance pass (ui-capture, both themes, EN+DE, off+on states).
8. `docs/dom-contract.md` note: the fax compose form is **conditional** now — it is not in the pinned id block, but the stack's browser E2E (which I could only grep, not run) should be re-run after any markup change per the release-runbook obligation.
9. Reconciler support for Telnyx fax applications (`/fax_applications` in `scripts/telnyx_reconcile.py` + `telnyx/desired.json`) — only needed if the answer to Q3 is "Programmable Fax".
10. A stack-side option for outbound fax (today: flip `settings.fax.send` by hand in a host).
11. Deleting the stale `pbx-artmann/.pytest_cache` node id referencing `tests/test_telnyx_bridge.py`, which now lives in the stack repo.
12. Any owner-facing announcement ("fax is not available here, here is why") beyond the in-app copy.

---

## d) TOTALLY FUCKED UP (mine, then fixed — and the ones that are pre-existing)

**Mine during this session:**

1. **First design was wrong and I nearly shipped it.** I modelled `fax.send/receive` as plain bools with `defaults() = true`. Consequences: every config **literal** that forgot the field (exactly what the server test harness does) silently rendered a fax-less UI — two tests went red and the *shape* of the bug was precisely "the UI quietly claims no fax". I inverted it to **pointer bools where unset = on**; the zero value is now the safe, complete product. The red tests were the honest signal, not noise — worth remembering that the first failing run was *right*.
2. **Sloppy first `fax.templ` edit.** I wrote a truncated comment ("…is a lie on") and a stuttering paragraph, then had to re-edit. A five-second re-read before saving would have avoided it.
3. **Scope drift by design, priced badly.** I answered a *research* question, then audited a sibling repo, then changed two repos in one sitting. The finding justified it, but I never stopped to ask whether the owner wanted the honesty fix in the same breath as the answer — and the push/lock-bump chain is now half-open because of that.
4. **Claimed "one step left" without re-verifying drift.** I said the daemon push "stalled ~3h45m"; that number came from a single measurement and I did not re-check before writing it into the wrap-up.

**Pre-existing, surfaced by this session (not caused by it):**

5. **The product lied for weeks.** A complete-looking Fax tab on a host that cannot fax, plus `FEATURES.md` `🟢 FULLY_FUNCTIONAL` with no deployment qualifier. The honesty contract in AGENTS ("the UI never shows a state the network hasn't confirmed") had no analogue for *capabilities* — only for *confirmations*.
6. **Two hosts disagree.** Stack example host enables fax; the real host does not. Nothing compares them, so the "it's shipped" belief was cheap to hold.
7. **Unknown config keys are silently ignored.** `koanf` accepts `settings.fax.*` from a newer stack into an older binary with no warning, so a config/version mismatch is invisible. That is exactly how a "declared" capability can evaporate.
8. **No test asserted the deployment's ability to fax** — the whole suite tested the code path with a fake provider, which is correct but was read (by humans) as ability.

---

## e) WHAT WE SHOULD IMPROVE

1. **Every optional capability should have a declaration seam like this one**, not a discovered-at-503 story. Candidates that are structurally identical: Paperless archive, CRM, ASR, push. ASR/push *do* gate their affordances; Paperless and CRM only hide rows — no one can tell from the UI whether they work.
2. **Make the capability one type, not one string.** Today `views.FaxCapability` is booleans; the same idea for other seams would be a small `views.Capability` (on/off/degraded). Careful: do not generalise before the second real user (YAGNI) — but note the shape.
3. **Config keys the binary does not know should warn, not vanish.** A one-line boot warning listing unknown keys under `settings.*` would have made the pin mismatch loud instead of silent.
4. **A "deployment facts" contract test in the stack** that compares the *example host* against the *real host* (`pbx-artmann`) for the lanes that matter (fax, recording, dashboard). Their disagreement is what seeded the false belief.
5. **Push liveness needs an owner-visible signal.** The manual-push bar exists; the stall was invisible. `scripts/push-stall-watch.sh` exists for exactly this and was not scheduled/watched.
6. **Write the copy once, render everywhere.** The off-state copy currently lives in 11 i18n keys. A second seam will want the same shapes; consider `views.capabilityNotice(missing []string)` when the second consumer appears.
7. **Run `buildflow` at train start**, as AGENTS demands — I skipped it and therefore cannot state the lint/dup baseline for this train.
8. **Test the *pin*, not just the code.** A flake-level check that the stack's `webphone` input is newer than the config keys it writes would have caught (1) above without a human.
9. **Prefer "state the consequence" copy over "not configured".** "Sending is not configured on this server: no outbound fax provider is wired. Received faxes still appear below. Ask the owner…" is close; the DE version should be read by a native speaker before shipping.
10. **Small cleanups noticed:** the stack's `hosts/pbx-prod/default.nix` fax comment says "provider must send that number to the PBX" while `fax.enable = true` with no `faxDid` — a config that enables rxfax but routes nothing; the pbx-artmann pytest cache in the wrong repo; `AGENTS.md` grew a rule but the *capability* doctrine deserves one home rather than a new bullet per seam.

---

## f) Up to 50 things to get done next

**Ship the change (1–6)**
1. Push webphone `main` (manual-push bar met: first unpushed commit 2026-10-29T22:31, >60 min).
2. Check CI on that push (`gh run list`) and watch to green.
3. Stack: `nix flake update webphone` (per-train lock bump), rebuild `telephony-eval` + `webphone-module`.
4. pbx-artmann: relock to the new stack rev, `nixos-rebuild switch` on the host.
5. Verify live: `/fax` on the deployed host shows the notice and **no** compose form; `/settings` shows the fax row.
6. Verify live: send attempt is impossible from the UI (no form) — then spoil it deliberately via `/fax/send` to confirm the 503 path still reads honestly.

**Verify what I did not (7–13)**
7. Run `buildflow` (full, `BUILDFLOW_NO_RESULT_CACHE=1`) and record findings.
8. Run `nix flake check` on webphone (island-lint, island-js, webphone-module, backup VM if KVM is free).
9. Run the stack's full `nix flake check` (or at least `webphone-module` + VM suites).
10. Run `scripts/ui-capture.py` (both themes) and eyeball `07-fax-light` / `20-fax-dark`; add an off-state capture (boot with `WEBPHONE_FAX__SEND=false`).
11. Re-run the stack's `tests/browser-e2e.py` against the deployment after the markup change (release-runbook obligation).
12. Native-speaker read of the German fax copy (4 subtitles, 4 empty states, 3 notices, 4 Settings strings).
13. Confirm on the Telnyx account whether `+1727289311` is bound to the FQDN connection and what a fax-enabled DID would cost.

**Close the honesty gaps in other seams (14–21)**
14. Audit Paperless: can the UI tell an operator whether inbound archiving works? (Today: no.)
15. Audit CRM: the Settings row is a bounded probe — is "connected" enough, or should it say *what* it is connected to?
16. Audit ASR: does the Settings row state that transcription is *local* vs *remote* everywhere it appears?
17. Audit push: `/api/push-devices` 404 is read as "cannot ring closed apps" — is that visible when it matters (Settings) or only in logs?
18. Decide whether a capability declaration belongs in `/config.js` for the island (today the island learns nothing about fax).
19. Add an unknown-config-key boot warning (koanf leftovers) naming the keys.
20. Add a stack eval check comparing example-host vs real-host lanes (or delete the example host's fax enablement if it is aspirational).
21. Write the "capability declaration" doctrine into exactly one doc (probably AGENTS § Failure → feedback) and point the per-seam rules at it.

**Enable fax, if the owner wants it (22–30)**
22. Owner decision: dedicated fax DID yes/no, which country/number type.
23. Buy/port the DID; confirm fax capability with the carrier.
24. Set `services.telephony.gateways.telnyx.faxDid` + `fax.feed.enable` + `feed.secretFile` + `feed.from = <faxDid>`.
25. Prove inbound: send a real fax, find the TIFF under `/var/lib/telephony/recordings/fax`, then the PDF in the Fax tab.
26. Decide outbound: bridge implementation of `/fax` vs Telnyx Programmable Fax app.
27. If Programmable Fax: extend `scripts/telnyx_reconcile.py` + `telnyx/desired.json` with `fax_application_name` / `fax_did` / OVP attachment.
28. Implement `/gateway/fax` in `telnyx-webhooks.py` (multipart → `POST /faxes`; webhook → `/hooks/fax/status`).
29. Flip `settings.fax.send = true` in the host only after 28 is green.
30. Add a smoke scenario for a real outbound fax against the staged bridge (avoid a webhook-lane smoke per the decided-against rule; stage it as a VM/eval test instead).

**Robustness / tests (31–38)**
31. Test `fax.receive=false` + a stored inbound fax (history keeps rendering, notice still names the missing lane).
32. Test the SSE push payload for a fax-off deployment (a fake inbound event → empty state copy, not "send a PDF above").
33. Property test: every `views.FaxCapability` combination produces exactly one notice or none.
34. Add a `TestSettingsFaxRowAbsentWhenBothLanesOn` (byte-parity guard for the Settings panel).
35. Add a docpin entry for the two new config keys in the *example* JSON if one exists (docpin already passes; verify the example block).
36. Consider a `config.Fax.Validate()` rejection for `fax.send=true` while `gateway.mode=loopback` outside dev? (Probably NOT — loopback is the documented demo. Record the ruling.)
37. Lint: run `erraudit` to be sure the new constructors did not add stale nolints (none added, but the 41-of-109 decay lesson).
38. Coverage read: did `faxCapability()` and `faxUnavailableKey` get covered on all branches? (View tests suggest yes; confirm with `go test -cover`.)

**Operations / process (39–46)**
39. Schedule `scripts/push-stall-watch.sh` on a local timer (it is LOCAL-only by design).
40. Clean `pbx-artmann/.pytest_cache` and `.mypy_cache` stale artifacts referencing moved tests.
41. Update `docs/release-runbook.md` with the "pin predates config key" trap and the verify step (`/fax` notice).
42. Add the fax notification to the owner-facing runbook (`docs/runbooks/`) so the next operator does not re-discover it.
43. Reconcile the stack's `hosts/pbx-prod/default.nix` fax block with reality (either wire `faxDid` or stop enabling fax in the example).
44. Decide whether `fax.enable` without any `faxDid` should be an eval assertion (it currently enables rxfax nothing routes to).
45. Re-check the AGENTS size cap after the new rule (188/377 — fine, but the doctrine bullet should replace, not accrete, next time).
46. Add the faxcap scenario to any CI job that runs the smoke (verify where the smoke runs; today it is manual/local).

**Optional polish (47–50)**
47. Nicer off-state affordance: a single inline link to Settings instead of only prose? (Design call.)
48. Show the fax lanes in Settings *even when on*, as "enabled" — for symmetry with the gateway row? (Tension with byte-parity; needs an owner call.)
49. Consider renaming `settings.gateway` from "Message & fax gateway" now that fax is declared separately (it is currently a small lie in the other direction).
50. Add an entry to `docs/dom-contract.md` § Notes naming the conditional `.wp-compose-fax` / `.wp-fax-off` pair so the stack E2E never assumes the form.

---

## g) Three questions I CANNOT answer myself

1. **Should the Fax tab exist at all when both lanes are off?** I kept it visible so stored faxes and their documents stay reachable, and made it say "Fax is not available on this server". The alternative is hiding the tab and (perhaps) moving history behind Settings → export. This is a product/UX call with data-access consequences — I will implement either.
2. **Push and relock now, or leave it to the daemon/train?** The manual-push bar is met (first unpushed commit >60 min) and the deployed effect needs webphone push → stack lock bump → pbx-artmann relock → switch. I did not push because you have not asked me to. Say the word and I will run the whole chain and CI-watch it to green.
3. **Is fax a product you want, or a feature you want gone?** If yes: dedicated fax DID (which number/country, and who pays) plus an outbound decision (bridge `/fax` implementation vs a Telnyx Programmable Fax application on a *separate* number). If no: I can delete the Fax tab, the stack's fax options, and the bridge's 503 in one train and keep only the export path. Both are clean outcomes; the current middle state is the only one that costs attention forever.

---

## UPDATE — 2026-10-30 14:30 CET (the follow-through session)

Snapshot-rot warning applies: timestamps below are per-claim; re-check
push/CI before quoting.

### The three open questions (§g) — ANSWERED by the owner (~04:15 CET)

1. **Fax tab visibility** → KEEP the tab (history/downloads live there).
2. **Push/relock** → executed (see the chain below; the daemon's
   push-leg stalled 9h20m — the manual-push bar was invoked with anchor
   04:34:43+01:00, and SystemNix gained an hourly local watchdog).
3. **Fax future** → KEEP as-is, declared off. No DID purchase, no
   outbound path, no deletion. Full record + consequences:
   docs/planning/2026-10-30_04-15_fax-decision.md. Settings lanes go
   always-on (owner decision 2), shipped in this session.

### Gates that never ran at 02:25 — ALL RUN NOW

- `buildflow` full (`BUILDFLOW_NO_RESULT_CACHE=1`): exit 0; warnings
  only, all pre-existing classes (vulnix toolchain advisories, vulture
  false positives on ui-capture's HTTP handler methods, the retired
  baseline-file notice, a go.sum drift that `go mod tidy` then healed —
  vendorHash verified by `nix build .#webphone.goModules --rebuild`).
- webphone island gates: island-lint, island-js, island-typecheck,
  webphone-module — all green.
- `ui-capture.py`: 24-shot matrix green (parity path intact) AND the new
  `--fax-off` mode: 4 shots (fax + settings, light+dark) with the
  TIGHTENED assertion — notice present, compose form ABSENT. The
  off-state panel reads informational, not error.
- erraudit nolint-audit: 68 directives, 68 needed, 0 stale.
- Smoke on a fresh HEAD binary: 49 + restart 4 + boot 8 + asr 9 +
  dialgate 9 + faxcap 8 (new: Settings states the available lanes).

### The verification gaps — CLOSED

- **Unknown config keys now WARN at boot** (`internal/config/unknown.go`):
  one WARN names the unknown key paths (never values; capped at 20 with
  the full count), zero unknown keys log nothing. Catches the exact
  silent-pin-gap that kept this deployment lying: a config newer than
  its binary names what the binary ignores. Found and fixed a real wart
  while testing it: WEBPHONE_CONFIG itself used to leak into koanf as a
  junk `config` key (the warning would have fired on EVERY boot).
- **Stack eval guard**: `fax.enable` with a PSTN trunk but no
  `gateways.<n>.faxDid` now FAILS EVAL (negative fixture in
  tests/eval.nix; a gateway-less lab node stays legal — its fax
  extension is SIP-reachable, not a PSTN lie). The example host
  (hosts/pbx-prod) flipped `fax.enable = true → false` with the real
  enablement recipe in a comment: the example can no longer disagree
  with the truth.
- **CI flake fixed** (tests/fax.nix): the `mod_spandsp` assertion waits
  (wait_until_succeeds, 60s) instead of probing once — CI run
  40986572102 failed it on a loaded shared runner while the
  byte-identical derivation passed locally with KVM.

### Owner decisions shipped (Settings symmetry)

The fax Settings row is now UNCONDITIONAL and states both lanes in all
four combinations ("sending and receiving available" / the named
missing half with its reason), pill class following whether any lane is
up. This deliberately ends the Settings byte-parity era — pinned by
`TestSettingsFaxRowAlwaysStatesTheLanes` + the smoke's new faxcap
assertion. DE copy got a best-effort non-native pass (one precision
fix: "Der Faxversand ist …"); the record says a native read remains the
real bar.

### Deploy chain state (as of this snapshot)

- webphone `main` = `9dc9b15a`, CI green (run 41058944754).
- stack `main` = `ad122b9` (faxDid assertion scoped to trunk-hosts after
  it honestly red-ed the gateway-less VM fixtures in run 41060217586;
  refined + locally proven: telephony-eval AND telephony-fax green).
  CI on `ad122b9`: in progress at snapshot time — CHECK BEFORE RELOCK.
- pbx-artmann: **RELOCKED, STAGED, GUARDS GREEN, PUSHED (15:5x CET)**.
  `nix run .#relock` moved telephony `0a5ac22 → ad122b9` (stack CI green,
  run 41062742902) and the webphone input `b43d1ab → b035e74e` (the
  sibling's vetted rev). Real-host eval verdict:
  `{"receive":false,"send":false}`. The staging ritual survived two
  guard stops, both instructive: (1) `/tmp/pbx-toplevel-current` was
  wiped by a reboot — rebuilt from the DEPLOYED commit `376db4a` via a
  git worktree and the store path came back BYTE-IDENTICAL to the live
  closure (`yyib5k9g…`, same derivation = same path, the repair recipe
  works); (2) the concurrent mail-train commit (`0fce4f0`) landed
  mid-ritual, so the staged root carries BOTH trains and the CHANGELOG
  baseline bullet names the combined diff story (docs-churn restages
  are same-plan children; readlink is the truth). Final state: staged
  root `dw8zhd79…`, deploy-freshness FRESH, secrets-preflight OK,
  remote master == local (`1245583`). Switch sheet:
  `pbx-artmann/docs/runbooks/owner-switch-2026-10-30-fax-truth.md`.
  The deploy app's printed owner legs (no-ssh doctrine — the assistant
  never ssh'd the PBX):
  ```console
  0. push the train (DONE — remote master 1245583)
  1. push secrets if any changed (NONE changed in this train)
  2. nixos-rebuild switch --flake .#pbx --target-host root@pbx.artmann.tech
  3. ssh root@pbx.artmann.tech 'systemctl restart telephony-secrets-perms webphone telephony-agent'
  4. ./tests/installer/verify-live.sh
  ```
  (The sheet's test→verify→switch order is the conservative superset;
  the printed legs are the app's canonical four.)

### RIDE COMPLETE — 18:2x CET (the staged root carries today's full webphone)

The ride-first recommendation was EXECUTED: stack relocked
`62b52069 -> e6fb14db` (the intermediate bump's CI was cancelled by the
Dependabot github_actions race; the pin rode straight to the settled
tip), stack CI GREEN on `f64f4fd`, pbx-artmann relocked and RESTAGED —
final staged root `a2slrp3c…`, ALL GUARDS GREEN, pushed (`cc1e3f3`).
The staged closure now carries: the fax-truth ride + always-on Settings
row + Message-gateway rename + T20 tests + the Paperless archive
Settings row + the smoke-in-CI lane + the concurrent mail-fleet train.
Also shipped while the CI cooked: `nix flake check` FULL green (T5,
including the KVM backup VM suite), the smoke-in-CI job (T6, first
green run on `e6fb14db`; the first attempt died on the missing `trash`
binary — the cleanup now degrades to shutil.rmtree for the script's own
scratch), and the Paperless capability row (T7, conditional, EN/DE,
tested). Remaining: the owner's four switch legs.

### Other attributed findings

- Stack CI's two "runner lost" reds (10-29) and one cancel: infra, not
  code; the runs around them were green.
- SystemNix `btrfs-scrub-staleness-fixture` red: pre-existing since
  530ff33ee (10-28), unrelated to this train's push-stall timer (which
  evals clean); attributed, not fixed — wrong repo, wrong train.
- Paperless is the next silent-seam candidate (capability-seam audit:
  docs/planning/2026-10-30_14-05_capability-seam-audit.md) — recorded,
  NOT built (plan §8: findings first).

→ **Continued 20:29 CET:** the pre-switch sweep + owner-sheet repair
(18:3x snapshot) is
`2026-10-30_20-29_fax-truth-pre-switch-sweep-and-owner-sheet-repair.md`
— the owner sheet had gone stale after the second ride (it still
demanded relock `05a442e` on top) and was fixed (`d540f4b` in
pbx-artmann); the switch was still pending as of that report.
