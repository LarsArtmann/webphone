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

    memoryMax = lib.mkOption {
      type = with lib.types; nullOr str;
      default = null;
      example = "512M";
      description = ''
        Memory cap for the service (systemd MemoryMax, e.g. "512M").
        null leaves the service uncapped.
      '';
    };

    nginx = {
      enable = lib.mkEnableOption "an nginx vhost that terminates TLS and proxies HTTP and the SIP WebSocket";
      hostName = lib.mkOption {
        type = lib.types.str;
        example = "phone.example.org";
        description = "Virtual host name for the generated nginx vhost.";
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
      csrf = lib.mkIf cfg.nginx.enable {
        trusted_proxies = lib.mkDefault [ "127.0.0.1" ];
        trusted_origins = lib.mkDefault [ "https://${cfg.nginx.hostName}" ];
      };
    };

    users.users.webphone = {
      isSystemUser = true;
      group = "webphone";
      description = "webphone service user";
    };
    users.groups.webphone = { };

    systemd.services.webphone = {
      description = "webphone unified communications (calls, messages, fax, voicemail)";
      wantedBy = [ "multi-user.target" ];
      after = [ "network-online.target" ];
      wants = [ "network-online.target" ];

      environment = {
        WEBPHONE_CONFIG = configFile;
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

    services.nginx = lib.mkIf cfg.nginx.enable {
      enable = lib.mkDefault true;
      recommendedProxySettings = lib.mkDefault true;
      virtualHosts.${cfg.nginx.hostName} = {
        locations = {
          "/" = {
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
