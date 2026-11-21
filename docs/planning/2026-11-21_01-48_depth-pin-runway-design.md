# Depth-pin runway design — the next split, before a red budget forces it

> Source row: TODO_LIST § 2026-11-19 evening skills harvest, F1 extended
> by F4/F5 (architecture review #6). This note is the design the row
> asked for: what the NEXT carve of each near-ceiling package looks
> like, so the decision is never made mid-train with a red gate in
> hand. No code rides this note; each carve is its own train when its
> pin actually bites.

## Why now — the measured runway (2026-11-21, HEAD 752dd1cc)

Fresh measurements (the arch test's own definition: hand-written
non-test, non-templ Go lines; exports = top-level capitalized decls):

| Package             | Measured | Ceiling | Headroom | Notes                        |
| ------------------- | -------- | ------- | -------- | ---------------------------- |
| `internal/app`      | 1681     | 1700    | **19**   | nearest bite                 |
| `internal/config`   | 1487     | 1500 \* | **13**   | \* default, UNPINNED         |
| `internal/server`   | 5271     | 5300    | 29       | html hub residual, named pin |
| `internal/server/api` | 1844   | 1900    | 56       | carved JSON surface          |
| `internal/server/platform` | 1424 | 1500 \* | 76    | \* default, UNPINNED         |
| `internal/domain`   | 2282     | 2500    | 218      | exports axis: **162/180**    |
| `internal/cqrs`     | 5763     | 6000    | 237      | the ONE-journal world        |

The 2026-11-19/20 trains (DLQ decorator, export-legacy, note-join,
moduledrift pins) each consumed 20–60 lines of headroom in the top
five. Two more trains of that size and `app` or `config` reds the
budget gate mid-work — the exact forcing function this note retires.

## The standing valve (ratified here): pin-riding-with-reason

The 2026-11-21 note-join train needed +21 lines in `server` and raised
the pin in the SAME commit with the reason in the exemption table —
that is the standing valve for packages whose carve is RULED OUT or
not yet due: **the number rides WITH the feature, never as silent
drift, and the table comment names the feature.** The ratchet's
direction stays DOWN; riding up is a loan the next carve repays. This
ratifies what the 2026-11-18 line-ownership and 2026-11-21 note-join
commits already did, so a future session does not re-derive it.

## The next split per package

### `internal/app` (adapters extraction — the nearest bite)

app.go is the composition root's WIRING (the container literal, the
lifecycle, the health/probe assembly) plus a growing set of ADAPTERS
(type adapters the injector cannot express: `app.PasskeyRuntime`,
`app.Trace`, the DLQ decorator, `mappedUsers` joins). The seam: move
the adapters into `internal/app/adapters` (same module-internal
package tree, adapters own no wiring) — the injector doctrine already
forbids `MustInvoke*` outside the root, and adapters are the
framework-free half the doctrine keeps promising. Trigger: the pin
bites (next +20-line feature). Est. motion: ~300–400 LOC out.

### `internal/config` (validate-lane seams per section)

config.go carries the Load pipeline plus per-section validation that
grew with every seam (identities, asr, push, chat, dial). The seam:
one validate file per SECTION (`validate_telephony.go`,
`validate_chat.go`, …) — not a package split; config stays ONE package
(one struct, one loader) and the depth budget counts the package, so
this does NOT lower the number. The honest lever is the pin: config is
now 1487 — it needs a NAMED pin at 1600 with the reason "per-section
validation rides with each new seam; the split candidate is
section-seams, ruled in this note" — OR the validate-lane extraction
into `internal/config/validate` (a subpackage IS countable separately).
Ruling: extract the subpackage when the pin bites; the sections are
already named by the struct.

### `internal/server/api` (resource-family subpackages — conditional)

The 2026-11-12 carve created `server/api` as the JSON surface; it has
since grown resource families (contacts, transcripts, push, notes,
calls, latency, session). The next split is per-family subpackages
(`server/api/contacts`, …) — but ONLY conditionally: each family is
50–150 LOC and the shared harness (writeJSONBody, requireSession,
limiters) is the cohesion. Ruling: do NOT split by file count; split
when a family needs its own test-fixture mass (the depth-budget test's
own carve trigger). Until then, pin-ride.

### `internal/server/platform` (already file-shaped)

assets.go, metrics.go, trace.go, directory.go, configjs.go are
file-shaped families; the package is cohesive and UNPINNED at 1424.
Ruling: name the pin at 1500 with reason "file-shaped families; the
next family is a new FILE, and 1500 is the ceiling the depth default
already provides" — i.e., promote the default to a named pin so the
exemption table tells the story (unpinned defaults are invisible in
the table; F4's finding).

### `internal/domain` (the exports axis — annotate, don't split)

162/180 top-level exports. The axis is UNANNOTATED: nothing says what
the 181st export costs. Ruling recorded here: domain's exports are the
ubiquitous vocabulary — the cap exists to force a CONVERSATION when
the vocabulary grows, not to prevent it. The 181st export requires
either a machine-table home (a new transition table consolidating
exports) or a one-line reason in the exemption table's commit. LOC
(2282/2500) is not the binding axis; exports is.

### `internal/cqrs` (no action)

6000 named pin, 5763 measured, the ONE-journal world. The aggregate
files are already one-file-per-aggregate; headroom is healthy. The
reopen question (per-aggregate subpackages) died with the single-go.mod
ratification — do not re-litigate.

## What this note is NOT

- Not a carve train: no file moves ride this commit.
- Not a pin change: the numbers above are measurements, not edits; the
  app/config pin decisions land when their bites occur (or with the
  next train that needs the headroom, whichever comes first).
- Not re-opening the html hub question (ruled out 2026-11-12; only a
  real architecture change reopens it — the valve above is the
  standing answer meanwhile).

_Generated 2026-11-21 01:48 CET · measurements at HEAD 752dd1cc ·
source: architecture review #6 F1/F4/F5 + the 2026-11-21 note-join
train's own pin-ride._
