# Status — the 20-skill review mega-train: full honest accounting

**Snapshot-rot header: claims verified as of 2026-11-19 14:45 CET.** Push/CI
statements rot fastest: at this minute, origin/main = `c89f6687` CI **RED**
(docs-gates: the daemon pushed a snapshot that predated my §f/harvest fixes),
while local HEAD `978576c1` carries the **green** tree (full battery PASS at
15:3x local rerun… corrected: PASS verified minutes before this header; the
run log of my final `docs-gates --check` is in the session transcript) and is
**unpushed**. A concurrent session (the "ring speaks names" island train +
the 13:37 pbx-lens report) worked beside this train the whole day.

## What did you forget? What could you have done better? What could you still improve?

- **Forgot (during the run):** that "local green" is not "remote green" when
  a daemon races you — I verified my tree, pushed, and the daemon immediately
  shipped ANOTHER commit (`c89f6687`) cut from the pre-fix snapshot, re-redding
  CI within minutes of my green. I should have pushed the FINAL tree state
  (post-§f, post-harvest) or waited for the daemon to catch up before
  declaring victory. Also forgot, until the brutal review forced the
  accounting: the release-readiness report itself (written late, below), and
  the fcr-findings TODO harvest (done, but only after self-flagging).
- **Could have done better:** dispatched 15 subagents in parallel without
  probing the provider cap → 11 agent deaths, hours of serial resumption, and
  finally a hard 5-hour usage limit that forced me to execute the last four
  skills myself. The correct probe was 2 concurrent for five minutes. The
  usage-limit warning signs were visible an hour before it hit.
- **Could still improve:** pre-draw the depth-pin split sketches (server
  5212/5250 = 38 lines of headroom) before a red budget forces one mid-train;
  the importContacts toast-half lacks a filename-specific pin; the favicon
  green literal is a split brain I created; the cadence drain is still a
  manual ceremony that can red CI again any day.

## a) FULLY DONE (verified this session)

1. **All 20 skills executed to their procedures** (improve-codebase-architecture
   is architecture-deepening's local successor — one skill, both names).
2. **15+ report artifacts**: architecture review #5 (4.36/5), deepening brief +
   grilled ThreadGate design, 6 D2/SVG architecture diagrams (+INDEX rows),
   full-code-review ×4 parts (~350 files), data-model review #6, naming review
   (12 findings), loginpage deep-dive (92/100), brutal-self-review #3 (HTML),
   samber/do audit, nix review, golang-policy report, code-quality+dedup
   sweep, dependency currency sweep, modularize verification #6,
   release-readiness memo (this train's §f home).
3. **8 code fixes, each test-verified**: SSE shutdown log honesty,
   `/metrics` nil-panic, `/login` CSP reporting tail, contacts-import
   double-escape, UTF-8 transcript cap (+new pin test), disposition
   echo-guard (+6-spec Ginkgo suite — the repo's 2nd Ginkgo lane), dial-clock
   seam, Start-failure Shutdown cascade (samber/do DO-2). Plus the green
   login favicon (brand fix) and 2 nix package `meta` completions.
4. **The 7.5h pre-existing CI red healed at the root**: cadence ARCHIVE-OVERDUE
   drained (14 reports, gates-first: check-rows --strict + bare-rows BEFORE
   every git mv), 11 broken relative links repaired (archived + planning
   corpora), TODO/ROADMAP pointers repointed, lychee green.
5. **Push under the ratified manual-push bar** (anchor: oldest unpushed commit
   2026-11-19T07:44:29+01:00, 365 min ≫ 60-min bar; red main + verified fix):
   `25301790..d0a9a3be` → **CI GREEN** (run 44529291480, verified).
6. **Harvests**: the concurrent session's 3 unharvested §f items + 7
   review-finding rows landed in TODO_LIST (all pre-flighted SAFE).
7. **Gates at close**: go test 37 pkgs green ×2, island 369/369, tsc green,
   format-check green, dprint green, docs-gates --check FULL matrix PASS,
   changelog-guard PASS, release-hygiene lychee ok.

## b) PARTIALLY DONE

- **go-release**: readiness analysis complete (v2.15.0 minor recommended; 422
  commits/176 human since v2.14.0; blockers + owner sequence documented) —
  the tag/push/release.sh ritual deliberately NOT run (owner hands).
- **go-ecosystem-upgrade**: upstream half complete (10 consumed-surface
  behinds, all additive; 4-train plan spelled out; zero fired reopen
  triggers) — the bumps themselves deliberately not executed (owner's
  tidy→vendorHash→push→CI-watch ritual cannot ride this session).
- **Review-finding fix-forward**: the surgical oranges/small fixed; the
  medium findings (fax M1, catalog M2) harvested as TODO rows, not fixed.
- **Remote greenness**: green at `d0a9a3be`, red again at `c89f6687` (the
  daemon's race — see §d); the healing tree sits unpushed in `978576c1`.

## c) NOT STARTED (deliberately deferred, all owned)

The fax Resend×retention fix (M1), the catalog ReviewSaved row + derived
parity (M2), the naming adjudication trains (FindThread first), the
asr.Locality/ASPRovider/LoginPageExtra type tails, the apidocs comment lie,
the favicon token unification, the cadence pre-drain automation, the 4
dependency bump trains, the v2.15.0 release itself, the messaging
read-surface collapse (deepening Strong pick — 5 owner questions attached),
the httpclient roundtrip skeleton, the calls.js surface split.

## d) TOTALLY FUCKED UP

- **The 15-agent burst**: walked 15 concurrent subagents straight into the
  provider's request cap. Eleven died mid-flight (some twice), ~2h of
  serialization followed, and the late 5-hour usage limit killed the last
  two (library-deep-dive, go-release) which I then redid in-session. Cost:
  roughly a third of the train's wall-clock and several dollars of dead
  agent starts.
- **The daemon race re-redded CI**: my green push was followed within
  minutes by the daemon shipping the pre-fix snapshot (`c89f6687`) — I
  treated "pushed and green" as the end state while my later §f/harvest
  fixes were still local. The red is mine to own even though the daemon
  committed it: I knew it was racing me.
- **Small, admitted**: the importContacts fix's toast half is untested for
  the filename case; I created the #17a467 favicon split brain; my first
  archive batch nearly shipped without the release-report §f (caught by the
  gate, fixed same minute).

## e) WHAT WE SHOULD IMPROVE

1. **Cadence auto-drain** — the 65-pin vs report-birthrate class reds CI for
   hours and only humans drain it (TODO row landed).
2. **Subagent discipline** — hard cap 2 concurrent on this account until the
   limiter is understood; probe before fan-out.
3. **Daemon push gate** — the daemon ships mid-edit/pre-fix trees; a cheap
   local-gate check before its push leg would have prevented both of this
   week's red windows.
4. **M1 first** — the one finding with a data-loss shape should not sit
   behind lower-value trains.

## f) Up to 50 things we should get done next

1. Push `978576c1` (the green tree) — red main + verified fix in hand; the
   manual-push bar's first prong is met NOW (see question 1).
2. Fax Resend×retention document-integrity fix + shared-path sweep test
   (fcr-02 M1, TODO row landed).
3. v2.15.0 release decision + ritual (readiness memo is the recipe).
4. T1 bump train: cqrs-lite system/metaengine/sqliteengine (fix-class
   crash-loop repair rides it).
5. T2 templ v0.3.1118 + CLI-matched regen train.
6. T3 modernc sqlite v1.62.0 train (backup-drill + vulnix gates).
7. T4 small-SDKs train (httputil, paperless, idempotency WithClock — retire
   TimingScale 5 and do NOT post the draft filing, nanoid, oauth2).
8. Catalog ReviewSaved row + exhaustive TestCatalogSchemaParity (M2).
9. Naming train 1: FindThread → ResolveThreadID (gate iface + 2 impls +
   messaging call sites).
10. Naming train 2: views.VoicemailItem, authmap.Map → Ownership, Shell ×2.
11. cqrs Mast* → Last* rename + replay-test regen.
12. asr.Locality constants (F1) + LocalityOf direct tests (fcr-03 L5).
13. ASRProvider enum (F4 — trigger already fired).
14. LoginPageExtra positional bools → question-form params (F5).
15. apidocs.go + AGENTS response-header comment lie fix.
16. Favicon #17a467 token unification (kill the split brain).
17. Cadence pre-drain automation (TODO row landed).
18. Depth-pin split sketches: server, server/api, app (arch review P2).
19. templ sources under the depth ratchet (arch review P3, one-commit
    registry act).
20. Server-root test-package drain riding future carves (arch review P3).
21. Messaging read-surface collapse onto ThreadGate (deepening pick — after
    the owner answers its 5 design questions).
22. Disposition echo-guard: add the server-side changed-verdict pre-read
    (Notes.Set shape) so SetDisposition can skip the dispatch entirely.
23. importContacts toast-half filename pin (close the admitted test gap).
24. push.Fan panic-containment test pin (fcr-03 finding 7).
25. FCM token-mint single-flight (fcr-03 finding 4 — sso's shape).
26. Signal data-URI filename escaping (fcr-03 finding 2).
27. vcard group-prefix support (`item1.TEL`, Apple shape) + fixture (L5).
28. Gateway webhook bytes.Buffer swap (~50-70 MiB peak, L6).
29. pbx do() UpstreamURL delegation (L7).
30. parseID brand verification (L1 — "Fax:<21 chars>" parses as ThreadID).
31. Dead zero-predicate sweep (L2: RelPath.IsZero, Address/Phone/Token.IsZero).
32. Fax inbound byte cap at the service seam (L3).
33. Blob brand-before-write (L4).
34. F3 twin: transcript 4000 bytes-vs-runes unification (fcr-01 finding 7).
35. Call-ID ≤128 bound: one home (4 sites, fcr-01 finding 6).
36. Root/open dual-mux parity pin (fcr-01 finding 8).
37. messages.templ hardcoded back-link → threads.back key (fcr-01 finding 9).
38. Reviews() full-queue fetch at boot probe — cache or slim (fcr-01 finding 10).
39. MarkRead error logging (fcr-01 residue).
40. Admin per-line CDR fan parallelism (fcr-01 residue).
41. cqrs boot-helper dedup (9 copies of DeploymentConfig boot, fcr-04 L7).
42. attachAttachments full-scan → per-thread filter (fcr-04 L5).
43. Seq-as-retention-clock comment (fcr-04 L4).
44. humanScaleVolume dual-home merge (fcr-04 split-brain list).
45. AGENTS stale citation: readmodels/admin_reads_test.go → cqrs pin (fcr-04).
46. arch services-never-import-server hardcoded-package gap (fcr-04 L1).
47. matrix ingest-before-Save dedup-by-event-id (fcr-04 info 8).
48. ntfy 404⇒prune semantics vs FCM runbook note (fcr-03 watch).
49. config family_test scrubEnv discipline (fcr-03 L6).
50. Owner ticks the pbx-lens 5-question register (harvested row — owner
    prompt, not agent work).

## g) Questions I cannot answer myself

1. **Push now?** `978576c1` is the green tree and main is red — the
   manual-push bar's first prong (red main + verified fix in hand) is met,
   but you said WAIT, so it waits: do you want me to push it, or let the
   daemon flush it (it may race again and ship it mid-edit of whatever comes
   next)?
2. **v2.15.0 order:** ship the release on current pins now, or run the
   4-train dependency plan first and fold those changelog lines in? (The
   readiness memo holds both recipes; the cqrs-lite trio carries a
   fix-class worker crash-loop repair, which argues for bumps-first.)
3. **The deepening pick:** execute the messaging read-surface collapse onto
   ThreadGate as the next structural train (my recommendation — it finishes
   the 11-11 Strong pick), or answer its five design questions (Q1–Q5 in
   the design doc) and redirect first?
