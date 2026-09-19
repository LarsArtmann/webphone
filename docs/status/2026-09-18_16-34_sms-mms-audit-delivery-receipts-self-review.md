# SMS/MMS Audit, Delivery Receipts, and Brutal Self-Review

Point-in-time snapshot: 2026-09-18 16:34 CEST. Scope: this session only
(the "How are we doing on SMS/MMS?" question and the work it produced).
No fresh research beyond what was read and verified in-session.

## What this session did

1. Audited the whole messaging subsystem against its docs (FEATURES,
   TODO_LIST, CHANGELOG, README claims vs. code), reading domain model,
   store, service, gateway, webhook handlers, actions, views, and tests.
2. Found exactly one gap: delivery receipts. `StatusDelivered` and the
   UI badge existed, but nothing ever flipped them.
3. Closed it: `Messages.MessageByProviderRef` (store),
   `messaging.Service.DeliveryReceipt` (service), `POST
   /hooks/message/status` (handler, route, contract comment), tests at
   both store and server level, docs synced (FEATURES, TODO_LIST,
   CHANGELOG Unreleased, README hooks contract).
4. Verified: `GOEXPERIMENT=jsonv2 go test ./...` green (baseline first,
   then after changes), `go vet` clean, `nix flake check` all checks
   passed, `nix fmt` clean.

## a) FULLY DONE

- SMS/MMS subsystem audit: every FEATURES messaging claim verified
  against code and tests (limits ≤1600 chars / ≤5 attachments / ≤10 MiB,
  fail-closed Bearer gate with 503, owner-scoped attachment streaming,
  unread upsert, swap-safe SSE fragments, 200-message window).
- Delivery receipts end-to-end: send → `provider_ref` receipt → provider
  callback → status flip → live SSE badge update. Store, service,
  handler, route, contract comment.
- Tests: `TestMessageStatusWebhookUpdatesTranscript` (202 → "delivered"
  rendered in transcript; unknown ref → 404; empty ref, `sent`, `queued`
  statuses → 400) and store round-trip (by-ref lookup, inbound-ref
  refusal, unknown-ref refusal, delivered persistence).
- Docs honesty restored: FEATURES messaging row now 🟢, TODO_LIST row
  deleted, CHANGELOG `[Unreleased]` entry, README hooks contract lists
  the new endpoint.
- Hygiene: fixed a pre-existing errcheck warning (`resp.Body.Close()` in
  the fax test), stale-LSP warning disproven with `go vet`.

## b) PARTIALLY DONE

- ~~FEATURES wording on delivery receipts: I marked it FULLY_FUNCTIONAL,
  but loopback-mode messages can never reach `delivered` (loopback marks
  `sent` and stops). True only for webhook providers that implement the
  callback. The row needs a caveat, or loopback needs a decision (see g2).~~ — caveat added to FEATURES 2026-09-19; loopback semantics → ROADMAP open questions (g2).
- ~~README "Bridging FreeSWITCH" example still says "or a message-status
  hook with that id" — I updated the hooks block but left this sentence
  vague instead of naming `/hooks/message/status`.~~ — fixed 2026-09-19: the example now names `/hooks/fax/status` / `/hooks/message/status`.
- Test coverage of the receipt path: the SSE `thread` event fires via
  `notify()`, but the new test asserts the badge only via HTTP re-fetch
  of the partial, not via the SSE event itself. → ROADMAP (testing long tail)

## c) NOT STARTED (observed this session, untouched by it)

- All pre-existing TODO_LIST rows: deployment ownership decision,
  upstream switchover + browser E2E re-run, German tab translations,
  login rate limiting, session persistence, message pagination beyond
  200, fax page-count parsing, vCard import/export, history search,
  keyboard shortcuts, manual theme toggle, sip.js 0.22 evaluation,
  `countUnread` caching.
- Session-specific: SSE-event assertion for delivery receipts, `-race`
  run over the new webhook path, idempotent repeat-callback test
  (provider retry → same status → still 202, unasserted).
- golangci-lint via buildflow was not run this session (fast gates only:
  go vet, nix flake check).

## d) TOTALLY FUCKED UP

Nothing. No broken state left behind; suite and flake check green at
close. Two honest dents, neither fatal:

- One sloppy first-pass compile error in the new test (`c.do` result
  passed straight into `FindSubmatch`), caught by diagnostics within
  seconds and fixed before any run.
- Git history for this feature is mangled by the auto-commit daemon: the
  work is scattered across four "chore: auto-commit N file(s)
  (heuristic)" commits with zero narrative. Not my commit to make
  (harness forbids), but the history does not tell the story.

## e) WHAT WE SHOULD IMPROVE

1. ~~**`provider_ref` is load-bearing but unindexed-unique.**~~ — half done: `fax_jobs` got its UNIQUE partial index (`internal/store/db.go:85`); the `messages` side is TODO_LIST.
2. ~~**Sibling-hook asymmetry (small split brains):** empty `provider_ref`
   → messages 400, faxes 404. Failure detail → faxes persist error text,
   messages only log it. Deliberate hardening on one side, drift on the
   other; pick one contract per pair and align.~~ → TODO_LIST (`hookFaxStatus` row covers both the mapping split and empty-ref alignment); message failure-detail persistence → ROADMAP.
3. ~~**Delivered vs. sent look identical** (`wp-status-sent` class serves
   both).~~ → TODO_LIST (distinct delivered style).
4. **Prove receipts with the real binary.** — still open as written (cheap loopback curl; superseded in value by the in-repo smoke-suite TODO row).
5. ~~**Test rigor:** run `-race` locally for the webhook paths~~ done (17:46 session: full suite green under `-race`); ~~assert the SSE event, not just the re-rendered partial~~ → ROADMAP.
6. ~~**FEATURES caveat** for delivery receipts under loopback mode (see g2).~~ done 2026-09-19.
7. **Commit narrative:** — daemon-owned history accepted; per-task commits when explicitly authorized (standing rule).

## f) UP TO 50 THINGS TO GET DONE NEXT

Sorted by impact; [N] = new from this session, [T] = already in
TODO_LIST, [F] = already in FEATURES planned table.

1. [T] ~~Deployment ownership decision + NixOS module (High, blocks ops story)~~ done (18:50 #5)
2. [T] ~~Upstream switchover in nix-international-telephony + browser E2E re-run (High)~~ done (E2E green after `00f13fe`)
3. [N] ~~Fix README bridge-example sentence to name `/hooks/message/status` (S)~~ done 2026-09-19
4. [N] ~~FEATURES caveat: delivery receipts require a callback-capable provider (S)~~ done 2026-09-19
5. [N] ~~UNIQUE partial index on `messages.provider_ref` / `fax_jobs.provider_ref` (S–M)~~ fax side done (`db.go:85`); messages side → TODO_LIST
6. [N] SSE `thread`-event assertion for the delivery receipt (S) → ROADMAP
7. [N] ~~Idempotency test: repeated delivery callbacks stay 202/no-op (S)~~ done at `232795e` (`hooksIdem` replay dedupe + tests)
8. [N] ~~Run messaging/server tests with `-race` once, fix fallout (S)~~ done (17:46 session, full suite)
9. [N] Distinct CSS for delivered vs. sent badge (S) → TODO_LIST
10. [N] Decision + implementation: loopback simulating `delivered` (see g2) (S) → ROADMAP open questions
11. [N] Align empty-`provider_ref` handling between message and fax hooks (S) → TODO_LIST (hookFaxStatus row)
12. [T] ~~Login rate limiting on `/api/session` and hooks (Medium, S)~~ done (18:50 #4)
13. [T] ~~German translations for server-rendered tabs (Medium, M)~~ done (18:50 #14)
14. [T] ~~`countUnread` cache (Low, S)~~ done (18:50 #6)
15. [N] Persist + display failure reason for failed messages (parity with fax; schema change) (M) → ROADMAP
16. [N] Loopback provider-ref format: use random IDs, not UnixNano (S) → ROADMAP
17. [T] ~~Session persistence across restarts (Low, M — security tradeoff, needs g3-style decision)~~ decided: in-memory by design (FEATURES WORTH_CONSIDERING)
18. [T] ~~Message pagination/virtualization beyond 200 (Low, M)~~ done (18:50 #13)
19. [T] ~~Fax page-count parsing from status payloads (Low, S)~~ done (18:50 #7)
20. [T] ~~vCard contact import/export (Low, M)~~ done (18:50 #12)
21. [T] ~~History search/filter (Low, M)~~ done (18:50 #10)
22. [T] ~~Keyboard shortcuts + media keys (Low, M)~~ done (18:50 #11)
23. [T] ~~Manual theme toggle (Low, S)~~ done (18:50 #9)
24. [T] ~~sip.js 0.22 evaluation via `./update.sh` (Low, M)~~ done (evaluation report — stay on 0.21.2)
25. [F] Retention/cleanup job for blobs and old rows (data grows unbounded) (M) — FEATURES WORTH_CONSIDERING
26. [F] ~~Richer `/healthz` (store ping, gateway mode) (S)~~ done (store side: `cqrshtmx.ReadinessHandler`)
27. [N] ~~Document the provider-callback retry expectation in README hooks section (S)~~ done 2026-09-19 (idempotent-replay note added)
28. [N] ~~Consider wiring the receipt hook into the README demo bridge sketch (S)~~ done 2026-09-19 (bridge example names `/hooks/message/status`)
29. [T] Video calls (WORTH_CONSIDERING; sip.js supports it) — FEATURES
30. [F] PWA offline shell (WORTH_CONSIDERING; must respect strict CSP) — FEATURES

Items 31–50 deliberately left empty: this session surfaced no further
specific, evidence-backed tasks, and padding a list with invented work
is exactly the kind of dishonesty this report refuses.

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Failure detail for failed messages:** faxes persist the provider's
   error text and show it; messages only log it. → ROADMAP (persist + display failure reasons).
2. **Loopback semantics:** should the loopback gateway simulate
   `delivered` (like fax loopback resolves `transmitted`) so demos and
   E2E exercise the receipt path, or is `sent` the honest terminal state
   for dev mode? → ROADMAP open questions.
3. **Commit narrative:** the daemon scattered this feature across four
   heuristic "chore" commits. → accepted tradeoff, documented in AGENTS.md (daemon commits AND pushes).
