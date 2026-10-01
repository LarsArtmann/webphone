# NixOS module for the webphone service (Go binary: HTTP + SIP-over-WSS).
#
# Deployment decision (2026-09-18, resolves the open TODO_LIST question):
# this repo SHIPS the module so the binary and its deployment shape stay
# in sync. The consuming nix-international-telephony stack keeps both
# options: import this module, or keep reverse-proxying to the binary
# itself (the README's documented default) — the module is additive.
#
# Secrets: keep gateway.webhook_secret out of the world-readable config
# JSON; put WEBPHONE_GATEWAY__WEBHOOK_SECRET into an EnvironmentFile
# (services.webphone.environmentFile) instead.
{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.webphone;

  configFile = (pkgs.formats.json { }).generate "webphone-config.json" cfg.settings;

  # "127.0.0.1:8080" / ":8080" → "8080" (the Caddy upstream needs the port).
  listenPort = lib.last (lib.splitString ":" cfg.settings.addr);

  # The two embedded programs — the Caddy vhost body and the backup
  # shell script — live in their own files so this module's wiring
  # stays readable (nix-review monolith guideline).
  caddyVhost = import ./caddy-vhost.nix { inherit lib cfg listenPort; };
  backupScript = import ./backup-script.nix { inherit lib cfg; };
in
{
  imports = [ ./options.nix ];

  config = lib.mkIf cfg.enable {
    assertions = [
      {
        assertion = lib.hasPrefix "/var/lib/" cfg.dataDir && cfg.dataDir != "/var/lib/";
        message = ''
          services.webphone.dataDir must name a directory under /var/lib/
          (got `${cfg.dataDir}`): systemd's StateDirectory manages exactly that
          tree, and anything else would yield an invalid relative
          StateDirectory value for the webphone unit.
        '';
      }
    ]
    ++ lib.optionals cfg.backup.enable [
      {
        assertion = lib.hasPrefix "/var/lib/" cfg.backup.destDir && cfg.backup.destDir != "/var/lib/";
        message = ''
          services.webphone.backup.destDir must name a directory under /var/lib/
          (got `${cfg.backup.destDir}`): the backup unit's StateDirectory manages
          exactly that tree, and anything else would yield an invalid relative
          StateDirectory value for the webphone-backup unit.
        '';
      }
    ];

    services.webphone.settings = {
      data_dir = lib.mkDefault cfg.dataDir;
      # Fronted shape: the generated vhost terminates TLS, so the browser's
      # Origin is https://<hostName> while the listener sees plain HTTP from
      # the local Caddy. Without these the CSRF middleware reads the truthful
      # Origin as a forged same-origin attestation and 403s every POST.
      # Precedence per key (pinned by the flake check's csrf-conflict case):
      # typed csrf.* (mkForce) > raw settings.csrf.* (plain) > the
      # caddy-derived defaults (mkDefault).
      csrf = lib.mkMerge [
        (lib.mkIf cfg.caddy.enable {
          trusted_proxies = lib.mkDefault [ "127.0.0.1" ];
          trusted_origins = lib.mkDefault [ "https://${cfg.caddy.hostName}" ];
        })
        (lib.mkIf (cfg.csrf.trustedProxies != [ ]) {
          trusted_proxies = lib.mkForce cfg.csrf.trustedProxies;
        })
        (lib.mkIf (cfg.csrf.trustedOrigins != [ ]) {
          trusted_origins = lib.mkForce cfg.csrf.trustedOrigins;
        })
      ];
    };

    users.users.webphone = {
      isSystemUser = true;
      group = "webphone";
      description = "webphone service user";
    };
    users.groups.webphone = { };

    systemd = {
      services = {
        # Deliberately Type=simple (the default): readiness is the
        # /startupz latch, not sd_notify — see README "Readiness vs
        # systemd". Do not add Type=notify or watchdog restarts.
        webphone = {
          description = "webphone unified communications (calls, messages, fax, voicemail)";
          wantedBy = [ "multi-user.target" ];
          after = [ "network-online.target" ];
          wants = [ "network-online.target" ];

          environment = {
            WEBPHONE_CONFIG = configFile;
          }
          // lib.optionalAttrs cfg.serverTiming.enable {
            WEBPHONE_DEBUG_TIMING = "1";
          };

          serviceConfig = {
            ExecStart = lib.getExe cfg.package;
            EnvironmentFile = lib.mkIf (cfg.environmentFile != null || cfg.environmentFiles != [ ]) (
              (lib.optional (cfg.environmentFile != null) cfg.environmentFile) ++ cfg.environmentFiles
            );
            User = "webphone";
            Group = "webphone";
            StateDirectory = builtins.replaceStrings [ "/var/lib/" ] [ "" ] cfg.dataDir;
            # Private data (message threads, faxes, voicemail blobs): files
            # the app creates stay owner-only and the state dir drops to
            # 0750 — the group is webphone-only and Caddy never reads here.
            StateDirectoryMode = "0750";
            UMask = "0077";

            # Hardening: the service needs almost nothing — network, its state
            # directory, and nothing else.
            NoNewPrivileges = true;
            PrivateTmp = true;
            PrivateDevices = true;
            ProtectSystem = "strict";
            ProtectHome = true;
            ProtectClock = true;
            ProtectHostname = true;
            ProtectKernelLogs = true;
            ProtectKernelModules = true;
            ProtectKernelTunables = true;
            ProtectControlGroups = true;
            ProtectProc = "invisible";
            RestrictAddressFamilies = [
              "AF_INET"
              "AF_INET6"
              "AF_UNIX"
            ];
            RestrictNamespaces = true;
            RestrictRealtime = true;
            RestrictSUIDSGID = true;
            LockPersonality = true;
            MemoryDenyWriteExecute = true;
            CapabilityBoundingSet = "";
            SystemCallArchitectures = "native";
            SystemCallFilter = [
              "@system-service"
              "~@privileged @resources"
            ];
            Restart = "on-failure";
            RestartSec = 5;
            MemoryMax = lib.mkIf (cfg.memoryMax != null) cfg.memoryMax;
          };
        };

        webphone-backup = lib.mkIf cfg.backup.enable {
          description = "webphone online backup (sqlite .backup + blob rsync)";
          # Order after the phone service when both start together (a boot
          # with a Persistent timer catch-up fire): the oneshot reads the
          # live webphone.db and must not race its creation.
          after = [ "webphone.service" ];
          serviceConfig = {
            Type = "oneshot";
            User = "webphone";
            Group = "webphone";
            StateDirectory = builtins.replaceStrings [ "/var/lib/" ] [ "" ] cfg.backup.destDir;
            StateDirectoryMode = "0750";
            UMask = "0077";
            NoNewPrivileges = true;
            PrivateTmp = true;
            PrivateDevices = true;
            ProtectSystem = "strict";
            ProtectHome = true;
            ReadWritePaths = [ cfg.backup.destDir ];
          };
          # Online copy: sqlite's .backup API takes a consistent snapshot
          # while the service runs, so the phone never restarts for a
          # backup (the cold-copy path is the drill-verified fallback).
          path = [
            pkgs.sqlite
            pkgs.rsync
          ];
          script = backupScript;
        };
      };

      timers.webphone-backup = lib.mkIf cfg.backup.enable {
        description = "Daily webphone online backup";
        wantedBy = [ "timers.target" ];
        timerConfig = {
          OnCalendar = cfg.backup.calendar;
          Persistent = true;
          Unit = "webphone-backup.service";
        };
      };
    };

    # The Caddy front: a catch-all reverse_proxy plus the paths that
    # need explicit treatment — the SIP WebSocket bridge (only when an
    # upstream is configured; the app never terminates the wss itself),
    # SSE (explicit flush_interval -1 so /events bytes reach the
    # browser unbuffered; the /health/* dashboard subtree streams the
    # same way), and the JSON probes (/healthz /livez /startupz
    # /metrics) which need no handle of their own: fence scrapers with
    # a remote_ip matcher in extraConfig instead of nginx-style
    # per-location blocks.
    services.caddy = lib.mkIf cfg.caddy.enable {
      enable = lib.mkDefault true;
      virtualHosts.${cfg.caddy.hostName}.extraConfig = caddyVhost;
    };
  };
}
