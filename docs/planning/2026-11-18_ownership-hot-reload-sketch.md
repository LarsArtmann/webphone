# Ownership hot-reload — design sketch (NOT IMPLEMENTED)

> STATUS: SKETCH, minted 2026-11-18 by the line-ownership follow-up
> train (TODO_LIST § "Line-ownership model residue", f.16). Nothing in
> this doc is implemented. The SIGHUP-vs-file-watch choice is an
> OWNER CALL — do not implement either arm without a ruling.

## The problem

The ownership plane — `auth.users` (email → lines), `identities`
(extension → DID), and the `extension_password_files` inventory — is
read EXACTLY ONCE, at boot, by `internal/app`:

```
cfg.Auth.Users + cfg.Identities + cfg.Auth.ExtensionPasswordFiles
  → mappedUsers()                      (app.go — the one join)
  → authmap.New(...)                   (immutable snapshot)
      → userauth (passkey layer)       holds *authmap.Map
      → SSO runtime config             mappedUsers() again
      → AdminService() → admin.New    holds its own *authmap.Map
```

Every REQUEST-side read already resolves live against that snapshot
(`MappedIdentityWith` in shell_seam.go, `server.isAdminSession`,
whoami, the switcher, Settings' account section) — the 2026-11-04
admin precedent ("a demotion bites on the next request") works
because resolution is per-request. But the SNAPSHOT itself is
boot-pinned: editing the deployment's config and re-activating the
NixOS module restarts the unit, and a restart drops every live SIP
registration and call. The island never unloads by design; the
process behind it should not have to either, just because an
operator moved a line between two humans.

What already reloads by construction (no work): SIP directory
passwords are read per login (`authmap.SIPPassword`, never cached),
and FreeSWITCH owns the credentials anyway — minted sessions stay
valid until expiry regardless of what the map says afterwards.

## The re-join seam (the actual design problem)

`authmap.Map` is immutable after `New`; every consumer holds a
`*Map` for its lifetime, so a reload that builds a second Map is
invisible to them. The seam needs ONE indirection:

- **`MapSource`** — a tiny holder over `atomic.Pointer[authmap.Map]`
  (and the identities map beside it, same swap). One home in
  `internal/authmap`.
- All four boot consumers (`userauth`, the SSO runtime,
  `AdminService`, plus `identityFor`'s identities feed) hold the
  SOURCE, not the Map, and resolve per request — they already do;
  only the field they read from changes.
- Reload = build the new snapshot OFFLINE (same validation as
  boot), then swap the pointer. Readers see old-then-new with no
  lock; no request ever observes a half-built map.

## Validation doctrine (carried over verbatim)

The reload runs the SAME checks boot runs (`authmap.New` +
`AdminService`'s admin-doctrine). On failure the reload is REFUSED,
the old snapshot keeps serving, and one loud WARN names the reason —
a bad edit must never drop the process, because the restart is the
thing we are avoiding. `extension_password_files` re-validates the
inventory (every mapped extension still has a readable file).

## Live-session semantics after a swap

- A session whose line left its user: `MappedIdentityWith` returns
  ok=false on the next request → the surface degrades to the
  classic extension login (Settings' unmapped branch already
  synthesizes `Line{ext, identityFor(ext)}`; the switcher 403s).
  No session is revoked mid-flight; the cookie expires on its own
  schedule. Revocation-on-demotion is a SEPARATE owner call and is
  deliberately not assumed here.
- An admin demoted in the edit: loses the Admin tab on the next
  request (existing doctrine, already tested).
- A NEW line/user: usable at the next login; no push, no nudge
  needed (nothing server-rendered caches the inventory beyond the
  snapshot).

## The owner call: trigger mechanism

| Arm            | Shape                                                                                                                                           | For                                                                                                                                                       | Against                                                                                                                                                                                                      |
| -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **SIGHUP**     | `signal.Notify(SIG_HUP)` → reload; module wires `ExecReload = /bin/kill -HUP $MAINPID`                                                          | Explicit operator intent; zero deps; atomic w.r.t. config activation (systemd reload happens AFTER the file landed); matches postgres/nginx muscle memory | Operator must run `systemctl reload webphone` (or it rides `activationModules`) — forgetting means stale map with no signal                                                                                  |
| **file-watch** | fsnotify on the config file's DIRECTORY (NixOS activates configs by atomic RENAME — watching the file itself misses the swap) + ~250ms debounce | Zero operator action; the map is always current                                                                                                           | New dependency; half-written-file windows (foreign deployments without rename semantics); surprise factor — a config edit visibly changing auth mid-session; needs the debounce + rename dance to be correct |

**Sketch recommendation (not a ruling):** SIGHUP first. It is the
smaller, honest arm — ownership edits are rare operator acts, and
`systemctl reload` after `nixos-rebuild switch` is a teachable
one-liner the module can print in its activation message.
File-watch is additive later if the owner finds the reload step
annoys them; the `MapSource` seam serves both arms unchanged.

## Explicitly OUT of scope

- Hot-reloading any OTHER config plane (gateway, seams, addr,
  language) — those own sockets, clients and loops; the ownership
  plane owns none. A general "config hot-reload" is the
  cohesive-config question (19-21 report §g), which this sketch
  feeds but does not answer.
- Revoking live sessions on demotion (separate owner call above).
- The Android shell's cached line list — it re-reads the session
  identity at its own cadence; the wire already carries `lines[]`.

## Wiring sketch (for the implementing train, when ruled)

1. `authmap.NewSource(...)` + `(*Source).Snapshot()`; `Map` gains
   nothing (stays immutable).
2. `internal/app`: the three consumers take the source; the reload
   routine lives in `cmd/webphone` next to the shutdown plumbing
   (process concern), calling an exported `App.ReloadOwnership()`.
3. Tests: swap-under-request race spec (the admin-demotion pattern
   generalized), refused-reload-keeps-serving spec, and a smoke
   `ownership-reload` scenario (HUP the loopback boot, assert the
   account section relabels).
4. NixOS module: `ExecReload` + `reloadTriggers` on the config
   file's ownership-relevant keys so `nixos-rebuild switch` reloads
   instead of restarting when ONLY those keys changed.

**Effort:** M (SIGHUP arm, seam + tests + module) — matches the
TODO_LIST row's "M (design done) / L (implement)".
