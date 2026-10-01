# Shared evaluation for the webphone-module check: the NixOS
# caddy/systemd/users stand-in module, the base eval, and the derived
# config handles the case modules read. Split out of module-check.nix
# so that file plus the csrf/backup case files each stay focused
# (nix-review monolith guideline: files under ~300 lines).
{
  lib,
  pkgs,
  self',
}:
let
  # The stand-ins shared by the base evaluation and every variant
  # below. freeformType accepts any further top-level config key the
  # module writes, so a new option path never silently breaks the
  # check; the declared options below keep explicit types + defaults
  # for the paths the assertions read.
  moduleSet = extra: {
    modules = [
      { _module.args.pkgs = pkgs; }
      {
        freeformType = lib.types.attrsOf lib.types.anything;
        options = {
          services.caddy = {
            enable = lib.mkOption {
              type = lib.types.bool;
              default = false;
            };
            virtualHosts = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
          };
          systemd.services = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
          # The backup timer writes systemd.timers; the stand-in
          # must accept it like services.
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
        # recursiveUpdate, not `//`: extras like the HSTS variant nest
        # deeper (services.webphone.caddy.hsts) and a shallow merge
        # would drop the base attrs.
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
in
{
  inherit moduleSet evaluated;
  cfg = evaluated.config.services.webphone;
  vhost = evaluated.config.services.caddy.virtualHosts."phone.example.org";
}
