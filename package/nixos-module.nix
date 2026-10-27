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
#
# Structure (nix-review monolith guideline, files under ~300 lines):
# options.nix carries the core option surface; oauth2.nix and crm.nix
# carry the typed seams END TO END (options + assertions + settings +
# LoadCredential wiring); caddy-vhost.nix and backup-script.nix carry
# the two embedded programs. This file is the wiring that composes them.
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

  caddyVhost = import ./caddy-vhost.nix { inherit lib cfg listenPort; };
  backupScript = import ./backup-script.nix { inherit lib cfg; };

  # `settings` accepts Nix paths (types.json), and a path literal is
  # COPIED into the world-readable store when the config renders — a
  # silent secret leak. Every secret-file setting must be a runtime-file
  # STRING (a string already pointing into the store is equally wrong).
  # builtins.isPath keeps this check from coercing (and thereby copying)
  # the offending value itself.
  leaksSecretToStore =
    value:
    if builtins.isPath value then
      true
    else if builtins.isString value then
      lib.hasPrefix "/nix/store/" value
    else
      false;
  secretFilePaths = [
    (cfg.settings.auth.oauth2.client_secret_file or "")
  ]
  ++ lib.attrValues (cfg.settings.auth.extension_password_files or { })
  ++ lib.attrValues (cfg.settings.auth.passkey.extension_password_files or { })
  ++ lib.optional (cfg.crm.enable && cfg.crm.tokenFile != null) cfg.crm.tokenFile;
  leakedSecretFiles = lib.count leaksSecretToStore secretFilePaths;
in
{
  imports = [
    ./options.nix
    ./oauth2.nix
    ./crm.nix
  ];

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
    ]
    # The oauth2/crm seams carry their own eval-time mirrors of the
    # app's fail-closed boot validation (in oauth2.nix / crm.nix).
    # Secret leak guard: Nix path VALUES in settings (accepted by the
    # freeform JSON type) are copied into the world-readable store when
    # the config renders, and a string already pointing there is just as
    # wrong. Always an error, independent of the OIDC/CRM mode, so it
    # stays HERE — it spans every seam.
    ++ lib.optional (leakedSecretFiles > 0) {
      assertion = false;
      message = ''
        a secret file setting under services.webphone is a Nix path
        (or a /nix/store string): path values are copied into the
        world-readable Nix store when the config renders. Point
        auth.oauth2.client_secret_file, the extension_password_files and
        crm.tokenFile at runtime files as plain strings
        (e.g. /run/secrets/...).
      '';
    };

    services.webphone.settings = {
      data_dir = lib.mkDefault cfg.dataDir;
      # Fronted shape: the generated vhost terminates TLS, so the browser's
      # Origin is https://<hostName> while the listener sees plain HTTP from
      # the local Caddy. Without these the CSRF middleware reads the truthful
      # Origin as a forged same-origin attestation and 403s every POST.
      # Precedence per key (pinned by the flake check's csrf-conflict case):
      # typed csrf.* (mkForce) > raw settings.csrf.* (plain) > the
      # caddy-derived defaults (mkDefault).
      # The auth/oauth2 and crm lanes are contributed by their seam files.
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
            # The OIDC client secret and the CRM bearer token ride
            # LoadCredential entries contributed by oauth2.nix / crm.nix:
            # systemd copies the operator's file into the unit's private
            # credential directory at start (0400, owned by the webphone
            # user), so the app reads the copy and no host-side group
            # grant is needed. A missing source file fails the unit
            # start — earlier and louder than the first login.
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
            # Sandbox parity with the main unit (the oneshot reads the
            # same private data): everything the backup needs is local
            # files — AF_UNIX only, no capabilities, no exotic syscalls
            # (sqlite3/rsync/date/find all sit inside @system-service).
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
            RestrictAddressFamilies = [ "AF_UNIX" ];
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
