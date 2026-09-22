# Status Report: Call-Recording Support Investigation

**Date:** 2026-09-18 16:37
**Scope:** This session only — answering "do we support call recording?"
at both levels of the product: this repo (`webphone`, the island/UI + Go
server) and the consuming stack (`nix-international-telephony`, FreeSWITCH
PBX + nginx). Read-only research; **zero code changes were made**.

> Format note: the status-report skill's canonical format is HTML; the user
> explicitly requested Markdown — honored (same as the 15:25 v2-rebuild
> report).

---

## a) FULLY DONE ✅

Each item: what was done, evidence (verified this session by direct
grep/view at the cited locations), scope.

1. **Webphone-level recording audit** — established that the island and Go
   server contain **no audio-recording capability** whatsoever.
   Evidence: `MediaRecorder` / `createMediaStreamSource` /
   `startRecording` / `stopRecording` / `getUserMedia`-recording patterns
   have **zero matches** in all JS (`src/` era and
   `internal/web/assets/island/app/` alike). Audio flows only to output.
2. **All "record" hits correctly classified as CDR call history, not
   audio** — `recordHistory` (island, `internal/web/assets/island/app/panels.js:31`),
   "call detail record" rows (`internal/pbx/client.go:73`), "No calls
   recorded yet" panel copy (`internal/web/views/history.templ:29`).
   Scope: exact call-history (metadata) vs call-recording (audio)
   distinction now established for the repo.
3. **Webphone docs cross-check** — README capability table and the full
   `FEATURES.md` inventory (all 8 sections incl. PLANNED/WORTH_CONSIDERING)
   contain **no recording entry, not even as an idea**. Evidence: single
   `record` match in all `*.md` = README:28 (CDR row).
4. **Upstream-level audit: server-side recording CONFIRMED** —
   `nix-international-telephony` records calls at the PBX:
   `record_session` with `RECORD_STEREO=true` +
   `media_bug_answer_req=true`, writing `uuid_destination.wav` to
   `/var/lib/telephony/recordings` (`modules/freeswitch.nix:117-121`).
5. **Upstream option wiring verified** — `recording.enable` (default
   **true**, `modules/telephony/options.nix:577-580`), retention option
   (null = keep forever, timers prune), `recording.serve.enable` gating
   nginx `/recordings/` autoindex behind basic auth with runtime-rendered
   htpasswd (`modules/telephony/web.nix:184-191`,
   `modules/telephony/shared.nix:76-85`).
6. **Per-call opt-out + simulator parity verified** — `*97<ext>` dial
   skips recording (upstream FEATURES.md:45, VM-proven per
   `tests/pbx.nix`); the operator dialplan simulator treats
   `record_session` as a supported no-op action
   (`packages/telephony-operator/dialplan_sim.py:188`).
7. **Definitive two-level answer delivered** — PBX records every dialled
   call to stereo WAV, browsable at `https://<domain>/recordings/` behind
   operator basic auth; the webphone island has **no in-UI recording
   controls, indicator, or playback**. Upstream `FEATURES.md:46` marks
   recording 🟢 FULLY_FUNCTIONAL (VM-proven).

## b) PARTIALLY DONE 🟡

1. ~~**Cross-repo documentation linkage.**
   What works: the recording story is fully documented _upstream_
   (README:61, FEATURES.md:46, DOMAIN_LANGUAGE.md:33).
   What remains: webphone's own README/FEATURES/AGENTS.md never mention
   that recordings exist at the stack level or where they live — a fresh
   webphone session rediscovers nothing.~~ — done 2026-09-19: README capability row + FEATURES WORTH_CONSIDERING row + AGENTS.md hard-won bullet (two-level recording story).
2. ~~**Verification of the "VM-proven" claim.**
   What works: upstream FEATURES.md asserts `tests/pbx.nix` proves a
   dialled call grows `*_1001.wav` (and that `recording.enable = false`
   provisions no directory).
   What remains: I cited the claim but did not open `tests/pbx.nix` to
   confirm the assertions exist (verify-external-claims discipline applied
   only partially — the source repo was in reach).~~ — verified 2026-09-19: `tests/pbx.nix` in the stack repo carries the recording assertions (recorded calls land in the served dir; recording disabled → no files).
3. ~~**Session knowledge capture.**
   What works: findings delivered in-chat with exact file:line refs.
   What remains: per the memory-maintenance protocol, the durable
   two-level recording fact belongs in webphone `AGENTS.md`
   (Hard-won knowledge); not written.~~ — done 2026-09-19 (AGENTS.md hard-won knowledge).

## c) NOT STARTED ⚪

(Nothing in this session's mandate; listing what the investigation
revealed as unwritten-but-implied work.)

1. **Recording surface in the webphone UI** — list/play/delete recordings
   inside the island (mirroring the voicemail pattern). → ROADMAP raw ideas (product intent pending — g1).
2. **Live "REC" indicator during calls** — → ROADMAP raw ideas (tied to 1; consent dimension — g2).
3. **Per-extension recording access** — → ROADMAP raw ideas (design decision — g3).
4. ~~**Harvest of this report into TODO_LIST/ROADMAP**~~ done (2026-09-19 docs-health sweep: recording cluster routed to ROADMAP; docs items done).

## d) TOTALLY FUCKED UP 💥

**Radical honesty — one genuine miss, nothing destroyed:**

1. **The first answer was scoped to the repo, not the product.**
   "Do we support call recording?" got the answer **"No"** — correct for
   _this repository_, wrong for _the product the user runs_, because the
   consuming stack records server-side. The user had to spend a follow-up
   prompt ("On the nix-international-telephony level?") to get the truth.
   - Severity: answer-quality failure; cost one round trip; no data or
     code damage (session made zero mutations).
   - Root cause: AGENTS.md _literally documents_ the ownership split
     ("the module there owns config.js rendering…; this repo owns ONLY
     the UI") — the pointer that the product extends beyond this repo was
     in context and I did not apply it to scope the question.
   - Mitigation going forward: product-level questions ("do we support X")
     get answered at product scope: check the consuming stack whenever
     AGENTS.md describes a cross-repo contract. Recorded in (e).

**That is the complete list.** No broken builds, no failed tests (none
were runnable/relevant — read-only session), no lost work.

## e) WHAT WE SHOULD IMPROVE 🔧

1. **Scope answers to the product, not the checkout.** Pattern: questions
   phrased as "do we…" in a repo that is a _component_ of a documented
   larger stack must start from the stack. Fix: a session habit (and a
   line in webphone AGENTS.md): before answering a capability question,
   enumerate the levels (island / Go server / consuming PBX stack) and
   answer per level.
2. **Verify cited test files, not just cited claims.** I echoed
   "VM-proven in tests/pbx.nix" from upstream FEATURES.md without opening
   the file. Fix: one extra `view`/`glob` when a claim names a specific
   test — verify-before-echoing, the inbound cousin of
   verify-before-filing.
3. **Write discovered durable facts immediately.** The two-level
   recording story took a whole session to assemble and exists only in
   chat history. Fix: update webphone AGENTS.md "Hard-won knowledge" the
   moment a cross-repo fact is confirmed (item f.6).
4. **Surface integration gaps as findings, not footnotes.** "Recordings
   exist upstream but have no UI surface" is arguably the most valuable
   output of the session and appeared only in the final summary line.
   Fix: lead with gaps discovered during capability audits, and route
   them into (f).
5. **Docs drift between siblings is structural here.** webphone README's
   capability table is written as if this repo were the whole product.
   Fix: add one README row noting PBX-side features that consume the UI
   (recordings today; whatever comes next), or a "stack context" note.

## f) Up to 50 things we should get done next

Brainstorm per the skill's rule: a large N is ROADMAP fuel, not a
commitment list. Grouped; each with Impact / Effort / Category. Items 1-10
come directly from this session's discovery; 11-20 from the in-flight v2
state observed at session start; 21-50 from webphone's own
PLANNED/WORTH_CONSIDERING inventory (read this session).

**Recording integration (this session's core discovery):**

| #  | Task                                                                                                                                               | Impact | Effort | Category      |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1  | ~~Add a Recordings panel to the webphone UI (list WAVs, mirror the voicemail pattern)~~ → ROADMAP (recording cluster)                              | High   | M      | Feature       |
| 2  | ~~In-island playback (same-origin audio element; strict CSP already allows it)~~ → ROADMAP                                                         | High   | M      | Feature       |
| 3  | ~~Live "REC" indicator on the call card while the PBX records that leg~~ → ROADMAP                                                                 | High   | M      | Feature       |
| 4  | ~~Dialpad shortcut for `*97` no-record dial (per-call opt-out exists upstream, invisible in UI)~~ → ROADMAP                                        | Medium | S      | Feature       |
| 5  | ~~README: document stack-level recording + `/recordings/` location~~ done 2026-09-19 (capability row)                                              | Medium | S      | Documentation |
| 6  | ~~webphone AGENTS.md: record the two-level recording story as hard-won knowledge~~ done 2026-09-19                                                 | Medium | S      | Documentation |
| 7  | ~~FEATURES.md: add cross-reference row (recordings exist upstream; webphone has no UI)~~ done 2026-09-19                                           | Medium | S      | Documentation |
| 8  | ~~Per-extension recording access via the Go server (phone-api proxy pattern) replacing shared basic auth~~ → ROADMAP (open question: access model) | High   | L      | Feature       |
| 9  | ~~Surface retention in UI (recordings are pruned by timers upstream; users should know)~~ → ROADMAP                                                | Low    | M      | Feature       |
| 10 | ~~README privacy note: `RECORD_STEREO` captures both legs; consent/jurisdiction caveat~~ → ROADMAP (consent posture — g2)                          | Medium | S      | Documentation |

**In-flight v2 state observed at session start (uncommitted changes):**

| #  | Task                                                                                                                                                    | Impact | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 11 | ~~Review/verify the uncommitted v2 changes~~ done (v2.0.0 shipped; gates green through 07:43)                                                           | High   | M      | Quality       |
| 12 | ~~Re-run the upstream browser E2E after the v2 switchover~~ done at `00f13fe` (green 06:42)                                                             | High   | L      | Quality       |
| 13 | ~~German translations for the server-rendered tabs~~ done (18:50 #14)                                                                                   | Medium | M      | Feature       |
| 14 | ~~Bring the 15:25 v2-rebuild report current with the working tree~~ done (2026-09-19 docs-health annotate sweep)                                        | Medium | S      | Documentation |
| 15 | ~~Run `nix flake check` + `nix fmt` gate before handing the tree back~~ done (green through 06:42/07:43)                                                | High   | S      | Quality       |
| 16 | ~~Review the `.gitignore` modification sitting in the tree~~ done (swept into reviewed commits; tree clean since)                                       | Low    | S      | Cleanup       |
| 17 | ~~Review the `flake.lock` bump~~ done (nixpkgs → 2026-09-17, 06:42 #7)                                                                                  | Low    | S      | Cleanup       |
| 18 | ~~Review `update.sh` changes~~ done (shfmt-fixed + shellcheck-clean, 16:21)                                                                             | Low    | S      | Cleanup       |
| 19 | ~~Replace "heuristic" auto-commit history with per-task commits~~ **Won't implement —** daemon-owned history accepted; explicit commits when authorized | Low    | S      | Cleanup       |
| 20 | ~~Re-count the 35 asserted island DOM ids after v2 changes~~ done (16:21 fixed 34→35; contract test pins them)                                          | Medium | S      | Quality       |

**From webphone FEATURES.md PLANNED / WORTH_CONSIDERING (already inventoried):**

| #  | Task                                                                                              | Impact | Effort | Category |
| -- | ------------------------------------------------------------------------------------------------- | ------ | ------ | -------- |
| 21 | ~~Login rate limiting for `/api/session`~~ done (18:50 #4)                                        | High   | M      | Security |
| 22 | ~~Message delivery-receipt webhook~~ done (16:34 session)                                         | Medium | M      | Feature  |
| 23 | ~~Session persistence across restarts~~ decided: in-memory by design (FEATURES WORTH_CONSIDERING) | Medium | L      | Feature  |
| 24 | ~~Message pagination/virtualization (200-per-thread window)~~ done (18:50 #13)                    | Low    | M      | Quality  |
| 25 | ~~vCard contact import/export~~ done (18:50 #12)                                                  | Low    | M      | Feature  |
| 26 | ~~History search/filter~~ done (18:50 #10)                                                        | Medium | M      | Feature  |
| 27 | Retention/cleanup job (blobs grow unbounded) — FEATURES WORTH_CONSIDERING                         | Medium | M      | Quality  |
| 28 | ~~Richer `/healthz` (store, gateway mode)~~ done (store side: `cqrshtmx.ReadinessHandler`)        | Low    | S      | Quality  |
| 29 | ~~Manual theme override toggle~~ done (18:50 #9)                                                  | Low    | S      | Feature  |
| 30 | ~~Keyboard shortcuts / media keys~~ done (18:50 #11)                                              | Low    | M      | Feature  |
| 31 | Video calls (sip.js supports; UI needs a video surface) — FEATURES WORTH_CONSIDERING              | Low    | L      | Feature  |
| 32 | PWA offline shell (service worker must respect strict CSP) — FEATURES WORTH_CONSIDERING           | Low    | L      | Feature  |
| 33 | ~~sip.js 0.22 evaluation~~ done (evaluation report — stay on 0.21.2)                              | Medium | M      | Quality  |
| 34 | Voicemail transcripts (only if the PBX API ever provides them) — FEATURES WORTH_CONSIDERING       | Low    | S      | Feature  |
| 35 | ~~Fax richer page-count parsing from status payloads~~ done (18:50 #7)                            | Low    | S      | Feature  |

**Follow-on ideas grounded in this session's observations:**

| #  | Task                                                                                                                        | Impact | Effort | Category      |
| -- | --------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 36 | ~~Serve recordings through the existing phone-api reverse proxy (auth injected server-side)~~ → ROADMAP (recording cluster) | Medium | M      | Feature       |
| 37 | ~~Document the recording filename convention (`uuid_destination.wav`) in webphone docs~~ → ROADMAP                          | Low    | S      | Documentation |
| 38 | ~~Open `tests/pbx.nix` and confirm the recording assertions first-hand (close b.2)~~ done 2026-09-19 (assertions verified)  | Low    | S      | Quality       |
| 39 | ~~Add a "recorded" badge to CDR history rows when a WAV exists for that call~~ → ROADMAP (recording cluster)                | Medium | M      | Feature       |
| 40 | ~~Recordings browser inside the upstream `/operator/` window (today: nginx autoindex only)~~ → ROADMAP (upstream)           | Low    | M      | Feature       |
| 41 | ~~Upstream `natAddress`: real-NAT validation~~ → upstream repo                                                              | Medium | L      | Quality       |
| 42 | ~~Run docs-health HARVEST on this report so (f) does not die in a timestamped file~~ done (2026-09-19 sweep)                | High   | S      | Documentation |
| 43 | ~~Compress/prune stereo WAVs (Opus transcode job)~~ → ROADMAP (recording cluster)                                           | Low    | M      | Quality       |
| 44 | ~~Document `*97` in the island's user-facing help~~ → ROADMAP (recording cluster)                                           | Low    | S      | Documentation |
| 45 | ~~New recording UI strings ship en/de from day one~~ → ROADMAP (applies when the recording UI is built)                     | Low    | S      | Quality       |
| 46 | ~~Post-call SSE nudge: "recording available" when the WAV lands~~ → ROADMAP (recording cluster)                             | Low    | M      | Feature       |
| 47 | ~~Live start/stop recording control from the island (needs event-socket path — bigger design)~~ → ROADMAP                   | Low    | L      | Feature       |
| 48 | ~~Cross-link webphone README ↔ upstream recording docs section~~ done (README capability row references the stack)          | Low    | S      | Documentation |
| 49 | ~~Consider a stack-level capability matrix (island vs server vs PBX)~~ → ROADMAP                                            | Medium | S      | Documentation |
| 50 | ~~Re-verify this session's two answers after the v2 switchover lands~~ done (E2E green post-switchover; contract pinned)    | Low    | S      | Quality       |

Harvest note: items 1-10, 21-28, 33, 38, 42, 49 are the strongest
TODO_LIST candidates; the rest are ROADMAP fuel pending the answers to (g).

## g) Questions I cannot figure out myself

1. ~~**Do you want recordings surfaced inside the webphone UI at all?**~~ → ROADMAP open questions (product intent).
2. ~~**What is the legal/consent posture for recording?**~~ → ROADMAP open questions (consent/jurisdiction).
3. ~~**Should recording access stay one shared operator credential, or become per-extension?**~~ → ROADMAP open questions (access model).

---

_Point-in-time snapshot. Section (f) is the HARVEST input for
`TODO_LIST.md` / `ROADMAP.md` (item 42)._
