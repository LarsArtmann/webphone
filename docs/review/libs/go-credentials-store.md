# go-credentials-store — library utilization audit

**Verdict:** NO-GO — it is a stateful credential _service_ (encrypted store, master keys, audit trail, admin UI) for personal infrastructure; webphone's threat model delegates secret-keeping to the platform (systemd `LoadCredential` + the NixOS secrets pipeline) and reads credentials fail-closed at use. **Adoption: n/a (absent, by decision).**
**Pin:** none (absent from go.mod) · **Latest:** v0.4.3 (re-verified 2026-11-22; the v0.4.2→v0.4.3 delta is a templ v0.3.1100 coordinated regeneration — the attribute-sanitization consumer half of GHSA-94c2-xg24-pwhv — plus a fleet-aligned cmdguard v5.1.0 bump: adminui/release hardening, **no change to the stateful-service shape this ruling declines**; the v0.4 line through v0.4.2 was adminui host-nav links, a CSP-nonce bridge, generator-pinned templ regeneration, and a consumer-simulation pre-tag gate) · **Status:** not a dependency
**Audited:** 2026-11-04 · 2026-11-22 re-verify · **Prior reports:** none (first audit).

## Status

|             |                                               |
| ----------- | --------------------------------------------- |
| Module      | `github.com/larsartmann/go-credentials-store` |
| In webphone | absent (0 references)                         |

## What the library offers

Event-sourced, encrypted-at-rest credential store: `credentials.Build(ctx, Options{ConfigPath, MasterKeys, Version})` → container with `Resolver().Resolve(ctx, platform, kind)`; XChaCha20-Poly1305 sealing with KeyID rotation; lifecycle audit (created/rotated/probed/used/revealed); loopback admin UI with one-time reveal (`adminui.Mount`); SSE live updates; Prometheus metrics; standalone host binary. It solves credential _storage, rotation and audit_ as its own long-running component.

## Webphone overlap analysis (verified)

| Credential                                                      | Webphone's answer                                                                                                  | Evidence                                      |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ | --------------------------------------------- |
| SIP extension directory passwords                               | per-extension files read **per login, never cached**; empty/missing = loud Rejection (the 2026-10-01 outage class) | internal/authmap/authmap.go:194-219           |
| Webhook secret / credential-brokering secret / CRM token / paperless token | file reads at boot; NixOS module wires them through systemd `LoadCredential`                                       | internal/config/config.go:812,831,852,874     |
| OIDC client secret                                              | `ClientSecretFile` read at the sso seam; wired via systemd `LoadCredential`; rotation = unit restart                               | internal/sso/sso.go:306; AGENTS.md OIDC lane  |
| Session-carried directory password (for the `/phone-api` proxy) | SQLite session rows, swept on read; at-rest accepted deliberately                                                  | internal/session (store seam)                 |

Adopting the library would add a second stateful component with its own master-key rotation problem **inside** a single-binary product, to store secrets the OS already holds — strictly more keys to manage, zero capability gained. The one-time-reveal admin UI would also widen webphone's auth surface, which the platform currently fences.

## Ruling

**Decline.** The platform seam is the right seam for this deployment; the library is the right seam for hosts without systemd/NixOS secrets plumbing. The v0.4.x line (reviewed at the tags, 2026-11-13) is UI/release engineering — the premise stands. Re-open only if webphone must run somewhere without `LoadCredential`-class support AND accumulates secrets it must audit.

**Refresh trigger:** a non-systemd deployment target; a credential-audit compliance requirement; more than a handful of runtime-mutable secrets.

## 2026-11-22 re-verify

v0.4.3 (2026-11-13) is the tag new since the last currency check; its delta, read at the CHANGELOG: the templ v0.3.1070→v0.3.1100 coordinated wave (every `internal/ui/*_templ.go` regenerated with the pinned generator; pairing consumers MUST run templ ≥ v0.3.1100) and cmdguard/v5 v5.1.0 bumped fleet-aligned before the tag via the consumer-sim gate — release/UI hardening, the same class the v0.4.x line has shipped all along; the stateful-service shape is untouched. [Unreleased] carries one Changed item — the templ ↔ templ-components coordinated pair (v0.3.1100 + v1.27.0), "the dependency-currency plan's last open item" — again dependency motion, no service change. webphone's premise is intact: the platform still owns secrets (AGENTS.md OIDC lane — `LoadCredential` into `/run/credentials/webphone.service/`, rotation = unit restart; boot-time file reads at internal/config/config.go:812,831,852,874; per-login directory reads at internal/authmap/authmap.go:194-219 — refs refreshed), and `credentials-store` still replays 0 references. None of the flip triggers — a non-systemd deployment target, a credential-audit compliance requirement, a pile of runtime-mutable secrets — has fired. Verdict unchanged.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -rn "credentials-store" --type go go.mod internal/ | wc -l    # expect 0
rg -n "os.ReadFile" internal/authmap/authmap.go | head -2        # per-use secret reads
rg -n "LoadCredential" /home/lars/projects/nix-international-telephony/modules/telephony/*.nix 2>/dev/null | head -3
```
