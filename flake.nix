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
  };

  outputs =
    inputs@{ self, flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      imports = [ inputs.treefmt-nix.flakeModule ];

      # The NixOS module ships from this repo so the binary and its
      # deployment shape stay in sync; the consuming telephony stack may
      # import it or keep its own reverse proxy (see package/nixos-module.nix).
      flake.nixosModules = {
        default = import ./package/nixos-module.nix;
        # Alias for consumers that prefer an explicit name.
        webphone = import ./package/nixos-module.nix;
      };

      perSystem =
        {
          config,
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
            webphone-vulnix-triage =
              pkgs.writeShellApplication
                {
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
            # GOEXPERIMENT=jsonv2 is required by templ-components (encoding/
            # json/v2) until Go 1.27 ships it stable. webphoneVersion is the
            # single source: the package version AND the /version ldflags
            # injection — keep it in lockstep with the git tag at release.
            webphone =
              let
                webphoneVersion = "2.4.0";
              in
              pkgs.buildGoModule.override
                {
                  # Go >= 1.27.1: the fleet floor (see go.mod + devShell).
                  go = pkgs.go_1_27;
                }
                {
                  pname = "webphone";
                  version = webphoneVersion;

                  src = lib.fileset.toSource {
                    root = ./.;
                    # Exactly the build inputs: sources under cmd/ and
                    # internal/ (island assets + committed *_templ.go live
                    # there) plus the module definition files. Everything
                    # else (docs, scripts, package/) never reaches Go.
                    fileset = lib.fileset.unions [
                      ./cmd
                      ./go.mod
                      ./go.sum
                      ./internal
                    ];
                  };

                  vendorHash = "sha256-n8scPBKrJXjxinLixyJ7Amen600nRVSSRVr6KBdQy9A=";

                  proxyVendor = true;

                  env.GOEXPERIMENT = "jsonv2";

                  subPackages = [ "cmd/webphone" ];

                  ldflags = [
                    "-s"
                    "-w"
                    # /version reports the released version (v-prefixed, like
                    # the git tag) instead of Go's "(devel)".
                    "-X github.com/larsartmann/webphone/internal/server.buildVersion=v${webphoneVersion}"
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
                  cd ${./.}
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
              pkgs.runCommand "webphone-vulnix-triage-check" { }
                "${lib.getExe script} | tee $out";

            # Evaluate the NixOS module with a minimal config and build
            # the artifacts it would generate — catches option/syntax
            # breakage without a full NixOS evaluation.
            webphone-module =
              let
                # The NixOS nginx/systemd/users stand-ins shared by the
                # base evaluation and the HSTS opt-in variant below.
                moduleSet = extra: {
                  modules = [
                    { _module.args.pkgs = pkgs; }
                    # Minimal stand-ins for the NixOS nginx/systemd/users
                    # modules: the webphone module writes services.nginx,
                    # systemd.services, and users.{users,groups} config.
                    {
                      options = {
                        services.nginx = {
                          enable = lib.mkOption {
                            type = lib.types.bool;
                            default = false;
                          };
                          recommendedProxySettings = lib.mkOption {
                            type = lib.types.bool;
                            default = false;
                          };
                          virtualHosts = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                        };
                        systemd.services = lib.mkOption { type = lib.types.attrsOf lib.types.anything; };
                        # The backup timer writes systemd.timers; the
                        # stand-in must accept it like services.
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
                    (import ./package/nixos-module.nix)
                    {
                      # recursiveUpdate, not `//`: extras like the HSTS
                      # variant nest deeper (services.webphone.nginx.hsts)
                      # and a shallow merge would drop the base attrs.
                      services.webphone = lib.recursiveUpdate {
                        enable = true;
                        package = self'.packages.webphone;
                        nginx.enable = true;
                        nginx.hostName = "phone.example.org";
                        settings.sip_domain = "pbx.example.org";
                      } extra;
                    }
                  ];
                };
                evaluated = lib.evalModules (moduleSet { });
                cfg = evaluated.config.services.webphone;
                vhost = evaluated.config.services.nginx.virtualHosts."phone.example.org";
                locationNames = lib.attrNames vhost.locations;
                # Every location the DOM/SSE contract rides on must be
                # proxied by the module's own vhost — including the probe
                # triple, which ships as dedicated locations so fleet
                # scrapers can be fenced per location without touching the
                # app's "/".
                missingLocations = lib.filter (loc: !lib.elem loc locationNames) [
                  "/"
                  cfg.settings.websocket_path
                  "/events"
                  "/healthz"
                  "/livez"
                  "/startupz"
                ];
                unitPresent = evaluated.config.systemd.services ? "webphone";
              in
              pkgs.linkFarm "webphone-module-check" [
                {
                  name = "webphone-config.json";
                  path = (pkgs.formats.json { }).generate "webphone-config.json" cfg.settings;
                }
                {
                  name = "listen-port";
                  path = pkgs.writeText "listen-port" (lib.last (lib.splitString ":" cfg.settings.addr));
                }
                {
                  name = "vhost-locations";
                  path = pkgs.writeText "vhost-locations" (
                    if missingLocations == [ ] then
                      lib.concatStringsSep "\n" locationNames
                    else
                      throw "webphone-module check: vhost locations missing: ${toString missingLocations}"
                  );
                }
                {
                  name = "systemd-unit";
                  path = pkgs.writeText "systemd-unit" (
                    if unitPresent then
                      "systemd.services.webphone present"
                    else
                      throw "webphone-module check: systemd.services.webphone missing"
                  );
                }
                {
                  name = "csrf-fronted-origin";
                  path = pkgs.writeText "csrf-fronted-origin" (
                    if
                      cfg.settings.csrf.trusted_origins == [ "https://phone.example.org" ]
                      && cfg.settings.csrf.trusted_proxies == [ "127.0.0.1" ]
                    then
                      "csrf fronting defaults present"
                    else
                      throw "webphone-module check: csrf fronting defaults missing from the rendered settings"
                  );
                }
                {
                  # The typed csrf.* options must render into settings.csrf
                  # and BEAT the nginx-derived defaults when set.
                  name = "csrf-typed-override";
                  path = pkgs.writeText "csrf-typed-override" (
                    let
                      typedEvaluated = lib.evalModules (moduleSet {
                        csrf.trustedProxies = [ "10.9.8.7" ];
                        csrf.trustedOrigins = [ "https://alt.example.org" ];
                      });
                      typedCfg = typedEvaluated.config.services.webphone;
                    in
                    if
                      typedCfg.settings.csrf.trusted_proxies == [ "10.9.8.7" ]
                      && typedCfg.settings.csrf.trusted_origins == [ "https://alt.example.org" ]
                    then
                      "typed csrf options render and override"
                    else
                      throw "webphone-module check: typed csrf options did not render into settings.csrf over the nginx defaults"
                  );
                }
                {
                  # Precedence under conflict, pinned per key: typed csrf.*
                  # (mkForce) > raw settings.csrf.* (plain) > the nginx
                  # defaults (mkDefault). One case exercises all three
                  # lanes: typed proxies beat the raw proxy, and the raw
                  # origin (typed origins empty) beats the nginx default.
                  name = "csrf-conflict-precedence";
                  path = pkgs.writeText "csrf-conflict-precedence" (
                    let
                      conflictEvaluated = lib.evalModules (moduleSet {
                        csrf.trustedProxies = [ "10.9.8.7" ];
                        settings.csrf = {
                          trusted_proxies = [ "192.0.2.1" ];
                          trusted_origins = [ "https://raw.example.org" ];
                        };
                      });
                      conflictCfg = conflictEvaluated.config.services.webphone;
                    in
                    if
                      conflictCfg.settings.csrf.trusted_proxies == [ "10.9.8.7" ]
                      && conflictCfg.settings.csrf.trusted_origins == [ "https://raw.example.org" ]
                    then
                      "csrf conflict precedence: typed > raw > nginx default"
                    else
                      throw "webphone-module check: csrf conflict precedence broken (expected typed proxies and raw origins to win their lanes)"
                  );
                }
                {
                  # backup.enable must render the timer (OnCalendar + Unit)
                  # and the oneshot service.
                  name = "backup-timer";
                  path = pkgs.writeText "backup-timer" (
                    let
                      backupEvaluated = lib.evalModules (moduleSet {
                        backup.enable = true;
                      });
                      timer = backupEvaluated.config.systemd.timers.webphone-backup;
                      service = backupEvaluated.config.systemd.services.webphone-backup;
                    in
                    if
                      timer.timerConfig.OnCalendar == "*-*-* 04:30:00"
                      && timer.timerConfig.Unit == "webphone-backup.service"
                      && timer.wantedBy == [ "timers.target" ]
                      && service.serviceConfig.Type == "oneshot"
                    then
                      "backup timer + oneshot rendered"
                    else
                      throw "webphone-module check: backup.enable did not render the timer/oneshot pair"
                  );
                }
                {
                  # serverTiming.enable must set the env gate the middleware
                  # reads; without it the environment key stays absent.
                  name = "server-timing";
                  path = pkgs.writeText "server-timing" (
                    let
                      timingEvaluated = lib.evalModules (moduleSet {
                        serverTiming.enable = true;
                      });
                      timingUnit = timingEvaluated.config.systemd.services.webphone.environment;
                    in
                    if timingUnit ? WEBPHONE_DEBUG_TIMING && timingUnit.WEBPHONE_DEBUG_TIMING == "1" then
                      "server-timing env gate rendered"
                    else
                      throw "webphone-module check: serverTiming.enable did not set WEBPHONE_DEBUG_TIMING"
                  );
                }
                {
                  name = "hsts-opt-in";
                  path = pkgs.writeText "hsts-opt-in" (
                    let
                      hstsEvaluated = lib.evalModules (moduleSet {
                        nginx.hsts.enable = true;
                      });
                      hstsVhost = hstsEvaluated.config.services.nginx.virtualHosts."phone.example.org";
                    in
                    if lib.hasInfix "Strict-Transport-Security" hstsVhost.extraConfig then
                      "hsts header present when enabled"
                    else
                      throw "webphone-module check: nginx.hsts.enable did not produce an HSTS vhost header"
                  );
                }
              ];

            # Backup-timer VM test (fax-feed-test style, named after the
            # consuming stack's tests/fax-feed.nix): boot the REAL service
            # with backup.enable, run the oneshot, and assert the online
            # snapshot lands under destDir while the service keeps serving —
            # the production behaviors are the sqlite .backup consistency
            # and the no-restart claim, not just file existence.
            # kvm-gated like the stack's VM tests: `nix flake check` skips
            # it (with a warning) on machines without KVM instead of
            # degrading to multi-minute TCG boots.
            webphone-backup = pkgs.testers.runNixOSTest {
              name = "webphone-backup";

              requiredFeatures.kvm = true;

              nodes.machine =
                { pkgs, ... }:
                {
                  imports = [ ./package/nixos-module.nix ];
                  services.webphone = {
                    enable = true;
                    package = self'.packages.webphone;
                    backup.enable = true;
                  };
                  environment.systemPackages = [ pkgs.sqlite ];
                  system.stateVersion = "26.05";
                };

              testScript = ''
                machine.wait_for_unit("webphone.service")
                machine.wait_for_open_port(8080)

                # Deterministic run: start the oneshot directly (the timer
                # exists too, but the test asserts outcomes, not scheduler
                # timing — same posture as the stack's fax-feed test).
                machine.succeed("systemctl start webphone-backup.service")

                # The snapshot pair lands under destDir.
                machine.wait_until_succeeds(
                    "test -f /var/lib/webphone-backup/webphone.db",
                    timeout=30,
                )
                machine.succeed("test -d /var/lib/webphone-backup/files")

                # It is a consistent database, not a torn copy: sqlite's
                # .backup API ran to completion.
                machine.succeed(
                    "sqlite3 /var/lib/webphone-backup/webphone.db 'pragma integrity_check' | grep -q '^ok$'"
                )

                # The timer is wired to the oneshot on the default calendar
                # (systemctl show has no OnCalendar unit property — the
                # rendered unit file is the truth here).
                machine.succeed(
                    "systemctl cat webphone-backup.timer | grep -q '^OnCalendar=\\*-\\*-\\* 04:30:00'"
                )
                machine.succeed(
                    "systemctl cat webphone-backup.timer | grep -q '^Unit=webphone-backup.service$'"
                )

                # Online claim: the phone service never restarted for the
                # backup (uptime predates the oneshot run).
                machine.succeed(
                    "systemctl show webphone.service -p NRestarts | grep -q 'NRestarts=0'"
                )
              '';
            };

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
                  if oxlint -c internal/web/assets/island/oxlint.json internal/web/assets/island/app/ internal/web/assets/shell.js; then
                    echo "no-undef clean over:" >$out
                    ls internal/web/assets/island/app/ >>$out
                  else
                    echo "island no-undef gate FAILED" >&2
                    exit 1
                  fi
                '';
          };

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
              # json/v2 shipped stable in Go 1.27; the env stays harmless
              # (and matches the fleet's other flakes). local keeps the
              # shell's go_1_27 instead of downloading a toolchain behind
              # the user's back.
              GOEXPERIMENT = "jsonv2";
              GOTOOLCHAIN = "local";
            };
          };

          # Runtime-closure vulnerability scan: `nix run .#vulnix` (repo
          # root, network required for the advisory DB). Scans ONLY the
          # runtime closure — the build closure (bootstrap toolchains,
          # gcc, zlib) never deploys and drowns the signal. Networked by
          # nature, so it is an app, not a flake check.
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
                  "internal/web/assets/island/**/*.js"
                ];
              };
            };
          };
        };
    };
}
