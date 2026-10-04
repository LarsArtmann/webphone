# The interactive development shell (default): Go 1.27 (the go.mod
# floor), templ, the Go LSP pair, and the linters — plus a minimal `ci`
# shell with only the build/test tools. GOTOOLCHAIN=local in both
# forbids toolchain downloads behind the user's back.
{
  perSystem =
    {
      config,
      pkgs,
      ...
    }:
    let
      # Shared across both shells so CI and the interactive shell cannot
      # drift on the Go pin or the toolchain policy (nix-review batch 2).
      # go.mod carries the 1.27.1 fleet floor (go-health v0.3.0 and the
      # cqrs-htmx v4.11.x train both require it); GOTOOLCHAIN=local
      # forbids toolchain downloads, so the shell must provide 1.27
      # itself. json/v2 shipped stable in Go 1.27: no GOEXPERIMENT since
      # 2026-09-22 (it had become a footgun — callers outside the shell
      # copied it into docs and scripts that then failed in-train).
      goTools = [ pkgs.go_1_27 ];
      goEnv = {
        GOTOOLCHAIN = "local";
      };
    in
    {
      devShells.default = pkgs.mkShellNoCC {
        packages = [
          config.treefmt.build.wrapper
          # The Go LSPs run inside the shell via the project crushrc:
          # the host toolchain is older than go.mod's 1.27.1 floor,
          # so bare gopls/golangci-lint fail every go list.
          pkgs.gopls
          pkgs.golangci-lint-langserver
          pkgs.templ
          pkgs.golangci-lint
          pkgs.esbuild
          pkgs.jq
          pkgs.nil
          pkgs.oxlint
          pkgs.vulnix
          # go-licenses rebuilt on Go 1.27: the stock 2.0.1 rides go
          # 1.26.8, and its nixpkgs wrapper BAKES that GOROOT into the
          # binary — the package loader then dies (F1004, "crypto/mldsa
          # is not in std") on any dependency importing Go 1.27 stdlib,
          # which go-webauthn/webauthn (passkey train) does. Both args
          # matter: go pins the wrapper's GOROOT, buildGoModule the
          # compiler.
          (pkgs.go-licenses.override {
            go = pkgs.go_1_27;
            buildGoModule = pkgs.buildGo127Module;
          })
          # codespell in the shell so BuildFlow's on-demand step runs the
          # REAL binary (it honors .codespellrc; BuildFlow's built-in
          # fallback scanner does not — 3894 vendor/noise findings,
          # T14 2026-09-30). statix + deadnix: the nix-review follow-ups
          # (TODO row) want them one command away.
          pkgs.codespell
          pkgs.statix
          pkgs.deadnix
        ]
        ++ goTools;
        env = goEnv;
      };

      # A minimal CI shell: only the build/test tools (Go, templ) with
      # no interactive LSP/editor tooling, so the CI job enters a small
      # closure quickly.
      devShells.ci = pkgs.mkShellNoCC {
        packages = goTools ++ [
          pkgs.templ
          pkgs.golangci-lint
        ];
        env = goEnv;
      };
    };
}
