# UI redesign via templ-components — session status (2026-10-04 08:17 CEST)

Task: "I hate our design — use templ-components and build something proper."
A full visual overhaul of the webphone app: new design system + deeper
templ-components adoption, without breaking the pinned DOM/byte contracts.

## a) Fully done

1. **Research** (complete): DOM contract (43 ids + script/brand/toast pins),
   every view template, full app.css (1980 lines) + island/style.css, the
   JS-touched class/id inventory (shell.js + all island modules), the
   byte-exact test pins (panels_test, server_test, messages/fax/voicemail/
   history/threads_flags/sse tests), i18n key mechanics, the tw.css
   build recipe (docs/planning tailwind-coexistence verdict), the theme
   model (theme-preload.js: auto = no attribute, manual = data-theme),
   and the PINNED templ-components v1.19.4 API (Button/Input/Textarea/
   FormFieldWrapper/BaseProps — NOT the local repo's 195-commits-ahead API).
2. **"Before" evidence**: 14 screenshots in `ui-shots-before/`.
3. **Design plan locked** — "Graphite desk + signal green": hueless
   graphite surface ramp with real elevation contrast, one green brand
   ramp (telephony on-the-line semantics), tighter radii (7/10/14),
   stronger type hierarchy, theme-independent avatar tints, filled-green
   outbound bubbles (polarity by fill), danger-voiced failed sends.
4. **New token system** in app.css AND island/style.css, mirrored —
   dark + light-auto + light-manual blocks; `internal/web/assets` tests
   PASS (token mirror + logical properties).
5. **app.css restyle** (~25 rule groups): solid header, wider body grid,
   nav weights/hover, panel padding + bigger h2, rows transparent with
   hover fill, unread = accent wash, avatars = oklch tint+ink (dark ink
   override for both light paths), bubbles (filled outbound, inverted
   ink, failed = solid danger + inverted minis, optimistic dashed
   outline), unread divider accent, unified input/textarea/select fill,
   thread-head action docking, borderless row-flag buttons.
6. **island/style.css**: token block + input fill (rest inherits).
7. **Phase-1 visual checkpoint**: 14 screenshots captured and reviewed —
   light+dark both dramatically cleaner; two flaws found and FIXED
   (bare textarea fill, thread-head wrap).
8. **templ-components adoption (markup)**: display.Button for ALL primary
   CTAs — messages (new send, reply send, thread-head Call with data-dial
   Attrs, load-older as secondary/sm), contacts (save, import as
   secondary), fax (send, with the htmx indicator passed as Button Icon
   via new `faxSendIndicator` component), settings (add snippet);
   forms.Input for contacts name/number (pinned ids kept) and fax "to";
   forms.Textarea for the snippet body. Imports added (display/forms/utils).
9. **templ generate** ran clean over internal/web/views.

## b) Partially done

1. **Phase 2 wiring** — templ files edited + generated, but the tree does
   NOT build: `go build` fails with `cannot find module providing package
   github.com/larsartmann/templ-components/forms: import lookup disabled
   by -mod=vendor`. The vendored tree predates the forms import.
   FIX (next step): `nix develop -c go mod vendor` (tidy not needed —
   go.mod already lists forms as indirect).
2. **Visual verification of phase 2** — none yet (blocked on the build).
3. **Full test battery** — only internal/web/assets run so far. The
   server-side markup-pin tests (messages/fax/voicemail/contacts/
   settings/panels/server) are UNRUN after the component swaps.

## c) Not started

1. **tw.css rebuild** — required for the adopted components to have any
   styling: commit `internal/web/assets/tw.css.input` (@import tailwindcss,
   `@custom-variant dark` mapped to `:root[data-theme="dark"]`, @theme
   remapping Tailwind's blue/gray/red/amber ramps onto the webphone
   tokens, @source = internal/web/views + the vendored adopted files:
   display/button_go.go, button.templ, empty_state.templ, forms/input*,
   label.templ, textarea*) + `scripts/build-tw-css.sh` (mirror of
   build-health-css.sh) + run with `nix run nixpkgs#tailwindcss_4 --minify`.
2. **Dead CSS removal** — `button.wp-primary` rules and `.wp-older button`
   rule are now orphaned (zero template references).
3. **layout.templ ThemeColor** — still old teal (#0f766e/#10161a); update
   to the new accent/bg hexes.
4. **favicon check** — island/favicon.svg may carry hardcoded old-teal.
5. **i18n re-verify** — keys unchanged (same literals via Button Text) but
   TestNoUnusedDictionaryKeys/TestEveryReferencedKeyExists unrun.
6. **Island JS tests** (`nix run nixpkgs#nodejs -- --test
   internal/web/assets/island-tests/*.test.mjs`).
7. **`nix fmt`** — mandatory after island/css edits (BuildFlow's formatter
   skips island files; drift has failed gates before).
8. **buildflow** full gate + `nix develop -c go test -count=1 ./...`.
9. **Final visual capture + before/after review**, contrast spot-check.
10. **webphone-smoke.py** 48-check against the rebuilt binary.
11. **Docs** — AGENTS.md templ-components adoption paragraph now says
    "only Base + EmptyState"; the tw.css input/script and the adopted
    component list change that. CHANGELOG entry for the redesign.
12. **Stack browser E2E** — served markup changed → release-runbook
    obligation (needs the nix-international-telephony stack; note for the
    release train, cannot run from this repo).

## d) Totally fucked up

1. **Two malformed multiedits on messages.templ** — I drafted new_strings
   with stray tab-runs inside button elements. One produced valid-but-
   ugly bytes (false alarm), the other initially failed to match and I
   burned three tool calls diagnosing my own damage (cat -A showed the
   truth). Root cause: sloppy new_string drafting under speed. Net file
   state is correct; the cost was pure churn.
2. **Build break by foresight failure** — importing templ-components/forms
   in vendor-mode WITHOUT running `go mod vendor` first is a predictable
   failure I should have chained immediately after templ generate.
3. **Skill/repo version trap** — the templ-components SKILL.md describes
   the LOCAL repo (195 commits ahead: Checkbox/Toggle/AppShell exist
   there, NOT in pinned v1.19.4). Caught before coding against it, but I
   should have read the module cache FIRST, not the skill catalogue.
4. **Staleness-guard thrash** — two multiedit batches bounced off the
   "file modified since read" guard (auto-commit daemon commits my own
   edits mid-flight); each cost a re-View + re-apply. Should re-View
   immediately after every daemon commit instead of retrying blind.

## e) What we should improve

- **Pin-verification cadence**: run the full server test battery after
  EACH view swap, not after all of them — a red test names the exact
  overreach while the diff is one file.
- **Vendor discipline**: any new library import → `go mod vendor` in the
  same breath as templ generate.
- **Contrast verification**: the new palette was eyeballed; a tiny
  WCAG-ratio script over the token pairs (text/bg, on-accent/accent,
  muted/surface) would make the a11y floor provable, not vibes.
- **Phase checkpoints**: rebuild + capture after each phase boundary —
  the phase-1 capture happened BEFORE the two phase-1 fixes, so their
  visual effect is still unverified.
- **Version-pin awareness in skills**: the templ-components skill should
  open with "check the CONSUMER's pinned version first" — the catalogue
  drift between local repo and v1.19.4 is a real trap.

## f) What to do next (ordered)

1. `nix develop -c go mod vendor` — unbreak the build (forms into vendor).
2. `nix develop -c go build ./...` — confirm compile.
3. Write `internal/web/assets/tw.css.input` (dark variant mapped to
   data-theme + @theme token remap + @source set).
4. Write `scripts/build-tw-css.sh` (staged build like build-health-css.sh)
   and run it with `nix run nixpkgs#tailwindcss_4`.
5. Remove dead CSS: `button.wp-primary` blocks, `.wp-older button`.
6. Update layout.templ ThemeColor hexes; check island/favicon.svg tints.
7. `templ generate` (if any .templ touch since last generate) then
   `nix develop -c go test -count=1 ./...` — fix any byte-pin fallout.
8. Island node tests.
9. `nix fmt` (island/css/go formatting).
10. Rebuild binary + full 14-shot capture; compare against
    ui-shots-before; iterate on visual nits (bubble meta contrast,
    keypad, toasts, empty states).
11. buildflow full gate (BUILDFLOW_NO_RESULT_CACHE=1 if time allows).
12. webphone-smoke.py 48-check.
13. Update AGENTS.md (templ-components adoption paragraph + tw.css input
    note) + CHANGELOG.
14. Note the stack browser-E2E obligation for the release train.
15. Self-review pass (brutal): grep for leftover wp-primary, orphaned
    selectors, i18n parity, dedup-registry sweep-log line if needed.

## g) Questions I cannot answer myself

1. **Palette sign-off**: I stayed in the green family (brand continuity
   with the old teal, theme-color/favicon untouched concept) but shifted
   from muddy teal to a cleaner signal-green on graphite. Keep this
   direction, or do you want a different accent family entirely?
2. **Dependency bump in scope?** The local templ-components repo is 195
   commits ahead of the pinned v1.19.4 (has Checkbox/Toggle/AppShell,
   better forms). Ride the pin and use only v1.19.4 (what this session
   did), or bump the dependency in this train to unlock the newer form
   kit?
3. **Copy in scope?** All UI copy is unchanged (en/de dictionaries
   untouched). The redesign exposed some clunky strings (e.g. "What this
   server is wired to."). Visual-only train, or include a copy pass?
