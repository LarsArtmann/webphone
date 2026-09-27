# Dedup acceptance registry — the ONE home for clone rulings

Every art-dupl ACCEPT/EXTRACT ruling and its WHY lives here. This
supersedes the scatter across five point-in-time status reports
(2026-09-18, 09-23 01-12, 09-23 03-01, 09-23 15-58, 09-24 16-56 +
18-05 self-review) that the 18-05 self-review flagged as process debt;
those reports remain as historical provenance — annotate, never rewrite.

## Protocol (every dedup session)

1. Run art-dupl (working baseline `-t 3`; `-t 2` deep sweeps by owner
   request — baseline ratification is an open owner call).
2. Read THIS table before judging anything. A group that matches a
   standing ruling below is ACCEPTED — do not re-litigate it.
3. A NEW harmful group gets extracted (helper + micro-test in the same
   change set; templ refactors proven with an old-vs-new binary render
   diff), then a ruling row is added here.
4. Append one line to the sweep log. Nothing else.

A re-run regenerates the GROUPS cheaply; it cannot regenerate the
RULINGS — that is why this file exists (09-18 f.14 declared a
group-inventory register NOT-DO; the later sessions narrowed the need
to rulings, which the 18-05 e.6 proposal asked to consolidate — this
file implements that proposal and records the supersession).

## Standing rulings (ACCEPT unless the row says EXTRACTED)

| Clone shape (sites at last sighting)                                                                           | Ruling                                                                                          | Why                                                                                                                                                     |
| -------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Settings `dl` dt/dd rows (settings.templ) — surfaces every sweep                                               | ACCEPT, build-or-retire is an OWNER call (`settingsRow`)                                        | 7 rows with divergent value shapes (orUnset / plain / fmtInt / conditional / link); a component would half-cover. Owner asked to ratify 3× and counting |
| `wp-nav-badge` conditional spans (layout.templ nav ×2, messages thread row)                                    | ACCEPT                                                                                          | Deliberate hand-roll, pinned in AGENTS.md templ table; each site has a different condition + count source                                               |
| `T()`-paragraph windows (layout welcome block, settings themeNote, panelHead-internal lines)                   | ACCEPT                                                                                          | Different keys/classes/context; panelHead call sites ARE the dedup; themeNote is one of the three informational `wp-empty` occurrences                  |
| idempotency clock+lock prologue (idempotency.go `seen`/`record`)                                               | ACCEPT                                                                                          | In-code rationale at the site: check-then-record must stay separate steps; a lock helper would sever Lock from its deferred Unlock                      |
| `i++` / `continue` loop-skip branches (vcard unescape ×2, i18n_test %% skip)                                   | ACCEPT                                                                                          | Escape-handling idiom; the vcard branches handle different escape classes; the test site is unrelated logic, coincidental shape                         |
| Coincidental closing-`</span>` / muted-fragment display lines (contacts, messages thread row, attachment line) | ACCEPT                                                                                          | Different domain content behind a similar shape                                                                                                         |
| `@panelError(props.Error)` + `if !props.Enabled` needsAPI block (history, voicemail)                           | ACCEPT (error half EXTRACTED 09-24 → `panelError`)                                              | Call site = the dedup; needsAPI halves differ in keys and markup shape; a shared component takes 3 params for a 2-line body with an embedded `<code>`   |
| `wp-attach` empty dropzone div (fax + messages composers)                                                      | ACCEPT                                                                                          | One div per form; island JS wires each dropzone differently                                                                                             |
| `@panelHead(...)` + `@panelError(...)` prologue pair (fax, messages)                                           | ACCEPT                                                                                          | Both lines are extracted-helper call sites — the call sites ARE the dedup                                                                               |
| Row-head strong/muted lines (history CDRRow, voicemail VoicemailRow)                                           | ACCEPT                                                                                          | Different content expressions (cdrTarget/durationLabel vs vmCaller/vmWhen); abstraction would take more params than lines                               |
| `wp-segcount` + submit button pair (messages composers ×2)                                                     | ACCEPT                                                                                          | A component call is the same length; same-file widget shape                                                                                             |
| Conditional `@display.EmptyState(...)` (contacts panel, history, voicemail, threads)                           | ACCEPT (rationale updated 09-27 after the EmptyState adoption changed the old `wp-empty` shape) | Different conditions/keys/icons; a wrapper takes (condition, key, icon) — 3 params for a 1-call body                                                    |
| `data-i18n` spans (phone.templ island)                                                                         | ACCEPT — untouchable                                                                            | Island DOM contract, ported verbatim; island modules served verbatim + stack E2E greps them                                                             |
| Avatar rows, shared vs personal (contacts.templ)                                                               | ACCEPT                                                                                          | Different domain rules: read-only shared (tagged) vs mutable personal (delete affordance); avatar pair hand-roll is AGENTS-pinned                       |
| `contactSaveFailed(w); return` pair (actions.go, contacts_api.go)                                              | ACCEPT                                                                                          | Helper call sites ARE the dedup; contacts_api's site carries the extra ErrListFull 422 branch                                                           |

## Extracted on the record (the harmful ones, with their one-homes)

| What                                                                                                                                                                                 | One-home                                                                                                                                                                                                       | Sweep               |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------- |
| Inline session gates ×11                                                                                                                                                             | `server.requireSession`                                                                                                                                                                                        | 09-18               |
| Webhook multipart envelope                                                                                                                                                           | `gateway.providerForm`                                                                                                                                                                                         | 09-18               |
| Must-parser panics; missing provider timestamps; zero-rows UPDATE; language switch; CRM number collection; status-hook tail; call-log idem; contact-save 500 text; JSON 204 epilogue | `domain.must`, `domain.OrClock`, `store.updatedOrNotFound`, `views.formatFor`, `server.crmNumbers`, `server.applyStatusWebhook`, `server.recordCallIdem`, `server.contactSaveFailed`, `server.apiContactSaved` | 09-23 01-12 + 03-01 |
| Panel header block ×6                                                                                                                                                                | `views.panelHead`                                                                                                                                                                                              | 09-23 15-58         |
| `.wp-error` node / conditional wrapper / identity line ×8 sites                                                                                                                      | `views.errorBanner`, `views.panelError`, `views.identityLine`                                                                                                                                                  | 09-24 16-56         |

## Sweep log (append-only)

- 2026-09-18 22:55 — `-t 1`, 9 groups: 2 extracted, 7 accepted (§a.5).
- 2026-09-23 01:12 — `-t 1`, 13 groups + bonus: 9 helpers extracted, 6 accepted.
- 2026-09-23 03:01 — `-t 3`, 5 groups: `apiContactSaved` extracted, 3 accepted; re-run shown 3 = the accepted trio.
- 2026-09-23 15:58 — `-t 2`, 16 groups: `panelHead` extracted (6 sites), 15 accepted.
- 2026-09-24 16:56 — `-t 2`, 16 groups: `errorBanner`/`panelError`/`identityLine` extracted (8 sites), 13 accepted; re-run fully attributed.
- 2026-09-27 — `-t 2`, 15 groups, all verified at HEAD against this table: 0 new, 0 extracted, 15 attributed (EmptyState ruling rationale updated). Zero harmful duplication; re-run at the same flags confirmed the group set.

## Open owner calls (do not resolve unilaterally)

1. Baseline ratification: `-t 3` as the ritual vs `-t 2` deep sweeps.
2. `settingsRow`: build the component or accept the dt/dd rows permanently (surfaced in every sweep to date).
3. Ratify this registry as the ONE acceptance home (proposed 03-01 f.16, re-raised 15-58 e.4 + 18-05 e.6).
