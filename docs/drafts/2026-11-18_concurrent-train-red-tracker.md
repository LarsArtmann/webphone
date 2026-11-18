> [!NOTE]
> This filing was drafted by GLM-4.6 via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

## Problem

`main` is red on two legs, both from the concurrent admin/SSO train's landed state (commit `a4a77d94`, 51 files) — tracked here so the fleet knows the red is attributional, not a regression from another train:

1. **island-js**: `internal/web/assets/island-tests/postcall-note.test.mjs` — 2 failing tests, now FULLY LANDED (no dirty postcall files; the red is on main, not in-flight):
   - `a failed autosave keeps the draft, marks the line, announces once per episode` — announce count `actual: 4, expected: 1` (strictEqual)
   - `flushPendingCallNotes sends the unsaved tail without waiting`

~~2. **treefmt**: `internal/server/session_switch_test.go` — non-canonical bytes (in-flight at last observation; the file is part of the same train's edit set).~~ → SELF-HEALED: `nix build .#checks.x86_64-linux.format` green at 19:0x once the train's wave finished.

## Evidence (2026-11-18 18:4x)

- Island suite full run: `347 pass / 2 fail`, both failures the postcall spec above.
- Same spec polled 5× over ~15 minutes spanning the train's landing: stable red (6 pass / 2 fail in-file, landed state).
- `env -u GOEXPERIMENT go test -count=1 ./...` — **0 failures** on the healed tree (the train's Go-side compile breaks self-healed mid-session: `usersByExtension` → `views.AdminUserLine` → `ssoStub` missing `MappedByLogin`, each observed then healed).
- `nix flake check` aborts at `treefmt-check` (the file above); the previous run aborted at `island-js`.
- Origin CI on the landed tip: `failure` (fast flake checks).

## Status

The Go side of that train self-healed mid-session (`usersByExtension` → `views.AdminUserLine` → `ssoStub` missing `MappedByLogin`, each observed then healed) and treefmt followed; the postcall announce-count expectation is the one survivor — the spec pins `once per episode` while the code announces 4×. Everything outside the train's surfaces is green on the same tree, including the newly landed `webphone-oauth2` VM test (`65299eb9`).

💘 Generated with Crush

Assisted-by: Crush:GLM-4.6
