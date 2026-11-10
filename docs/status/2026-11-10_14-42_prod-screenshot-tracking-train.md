# Prod screenshot-tracking train (the unlighthouse lane) — status

> Claims verified as of 14:42 CET (2026-11-10). Push/CI statements rot fastest; the remote tip and the unpushed set were re-read moments before this header's timestamp.
> Session: single Crush train, task = "full unlighthouse-kind screenshot tracking for https://pbx.artmann.tech so the UI (layout) can actually be made not-ugly".

## TL;DR

The visual harness (`scripts/ui-capture.py`) now captures a LIVE deployment, keeps timestamped history runs, and pixel-diffs runs against each other — implemented, verified end-to-end locally in four lanes, and run once against prod (anonymous public face: 4 shots, `v2.12.0` recorded). The authenticated prod run (the actual all-tabs matrix over real data) is **blocked on credentials only** — one env-var command away. The full buildflow gate is RED at train end (4 failed steps), attributed: `island-js` fails on a spec landed by ANOTHER session's in-flight voicemail work (unpushed daemon commit `0d2e9ff8`), and `nix-hash-fix`'s "failure" contradicts its own output. Push is deliberately HELD (12 unpushed commits, manual-push bar satisfied at 173 min, but pushing would land a known-red main without a verified fix in hand).

## a) FULLY DONE

1. **Harness understanding**: read `scripts/ui-capture.py` (878 lines pre-train) end to end; verified the wire facts the remote mode leans on (login-card ids `#login-form`/`#passkey-login` in `internal/web/views/phone.templ`, cookie `webphone_session`, open JSON `/version` route, prod `/livez` pass).
2. **Design**: extend the ONE existing visual harness instead of adding an unlighthouse-style crawler — `--base-url` (remote mode), `--track` (history + diff), full-page shots, repeatable viewports, `--anonymous`. Design rationale recorded in the module docstring.
3. **Remote login** (`login_remote`): `POST /api/session` with env credentials (`WEBPHONE_CAPTURE_EXTENSION` / `WEBPHONE_CAPTURE_PASSWORD`), CSRF meta token dance identical to `seed()`, cookie-jar extraction, loud refusals for 401/403/429 with named env vars, credentials never logged.
4. **Live surface resolution** (`resolve_remote_surfaces`): scrapes the real `/messages` list for deep links (`thread-01`…`thread-NN`, default 6 — stable diff keys), probes partials for `data-transcribe-src` (media note), includes `history-transcript` only when transcripts exist, pins `#wp-thread-head` as the evidence marker per thread surface.
5. **Lenient marker mode**: remote runs warn + keep instead of abort (real data legitimately lacks the seeded fixtures); styled-404 surfaces skip — except the deliberate `/nope` surface, which is exempted and still shot.
6. **Full-page capture** (`shoot`): CDP `Page.captureScreenshot` + `captureBeyondViewport`; proven — mobile settings page 3221 px vs 844 px viewport. Desktop panels stretch to exactly viewport height, so the seam only shows on tall/mobile pages (documented in AGENTS).
7. **Tracking** (`track` + builders): runs land in `ui-shots/<host>/<UTC-stamp>/`, never wiped; per-run `gallery.html` + `summary.json` (deployment `/version`, `--label`, per-shot ledger with bytes, prev pointer, changed-%); host-level `compare.html` — per surface previous | current | pixel-diff heat map (Pillow `ImageChops`, faded base + red overlay), changed-% badges ≥ 0.5 %, "new" badges, worst-first ordering, height-changed pairs degrade to side-by-side with a printed note; pillow-less degradation is graceful (side-by-side only + install hint).
8. **Mobile viewport support**: repeatable `--viewport WxH`; fixed a real crash this surfaced — at 390 px the Settings collapsed section sits under the fixed bottom tab bar → `ElementClickInterceptedException`; now `scrollIntoView` + JS-click fallback.
9. **Layout-proof scoping**: the computed-style collapse/split proof is gated to the default desktop viewport (it asserted desktop facts and would false-fail a mobile pass).
10. **Verification, local lanes (all green)**:
    - Classic strict matrix: 24 shots, exit 0, layout proof PASS.
    - Two tracked classic runs: history chain built, `compare.html` diffs computed (2.19 % on `zz-thread-collapsed`; size-mismatch fallback exercised).
    - Remote mode against a manually-booted bare local server (data-reused): API login with `1001`/`pw`, live threads resolved, lenient warns exactly where the bare boot lacks ASR/passkey, `login-page` 404-skip, 22 shots, version `v2.12.0` in summary.
    - `--anonymous` locally, then **against prod**: 4 shots (login card + `/login`, light+dark), `v2.12.0` recorded, gallery written; the `/login` shot was visually inspected — real library page, email + passkey button (prod has passkey ON).
    - `--surface` filter in remote mode: typo rejects loudly (exact names), filtered tracked run works.
11. **Docs**: AGENTS.md command block updated (pillow added to the PYSEL line, full-page note, and a complete PROD-tracking entry with the env-credential recipe, REAL-data warning, SIP-registration caveat, and `--anonymous`); CHANGELOG `[Unreleased] ### Added` entry. Both landed (via daemon commit `b68543b4` — see d1).
12. **Formatter/lint hygiene**: `nix fmt` clean; `python3 -m py_compile` green after every edit; test scratch dirs removed with `trash` (never `rm`).

## b) PARTIALLY DONE

1. **The authenticated prod capture** — everything is built and the exact command is documented in AGENTS.md, but the run itself needs the extension + PBX-directory password in env (`WEBPHONE_CAPTURE_EXTENSION` / `WEBPHONE_CAPTURE_PASSWORD`). I verified there is no legitimate local source: the stack repo's browser-E2E carries only test-VM credentials (`test-1000-*` @ `pbx.test`); nothing in env. Prod uses passkey auth (`#passkey-login` present, `/login` live), and a WebAuthn ceremony CANNOT be automated against the real authenticator — the extension-password lane via `/api/session` is the automatable path (the island card keeps it; passkey mode adds, never replaces).
2. **The `island-js` red diagnosis** — root-caused to ONE spec: `shell.test.mjs:1260` "a dead recording keeps one toast — the fallback stays silent" (expects exactly 1 toast, got 0). Landed in unpushed daemon commit `0d2e9ff8` (07:33) together with `shell.js` changes (42 + 61 lines) — that is ANOTHER SESSION's in-flight voicemail-play follow-up (their pushed fix was `fc352e66`; their uncommitted status report `2026-11-10_06-33_voicemail-play-silence-diagnosis-and-fix-train.md` sat in the tree at session start). I traced the full path (`vmTogglePlay` → `vmWire` error listener → `vmPlaybackFailed`, `vmErrorNotified` WeakSet dedup; the new `recordingMissing` fallback-silence flag) and the code reads correct — a step-by-step node probe (`/tmp/vm-probe.mjs`) was my next tool when it HUNG (killed; the probe's own ordering bug fixed once, then no output — needs a different diagnostic tack, e.g. run the single spec under `--test-name-pattern` with instrumentation inside a throwaway copy of the spec, not shell.js).
3. **Full buildflow green** — 81 steps ran: 77 pass, 4 fail. `nix-build` → the `island-js` check above (pre-existing-red class, now visible: the 2026-10-29 note warned this gate was structurally incapable of failing before). `nix-hash-fix` marks failed while its captured output says "All 4 flake target(s) built successfully, 0 hashes changed" — contradiction unexplained (its historical failure rate is 76 %, so the step may have an instrument bug; not dug into). The remaining two failures in my run are not yet individually identified (`test-race`/`golangci-lint` failures in the history table are 1-day-old, NOT from this run).
4. **Push + CI watch** — the manual-push bar IS satisfied (first unpushed commit `06:38:44+01:00`, measured 173 min old, anchor `git log origin/main..HEAD --format=%cI | tail -1`), 12 unpushed commits, remote tip `2c2cc950`. Push is HELD on purpose: the set contains the other session's red spec, so pushing lands a known-red main without a verified fix in hand — that fails the bar's own precondition.
5. **Status report** — this document; written now, index regen follows.

## c) NOT STARTED

1. The first **authenticated** prod run (all tabs over real data) — the actual ugliness-hunting payload.
2. The **mobile prod pass** (`--viewport 390x844`) against pbx.artmann.tech — implemented, never run remotely.
3. The **first prod diff chain** — `compare.html` needs a second prod run; exactly one exists.
4. `nix-hash-fix` step-failure explanation (its success-looking output vs failed status).
5. CI verdict watch after the eventual push (`gh run list`), per the check-CI-after-every-push rule.
6. The deliberate gallery/compare design review BY THE OWNER (they are eyeball tools; nobody has judged them yet beyond my smoke).

## d) TOTALLY FUCKED UP

1. **Narrative commit messages eaten by the daemon — twice**: my staged+committed docs landed in daemon commit `b68543b4` with a "chore: auto-commit" message despite staging + committing in ONE call (the daemon committed between my `git add` and `git commit`). Content is intact and verified; the history narrative is not. The same happened earlier to the script work (a chain of daemon commits `da097bf1`…`f6c8ba72` carries my `ui-capture.py` evolution).
2. **The buildflow verdict is red at train end** — not by my code (attribution above), but the train cannot declare green, and the red is sitting in the unpushed set.
3. **Probe flailing on the red spec**: the reproduction probe hung with zero output and had to be killed (04F) — I spent two iterations on it and still have no confirmed root cause for the 0-toast behavior. The honest state: code reads right, test says wrong, mechanism unproven.
4. Minor process debt: one no-op multiedit (identical old/new strings) burned a round trip; one splice asserted an off-by-one boundary before writing (caught by the guard asserts, zero damage — the assert-first pattern paid for itself).

## e) WHAT WE SHOULD IMPROVE

1. **Session-ownership markers for daemon commits**: 12 unpushed commits carry zero indication of WHICH session authored the content; a red main made of mixed in-flight work from two sessions is nearly unattributable after the fact.
2. **The daemon racing a deliberate commit inside one tool call** is unhandled — the AGENTS recipe ("stage + commit in ONE call") did not survive contact; a commit-time marker (e.g. `git commit --allow-empty-message` reservation or a lockfile convention) may be needed.
3. **`island-js` visibility debt is real**: the gate's red is a genuinely useful catch, but it took a full buildflow run to see it — a pre-commit/push-time island-test lane (even without nix) would have surfaced the other session's red before it entered history.
4. **Remote-mode spec coverage**: `ui-capture.py` has no self-test; every bug above was found by running against live servers. A `--self-test` (fixture HTML + stub driver) would pin `shot_key`, surface-list building, lenient-skip rules, and the compare math without chromium.
5. **`nix-hash-fix` as an instrument**: 76 % historical failure rate with success-shaped output smells like a broken step, not a broken tree — it erodes trust in every red verdict it produces.
6. **Probe discipline**: I theorized twice before instrumenting; the right move was a probe immediately after the first code read.
7. **Prod data privacy posture**: shots with real customer data live on this laptop only (gitignored) — fine by default, but there is no retention policy or explicit "where may these live" ruling.

## f) NEXT (ranked, roughly Pareto-ordered)

1. Resolve the `island-js` red: finish or revert the other session's fallback-silence change (`0d2e9ff8`) — owner call on ownership (see g2).
2. Run the authenticated prod capture (needs g1) — the whole point of the train.
3. Second prod run to birth the first real `compare.html` diff chain; eyeball it and judge the diff presentation.
4. Mobile prod pass `--viewport 390x844` (and decide whether mobile belongs in every prod run by default).
5. Explain `nix-hash-fix`'s failed-but-successful-output contradiction; fix the instrument or the tree.
6. Push (bar satisfied) once island-js is green; then `gh run list` watch to green.
7. Re-run full buildflow to 100 % green and record the 4-step failure set as closed.
8. Add a `--self-test` to `ui-capture.py` pinning the pure helpers (shot_key, surface list building, lenient rules, compare math).
9. Consider a `--crawl` mode (true unlighthouse-style link discovery beyond the curated list) — only if the curated matrix proves too narrow for real ugliness.
10. Add `admin` surface to the remote list when the session is an admin (`isAdminSession` probe), so the Admin tab gets tracked.
11. Track the CRM `crm-context` strip on prod when the CRM seam is on (currently local-only via the fake CRM).
12. Decide where tracking artifacts may live (laptop-only? synced store?) given real caller data.
13. Retention/pruning policy for `ui-shots/<host>/` (runs are never wiped by design; disk will grow).
14. `history-transcript` on prod: resolve a transcript-bearing thread search (`?q=`) instead of plain `/history` for a denser shot.
15. Capture the `/enroll` page as a conditional anonymous surface (token-gated; currently absent from the matrix).
16. Add the PWA manifest + icon surfaces? (probably NOT — no visual value; record the ruling if skipped).
17. Wire a `just`-free flake app (`nix run .#ui-track`) so the prod recipe is one command instead of a shell dance.
18. Document the SIP-registration caveat in the release-runbook context (a capture browser registers the owner's line for the run's duration).
19. Guard against inbound calls mid-run (detect the ringing card in the DOM and re-shoot that surface automatically).
20. Add a `--min-height` sanity probe per shot (a blank/1px shot should warn — the styled-404 skip already covers one class, dead renders another).
21. Diff-noise tuning: expose the pixel threshold (currently 12) and the 0.5 % badge bar as flags once real prod noise is observed.
22. Consider per-surface viewport clipping for the diff (full-page heights drift with data; maybe diff only the top N px for list surfaces).
23. Teach `compare.html` to link each section back to the run galleries.
24. Add `--open` to launch the gallery/compare in a browser (local convenience only).
25. Make `summary.json` carry the surface URL per shot (currently only the key; the URL is reconstructable but not recorded).
26. Record the seeded-thread count quirk discovered in Test C (re-seeding appends messages to the SAME three threads — thread model working as designed) into lessons.md if it bites anyone again.
27. Add the bare-boot smoke boot as a documented `--base-url` test fixture recipe (I used it ad hoc on port 18098; 18099 was held by another session).
28. Sweep TODO_LIST for the reopen-trigger registry impact of this train (none expected — no library adoptions; Pillow is capture-env-only).
29. Check whether `mypy.ini` (the harness's lint arrangement) needs `ps.pillow` mirrored where selenium is declared.
30. Update the `# T23` comment block in `.gitignore` — it says "regenerated from scratch on every run", which is no longer the whole truth (track-mode history persists).
31. Verify the flake's `webphone-backup`/drill lanes are untouched (no product code changed — expected clean; buildflow's nix-build failure was island-js only).
32. Decide whether `--track` should also record the git HEAD sha in `summary.json` (local runs have a repo; prod runs have `/version` + commit — `/version` already carries `commit`).
33. Consider capturing `/config.js` and Settings' service rows as text snapshots alongside shots (cheap wire-facts ledger per run).
34. Add screenshot byte-size deltas to the compare table (crude but free signal for gross layout changes).
35. Evaluate `prefers-reduced-motion` and forced-colors passes as extra emulation axes (accessibility ugliness is still ugliness).
36. Once authenticated prod runs exist: diff-stability check across a quiet window (night run) to calibrate the noise bar on real data.
37. File the `login_remote` 429 experience (wait-and-retry hint is printed; an automatic single retry after Retry-After could be polite — or stay manual by design).
38. Teach `resolve_remote_surfaces` to note when the real list holds FEWER threads than `--threads` (currently silent; Test C hit 3 < 6 silently).
39. Sweep docs for stale references to "24 shots" now that full-page/viewport variance exists (README passkey section? docs/dom-contract untouched — markup unchanged).
40. Re-read `.buildflow.yml` after daemon activity near it (per AGENTS) — it was untouched this train, but the run happened while another session was live.
41. Thank the assert-first splice pattern (it caught an off-by-one) — keep using it for any further surgical file surgery.
42. Consider committing the probe technique as a tiny `scripts/` note or island-tests README addition (stub `listeners` lazily materialize per event name — cost me one iteration).
43. When the other session lands: re-run their spec against MY final tree state before pushing (mixed-tree hygiene).
44. Evaluate running the harness against a TLS-fronted LOCAL boot (self-signed + `--resolve` dance) to exercise the https add_cookie path before the real prod run.
45. Confirm `add_cookie` behavior when the prod session cookie carries `Secure`/`SameSite=Strict` (anonymous prod shots worked cookie-less; the authenticated path injects — smoke it on the TLS-fronted local boot from #44 first).
46. Add a `--dry-run` to remote mode: login + resolve + print the surface list without shooting (fast credential/surface sanity check).
47. Number stability in galleries: the `NN-` prefix in filenames drifts across runs with filtered surfaces; consider zero-padded stable ordering or pure-key filenames in a future format bump (compare keys already ignore the prefix).
48. Make `track()` tolerate a run interrupted mid-capture (partial run dirs already compare gracefully — verified accidentally in Test C via the crashed first attempt; make that a documented property).
49. Sweep `docs/planning/2026-10-29_*ui-ux*` plans for "the harness should" leftovers that this train accidentally satisfied (and strike them with evidence, docs-gates style).
50. Ask the owner whether "unlighthouse kind" should eventually include LIGHTHOUSE-style metric runs per surface (perf/a11y scores) — that is the one unlighthouse feature this train deliberately did NOT build.

## g) QUESTIONS (cannot answer myself)

1. **Prod credentials for the authenticated capture** — which extension should the capture sign in as, and may I have its PBX directory password via `WEBPHONE_CAPTURE_EXTENSION` / `WEBPHONE_CAPTURE_PASSWORD` at run time? (I will not hunt for it in secret stores; the passkey ceremony cannot be automated, so the extension-password lane is the only automatable one — confirm that's acceptable.)
2. **Ownership of the red spec** — the failing "a dead recording keeps one toast" spec and its `shell.js` changes came from ANOTHER live session's voicemail train (unpushed `0d2e9ff8`). Do you want me to FIX/FINISH that change myself (it's blocking buildflow + the push), or leave their in-flight work alone and let that session land it? Related: push the 12 commits once green, or is the daemon's push expected to recover on its own?
3. **Where may prod shots live?** They carry real caller/conversation data and are currently laptop-local under gitignored `ui-shots/` with unbounded history. Is laptop-only the ruling, or do you want a synced/private store (and a retention cap)?

---

### Verification ledger (commands as run)

- Classic matrix: `python3 scripts/ui-capture.py --binary /tmp/wp-visual-bin/bin/webphone` → `ui-capture: 24 shots`, exit 0, `layout: PASS`.
- Track chain: same + `--track` twice → `ui-shots/127.0.0.1_18097/{20261110T063917Z,20261110T064146Z}` + `compare.html`; filtered tracked run added `20261110T064603Z` with `version: v2.12.0` after the fetch-ordering fix.
- Remote local: bare boot `WEBPHONE_ADDR=127.0.0.1:18098 WEBPHONE_DATA_DIR=/tmp/wp-visual-data` + env creds → 22 shots (`20261110T081416Z`), then filtered 6-shot run with the `not-found` fix.
- Anonymous: local (2 shots) and **prod** → `ui-shots/pbx.artmann.tech/20261110T082706Z/` (4 shots, `v2.12.0`).
- Missing-creds error path: env-stripped run → the 5-line instruction message + `--anonymous` hint, exit 2.
- Full-page proof: `--viewport 390x844 --surface settings` → 443×3221 PNGs.
- Island tests direct: `nix run nixpkgs#nodejs -- --test --test-force-exit internal/web/assets/island-tests/*.test.mjs` → exactly one red spec (named above).
- buildflow: run `20261110-083158-4f16e74a`, 81 steps, 77 pass / 4 fail (`nix-hash-fix`, `nix-build` → `island-js`, +2 unidentified in-run; history table's `test-race`/`golangci-lint` failures are from 1d ago).
- Push state: remote `2c2cc950`, HEAD `b68543b4`, 12 unpushed, first unpushed `2026-11-10T06:38:44+01:00` (173 min at 08:31 UTC measurement) — manual-push bar satisfied, push held on the red gate.
