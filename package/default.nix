# The webphone static site: the UI in ../src plus a self-contained SIP.js
# bundle, installed under share/webphone in exactly the layout the nginx
# vhost copies to its document root (index.html, style.css, favicon.svg,
# app.js, sip.min.js, sip.min.js.LEGAL.txt). config.js is NOT part of the
# package: the serving PBX renders it at runtime (short-lived TURN
# credentials) — see README.md for the window.PBX_CONFIG contract.
{
  esbuild,
  fetchurl,
  lib,
  stdenv,
}:

let
  sipJs = fetchurl {
    url = "https://registry.npmjs.org/sip.js/-/sip.js-0.21.2.tgz";
    hash = "sha256-q783S8z1D90PZZ4k3XCoOgutCVU70W+ktslAIQU/Zss=";
  };
in
stdenv.mkDerivation {
  pname = "webphone";
  version = "0.1.0";

  src = ../src;

  nativeBuildInputs = [ esbuild ];

  buildPhase = ''
    runHook preBuild

    mkdir -p work
    tar -xzf ${sipJs} -C work

    # sip.js as a classic browser global (SIP.*). Pinned tarball, no CDN,
    # no runtime fetch — the served page is fully offline-capable.
    esbuild work/package/lib/index.js \
      --bundle \
      --minify \
      --format=iife \
      --global-name=SIP \
      --outfile=sip.min.js

    # The app itself: ES modules under app/ bundled to one classic script
    # (single same-origin asset; strict CSP serves no other script).
    # Deliberately NOT minified: the telephony stack's VM test asserts on
    # identifiers inside this bundle (blindTransfer, titleFlashStart, ...)
    # and an operator greps the served app.js when debugging. See
    # AGENTS.md (bundle contract) before changing either half.
    esbuild app/main.js \
      --bundle \
      --format=iife \
      --outfile=app.js

    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall

    mkdir -p $out/share/webphone
    cp ./index.html ./style.css ./favicon.svg ./app.js ./sip.min.js $out/share/webphone/
    cp work/package/LICENSE.md $out/share/webphone/sip.min.js.LEGAL.txt

    runHook postInstall
  '';

  meta = {
    description = "Standalone SIP.js WebRTC softphone UI, packaged as a static site";
    homepage = "https://github.com/LarsArtmann/webphone";
    license = lib.licenses.mit;
    platforms = lib.platforms.linux;
    maintainers = [
      {
        name = "Lars Artmann";
        github = "LarsArtmann";
      }
    ];
  };
}
