> ARCHIVED 2026-11-08 (docs-health drain, portfolio-plan P2-1): gates clean (check-rows --strict + bare-rows green), citations repointed — THE DURABLE RECORD of the JSContact REJECT + reopen trigger (a named consumer ingesting JSContact but not vCard, Ledger-domain first): reopening means re-running verify-external-claims on the RFC history and evaluating CardDAV before JMAP. The prior-art hole closed 2026-11-08 (repo-wide grep clean); the Q10 export-loop candidate is adjudicated by the 2026-10-05 dedup-registry boundary ruling.

# JMAP/JSContact Contacts-Interop Advisory Session — Status + Self-Review

- **Date**: 2026-10-16 20:41 CEST
- **Session shape**: 3 advisory Q&A turns (JMAP support? → JMAP as interchange format? → JSContact for webphone?). **Zero code edits, zero commits, zero gates touched.** This report is session-scoped per instruction: it covers only what this session did and noticed.
- **Session verdicts delivered**:
  1. webphone has **zero JMAP support**; contacts interop is vCard files in/out (`internal/vcard`, `/contacts/import`, `/contacts/export`, T26c zip) + `/api/contacts`.
  2. JMAP = excellent **sync protocol**, poor **interchange format** (category confusion); ecosystem thin for contacts (JMAP Contacts stripped from RFC 8621; adopters ≈ Fastmail/Stalwart/Cyrus; Gmail/MS absent).
  3. **JSContact: NO** for webphone — two-field contact model (`vcard.Card{Name, Number}`), interop runs the wrong way, collides with the Ledger-domain ruling; reopen trigger = a _named_ consumer that ingests JSContact but not vCard (≈ a JMAP AddressBook story, which is Ledger-domain first).

---

## a) FULLY DONE

| #  | Item                                                                                                        | Evidence                                                                                                                                                   |
| -- | ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~a1~~ | ~~Repo-wide JMAP absence confirmed (case-insensitive, all files incl. docs)~~ done — historical advisory fact; still true at HEAD (no JMAP code) | ~~grep `jmap\|JMAP\|Jmap` → zero matches~~ |
| ~~a2~~ | ~~Contacts interop surface fully mapped~~ done — historical; the surface is unchanged in kind (vcard in/out + /api/contacts) | ~~`internal/vcard/vcard.go`, `actions.go` import/export handlers, `export.go` zip lane, `contacts_api.go`, FEATURES.md:88 🟢 FULLY_FUNCTIONAL~~ |
| ~~a3~~ | ~~External fetch of jmap.io performed~~ done — historical; see §b 1 for its provenance limits | ~~core=RFC 8620, mail=RFC 8621 confirmed; contacts/calendars specs exist on the site~~ |
| ~~a4~~ | ~~Three verdicts delivered, each with a precise reopen trigger~~ done — the durable record is this archived report's banner (REJECT + reopen trigger) | ~~conversation + this report~~ |
| ~~a5~~ | ~~Report-time grounding checks (2 greps + 2 ROADMAP reads): verdicts **align with existing ratified rulings**~~ done — the M23 Ledger ruling still owns contacts depth (ROADMAP, grep this session) | ~~ROADMAP.md:402-403 (M23: contacts depth → Ledger; single-vCard export explicitly excluded), dedup-registry has NO ruling on the noticed export duplication~~ |

## b) PARTIALLY DONE

| #  | Item                               | Gap                                                                                                                                                                                                                                                                                                                                                                                                           |
| -- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~b1~~ | ~~External-claim verification~~ done — moot: the REJECT never rested on RFC history (cost-of-wrongness ~0, stated in g3); the verify discipline is now the mandatory verify-external-claims skill | ~~Only the jmap.io **homepage** was fetched. The two fetches that would have retired risk — `/spec/` (RFC vs draft status) and `/software/` (implementation list) — were NOT made. RFC-history claims (JMAP Contacts removed from RFC 8621; RFC 9553 = JSContact Card; JMAP-for-Voicemail RFC existence/number; Gmail/MS absence; Fastmail/Stalwart/Cyrus trio) are **memory-sourced, unverified this session**~~ |
| ~~b2~~ | ~~In-repo prior-art check~~ done — CLOSED: the repo-wide grep ran 2026-11-08 — zero prior art beyond a test-file reference (confirms the turn-3 premise) | ~~JMAP/CardDAV/vCard grepped (turn 1), but the string **"JSContact" was never grepped in any turn** — turn 3's "no prior art / no consumer" premise rests on inference, not search~~ |
| ~~b3~~ | ~~Durable recording of the rejection~~ done — this archive's banner carries the REJECT + reopen trigger; the M23 Ledger ruling owns the contacts domain (grep this session) | ~~The JSContact REJECT + reopen trigger lives only in conversation + this report. No TODO/ROADMAP row yet (gated on owner call, question g2)~~ |

## c) NOT STARTED

1. ~~`grep -i jscontact` repo-wide (closes the b2 hole; ~1 minute).~~ done — ran 2026-11-08: zero prior art beyond a test-file reference; premise confirmed.
2. ~~jmap.io `/spec/` + `/software/` + datatracker verification pass (closes b1).~~ done — NOT-DO: the advisory REJECT is archived with its reopen trigger; its RFC-history footnotes were never load-bearing.
3. ~~docs-health HARVEST of this report's (f) items into TODO_LIST/ROADMAP — skill-mandated on session continuation, gated by the WAIT instruction.~~ done — this archive IS the harvest; the operative items adjudicated in place.
4. ~~dedup-registry sweep-log entry for the noticed export duplication (d-side candidate, needs judging first — see e/Q10).~~ done — adjudicated by the 2026-10-05 registry ruling's boundary-class (dedup-registry:108, grep this session); no new entry needed.

## d) TOTALLY FUCKED UP

**Nothing repo-harming** — no file was touched, so no build/test/CI risk exists. The failures are **epistemic**, and they are real:

1. **Unearned confidence (worst failure of the session).** RFC-level history was asserted with verdict-bearing confidence from memory. The `verify-external-claims` skill is loaded in this environment, its trigger fires exactly on this shape ("before encoding any external claim into … documentation, code, or review"), and it was neither invoked nor manually emulated. Five-plus claims carried provenance they didn't have.
2. **The one fetch made was the wrong one.** The homepage fetch returned marketing copy (benchmarks, power-usage numbers) that was **never cited**; the fetches that mattered weren't made. Effort spent ≠ risk retired.
3. **Missing cheap grep.** Turn 3 could have run one `grep -i jscontact` to ground its central premise. Seconds. Skipped.
4. **Turn 3 ran entirely tool-free.** Legitimate for a judgment call on loaded context, but it silently presented grounded facts (two-field model, boundary trio) and memory facts (JSContact schema shape, mandatory `uid`) **at the same confidence level** — the reader had no way to tell them apart.

## e) WHAT WE SHOULD IMPROVE (brutal-self-review, Q1–Q11)

| Q                         | Answer                                                                                                                                                                                                                                                                                                                                                 |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ~~Q1 forgot?~~             | ~~The JSContact grep; the verify-external-claims discipline; recording the rejection durably~~ done — all three closed: grep run 2026-11-08, the skill is now mandatory, the banner carries the record                                                                                                                                                                                                                                                             |
| ~~Q2 stupid-we-do-anyway~~ | ~~Advisory answers mix loaded-context facts and memory facts at one confidence level; verdict speed prioritized over claim provenance~~ done — process observation, archived; the label discipline landed via the skill                                                                                                                                                                                                                    |
| ~~Q3 done better?~~        | ~~Fetch `/spec/`+`/software/` instead of the homepage; run one grounding grep per newly-named format; label external claims memory vs verified~~ done — process observation, archived                                                                                                                                                                                                           |
| ~~Q4 still improve?~~      | ~~The (f) list; a personal hard rule: no RFC/standard claim in a verdict without "verified" or "memory" label~~ done — the hard rule is the verify-external-claims skill's job now                                                                                                                                                                                                                                            |
| ~~Q5 lied?~~               | ~~**No fabricated facts detected** — but confidence inflation over unverified claims is the honest close cousin, flagged in (d)~~ done — historical self-audit, archived                                                                                                                                                                                                                          |
| ~~Q6 less stupid?~~        | ~~Make "verify or label" non-optional for standards claims that bear on accept/reject verdicts~~ done — superseded by the mandatory skill                                                                                                                                                                                                                                                           |
| ~~Q7 ghost systems?~~      | ~~None — nothing created this session; nothing unwired noticed~~ done — historical (zero-code session)                                                                                                                                                                                                                                                                                           |
| ~~Q8 scope creep?~~        | ~~**Avoided, correctly**: stayed advisory; did NOT implement a JSContact lane (would have collided with the M23 Ledger ruling)~~ done — historical; the M23 ruling stands                                                                                                                                                                                                                           |
| ~~Q9 removed useful things?~~ | ~~Nothing removed~~ done — historical                                                                                                                                                                                                                                                                                                                                        |
| ~~Q10 split brains?~~ | ~~**Noticed unjudged candidate**: the vCard-export card-building loop exists twice — `actions.go` (`exportContacts`, :698) and `export.go` (:169) both build `[]vcard.Card{Name, Number}` identically. Dedup-registry checked at report time: **no ruling**. Candidate for the next art-dupl sweep / one-home helper — NOT for unilateral action now~~ done — adjudicated by class: the 2026-10-05 registry ruling REJECTED consolidating the Name+Number boundary builders (dedup-registry:108, grep this session); both sites remain accepted similarity under that doctrine |
| ~~Q11 tests?~~             | ~~N/A (no code changed). Test-adjacent notice: the vcard-parser fuzz target remains ROADMAP-unstarted (ROADMAP.md:578-580; only `FuzzContactsAPISave` exists)~~ done — routed: the fuzz-target row lives in ROADMAP (grep this session)                                                                                                                                                                                            |

## f) NEXT (session-scoped; honest count is 10, not padded to 50 — most "50-item" filler would violate the no-unrelated-research instruction)

Sorted impact/effort:

1. ~~**Owner call → record the JSContact REJECT + reopen trigger durably** (ROADMAP rejected-ideas section or an AGENTS.md contacts-seam line) — ~10m; prevents future re-litigation (the exact failure class the dedup-registry pattern exists for).~~ done — the archive banner is the durable record (REJECT + reopen trigger); the M23 Ledger ruling owns the domain, making a separate doctrine row redundant.
2. ~~**Verify the RFC claims** (jmap.io `/spec/` + datatracker: 8621-contacts removal, 9553 identity, JMAP-for-Voicemail RFC number) — ~15m; retires the session's biggest epistemic debt if the verdict is to be treated as doctrine.~~ done — NOT-DO: the advisory REJECT is archived with its reopen trigger; its RFC-history footnotes were never load-bearing.
3. ~~**`grep -i jscontact` repo-wide** — ~1m; closes the turn-3 grounding hole (expected: zero).~~ done — ran 2026-11-08, clean (see §c 1).
4. ~~**Fetch jmap.io `/software/`** — ~5m; verifies the Fastmail/Stalwart/Cyrus adoption claim.~~ done — NOT-DO: same as §f 2; the adoption claim never bore weight.
5. ~~**HARVEST this report** (docs-health) once the owner green-lights — mostly routing items 1-2.~~ done — this archive IS the harvest.
6. ~~**Next art-dupl sweep: judge the dual export card-building sites**; if ruled a harmful clone → one-home `contactsCards` helper + mandatory registry sweep-log line.~~ done — adjudicated: the 2026-10-05 registry ruling's boundary-class covers it (dedup-registry:108, grep this session); accepted similarity.
7. ~~**vcard parser fuzz target** (already ROADMAP-fueled, listed here only because noticed — not new).~~ done — routed: ROADMAP carries the fuzz-target row (grep this session).
8. ~~**If contact sync is EVER reopened: evaluate CardDAV before JMAP** (session conclusion; ROADMAP-fuel).~~ done — archived fuel; the reopen trigger in the banner names the door.
9. ~~**Process improvement, zero repo cost**: adopt the claim-provenance label (memory/verified) in advisory answers.~~ done — superseded: the verify-external-claims skill is mandatory in this environment and owns the discipline.
10. ~~_Checked, no action_: T26c zip export already lives at ROADMAP.md:505; vCard feature status green in FEATURES.md:88.~~ done — historical grounding check, archived.

**Explicitly NOT recommended** (checked at report time, do not re-surface): single-contact vCard export — RULED OUT by M23 (ROADMAP.md:402-403, Ledger's domain); JSContact export alias — rejected this session pending the reopen trigger.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Is there a real consumer behind the JMAP/JSContact question** — a client you use, a Ledger/CRM sync plan — or was it pure exploration? The reopen trigger's sharpness depends on owner intent that is not derivable from the repo.
2. **Should the JSContact REJECT + trigger become durable doctrine** (and if so, ROADMAP vs an AGENTS.md contacts-seam line)? Placement is an owner call under the one-home rule; the Ledger ruling might make it moot.
3. **Is memory-sourced RFC history acceptable grounding for a REJECT verdict** whose cost-of-wrongness is near zero — or do you want the `/spec/` + datatracker verification pass (f2) done before the verdict is citable?

---

_Reported 2026-10-16 20:41 CEST. Auto-commit daemon will pick this file up; no manual commit per harness contract. Waiting for instructions._
