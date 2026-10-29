# CRM data-mesh train, Tier-0 closeout — full status update

**Snapshot:** 2026-10-29 21:31 CET. Claims verified by execution in this
session; every gate row below was run on the tree it names, not assumed.
**Scope of this report:** the Tier-0 tail of
[2026-10-29_19-44 closeout](2026-10-29_19-44_crm-data-mesh-train-closeout-status.md)
§f, i.e. the seven blocking/high-value items plus everything that fell out of
them. Nothing outside this train was researched.

| Repo | HEAD | State |
| --- | --- | --- |
| webphone | `366944d3` | clean, pushed, **CI `40942091650` green** (7m57s) |
| Kith CRM | `3590e01` | clean, **2 commits ahead of `origin/master`, NOT pushed** (owner question Q3 below) |

10 commits landed this session, 21 files, +622/−86.

---

## a) FULLY DONE

### 1. The detection-table row (was the audit's own broken promise)

`docs/error-contract.md` § "Detection column" promised at ship time that a new
quiet class gets a row. The context strip had shipped without one. Added:

- **CRM context absent** — why quiet: fail-open by design; a stale CRM (404), a
  revoked token (401) and a 5xx all degrade to NO markup, so a permanently
  broken seam is indistinguishable from a contact the CRM knows nothing about.
  Detection: `crm: context failed; no strip` DEBUG per fetch carrying the
  wrapped `crm.context_read`/`crm.context_decode` code; failures are not
  cached, so a healed CRM recovers on the next conversation open. The 401 arm
  logs `crm: context unauthorized; no strip` and caches absence 15 min; the
  404 arm caches absence silently, which IS the intended old-CRM handshake.
- Section intro corrected five → six classes; closing paragraph now says rows
  1 and 6 landed with their features (2–5 retroactive).

### 2. Dead fields deleted, and the read contract made structural

`crm.Context` carried `Name`, `Tags` and `Ref` with **no non-test reader**.

- All three removed. `Ref` is the interesting one: while it existed, a payload
  could have been trusted for identity; with it gone, strip redirection is
  *structurally* impossible rather than test-guarded. `toContext()` no longer
  takes a ref; the wire optionally carries producer identity fields that the
  consumer now provably ignores.
- New pin `TestEveryCRMContextFieldIsRead` (`internal/server/crm_context_test.go`):
  parses `crm.go` with `go/ast`, collects selectors taken off the `published`
  parameter of `crmContextView`, and reflects over `crm.Context`'s fields.
  **Verified it fails**: adding `Extra string` to the struct produced
  `crm.Context fields no surface reads … [Extra]`.
- The consumer fixture in `crm/context_test.go` deliberately KEEPS
  `id`/`name`/`tags`, so the forward-compatibility rule ("unknown fields
  tolerated") is now a test instead of a promise. The asymmetry is documented
  in the plan doc §4 and in AGENTS.

### 3. A real defect found while writing the guard pin (the biggest catch)

The context strip guarded `remote.Network() != domain.NetworkSMS`, but the
identity feed called the CRM for **every** thread's raw remote. So a Matrix
thread sent `@ada:example.org` to `/api/contacts/by-phone`, which Kith answers
with **400** (`normalizePhoneDigits` empty). The resolver does not cache
failures, so it repeated on every single render.

- New `domain.Address.IsPhoneShaped()` (sms | signal): the ONE home of "a
  phone-number index can match this", instead of per-consumer network
  enumerations.
- New `server.crmNumberFor(remote)`; the three thread call sites
  (`panels.go` ×2, `notifier.go` ×1) now feed `crmNumbers` through it, so the
  existing `""`-skip does the filtering.
- `crmContextFor` now uses the same predicate, which also fixed a
  user-invisible asymmetry: **Signal** conversations got a linked name and no
  context for no explainable reason. They are phone numbers, so they are
  eligible now.
- Pins: `TestAddressIsPhoneShaped` (4 networks + zero value),
  `TestCRMContextRefusesNonPhoneShapedRemotes` (matrix/session ⇒ nil **and zero
  CRM lookups**; sms/signal ⇒ a strip and exactly one lookup),
  `TestCRMNumberForAnswersOnlyPhoneShapedRemotes`, and
  `TestMatrixThreadNeverReachesTheCRM` (real partial path, seeded via
  `Messaging.Receive`, asserts the stub saw exactly the phone-shaped remote).
  **Verified it fails without the fix**: reverting one call site produced
  `the matrix address reached the CRM lookup: [@ada:example.org +493012345678]`.

### 4. Visual proof — and the real bug it caught

`scripts/ui-capture.py` had no CRM mode, which is why the audit could only say
"the strip probably looks right". Now:

- `--crm` boots a bearer-gated fake Kith machine API in a daemon thread
  (`by-phone`, `/context`, `calls`) beside the app, with `crm.url` +
  `crm.token` + `crm.contact_url` in the boot config.
- A new `crm-context` surface (same thread path) with **three DOM pins**:
  `.wp-crm-link` + "Ada Lovelace", `#wp-crm-context` + company, and the deal
  title. Marker normalisation was generalised to accept a string, a
  `(selector, copy)` tuple, or a LIST of pins.
- Running `--surface crm-context` **without** `--crm` exits 2 with an
  explanatory message, so the proof cannot silently degrade into shooting the
  no-CRM product.
- Full matrix re-run with `--crm`: **26 shots**, every assertion passing,
  layout proof included.

**What it caught:** the first shot rendered
`Ada Lovelace(Open in the CRM)` underlined INSIDE the 24px conversation
heading. The link's screen-reader suffix used a class I invented,
`wp-sr-only`; the owned utility is `.sr-only` (app.css), so the "hidden" text
rendered visibly and cramped. Fixed by moving the destination into
`aria-label` (unstyleable, cannot leak), pinned by
`TestCRMNameLinkIsTheOneLinkHome` including a negative pin that no
parenthesised suffix is visible. **No gate in this repo catches a wrong class
name; only a picture does.**

### 5. AA contrast measured (was: "not measured")

Computed from the live tokens, sRGB luminance, both themes:

| pair | ratio | verdict |
| --- | --- | --- |
| light `--muted` on `--surface` | 5.85:1 | AA |
| light `--muted` on `--bg` | 5.17:1 | AA (worst case) |
| light `--text` on `--surface` | 17.84:1 | AA |
| dark `--muted` on `--surface` | 6.89:1 | AA |
| dark `--muted` on `--bg` | 7.46:1 | AA |
| dark `--text` on `--surface` | 15.00:1 | AA |

`.wp-crm-key` / `.wp-crm-muted` clear 4.5:1 at their ~12px size in both themes.

### 6. `FEATURES.md` + `TODO_LIST.md` brought to the one-home rule

- FEATURES: the CRM row was rewritten (identity in the seam, vocabulary
  translation, phone-shaped-only) and a new row added for **deep links +
  contact context**, naming the new-tab rationale, the 2 min/15 min cache
  windows, and the absence-is-a-state contract.
- TODO_LIST: new dated section "CRM data-mesh train (2026-10-29) — deferred
  tail" listing the remaining edges; items 1–2 (visual proof, contrast) were
  executed and therefore **deleted**, not struck through.

### 7. Kith side documented (not pushed)

- `FEATURES.md` machine-API row, `README.md` usage example, `CHANGELOG.md`
  `[Unreleased]/Added` entry for `GET /api/contacts/{id}/context`.
- Discovered `/docs` is generated from the **EventCatalog** and covers no
  machine-API REST route, so no generated-wire entry was owed. That
  correction is written into TODO_LIST rather than left as a phantom task.
- Kith's own `dprint` owns markdown, so the FEATURES table was formatted and
  the em dash in my row replaced with a colon; `dprint check` exit 0,
  `go build ./...` and `internal/machineapi` tests green.

### 8. A KVM-gated VM test race found BY the full gate and fixed

`nix flake check` failed on `webphone-error-drill` twice, on a boot path this
session never touched: the check asserted
`systemctl show webphone-drill -p ExecMainStatus` = 1 **after** five
`journalctl` greps, against a unit running `Restart=always -p RestartSec=1`,
and systemd resets `ExecMainStatus` on every new cycle. Fixed by asserting the
exit taxonomy FIRST and wrapping it in `wait_until_succeeds` (every failed
cycle re-exposes it). Re-ran the check: green. Rule recorded in
`docs/lessons.md` § Nix: with an auto-restarting unit, per-cycle properties
must be RETRIED, never single-shot — and a red `nix flake check` is not
automatically your regression.

### 9. Gates actually run (frozen tree, 20:40)

| Gate | Result |
| --- | --- |
| `go build` + `go vet ./...` | clean |
| `go test -count=1 ./...` | **31/31 packages ok** |
| `tsc -p . --noEmit` | clean |
| island `node --test` | **283 pass, 0 fail** |
| `nix fmt` | 0 changed |
| 5 fast flake checks | green |
| **`nix flake check`** | **all checks passed** (KVM-gated backup, backup-drill, error-drill, module) |
| `art-dupl check -t 1` | 0 new clones, baseline 86 |
| `buildflow` | every step green (exit 0) |
| smoke | 49 + 4 + 8 + 9 + 9 passed, **0 failed** |
| `error-context.py --self-test` | 37/37 |
| visual matrix `--crm` | 26/26 shots, all DOM pins |
| CI `40942091650` | success |

---

## b) PARTIALLY DONE

1. **The producer's identity fields.** Kith publishes `id`, `name`, `email`,
   `tags`; webphone decodes none of them. I decided to KEEP them on the
   producer (a REST representation without its id is poor practice, and they
   are proven-ignored by the consumer fixture) and document the asymmetry —
   but that is a ruling, not a measurement. If the owner's data-mesh reading is
   "publish only what is consumed", this should be trimmed. **Half-decided.**
2. **`buildflow` warnings.** Exit 0, but three known noises persist:
   `.buildflow-baseline.json` has 78 accepted-noise entries matching nothing
   (third run in a row), six timing regressions (+85% on `test-race`, +217% on
   `golangci-lint`) from host load, and the `nix-build-verify` 86% failure-rate
   warning. Untouched by me, and now *measured* rather than folklore.
3. **Kith commit hygiene.** My three doc files landed as
   `chore: auto-commit …` (the daemon won the race twice); I authored one good
   message but the working tree was already clean. History reads worse than the
   change.
4. **The `crm-context` harness surface is undocumented in `docs/dom-contract.md`
   verification terms** — the *ids* are documented, but nothing says the visual
   harness asserts them. AGENTS carries the command; the doc contract does not
   cross-reference it.
5. **Em dash sweep.** I used one em dash in a Nix comment by accident and
   caught it; other new prose rows (Kith FEATURES) needed the same correction.
   I did not sweep the files I touched for pre-existing em dashes.

---

## c) NOT STARTED (all carried, all listed in TODO_LIST)

1. Two surfaces still render plain names: voicemail's `vmCaller` (mixed
   name/number text, must be split before wrapping) and the drafts panel's
   `contactLabel`.
2. "Add to CRM" affordance for a caller the CRM does not know.
3. Batch `by-phone?numbers=` identity product (owner-side, Kith).
4. Contract snapshot test pinning our exact `/context` request + parse against
   the producer's published shape (the vocabulary-break class mitigation).
5. The Kith push itself (blocked on Q3).
6. Pushing the CRM-context commits as a clean branch (the cherry-pick option of
   Q3) — not evaluated because it depends on the answer.

---

## d) TOTALLY FUCKED UP (honest list, ordered by cost)

1. **I shipped a made-up CSS class in the previous session** (`wp-sr-only`) and
   every gate was green. Only a screenshot found it. That means, by the repo's
   own ruling, the feature was not done when it was declared done — and the
   proof I now have took a harness feature that should have been built with the
   strip.
2. **I let the auto-commit daemon eat five narrative commits.** Six of this
   session's ten commits are `chore: auto-commit N changed file(s) (heuristic)`.
   AGENTS literally documents the fix ("stage + commit in ONE tool call… the
   daemon is adversarial to in-flight edits"); I still edited continuously and
   committed at phase boundaries, so the daemon won repeatedly. History value
   lost, not code.
3. **First `--surface crm-context --crm` run crashed** on a marker-unpacking
   bug (`ValueError: not enough values to unpack`) because I generalised the
   marker shape for `crm-context` without re-testing the *string* surfaces I
   had just broken. Caught by the very next run, but it is the same class of
   mistake as #1: changing a shared shape and only testing the new consumer.
4. **I burned two `nix flake check` runs (~8 min each) before reading the
   failing assertion's mechanics** — the second failure was fully diagnosable
   from the first log. Reading first would have saved one cycle.
5. **`nix build` of the error drill reused the cached red store path** after a
   python patch that silently failed an assertion (count mismatch, wrong
   indentation), so I briefly "verified" a fix that had not been applied. Only
   the traceback revealed it. Lesson: after a scripted edit, grep the file for
   the new text before running a 5-minute gate.
6. **Two `grep -v` filters hid the build/test exits** (`exit status 1` with no
   message) and I re-ran commands instead of reading raw output — the AGENTS
   trap about filtering test output by leading spaces is still biting.
7. Lower-severity: I pushed before recording `buildflow`'s verdict in the plan
   doc (order of operations sloppy), and my first Kith FEATURES row was not
   `dprint`-canonical, which would have red-ed their tree for a future run.

---

## e) WHAT WE SHOULD IMPROVE (process, not features)

1. **Build the proof harness WITH the feature, not after.** The rule the repo
   already has ("a screenshot proves a surface RENDERS") needs to become
   "a UI affordance ships with its assertion" — the `--crm` flag was 200 lines
   that should have existed before the strip was called done.
2. **Ban invented class names structurally.** A cheap, high-value gate: for the
   `wp-` namespace only, every class literal in `.templ` must appear in
   `app.css` **or** be listed in a small allowlist of JS/test-only hooks. I
   measured the shape: 158 `wp-` classes, 33 without CSS rules, of which
   exactly ONE (`wp-sr-only`) was a defect. An allowlist keeps the gate honest
   and would have caught it at commit time.
3. **Stop treating the daemon as a scheduling detail.** Either commit every
   coherent change immediately in the same tool call that verifies it, or work
   on a branch. Six heuristic commits is the observable cost of doing neither.
4. **Read the failing assertion before re-running a gate.** Add to the
   `docs/lessons.md` Nix section (done for the drill) and generalise: "a red
   gate on an untouched path is a test-mechanics question first".
5. **`buildflow`'s accepted-noise baseline should be refreshed or retired.**
   Third consecutive run reporting 78 unmatched entries is the same rot the
   branching-flow baseline had before it was retired on 2026-10-28.
6. **Measure before writing docs claims.** The 19:44 report labelled two
   claims "speculation" — both took under 30 minutes to measure. Speculation
   labels should be a to-do, not a footnote.
7. **Cross-repo doc formatting is part of the change.** Kith's dprint owns
   markdown; a doc edit that is not `dprint`-checked is a latent red tree.

---

## f) UP TO 50 THINGS TO GET DONE NEXT

Ordered roughly by value/cost. Nothing here is speculative research; each row
came out of this session's work.

**Correctness / contracts**
1. Contract snapshot test: pin the exact `/context` request and the parse
   against Kith's published shape (both repos' fixtures as one artifact).
2. Vocabulary-bridge test on the PRODUCER side: assert Kith rejects our OLD
   value (`answered`) and accepts the current set — a cross-repo pin, so the
   closed-set drift fails in CI.
3. Decide (owner) whether Kith keeps publishing `id`/`name`/`email`/`tags`
   while its only consumer ignores them.
4. Add a `crm.Context`-fields↔`views.CRMContext` mapping test so a new wire
   fact cannot land without a view slot.
5. Assert `crmNumberFor` is the ONLY path from a thread remote to the CRM (a
   source scan, so a new call site cannot bypass the filter).
6. Pin `crmContextFor`'s behavior when `deps.CRM` is nil (documented nil-safe,
   currently only exercised implicitly).
7. Error-code registry: confirm no new codes were needed for the phone-shaped
   refusal (it returns nil, no error) and record that reasoning.

**Verification debt**
8. Refresh or retire `.buildflow-baseline.json` (78 unmatched accepted-noise
   entries, 3 consecutive runs).
9. Investigate `nix-build-verify`'s 86% failure history (19/22) — either fix or
   exclude with a written reason.
10. Re-run the timing-regression comparison on an idle host to separate load
    noise from real regressions (`test-race` +85%).
11. Add the `crm-context` surface to whatever list the next visual pass uses as
    its "expected shots" baseline (currently `shots >= 22`, which would let a
    missing CRM shot pass unnoticed).
12. Record the AA measurement as a repeatable script (the numbers live in a
    status doc today; a 20-line script would make it a gate).

**UI completeness**
13. Voicemail `vmCaller`: split name vs number, then link the name.
14. Drafts panel `contactLabel`: link it.
15. "Add to CRM" affordance for misses (link into the CRM create form; still
    no browser→CRM traffic).
16. Show the CRM link on the History transcript SUMMARY or explicitly document
    why not (a link inside `<summary>` steals the toggle click) — today it is
    silent.
17. Consider a CRM badge on list rows so a linked name is not mistaken for a
    local contact.
18. Empty/degraded state for the strip when the CRM is known but publishes
    nothing (today: nothing at all; is that right for a *known* contact?).
19. German copy review of the eight `crm.*` keys (only English was eyeballed).
20. Dark-mode review of the dotted underline (contrast passes; the 1px dotted
    border may read as noise at 12px).

**Producer (Kith)**
21. Push the two local commits once Q3 is answered.
22. Batch `by-phone?numbers=` product for lists.
23. Decide whether `/context` should include `last_activity` for contacts with
    no journal rows (currently omitted vs zero, undocumented).
24. Document the `/context` cache windows (2 min / 15 min absence / 256 cap) in
    the README integration section, not only in webphone's plan doc.
25. Add a `/context` example to `docs/ops/RESTORE-DRILL.md` if that drill is
    meant to prove the phone integration end to end.
26. Kith: a machine-API OpenAPI (or a hand-written wire table) since `/docs`
    covers only the EventCatalog.

**Harness / tooling**
27. Add `--crm` to the AGENTS smoke section too (only the visual section
    carries it).
28. Make `ui-capture.py` fail when `--crm` is passed but no shot pins the link
    (guard against a future surface rename silently dropping the proof).
29. Give the fake CRM a second contact with NO context, and shoot the absence
    case (proves the "no strip" contract visually, not just in Go tests).
30. Seed a Matrix thread in the harness so the phone-shaped filter is visible
    in the product, not only in tests.
31. Extract the fake CRM into `scripts/fake_crm.py` if a second consumer
    appears (today it is 60 lines in the harness; not worth splitting yet).
32. `--surface` typo guard exists; extend the same loudness to a typo'd marker
    key (unknown surface name in `SURFACE_MARKERS` would `KeyError`).

**Process**
33. Commit per coherent change in the same tool call as its verification (stop
    losing messages to the daemon).
34. Add a `docs/status` snapshot-rot header convention check (the 19:44 report's
    push claims were stale within an hour).
35. Sweep pre-existing em dashes in files touched by a change (Kith FEATURES
    does it widely).
36. `dprint check` for Kith markdown belongs in any cross-repo doc edit
    checklist.
37. Run `nix flake check` at TRAIN START, not only at end (it surfaced a race
    that had nothing to do with this work).
38. After any scripted edit, grep for the new text before running a long gate.
39. Never pipe gate output through a filter that can eat the failure line.
40. Record gate verdicts in the plan doc BEFORE pushing.

**Owner decisions / blocked**
41. Q1: do email/SMS/fax facts reach the CRM timeline?
42. Q2: may the strip become actionable (complete a task / move a deal)?
43. Q3: how is Kith pushed (whole branch / cherry-pick / leave it)?
44. Decide whether CRM context should also appear on History call detail
    (currently defaulted NO).
45. Decide whether the phone may ever WRITE to the CRM beyond call metadata +
    notes.
46. Decide whether `crm.contact_url` stays a separate knob or is derived from a
    CRM vhost setting.

**Housekeeping**
47. `docs/dom-contract.md`: cross-reference the visual harness assertion for
    the optional CRM ids.
48. AGENTS: the CRM bullet is long; consider moving the vocabulary table into
    `docs/error-contract.md` or a `docs/crm-contract.md` (rule vs evidence).
49. Keep `ui-shots/` untracked (it is gitignored) but decide whether the two
    CRM shots should live in `docs/` as proof artifacts.
50. Re-read the 19:44 report's "Verified by reading only" list — three entries
    are now measurable and two were never re-verified.

---

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

**Q1 — Do message bodies ever reach the CRM?**
Today only call metadata (direction/number/seconds/outcome) crosses. Sending
SMS/voicemail/fax *content* would make the phone a producer of the CRM's
relationship history. That is a privacy and noise decision I cannot infer from
the code: the seam is built to carry it, the product does not.

**Q2 — May the strip become actionable?**
Right now it is read-only by construction, and every design statement in this
train says the phone never mutates CRM state. Completing a task or moving a
deal from the phone while a caller is on the line is obviously useful and would
flip the ownership rule this train is built on. I cannot decide which bar wins.

**Q3 — Who pushes Kith, and how?**
`origin/master` is now `d57e00f`; local `master` is 2 commits ahead with my
CRM-context docs. During this session the branch's earlier shape (16 commits of
mixed claim/outreach work from another session, with a red `internal/catalog`)
was rewritten upstream, so my commits are a clean fast-forward onto the current
tip. Push as-is, cherry-pick onto a fresh branch, or leave it to that session —
this is a shared-tree policy call, not a technical one.

---

## h) Verification vs speculation

**Verified by execution:** every gate row in §a/9; each new test's FAILURE mode
(reverting the fix for the matrix lookup; adding a dummy field for the
reflective scan); the 26-shot matrix; the AA numbers (computed from the live
tokens); Kith's `dprint check`, `go build`, `machineapi` suite; `git ls-remote`
for push state; `gh run view` for CI.

**Verified by reading only:** Kith's `GetContact`/`FindDeals`/`FindTasks`
shapes (no new test of mine exercises them beyond the seeded cases); that
`/docs` covers no machine-API route (grepped the catalog package; did not boot
the docs server and click through).

**Speculation, labelled:** that the daemon's history rewrite of the other
session's commits is intentional (I only observed the resulting state); that
the six timing "regressions" are host load (the workload was clearly loaded,
but I did not re-measure idle).

**Not checked at all:** the stack's browser E2E against the aria-label change
(owed by the release-runbook rule for markup changes — the stack tree was not
touched this session); the KVM backup VM test's own history beyond this run;
whether any consumer other than webphone reads Kith's `/context`.
