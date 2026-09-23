# Pareto cycle CLOSE — v2.6.0 shipped, all gates green (2026-09-23 21:15)

Session 3 of the pareto execution. Sessions 1–2 landed C9–C17,
C18–C20, C22–C23 and pre-cleared C1c/C6; this session closed the
load-gated remainder. The plan's verdict is filled at
`docs/planning/2026-09-23_04-29_SUPERB-pareto-execution-plan.md`
(Outcomes + Verdict sections). Every assistant-executable coarse
task is DONE and verified; what remains is owner-terminal by design.

## Gates, with evidence

- **Stack browser E2E ×2 GREEN** at stack `271f5ef` — 195s (plain)
  and 184s (`--rebuild`), both under the 445s budget by miles.
  REGISTRATIONS/calls/DTMF/ICE/transfer/reconnect/contacts all
  green. Two intermediate runs flaked at different post-drill steps
  (CONTACTS-ROUNDTRIP, INCOMING-SHOWN — the known tail variance,
  re-run per runbook).
- **v2.6.0 GitHub release PUBLISHED** — tag `807ca0c`, full
  CHANGELOG 2.6.0 body (`gh release view v2.6.0`).
- **Release smoke GREEN** — 41 + 4 restart checks, 0 failed,
  `/version` exactly `v2.6.0`. Note: `--expect-version` wants the
  RELEASED build; locally pass `--bin $(nix build .#webphone)`
  (a bare `go build` reports Go's pseudo-version and fails the
  check — that first "failure" was my invocation, not the build).
- **`nix flake check` GREEN** (24s) incl. the KVM backup-VM test.
- **pbx-artmann relock #4 + re-pin** — first pin `be876ae`
  (daemon-swept `8104448`, pushed with the concurrent session's
  docs at `438c348`), then re-pinned to the verified stack rev
  `271f5ef` with narrative commit `20b2a18`. lock-drift-probe
  green both times; both toplevels (x86_64 + cross aarch64) green
  both times; webphone ExecStart store path moved
  `lq5fj…-webphone-2.5.0` → `7bm0h…-webphone-2.6.0` at the first
  pin (byte-identical across the re-pin — both stack revs lock
  train `7197f1c`).

## Two incidents worth knowing about

1. **The FOUC E2E scenario had never actually run.** Its first live
   run dead-asserted, and the fix took four stack commits
   (`784126c` → `9fb0539` → `f42cf9d` → `271f5ef`), each rejected
   or confirmed by instrumented evidence rather than theory. Two
   independent causes: (a) a soft `location.reload()` serves
   subresources from cache, where `Network.setBlockedURLs` cannot
   intercept — the "blocked" preload kept executing; fixed with
   `Page.reload{ignoreCache}`. (b) chromedriver executes no scripts
   against a document mid-navigation, so driver-side polling is
   structurally blind to the flash window — three runs sampled
   unbroken `dark` while nginx proved the reload happened. The
   scenario now counts theme ticks IN-PAGE
   (`Page.addScriptToEvaluateOnNewDocument`); green evidence:
   `rafUnthemed=106` flash ticks, then settle-dark; pair 2 zero
   unthemed ticks. ROADMAP g3 is answered: the scenario costs ~15s
   — no E2E budget bump needed.
2. **webphone main had a broken `nix build` for ~2h.** The
   daemon-swept go-etag v0.6.0 bump (`e85923d`, split into
   entitytag/server submodules) changed go.mod without the modules
   hash; the flake check caught it (`go-etag/server@v0.6.0: no such
   file`). vendorHash recomputed and pushed (`0a7a732`). The stack
   trains were never affected (they lock webphone `7197f1c`, before
   the bump). Lesson queued for the owner-calls batch: dep bumps
   swept by the daemon should trigger the vendorHash roundtrip
   reflex immediately.

## OWNER items (the only remainder)

1. **Deploy** — chain ready end-to-end (webphone `7197f1c` → stack
   `271f5ef` → pbx-artmann `20b2a18`):
   `cd ~/projects/pbx-artmann && nixos-rebuild test --flake .#pbx --target-host root@pbx.artmann.tech`
   → smoke → `nixos-rebuild switch`.
2. **Post-deploy probes** —
   `python3 scripts/webphone-smoke.py --base https://pbx.artmann.tech --expect-version 2.6.0`
   (0 failed), plus the rejection-banner check with a real
   extension session (self-send an SMS to the PBX DID, expect the
   persisted provider reason in the failed bubble).
3. **SMS-lane root cause** (prod bridge): journal the
   telnyx-webhooks unit (`journalctl -u telnyx-webhooks --since
   today | grep -i 'sms\|422\|error'`), fix creds/bridge, send a
   test SMS, record the cause in the stack runbook. Webphone-side
   classification (`1d53f44`) and refusal-422 (`6ac8962`) ride this
   deploy.
4. **Owner-calls batch** — briefing at
   `docs/planning/2026-09-22_13-50_owner-calls-briefing.md`. Since
   the last revision: g1 (force-push ratification) and g3 (E2E
   budget — answered, no bump) can be closed from evidence; the CRM
   option shape to ratify is now live as
   `crm.{enable,url,tokenFile}` in stack `be876ae`+; g2
   (KVM-timeout policy) still open; add the dep-bump→vendorHash
   reflex question.
5. **Announcements** — v2.6.0 drafts at
   `docs/announcements/2026-09-23_v2-6-0_drafts.md` (plus the
   v2.1–v2.3, v2.5.0 back-catalog there); owner picks channels and
   the disclosure posture.

Known noise, unchanged: the STACK repo's buildflow vulnix step
crashes on NVD's retired 2.0 feed (404) — documented in its
AGENTS.md, repo-independent. webphone's own `nix run .#vulnix`
verdict is zero real advisories.
