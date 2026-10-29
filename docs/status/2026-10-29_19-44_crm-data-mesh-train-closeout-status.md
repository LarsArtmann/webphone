# Status — CRM data-mesh train, CLOSED (what I forgot, what to fix next)

**Written:** 2026-10-29 19:44 CET
**Verified as of:** 19:44 (git state + CI re-checked seconds before writing;
gate numbers are from the runs recorded in the plan doc's §6)
**webphone:** `main` @ `5222095d`, tree clean, CI run **40928206449 = success**
(the code run **40927955113 = success** too)
**Kith (CRM):** `master` @ `d57e00f`, **16 commits ahead of `origin/master`,
NOT pushed**
**Plan of record:** [2026-10-29_18-13_SUPERB-crm-data-mesh-completion.md](../planning/2026-10-29_18-13_SUPERB-crm-data-mesh-completion.md)
**Supersedes:** [2026-10-29_18-06_crm-data-mesh-midtrain-status.md](2026-10-29_18-06_crm-data-mesh-midtrain-status.md)

---

## 0. Verdict

The train works and is green, in both repos and in CI. A caller the CRM
knows is now **named, linkable, and contextualised** in the phone, and the
phone still stores nothing the CRM owns. That is the whole point and it is
done.

What is **not** done is everything around the edges: two fields I carried
"for free" and never rendered, one quiet-failure class that skipped its
required doc row, two repo-convention documents that were never touched at
all, one visual proof I decided not to attempt, and a producer change that
sits unpushed in a sibling repo because another session's work is in flight
there. None of those break the product; all of them are the reasons a future
session would rewrite something.

---

## a) FULLY DONE (verified by a gate, not by intent)

| #   | Thing                                                                                                                                                                                                                                                                                  | Evidence                                                                          |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| A1  | **Root cause fixed**: the seam's type was a lossy `number → string`, so the CRM contact id died at the boundary and nothing could link to a named caller. Now `crm.Person{Ref, Name}` + `Resolver.People`.                                                                             | `internal/crm/client.go`, `resolver.go`; crm package tests green                  |
| A2  | **Deep links** `crm.contact_url` (template with required `{id}`, requires `crm.url`, absolute http(s), fails the boot otherwise), rendered in a NEW TAB so the SIP island is never unloaded.                                                                                           | 5 config cases in `TestLoadValidatesCRMConfig`; `views.crmNameLink`; 3 view tests |
| A3  | **One conversion home** `server.crmContacts` + **one link home** `views.crmNameLink` + **one row-label home** `views.remoteLabel`; views never import `internal/crm`.                                                                                                                  | `go build`, server tests                                                          |
| A4  | **Context product, producer side**: `GET /api/contacts/{id}/context` in Kith (name, email, company, tags, open deals capped at 5 + true count, open task count, next due task with the producer's own overdue verdict, last activity). Bearer-gated, owner-computed, no consumer join. | Kith `internal/machineapi/context.go`; **52 BDD specs + the new suite green**     |
| A5  | **Context product, consumer side**: `crm.Context`, `Client.ContactContext` (404 ⇒ `ErrNoContext`, 401 ⇒ `ErrUnauthorized`), `Resolver.ContextOf` with a 2-min positive cache, 15-min absence cache, 256-entry bound, and **no caching of transport failures**.                         | 6 new tests in `internal/crm/context_test.go`                                     |
| A6  | **The strip is reachable and rendered**: `ThreadViewProps.CRMContext` is populated in `threadPanelPrefilled` for telephony threads only, degrading to `nil` (no markup at all) on every miss.                                                                                          | `TestThreadPanelRendersTheCRMContextStrip` + 2 degradation tests                  |
| A7  | **CSS exists** for the link and the strip (tokens only, both themes, focus ring), and the artifacts were rebuilt the mandatory way.                                                                                                                                                    | `.wp-crm-*` rules in `app.css`; `scripts/build-tw-css.sh`; `treefmt` check green  |
| A8  | **Vocabulary bridge**: the CRM closed its outcome set to `connected                                                                                                                                                                                                                    | voicemail                                                                         |
| A9  | **Gates**: full `go test ./...` green, `-race` green (crm/views/server), `tsc` green, island **283/283**, five flake checks green, smoke **49/4/8** with 0 failures, `buildflow` **69 success / 0 failed**, `art-dupl` **0 new** (83 → 86 after two real extractions), CI green twice. | plan doc §6                                                                       |
| A10 | **Docs**: error-code registry (+4 codes), AGENTS data-mesh invariant, CHANGELOG (3 Added / 1 Changed), dom-contract optional ids, README (config row + both integration bullets), dedup sweep-log line, plan doc with mermaid + a filled ledger + a gate-evidence table.               | `ee6628e4`, `5222095d`                                                            |

---

## b) PARTIALLY DONE

| #  | Thing                                  | What exists                                                                                                             | What is missing                                                                                                                                                                                                                       |
| -- | -------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| B1 | **Producer is committed, not shipped** | Kith has my `context.go`, its tests, the route, and two stale-spec fixes, all committed by that repo's daemon           | **Not pushed** (`master` 16 ahead of `origin/master`) and **not mentioned in Kith's CHANGELOG or its generated wire reference** (`/docs`, OpenAPI/AsyncAPI) — the data product is undocumented where that repo documents its products |
| B2 | **Visual proof**                       | Server tests assert the rendered HTML; the CSS was reviewed against the house design language (flat row, no box-in-box) | **No screenshot.** M13 was planned and not attempted; `ui-capture.py` has no fake CRM, so the harness cannot shoot the strip today                                                                                                    |
| B3 | **Accessibility of the new text**      | Semantic markup (`aside` + `aria-label`), muted-but-token colours, focus-visible ring                                   | **Contrast never measured.** `.wp-crm-key` is 0.85em uppercase muted on `--surface`; the repo claims AA elsewhere, and this is the one claim I did not verify                                                                         |
| B4 | **Dead fields tolerated**              | `crm.Context.Tags` and `crm.Context.Name` are populated from the wire                                                   | **Neither is read anywhere in webphone** (only by tests). Two speculative fields carried "because they were free"                                                                                                                     |
| B5 | **Error-contract narrative**           | The generated registry table carries all 4 new codes                                                                    | **No row in "Detection column — how anyone would KNOW a quiet class happened"**, whose own contract says a new quiet class gets a row _in the same commit that ships it_. "CRM context absent / unreachable" is exactly such a class  |
| B6 | **Commit narration**                   | Two narrative commits of mine (`ee6628e4`, `5222095d`)                                                                  | The **code** history is 4+ `chore: auto-commit N changed file(s) (heuristic)` messages: the daemon won the race because I committed per-phase instead of per-task, which the runbook explicitly warns about                           |

---

## c) NOT STARTED

| #  | Thing                                                                                                                          | Why it matters                                                                                                                                                                    |
| -- | ------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| C1 | Links on the remaining safe surfaces: voicemail's `vmCaller` chain (mixed name/number text), the drafts panel's `contactLabel` | Three surfaces link, two do not — an inconsistency a user will notice before I do                                                                                                 |
| C2 | Context anywhere other than the thread detail (History's per-call detail, the incoming-call card)                              | Owner question Q4 defaults to NO for History (fetch multiplication); the call card is genuinely undecided                                                                         |
| C3 | "Not in the CRM — add them" affordance for misses                                                                              | Today a miss is a dead end; a deep link into the CRM's new-contact form would close the loop without webphone writing anything                                                    |
| C4 | Batch lookup on the producer (`by-phone?numbers=`)                                                                             | Every panel currently costs up to N requests (bounded at 8 concurrent), all cache-missed on first paint                                                                           |
| C5 | `FEATURES.md` and `TODO_LIST.md`                                                                                               | **Never touched.** The repo's rule is explicit: FEATURES inventories status, TODO_LIST holds open work. The deferred tail currently lives only in a planning doc — the wrong home |
| C6 | Full `nix flake check` as ONE gate (module check, error drill, KVM-gated backup + backup drill)                                | Same gap as the previous train; I ran the five cheap checks + buildflow instead                                                                                                   |
| C7 | Kith's own gates (buildflow/CI there) and its catalog entry for the new endpoint                                               | Its `internal/catalog` gate is currently RED from another session's claim/outreach events, so I could not get a clean whole-repo verdict                                          |

---

## d) TOTALLY FUCKED UP

| #  | Fuck-up                                                                                                                                                                                                                                                                                                                                                      | Severity        | Detail                                                                                                                                                                                                                              |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D1 | **I shipped a live integration break and only found it by accident.** The CRM's closed outcome vocabulary rejects `answered`; webphone sent exactly that. Every answered call would have journaled as a 400 → island 502 toast. I found it because a _test fixture_ I wrote used `"answered"` and got a 400 — not because any gate, doc, or alert caught it. | **High**        | Fixed (A8), but the lesson is the failure mode: a sibling repo changed a wire contract and nothing in this repo would have noticed. There is still **no contract test against a real CRM** — the stub is mine, so it agrees with me |
| D2 | **I broke a documented doc contract in the same commit I shipped the feature.** `error-contract.md` § "Detection column" says a new quiet class gets a row at ship time. I added four registry codes and skipped the row.                                                                                                                                    | Medium          | B5. The class here: _silent absence_ (context 404 / CRM down / not published) — the exact thing that section exists for                                                                                                             |
| D3 | **Two dead fields.** `Context.Tags` and `Context.Name` ride the wire and are read by nobody. I violated my own "every field has a reason you can articulate" rule — "it was free" is not a reason.                                                                                                                                                           | Low-Medium      | Cost: a reader assumes they are rendered somewhere; the next session either builds UI to justify them or deletes them                                                                                                               |
| D4 | **I let the daemon write the code history.** Four `chore: auto-commit` commits carry the entire seam rebuild, the config knob, the strip, the CSS and the tests. The two commits that read like a human wrote them are docs-only.                                                                                                                            | Medium          | Preventable exactly as the runbook says: stage+commit in the SAME tool call as the verification. I verified 12 times and committed twice                                                                                            |
| D5 | **`buildflow` ran against a tree I was actively editing**, failed `test-race`, and cost ~10 minutes of re-running plus a false alarm.                                                                                                                                                                                                                        | Low             | My own interference, not a code defect. Freeze the tree before a 10-minute gate                                                                                                                                                     |
| D6 | **One edit was silently lost mid-session.** My `client.go` payload change was reverted by daemon activity and survived only because a test failed loudly and I re-checked with `grep`.                                                                                                                                                                       | Low (recovered) | The repo documents this hazard; the mitigation is "re-read before each edit + verify with grep after", which I did only after being burned                                                                                          |
| D7 | **Kith was left with two red gates I caused-or-touched** and I fixed both without a narrative commit: a `byte`/`int` type error from a parallel edit (blocked every build) and two BDD specs asserting freeform outcomes (now obsolete).                                                                                                                     | Low             | Both were minimal, intent-preserving fixes, but they are now anonymous `chore:` commits in a repo I do not own                                                                                                                      |
| D8 | **Wrong-ish citation in AGENTS**: I wrote `crm.outcomeForDisposition` (the table is `outcomeForDisposition`, the function is `crmOutcome`). Stable-name citing is the rule; this one is ambiguous.                                                                                                                                                           | Low             | Trivial to fix, but it is the kind of imprecision that sends a reader to the wrong symbol                                                                                                                                           |

---

## e) WHAT SHOULD IMPROVE — the systemic reading

1. **Cross-repo contracts are checked by hand.** The vocabulary break (D1)
   was caught by luck. There is a `contracts-fresh` check for _our_ generated
   contracts and nothing for _the neighbour's_ wire. Options: a recorded
   fixture of the producer's responses checked in beside the client (a
   "contract snapshot"), or a Kith-side test that our exact payload is
   accepted. The second is stronger and belongs in Kith, where the vocabulary
   is owned.
2. **"Quiet by design" is a doc contract nobody enforces.** D2 happened because
   the rule is prose. A one-line test could assert that each code in the
   registry appears either in the detection table or in an explicit
   `loudByDesign` list. That is the same shape as the config-keys-are-documented
   test that already exists and works.
3. **Dead wire fields should fail a test.** A "every field of `crm.Context` is
   read by at least one non-test caller" assertion would have caught D3. Cheap:
   a reflect-based reader in the package's test.
4. **Committing is the weakest link in an otherwise gated workflow.** The
   daemon is faster than I am, and the workaround (stage+commit in the same
   call as the verification) is known and written down. I ignored it.
5. **A visual gate would have caught B2/B3 for free.** The harness already
   boots the product and asserts DOM per shot; giving it a fake CRM is a
   ~20-line addition. Without it, every new visual affordance ships
   screenshot-unverified (`ui-shots/` is untracked now, so nobody notices).
6. **Two repos, two daemons, one narrative.** A train that spans repos ends up
   with chore commits in both. If cross-repo trains become normal, the
   release runbook needs a "who commits what, in which repo, in what order"
   paragraph (webphone first, then Kith, then the stack's re-pin).

---

## f) THE NEXT 50, ordered by value ÷ effort

**Tier 0 — close this train's own holes (~2 h)**

| # | Task                                                                                                             | Min |
| - | ---------------------------------------------------------------------------------------------------------------- | --- |
| 1 | Add the "CRM context absent/unreachable" row to `error-contract.md` § Detection column (D2)                      | 10  |
| 2 | Push Kith (`master` is 16 ahead) after its catalog gate is green, so the producer actually ships                 | 15  |
| 3 | Kith CHANGELOG + wire-reference entry for `GET /api/contacts/{id}/context`                                       | 25  |
| 4 | Delete or render `Context.Tags` / `Context.Name` (D3) — my default: delete, add back with the UI that needs them | 20  |
| 5 | Add `FEATURES.md` rows for both capabilities (C5)                                                                | 15  |
| 6 | Move the deferred tail from the plan doc into `TODO_LIST.md` (C5)                                                | 20  |
| 7 | Fix the AGENTS citation `crm.outcomeForDisposition` → `crm.crmOutcome` (D8)                                      | 5   |
| 8 | Test that `crmContextFor` refuses matrix/session addresses (the guard exists, nothing pins it)                   | 15  |

**Tier 1 — make the new thing provable (~3 h)**

| #  | Task                                                                                  | Min |
| -- | ------------------------------------------------------------------------------------- | --- |
| 9  | Give `scripts/ui-capture.py` a fake CRM (by-phone + /context) behind a flag           | 30  |
| 10 | `--surface thread` shots with the strip on, both themes; assert the classes           | 25  |
| 11 | Shots with `crm.contact_url` set, to see the heading link                             | 15  |
| 12 | Measure AA contrast for `.wp-crm-key` / `.wp-crm-muted` on `--surface` in both themes | 20  |
| 13 | German eye-pass over the strip (8 new keys)                                           | 15  |
| 14 | Contract snapshot test: recorded Kith responses decoded by the client (D1 mitigation) | 45  |
| 15 | Kith-side test: the exact payload webphone sends is accepted (stronger twin of 14)    | 30  |
| 16 | A reflective "every `crm.Context` field is read by a non-test caller" test (D3 class) | 30  |

**Tier 2 — finish the surfaces (~4 h)**

| #  | Task                                                                                                       | Min |
| -- | ---------------------------------------------------------------------------------------------------------- | --- |
| 17 | Link the voicemail caller label (needs `vmCaller` split into name + fallback text)                         | 30  |
| 18 | Link the drafts panel's contact label                                                                      | 20  |
| 19 | Decide + implement the call card's CRM context (or record the NO)                                          | 45  |
| 20 | "Add to CRM" deep link on a miss (config template + view affordance, no write)                             | 45  |
| 21 | Batch `by-phone?numbers=` on the producer, consumer joins it behind a feature test                         | 60  |
| 22 | Cache-key normalisation: the same contact reached via two number spellings costs two context entries today | 30  |

**Tier 3 — governance and tooling (~4 h)**

| #  | Task                                                                                                       | Min |
| -- | ---------------------------------------------------------------------------------------------------------- | --- |
| 23 | Test: every registered error code appears in the detection table or an explicit loud list (e)              | 45  |
| 24 | Run `nix flake check` as one gate and record the verdict (C6)                                              | 60  |
| 25 | Kith buildflow + CI green (blocked on its catalog drift, which is another session's)                       | 45  |
| 26 | NixOS module: derive `crm.contact_url` from a CRM vhost (owner Q2)                                         | 40  |
| 27 | Module option for `crm.contact_url` + `webphone-module` stand-in update                                    | 30  |
| 28 | Stack browser-E2E: re-run with the new ids (already a standing obligation)                                 | 60  |
| 29 | Smoke lane with a CRM stub, so `webphone-smoke.py` exercises the strip                                     | 45  |
| 30 | Record a `crm.*` metrics counter for context lookups (hit/absent/failure) so the quiet class is measurable | 30  |

**Tier 4 — product depth (needs owner input, not mine)**

| #  | Task                                                                                                                | Min |
| -- | ------------------------------------------------------------------------------------------------------------------- | --- |
| 31 | Decide Q3: SMS/voicemail/fax facts into the CRM timeline                                                            | —   |
| 32 | If GO: metadata-only activity rows (no bodies) + privacy note                                                       | 90  |
| 33 | If GO: SMS bodies behind an explicit opt-in key, per-extension                                                      | 90  |
| 34 | Deal-stage semantics in the strip (e.g. highlight "negotiation" as actionable)                                      | 30  |
| 35 | Click-through from a deal chip into the CRM's deal page (second template key)                                       | 30  |
| 36 | Show the contact's last NOTE, not just a timestamp (needs a producer field)                                         | 60  |
| 37 | Owner-scoped context: today the CRM is single-user, but the seam leaks no owner — document or fix before multi-user | 45  |
| 38 | Tag chips (uses the field D3 would otherwise delete)                                                                | 20  |
| 39 | Overdue-task emphasis in the thread head (a badge, not a muted phrase)                                              | 30  |
| 40 | Decide whether the strip belongs above or below the self-send notice                                                | 10  |

**Tier 5 — debt noticed this session, unrelated to the feature**

| #  | Task                                                                                                                                              | Min |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------- | --- |
| 41 | `.buildflow-baseline.json` is stale (78 entries matching nothing for 3 runs) — refresh or retire, same ruling as the branching-flow baseline      | 30  |
| 42 | `test-race` 55 s → 102 s and `golangci-lint` 1.3 s → 4 s: confirm host load vs a real regression                                                  | 30  |
| 43 | Kith: the two specs I rewrote should be reviewed by whoever owns the closed-vocabulary train                                                      | 15  |
| 44 | Kith: `internal/catalog` drift from the concurrent claim/outreach work (their gate, my discovery)                                                 | —   |
| 45 | Kith: `projections.go` was missing a `time` import — check whether the same edit left other build breaks                                          | 20  |
| 46 | `views.CRMContact` vs `crm.Person` — re-read in 6 months; if a second consumer appears, that projection is the one to question                    | 20  |
| 47 | `docs/status/2026-10-29_18-06_*` is annotated superseded; sweep the status dir for other stale snapshots                                          | 30  |
| 48 | Add the "two repos, two daemons" paragraph to the release runbook (e)                                                                             | 25  |
| 49 | Re-read `docs/planning/2026-10-29_17-52_*` (the design doc) against what shipped — it predates the vocabulary finding                             | 20  |
| 50 | Quarterly dedup audit is due 2027-01-11; the 86-entry baseline gained 3 markup-shape groups today, all candidates for an upstream art-dupl report | 30  |

---

## g) THREE QUESTIONS ONLY THE OWNER CAN ANSWER

**Q1 — Should SMS / voicemail / fax facts reach the CRM contact timeline?**
Today only calls + dispositions journal. This is a privacy decision (message
bodies are personal data held by a different processor) and a product decision
(a timeline flooded by OTP codes is noise). It also decides whether the phone
becomes a _producer_ of more than call metadata — the largest remaining axis of
this integration, and not one I should pick.

**Q2 — Is the phone's CRM context allowed to become ACTIONABLE, or forever
read-only?** Right now the strip tells you a follow-up is due but you cannot
complete it, and a deal is in negotiation but you cannot move it. Making it
actionable means webphone writing to the CRM's aggregate (beyond call journaling),
which flips the ownership rule this train was built on. A deliberate yes/no.

**Q3 — Who pushes Kith?** The producer change is committed locally on `master`
(16 ahead) and mixed with another session's in-flight claim/outreach work whose
catalog gate is red. I will not push another session's unreviewed work, and I
will not rewrite their commits out. Do you want me to push the whole branch
as-is, cherry-pick only the CRM-context commits onto a clean branch, or leave it
for that session to land?

---

## h) Corrections ledger — where I caught myself mid-flight

| # | Claim I nearly made                                | Correction                                                                                                                                                                                    |
| - | -------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | "the strip works"                                  | It compiled, was wired, and rendered nothing until the producer existed: I only called it done after the server test asserted the HTML                                                        |
| 2 | "the nil guard is fine"                            | `if ctx != nil` compiled against templ's own render context (always true) and panicked on the first nil render; the test found it                                                             |
| 3 | "money formatting is trivial"                      | `formatThousands(-12000)` produced `-,999`; the sign was being grouped as a digit                                                                                                             |
| 4 | "sending our disposition verbatim is right"        | The CRM closed its vocabulary; verbatim was a 400 on every call. Corrected with a table + a two-sided test                                                                                    |
| 5 | "extract `statusFamily` and the registry is happy" | It red-lined `TestErrorCodeRegistryIsFresh` because the scan cannot see a code whose family argument is a function call; the family now rides a local variable, with the reason at both sites |
| 6 | "my grep filter is harmless"                       | `grep -v "^  "` silently ate every Go test failure message (they are indented). Twice.                                                                                                        |
| 7 | "the client edit landed"                           | It had been reverted by daemon activity; a failing test caught it, `grep` confirmed, re-applied                                                                                               |

---

## i) Verification vs speculation

**Verified by execution:** every gate row in §a/A9; Kith's machineapi suite
(52 specs + 4 new tests); Kith's whole-repo suite (only `internal/catalog` red,
from another session); `git ls-remote` for push state; `gh run view` for the two
CI conclusions; grep counts for the doc gaps (D2, D3, C5).

**Verified by reading only:** Kith's `app.System` read helpers and view shapes
(no test of my own exercises `FindDeals`/`FindTasks` beyond the seeded cases);
the CSS token values; the `wp-msg-detail` flex gap.

**Speculation, labelled:** that the strip looks right (no screenshot); that the
contrast passes AA (not measured); that the 15-min absence cache is the right
window (a product guess, not a measurement).

**Not checked at all:** the stack's browser E2E against the new markup; the
KVM-gated backup checks; whether Kith's catalog drift also needs an entry for
my endpoint once their train lands.

---

## Addendum — the Tier-0 tail, executed the same evening (20:30)

Everything this report called "speculation, labelled" and the §f items 1–2 are
now measured rather than guessed. Superseded claims are corrected inline below
rather than silently edited:

| Report claim                                 | Now                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| -------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "that the strip looks right (no screenshot)" | **Measured.** `scripts/ui-capture.py --crm` boots a bearer-gated fake Kith machine API beside the app and adds the `crm-context` surface. Two shots (`ui-shots/0*-crm-context-{light,dark}.png`) and three DOM pins: `.wp-crm-link` + the name, `#wp-crm-context` + company, and the deal title. Running it WITHOUT `--crm` exits 2 with a message, so the proof cannot silently degrade.                                                                |
| "that the contrast passes AA (not measured)" | **Measured**, worst case 5.17:1. `--muted` on `--surface` is 5.85:1 (light) / 6.89:1 (dark); on `--bg` 5.17:1 / 7.46:1. All ≥ 4.5:1, so `.wp-crm-key`/`.wp-crm-muted` clear AA at their ~12px size in both themes.                                                                                                                                                                                                                                       |
| the visual gap was cosmetic risk only        | **It was a real defect.** The first shot showed `Ada Lovelace(Open in the CRM)` underlined INSIDE the 24px thread heading: the link's screen-reader suffix used a made-up `wp-sr-only` class (the owned utility is `.sr-only`), so the hidden text rendered visibly. Fixed by moving the destination into `aria-label` (unstyleable, cannot leak), pinned by `TestCRMNameLinkIsTheOneLinkHome`. No gate catches a wrong class name; only a picture does. |

Other corrections carried in this pass:

- `crm.Context` lost `Name`, `Tags` **and `Ref`**: nothing read them. `Ref` is
  the interesting one — dropping it makes strip redirection structurally
  impossible (the value no longer exists off the wire), which is stronger than
  the test that used to guard it. `TestEveryCRMContextFieldIsRead` (AST-precise,
  sees only selectors off `published`) now fails when a field lands unread.
- The telephony guard was SMS-only while the identity feed asked about EVERY
  network: a Matrix thread's raw address reached `/api/contacts/by-phone`,
  which answers **400** for a value with no digits — a failure the resolver does
  not cache, so every render repeated it. Found while writing the pin for item
  7; `domain.Address.IsPhoneShaped()` is now the one rule, `crmNumberFor` the
  one projection, and `TestMatrixThreadNeverReachesTheCRM` fails without the fix
  (verified by reverting it: `[@ada:example.org +493012345678]`).
- Signal is now ELIGIBLE for the strip (its addresses are phone numbers) —
  previously it got a linked name from the identity product and no context, an
  asymmetry with no user-visible explanation.
- `docs/error-contract.md` gained the missing quiet-class row (CRM context
  absent), and its closing count corrected from five to six.

### Gate evidence for this addendum (frozen tree, 20:40)

`go build` + `go vet` clean · `go test -count=1 ./...` green over 31 packages
(the reflective and guard pins included) · `tsc -p . --noEmit` clean · island
283/283 · `nix fmt` 0 changed · five fast flake checks green · **`nix flake
check` all checks passed** (this is the §f item 5 "one gate" verdict: the
KVM-gated `webphone-backup`, `webphone-backup-drill`, `webphone-error-drill`
and `webphone-module` all ran) · `art-dupl` 0 new clones (baseline 86) ·
`buildflow` every step green · smoke 49 + 4 + 8 + 9 + 9 passed, 0 failed.

One defect surfaced by the full gate, not by the diff: `webphone-error-drill`
failed twice on an untouched boot path because a single-shot
`systemctl show … ExecMainStatus` read raced the unit's own `Restart=1s` loop.
Fixed in the check (retried assertion, taxonomy asserted first); rule recorded
in docs/lessons.md § Nix. No product change.
