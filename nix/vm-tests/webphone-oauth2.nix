# The KVM-gated OIDC-lane VM test: the typed oauth2 lane against a real stub IdP.
_: {
  perSystem =
    {
      pkgs,
      self',
      ...
    }:
    {
      checks = {
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
              # 3650 days: a runCommand cert is CACHED by its input hash —
              # a short -days is a time bomb that reds the check N days
              # after its first build with zero tree change (found
              # 2026-11-21: the -days 3 cert expired under the suite).
              idpCert = pkgs.runCommand "webphone-oauth2-idp-cert" { } ''
                mkdir $out
                ${pkgs.openssl}/bin/openssl req -x509 -newkey rsa:2048 -nodes \
                  -keyout $out/key.pem -out $out/cert.pem -days 3650 \
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
              imports = [ ../../package/nixos-module.nix ];
              services.webphone = {
                enable = true;
                package = self'.packages.webphone;
                # CADDY-FREE (2026-11-15 ruling): port 443 belongs to the
                # stub IdP and real ACME would be network noise; the
                # caddy-keyed derivations (callback, passkey rp) stay
                # covered by nix/testdata/module-output-oauth2.golden, so the VM
                # pins the callback explicitly instead.
                oauth2 = {
                  enable = true;
                  issuer = "https://idp.example.org";
                  clientId = "webphone-vm";
                  clientSecretFile = "/etc/webphone-oidc-source";
                  redirectUrl = "https://phone.example.org/api/auth/oauth2/callback";
                };
                settings = {
                  auth = {
                    users."lars@example.org".extensions = [ "1000" ];
                    extension_password_files."1000" = "/etc/webphone-ext-1000";
                    # oauth2 turns on the shared identity layer, whose
                    # passkey-mode derivation must resolve even with the
                    # ceremony off — without caddy the rp keys are set
                    # explicitly (same origin the front would serve).
                    passkey.rp_id = "phone.example.org";
                    passkey.rp_origins = [ "https://phone.example.org" ];
                  };
                };
              };
              # The stub IdP: the issuer name resolves to loopback over a
              # locally-trusted TLS cert — the app's discovery leg runs
              # for real, no external network.
              # The stub IdP as a boot-started unit (the house pattern —
              # the stack suites' itsp/gemini/telnyx stubs): a bare
              # `python3 /etc/idp-stub.py &` through the driver's shell
              # session DIES with the session — the v2.15.0 preflight
              # found discovery waiting 60s on a stub that never lived
              # (a failed unit fails the suite with its journal instead).
              systemd.services.idp-stub = {
                description = "Loopback stub OIDC issuer (webphone-oauth2)";
                after = [ "network.target" ];
                wantedBy = [ "multi-user.target" ];
                serviceConfig.ExecStart = "${pkgs.python3}/bin/python3 /etc/idp-stub.py";
              };
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

            # The stub IdP is a boot-started unit (see the node config):
            # discovery is boot-lazy but BEGIN-EAGER (sso.Challenge
            # fetches the well-known doc on every door push), so an
            # honest redirect assert needs a reachable issuer. Loopback
            # https over the locally-trusted cert; the name resolves via
            # /etc/hosts.
            machine.wait_for_unit("idp-stub.service")
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

            # The IdP DYING is invisible to the CACHED door: the positive
            # discovery cache (15 min TTL, successes only) still composes
            # the redirect — the failure would surface at the callback's
            # token exchange, not at begin.
            machine.succeed("systemctl stop idp-stub.service")
            machine.wait_until_fails(
                "curl -sf https://idp.example.org/.well-known/openid-configuration",
                timeout=60,
            )
            cached = machine.succeed(
                "curl -s -o /dev/null -w '%{redirect_url}' "
                "http://127.0.0.1:8080/api/auth/oauth2/begin"
            )
            assert cached.startswith("https://idp.example.org/authorize"), f"cached begin lost the composed redirect: {cached}"

            # A FRESH process with a dead IdP fails the ATTEMPT, not the
            # boot: discovery is boot-lazy by design (the PBX phone must
            # keep answering while the IdP is down), the restart proves
            # the app comes up WITHOUT a reachable issuer, and failures
            # are never cached — this begin re-fetches and lands the
            # browser on the login door with the unavailable reason code.
            machine.succeed("systemctl restart webphone.service")
            machine.wait_for_unit("webphone.service")
            machine.wait_for_open_port(8080)
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
            # The unit CYCLES on Restart=always (restart counter climbing,
            # ~100ms failed windows), so steady-state probes race the
            # cycle — assert the ESSENCE instead: start attempts keep
            # failing, the journal names the exit-code result, and the
            # app never comes up (the error-drill 2026-10-29 lesson in
            # its honest form for an always-restarting unit).
            machine.wait_until_succeeds(
                "test \"$(systemctl show webphone.service -p NRestarts --value)\" -gt 2",
                timeout=30,
            )
            machine.succeed(
                "journalctl -u webphone.service --no-pager "
                "| grep -q \"Failed with result 'exit-code'\""
            )
            machine.fail("curl -sf http://127.0.0.1:8080/healthz")
            machine.succeed("mv /etc/webphone-oidc-source.saved /etc/webphone-oidc-source")
            machine.succeed("systemctl restart webphone.service")
            machine.wait_for_open_port(8080)
            machine.succeed("curl -sf http://127.0.0.1:8080/healthz")
          '';
        };
      };
    };
}
