# The webphone-module check: evaluate the NixOS module with a minimal
# config and build the artifacts it would generate — catches
# option/syntax breakage without a full NixOS evaluation.
{
  perSystem =
    {
      lib,
      pkgs,
      self',
      ...
    }:
    {
      checks.webphone-module =
        let
          # The NixOS caddy/systemd/users stand-ins shared by the
          # base evaluation and the variants below.
          moduleSet = extra: {
            modules = [
              { _module.args.pkgs = pkgs; }
              # Minimal stand-ins for the NixOS caddy/systemd/users
              # modules: the webphone module writes services.caddy,
              # systemd.services, and users.{users,groups} config.
              {
                options = {
                  services.caddy = {
                    enable = lib.mkOption {
                      type = lib.types.bool;
                      default = false;
                    };
                    virtualHosts = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                  };
                  systemd.services = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                  # The backup timer writes systemd.timers; the
                  # stand-in must accept it like services.
                  systemd.timers = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                  users.users = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                  users.groups = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                  # NixOS's modules.nix normally provides this.
                  assertions = lib.mkOption {
                    type = lib.types.listOf lib.types.anything;
                    default = [ ];
                  };
                };
              }
            ]
            ++ [
              (import ../package/nixos-module.nix)
              {
                # recursiveUpdate, not `//`: extras like the HSTS
                # variant nest deeper (services.webphone.caddy.hsts)
                # and a shallow merge would drop the base attrs.
                services.webphone = lib.recursiveUpdate {
                  enable = true;
                  package = self'.packages.webphone;
                  caddy.enable = true;
                  caddy.hostName = "phone.example.org";
                  settings.sip_domain = "pbx.example.org";
                } extra;
              }
            ];
          };
          evaluated = lib.evalModules (moduleSet { });
          cfg = evaluated.config.services.webphone;
          vhost = evaluated.config.services.caddy.virtualHosts."phone.example.org";
          # Every streaming-critical directive the DOM/SSE contract rides
          # on must be present in the generated extraConfig: the
          # unbuffered /events handle, compression, and the reverse_proxy
          # to the app's listen port. The SIP WebSocket bridge renders
          # only when caddy.sipUpstream is set (the app never terminates
          # the wss — pinned by the vhost-sip-bridge case below); probe
          # fencing is a remote_ip matcher concern (Caddy has no
          # per-location blocks), so no probe handles are asserted here.
          requiredExtras = [
            "handle /events"
            "flush_interval -1"
            "encode zstd gzip"
            "reverse_proxy 127.0.0.1:${lib.last (lib.splitString ":" cfg.settings.addr)}"
          ];
          missingExtras = lib.filter (frag: !lib.hasInfix frag vhost.extraConfig) requiredExtras;
          unitPresent = evaluated.config.systemd.services ? "webphone";
        in
        pkgs.linkFarm "webphone-module-check" [
          {
            name = "webphone-config.json";
            path = (pkgs.formats.json { }).generate "webphone-config.json" cfg.settings;
          }
          {
            name = "listen-port";
            path = pkgs.writeText "listen-port" (lib.last (lib.splitString ":" cfg.settings.addr));
          }
          {
            name = "vhost-extras";
            path = pkgs.writeText "vhost-extras" (
              if missingExtras == [ ] then
                "all streaming-critical Caddy directives present"
              else
                throw "webphone-module check: vhost extraConfig missing: ${toString missingExtras}"
            );
          }
          {
            # caddy.sipUpstream gates the SIP WebSocket bridge: null (the
            # default) must render NO /sip handle (the app cannot
            # terminate the wss — proxying it there would dead-end every
            # call), a set upstream must render the bridge handle.
            name = "vhost-sip-bridge";
            path = pkgs.writeText "vhost-sip-bridge" (
              let
                sipHandle = "handle ${cfg.settings.websocket_path}";
                bridgedEvaluated = lib.evalModules (moduleSet {
                  caddy.sipUpstream = "https://pbx.example.org:7443";
                });
                bridgedVhost = bridgedEvaluated.config.services.caddy.virtualHosts."phone.example.org";
              in
              if
                !lib.hasInfix sipHandle vhost.extraConfig
                && lib.hasInfix sipHandle bridgedVhost.extraConfig
                && lib.hasInfix "reverse_proxy https://pbx.example.org:7443" bridgedVhost.extraConfig
              then
                "sip bridge renders exactly when sipUpstream is set"
              else
                throw "webphone-module check: caddy.sipUpstream did not gate the SIP bridge handle correctly"
            );
          }
          {
            name = "systemd-unit";
            path = pkgs.writeText "systemd-unit" (
              if unitPresent then
                "systemd.services.webphone present"
              else
                throw "webphone-module check: systemd.services.webphone missing"
            );
          }
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
          {
            # backup.enable must render the timer (OnCalendar + Unit)
            # and the oneshot service.
            name = "backup-timer";
            path = pkgs.writeText "backup-timer" (
              let
                backupEvaluated = lib.evalModules (moduleSet {
                  backup.enable = true;
                });
                timer = backupEvaluated.config.systemd.timers.webphone-backup;
                service = backupEvaluated.config.systemd.services.webphone-backup;
              in
              if
                timer.timerConfig.OnCalendar == "*-*-* 04:30:00"
                && timer.timerConfig.Unit == "webphone-backup.service"
                && timer.wantedBy == [ "timers.target" ]
                && service.serviceConfig.Type == "oneshot"
              then
                "backup timer + oneshot rendered"
              else
                throw "webphone-module check: backup.enable did not render the timer/oneshot pair"
            );
          }
          {
            # backup.retentionDays: null (default) must render the
            # plain single-snapshot script; a number must add the
            # dated-history branch (snapshots dir, link-dest basis,
            # bounded prune).
            name = "backup-retention";
            path = pkgs.writeText "backup-retention" (
              let
                plainScript =
                  (lib.evalModules (moduleSet {
                    backup.enable = true;
                  })).config.systemd.services.webphone-backup.script;
                retentionScript =
                  (lib.evalModules (moduleSet {
                    backup.enable = true;
                    backup.retentionDays = 7;
                  })).config.systemd.services.webphone-backup.script;
              in
              if
                !lib.hasInfix "snapshots" plainScript
                && lib.hasInfix "snapshots" retentionScript
                && lib.hasInfix "--link-dest" retentionScript
                && lib.hasInfix "-mtime +7" retentionScript
              then
                "backup retention branch renders per option"
              else
                throw "webphone-module check: backup.retentionDays did not gate the history/prune script correctly"
            );
          }
          {
            # backup.destDir is policed like dataDir: an off-/var/lib
            # path must trip exactly the destDir assertion (and the
            # base backup eval must stay free of failed assertions).
            name = "backup-destdir-assertion";
            path = pkgs.writeText "backup-destdir-assertion" (
              let
                failedAssertionsOf =
                  extra:
                  let
                    evaled = lib.evalModules (moduleSet extra);
                  in
                  lib.filter (a: !a.assertion) evaled.config.assertions;
                bad = failedAssertionsOf {
                  backup.enable = true;
                  backup.destDir = "/tmp/webphone-backup";
                };
                clean = failedAssertionsOf { backup.enable = true; };
              in
              if lib.length bad == 1 && lib.length clean == 0 then
                "backup.destDir assertion fires exactly off-/var/lib"
              else
                throw "webphone-module check: backup.destDir assertion mis-fires (bad=${toString (lib.length bad)}, clean=${toString (lib.length clean)})"
            );
          }
          {
            # Eval-time pins for the private-data hardening (2026-09-29
            # nix-review train): BOTH units must keep UMask=0077 +
            # StateDirectoryMode=0750 so message threads, faxes,
            # voicemail blobs and the session DB never land
            # world-readable. A regression here fails the check, not
            # prod.
            name = "unit-hardening-pins";
            path = pkgs.writeText "unit-hardening-pins" (
              let
                mainUnit = evaluated.config.systemd.services.webphone.serviceConfig;
                backupUnit =
                  (lib.evalModules (moduleSet {
                    backup.enable = true;
                  })).config.systemd.services.webphone-backup.serviceConfig;
              in
              if
                mainUnit.UMask == "0077"
                && mainUnit.StateDirectoryMode == "0750"
                && backupUnit.UMask == "0077"
                && backupUnit.StateDirectoryMode == "0750"
              then
                "both units pin UMask=0077 + StateDirectoryMode=0750"
              else
                throw "webphone-module check: private-data hardening regressed (UMask/StateDirectoryMode)"
            );
          }
          {
            # dataDir gets the same dedicated assertion coverage destDir
            # has: an off-/var/lib path must trip exactly the dataDir
            # assertion, and the base eval stays assertion-clean.
            name = "datadir-assertion";
            path = pkgs.writeText "datadir-assertion" (
              let
                failedAssertionsOf =
                  extra:
                  let
                    evaled = lib.evalModules (moduleSet extra);
                  in
                  lib.filter (a: !a.assertion) evaled.config.assertions;
                bad = failedAssertionsOf {
                  dataDir = "/tmp/webphone-data";
                };
                clean = failedAssertionsOf { };
              in
              if lib.length bad == 1 && lib.length clean == 0 then
                "dataDir assertion fires exactly off-/var/lib"
              else
                throw "webphone-module check: dataDir assertion mis-fires (bad=${toString (lib.length bad)}, clean=${toString (lib.length clean)})"
            );
          }
          {
            # serverTiming.enable must set the env gate the middleware
            # reads; without it the environment key stays absent.
            name = "server-timing";
            path = pkgs.writeText "server-timing" (
              let
                timingEvaluated = lib.evalModules (moduleSet {
                  serverTiming.enable = true;
                });
                timingUnit = timingEvaluated.config.systemd.services.webphone.environment;
              in
              if timingUnit ? WEBPHONE_DEBUG_TIMING && timingUnit.WEBPHONE_DEBUG_TIMING == "1" then
                "server-timing env gate rendered"
              else
                throw "webphone-module check: serverTiming.enable did not set WEBPHONE_DEBUG_TIMING"
            );
          }
          {
            name = "hsts-opt-in";
            path = pkgs.writeText "hsts-opt-in" (
              let
                hstsEvaluated = lib.evalModules (moduleSet {
                  caddy.hsts.enable = true;
                });
                hstsVhost = hstsEvaluated.config.services.caddy.virtualHosts."phone.example.org";
              in
              if lib.hasInfix "Strict-Transport-Security" hstsVhost.extraConfig then
                "hsts header present when enabled"
              else
                throw "webphone-module check: caddy.hsts.enable did not produce an HSTS vhost header"
            );
          }
        ];
    };
}
