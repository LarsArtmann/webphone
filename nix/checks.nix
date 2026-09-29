# The fast checks: the package itself, the formatter gate, the island
# JS tests, the vulnix-triage fixture smoke, the Nix linters, and the
# island no-undef gate. The module-eval check and the VM-backed tests
# live in ./module-check.nix and ./vm-tests.nix.
# `self` is a TOP-level flake-parts module arg (not perSystem) — it is
# declared here and reaches the perSystem body via lexical closure.
{
  self,
  ...
}:
{
  perSystem =
    {
      config,
      lib,
      pkgs,
      self',
      ...
    }:
    {
      checks = {
        webphone = self'.packages.webphone;
        format = config.treefmt.build.check self;

        # Island JS tests under node:test — the runner the island never
        # had (toast rendering and i18n parity were Go asset-grep
        # tripwires only). Tests live in island-tests/, a SIBLING of
        # island/, so the `all:island` embed never ships them to
        # browsers; --test-force-exit keeps the toast lifetime timers
        # from holding the event loop open.
        island-js =
          pkgs.runCommand "island-js-check"
            {
              nativeBuildInputs = [ pkgs.nodejs ];
            }
            ''
              cd ${self}
              node --test --test-force-exit \
                internal/web/assets/island-tests/*.test.mjs | tee $out
            '';

        # Fixture smoke of the vulnix triage CLI — the 2026-09-20
        # regression class (grep-in-pipeline under set -e inverted every
        # verdict) would only have surfaced at the next train. Runs the
        # SAME derivation the `.#vulnix` app calls, against fixtures.
        vulnix-triage =
          let
            script = pkgs.writeShellApplication {
              name = "webphone-vulnix-triage-check";
              runtimeInputs = [
                pkgs.coreutils
                pkgs.gnugrep
                self'.packages.webphone-vulnix-triage
              ];
              text = ''
                tmp=$(mktemp -d)
                trap 'rm -rf "$tmp"' EXIT
                printf '%s\n' "upstream fix mentioning CVE-2026-5450" > "$tmp/patched.patch"
                printf '%s\n' "$tmp/patched.patch" > "$tmp/patches.lst"

                # 1. glibc finding, patch present → triaged, exit 0.
                printf '%s\n' "/nix/store/x-glibc-2.42.drv" "CVE-2026-5450" > "$tmp/scan1"
                if out=$(webphone-vulnix-triage "$tmp/scan1" "$tmp/patches.lst" testrev); then
                  grep -q "distro-patched" <<< "$out" || {
                    echo "case1 verdict wrong: $out" >&2
                    exit 1
                  }
                else
                  echo "case1: patched finding must exit 0" >&2
                  exit 1
                fi

                # 2. glibc finding, NO patch → REAL finding, exit 1.
                printf '%s\n' "/nix/store/x-glibc-2.42.drv" "CVE-2099-0001" > "$tmp/scan2"
                if out=$(webphone-vulnix-triage "$tmp/scan2" "$tmp/patches.lst" testrev 2>&1); then
                  echo "case2: real finding must exit 1: $out" >&2
                  exit 1
                elif ! grep -q "REAL finding" <<< "$out"; then
                  echo "case2 verdict wrong: $out" >&2
                  exit 1
                fi

                # 3. unparseable output → parse-failure verdict, exit 1.
                printf '%s\n' "vulnix exploded" > "$tmp/scan3"
                if out=$(webphone-vulnix-triage "$tmp/scan3" "$tmp/patches.lst" testrev 2>&1); then
                  echo "case3: garbage must exit 1: $out" >&2
                  exit 1
                elif ! grep -q "parseable findings" <<< "$out"; then
                  echo "case3 verdict wrong: $out" >&2
                  exit 1
                fi

                # 4. non-glibc derivation flagged → no automated triage, exit 1.
                printf '%s\n' "/nix/store/x-openssl-3.5.1.drv" "CVE-2026-1111" > "$tmp/scan4"
                if out=$(webphone-vulnix-triage "$tmp/scan4" "$tmp/patches.lst" testrev 2>&1); then
                  echo "case4: non-glibc must exit 1: $out" >&2
                  exit 1
                elif ! grep -q "non-glibc derivations flagged" <<< "$out"; then
                  echo "case4 verdict wrong: $out" >&2
                  exit 1
                fi

                # 5. mixed patched + real: the loop must reach BOTH
                # verdicts and exit 1 (the original bug aborted mid-loop).
                printf '%s\n' "/nix/store/x-glibc-2.42.drv" "CVE-2026-5450" "CVE-2099-0002" > "$tmp/scan5"
                if out=$(webphone-vulnix-triage "$tmp/scan5" "$tmp/patches.lst" testrev 2>&1); then
                  echo "case5: mixed findings must exit 1: $out" >&2
                  exit 1
                elif ! grep -q "distro-patched" <<< "$out" || ! grep -q "REAL finding" <<< "$out"; then
                  echo "case5: loop did not reach both verdicts: $out" >&2
                  exit 1
                fi

                echo "vulnix-triage: all 5 fixture cases green"
              '';
            };
          in
          pkgs.runCommand "webphone-vulnix-triage-check" { } "${lib.getExe script} | tee $out";

        statix =
          pkgs.runCommand "statix-check"
            {
              nativeBuildInputs = [ pkgs.statix ];
              meta.description = "statix lint over the Nix files";
            }
            ''
              cd ${self}
              statix check -o errfmt . 2>&1 | tee $out
            '';
        deadnix =
          pkgs.runCommand "deadnix-check"
            {
              nativeBuildInputs = [ pkgs.deadnix ];
              meta.description = "deadnix scan over the Nix files";
            }
            ''
              cd ${self}
              deadnix --fail --no-lambda-pattern-names . 2>&1 | tee $out
            '';

        # no-undef over the SIP island modules: a call to an undefined
        # identifier used to surface only as a silent browser
        # ReferenceError (the accept/reject bug class) — this gate is
        # the local tripwire. The island is excluded from BuildFlow,
        # so the gate lives here with its config beside the sources.
        island-lint =
          pkgs.runCommand "island-lint-check"
            {
              nativeBuildInputs = [ pkgs.oxlint ];
              meta.description = "oxlint no-undef over the SIP island modules";
            }
            ''
              cd ${self}
              if oxlint -c internal/web/assets/island/oxlint.json internal/web/assets/island/app/ internal/web/assets/shell.js internal/web/assets/theme-preload.js; then
                echo "no-undef clean over:" >$out
                ls internal/web/assets/island/app/ >>$out
                echo internal/web/assets/shell.js >>$out
                echo internal/web/assets/theme-preload.js >>$out
              else
                echo "island no-undef gate FAILED" >&2
                exit 1
              fi
            '';
      };
    };
}
