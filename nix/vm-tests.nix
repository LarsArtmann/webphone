# The heavyweight checks: the KVM-gated backup VM test and the
# sandboxed backup/restore drill.
# `self` is a TOP-level flake-parts module arg (not perSystem) — it is
# declared here and reaches the perSystem body via lexical closure.
{
  self,
  ...
}:
{
  perSystem =
    {
      pkgs,
      self',
      ...
    }:
    {
      checks = {
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
              imports = [ ../package/nixos-module.nix ];
              services.webphone = {
                enable = true;
                package = self'.packages.webphone;
                backup.enable = true;
                backup.retentionDays = 7;
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

            # The metaengine read-model database rides the same snapshot
            # (it carries user data since the metaengine train: contacts,
            # snippets, faxes, transcripts) — present, consistent.
            machine.succeed(

            )
            machine.succeed(

            )

            # The system composition root's journal database rides the
            # same snapshot (command/event streams, projections,
            # checkpoints, DLQ — user data from the SUPERB phases on).
            machine.succeed(
                "test -f /var/lib/webphone-backup/system.db"
            )
            machine.succeed(
                "sqlite3 /var/lib/webphone-backup/system.db 'pragma integrity_check' | grep -q '^ok$'"
            )

            # Private-data hardening (2026-09-29 train, pinned 2026-09-30):
            # the oneshot succeeded, the state dir is group-readable at
            # most, and backup artifacts are owner-only (UMask=0077 ->
            # files 600, dirs 700) — comms data never world-readable.
            machine.succeed(
                "systemctl show webphone-backup.service -p Result | grep -q 'Result=success'"
            )
            machine.succeed(
                "test \"$(stat -c %a /var/lib/webphone)\" = 750"
            )
            machine.succeed(
                "stat -c %a /var/lib/webphone-backup/webphone.db | grep -qE '^(600|700)$'"
            )
            machine.succeed(
                "stat -c %a /var/lib/webphone-backup/files | grep -qE '^(700|750)$'"
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

            # Retention (retentionDays = 7): plant a stale dated
            # snapshot AFTER a first successful run (so the StateDirectory
            # and snapshots/ tree exist under the service user), rerun the
            # oneshot — the stale directory is pruned, today's dated
            # snapshot survives with an intact db, and the same-day rerun
            # must not have used itself as the rsync basis.
            machine.succeed(
            "install -d -o webphone -g webphone /var/lib/webphone-backup/snapshots/2000-01-01"
            )
            machine.succeed(
            "touch /var/lib/webphone-backup/snapshots/2000-01-01/stale.db && touch -d '2000-01-01' /var/lib/webphone-backup/snapshots/2000-01-01"
            )
            machine.succeed("systemctl start webphone-backup.service")
            machine.succeed("test ! -e /var/lib/webphone-backup/snapshots/2000-01-01")
            machine.succeed(
            "test -f '/var/lib/webphone-backup/snapshots/'$(date +%F)'/webphone.db'"
            )
            machine.succeed(
            "sqlite3 '/var/lib/webphone-backup/snapshots/'$(date +%F)'/webphone.db' 'pragma integrity_check' | grep -q '^ok$'"
            )
            machine.succeed(

            )
            machine.succeed(
            "test -f '/var/lib/webphone-backup/snapshots/'$(date +%F)'/system.db'"
            )
          '';
        };

        # Error-contract DRILL (SUPERB error-agents M14-M15, 2026-10-26):
        # the failure SHAPE the product owes, asserted in the REAL
        # module/unit/journald context the smoke cannot see. Scenario A:
        # an unwritable data dir under a regular file must produce the
        # designed boot failure — five-part operator block on the
        # journal, exit taxonomy 1, and an honest systemd restart loop.
        # Plus the two runtime shape asserts: /healthz degrades to 503
        # NAMING the failing leg when the blob store becomes unwritable
        # (and recovers when unwedged), and /hooks/* fail CLOSED (503)
        # with no gateway secret configured. KVM-gated like the backup
        # VM test — consumer CI (the fast-checks list) never runs it.
        # The gateway-lane and surface-shape scenarios of the plan stay
        # smoke-owned (scripts/webphone-smoke.py runs them per train on
        # the same binary; one home per fact).
        webphone-error-drill = pkgs.testers.runNixOSTest {
          name = "webphone-error-drill";

          requiredFeatures.kvm = true;

          nodes.machine =
            { pkgs, ... }:
            {
              imports = [ ../package/nixos-module.nix ];
              services.webphone = {
                enable = true;
                package = self'.packages.webphone;
              };
              environment.systemPackages = [ pkgs.curl ];
              system.stateVersion = "26.05";
            };

          testScript = ''
            # The healthy module boots first: the runtime-shape asserts
            # need a live, listening service.
            machine.wait_for_unit("webphone.service")
            machine.wait_for_open_port(8080)

            # Health is honest while healthy.
            machine.succeed("curl -sf http://127.0.0.1:8080/healthz")

            # /hooks/* without a configured secret fail CLOSED.
            machine.succeed(
              "test \"$(curl -s -o /dev/null -w '%{http_code}' -X POST "
              "http://127.0.0.1:8080/hooks/message -d '{}')\" = 503"
            )

            # Wedge the blob store under the running service: the
            # healthz leg must name its failure, not lie with a 200.
            machine.succeed("chmod 000 /var/lib/webphone/files")
            machine.wait_until_succeeds(
              "test \"$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/healthz)\" = 503",
              timeout=30,
            )
            machine.succeed(
              "curl -s http://127.0.0.1:8080/healthz | grep -q 'blob-dir'"
            )
            machine.succeed("chmod 750 /var/lib/webphone/files")
            machine.wait_until_succeeds(
              "curl -sf http://127.0.0.1:8080/healthz",
              timeout=30,
            )

            # Scenario A: the designed boot failure under systemd —
            # DATA_DIR parked under a regular file (ENOTDIR in the
            # storage phase), Restart=always so the loop is real.
            machine.succeed("touch /tmp/blocked")
            machine.succeed(
              "systemd-run --unit=webphone-drill -p Restart=always -p RestartSec=1 "
              "--setenv=WEBPHONE_ADDR=127.0.0.1:18099 "
              "--setenv=WEBPHONE_DATA_DIR=/tmp/blocked/data "
              "${self'.packages.webphone}/bin/webphone"
            )
            machine.wait_until_succeeds(
              "journalctl -u webphone-drill --no-pager | grep -q 'boot failed (class=data-dir'",
              timeout=60,
            )
            # Exit taxonomy FIRST, and RETRIED: Restart=always resets
            # ExecMainStatus to 0 the moment a new cycle starts, so a
            # single-shot read races the 1s restart loop (the 2026-10-29
            # flake: five journalctl greps in between were enough to lose
            # the window). Every failed cycle re-exposes ExecMainStatus=1,
            # so retrying observes one instead of hoping to win a race.
            machine.wait_until_succeeds(
              "systemctl show webphone-drill -p ExecMainStatus "
              "| grep -q 'ExecMainStatus=1'",
              timeout=30,
            )
            for marker in ("WHAT:", "REASSURE:", "WHY:", "FIX:", "ESCAPE:"):
                machine.succeed(
                  f"journalctl -u webphone-drill --no-pager | grep -q '{marker}'"
                )
            machine.wait_until_succeeds(
              "test \"$(systemctl show webphone-drill -p NRestarts --value)\" -gt 1",
              timeout=30,
            )
            machine.execute("systemctl stop webphone-drill")
          '';
        };

        # OIDC lane VM test (plan T09, 2026-11-15): the typed
        # services.webphone.oauth2 lane gets its RUNTIME proof, not just
        # the eval assertions. The operator's secret file is handed to
        # systemd as a LoadCredential, the unit-private copy lands 0440
        # under /run/credentials, the app's rendered config points at
        # that copy, the SSO door is LIVE against a real stub IdP
        # (discovery is begin-eager, so the VM runs a loopback https
        # issuer with a locally-trusted cert — begin 302s to the
        # DISCOVERED authorization endpoint with client id + PKCE), a
        # dead IdP fails the ATTEMPT to /login?sso_error=unavailable
        # (boot-lazy discovery, never cached), and the chain FAILS
        # CLOSED: remove the source file and the unit start fails (the
        # crm.nix doctrine "a missing source file fails the UNIT START"
        # proven live). CADDY-FREE by ruling: port 443 belongs to the
        # stub, and the caddy-keyed derivations stay covered by the
        # module-output-oauth2 golden. KVM-gated like the backup VM test.
        webphone-oauth2 = pkgs.testers.runNixOSTest {
          name = "webphone-oauth2";

          requiredFeatures.kvm = true;

          nodes.machine =
            { pkgs, ... }:
            let
              # Self-signed leaf for the stub IdP; the VM's system trust
              # store adopts it so the app's Go HTTP client (system pool)
              # verifies https://idp.example.org without any app-side knob.
              idpCert = pkgs.runCommand "webphone-oauth2-idp-cert" { } ''
                mkdir $out
                ${pkgs.openssl}/bin/openssl req -x509 -newkey rsa:2048 -nodes \
                  -keyout $out/key.pem -out $out/cert.pem -days 3 \
                  -subj "/CN=idp.example.org" \
                  -addext "subjectAltName=DNS:idp.example.org"
              '';
              # Minimal OIDC discovery document over loopback https: the
              # only leg the app ever contacts on the begin path. The
              # token/jwks endpoints are never reached by this test (the
              # assert is the app's redirect, not the IdP's answer).
              idpStub = pkgs.writeText "webphone-oauth2-idp-stub.py" ''
                import http.server
                import json
                import ssl

                DISCO = json.dumps({
                    "issuer": "https://idp.example.org",
                    "authorization_endpoint": "https://idp.example.org/authorize",
                    "token_endpoint": "https://idp.example.org/token",
                    "jwks_uri": "https://idp.example.org/jwks.json",
                    "response_types_supported": ["code"],
                    "subject_types_supported": ["public"],
                    "id_token_signing_alg_values_supported": ["RS256"],
                    "scopes_supported": ["openid", "email", "profile"],
                }).encode()

                class Handler(http.server.BaseHTTPRequestHandler):
                    def do_GET(self):
                        if self.path.startswith("/.well-known/openid-configuration"):
                            self.send_response(200)
                            self.send_header("Content-Type", "application/json")
                            self.send_header("Content-Length", str(len(DISCO)))
                            self.end_headers()
                            self.wfile.write(DISCO)
                        else:
                            self.send_response(404)
                            self.end_headers()

                    def log_message(self, *args):
                        pass

                srv = http.server.HTTPServer(("127.0.0.1", 443), Handler)
                ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
                ctx.load_cert_chain("/etc/idp-cert.pem", "/etc/idp-key.pem")
                srv.socket = ctx.wrap_socket(srv.socket, server_side=True)
                srv.serve_forever()
              '';
            in
            {
              imports = [ ../package/nixos-module.nix ];
              services.webphone = {
                enable = true;
                package = self'.packages.webphone;
                # CADDY-FREE (2026-11-15 ruling): port 443 belongs to the
                # stub IdP and real ACME would be network noise; the
                # caddy-keyed derivations (callback, passkey rp) stay
                # covered by nix/module-output-oauth2.golden, so the VM
                # pins the callback explicitly instead.
                oauth2 = {
                  enable = true;
                  issuer = "https://idp.example.org";
                  clientId = "webphone-vm";
                  clientSecretFile = "/etc/webphone-oidc-source";
                  redirectUrl = "https://phone.example.org/api/auth/oauth2/callback";
                };
                settings = {
                  auth.users."lars@example.org".extensions = [ "1000" ];
                  auth.extension_password_files."1000" = "/etc/webphone-ext-1000";
                };
              };
              # The stub IdP: the issuer name resolves to loopback over a
              # locally-trusted TLS cert — the app's discovery leg runs
              # for real, no external network.
              security.pki.certificateFiles = [ "${idpCert}/cert.pem" ];
              networking.hosts."127.0.0.1" = [ "idp.example.org" ];
              environment.etc = {
                "webphone-oidc-source" = {
                  text = "oidc-vm-secret";
                  mode = "0600";
                };
                "webphone-ext-1000" = {
                  text = "ext-secret";
                  mode = "0600";
                };
                "idp-cert.pem" = {
                  source = "${idpCert}/cert.pem";
                  mode = "0644";
                };
                "idp-key.pem" = {
                  source = "${idpCert}/key.pem";
                  mode = "0600";
                };
                "idp-stub.py" = {
                  source = idpStub;
                };
              };
              environment.systemPackages = with pkgs; [
                curl
                python3
              ];
              system.stateVersion = "26.05";
            };

          testScript = ''
            machine.wait_for_unit("webphone.service")
            machine.wait_for_open_port(8080)

            # The credential landed: the unit-private copy exists, is
            # owner+group readable (0440 — systemd installs LoadCredential
            # snapshots at 0440, pinned EMPIRICALLY; not the 0400 of the
            # encrypted-credential mental model), and carries the bytes.
            machine.succeed(
                "test -f /run/credentials/webphone.service/webphone-oidc-client-secret"
            )
            cred_mode = machine.succeed(
                "stat -c %a /run/credentials/webphone.service/webphone-oidc-client-secret"
            ).strip()
            assert cred_mode == "440", f"credential mode is {cred_mode!r}, expected 440"
            machine.succeed(
                "grep -q 'oidc-vm-secret' /run/credentials/webphone.service/webphone-oidc-client-secret"
            )

            # The app READS the copy: the rendered config JSON (its store
            # path rides WEBPHONE_CONFIG) names the credential path — the
            # app must never see the host path in place.
            cfg_path = machine.succeed(
                "systemctl show webphone.service -p Environment "
                "| tr ' ' '\\n' | grep WEBPHONE_CONFIG= | cut -d= -f2-"
            ).strip()
            machine.succeed(
                f"grep -q '/run/credentials/webphone.service/webphone-oidc-client-secret' {cfg_path}"
            )
            machine.succeed(
                f"grep -q 'webphone-vm' {cfg_path}"
            )

            # The stub IdP goes LIVE before the begin probe: discovery is
            # boot-lazy but BEGIN-EAGER (sso.Challenge fetches the
            # well-known doc on every door push), so an honest redirect
            # assert needs a reachable issuer. Loopback https over the
            # locally-trusted cert; the name resolves via /etc/hosts.
            machine.succeed("python3 /etc/idp-stub.py >/tmp/idp-stub.log 2>&1 &")
            machine.wait_until_succeeds(
                "curl -sf https://idp.example.org/.well-known/openid-configuration "
                "| grep -q authorization_endpoint",
                timeout=60,
            )

            # The SSO door is LIVE: begin 302s to the DISCOVERED
            # authorization endpoint with the client id, PKCE method and
            # nonce in the query — the full config → parse → discover →
            # compose chain, proven over real TLS.
            location = machine.succeed(
                "curl -s -o /dev/null -w '%{redirect_url}' "
                "http://127.0.0.1:8080/api/auth/oauth2/begin"
            )
            assert location.startswith("https://idp.example.org/authorize"), f"begin did not redirect to the issuer: {location}"
            assert "client_id=webphone-vm" in location, f"begin lost the client id: {location}"
            assert "code_challenge_method=S256" in location, f"begin lost the PKCE method: {location}"
            assert "nonce=" in location, f"begin lost the nonce: {location}"

            # A DEAD IdP fails the ATTEMPT, not the process: discovery is
            # boot-lazy by design (the PBX phone must keep answering while
            # the IdP is down), and failures are never cached — this begin
            # re-fetches and lands the browser on the login door with the
            # unavailable reason code.
            machine.succeed("pkill -f idp-stub.py")
            machine.wait_until_fails(
                "curl -sf https://idp.example.org/.well-known/openid-configuration",
                timeout=60,
            )
            dead = machine.succeed(
                "curl -s -o /dev/null -w '%{redirect_url}' "
                "http://127.0.0.1:8080/api/auth/oauth2/begin"
            )
            assert "sso_error=unavailable" in dead, f"begin with an unreachable IdP did not fail closed to the login door: {dead}"

            # The chain FAILS CLOSED: remove the operator's source file
            # and the unit start dies (systemd cannot snapshot the
            # credential) — the crm.nix doctrine proven live, not the
            # silent-dark-seam shape.
            machine.succeed("mv /etc/webphone-oidc-source /etc/webphone-oidc-source.saved")
            machine.fail("systemctl restart webphone.service")
            machine.succeed("systemctl is-failed webphone.service")
            machine.succeed("mv /etc/webphone-oidc-source.saved /etc/webphone-oidc-source")
            machine.succeed("systemctl restart webphone.service")
            machine.wait_for_open_port(8080)
            machine.succeed("curl -sf http://127.0.0.1:8080/healthz")
          '';
        };

        # Backup/restore DRILL (plan T23, 2026-09-20): the VM test above
        # proves the online snapshot; this proves the RESTORE — tar the
        # data dir, reboot in a fresh location, pull a byte-identical
        # attachment back out of the restored store. Loopback-only, so
        # it runs in the sandbox (no KVM gate).
        webphone-backup-drill =
          pkgs.runCommand "webphone-backup-drill-check"
            {
              nativeBuildInputs = [ pkgs.python3 ];
              meta.description = "end-to-end backup/restore drill over a live webphone instance";
              meta.timeout = 300;
            }
            ''
              export WP_BIN=${self'.packages.webphone}/bin/webphone
              python3 ${self}/scripts/webphone-backup-drill.py
              echo "drill passed" > $out
            '';
      };
    };
}
