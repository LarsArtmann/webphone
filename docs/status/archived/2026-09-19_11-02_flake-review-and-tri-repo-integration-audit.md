# Status Report — flake review + tri-repo integration audit

**Date:** 2026-09-19 11:02 (CEST)
**Session scope:** read-only analysis in `/home/lars/projects/webphone` and its two
consumer repos (`nix-international-telephony`, `pbx-artmann`). **No code was changed
in any repo this session** — both work items are reviews with recommended (unapplied)
actions.

Work performed:

1. `flake.nix` + `package/nixos-module.nix` review (nix-review skill, full checklist).
2. Full integration audit webphone ↔ nix-international-telephony ↔ pbx-artmann
   (module seam, vhost seams, config.js contract, phone-api contract, E2E DOM
   contract, lock cadence, deployment gaps).

---

## a) FULLY DONE

### Flake review (webphone repo)

- All nix-review checklist categories walked; verdict "already excellent", 5 findings.
- **3 suspected criticals probed and dismissed with evidence** (not vibes):
  - `statix | tee` / `deadnix | tee` pipeline masking: built a probe derivation
    against the locked nixpkgs — stdenv **does** set `pipefail`, gates are real.
  - statix + deadnix run live: both exit 0 (clean).
  - SSE idle-drop behind nginx default 60s read timeout: cqrs-htmx v4.9.0
    heartbeats every 15s (`sse_broadcaster.go:16`) — no idle drop.
- Rejected migrations with reasoning: `go-standard` (3 inputs already; perSystem
  bulk is bespoke checks, not the boilerplate it removes), `fileset` src narrowing
  (harmless store-size only).

### Tri-repo integration audit (all claims source-verified this session)

- **Module seam**: stack imports `inputs.webphone.nixosModules.default` and defaults
  `services.telephony.webphone.package` from its input — binary + deployment shape
  in sync from one source, as designed.
- **vhost seams**: catch-all → webphone; `= /config.js` shadow (48h TURN REST creds,
  daily timer, no app restart so in-memory sessions survive); `= /sip` exact-match →
  sofia :7443 (prefix-trap for `/sip.min.js` and Via/WSS transport both documented
  in-stack); `/events` with `proxy_buffering off` + 3600s.
- **config.js contract**: stack's rendered key set matches webphone's
  `configjs.go` exactly (`sipDomain`, `websocketPath`, `iceServers`, `phoneApi`,
  `contacts`).
- **phone-api contract**: webphone's proxy (session extension creds as Basic auth,
  `proxy.go`) → operator API; routes verified in `api.py`
  (`/phone-api/voicemail/<ext>/summary|messages|<uuid>/audio`, DELETE,
  `/phone-api/history?limit`); per-extension authz enforced (authed ext == path ext).
- **DOM contract**: all 12 E2E-greppable island strings grep-verified present at
  webphone HEAD.
- **Coupling cadence**: 4 webphone re-pins in the stack within ~2 days; current pin
  `4a1266d` bumped today 10:34 (`409fa53`), 36 commits / 55 files past v2.0.0
  (incl. `messages.templ` rewrite).
- **Deployment reality**: `webphone.enable` defaults **true** in the stack, so
  pbx-artmann (which sets no webphone options) runs the webphone UI live —
  corroborated by its TODO_LIST live-200 check of 2026-09-18.
- **Messaging seam unwired** (re-verified clean this session): zero matches for a
  `/hooks/message` sender in stack or pbx-artmann code; `telnyx-webhooks.py` is a
  logger; no `/message`/`/fax` gateway receiver exists; webphone's gateway stays
  loopback default. Inbound SMS never reaches webphone; outbound is dev-mode.

## b) PARTIALLY DONE

- **Flake improvement plan**: reported (5 fixes, severities, one-line hows),
  offered to apply — not applied; user moved to the next question.
- **Integration action plan**: 3 Pareto steps recommended (pbx-artmann phoneApi +
  contacts; stack E2E re-run + 2 render fixes; Telnyx↔webphone messaging bridge) —
  not executed.
- **"E2E unverified against new pin"**: correctly flagged but not resolved — I did
  not query the stack's CI (`gh run list`) nor run the local E2E build to establish
  last-green vs the 10:34 pin.

## c) NOT STARTED

- No edits in any of the three repos (both tasks were analyses).
- No `AGENTS.md` / `TODO_LIST.md` updates from session findings (memory protocol
  says: update at discovery time — not done; see e).
- No HARVEST of this report's section (f) into `TODO_LIST.md` (docs-health) —
  deliberately deferred to owner instruction per "report, then wait".

## d) TOTALLY FUCKED UP (honest ledger)

1. **`rg -rn` flag misuse, twice**: `-rn ""` parses as replace-with-`n` — the
   displayed output of two greps was corrupted (`.gitnpre-commit`,
   `legacyPackages.n`). My "no hooks sender" conclusion happened to survive a clean
   re-verification (zero matches, `rg -c`, no pipes), but I initially shipped a
   load-bearing claim on corrupted evidence and only noticed during this
   self-review. Same failure family as the AGENTS pipeline-masking lesson — I
   re-committed it this session.
2. **Exit-code-through-pipe, again**: `rg … | head; echo $?` reports `head`'s
   status, not rg's. Did it once mid-session, then AGAIN in the first
   re-verification before fixing. Third strike avoided by a no-pipe re-run.
3. **One materially wrong claim to the user, caught post-hoc**: I said pbx-artmann
   "has not yet absorbed" today's webphone re-pin. False — its path-lock
   `lastModified` decodes to **2026-09-19 10:42**, _after_ the stack's 10:34 bump.
   I eyeballed a vague "~2026-09-18/19" instead of running one `date -d @…`.
   Corrected here; the integration report's other claims stand as verified.
4. No intentional lies; no data loss; no unauthorized writes (read-only session).

## e) WHAT WE SHOULD IMPROVE (self-review answers)

**What did I forget?**

- **webphone's own `nixos-module.nix` vhost has no `/events` location** — no
  `proxy_buffering off`, no long read timeout. The stack fixed this on ITS vhost
  (`web.nix:227`), but any non-stack consumer using the module's `nginx.enable`
  gets SSE through default nginx buffering. Missed during the flake review; found
  in retrospect. This is finding #6 below.
- Checked `.buildflow.yml` nowhere: the devShell `GOEXPERIMENT` recommendation was
  made without confirming buildflow doesn't already inject it (recommendation
  stands, but it should have been checked first).
- Cited the stack's `tests/webphone.nix` and `browser-e2e.py` from AGENTS.md
  instead of reading them at source (12-string grep was verified, full test files
  were not).

**What could I have done better?**

- Decode timestamps instead of approximating them (would have prevented d.3).
- Probe CI state (`gh run list -w …` in the stack) instead of shipping "unverified".
- Clean grep/pipe hygiene from the first command.

**What could we still improve (product-level, from the audit)?**

- The **entire messaging/fax product surface is dark in production**: webphone's
  inbound-hooks + webhook-gateway architecture is fully implemented and tested
  webphone-side, with **zero counterparts** stack-/deployment-side. This is a ghost
  system by deployment, not by code — the highest-value integration work available.
- Production runs with `phoneApi` off, so History + Voicemail panels are dark while
  the stack ships the API and its own template host enables it (one config line).
- Cross-repo contracts (config.js keys, phone-api routes, E2E DOM ids) are held by
  convention + scattered tests; there is no single cross-repo contract test.
- The stack rides webphone **main** (36 unreleased commits on a live PBX path today)
  with browser E2E as manual dispatch only — the regression gate lags the pin.

## f) Things to get done next (impact-ordered, not a commitment list)

**webphone repo**

1. ~~Apply flake fix 1: devShell `GOEXPERIMENT = "jsonv2"` + `GOTOOLCHAIN = "local"`~~ done at `4c6bba1`
   ~~(check `.buildflow.yml` first for overlap).~~
2. ~~Apply flake fix 2: `recommendedProxySettings = true;` on the module's websocket~~ done at `4c6bba1`
   ~~location (XFF readiness for the planned `KeyExtractorFromClientIP` flip).~~
3. ~~NEW (self-review find): add an `/events` location (buffering off, long read~~ done at `4c6bba1`
   ~~timeout) to the module's own nginx vhost.~~
4. ~~Apply flake fix 3: `services.webphone.memoryMax` option + `serviceConfig.MemoryMax`.~~ done at `4c6bba1`
5. ~~Apply flake fix 4: assertion that `dataDir` is under `/var/lib`.~~ done at `4c6bba1`
6. ~~Apply flake fix 5 (style): `lib.*` instead of `pkgs.lib.*`; drop `with pkgs;`.~~ done at `4c6bba1`
7. ~~Consider `nixosModules.webphone` alias next to `.default`.~~ done at `4c6bba1`
8. ~~Optional: narrow `src` via `lib.fileset` (store size only).~~ done at `4c6bba1`
9. ~~Update webphone `AGENTS.md` with the integration-state facts (stack rides main;~~ done at `4c6bba1`
   ~~pbx-artmann defaults; messaging seam unwired; pin `4a1266d` absorbed 10:42).~~
10. ~~`TODO_LIST.md` HARVEST from this report (after owner go-ahead).~~ done at `4c6bba1`
11. ~~`webphoneVersion` drift: HEAD reports `v2.0.0` while 36 commits newer — next~~ done at `d815004`
    ~~release per runbook resolves; keep in mind when reading `/version`.~~

**stack repo (nix-international-telephony)**

12. ~~Re-run browser E2E against `4a1266d` (markup changed since the tag: the~~ done at `28d4688`
    ~~`messages.templ` rewrite).~~
13. ~~Fix `contactsJson` double-escape in `web.nix` (`escapeJs` + `toJSON` together).~~ done at `28d4688`
14. ~~Fix rendered `phoneApi` flag: mirror `cfg.webphone.phoneApi.enable` only, not~~ done at `28d4688`
    ~~`|| cfg.operator.enable` (else operator-only config yields erroring panels).~~
15. Promote browser E2E CI from `workflow_dispatch` to periodic/per-push (their TODO
    row 31; owner's call).
16. Operator API: HTTP Range support in `send_file` (voicemail audio seek).
17. Operator API: mark-read flip via `vm_read` (unread badges).
18. Operator API: auth-failure lockout for `/phone-api`.

**pbx-artmann repo**

19. ~~`services.telephony.webphone.phoneApi.enable = true;` — lights History +~~ done (deployed to prod, then re-deployed with the d5026d1 credentials fix (see appendix))
    ~~Voicemail panels in production (biggest single-line win found).~~
20. ~~`webphone.contacts = […]` — surface 1000 (Lars), 1001 (Alice), 2000 (ring~~ done (deployed to prod with the same pass (see appendix))
    ~~group) as shared contacts.~~
21. Owner decisions: `operator.enable`, `recording.serve.enable` on prod.
22. ~~Build the messaging bridge, inbound leg: `telnyx-webhooks.py` POSTs normalized~~ done (bridge built 2026-09-19, inbound verified end-to-end (see appendix))
    ~~inbound SMS/MMS to webphone `/hooks/message` (shared secret, `/hooks/fax` for~~
    ~~TIFFs).~~
23. ~~Messaging bridge, outbound leg: serve `{url}/message` + `/fax` (multipart,~~ done (outbound leg live, waiting only on Telnyx portal steps (see appendix))
    ~~Bearer secret, `{"provider_ref"}` receipts) so webphone's webhook gateway mode~~
    ~~goes live; wire status-hook callbacks.~~
24. ~~Wire `WEBPHONE_GATEWAY__WEBHOOK_SECRET` via~~ done (shipped with the bridge wiring (see appendix))
    ~~`services.webphone.environmentFile` + add to `generate.sh`/`push-secrets.sh`~~
    ~~lists (their documented invariant).~~
25. ~~Re-verify prod after the phoneApi flip (eval + optionally the stack's~~ done (prod re-verified after the phoneApi flip (see appendix))
    ~~`telephony-webphone` VM test).~~
26. Owner/portal BLOCKEDs: Telnyx messaging-profile DID attach (their TODO row 42),
    Warsaw DID re-purchase — prerequisites for real inbound SMS.

**cross-repo / policy**

27. Decide stack→webphone pin policy: tags at release boundaries vs riding main
    (current de facto) — question g.1 below.
28. Evaluate replacing pbx-artmann's `path:` input with a github pin (the LIVE path
    input already burned them 2026-09-18), or keep path + lock discipline.
29. ~~Add a cross-repo contract test: stack's rendered config.js key set vs webphone's~~ done (stack VM test pins the config.js key set (see appendix))
    ~~`configjs.go` (today only convention + two scattered renderers hold it — a~~
    ~~nascent split brain).~~
30. ~~Consider teaching webphone's module eval-check to assert the vhost it generates~~ done (module-check asserts vhost locations plus the webphone unit (see appendix))
    ~~(today it linkFarms config.json + port only).~~

## g) Questions I cannot figure out myself

1. ~~**Pin policy**: should the stack keep riding webphone `main` (velocity — 4 pins~~ done (superseded - the P8 pin-policy decision memo in the SUPERB plan (docs/planning/2026-09-19_15-37) owns it, owner call pending)
   ~~in 2 days, but a live PBX runs unreleased commits) or pin release tags and bump~~
   ~~per the release runbook? This is an owner risk-appetite call, not discoverable.~~
2. ~~**Messaging bridge priority**: build the Telnyx↔webphone bridge now (costs Telnyx~~ done (answered by events - bridge built 2026-09-19 (see appendix))
   ~~messaging profile + portal steps, lights up Messages/Fax in production), or is~~
   ~~the Messages tab deliberately parked until the DID/messaging-profile BLOCKEDs~~
   ~~clear? Product sequencing only you know.~~
3. ~~**Prod trust posture**: enabling `phoneApi` proxies each signed-in extension's~~ done (answered by events - one-liner shipped per owner blanket go (see appendix))
   ~~SIP password as Basic auth to the loopback operator API, which currently has no~~
   ~~auth-failure lockout (stack TODO row 42). Ship the one-liner now and harden~~
   ~~after, or harden first? Risk tolerance is yours.~~

---

_Report is a point-in-time snapshot. No repos were modified this session; the
auto-commit daemon will pick up this file. HARVEST into TODO_LIST deferred pending
owner instruction._

---

## Appendix (2026-09-19, SUPERB execution): outcome of this report's §f list

Executed same day under `docs/planning/2026-09-19_11-51_SUPERB-tri-repo-functional-completion.md`
(owner: "MAKE IT PERFECT AND FUNCTIONAL! ALL OF IT!"). Outcome against §f:

- **f.1–f.8, f.30 (webphone flake/module)**: DONE — memoryMax, `/events`
  vhost location, websocket `recommendedProxySettings`, dataDir
  assertion, devShell env attrs, `lib.*` style, `nixosModules.webphone`
  alias, fileset src, module-check vhost+unit assertions. All gates
  green (buildflow, go tests, flake check, aarch64, smoke 21/21);
  pushed as `4c6bba1`.
- **f.9, f.10**: DONE (AGENTS integration section; TODO_LIST harvest).
- **f.11**: open by design — v2.1.0 release ceremony is the owner's.
- **f.12–f.14 (stack)**: DONE — contactsJson double-escape and the
  phoneApi flag fixed; the VM test now pins the full config.js key set
  - a quoted-name contact; browser E2E green against the new pin
    (`28d4688`).
- **f.19–f.20, f.22–f.25 (pbx-artmann)**: DONE and DEPLOYED to
  pbx.artmann.tech (toplevel `h8y8i78…`), then re-deployed with the
  `$CREDENTIALS_DIR` fix (`xv3infy…`). phoneApi live
  (`/config.js` → `"phoneApi": true`), contacts live, operator unit
  running, bridge green.
- **f.21, f.26 (owner/portal)**: still blocked on owner — unchanged.
- **f.27–f.28 (policy)**: still owner decisions — now also as TODO rows.
- **f.29 (cross-repo config.js contract)**: DONE via the stack's VM-test
  key-set assertion (the split brain now fails a check, not a user).
- **g.2 answered by events**: the bridge was built; inbound verified
  end to end with a marked message (`[bridge-verify 2026-09-19]`, 202).
  Outbound waits only on the Telnyx portal steps (messaging profile +
  real API key at `/var/lib/telephony-secrets/telnyx_api_key`, then
  `systemctl restart telnyx-webhooks`).
- **g.3 answered by events**: shipped the one-liner (`phoneApi.enable`)
  per the owner's blanket go; the operator-API hardening items remain
  in the stack's TODO.

## Appendix (2026-09-19 evening): the CSRF fronting bug, timeline

- **2026-09-18 09:23** — CSRF middleware wired into the web server
  (cqrs-htmx adoption). Correct for direct-HTTP listeners; the fronted
  shape had no test.
- **2026-09-18** — the stack's `forceSSL` vhost ships; the Origin
  `https://` vs plain-HTTP-listener mismatch becomes production's
  reality, unnoticed.
- **2026-09-19 morning** — the v2.0.0-era build is live on
  pbx.artmann.tech: every browser login 403s
  (`contradictedAttestationOrigin`, scheme-only mismatch). Calls keep
  working (the island is serverless), so nothing else screams; the
  stack browser E2E stays green because it never asserted the session
  POST.
- **2026-09-19 early afternoon** — v2.1.0 (`d815004`) released with
  both halves of the fix: `csrf.trusted_proxies`/`trusted_origins`
  (the trusted proxy's `X-Forwarded-Proto` is believed) and login CSRF
  rotation with island adoption via `GET /api/csrf`.
- **2026-09-19 evening** — the stack browser E2E gains the
  SESSION-CREATED gate (per-browser, negative-tested: stripping the
  csrf settings fails the run with `SESSION-GATE-FAILED: session=403`).
  The silent-breakage class dies. The prod redeploy is the owner's one
  command (SUPERB plan P1; the pbx-artmann lock now rides the v2.1.0
  chain as of `019093f` there).
