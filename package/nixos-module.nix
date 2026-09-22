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

  # "127.0.0.1:8080" / ":8080" → "8080" (the nginx upstream needs the port).
  listenPort = lib.last (lib.splitString ":" cfg.settings.addr);
in
{
  options.services.webphone = {
    enable = lib.mkEnableOption "webphone, the self-hosted unified-communications web app (calls, SMS/MMS, fax, voicemail)";

    package = lib.mkOption {
      type = lib.types.package;
      description = ''
        The webphone package. Point this at the flake's package:
        `inputs.webphone.packages.\${pkgs.system}.webphone`.
      '';
    };

    dataDir = lib.mkOption {
      type = lib.types.str;
      default = "/var/lib/webphone";
      description = "State directory for the SQLite database and the blob store.";
    };

    # Free-form mapping rendered to the JSON config file; every nested key
    # the server understands is allowed (see README config reference).
    settings = lib.mkOption {
      type = lib.types.submodule {
        freeformType = (pkgs.formats.json { }).type;
        options = {
          addr = lib.mkOption {
            type = lib.types.str;
            default = "127.0.0.1:8080";
            description = "Listen address of the HTTP server.";
          };
          websocket_path = lib.mkOption {
            type = lib.types.str;
            default = "/sip";
            description = "Path the SIP WebSocket is served on (the nginx vhost proxies it).";
          };
        };
      };
      default = { };
      description = "Structured webphone configuration (rendered to JSON).";
    };

    environmentFile = lib.mkOption {
      type = with lib.types; nullOr path;
      default = null;
      description = ''
        EnvironmentFile loaded by systemd — the place for
        WEBPHONE_GATEWAY__WEBHOOK_SECRET and other secrets.
      '';
      example = "/run/secrets/webphone-env";
    };

    # Typed front for settings.csrf.trusted_*: same values the freeform
    # settings accept, but discoverable and checkable as module options.
    # Precedence per key: typed csrf.* (mkForce) beats a raw
    # settings.csrf.* value, which beats the nginx.enable defaults
    # (mkDefault). Empty typed lists never clobber raw values.
    csrf = {
      trustedProxies = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        example = [ "127.0.0.1" ];
        description = ''
          Proxies whose X-Forwarded-Proto header the CSRF middleware may
          believe (IP or CIDR entries). Renders into
          `settings.csrf.trusted_proxies`; wins over BOTH a raw
          `settings.csrf.trusted_proxies` value and the `nginx.enable`
          default when non-empty.
        '';
      };
      trustedOrigins = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        example = [ "https://phone.example.org" ];
        description = ''
          Browser-facing origins counted as same-origin by the CSRF
          middleware (the TLS vhost). Renders into
          `settings.csrf.trusted_origins`; wins over BOTH a raw
          `settings.csrf.trusted_origins` value and the `nginx.enable`
          default when non-empty.
        '';
      };
    };

    serverTiming = {
      enable = lib.mkEnableOption ''
        the Server-Timing response header (W3C `total;dur=…` on every
        response) for the live host: sets WEBPHONE_DEBUG_TIMING=1 in the
        unit environment, the same gate the middleware reads. Diagnostic
        only — header values are sanitized against CRLF injection and the
        timings ride the existing request log.
      '';
    };

    memoryMax = lib.mkOption {
      type = with lib.types; nullOr str;
      default = null;
      example = "512M";
      description = ''
        Memory cap for the service (systemd MemoryMax, e.g. "512M").
        null leaves the service uncapped.
      '';
    };

    backup = {
      enable = lib.mkEnableOption ''
        a daily online-backup timer: sqlite's `.backup` API copies the
        database while the service keeps running (no phone downtime),
        and the blob tree is rsynced beside it. The skeleton writes
        into destDir and leaves retention and off-machine copies to
        the operator's existing backup tooling - it is a starting
        point, not a restic replacement. Restores were drill-verified:
        stop the service, copy db + blobs back, start it.
      '';
      destDir = lib.mkOption {
        type = lib.types.str;
        default = "/var/lib/webphone-backup";
        description = "Directory the daily backup snapshot lands in.";
      };
      calendar = lib.mkOption {
        type = lib.types.str;
        default = "*-*-* 04:30:00";
        example = "*-*-* *:00/15:00";
        description = "systemd OnCalendar schedule for the backup timer.";
      };
      retentionDays = lib.mkOption {
        type = lib.types.nullOr lib.types.ints.positive;
        default = null;
        example = 30;
        description = ''
          When set, the daily run also keeps a dated history: each run
          writes a snapshot under `<destDir>/snapshots/<date>/` (blob
          files hardlinked against the previous snapshot, so unchanged
          blobs cost no extra space) and deletes snapshot directories
          whose modification time is older than this many days.
          null (the default) keeps only the single latest snapshot
          under destDir — the historical behavior. Point-in-time
          restore: stop the service, copy db + files back from the
          chosen snapshot directory, start the service.
        '';
      };
    };

    nginx = {
      enable = lib.mkEnableOption ''
        an nginx vhost that terminates TLS and proxies HTTP and the SIP
        WebSocket. Enabling it also defaults settings.csrf to trust the
        loopback proxy and the https://<hostName> origin; without that
        fronting shape (or a hand-rolled equivalent in settings.csrf)
        the CSRF middleware rejects every browser POST behind TLS.'';
      hostName = lib.mkOption {
        type = lib.types.str;
        example = "phone.example.org";
        description = "Virtual host name for the generated nginx vhost.";
      };
      hsts = {
        enable = lib.mkEnableOption ''
          Strict-Transport-Security on the generated vhost. Default off:
          HSTS pins browsers to https for maxAge seconds, so flipping it
          on before the deployment is genuinely https-only (ACME working,
          no http-only tooling left) can brick the domain for that window.'';
        maxAge = lib.mkOption {
          type = lib.types.ints.positive;
          default = 63072000;
          example = 31536000;
          description = "HSTS max-age in seconds (default 2 years, the common baseline).";
        };
      };
    };
  };

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
    ];

    services.webphone.settings = {
      data_dir = lib.mkDefault cfg.dataDir;
      # Fronted shape: the generated vhost terminates TLS, so the browser's
      # Origin is https://<hostName> while the listener sees plain HTTP from
      # the local nginx. Without these the CSRF middleware reads the truthful
      # Origin as a forged same-origin attestation and 403s every POST.
      # Precedence per key (pinned by the flake check's csrf-conflict case):
      # typed csrf.* (mkForce) > raw settings.csrf.* (plain) > the
      # nginx-derived defaults (mkDefault).
      csrf = lib.mkMerge [
        (lib.mkIf cfg.nginx.enable {
          trusted_proxies = lib.mkDefault [ "127.0.0.1" ];
          trusted_origins = lib.mkDefault [ "https://${cfg.nginx.hostName}" ];
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
            EnvironmentFile = lib.mkIf (cfg.environmentFile != null) [ cfg.environmentFile ];
            User = "webphone";
            Group = "webphone";
            StateDirectory = builtins.replaceStrings [ "/var/lib/" ] [ "" ] cfg.dataDir;

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
          serviceConfig = {
            Type = "oneshot";
            User = "webphone";
            Group = "webphone";
            StateDirectory = builtins.replaceStrings [ "/var/lib/" ] [ "" ] cfg.backup.destDir;
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
          script =
            let
              dest = cfg.backup.destDir;
              history = cfg.backup.retentionDays != null;
            in
            ''
              sqlite3 ${cfg.dataDir}/webphone.db ".backup '${dest}/webphone.db'"
              rsync -a --delete ${cfg.dataDir}/files/ ${dest}/files/
            ''
            + (lib.optionalString history ''
              # Dated history: snapshot today's state under snapshots/,
              # hardlinking unchanged blobs against the newest previous
              # snapshot (first run has no basis — plain copy; a same-day
              # rerun never uses itself as the basis).
              today=$(date +%F)
              snap="${dest}/snapshots/$today"
              basis=$(
                find ${dest}/snapshots -mindepth 1 -maxdepth 1 -type d \
                  -name '????-??-??' ! -name "$today" 2>/dev/null | sort | tail -1 || true
              )
              mkdir -p "$snap"
              sqlite3 ${cfg.dataDir}/webphone.db ".backup '$snap/webphone.db'"
              if [ -n "$basis" ]; then
                rsync -a --delete --link-dest="$basis" ${cfg.dataDir}/files/ "$snap/files/"
              else
                rsync -a --delete ${cfg.dataDir}/files/ "$snap/files/"
              fi
              # Prune: only the dated directories themselves, never the
              # latest top-level snapshot pair, and never anything that
              # is not a YYYY-MM-DD name.
              find ${dest}/snapshots -mindepth 1 -maxdepth 1 -type d \
                -name '????-??-??' -mtime +${toString cfg.backup.retentionDays} \
                -exec rm -rf -- {} +
            '');
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

    services.nginx = lib.mkIf cfg.nginx.enable {
      enable = lib.mkDefault true;
      recommendedProxySettings = lib.mkDefault true;
      virtualHosts.${cfg.nginx.hostName} = {
        extraConfig = lib.mkIf cfg.nginx.hsts.enable ''
          add_header Strict-Transport-Security "max-age=${toString cfg.nginx.hsts.maxAge}" always;
        '';
        locations = {
          # "/" carries the whole app. The JSON probe endpoints have their
          # OWN locations below so a fleet health hub can be allowlisted or
          # restricted per location without touching the app's.
          "/" = {
            recommendedProxySettings = true;
            proxyWebsockets = false;
            proxyPass = "http://127.0.0.1:${listenPort}";
          };
          # The probe triple, as dedicated locations: GET /healthz
          # (readiness: sqlite + blob-dir, bounded checks), GET /livez
          # (process liveness, fetch-free) and GET /startupz (503 until the
          # backing resources first pass, then latched). All three are
          # session-free GETs whose bodies name checks and statuses only,
          # never secrets — safe to expose or scrape by a fleet health hub;
          # override one with extraConfig (allow/deny) to fence scrapers.
          "/healthz" = {
            recommendedProxySettings = true;
            proxyWebsockets = false;
            proxyPass = "http://127.0.0.1:${listenPort}";
          };
          "/livez" = {
            recommendedProxySettings = true;
            proxyWebsockets = false;
            proxyPass = "http://127.0.0.1:${listenPort}";
          };
          "/startupz" = {
            recommendedProxySettings = true;
            proxyWebsockets = false;
            proxyPass = "http://127.0.0.1:${listenPort}";
          };
          # The SIP WebSocket path: upgrade + no read timeout (calls are
          # long-lived; the island's reconnect watchdog handles drops).
          ${cfg.settings.websocket_path} = {
            proxyPass = "http://127.0.0.1:${listenPort}";
            recommendedProxySettings = true;
            proxyWebsockets = true;
            extraConfig = ''
              proxy_read_timeout 3600s;
            '';
          };
          # Server-sent events: unbuffered, HTTP/1.1, long read timeout so
          # the event stream stays open for the whole session.
          "/events" = {
            proxyPass = "http://127.0.0.1:${listenPort}";
            recommendedProxySettings = true;
            extraConfig = ''
              proxy_buffering off;
              proxy_read_timeout 3600s;
              proxy_http_version 1.1;
            '';
          };
        };
      };
    };
  };
}
