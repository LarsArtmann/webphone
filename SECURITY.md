# Security Policy

## Reporting a vulnerability

**Preferred (sensitive):** GitHub private vulnerability reporting — open
this repository, then **Security → Report a vulnerability**. It reaches
the maintainer privately, supports back-and-forth discussion, and can
publish a security advisory with the fix. No email or GPG setup needed.

**Only for non-sensitive, already-public problems:** open a regular
GitHub issue.

Please include: the version string the `/version` endpoint reports (or
the commit), the surface involved — the SIP.js call island, a
server-rendered tab, the `/phone-api` session proxy, the bundle, or the
NixOS module — and a minimal reproduction. Say what fronts the binary
in your deployment (nginx TLS + WSS proxy in the reference stack,
something else, or nothing): transport misconfigurations usually live
in the fronting proxy, not this service.

## Threat model and security posture

Facts live with their code; this section points, it does not duplicate:

- One binary, one page, one login: the PBX extension + directory
  password — there is no user table by design.
- Strict same-origin CSP; assets served in-repo, no CDN, no plugins.
- CSRF protection on all mutations, security headers, owner-scoped
  queries (see FEATURES.md, "Security posture").
- TURN REST credentials are derived per response from a server-side
  secret (`turn_rest` config; draft-uberti-behave-turn-rest) — the
  secret never ships to the browser.
- The NixOS module runs the service as a dedicated system user with
  `StateDirectoryMode = "0750"` and `UMask = "0077"`.

Transport security (TLS, WSS termination) belongs to the deploying
stack; the reference shape is documented in
[nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony).

## Supported versions

Only the latest tagged release receives fixes:
https://github.com/LarsArtmann/webphone/releases
Tags trail the binary's version literal; when in doubt, report against
the revision your consuming stack's `flake.lock` pins.
