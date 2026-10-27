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

  # --- OIDC/SSO wiring (services.webphone.oauth2) ---
  # The secret is handed to systemd as a credential; systemd exposes
  # credentials at $CREDENTIALS_DIRECTORY, which is the unit's
  # /run/credentials path. The app reads THIS copy, never the
  # operator's host file, so no group grant is required.
  oauth2CredentialName = "webphone-oidc-client-secret";
  oauth2SecretPath = "/run/credentials/webphone.service/${oauth2CredentialName}";

  # --- CRM wiring (services.webphone.crm) — same credential doctrine.
  crmCredentialName = "webphone-crm-token";
  crmSecretPath = "/run/credentials/webphone.service/${crmCredentialName}";

  # null = derive from the vhost; eval fails closed when neither is
  # available (an external front must declare the registration).
  oauth2RedirectUrl =
    if cfg.oauth2.redirectUrl != null then
      cfg.oauth2.redirectUrl
    else if cfg.caddy.enable then
      "https://${cfg.caddy.hostName}/api/auth/oauth2/callback"
    else
      null;

  # The identity map is shared by both passwordless modes and lives in
  # the app's ONE home for it (settings.auth.*, folded into the passkey
  # shape by the app). The module reads the same merged value so its
  # assertions reflect what the binary will actually see.
  oauth2Users = cfg.settings.auth.users or { };
  oauth2PasswordFiles =
    (cfg.settings.auth.extension_password_files or { })
    // (cfg.settings.auth.passkey.extension_password_files or { });
  oauth2MappedExtensions = lib.unique (
    lib.concatMap (user: user.extensions or [ ]) (lib.attrValues oauth2Users)
  );
  oauth2MissingPasswordFiles = lib.filter (
    ext: (oauth2PasswordFiles.${ext} or "") == ""
  ) oauth2MappedExtensions;

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

  # A typed OIDC config derives passkey mode in the app (the shared
  # users map feeds both modes), which then DEMANDS rp_id/rp_origins.
  # Behind the generated vhost they follow from caddy.hostName; a raw
  # value still wins (mkDefault), so an operator pin stays authoritative.
  # Deliberately keyed on the TYPED lane only: reading the merged
  # settings.auth.users from inside the settings definition would recurse.
  deriveWebAuthnRP = cfg.caddy.enable && cfg.oauth2.enable;

  bareHTTPS = url: lib.match "https://[^/?#]+(/[^?#]*)?" url != null;
in
{
  imports = [
    ./options.nix
    ./options-oauth2.nix
    ./options-crm.nix
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
    # Eval-time mirror of the app's fail-closed OIDC validation: an
    # enabled-but-half SSO config must fail `nixos-rebuild`, never a
    # reboot into a login card whose every attempt dies.
    ++ lib.optionals cfg.oauth2.enable [
      {
        assertion = cfg.oauth2.issuer != null && bareHTTPS cfg.oauth2.issuer;
        message = ''
          services.webphone.oauth2.issuer must be a bare https URL with no
          query or fragment (got `${
            if cfg.oauth2.issuer == null then "unset" else cfg.oauth2.issuer
          }`): the IdP's `iss` claim is matched verbatim.
        '';
      }
      {
        assertion = cfg.oauth2.clientSecretFile != null;
        message = ''
          services.webphone.oauth2.clientSecretFile is required when
          enable = true — the client secret never lives in the world-readable
          config JSON or the Nix store; point it at the runtime file.
        '';
      }
      {
        assertion = lib.elem "openid" cfg.oauth2.scopes;
        message = ''
          services.webphone.oauth2.scopes must include "openid" — without it
          the IdP returns no id_token and no login can complete.
        '';
      }
      {
        assertion = oauth2RedirectUrl != null && bareHTTPS oauth2RedirectUrl;
        message = ''
          services.webphone.oauth2: the callback cannot be derived — enable
          caddy (derives https://<caddy.hostName>/api/auth/oauth2/callback)
          or set redirectUrl to the registered absolute https callback.
        '';
      }
      {
        assertion = oauth2Users != { };
        message = ''
          services.webphone.oauth2.enable requires settings.auth.users (the
          shared email→extension mapping) — without it every SSO login is
          unmapped and denied.
        '';
      }
      {
        assertion =
          (cfg.settings.auth.passkey.rp_id or "") != ""
          && (cfg.settings.auth.passkey.rp_origins or [ ]) != [ ];
        message = ''
          services.webphone.oauth2.enable turns on the shared identity layer,
          which the app derives passkey mode from — settings.auth.passkey.rp_id
          and rp_origins must resolve. Enable caddy to derive them from
          caddy.hostName, or set both explicitly behind an external front.
        '';
      }
      {
        assertion = oauth2MissingPasswordFiles == [ ];
        message = ''
          services.webphone.oauth2: every mapped extension needs a SIP
          directory password file in settings.auth.extension_password_files
          (missing: ${lib.concatStringsSep ", " oauth2MissingPasswordFiles}).
        '';
      }
    ]
    ++ lib.optionals (!cfg.oauth2.enable) [
      {
        assertion =
          cfg.oauth2.issuer == null
          && cfg.oauth2.clientSecretFile == null
          && cfg.oauth2.redirectUrl == null
          && cfg.oauth2.allowedGroups == [ ];
        message = ''
          services.webphone.oauth2: set enable = true — issuer,
          clientSecretFile, redirectUrl or allowedGroups without enable would
          silently do nothing.
        '';
      }
    ]
    # Eval-time mirror of the app's fail-closed CRM validation: an
    # enabled-but-half CRM config must fail `nixos-rebuild`, never boot
    # a dark seam the operator believes is on.
    ++ lib.optionals cfg.crm.enable [
      {
        assertion = cfg.crm.url != null && bareHTTPS cfg.crm.url;
        message = ''
          services.webphone.crm.url must be an absolute https URL (got ${
            if cfg.crm.url == null then "unset" else cfg.crm.url
          }): the machine API answers 401 without TLS in any real
          deployment, and a relative URL cannot resolve.
        '';
      }
      {
        assertion = cfg.crm.tokenFile != null;
        message = ''
          services.webphone.crm.tokenFile is required when enable = true —
          the bearer token never lives in the world-readable config JSON or
          the Nix store; point it at the runtime file.
        '';
      }
    ]
    ++ lib.optionals (!cfg.crm.enable) [
      {
        assertion = cfg.crm.url == null && cfg.crm.tokenFile == null;
        message = ''
          services.webphone.crm: set enable = true — url or tokenFile
          without enable would silently do nothing.
        '';
      }
    ]
    # Secret leak guard: Nix path VALUES in settings (accepted by the
    # freeform JSON type) are copied into the world-readable store when
    # the config renders, and a string already pointing there is just as
    # wrong. Always an error, independent of the OIDC mode.
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
    }
    // lib.optionalAttrs cfg.oauth2.enable {
      auth = lib.mkMerge [
        (lib.mkIf deriveWebAuthnRP {
          # mkDefault: an explicitly declared settings.auth.passkey.*
          # value wins, so an operator pin stays authoritative.
          passkey = {
            rp_id = lib.mkDefault cfg.caddy.hostName;
            rp_origins = lib.mkDefault [ "https://${cfg.caddy.hostName}" ];
          };
        })
        (lib.mkIf cfg.oauth2.enable {
          # PLAIN definitions, deliberately: a raw
          # settings.auth.oauth2.* of the same key makes the eval fail
          # with a conflict (declare the OIDC client shape ONCE) instead
          # of one lane silently winning.
          oauth2 = {
            issuer = cfg.oauth2.issuer;
            client_id = cfg.oauth2.clientId;
            client_secret_file = oauth2SecretPath;
            redirect_url = oauth2RedirectUrl;
            scopes = cfg.oauth2.scopes;
            allowed_groups = cfg.oauth2.allowedGroups;
          };
        })
      ];
    }
    // lib.optionalAttrs cfg.crm.enable {
      # PLAIN definitions, deliberately: a raw settings.crm.* of the same
      # key makes the eval fail with a conflict (one home per fact), and
      # when the typed lane is OFF this block is absent entirely — the
      # raw lane stays untouched.
      crm = {
        url = cfg.crm.url;
        token_file = crmSecretPath;
      };
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
            # OIDC client secret: systemd copies the operator's file into
            # the unit's private credential directory at start (0400,
            # owned by the webphone user), so the app reads the copy and
            # no host-side group grant is needed. A missing source file
            # fails the unit start — earlier and louder than the first
            # SSO login.
            LoadCredential = lib.mkMerge [
              (lib.mkIf cfg.oauth2.enable [
                "${oauth2CredentialName}:${cfg.oauth2.clientSecretFile}"
              ])
              # CRM bearer token: the same credential doctrine — systemd
              # copies the operator's file into the unit's private
              # credential directory at start, the app reads the copy.
              (lib.mkIf cfg.crm.enable [
                "${crmCredentialName}:${cfg.crm.tokenFile}"
              ])
            ];
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
