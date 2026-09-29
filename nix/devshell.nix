# The interactive development shell: Go 1.27 (the go.mod floor), templ,
# the Go LSP pair, and the linters. GOTOOLCHAIN=local forbids toolchain
# downloads behind the user's back.
{
  perSystem =
    {
      config,
      pkgs,
      ...
    }:
    {
      devShells.default = pkgs.mkShellNoCC {
        packages = [
          config.treefmt.build.wrapper
          # go.mod carries the 1.27.1 fleet floor (go-health v0.3.0 and
          # the cqrs-htmx v4.11.x train both require it);
          # GOTOOLCHAIN=local forbids toolchain downloads, so the shell
          # must provide 1.27 itself.
          pkgs.go_1_27
          pkgs.templ
          pkgs.golangci-lint
          # The Go LSPs run inside the shell via the project crushrc:
          # the host toolchain is older than go.mod's 1.27.1 floor,
          # so bare gopls/golangci-lint fail every go list.
          pkgs.gopls
          pkgs.golangci-lint-langserver
          pkgs.esbuild
          pkgs.jq
          pkgs.nil
          pkgs.oxlint
          pkgs.vulnix
          pkgs.go-licenses
        ];
        env = {
          # json/v2 shipped stable in Go 1.27: no GOEXPERIMENT since
          # 2026-09-22 (it had become a footgun — callers outside the
          # shell copied it into docs and scripts that then failed
          # in-train). local keeps the shell's go_1_27 instead of
          # downloading a toolchain behind the user's back.
          GOTOOLCHAIN = "local";
        };
      };
    };
}
