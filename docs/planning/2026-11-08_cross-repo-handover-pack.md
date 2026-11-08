# Cross-repo handover pack — the four standing legs, copy-pasteable (2026-11-08)

**Status:** HANDOVER (verified handovers only; no assistant ssh) · Every
stack-side path below was READ-verified today against
`~/projects/nix-international-telephony` (read-only); every webphone-side
path against this repo at the paperless-parity train. The owner command
sheet (`docs/planning/2026-09-24_19-25_owner-terminal-command-sheet.md`)
remains the sheet for the commands that predate this pack; this pack owns
the four legs the execution plan named.

## Leg 1 — Stack: `crm.caFile` pass-through (the private-CA lane)

> **RESOLVED 2026-11-08 (before this pack's steps ran):** the stack side
> was ALREADY LANDED when the legs train arrived — `options.nix` declares
> `crm.caFile`, `web.nix` threads it into the webphone settings, and the
> eval arm pins the rendered `ca_file` host path. Nothing was owed here;
> the pack's steps below are kept as the original plan record.

**Webphone side: SHIPPED.** The app reads `crm.ca_file` (host path, plain
file — a CA bundle is public material, never a credential), and webphone's
own NixOS module carries `services.webphone.crm.caFile`, pinned by
`nix/module-check-crm.nix` (renders `settings.crm.ca_file` as a host path
and never routes it through LoadCredential).

**Stack side: OWED.** `modules/telephony/default.nix` exposes
`services.telephony.webphone.crm.{enable,url,tokenFile}` (the assertion at
line ~179 demands url+tokenFile together) but has **no `caFile` option** —
a stack deployment pointing the webphone at a private-CA CRM vhost has no
way to pass the bundle through.

**Steps (stack repo):**

1. In `modules/telephony/options.nix` (the telephony option tree), add
   `crm.caFile = mkOption { type = types.nullOr types.path; default = null;
   description = "PEM bundle vouching for the CRM vhost's TLS chain (host
   path, not a credential)"; }`.
2. In `modules/telephony/default.nix`, thread it into the webphone module
   config the same way `crm.url`/`crm.tokenFile` are threaded
   (`services.webphone.crm.caFile = mkIf (cfg.webphone.crm.caFile != null)
   cfg.webphone.crm.caFile;`).
3. Extend the existing both-or-neither assertion only if desired — the app
   itself fails the boot on a missing/empty/certificate-free bundle
   (`internal/crm/transport.go`), so the stack does NOT need to duplicate
   that check.
4. Verify: `nix flake check` in the stack (the `telephony-eval` check
   catches option renames), then the vm-test that boots the CRM vhost if
   one exists (`tests/agent.nix`, `tests/boot.nix` are the boot-shaped
   ones).

**Acceptance:** a stack config with `services.telephony.webphone.crm.caFile`
evaluates, the rendered webphone settings carry `crm.ca_file`, and the app's
own CA-lane test (`internal/config/config_test.go` +
`internal/crm/transport_test.go` shapes) covers the app half.

## Leg 2 — Stack: the `/hooks/ring` dialplan leg (ring-while-closed)

> **RESOLVED 2026-11-08 (as-built, superseding the steps below):** the
> lane shipped as a LOOPBACK SHIM, not a dialplan-side secret: the
> dialplan (store-rendered, world-readable by construction) carries NO
> secret — mod_curl posts compact JSON (quotes percent-encoded for the
> one-token grammar, source-verified against mod_curl.c) to
> `ringhook.py`, which holds `webphone_secret` via LoadCredential and
> Bearer-signs the forward to `/hooks/ring`. Wired as
> `services.telephony.ringhook.{enable,port,webphoneUrl,webhookSecretFile}`
> (`modules/telephony/ring.nix`: hardened systemd unit, secret defaults
> to the messaging bridge's gatewaySecretFile), dialplan arms in
> `freeswitch.nix` render ONLY when enabled (off = byte-identical).
> Verified: `ringhookCheck` eval arm (unit + credential derivation +
> rendered curl arms + off-dialplan carries none) and
> `tests/test_ringhook.py` (7 specs; 351-test discovery green).
> Boundary: the arms ride each extension's default-context entry — the
> RING-GROUP path is UNCOVERED (same question the legs report §g-1
> asks the owner).

**Webphone side: SHIPPED.** `POST /hooks/ring` shares the webhook secret,
payload `{owner, from, call_id, state: ring|end}`, answers 202 fire-and-
forget (the dialplan must never depend on push health), prunes pruned
devices on `ErrUnregistered`. Server contract pinned:
`internal/server/hooks/ring_test.go`, `internal/server/push_api_test.go`;
transports in `internal/push`.

**Stack side: OWED.** No `hooks/ring` reference exists anywhere in the
stack's telephony module today (verified by grep).

**Steps (stack repo):**

1. In the FreeSWITCH dialplan (the stack's dialplan XML/ls files under
   `modules/telephony/`), add a POST on inbound-call RING and RING-END:
   `curl`-style via mod_xml_curl's sibling or FS's `mod_http`-class
   (whatever the stack already uses for the Telnyx webhook bridge — the
   bridge pattern in `telnyx-webhooks.py` (pbx-artmann) is the documented
   precedent; reuse its auth shape).
2. Auth: the webhook secret env the stack already carries for the Telnyx
   bridge (`/var/lib/telephony-secrets/` slot — same PATH as the existing
   bridge's Bearer), header `Authorization: Bearer <secret>`.
3. Body: `{"owner":"<extension>","from":"<caller>","call_id":"<uuid>",
   "state":"ring"}` and `..."state":"end"}` on the hangup path.
4. Fire-and-forget contract: the dialplan must NOT block on the response —
   a 202 is an ack, a timeout is a no-op (the push transports prune their
   own dead devices).
5. Verify: `tests/` grows the ring-hook scenario next to the existing
   dialplan tests (`tests/cdr-visibility.nix` is the shape template).

**Acceptance:** an inbound call to a registered extension with a push
device registered rings the device while the browser is CLOSED; the ring
hook timing (ring + end both posted) is asserted in the VM test.

## Leg 3 — Stack: the FreeSWITCH `mod_enum` build repair (the CI 1h ceiling)

> **STEP 2 EXECUTED 2026-11-08 (exclusion, the honest arm):**
> `applications.enum` is DELIBERATELY EXCLUDED from the mirrored module
> set — grep-verified ZERO enum call sites across the stack's dialplan
> XML, `dialplan_sim.py` and `pbx.nix` first, with the ruling comment at
> the exclusion site. `telephony-eval` green; the pbx toplevel builds
> locally. RUN 42529951959 (2026-11-08 23:22 UTC): aarch64 VM test ✓,
> stdlib python suites ✓, the nix-flake-check leg ran 46m40s under the
> old ceiling — under it on TIME, red on ONE suite:
> `telephony-voicemail`'s quota-exceeded log assertion — the sibling
> session-2's OWN named open red (their report title carries it: "quota
> model correction red"), not the legs' work (ringhook rides a
> default-off byte-identical dialplan there). ACCEPTANCE (twice-green)
> is blocked on that named red, owned by the sibling's MWI/quota train.

**State:** the stack's mirrored module set includes `applications.enum`
(`modules/telephony/freeswitch-modules.nix:29`); its build produces the
stack CI 1-hour ceiling that cancels `nix flake check` runs and gates every
relock (TODO §Cross-repo, standing).

**Steps (stack repo):**

1. Reproduce: `nix build` the FS derivation with enum enabled, watch the
   hang point (the prior evidence lives in the stack's CI logs — the
   ceiling number is the observable, not a local repro).
2. Either FIX the module build (the usual classes: a missing input to the
   mirrored module set, an autotools fetch inside the sandbox) or EXCLUDE
   `applications.enum` from the mirrored set if no dialplan path uses enum
   resolution (grep the stack dialplan for enum URLs first — `mod_enum`
   resolves e164 via DNS/ENUM; if the dialplan never calls it, exclusion is
   honest).
3. Verify: the stack's CI run completes under the ceiling twice
   consecutively (the E2E budget rule: two consecutive greens).

**Acceptance:** stack `nix flake check` finishes < ceiling with enum either
building or honestly excluded, twice in a row.

## Leg 4 — Kith CRM live e2e (#9 — OWNER-GATED, `blocked`)

The pbx-artmann host wiring LANDED (`5df19b5`: dnsblockd CA into the system
trust bundle via `security.pki.certificateFiles`, `services.telephony.webphone.crm`
at https://crm.home.lan, the `webphone_crm_token` secret slot in
generate.sh). The issue #9 comment corrected the option name (nixpkgs has no
`trustedCertificateAuthorities`). What remains is the owner sequence:

1. Mint the Kith API token.
2. Write it to `~/.pbx-prod-secrets/webphone_crm_token`.
3. `scripts/push-secrets.sh` (pbx-artmann).
4. Deploy the pbx-artmann train.
5. Verify the acceptance bar:
   - an inbound caller the CRM knows renders the linked NAME in the island
     (the `crm-context` smoke surface: `scripts/ui-capture.py --crm`
     asserts the strip in the DOM — the only visual proof that exists);
   - a call lands `call_logged` on the contact journal (CRM side);
   - revoking the token turns the Settings CRM tri-state red AND `/healthz`
     names the `crm` leg (this repo's `internal/app/health_crm_test.go`
     pins the readiness shape — the strict half is deliberate, pack row B8).

## Sequencing + the standing rules

> **2026-11-08:** Legs 1–3 executed (Leg 1 pre-landed, Leg 2 as-built
> above, Leg 3 excluded + acceptance pending CI); Leg 4 stays owner-
> gated. The order below is kept as written at handover time.

- Order by blast radius per unit of effort: Leg 3 (unblocks every relock) →
  Leg 2 (unlocks ring-while-closed E2E) → Leg 1 (private-CA option, small)
  → Leg 4 (owner ritual, gated on the Kith branch's own cleanliness — its
  branch mixes another session's work and its `internal/catalog` gate is
  red; Q3 in TODO §CRM tail owns the push decision).
- webphone pushes BEFORE the stack re-pins (tri-repo rule); the stack tree
  must be CLEAN before pbx-artmann re-locks.
- No assistant ssh on the stack hosts; every leg above is either an
  assistant-executable repo change (Legs 1–3 in the stack repo) or an owner
  ritual (Leg 4).
