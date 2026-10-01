# csrf.* precedence cases for the webphone-module check: the caddy
# fronting defaults, the typed csrf.* options, and the raw→typed→default
# precedence lane. Extracted from module-check.nix (monolith split).
{ lib, pkgs, base }:
let
  inherit (base) moduleSet cfg;
in
[
  {
    name = "csrf-fronted-origin";
    path = pkgs.writeText "csrf-fronted-origin" (
      if
        cfg.settings.csrf.trusted_origins == [ "https://phone.example.org" ]
        && cfg.settings.csrf.trusted_proxies == [ "127.0.0.1" ]
      then
        "csrf fronting defaults present"
      else
        throw "webphone-module check: csrf fronting defaults missing from the rendered settings"
    );
  }
  {
    # The typed csrf.* options must render into settings.csrf
    # and BEAT the caddy-derived defaults when set.
    name = "csrf-typed-override";
    path = pkgs.writeText "csrf-typed-override" (
      let
        typedEvaluated = lib.evalModules (moduleSet {
          csrf.trustedProxies = [ "10.9.8.7" ];
          csrf.trustedOrigins = [ "https://alt.example.org" ];
        });
        typedCfg = typedEvaluated.config.services.webphone;
      in
      if
        typedCfg.settings.csrf.trusted_proxies == [ "10.9.8.7" ]
        && typedCfg.settings.csrf.trusted_origins == [ "https://alt.example.org" ]
      then
        "typed csrf options render and override"
      else
        throw "webphone-module check: typed csrf options did not render into settings.csrf over the caddy defaults"
    );
  }
  {
    # Precedence under conflict, pinned per key: typed csrf.*
    # (mkForce) > raw settings.csrf.* (plain) > the caddy
    # defaults (mkDefault). One case exercises all three
    # lanes: typed proxies beat the raw proxy, and the raw
    # origin (typed origins empty) beats the caddy default.
    name = "csrf-conflict-precedence";
    path = pkgs.writeText "csrf-conflict-precedence" (
      let
        conflictEvaluated = lib.evalModules (moduleSet {
          csrf.trustedProxies = [ "10.9.8.7" ];
          settings.csrf = {
            trusted_proxies = [ "192.0.2.1" ];
            trusted_origins = [ "https://raw.example.org" ];
          };
        });
        conflictCfg = conflictEvaluated.config.services.webphone;
      in
      if
        conflictCfg.settings.csrf.trusted_proxies == [ "10.9.8.7" ]
        && conflictCfg.settings.csrf.trusted_origins == [ "https://raw.example.org" ]
      then
        "csrf conflict precedence: typed > raw > caddy default"
      else
        throw "webphone-module check: csrf conflict precedence broken (expected typed proxies and raw origins to win their lanes)"
    );
  }
]
