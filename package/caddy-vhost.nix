# The generated Caddy vhost body for services.webphone.caddy: a
# catch-all reverse_proxy plus the paths that need explicit treatment
# (the SIP WebSocket bridge, gated on sipUpstream; unbuffered /events
# and /health/* so SSE bytes reach the browser). Extracted from
# nixos-module.nix so the module wiring and this embedded Caddy program
# stay separately reviewable (nix-review monolith guideline).
{
  lib,
  cfg,
  listenPort,
}:
''
  encode zstd gzip
  ${lib.optionalString cfg.caddy.hsts.enable ''
    header Strict-Transport-Security "max-age=${toString cfg.caddy.hsts.maxAge}"
  ''}
  ${lib.optionalString (cfg.caddy.sipUpstream != null) ''
    handle ${cfg.settings.websocket_path} {
      reverse_proxy ${cfg.caddy.sipUpstream}
    }
  ''}
  handle /events {
    reverse_proxy 127.0.0.1:${listenPort} {
      flush_interval -1
    }
  }
  # The health dashboard subtree (only served when the operator
  # enabled dashboard.enable in settings): same unbuffered
  # proxying as /events — the dashboard's SSE stream must not
  # buffer behind the front.
  handle /health/* {
    reverse_proxy 127.0.0.1:${listenPort} {
      flush_interval -1
    }
  }
  handle {
    reverse_proxy 127.0.0.1:${listenPort}
  }
''
