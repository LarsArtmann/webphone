# UI Redesign Train — Brutal Self-Review & Status

_Date: 2026-10-12 06:44 CEST · Session: single Crush session (plus one concurrent
sipcred/passkey session on shared main) · Base: 0aa89d0-era tree, HEAD now
`4547c62`, CI green (`38203650017`, success)._

The task: "You never looked at the screenshots? I think it's ugly!" — look,
judge, redesign, verify. Verdict up front: **the redesign itself is done and
shipped**, every gate is green, but the session bled **~3 wall-clock hours on a
self-inflicted hang**, briefly **weakened the findings baseline**, and pushed a
**red-main window** onto shared main. Details, no varnish.

---

## a) FULLY DONE

1. **Looked at the actual UI** — reviewed all 9 surfaces × 2 themes (18 shots)
   before judging; formed the critique from evidence, not vibes.
2. **Design direction chosen and committed to**: "engineered instrument /
   desk-phone vernacular" — tighter radii, disciplined weights, one line-light
   motif, no template-default look. Avoided all three AI-default aesthetics
   (cream+serif, acid-green-on-black, broadsheet).
3. **Token refresh** in all 5 theme blocks of both stylesheets, mirrored:
   radii 7/10/14 → 6/9/12, light bg `#f2f4f7` → `#eef1f5`, stronger light
   card shadow, deeper borders. `TestCSSTokenBlocksAreMirrored` green.
4. **Emoji eradicated** (📞📌📥🔔🔇📎🗑⚙⏸▶): server renders
   `icons.Icon(name, "wp-icon")` from templ-components; JS-created buttons use
   SVG **string** builders (shell.js `WP_PLAY_SVG`/`WP_PAUSE_SVG`/
   `WP_ATTACH_SVG`, panels.js `playIcon`/`trashIcon`). Typographic glyphs
   (✓ ✕ ⧉ ☆ ✖ ↑ ↓ ✱ #) deliberately kept.
5. **Avatar fix**: unknown numbers now show the number's **last two digits**
   (`avatarFor`, helpers.go) instead of a country signum every same-country
   peer shared ("+4" four times); tests renamed/updated.
6. **Desktop desk-surface layout**: panels stretch to viewport height
   (>901px), transcript fills the stretch, reply composer pinned to the panel
   bottom via `#messages-panel > .wp-compose` (scoped — contacts/fax forms
   excluded after the first capture caught the over-broad version).
7. **Chrome polish**: glass (blurred) sticky header, weight hierarchy
   (nav 500/active 650), the **line-light glow** on the active tab underline,
   uppercase eyebrow section headings (`wp-settings-head`,
   `wp-transcripts h3`), dashed-quiet empty states, neutral file-picker
   buttons (killed the green "Choose Files" tint), offline banner as a danger
   rail with faint wash (was a pink box), `#wp-sse-live` halo when live.
8. **Contrast fix**: quiet minis on filled outbound bubbles (Transcribe) ride
   inverted ink — was muted-on-accent, near-invisible.
9. **Three capture-verified fix loops**: my own re-captures caught (a) the
   stretch selector missing (`#tab-content` interposed), (b) the composer pin
   spreading gaps into Fax/Contacts, (c) attachment names silently dropped by
   a templ children gotcha. All fixed and re-verified in shots.
10. **All tests green at push**: island 213/213 (node, `--test-force-exit`),
    full Go suite `./...` exit 0, `buildflow` exit 0, whitespace-drift 0,
    `nix build .#webphone.goModules` green, CI **success** on final HEAD.
11. **Memory/docs trail**: CHANGELOG `[Unreleased]` entry (incl. the stack-E2E
    obligation sentence); AGENTS.md now carries the icons rule, the
    stretch-seam note, and the avatar rule.
12. **Tail work on shared main**: `vendorHash` re-pinned and FOD-verified;
    error-code registry regenerated with the test's own `-update` after the
    concurrent train left it stale; pushed per the manual bar and watched CI
    to green.

## b) PARTIALLY DONE

1. **Visual verification is desktop-only**: the capture matrix is 1280-wide;
   no mobile (<900px), no de-locale, no passkey/enroll, no lightbox, no
   command palette, no incoming-call/keypad shots. The stretch layout is
   explicitly >901px and has never been eyeballed on a phone.
2. **Baseline hygiene**: my full `--save-baseline` (189 accepted) was caught
   by diffing and reverted to the ratified 19 — but the staleness WARN
   ("19 entries match nothing in the last 3 runs") remains open. The sites'
   `(baselined)` comments still exist; whether the WARN is a scoped-run
   artifact or a real ID drift is unresolved. Ratification is owner-only, so
   I stopped at "revert and report".
3. **Two-vs-one icon sources**: server trash = templ-components `Trash`,
   island trash = my hand path in panels.js; server attachments =
   `Document` icon, shell.js chip = my paperclip path. Both render fine but
   the glyph sets are not the same lineage. Unification pending.
4. **Docs trail**: CHANGELOG + AGENTS done; **docs/lessons.md has neither the
   `--test-force-exit` hang story nor the templ children gotcha yet** (both
   are this session's lessons; added to the report, entries still owed).
5. **The Send-button color**: templ-components Tailwind green vs token green
   — visibly two greens in dark mode. Deliberately left (adopted-system
   doctrine) but **undocumented as a ruling** — needs a decision, not a
   silent drift.
6. **Verification of the concurrency aftermath**: CI green and full Go suite
   green at HEAD, but a full local `nix build .#webphone` (binary closure)
   and `nix flake check` have not been run after the sipcred train landed.

## c) NOT STARTED

1. **Stack browser E2E re-run** — owed for a served-markup change
   (recorded in the CHANGELOG entry; release-runbook obligation).
2. **Mobile capture pass** in `scripts/ui-capture.py` (375-width × 2 themes).
3. **German capture pass** (the i18n dictionary changed: `media.preview`).
4. **Accessibility audit** of the new surfaces: contrast ratios on the
   glow/rail/empty styles, focus-visible on icon-only buttons, screen-reader
   pass over the icon buttons (aria-labels kept, but never listened to).
5. **Pixel-diff before/after report** — `scripts/render-diff.py` exists; I
   eyeballed only.
6. **Emoji drift guard** — a test banning emoji codepoints in served views/JS
   so the cheap look can't creep back.
7. **`nix flake check`** full run (KVM-gated backup VM test included) at a
   calm moment.
8. **Smoke script** (`webphone-smoke.py`) restart/boot-failure scenarios after
   the train noise.

## d) TOTALLY FUCKED UP

1. **The ~3-hour hang (worst failure of the session).** I ran the island
   tests WITHOUT the documented `--test-force-exit` — twice — and the node
   runner hung on a keep-alive handle. Then, instead of killing the hung
   jobs within seconds, I **waited on the background shells** (twice), so a
   6-second suite cost the session hours of wall-clock and your patience.
   The command WITH the flag is written in AGENTS.md Commands — I failed to
   read it. Double failure: wrong invocation AND bad hung-job hygiene.
2. **Weakened the findings baseline.** `buildflow --save-baseline` full-run
   accepted **189 findings** as reviewed noise (85 duplication, migration,
   configuration — none ratified) where the curated file holds exactly **19**
   branching-flow placements. Caught by diffing the JSON, reverted via
   `git restore`; the correct scoped form (`buildflow -s branching-flow
   --save-baseline`) was in AGENTS.md the whole time. Same root cause as #1:
   didn't re-read the documented procedure before running it.
3. **A red-main window on the manual push.** At bar expiry I pushed HEAD
   (`3c56543`) which carried the concurrent session's MID-TRAIN commits —
   registry drift + vendor bump half-landed — so main went red for ~10
   minutes. A finer-grained `git push origin 0477af9:main` (last known-good,
   fast-forward) would have kept main green while their tail landed. The bar
   sanctioned *a* push; it didn't force pushing *the tip*.
4. **A false closing claim.** My handoff said the `--test-force-exit` lesson
   is "in AGENTS.md now" — it was ALWAYS in AGENTS.md; I simply hadn't read
   it. The honest statement is "I didn't read the documented command", not
   "memory captured the lesson".

## e) WHAT WE SHOULD IMPROVE

1. **Read the runbook line before running the command** — both fuckups above
   are "proceeded from memory of a similar command" failures. The rule that
   already existed was right.
2. **Hung-job reflex**: a background job with no output after ~60s gets
   `job_kill` + a different approach, never a passive wait. (Also the
   Kill-builtin is broken in this shell — use python `os.kill`.)
3. **Diff-check destructive-ish tool output against the ratified state** —
   the 189-vs-19 baseline catch came from `git show HEAD:file | jq`-style
   comparison. Make that reflex immediate, not post-hoc.
4. **Scoped gates by default**: `buildflow -s <rule> --save-baseline`, never
   a bare full save; likewise `git add -A` twice reached for their in-flight
   files before I noticed — pathspec commits (`git commit -- <paths>`) were
   the correct default all along.
5. **Preferring known-green pushes**: when bar-pushing a mixed daemon queue,
   push the last known-green commit first (`git push origin <sha>:main`),
   let the train land, push the tip.
6. **Verify the "before" is preserved**: I copied old shots to /tmp
   (ephemeral). Git history preserved them via daemon commits, but by luck of
   timing, not plan.
7. **Self-review with the artifact, always**: the three real design bugs were
   all caught by re-capturing and *looking* — that loop is the reason the
   shipped state is clean. Make the re-capture loop non-negotiable for any
   visual change.

## f) NEXT 50 (ordered roughly by impact, UI-train first)

1. Stack browser E2E re-run (owed, release-runbook).
2. Mobile capture pass in `ui-capture.py` (375×812, both themes).
3. German capture pass (i18n strings changed this train).
4. Emoji drift guard: views/island test banning emoji codepoints in served
   markup and JS strings.
5. Unify trash glyph: use templ-components `Trash` path data in panels.js.
6. Unify attachment glyph: shell.js chip should carry the same `Document`
   path the server bubbles render (currently a paperclip shape).
7. Accessibility audit: contrast (glow, rail, dashed empties, inverted-ink
   minis), focus-visible on `wp-icon-btn`, screen-reader pass.
8. Full local `nix build .#webphone` at HEAD (binary closure, post-train).
9. Full `nix flake check` at a calm host moment (includes KVM backup VM).
10. `webphone-smoke.py` restart/boot-failure scenarios at HEAD.
11. Keyboard-only walkthrough + screenshots (nav → rows → composer →
    palette).
12. Lightbox (`wp-lightbox`) visual check against new tokens (not in matrix).
13. Command palette overlay visual check (M8; not in matrix).
14. Incoming-call card + keypad visual pass (needs a real SIP call).
15. Voicemail player wave on dark (needs phone API; stack env).
16. Enroll page (passkey-gated) token pass — not in the matrix.
17. `health.css` token-drift check (separate Tailwind build, untouched).
18. `theme-color` meta vs new `--bg` (theme-preload interplay).
19. Header blur on iOS Safari (`-webkit-backdrop-filter` fallback verify).
20. Row hover-reveal on coarse pointers (44px targets) — touch-device pass.
21. Thread-head icon cluster alignment with h2 baseline + Call button.
22. Dark inbound bubble separation (surface-2 on surface reads flat).
23. Settings dl: hairline separators between rows for scan-ability.
24. Composer wrap review at ~1000px width (file input truncation point).
25. `wp-thread-side` timestamp alignment on pinned rows (visual polish).
26. Capture-harness seed determinism: duplicate identical bubbles across
    shots (2/3/4 sends of the same text in one run) — explain or fix.
27. `render-diff.py` before/after pixel report for this redesign (evidence).
28. Decide: `ui-shots-before/` folder is stale (14 shots, pre-T23) — refresh
    or delete.
29. Baseline decision: scoped `branching-flow` re-save (kills the staleness
    WARN) — needs owner ratification.
30. Two-greens ruling: token-align the templ-components Send button or bless
    the library green as "system action color" — needs owner decision.
31. `--danger-strong` is island-only (app.css lacks it) — document or
    converge the token sets.
32. Avatar collision note: last-two-digits collide across peers ending alike
    — accept and document, or widen to 3 on collision.
33. FEATURES.md: decide whether a visual redesign warrants a line (docs-health
    says features are functional inventory — probably skip, record the skip).
34. docs/lessons.md: the `--test-force-exit` hang story + the templ children
    gotcha (both owed from this train).
35. Release dance when ready: CHANGELOG cut → v2.9.0 train (release.sh gates:
    vulnix, whitespace, E2E).
36. Check `#wp-sse-live` halo vs the mobile bottom tab bar overlap.
37. Eyeball snippet-picker open list (opens upward) against new shadows.
38. `wp-snippet-row`/manage list in Settings — visual pass with ≥3 snippets.
39. Toast stack visual (ok/warn/error rails) — not in the matrix.
40. History transcript `details` open state (transcript text) — not in
    matrix; eyeball.
41. Narrow-viewport island order rules (`:has(.call-card)` sticky) — visual
    verify with a live call on a phone width.
42. Consider `prefers-reduced-contrast` handling for the glow (currently only
    reduced-motion is honored).
43. Consistent transition durations audit (120ms everywhere — confirm no
    stragglers).
44. `ui-capture.py`: assert the eyebrow styling markers (e.g., uppercase
    `REPLY SNIPPETS`) so hierarchy can't silently regress.
45. Confirm backdrop-filter needs no CSP fiddling behind Caddy on prod.
46. Document the text-glyph set (✓ ✕ ⧉ ☆ ✖ ↑ ↓ ✱ #) as RATIFIED so the next
    audit doesn't re-litigate them.
47. Consider a `prefers-color-scheme` transition (none today — flash-free via
    theme-preload; verify after token changes).
48. Perf: `backdrop-filter` on the header — measure paint cost on low-end.
49. Move `ui-shots-prev` preservation into git (tag or branch the shots) if
    before/after evidence matters long-term.
50. Post-train docs-health sweep: AGENTS/CHANGELOG/TODO_LIST freshness after
    the sipcred + redesign trains.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **The two-greens ruling**: should the templ-components Send button be
   forced onto `--accent` (one green everywhere, overriding the adopted
   library's utilities from app.css), or is the brighter library green the
   intended "system action" accent? This is doctrine + taste; both are
   defensible and I won't guess.
2. **Capture-harness budget**: are the mobile-width and de-locale passes
   worth adding to `ui-capture.py` (roughly doubles runtime, still
   local-only), or does the stack E2E own those dimensions?
3. **Baseline ratification**: should I run the scoped
   `buildflow -s branching-flow --save-baseline` to clear the staleness WARN
   (shrinking the accepted set to whatever still fires), or freeze the
   current 19-entry file until the error_rules-bearing BuildFlow release and
   treat the WARN as noise?

---

_Arte in Aeternum — and read the documented command before running it._
