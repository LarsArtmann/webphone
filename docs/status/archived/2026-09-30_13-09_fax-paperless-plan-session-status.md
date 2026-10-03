# Status — fax→Paperless plan session (2026-09-30 13:09 CEST)

**Scope:** THIS session only (started ~12:45, snapshot 13:09). Per the owner's
instruction the report is scoped to what this session did and noticed — not a
whole-project audit. **Format override:** status-report skill defaults to a
styled HTML dashboard; the owner explicitly requested `.md`, so this is flat
Markdown by instruction.

**Session arc:** owner floated "Optional integration for fax with
/home/lars/projects/go-paperless ??" → exploration (go-paperless repo +
webphone fax/config internals) → verdict + design in chat → plan doc →
TODO_LIST row. No code changed, no tests broken, nothing deployed.

> ARCHIVED 2026-10-03 (docs-health v6 sweep): the plan shipped in full as
> the T14 train (2026-10-02) — config both-or-neither, the fax.Archiver
> seam, the internal/paperless adapter with family + stub-test pins,
> the fire-and-forget Receive hook, Deps wiring, config-off zero-delta,
> and the full doc set; outbound archiving stays out-of-scope by design.
> Per-item verdicts inline.

**Concurrent-session notice (observed, untouched, per AGENTS rules):** the
auto-daemon commit `b1d494f` (12:55) carried my two files AND a
`package/nixos-module.nix` change (143 lines) I did not author;
`e68d71a` (13:03) carried dedup-registry + two OTHER sessions' status reports
(go-cqrs-lite-question, dedup-sweep-t2); `nix/module-check.nix` is modified in
the working tree right now, not by me. At least two sibling sessions were live
during this window.

---

## a) FULLY DONE

~~1. **Fax→go-paperless feasibility assessment** — read go-paperless (README,~~ done — this session (report of record)
layout, `UploadRequest`/`Upload` in client.go, go.mod) and webphone fax
internals (`fax.Service` Send/Receive/UpdateProviderStatus, `gateway`
FaxWebhook/FaxGateway, the CRM both-or-neither config validation +
`TestLoadValidatesCRMConfig`, both go.mod Go floors). Verdict delivered:
clean fit, CRM seam is the template. Evidence: conclusions encoded in the
plan doc; SDK facts verified against the LOCAL checkout (go 1.27, json/v2,
client-only scope statement, two named consumers).
~~2. **Plan doc written** —~~ done — this session (report of record)
`docs/planning/2026-09-30_12-53_fax-paperless-integration-plan.md`
(134 lines): config block, webphone-owned `fax.Archiver` seam,
`internal/paperless` adapter, fire-and-forget semantics with the
webhook-sync REJECTION argument (Receive has no provider-ref dedupe —
verified by reading `Receive`), non-goals, named follow-ups, gates,
sizing (~400–500 lines incl. tests). Committed by daemon in `b1d494f`.
~~3. **TODO_LIST.md row added and verified** — 🟡 `PLANNED` / Medium / S-M row~~ done — this session (report of record)
inserted into the Medium cluster (between `/version` enrichment and AGENTS
compaction), evidence column links the plan doc. Table structure verified
mechanically (5 cells, no raw pipes). Committed in `b1d494f`.

## b) PARTIALLY DONE

~~1. **The integration itself is DESIGNED, not built** — what exists: plan doc~~ done — the integration SHIPPED as the T14 train (2026-10-02; CHANGELOG Unreleased)

- TODO row. What remains: every implementation step (section c). Blocker:
  owner go + the 3 questions in (g). Effort: S-M (~400–500 lines). This is
  intentional staging, not drift — but it means the session's only durable
  artifact is a proposal.
  ~~2. **SDK claim verification is HALF-done** — verified locally: go.mod `go 1.27`,~~ resolved by events — the dep pulled + built; every gate since
  json/v2 requirement, `UploadRequest` shape (Title/Created/CorrespondentID/
  TagIDs/DocumentTypeID/CustomFields), `Ensure*` helpers, task polling +
  `TaskOutcome.Duplicate`. NOT verified: that a tagged release is actually
  pullable from the module proxy (`go list -m …@latest` never run; latest
  tag never checked). The plan's `go get` line is therefore an unverified
  claim in a committed doc. Effort to close: S.
  ~~3. **`fax.New` call-site enumeration** — the design grows the `fax.New`~~ superseded — the signature grew + landed (T14)
  signature (archiver param) but I never grepped its call sites; the "half a
  session" sizing rests on AGENTS' `Deps.CRM` precedent, not on counted
  consumers. Effort to close: S (one grep + wiring diff read).
  ~~4. **Naming due diligence** — coined `fax.Archiver` / `internal/paperless` /~~ not adopted — the names landed as coined; the seam lives in the AGENTS Paperless bullet, no glossary entry
  `webphone-fax-id` WITHOUT consulting `docs/DOMAIN_LANGUAGE.md` (the
  naming-is-architecture rule). Names may collide with existing domain
  vocabulary. Effort to close: S.

## c) NOT STARTED (all implementation; gated on owner go)

Everything below is planned in the doc, zero code written. Priority: waiting
on the (g) answers — not deprioritized.

- `Config.Paperless` block + env keys + both-or-neither/URL validation + test.
- `fax.Archiver` interface in `internal/fax`, nil-safe field, `fax.New` param.
- `internal/paperless` package: client construction, lazily-cached `Ensure*`
  metadata, `ArchiveFax` upload, error-family codes + `family_test.go` pin.
- Fire-and-forget hook in `Receive` (detached goroutine, own timeout, WARN).
- Server `Deps` wiring; nil-archiver zero-delta proof.
- vendorHash roundtrip in the same breath as the go.mod addition; full gates
  (`go test -count=1 ./...`, buildflow); config-off smoke zero-delta.
- Post-landing doc set: AGENTS seam bullet, FEATURES row, CHANGELOG.

## d) TOTALLY FUCKED UP

**Nothing this session.** Docs-only session: no code paths touched, no gate
skipped that applies (markdown is outside treefmt scope; markdown-lint is
detect-only corpus posture), nothing deployed, nothing reverted, no concurrent
session's file touched.

Radical-honesty near-miss (not (d) severity, named so it doesn't rot): the
committed plan doc contains the unverified publishability claim from (b)2 —
if `go-paperless` had an untagged/poisoned proxy state, the first implementer
would burn a session discovering it. It doesn't block anything until
implementation starts.

## e) WHAT WE SHOULD IMPROVE

~~1. **Verify-external-claims applies to PLAN DOCS too** — I gated the claim~~ process record — skill-trigger idea (verify-external-claims lives in the owner's skill repo)
mentally on "implementation will check", but a committed plan doc IS an
encoded claim. Fix: run `go list -m github.com/larsartmann/go-paperless@latest`

- check git tags BEFORE writing the `go get` line; pin the version in the
  doc. (Pattern → worth a line in the verify-external-claims skill trigger:
  "plan/TODO docs naming a dependency".)
  ~~2. **Consult `docs/DOMAIN_LANGUAGE.md` before coining seam names** — the~~ process record
  naming reflection exists in AGENTS; I skipped the domain-glossary lookup
  step for `Archiver`/`paperless`. Impact: a rename after landing costs a
  cross-file churn the glossary check costs nothing.
  ~~3. **Plan docs should inherit the CRM precedent wholesale** — the CRM seam~~ superseded — parity landed as the zero-delta gate + WARN observability; environmentFile rides the module
  set three conventions my plan omitted until asked: observability
  (`webphone_crm_lookups_total` → plan has no metrics story), secret-at-rest
  mechanism (`environmentFile` → plan never names it), and an explicit
  "config-off = zero-delta" gate line (I had it in chat, not in the doc's
  gate list). Fix: a short "CRM-parity checklist" when authoring any new
  optional-integration plan.
  ~~4. **Metadata spec gap found in self-review:** the v1 metadata carries the~~ not adopted — provenance carries the fax id only (paperless.go:90)
  fax ID as provenance but NOT the owning extension (`job.Owner`) — in a
  multi-DID deployment Paperless can't tell whose fax it was. Minor for the
  current single-user stack; a one-line plan amendment (add extension to the
  provenance custom field or a second field).
  ~~5. **Ask scope-shaping questions before authoring, or mark sections~~ process record
  decision-gated** — inbound-only v1 shapes half the doc; the owner hadn't
  confirmed. The doc does say "awaits owner go", but the inbound-only choice
  reads as decided, not proposed. Cheap fix: a "decisions assumed (owner may
  flip)" list at the top of future plan docs.

## f) UP TO 50 THINGS TO GET DONE NEXT

Ranked by impact within the session's scope. **Routing column:** where the
item already lives. Rows 25–44 were NOTICED this session (full TODO_LIST
read) and are already routed there — listed so HARVEST doesn't duplicate
them; the >25 extras are brainstorm-grade per the skill (ROADMAP fuel).

| #  | Task | Impact                                                                                                    | Effort | Category | Routing       |
| -- | ---- | --------------------------------------------------------------------------------------------------------- | ------ | -------- | ------------- |
| ~~ | 1    | Answer (g) Q1–Q3 (go/no-go, Paperless hosting, v1 scope) — unblocks everything else in 1–24               | High   | S        | Decision      |
| ~~ | 2    | Verify go-paperless publishability (`go list -m @latest` + git tags), pin the version in the plan doc     | High   | S        | Quality       |
| ~~ | 3    | Check `docs/DOMAIN_LANGUAGE.md` for `Archiver`/`paperless` collisions before implementation               | Medium | S        | Quality       |
| ~~ | 4    | Amend plan doc: CRM-parity items — smoke gate line, `environmentFile` token-at-rest, observability story  | Medium | S        | Documentation |
| ~~ | 5    | Amend plan doc: extension in the provenance metadata (owner-DID scoping in Paperless)                     | Medium | S        | Documentation |
| ~~ | 6    | Grep `fax.New(` call sites; confirm the signature-growth blast radius claimed by the sizing               | Medium | S        | Quality       |
| ~~ | 7    | `Config.Paperless` block + env keys + both-or-neither + URL validation                                    | High   | S-M      | Feature       |
| ~~ | 8    | `TestLoadValidatesPaperlessConfig` beside the CRM pin                                                     | High   | S        | Quality       |
| ~~ | 9    | `fax.Archiver` interface in `internal/fax` + nil-safe service field + `fax.New` param                     | High   | S        | Feature       |
| ~~ | 10   | Nil-archiver zero-delta test (every existing fax test passes untouched)                                   | High   | S        | Quality       |
| ~~ | 11   | `internal/paperless` package skeleton: client construction from config                                    | High   | S        | Feature       |
| ~~ | 12   | Lazily-cached `EnsureTag`/`EnsureDocumentType`/`EnsureCustomField` with stub-server test                  | High   | M        | Feature       |
| ~~ | 13   | `ArchiveFax` upload: title format, `Created` from job, provenance field, ≤20 MiB buffer                   | High   | M        | Feature       |
| ~~ | 14   | Fire-and-forget hook in `Receive` after persist+notify (detached goroutine, own timeout, WARN)            | High   | S        | Feature       |
| ~~ | 15   | `internal/paperless/family_test.go` error-family pin (codes at origin, per the 2026-09-30 convention)     | High   | S        | Quality       |
| ~~ | 16   | Adapter httptest stub: upload form fields + task-poll verdict + duplicate-refusal path                    | High   | M        | Quality       |
| ~~ | 17   | Server `Deps` wiring (nil when config absent)                                                             | High   | S        | Feature       |
| ~~ | 18   | vendorHash roundtrip same-breath as go.mod addition + `nix build` proof                                   | High   | S        | Feature       |
| ~~ | 19   | Full gates: `nix develop -c go test -count=1 ./...` + buildflow                                           | High   | S        | Quality       |
| ~~ | 20   | Config-off smoke zero-delta (fresh binary boots, fax lane unchanged)                                      | Medium | S        | Quality       |
| ~~ | 21   | Decide upload-outcome metrics (`webphone_paperless_uploads_total{outcome}`) or reject with rationale      | Medium | S        | Decision      |
| ~~ | 22   | Post-landing docs: AGENTS seam bullet + FEATURES row + CHANGELOG Unreleased                               | Medium | S        | Documentation |
| ~~ | 23   | Verify NixOS module needs no change (`settings` freeform carries the new keys) — note the verdict         | Low    | S        | Documentation |
| ~~ | 24   | Outbound-fax archiving on `transmitted` (same seam, direction-flavored title/tag)                         | Low    | M        | Feature       |
| ~~ | 25   | `archive_status` column + boot-time sweep (self-healing upload queue; closes the SIGTERM gap)             | Low    | M-L      | Feature       |
| ~~ | 26   | Retry-posture decision: SDK `WithRetry` wrapper vs fire-and-forget-forever for a routinely-down Paperless | Low    | S        | Decision      |
| ~~ | 27   | Deep-link "open in Paperless" affordance from the fax panel (needs document ID round-trip — post-#25)     | Low    | M        | Feature       |
| ~~ | 28   | v2.7.0+ release tail + deploy (OWNER terminal; everything release-ish queues behind it)                   | High   | S-M      | Release       |
| ~~ | 29   | cqrs-htmx setup-shell adoption train (upstream gaps first)                                                | High   | L        | Feature       |
| ~~ | 30   | Prod outbound-SMS 422 root cause (OWNER journal/ssh on pbx)                                               | High   | S        | Bug           |
| ~~ | 31   | erraudit tier-2 re-measure (next due 2026-10-22; must stay 0)                                             | High   | S        | Quality       |
| ~~ | 32   | Nix-review follow-ups batch (eval pins, VM asserts, release.sh preflight)                                 | High   | S-M      | Quality       |
| ~~ | 33   | OWNER-calls batch session (~15+ pending decisions incl. release number)                                   | High   | S        | Decision      |
| ~~ | 34   | Review-series deltas: `Receipt.Resolution` enum (kills the `*gateway.Loopback` type-assert)               | Medium | S        | Quality       |
| ~~ | 35   | Review-series deltas: config fail-closed typing for `identities`/shared contacts                          | Medium | S        | Quality       |
| ~~ | 36   | Review-series deltas: `schema_version` table before first ALTER (gated on its arrival)                    | Medium | S        | Quality       |
| ~~ | 37   | Gateway honest-Content-Type follow-ups (header-block golden, compat matrix, webhook-mode smoke probe)     | Medium | S        | Quality       |
| ~~ | 38   | `/version` enrichment via ldflags + vcs.buildinfo fallback                                                | Medium | M        | Feature       |
| ~~ | 39   | internal/server god-package carve (trigger-based: next file added)                                        | Medium | M        | Cleanup       |
| ~~ | 40   | templ-components release-gate leg (stack E2E rides the release tail)                                      | Medium | S        | Quality       |
| ~~ | 41   | AGENTS.md compaction ≤377 lines (owner-permission gate)                                                   | Medium | M        | Documentation |
| ~~ | 42   | Tooling hygiene batch (markdown-lint posture, codespell policy, render-diff script commit)                | Low    | S        | Cleanup       |
| ~~ | 43   | Post release announcements (v2.1–v2.7 drafts exist)                                                       | Low    | S        | Documentation |
| ~~ | 44   | Standing watches quarterly re-check (sip.js, templ-components, E2E budget; due 2026-12-20)                | Low    | S        | Quality       |

**HARVEST note:** items 1–6, 21, 26–27 are NOT yet routed anywhere durable
(1 is answered by replying, 2–6/21 are plan-doc amendments, 26–27 are new).
Per the owner's "THEN WAIT FOR INSTRUCTIONS" I am NOT running docs-health
HARVEST now — say the word and 2–6 fold into the plan doc / TODO row, 26–27
go to ROADMAP.

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

~~1. **Go/no-go + ordering:** implement the fax→Paperless integration now~~ resolved by events — shipped as T14 after the v2.8.0 release (train ordering answered by events)
(S-M), or park as PLANNED? If go — before or after the v2.7.0+ release
tail ships? (I could start immediately; what I can't decide is whether it
jumps the release queue — that's a train-ordering owner call.)
~~2. **Where does Paperless-ngx live?** Is there an existing instance (I have~~ routed — the stack leg (services.webphone.paperless module option + smoke arm) lives in the cross-repo row
NOT looked into `~/projects/nix-international-telephony` per your
don't-research instruction), or should the NixOS stack grow a paperless
service — making this a fourth tri-repo surface (module option,
`environmentFile` token, nginx)? The hosting decision changes whether
webphone's config stays app-level or the stack runbook gains a leg.
~~3. **v1 scope confirm:** inbound-only as designed, or fold~~ resolved by events — inbound-only shipped as v1 (outbound out-of-scope, CHANGELOG)
outbound-transmitted archiving into v1 (same seam, roughly +50 lines)?
I assumed inbound-only because received documents are the filing problem;
flip it and the plan's scope section changes before implementation.
