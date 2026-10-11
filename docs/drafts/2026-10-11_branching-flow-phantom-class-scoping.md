# PHANTOM_TYPE: scope to identity/transpose boundaries — non-identity plumbing params dominate the report

TL;DR: `phantom` flags fn-param pairs like `(baseURL, token)` and `(page, limit)` where a named type is churn, not safety — 13 of 14 rows in one webphone triage were non-identity plumbing. Ask: a class/pattern-level knob (art-dupl's `--disable-pattern` equivalent, or a severity split) so the valuable identity-boundary rows can gate without the plumbing noise.

branching-flow 0.6.4 (192031d), `phantom all .` against
LarsArtmann/webphone @ `85de30c`.

## Problem

The phantom analysis treats every same-underlying-type fn-param pair as
transpose risk. In webphone's 2026-10-11 triage, 14 rows surfaced and
13 were non-identity plumbing:

- `crm.NewClient(baseURL, token string)`, `asr.NewClient(baseURL,
  token, model string)`, `paperless.NewArchiver(baseURL, token …)` —
  single-caller constructors, config-validated both-or-neither at
  boot.
- `blob.Save(subDir, ext string)`, `pbx.ResolvePath(rest, rawQuery
  string)` — literal-arg plumbing, failure immediately visible.
- `Messages.UpdateOutboundStatus(… providerRef, failureKind,
  failureDetail string)` / `Faxes.UpdateStatus(… providerRef,
  errorText string, pages int)` — adjacent strings, but the failure
  vocabulary is documented at the domain field and the fix would touch
  webhook JSON DTOs + a partial unique index for zero behavioral gain.
- `ListMessagesPage(… page, limit int)` — mis-order degrades visibly
  and `page < 0` is guarded.

The one row that was WORTH fixing had a different signature: a closed
constant set (`sseEventThreads…` → defined `server.EventName` type;
`Publish(extension, eventName, html)` transpose no longer compiles).
That is the valuable class: identity-boundary / closed-vocabulary
params.

`.buildflow.yml` in webphone carries the standing triage note ("wants
phantom types for display fields … maximalist churn with zero safety
gain for non-identity data", 2026-09-18/09-22) and the tool is
skip_steps'd because of it.

## What already exists (checked)

- `//nolint:branching-flow` works per-finding (live-verified 2026-10-11:
  a flagged panic row disappears with the directive).
- `all --baseline` (Sarif) gates on NEW findings only.

Neither scopes a whole CLASS. Per-finding nolints on ~60 advisory rows
across a fleet is churn; the only class-level knob today is skipping
the whole tool, which also discards the valuable identity-boundary
rows.

## Ask

A class/pattern-level switch, whichever shape is cheapest:

1. `--disable-pattern`-style labels like art-dupl's (`--list-patterns`
   → suppress by label), e.g. `PHANTOM_TYPE_PLUMBING`; or
2. a severity split: identity/config-boundary params (constructors of
   http clients, wire seams) stay `warning`-gateable; single-caller
   plumbing drops to `info`.

That would let webphone un-skip branching-flow and gate the real
findings.

## Repro

```
git clone https://github.com/LarsArtmann/webphone && cd webphone
git checkout 85de30c
branching-flow all . --format markdown   # phantom: 14 rows, 13 plumbing
```

Triage record: `docs/status/2026-10-11_02-19_branching-flow-report-triage-status.md` (every row attributed).

💘 Generated with Crush
