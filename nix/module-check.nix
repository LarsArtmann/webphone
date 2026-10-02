# The webphone-module check: evaluate the NixOS module with a minimal
# config and build the artifacts it would generate — catches
# option/syntax breakage without a full NixOS evaluation. The eval
# helpers live in module-check-base.nix; the csrf and backup case
# groups live in their own files so each stays focused (nix-review
# monolith guideline: files under ~300 lines).
{
  perSystem =
    {
      lib,
      pkgs,
      self',
      ...
    }:
    let
      base = import ./module-check-base.nix { inherit lib pkgs self'; };
      inherit (base)
        moduleSet
        evaluated
        cfg
        vhost
        ;

      csrfCases = import ./module-check-csrf.nix { inherit lib pkgs base; };
      backupCases = import ./module-check-backup.nix { inherit lib pkgs base; };

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

      # Full-text rendering of the generated module output (Caddy vhost,
      # the retention=7 backup script, and the settings JSON). The
      # substring cases below catch missing pieces; the golden case
      # catches reordering/whitespace drift they would miss (the same
      # idea as the gateway part-header golden). The renderer lives in
      # nix/module-output.nix; the fixture is nix/module-output.golden.
      renderedModuleOutput = import ./module-output.nix { inherit lib base; };

      coreCases = [
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
        {
          # Full-text golden of the generated vhost + retention=7 backup
          # script + settings JSON. Substring checks miss reordering,
          # whitespace, or a dropped line; this pins the whole document.
          name = "module-output-golden";
          path = pkgs.runCommand "module-output-golden" { } ''
            if diff -u ${./module-output.golden} ${
              pkgs.writeText "module-output.actual" renderedModuleOutput
            } > $out; then
              echo "module output matches the committed golden (vhost + backup script + settings JSON)" >> $out
            else
              cat $out >&2
              echo "webphone-module check: module output drifted from nix/module-output.golden" >&2
              exit 1
            fi
          '';
        }
      ];
    in
    {
      checks.webphone-module = pkgs.linkFarm "webphone-module-check" (
        coreCases ++ csrfCases ++ backupCases
      );
    };
}
