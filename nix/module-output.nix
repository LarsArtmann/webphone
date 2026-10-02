# The full-text rendering of the module's generated output: the Caddy
# vhost body, the retention=7 backup script, and the settings JSON, in
# one deterministic document. This is the SINGLE source both the
# `module-output-golden` check and the golden fixture come from (the
# check diffs the rendered text against nix/module-output.golden).
# Regenerate the fixture after an intentional change with:
#   nix eval --impure --raw --expr 'let
#     flake = builtins.getFlake "path:'"$PWD"'";
#     system = builtins.currentSystem;
#     pkgs = flake.inputs.nixpkgs.legacyPackages.''${system};
#     base = import (flake.outPath + "/nix/module-check-base.nix") {
#       inherit pkgs; lib = pkgs.lib;
#       self' = { packages = flake.packages.''${system}; };
#     };
#   in import (flake.outPath + "/nix/module-output.nix") { inherit pkgs; lib = pkgs.lib; inherit base; }' \
#     > nix/module-output.golden
{ lib, base }:
let
  vhost = base.vhost.extraConfig;
  backupScript =
    (lib.evalModules (
      base.moduleSet {
        backup.enable = true;
        backup.retentionDays = 7;
      }
    )).config.systemd.services.webphone-backup.script;
in
''
  == Caddy vhost extraConfig ==
  ${vhost}
  == backup script (retentionDays = 7) ==
  ${backupScript}
  == settings JSON ==
  ${builtins.toJSON base.cfg.settings}
''
