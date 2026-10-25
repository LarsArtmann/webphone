# The services.webphone.oauth2 option surface: the typed front for the
# app's `auth.oauth2` config (OpenID Connect login; Pocket ID in the
# telephony stack). Split out of options.nix so each file stays focused
# (nix-review monolith guideline); imported by nixos-module.nix.
#
# Why a typed front at all: the app fails closed at BOOT on half an OIDC
# configuration, and enabling OIDC turns on the shared email→extension
# mapping, which the app derives passkey mode from (and then demands
# rp_id/rp_origins/password files for). Both facts surface at EVAL time
# here instead — an incomplete SSO config fails `nixos-rebuild` with a
# legible message, never a reboot into a dead login card.
{ lib, ... }:
{
  options.services.webphone.oauth2 = {
    enable = lib.mkEnableOption ''
      OpenID Connect (SSO) login: the login card gains a
      "Continue to Pocket ID" entry that runs the authorization-code +
      PKCE dance against `issuer`; the server exchanges the code,
      verifies the ID token, maps the verified email through the shared
      `settings.auth.users` map and mints the same webphone session the
      passkey and extension logins mint. The extension login stays
      available, so a dead IdP never locks the phone.

      Enabling this also requires the shared mapping
      (`settings.auth.users`) and one SIP-password file per mapped
      extension (`settings.auth.extension_password_files`) — both
      asserted. When `caddy.enable` is set the module derives the
      WebAuthn rp_id/rp_origins from `caddy.hostName` and the callback
      from the same host, so the vhost and the OIDC registration cannot
      drift apart.
    '';

    issuer = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "https://auth.example.org";
      description = ''
        The provider's OIDC discovery base URL: a bare https URL with no
        query or fragment (the ID token's `iss` claim is matched
        verbatim). Required when `enable` is set; discovery stays lazy
        inside the app, so an unreachable IdP costs a login, never the
        boot.
      '';
    };

    clientId = lib.mkOption {
      type = lib.types.str;
      default = "webphone";
      example = "webphone";
      description = ''
        Registered OIDC client id — the confidential client the IdP
        provisioner creates. Client creation stays an IdP concern; this
        module only consumes the registration.
      '';
    };

    clientSecretFile = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "/run/secrets/webphone-oidc-client-secret";
      description = ''
        HOST path of the client secret file (single line; trailing
        whitespace trimmed). Required when `enable` is set.

        The module hands this file to systemd as a credential
        (`LoadCredential`): systemd copies it into the unit's private
        credential directory at every unit start (mode 0400, owned by
        the service user), so the app reads the copy and never the host
        path in place — no host-side group grant needed, and a missing
        source file fails the UNIT START instead of the first SSO login.

        Declared as a STRING, never a Nix `path`: a path value would be
        copied into the world-readable Nix store and publish the secret
        to every user on the machine.
      '';
    };

    redirectUrl = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "https://phone.example.org/api/auth/oauth2/callback";
      description = ''
        The registered callback, absolute https; it must match the IdP
        registration verbatim or every login dies at the token exchange
        (invalid_redirect_uri).

        null (the default) derives
        `https://<caddy.hostName>/api/auth/oauth2/callback` when
        `caddy.enable` is set, so the vhost and the callback cannot
        drift. Behind an external front set it explicitly; eval fails
        closed when neither is available.
      '';
    };

    scopes = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [
        "openid"
        "email"
        "profile"
      ];
      description = ''
        OIDC scopes. MUST include "openid" (without it the IdP returns
        no id_token and no login can complete); "email" is what the
        identity mapping consumes.
      '';
    };

    allowedGroups = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [ ];
      example = [ "pbx-phone" ];
      description = ''
        In-app IdP group allowlist. Empty (the default) disables the
        check — admission stays with the users mapping and the IdP-side
        client-group link. Non-empty, every SSO login must carry at
        least one of these groups in its token claims or it is denied
        (fail-closed: a token with no groups claim at all is denied
        too). A second gate on top of the IdP link, not a replacement.
      '';
    };
  };
}
