# The runtime-closure vulnerability scan: `nix run .#vulnix` (repo
# root, network required for the advisory DB). Networked by nature,
# so it is an app, not a flake check.
{
  perSystem =
    {
      lib,
      pkgs,
      self',
      ...
    }:
    {
      # Scans ONLY the runtime closure — the build closure (bootstrap
      # toolchains, gcc, zlib) never deploys and drowns the signal.
      #
      # Triage built in (the documented manual procedure, automated):
      # NVD range-matches distro-patched versions (every glibc CVE
      # against 2.42-84), so a finding only counts when the CVE id is
      # NOT present in the LOCKED nixpkgs rev's patch set for the
      # flagged package. All-glibc + all-patched → exit 0 with a
      # TRIAGED banner; anything else fails the train.
      apps.vulnix =
        let
          script = pkgs.writeShellApplication {
            name = "webphone-vulnix";
            runtimeInputs = [
              pkgs.nix
              pkgs.vulnix
              pkgs.jq
              pkgs.coreutils
              self'.packages.webphone-vulnix-triage
            ];
            text = ''
              # Cwd guard: the triage below reads flake.lock from the
              # current directory; run from anywhere else it dies with a
              # jq parse error instead of an actionable message.
              if [ ! -f flake.lock ]; then
                echo "webphone-vulnix: no flake.lock in $(pwd) — run from the webphone repo root" >&2
                exit 1
              fi
              out=$(nix build --no-link --print-out-paths .#webphone)
              echo "webphone-vulnix: scanning the runtime closure of $out ($(nix-store -qR "$out" | wc -l) derivations)"
              # --closure: runtime dependencies ONLY. Without it vulnix
              # closes over BUILD inputs (bootstrap toolchains, gcc,
              # binutils) and drowns the signal.
              scan=$(vulnix --closure "$out" 2>&1) && {
                echo "$scan"
                echo "webphone-vulnix: no known advisories in the runtime closure"
                exit 0
              }
              echo "$scan"
              rev=$(jq -r '.nodes.nixpkgs.locked.rev' flake.lock)
              patches=$(nix eval "github:NixOS/nixpkgs/$rev#glibc.patches" --json | jq -r '.[]')
              scan_file=$(mktemp)
              patches_file=$(mktemp)
              trap 'rm -f "$scan_file" "$patches_file"' EXIT
              printf '%s\n' "$scan" > "$scan_file"
              printf '%s\n' "$patches" > "$patches_file"
              # Verdict logic lives in the shared CLI so the fixture
              # check (checks.vulnix-triage) exercises the same code.
              webphone-vulnix-triage "$scan_file" "$patches_file" "$rev"
            '';
          };
        in
        {
          type = "app";
          program = lib.getExe script;
        };
    };
}
