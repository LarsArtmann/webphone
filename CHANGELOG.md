# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-09-17

### Added

- Standalone webphone UI, extracted from
  [nix-international-telephony](https://github.com/LarsArtmann/nix-international-telephony)
  `packages/webphone` (the UI lived there since v0.1.0 of that stack).
  That stack now consumes this flake as its `webphone` input.
- The 1590-line single-file app is restructured into ES modules
  (`src/app/`: config, i18n, ui, state, auth, audio, notify, panels,
  ice, connection, calls, main) bundled by esbuild into one
  same-origin `app.js` — unchanged behavior, reviewable pieces.
- Light theme alongside dark: full design-token system following
  `prefers-color-scheme` (previously dark-only).
- LED-style status pill, tactile keypad/button press states, focus
  rings, `prefers-reduced-motion` support, incoming-call pulse, stricter
  toast styling — same DOM contract as before (see AGENTS.md).
- `package/update.sh` for repinning the bundled sip.js tarball.

[0.1.0]: https://github.com/LarsArtmann/webphone/releases/tag/v0.1.0
