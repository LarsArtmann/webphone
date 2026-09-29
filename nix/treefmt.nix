# Formatting gate: nixfmt for Nix, gofmt for Go, prettier for the
# hand-rolled frontend assets only (shell.js, theme-preload.js, the
# island sources — BuildFlow's oxfmt owns island-tests and the rest).
{
  perSystem =
    { config, ... }:
    {
      treefmt = {
        projectRootFile = "flake.nix";
        programs = {
          nixfmt.enable = true;
          gofmt.enable = true;
          prettier = {
            enable = true;
            # No *.html here: the only HTML in the repo is historical
            # docs/status + docs/planning snapshots (point-in-time,
            # never reformatted — and prettier cannot parse one).
            includes = [
              "*.css"
              "internal/web/assets/shell.js"
              "internal/web/assets/theme-preload.js"
              "internal/web/assets/island/**/*.js"
            ];
          };
        };
      };
    };
}
