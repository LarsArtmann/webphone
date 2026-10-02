# Caddy Front + Diagram Rework — Status & Self-Review — 2026-09-30 15:49

Scope: this session's second phase (13:00–13:45: "why any nginx?" → Caddy
module front, diagram rebuild) plus what was noticed at report time
(15:49). The morning review-series phase has its own report at
`2026-09-30_12-46_review-series-status-and-self-review.md`. Only
session-authored work and session-noticed state below.

> ARCHIVED 2026-10-03 (docs-health v6 sweep): the caddy front shipped in
> v2.8.0 (Breaking NixOS module; the stack guard flip `5067ef4` rode the
> relock as designed), the AGENTS trim debt was paid at the 2026-10-02
> compaction, and the verification debt (validate/adapt, caddy VM) rides
> the ROADMAP stragglers; the owner eyeballs (diagrams, stack-front,
> HSTS) stay unanswered. Per-item verdicts inline.


## Context noticed at report time (not my work)

Concurrent sessions shipped **v2.8.0** (tag exists; stack commit
`3afcf57` pins it) — and the release carries this session's module
change. The stack relock included EXACTLY the 2-line nginx→caddy guard
flip this session documented in the release-tail TODO row
(`5067ef4 fix(telephony): flip webphone vhost guard to caddy`) — the
delegation-by-documentation path worked. The review roadmap deltas
(Receipt.Resolution, config fail-closed typing) also landed
(`0aab677 feat: post-v2.8.0 train`). No live fire from my change.

## a) FULLY DONE

~~1. Answered "why any nginx?" with evidence, not opinion: the module's~~ done — this session (report of record; shipped in v2.8.0 Breaking NixOS module)
   nginx generator had ZERO users — the stack force-disables it
   (web.nix:176) AND asserts it off (default.nix:105); the real front
   is the stack's own vhost.
~~2. Found + fixed a LATENT BUG while porting: the old vhost proxied the~~ done — this session (report of record; the latent wss dead-end fix, caddy.sipUpstream)
   SIP WebSocket path to the Go app, which never terminates wss
   (verified: zero websocket deps in go.mod, no Hijack). New
   `caddy.sipUpstream` gates the bridge; null leaves routing to the
   deployment (the stack's shape).
~~3. Module: `nginx.*` → `caddy.{enable,hostName,sipUpstream,hsts}` —~~ done — this session (report of record)
   automatic HTTPS, `encode zstd gzip` unconditionally (gzip option
   dropped with its rationale), `/events` flushed unbuffered
   (`flush_interval -1`), csrf fronting defaults re-derived from
   caddy.enable/hostName with the pinned precedence ladder.
~~4. `nix/module-check.nix` rewritten: caddy stand-ins; `vhost-extras`~~ done — this session (report of record)
   pins the streaming directives; NEW `vhost-sip-bridge` case pins the
   bridge gate BOTH ways (no handle by default; handle + upstream when
   set); nginx-gzip case removed with its option; csrf tests reworded.
~~5. Rendered the generated Caddyfile via a standalone `nix eval` and~~ done — this session (report of record)
   eyeballed it (clean: encode → optional HSTS → optional sip bridge →
   /events flush → catch-all).
~~6. Docs swept: README (8 edits incl. the Deployment example →~~ done — this session (report of record)
   Caddyfile), AGENTS module bullets, CHANGELOG **Breaking (NixOS
   module)** entry, FEATURES row, ROADMAP HSTS line, smoke.py
   docstring. SECURITY.md deliberately untouched (it truthfully
   describes the STACK's nginx).
~~7. Stack repo deliberately NOT edited (would break its then-current~~ done — this session (report of record; the 2-line stack flip landed via the relock `5067ef4`)
   lock); the required 2-line fix recorded in the release-tail TODO
   row → landed via the relock train as designed.
~~8. Diagrams rebuilt end-to-end: color-class system matching the report~~ done — this session (report of record)
   palette, `sql_table` storage, `stored_data` blob tree, ≤3 depth
   levels, verb-labeled edges; ELK chosen over dagre by MEASURED aspect
   ratio (1.25 vs 2.3); label + color presence verified in the SVGs.
~~9. Gates: `nix fmt` clean; targeted `nix build .#checks…webphone-module`~~ done — this session (report of record; full nix flake check exit 0)
   green across 3 iterations; **full `nix flake check` exit 0**
   (treefmt, deadnix, statix, island-lint, island-js, backup drill,
   sandboxed go tests, the rewritten module check).

## b) PARTIALLY DONE

~~1. **Generated Caddyfile is string-asserted, never runtime-validated.**~~ routed — ROADMAP 2026-10-02 stragglers (caddy validate/adapt + caddy VM coverage)
   No `caddy validate/adapt` pass over the rendered extraConfig; no VM
   test boots caddy + webphone and exercises TLS proxy, /events
   streaming, or the sip bridge. (The nginx front had the same gap —
   but "the old one was also untested" is an excuse, not a fix.)
~~2. **Live runbooks not re-swept.** My stale-ref grep excluded `docs/`~~ done — the live docs carry no stale module-front refs (runbook grep clean; the stack flip rode the v2.8.0 tail)
   as history; `docs/release-runbook.md` is LIVE and could reference
   the module front. Unchecked at report time.
~~3. **Diagrams: structurally verified only** (render success, geometry,~~ owner — g1 never answered; the SVGs stand geometry-verified
   label/color presence). No pair of eyes — mine included — has judged
   them against the owner's "YES! That's a diagram!" bar.
~~4. **`caddy.hsts` caveat documented but unverified**: the option text~~ not adopted — below the bar (the fencing example never pinned)
   warns Caddy serves the header on both schemes unless fenced; no test
   or example pins the https-only fencing pattern.

## c) NOT STARTED

~~1. Stack-front migration nginx→Caddy (the stack's OWN vhost — separate~~ owner — g2 unanswered; stack-repo architecture call, no webphone leg
   repo, separate blast radius: fail2ban nginx scanner, E2E, runbook).
   Out of this repo's scope by design; owner appetite unknown.
~~2. A caddy-equipped VM test (KVM harness exists for backup; no front~~ routed — ROADMAP 2026-10-02 stragglers (caddy VM test)
   coverage).
~~3. The d2-syntax skill reference fix (crush-config repo — stale~~ other repo — crush-config owns the d2-syntax reference; not filed
   `border-dashed` keyword; carried from the morning report).

## d) TOTALLY FUCKED UP

~~1. **Two phase boundaries lost their narrative commits to daemon~~ process record — the narrative-commit lesson
   races.** Both the caddy change and the diagram rebuild landed as
   `chore: auto-commit (heuristic)` because I batched edits too wide
   before committing. The detailed rationale lives only in CHANGELOG +
   this report. Fix pattern: commit after EACH verified sub-step, not
   after the whole phase.
~~2. **AGENTS.md cap violation**: the "line-neutral reword" I planned~~ resolved by events — the 2026-10-02 closeout compaction (404→377) paid the trim debt
   became ~+7 lines (the two bullets now carry the why-nginx-died
   rationale). The 377-line cap rule says an add must pay for itself;
   it didn't. Next AGENTS touch owes a compensating trim.
~~3. **66 MiB imagemagick fetched to produce a JPG this model cannot~~ process record
   see.** Discovered the view limitation AFTER downloading, rendering,
   converting. Should have confirmed image-viewing capability BEFORE
   pulling heavy tools. (Bonus: the fallback — objective geometry
   metrics — was the right pivot, but it should have been the FIRST
   plan.)
~~4. Minor, caught in-flight: one multiedit reported "10 applied, 1~~ process record
   failed" (csrf comment) — caught by grep immediately; the sed-edit
   tripped the edit-guard once — recovered by re-reading. No damage.

## e) WHAT WE SHOULD IMPROVE

~~1. Commit cadence vs the daemon: smaller commit windows; the daemon~~ process record
   will always win a race against a 15-minute batch.
~~2. Capability check before tool acquisition (image view, model limits).~~ process record
~~3. Runtime validation tier for generated config: a `caddy adapt~~ routed — ROADMAP stragglers (caddy validate/adapt in the module check)
   --validate` (or `nix run nixpkgs#caddy -- validate`) pass in the
   module check would catch Caddyfile syntax drift for free.
~~4. Live-doc sweep should be a distinct checklist item from~~ process record
   code+top-docs (runbooks, error-contract, lessons are live docs).
~~5. Diagram QA protocol for agents who cannot see: state it UP FRONT~~ process record
   ("I will verify geometry + labels; you judge aesthetics") instead of
   discovering it mid-task.

## f) NEXT (prioritized)

P1 — close this change's verification debt:

~~1. `caddy validate`/`adapt` pass over the rendered extraConfig (add to~~ routed — ROADMAP 2026-10-02 stragglers (caddy validate/adapt)
   module-check or a devshell alias).
~~2. Check `docs/release-runbook.md` (and other live docs) for stale~~ done — no stale module-front refs in the live docs (runbook grep clean)
   module-front references; fix.
~~3. Owner eyeballs the two new diagrams (see question 1); iterate on~~ owner — g1 unanswered; iterate-on-feedback never triggered
   concrete feedback only.
~~4. caddy-equipped VM test: boot webphone+caddy.enable+sipUpstream~~ routed — ROADMAP 2026-10-02 stragglers (caddy VM test)
   stand-in; assert /healthz through TLS, /events streams, sip bridge
   hits the fake PBX port.
   P2 — hardening + follow-through:
~~5. Example `remote_ip` fencing snippet for probes/metrics in README~~ done in part — README carries the Caddy remote_ip fencing pattern for the health hub (:447); a generic probes/metrics snippet not added
   (the option docs reference the pattern but show nothing).
~~6. HSTS fencing example (https-only matcher) next to `caddy.hsts`.~~ not adopted — below the bar (the hsts caveat stays documented in the option text)
~~7. AGENTS.md compensating trim (-7 lines owed).~~ done — the 2026-10-02 closeout compaction (404→377) paid the debt
~~8. d2-syntax skill fix at crush-config source (border-dash →~~ other repo — crush-config owns the reference; not filed
   stroke-dash; note PNG-scale + geometry-metric workflow).
~~9. Stack-front nginx→Caddy migration decision (owner appetite; would~~ owner — g2 unanswered (stack-front migration; fail2ban/E2E/runbook ride along)
   retire the fail2ban nginx scanner dependency or need a caddy
   equivalent).
   P3 — carried from the morning report, still open:
~~10. Data-model review honesty line (read vcard/blob/session-sqlite/~~ record stands — the honesty line stays on the immutable HTML (LEAVE precedent); the deltas shipped; this sweep annotated the review series
    store-row-shapes/Deps or soften "every type file was read") — note
    the deltas themselves (Receipt.Resolution, config typing) are NOW
    IMPLEMENTED by the post-v2.8.0 train; the review-of-record should
    get an ANNOTATE pass, not a rewrite.
~~11. Embed/link D2 SVGs in the HTML reports.~~ not adopted — below the bar
~~12. Devil's-advocate score re-derivation + since-09-19 ripple window.~~ not adopted — below the bar
~~13. docs-health VERIFY pass over both review reports against the~~ done — this v6 sweep (VERIFY + ANNOTATE over the review series, 2026-10-03)
    post-v2.8.0 code (several findings are now RESOLVED — annotate).

## g) QUESTIONS (cannot be answered from the repo)

~~1. **Diagrams**: do the rebuilt current + improved SVGs actually hit~~ owner — unanswered; the SVGs stand geometry-verified
   your bar ("YES! That's a diagram!")? I can verify geometry and
   content but cannot SEE them — if anything reads as cluttered,
   misaligned, or ugly, name the spot and I will rework it
   specifically.
~~2. **Stack front**: should the STACK's own vhost also migrate~~ owner — unanswered; stack-repo call
   nginx→Caddy (your SystemNix standard), or does Caddy stay
   module-side only? This is a stack-repo architecture call (fail2ban
   nginx scanner + E2E + runbook ride along).
~~3. **HSTS posture**: with Caddy's automatic HTTPS the friction of~~ owner — unanswered; the opt-in posture stands (AGENTS caddy.hsts)
   `caddy.hsts.enable` is near zero — do you want it DEFAULT-ON for the
   module (breaking-ish posture change) or keep the opt-in nginx
   inherited?

— Session artifacts: module + check rewrite, 6 doc files swept, 2
rebuilt diagram pairs, 1 smoke docstring, 1 TODO train note. Shipped in
v2.8.0 via the concurrent relock train. All gates green at close.
