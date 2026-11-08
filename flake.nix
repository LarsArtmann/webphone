{
  description = "Webphone: self-hosted unified-communications web app (calls, SMS/MMS, fax, voicemail) in one Go binary";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # The docs-health battery (freshness / audit-all / annotate / routing /
    # cadence) as ONE stdlib-only Go binary, built here so CI and the devShell
    # cannot disagree about which rules ran. `flake = false` is deliberate:
    # the upstream flake pulls a PRIVATE erraudit input over ssh, which a
    # public-repo CI job cannot fetch, and this repo only needs the source
    # tree. The rev is locked, so rule behaviour cannot drift under us; a
    # deliberate bump is the only way the gates change.
    docs-gates = {
      url = "github:Artmann-Technologies/docs-gates/2278406afdfd1b8df7674e2e594b6ea880694c90";
      flake = false;
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    let
      # The release version — the package version AND the /version
      # ldflags injection in one place. release.sh seds THIS binding
      # (grep "webphoneVersion = " flake.nix), so keep it in this file,
      # in this exact shape, in lockstep with the git tag at release.
      webphoneVersion = "2.10.0";
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      imports = [
        inputs.treefmt-nix.flakeModule
        # The release version rides into the packages module as a
        # module arg (visible at the module's top-level function).
        { _module.args.webphoneVersion = webphoneVersion; }
        ./nix/packages.nix
        ./nix/apps.nix
        ./nix/checks.nix
        ./nix/module-check.nix
        ./nix/vm-tests.nix
        ./nix/devshell.nix
        ./nix/treefmt.nix
      ];

      # The NixOS module ships from this repo so the binary and its
      # deployment shape stay in sync; the consuming telephony stack may
      # import it or keep its own reverse proxy (see package/nixos-module.nix).
      flake.nixosModules = {
        default = import ./package/nixos-module.nix;
        # Alias for consumers that prefer an explicit name.
        webphone = import ./package/nixos-module.nix;
      };
    };
}
