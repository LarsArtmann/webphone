# Status Report — Visual Verification Loop (delta to 12:08 typography report)

**Written:** 2026-10-01 01:21 CEST
**Session scope:** The run since `docs/status/2026-09-30_12-08_typography-font-design-train.md`:
building the screenshot-based visual verification the owner asked for (pointer:
`~/projects/vision-review-agent` + my own image loading), capturing the webphone
UI matrix, and reviewing the first shots. No other research.
**Format note:** `.md` per the owner's standing instruction; skill's HTML default
overridden (flagged once, not re-litigated).

> ARCHIVED 2026-10-03 (docs-health v6 sweep): every finding closed by the
> 2026-10-01/02 trains — the language split brain fixed (i18n.js reads
> wp-lang first) with both pins landed, the harness persisted as
> scripts/ui-capture.py (T23) and the full 14-shot matrix captured +
> DOM-asserted; the owner legs (vision cross-check, shot disposition)
> ride the visual-harness TODO row. Per-item verdicts inline.

---

## Headline

The visual gate paid for itself on the first four screenshots: the `de_light`
capture shows **German server chrome (Nachrichten / Mailbox / Anrufe) beside an
ENGLISH island (Sign in / Extension / Password / Connect) — in one viewport**.
Root cause found in source: the island boot resolves language as
`localStorage → navigator.language` and **never reads the `wp-lang` cookie** the
server uses (`i18n.js:236-238` vs `pages.go:206-213`), and `applyI18n()`
additionally overwrites `document.documentElement.lang` at runtime — so for
cookie ≠ navigator cases the island both renders the wrong language and
clobbers the `html lang` the typography train just fixed. Pre-existing split
brain, exposed by the new gate, NOT caused by it.

---

## Self-Review (the hard questions, this run)

**What did I forget?**

- An identity check before trusting a port. I reused 18099 from the earlier
  lang-check without verifying WHAT answered; my seed POSTs got 403 from a
  foreign server, costing a full run before I diagnosed it from the response's
  CSP fingerprint (cqrs-htmx defaults: `nonce-…`, `unsafe-inline` styles,
  `report-uri` — not webphone's CSP) plus `/version` serving HTML instead of
  webphone's JSON probe.
- Durable output. The capture run's stdout lived only in a background-job
  buffer; the 13-hour session gap made it unfetchable, so the shot-5 death is
  undiagnosed. Should have teed to a file.

**What is stupid that we do anyway?**

- This shell's `kill` builtin **silently fails** (`kill: unsupported builtin`,
  hidden by my `2>/dev/null`) — every cleanup kill this session was a no-op. A
  stray dev server of mine survived ~13 hours until `pkill -9` did the job.
- `/tmp` one-shot tooling: the capture harness works but lives in /tmp with env
  vars for binaries; one session away from being lost.

**What could I have done better?**

- View the shots in the same breath as capturing them (the review IS the
  payoff; I captured 4, then got interrupted before looking).
- Capture chromedriver 4xx bodies from the start (I patched `http()` to raise
  with the body only after the first opaque 400).

**Did I lie?** No. But the 12:08 report's "html lang fix — fully done" deserves
a downgrade annotation: server-side it is correct and live-verified; the island
can still override `documentElement.lang` at runtime (i18n.js:258) for
cookie ≠ navigator users. The fix is incomplete without the island boot change.

**Ghost systems?** None created. The capture harness is /tmp-ephemeral — it is
NOT yet a repo tool (that is task 5, not a ghost).

**Split brains?** One REAL one found and pinned to source lines (the headline).
It predates this session; the typography train's Locale fix reduced its blast
radius server-side but the island half is still open.

**Tests?** Still the same gap as the 12:08 report — no new automated tests this
run; the finding above adds a second pin that must land with the fix (island
boot prefers the cookie).

**Scope creep?** Held: I researched vision-review-agent only far enough to
learn it analyzes but does not capture, then built the minimal capture path.

---

## a) FULLY DONE

~~1. **vision-review-agent recon** (`~/projects/vision-review-agent`): it is an~~ done — this session (report of record)
ANALYZE side — Go SDK + `vision` CLI (takes PNGs) + `visionreviewd` daemon
(reviews screenshots others capture: `view.captured` events, before/after
compare). No browser capture capability. Consequence: capture =
chromedriver REST, review = my own multimodal reading (used below), optional
later cross-check via the `vision` CLI.
~~2. **Capture harness written and proven end-to-end**: `/tmp/capture_webphone.py`~~ done — this session (report of record)
— pure-stdlib python driving chromedriver's REST API (no selenium);
seeds threads via `POST /hooks/message` (JSON + Bearer secret, per the
contract in webhooks.go:42-52) including a 190-char German message and an
MMS attachment; logs in through the island form (loopback mode skips PBX
verification); sends an outbound message through the app itself (in-page
`fetch` + CSRF from the meta tag, loopback gateway accepts); sets theme via
`data-theme` attribute and language via the `wp-lang` cookie; desktop
1440×1600 and mobile 375×812 rects.
~~3. **Four welcome shots captured** (en/de × dark/light, 88–92 KB PNGs in~~ done — this session (report of record)
`/tmp/webphone-shots/`) — proof the whole pipeline works: server boot →
seed → browser → theme/lang control → PNG.
~~4. **Port-squatter diagnosed and respected**: 18099 is held by a foreign app~~ done — this session (report of record)
(concurrent session's; identified by CSP fingerprint + HTML `/version`).
Left running per the AGENTS concurrent-session rule; moved to 18071 with a
`/version` identity check before seeding.
~~5. **Stray-process hygiene restored**: found my own 13-hour-old dev server~~ done — this session (report of record)
(442168), discovered the shell's `kill` builtin silently no-ops, killed it
with `pkill -9`. Cleanup now actually cleans.
~~6. **Visual verdict on the four welcome shots** (my own review):~~ done — this session (report of record)

- The desk-phone vernacular reads well in both themes: quiet slate, teal
  accent, LED-style uppercase OFFLINE pill, tidy hairline rows.
- `text-wrap: balance` visibly working: the welcome title breaks into two
  clean lines in BOTH languages ("One desk for calls, / messages, fax and
  voicemail." / "Ein Platz für Anrufe, / Nachrichten, Fax und Mailbox.").
- Server-side i18n verified in pixels: German nav + welcome copy render.
- Defect found: the island-language split brain (headline).
- Minor: the "remember extension" checkbox rides high against its two-line
  label (both themes).
- Caveat recorded: the capture box's fontconfig resolves system-ui to a
  DejaVu-class face with 400/700 only — antialiasing and intermediate-weight
  (550/650/750) verdicts from these PNGs are indicative, not final; layout,
  scale, spacing, contrast, and language verdicts stand.

## b) PARTIALLY DONE

~~1. **The capture matrix: 4 of 12 shots.** The run died at the login step~~ done — the T23 harness completed the 14-shot matrix (12:57 closeout; unique checksums)
(shots 5–12 missing: messages list, thread with bubbles, island keypad +
event log, settings, history, two mobile views). Error output lost to the
session gap; harness itself unproven past shot 4. Remaining: rerun with
output teed to a file; diagnose; finish. Effort: S.
~~2. **The visual review: 4 of 12 reviewed.** The substantive surfaces (bubbles~~ done — all surfaces captured + DOM-asserted (14/14)
with `text-wrap: pretty`, mono event log, keypad, settings tabular-nums,
mobile) do not exist as PNGs yet. Effort: S after (1).
~~3. **The language split brain: diagnosed, not fixed.** Source pinned~~ done — the island boot reads wp-lang first (i18n.js cookieLang + rationale comment); the split brain closed
(i18n.js:236-238 boot; :254 selector writes the cookie; :258 runtime
`documentElement.lang` override; pages.go:206 server cookie read). Fix
sketched: boot order cookie → localStorage → navigator.language. Not
implemented (report-first instruction). Effort: S + test.

## c) NOT STARTED

- The island boot fix + its two regression pins (island i18n test + the
  `html lang` test still owed from the typography train).
- Shots 5–12, their review, and any fixes they surface.
- Persisting the harness into the repo (`scripts/ui-capture.py` + AGENTS note)
  — currently /tmp-ephemeral.
- vision CLI cross-check (needs provider/key decision — g-1).
- Everything carried open from the 12:08 report (its §f remains valid and
  unharvested; see (f) here for the merged ranking).

## d) TOTALLY FUCKED UP

~~1. **The server↔island language split brain (product, pre-existing).** A user~~ done — fixed (i18n.js cookieLang reads the cookie before localStorage/navigator)
whose `wp-lang` cookie says `de` but whose navigator/localStorage say `en`
gets German tabs under an English phone, and their `html lang` flips to
`en` at runtime — screen readers announce German content with English
phonemes. Severity: medium (correctness/a11y, no data risk). Root cause:
i18n.js:236-238 + :258. Mitigation: the one-boot-line fix + pins (task 1-2).
~~2. **My process hygiene this run (mine, embarrassing).** (a) Trusted a port~~ process record
without an identity check → one wasted run + a foreign server's 403s to
decode; (b) cleanup kills that silently never executed (unsupported
builtin) → a stray server for ~13h; (c) capture stderr left in an evanescent
job buffer → the one failure I most needed to see is gone. All three have
named, cheap mitigations now (identity check; `pkill -9`; tee to file).
~~3. **Nothing in shipped code.** The repo tree is unchanged this run (the only~~ record stands
writes were /tmp harness + this report); last push `7008874` still HEAD.

## e) WHAT WE SHOULD IMPROVE

~~1. **Make the visual gate official**: persist the harness to~~ done — scripts/ui-capture.py persisted (T23) + the AGENTS run-recipe entry
`scripts/ui-capture.py`, document the matrix (light/dark × en/de ×
desktop/mobile) in AGENTS.md, and make "capture + look" a required step of
every CSS/markup train. It found in four shots what three test suites
could not see.
~~2. **Identity-check every endpoint before acting on it**: one `/version` (or~~ process record
known-body) probe would have saved the wasted run. Ports outlive sessions;
so do assumptions about them.
~~3. **Ephemeral outputs get files, not buffers**: background job stdout is not~~ process record
a log. Tee harness runs to `/tmp/<name>.log` (or a reports dir).
~~4. **In this shell, `kill` is a loaded gun with no bullet**: use~~ not adopted — the pkill lesson never landed in lessons.md; below the bar
`pkill -9 -f <pattern>` (or verify with `ps` after every kill). Worth a
line in docs/lessons.md.
~~5. **Report-first vs fix-on-sight for visual findings**: the split brain sat~~ owner — g3 unanswered; train practice settled on fix-on-sight
one `edit` away from fixed; the instruction said report. Owner should pick
the policy (g-3) — both are defensible, guessing is not.

## f) 50 things to get done next

New items from this run first (1–12), then the still-open carried items from
the 12:08 report (13–50, compressed; full rationale lives there). Impact
Critical/High/Medium/Low; Effort S (<30min)/M (30min–2h)/L (>2h). This section
is the HARVEST source for TODO_LIST.md / ROADMAP.md.

| #  | Task | Impact                                                                                                                                            | Effort   | Category |
| -- | ---- | ------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | -------- |
| ~~ | 1    | Fix island boot language: read `wp-lang` cookie first, then localStorage, then navigator (i18n.js:236-238) — closes the server↔island split brain | Critical | S        |
| ~~ | 2    | Pin it: island i18n test (cookie=de + empty localStorage → German) + `TestShellHtmlLangFollowsSessionLang` (still owed from typography train)     | Critical | S        |
| ~~ | 3    | Re-run the capture matrix with output teed to a file; diagnose the shot-5 login-step death                                                        | High     | S        |
| ~~ | 4    | Capture + review the remaining 8 shots (messages, thread bubbles, island keypad/log mono, settings, history, 2× mobile)                           | High     | S        |
| ~~ | 5    | Persist the harness: `scripts/ui-capture.py` + AGENTS.md "visual gate" note for CSS/markup trains                                                 | High     | M        |
| ~~ | 6    | Decide the vision CLI cross-check: provider/model/key for `vision <shots>` reviews, or own-eyes only (g-1)                                        | Medium   | S        |
| ~~ | 7    | Port hygiene habit: identity-check (`/version` body) before seeding/assuming on ANY local port; document in AGENTS.md                             | Medium   | S        |
| ~~ | 8    | Record this run's lessons in docs/lessons.md: unsupported `kill` builtin → `pkill -9`; foreign-CSP fingerprinting for port squatters              | Low      | S        |
| ~~ | 9    | ANNOTATE the 12:08 report: downgraded "lang fix fully done" (island runtime override open), link this report                                      | Medium   | S        |
| ~~ | 10   | Extend the stack browser-e2e with a de pass asserting `html lang` AND island language agreement                                                   | Medium   | M        |
| ~~ | 11   | Fix the "remember extension" checkbox alignment vs its two-line label (both themes)                                                               | Low      | S        |
| ~~ | 12   | Visually verify the styled 404 + `.wp-error` banner surfaces once (never eyeballed)                                                               | Low      | S        |
| ~~ | 13   | Run full `nix flake check` to close the typography train (treefmt, island-lint, module check, KVM backup VM)                                      | High     | M        |
| ~~ | 14   | CHANGELOG entry for the typography train (+ this fix when it lands)                                                                               | High     | S        |
| ~~ | 15   | Collapse font-family duplication: `--font-sans`/`--font-mono` tokens in app.css, island consumes `var()`                                          | High     | S        |
| ~~ | 16   | Token-parity test: app.css ↔ island/style.css mirrored blocks stay identical (island-only `--led`/`--key`/`--key-down` documented)                | High     | M        |
| ~~ | 17   | Contrast sweep at real sizes (muted on surface-2/3, 0.72–0.9em) vs WCAG 4.5:1                                                                     | High     | S        |
| ~~ | 18   | Settle the go.mod `go`-line flipflop (doctor: 8 changes/20 commits)                                                                               | High     | M        |
| ~~ | 19   | `hyphens: auto` on `.wp-bubble-body` + welcome body (lang now correct); verify German compounds                                                   | Medium   | S        |
| ~~ | 20   | `#log` 0.72rem readability bump + wrap check (keep English + greppable contract)                                                                  | Medium   | S        |
| ~~ | 21   | tabular-nums audit: voicemail `.len`, fax pages, settings `dd`, history `.when`                                                                   | Medium   | S        |
| ~~ | 22   | govalid-generate parallel contention: report upstream (verify-before-filing first), skip_steps, or tolerate (g-3 of 12:08)                        | Medium   | M        |
| ~~ | 23   | Investigate the 11:48:22 zero-byte mtime bumps on both CSS files                                                                                  | Medium   | S        |
| ~~ | 24   | AGENTS.md diet: 552 → ~≤400 lines; war stories → docs/lessons.md (after concurrent edits land)                                                    | Medium   | M        |
| ~~ | 25   | HARVEST both status reports into TODO_LIST.md / ROADMAP.md (docs-health)                                                                          | Medium   | S        |
| ~~ | 26   | README: Accessibility note (root scales with browser setting, lang per session)                                                                   | Medium   | S        |
| ~~ | 27   | vulnix "0/5 retries recovered": fix or gate vulnix to release builds                                                                              | Medium   | S        |
| ~~ | 28   | Focus-visible normalization (island input offset −1px vs 2px elsewhere)                                                                           | Low      | S        |
| ~~ | 29   | Update stale BuildFlow binary (e881e96 → 8dd634e)                                                                                                 | Low      | S        |
| ~~ | 30   | VACUUM buildflow cache.db (0.81 GB) / review state db (0.26 GB)                                                                                   | Low      | S        |
| ~~ | 31   | `text-wrap: pretty` on `.wp-welcome-body` / `.wp-welcome-hint`                                                                                    | Low      | S        |
| ~~ | 32   | Dark-mode +50 weight trial on a VF-capable box; adopt only if visibly better                                                                      | Low      | S        |
| ~~ | 33   | Keypad `✱` (U+2731) glyph sweep across platforms                                                                                                  | Low      | S        |
| ~~ | 34   | Heading-hierarchy audit of tab partials                                                                                                           | Low      | S        |
| ~~ | 35   | ANNOTATE reports as items close (docs-health ANNOTATE mode)                                                                                       | Low      | S        |
| ~~ | 36   | CSS-level structural checks beyond token parity (no px font-size; uppercase ⇒ letter-spacing)                                                     | Medium   | M        |
| ~~ | 37   | Prove island node:test suite has zero CSS coupling (one recorded run)                                                                             | Low      | S        |
| ~~ | 38   | Document island-only tokens (`--led`, `--key`, `--key-down`) in AGENTS.md mirror rule                                                             | Low      | S        |
| ~~ | 39   | Named type-scale tokens — only when a third stylesheet consumer appears (YAGNI)                                                                   | Low      | M        |
| ~~ | 40   | `prefers-contrast: more` adjustments                                                                                                              | Low      | M        |
| ~~ | 41   | Self-hosted display font for the brand wordmark only (owner taste call)                                                                           | Low      | L        |
| ~~ | 42   | `#wp-sse-live` light-mode visibility (0.45-opacity muted dot)                                                                                     | Low      | S        |
| ~~ | 43   | Cross-font check of the `·` separator in `.wp-signed-in-did`                                                                                      | Low      | S        |
| ~~ | 44   | Evaluate `font-variant-numeric: slashed-zero` for mono diagnostics                                                                                | Low      | S        |
| ~~ | 45   | og:locale per-session values: confirm no cache layer ever keys on them (noindex today)                                                            | Low      | S        |
| ~~ | 46   | Retire `/tmp/webphone-typo` + `/tmp/wp-*` artifacts when the train closes                                                                         | Low      | S        |
| ~~ | 47   | Document a dev-server port convention (never reuse across sessions; 180xx range)                                                                  | Low      | S        |
| ~~ | 48   | Owner verdict: visual-gate cadence (every CSS train vs on-demand) — g-1 here                                                                      | High     | S        |
| ~~ | 49   | Owner verdict: fix-on-sight vs report-first for defects found during visual review — g-3 here                                                     | High     | S        |
| ~~ | 50   | Keypad sub-label (0.55rem) contrast check                                                                                                         | Low      | S        |

Rejected-with-rationale list: unchanged from the 12:08 report §f 36–50 (palette
churn, webfonts, fluid scale, sentence-case pill, justified text, extra motion,
Tailwind-panels, island markup hooks, font-size-adjust, balance-on-body,
scrollbars, body tracking, VF optical sizing, preview hyphens, anti-prettier).

## g) Three questions I cannot answer myself

~~1. **Should the visual review spend AI tokens?** My own multimodal read caught~~ owner — routed TODO visual-harness row (vision-CLI provider/key)
the split brain without any API cost. Do you want a `vision` CLI cross-check
(vision-review-agent) as a second opinion on each matrix — and if so, which
provider/model/key am I allowed to burn? (I will not pick a paid endpoint
for you.)
~~2. **Confirm the 18099 occupant is yours/concurrent-session's and stays~~ resolved by events — the occupant was the concurrent session's; moot since
untouched.** I inferred it from the cqrs-htmx-default CSP fingerprint, the
HTML `/version`, and the AGENTS rule — but I am one wrong inference away
from killing a colleague's process someday. Say the word and I will treat
that fingerprint as "hands off" permanently.
~~3. **Fix-on-sight or report-first for defects found during visual review?**~~ owner — unanswered; train practice settled on fix-on-sight
The split brain waited ~13h in this report that one boot-line edit would
have closed. Which policy do you want for VISUAL-review findings (code
defects with pinned root cause), given the standing report cadence?

---

_Point-in-time snapshot, 2026-10-01 01:21 CEST. Artifacts: 4 PNGs in
`/tmp/webphone-shots/` (ephemeral — re-capture via the harness, do not archive
from /tmp), harness at `/tmp/capture_webphone.py` (task 5 persists it).
Annotate, never rewrite; (f) is the HARVEST source._
