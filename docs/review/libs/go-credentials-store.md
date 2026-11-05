# go-credentials-store — library utilization audit

**Verdict:** NO-GO — it is a stateful credential _service_ (encrypted store, master keys, audit trail, admin UI) for personal infrastructure; webphone's threat model delegates secret-keeping to the platform (systemd `LoadCredential` + the NixOS secrets pipeline) and reads credentials fail-closed at use. **Adoption: n/a (absent, by decision).**
**Pin:** none (absent from go.mod) · **Latest:** v0.3.1 (local checkout tag) · **Status:** not a dependency
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

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
| SIP extension directory passwords                               | per-extension files read **per login, never cached**; empty/missing = loud Rejection (the 2026-10-01 outage class) | internal/authmap/authmap.go:118-139           |
| Webhook secret / credential-brokering secret / CRM token        | file reads at boot; NixOS module wires them through systemd `LoadCredential`                                       | internal/config/config.go:691,710,731,334,723 |
| OIDC client secret                                              | `LoadCredential` path `/run/credentials/webphone.service/…`; rotation = unit restart                               | internal/sso/sso.go:297; AGENTS.md OIDC lane  |
| Session-carried directory password (for the `/phone-api` proxy) | SQLite session rows, swept on read; at-rest accepted deliberately                                                  | internal/session (store seam)                 |

Adopting the library would add a second stateful component with its own master-key rotation problem **inside** a single-binary product, to store secrets the OS already holds — strictly more keys to manage, zero capability gained. The one-time-reveal admin UI would also widen webphone's auth surface, which the platform currently fences.

## Ruling

**Decline.** The platform seam is the right seam for this deployment; the library is the right seam for hosts without systemd/NixOS secrets plumbing. Re-open only if webphone must run somewhere without `LoadCredential`-class support AND accumulates secrets it must audit.

**Refresh trigger:** a non-systemd deployment target; a credential-audit compliance requirement; more than a handful of runtime-mutable secrets.

## How to verify

All commands run from the **webphone checkout root** (repo-relative paths; `git -C` lines carry their own directory).

```console
rg -rn "credentials-store" --type go go.mod internal/ | wc -l    # expect 0
rg -n "os.ReadFile" internal/authmap/authmap.go | head -2        # per-use secret reads
rg -n "LoadCredential" /home/lars/projects/nix-international-telephony/nix/modules/telephony/*.nix 2>/dev/null | head -3
```
