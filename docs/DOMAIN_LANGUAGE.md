# Domain language — webphone's ubiquitous vocabulary

> STATUS: DRAFT created 2026-09-23 by the pareto execution session,
> pending owner ratification (ROADMAP open question g2: create vs
> record accepted absence). If the owner prefers absence, delete this
> file — nothing links to it except ROADMAP/TODO rows.

One term, one meaning, everywhere (code, docs, UI copy, commits).
Terms the SERVER owns are Go-stable names; terms the USER sees are the
en/de UI strings' English anchors.

## The product and its surfaces

- **webphone** — the single binary: SIP call island + server-rendered
  tabs (Messages, Fax, Voicemail, History, Contacts, Settings) on one
  page.
- **the island** — the always-loaded SIP.js call surface OUTSIDE the
  tab region; it never unloads, so calls survive tab switches.
- **tab region** (`#tab-content`) — the swapped area; navigation
  morphs partials in place, never a full page load after boot.
- **shell** — the page frame around both: nav, badges, toasts, error
  banner. shell.js works even when island scripts fail.

## Identity and sessions

- **extension** — the PBX phone identity ("1001"): the product's ONLY
  user concept. No user table exists by design (split-brain refusal of
  the cqrs-htmx `setup` bundle).
- **directory password** — the PBX FreeSWITCH directory credential;
  proven by the island's SIP REGISTER, stored in the session row for
  the `/phone-api` proxy.
- **session** — the cookie-backed tab session; slides on activity,
  capped absolutely. `requireSession` self-gates handlers.
- **identity / DID** — the operator-configured `identities` map:
  which presented number an extension sends as. Feeds the self-send
  warning.

## Telephony

- **call** — one island call leg. States: ringing/established/ending
  (the `data-state` chip).
- **CDR** — call detail record, served by the PBX's API; History's
  remote source. History is deliberately hybrid (local log + CDR).
- **DTMF** — in-call digits (`application/dtmf-relay`, `Signal=`
  form).
- **REFER / transfer** — server-side executed by FreeSWITCH; the
  browser sends REFER and reads the NOTIFY sipfrag verdict.
- **TURN REST credentials** — short-lived TURN passwords derived from
  the shared `turn_rest.secret` (HMAC); served via `/config.js`.

## Messaging, fax, voicemail

- **thread** — one remote number's message conversation (SMS+MMS).
- **bubble** — one message row in a thread; a **failed bubble** keeps
  the persisted provider reason + kind and retries only where
  retryable.
- **verdict hook** — the provider's post-send verdict
  (`/hooks/{message,fax}/status`): delivered | failed + reason,
  keyed by **provider_ref**.
- **fax job** — one outbound fax attempt with pages, PDF document,
  provider status (transmitted/failed).
- **voicemail summary / messages** — read-only views over the PBX's
  mailbox API (`*97`).

## The gateway seam

- **gateway** — the outbound transport: **loopback** (dev; everything
  "succeeds") or **webhook** (multipart POST + Bearer secret to a
  bridge; `{"provider_ref"}` receipt). One seam, two modes.
- **bridge** — the stack-side `telnyx-webhooks` service that stands
  between webphone's webhook gateway and Telnyx. Its error strings are
  a wire contract (rendered verbatim).
- **refusal vs outage** — a provider 4xx ANSWER is user feedback (422
  with the provider's reason); transport failures and provider 5xx
  answers are system-side (502, retryable).

## Stores and their invariants

- **one-home helpers** — dedup-train extractions with exactly one
  definition and one contract test (listRows, pbx.do, makeSession,
  requireMultipartTo, applyStatusWebhook, …).
- **owner scoping** — every store row belongs to an extension; every
  query is extension-scoped.
- **personal contacts** — ONE home: the per-extension SQLite store;
  mutations answer bare 204 and the LIST is the only id source.
- **shared contacts** — operator config, read-only for users.
- **idem key** — the in-memory 1h dedupe memory: `hooksIdem`
  (`<kind>/<provider_ref>`) and `callsIdem` (`<ext>:<uuid>`); only
  successes are recorded, failures stay retryable.

## Live updates

- **hub / SSE events** — per-extension server-sent events
  (`/events`): `threads`, `thread`, `fax`, `voicemail`, `contacts`,
  `connected`. Payload-less events are **nudges**: consumers re-fetch
  with their own credentials.
- **morph surface** — any live-updated region swaps via
  `hx-swap="morph:innerHTML"`; stateful nodes inside must carry
  stable ids (idiomorph persists by id).

## Integrations

- **CRM seam** — optional read-only enrichment (`crm.url`/`crm.token`)
  + idempotent post-call journaling; a dead CRM never breaks a page.
- **consuming stack** — nix-international-telephony: fronts the
  binary with TLS, proxies `/sip`, `/events`, `/phone-api`; pbx
  -artmann consumes the stack via a rev pin. Tri-repo chain.
