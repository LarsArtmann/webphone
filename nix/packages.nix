# Package definitions: the webphone binary and the vulnix triage CLI.
# webphoneVersion arrives as a module arg from flake.nix — release.sh
# seds the binding there (grep "webphoneVersion = "), so it must not
# move into this file. self is the flake source: its rev/dirtyRev and
# lastModifiedDate feed /version's commit enrichment (commit-time facts
# only — no build clock, byte-reproducibility preserved).
{
  self,
  webphoneVersion,
  ...
}:
{
  perSystem =
    {
      lib,
      pkgs,
      self',
      ...
    }:
    {
      packages = {
        default = self'.packages.webphone;

        # The vulnix triage verdict logic as a standalone CLI so the
        # `.#vulnix` app and the fixture check exercise the SAME code:
        # webphone-vulnix-triage <scan-file> <patches-file> <rev>.
        # Exit 0 = every flagged CVE is distro-patched in the locked
        # nixpkgs; exit 1 = unparseable output, a non-glibc derivation
        # flagged, or a real (unpatched) finding.
        webphone-vulnix-triage = pkgs.writeShellApplication {
          name = "webphone-vulnix-triage";
          runtimeInputs = [ pkgs.gnugrep ];
          text = ''
            scan_file="$1"
            patches_file="$2"
            rev="$3"

            flagged_drvs=$(grep -oE '/nix/store/[^ ]+\.drv' "$scan_file" | sort -u || true)
            non_glibc=$(printf '%s\n' "$flagged_drvs" | grep -v -- '-glibc-' || true)
            cves=$(grep -oE 'CVE-[0-9]{4}-[0-9]+' "$scan_file" | sort -u || true)

            if [ -z "$flagged_drvs" ] || [ -z "$cves" ]; then
              echo "webphone-vulnix: vulnix failed without parseable findings — inspect the output above" >&2
              exit 1
            fi

            if [ -n "$non_glibc" ]; then
              echo "webphone-vulnix: non-glibc derivations flagged (no automated triage):" >&2
              echo "$non_glibc" >&2
              exit 1
            fi

            # Hit-check reads patch files directly (no grep-in-pipeline
            # subshell: the 2026-09-20 regression class — set -e plus
            # a non-matching grep in a pipeline aborted the loop and
            # inverted every verdict).
            untriaged=0
            while read -r cve; do
              [ -n "$cve" ] || continue
              hit=""
              while read -r p; do
                [ -n "$p" ] || continue
                if grep -q "$cve" "$p" 2>/dev/null; then
                  hit=1
                  break
                fi
              done < "$patches_file"
              if [ -n "$hit" ]; then
                echo "webphone-vulnix: $cve — distro-patched in locked nixpkgs $rev (range-match noise)"
              else
                echo "webphone-vulnix: $cve — NOT found in the locked glibc patches: REAL finding, act on it" >&2
                untriaged=1
              fi
            done <<EOF
            $cves
            EOF

            if [ "$untriaged" -eq 0 ]; then
              echo "webphone-vulnix: all findings triaged as distro-patched — zero real advisories"
              exit 0
            fi
            exit 1
          '';
        };

        # One Go binary: templ shell + embedded island assets + SQLite.
        # (json/v2 went stable in Go 1.27 — no GOEXPERIMENT anywhere
        # since 2026-09-22.) webphoneVersion (from flake.nix) is the
        # single source: the package version AND the /version ldflags
        # injection — keep it in lockstep with the git tag at release.
        webphone =
          pkgs.buildGoModule.override
            {
              # Go >= 1.27.1: the fleet floor (see go.mod + devShell).
              go = pkgs.go_1_27;
            }
            {
              pname = "webphone";
              version = webphoneVersion;

              src = lib.fileset.toSource {
                root = ./..;
                # Exactly the build inputs: sources under cmd/ and
                # internal/ (island assets + committed *_templ.go live
                # there) plus the module definition files. Everything
                # else (docs, scripts, package/) never reaches Go.
                fileset = lib.fileset.unions [
                  ../cmd
                  ../go.mod
                  ../go.sum
                  ../internal
                ];
              };

              vendorHash = "sha256-hyfVz9Ud0LZfPjp+fRuTG698fXRtCfW98nTyhjiD+yA=";

              proxyVendor = true;

              subPackages = [ "cmd/webphone" ];

              ldflags = [
                "-s"
                "-w"
                # /version reports the released version (v-prefixed, like
                # the git tag) instead of Go's "(devel)".
                "-X github.com/larsartmann/webphone/internal/server.buildVersion=v${webphoneVersion}"
                # Commit enrichment (T12): self.rev when clean, dirtyRev
                # when the tree is dirty, lastModifiedDate = the commit
                # timestamp — all commit-time facts, so the aarch64
                # byte-reproducibility assert stays honest.
                "-X github.com/larsartmann/webphone/internal/server.buildCommit=${
                  if self ? rev && self.rev != null then self.rev else (self.dirtyRev or "unknown")
                }"
                "-X github.com/larsartmann/webphone/internal/server.buildCommitDate=${
                  let
                    d = self.lastModifiedDate;
                  in
                  "${builtins.substring 0 4 d}-${builtins.substring 4 2 d}-${builtins.substring 6 2 d}T${builtins.substring 8 2 d}:${builtins.substring 10 2 d}:${builtins.substring 12 2 d}Z"
                }"
              ];

              doCheck = true;

              meta = {
                description = "Self-hosted unified-communications web app: calls, SMS/MMS threads, fax, voicemail";
                homepage = "https://github.com/LarsArtmann/webphone";
                license = lib.licenses.mit;
                mainProgram = "webphone";
                platforms = lib.platforms.linux;
                maintainers = [
                  {
                    name = "Lars Artmann";
                    github = "LarsArtmann";
                  }
                ];
              };
            };
      };
    };
}
