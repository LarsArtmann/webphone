# The services.webphone option surface (the module's interface),
# split out of nixos-module.nix so the options and the config each
# stay under the nix-review monolith guideline. Imported by
# nixos-module.nix.
{ lib, pkgs, ... }:
{
  options.services.webphone = {
    enable = lib.mkEnableOption "webphone, the self-hosted unified-communications web app (calls, SMS/MMS, fax, voicemail)";

    package = lib.mkPackageOption pkgs "webphone" {
      # No default: the consumer points this at its own flake package
      # (there is no pkgs.webphone). `default = null` makes the option
      # required while mkPackageOption still supplies the type and the
      # "The webphone package to use." description.
      default = null;
      extraDescription = ''
        Point this at the flake's package:
        `inputs.webphone.packages.''${pkgs.system}.webphone`.
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
            description = "Path the SIP WebSocket is served on (the Caddy vhost proxies it).";
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

    environmentFiles = lib.mkOption {
      type = with lib.types; listOf path;
      default = [ ];
      description = ''
        Additional EnvironmentFiles loaded AFTER environmentFile — the
        seam a fronting module (e.g. the telephony stack) uses to inject
        its own rendered secrets (WEBPHONE_TURN_REST__SECRET,
        WEBPHONE_CRM__TOKEN, …) without owning the operator's primary
        environmentFile. Later files win on duplicate keys (systemd
        semantics).
      '';
      example = [ "/var/lib/telephony/webphone-env" ];
    };

    # Typed front for settings.csrf.trusted_*: same values the freeform
    # settings accept, but discoverable and checkable as module options.
    # Precedence per key: typed csrf.* (mkForce) beats a raw
    # settings.csrf.* value, which beats the caddy.enable defaults
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
          `settings.csrf.trusted_proxies` value and the `caddy.enable`
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
          `settings.csrf.trusted_origins` value and the `caddy.enable`
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

    caddy = {
      enable = lib.mkEnableOption ''
        a Caddy vhost that terminates TLS (automatic HTTPS by default)
        and proxies the app's HTTP surface. Enabling it also defaults
        settings.csrf to trust the loopback proxy and the
        https://<hostName> origin; without that fronting shape (or a
        hand-rolled equivalent in settings.csrf) the CSRF middleware
        rejects every browser POST behind TLS.'';
      hostName = lib.mkOption {
        type = lib.types.str;
        example = "phone.example.org";
        description = "Virtual host name for the generated Caddy vhost.";
      };
      sipUpstream = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        example = "https://pbx.example.org:7443";
        description = ''
          Upstream the SIP WebSocket path is bridged to (the PBX's
          sofia wss binding; an https:// scheme makes Caddy speak TLS
          to the upstream). The app itself never terminates the SIP
          WebSocket — the island's wss connection must reach the PBX,
          so this bridge is what makes calls work behind the vhost.
          null (default) renders no /sip handle: the deployment then
          routes the path itself (the telephony stack does exactly
          that with its own front).'';
      };
      hsts = {
        enable = lib.mkEnableOption ''
          Strict-Transport-Security on the generated vhost. Default off:
          HSTS pins browsers to https for maxAge seconds, so flipping it
          on before the deployment is genuinely https-only (ACME working,
          no http-only tooling left) can brick the domain for that window.
          Caddy serves the header on both schemes unless fenced — verify
          the https-only story before enabling.'';
        maxAge = lib.mkOption {
          type = lib.types.ints.positive;
          default = 63072000;
          example = 31536000;
          description = "HSTS max-age in seconds (default 2 years, the common baseline).";
        };
      };
    };
  };
}
