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

  configFile = pkgs.generateText "webphone-config.json" {
    inherit (cfg) settings;
  };

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
    services.webphone.settings.data_dir = lib.mkDefault cfg.dataDir;

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
      };
    };

    services.nginx = lib.mkIf cfg.nginx.enable {
      enable = lib.mkDefault true;
      recommendedProxySettings = lib.mkDefault true;
      virtualHosts.${cfg.nginx.hostName} = {
        locations."/" = {
          recommendedProxySettings = true;
          proxyWebsockets = false;
          proxyPass = "http://127.0.0.1:${listenPort}";
        };
        # The SIP WebSocket path: upgrade + no read timeout (calls are
        # long-lived; the island's reconnect watchdog handles drops).
        locations.${cfg.settings.websocket_path} = {
          proxyPass = "http://127.0.0.1:${listenPort}";
          proxyWebsockets = true;
          extraConfig = ''
            proxy_read_timeout 3600s;
          '';
        };
      };
    };
  };
}
