# go-filewatcher — library utilization audit

**Verdict:** NO-GO — webphone has nothing to watch. Config is boot-immutable, secrets are per-use reads or systemd `LoadCredential`, rotation is restart-by-doctrine. **Adoption: n/a (absent, by decision).**
**Pin:** none (absent from go.mod) · **Latest:** v2.4.3 (local checkout tag) · **Status:** not a dependency (fsnotify v1.10.1 rides indirect via koanf's file provider)
**Audited:** 2026-11-04 · **Prior reports:** none (first audit).

## Status

| | |
|---|---|
| Module | `github.com/larsartmann/go-filewatcher` |
| In webphone | absent (0 references) |

## What the library offers

High-level fsnotify wrapper: recursion, 17 composable filters (extensions, gitignore, generated-code), 18 middleware (recovery, logging, OTel), global/per-path debounce, NFS polling fallback, ENOSPC self-heal, content hashing, `Event{Path,Op,Size,ModTime,Hash}` stream over `Watch(ctx)`.

## Webphone overlap analysis (verified)

- `rg "fsnotify|watch|reload|SIGHUP" --type go internal/ cmd/` → only UI reload strings and test fixtures. **No watcher, no hot-reload, no SIGHUP handler exists.**
- Config (internal/config/config.go, koanf) loads once at boot; unknown keys WARN (the pin-gap alarm).
- Secrets: `authmap.SIPPassword` reads each extension's password file **per login, never cached** (internal/authmap/authmap.go:118-139) — rotation is effective on the next login without a watcher; OIDC secret arrives via systemd `LoadCredential` and rotation = unit restart (systemd snapshots credentials at start — a watcher could not see rotations anyway).
- A hot-reload feature would also fight the boot-failure doctrine (WRONG config fails boot loudly; a watcher would introduce a second, quieter config lifecycle).

## Ruling

**Decline.** The library is well-built for services that own watched workspaces (generated files, config dirs, inbox folders); webphone owns none. `fsnotify` stays indirect-only inside koanf. Re-open only if a feature genuinely needs filesystem reactivity (e.g. a drop-folder fax intake — for which a polling loop over the existing pbx/gateway seams is the more likely home anyway).

**Refresh trigger:** a drop-folder/intake feature; an operator demand for live config reload that outweighs the boot-fail doctrine.

## How to verify

```console
rg -rn "filewatcher|fsnotify" --type go internal/ cmd/ | wc -l   # expect 0
rg -n "SIPPassword" internal/authmap/authmap.go | head -3        # per-login read, no cache
rg -n "LoadCredential" internal/config/config.go | head -3       # systemd owns rotation
```
