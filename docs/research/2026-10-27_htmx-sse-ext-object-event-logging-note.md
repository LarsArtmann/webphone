# Upstream note draft: htmx SSE extension logs "[object Event]" on every connection error

**Status: DRAFT — not filed.** Filing is an owner action (verify-before-filing gate
passed on the diagnosis below; wording should go through github-voice before filing).
Target: the htmx SSE extension shipped by cqrs-htmx (`extensions/sse.min.js`), which
mirrors upstream htmx `ext/sse.js`.

## Verified diagnosis (source read 2026-10-27, cqrs-htmx v4.14.1)

The error chain, with the exact lines:

1. `extensions/sse.min.js:112-113` (the `EventSource.onerror` handler):

   ```js
   s.onerror = function (e) {
       g.triggerErrorEvent(r, "htmx:sseError", { error: e, source: s });
   ```

   `e` here is the raw `Event` object the browser hands `onerror`.

2. `triggerErrorEvent(elt, eventName, detail)` (htmx core, minified name `fe`)
   forwards the detail into the event pipeline with `error: t` preserved.

3. htmx core's event function (minified `ae`) stringifies it:

   ```js
   if (n.error) {
       H(n.error + (n.target ? ", " + n.target : ""));
   ```

4. `H` is `console.error(e)`.

Net effect: **every SSE connection error prints exactly `console.error("[object Event]")`
to the operator's console — no event name, no source URL, no state.** During a deploy
restart window (webphone releases `/events` instantly by design) this fires several
times per tab, which is how the 2026-10-27 pbx.artmann.tech incident triage ran into it.

## Suggested upstream fix (one line, htmx core or the ext)

Stringify something readable instead of the raw event, e.g. in `ae`:

```js
H(typeof n.error === "string" ? n.error : (n.error && n.error.type ? "htmx error event: " + n.error.type : String(n.error)));
```

or, scoped to the ext, pass a readable payload:

```js
g.triggerErrorEvent(r, "htmx:sseError", { error: e, source: s, url: r });
```

The `source`/URL in the detail would let downstream handlers log which feed dropped.

## What webphone does meanwhile (no served-bundle patch)

The island adds its own readable `htmx:sseError` log line (failure counter + retry
message) — island-side only; the ext bundle stays upstream-owned and unpatched.

## Relationship to the incident

`docs/status/2026-10-27_15-57_sse-502-incident-triage-and-error-surface-ruling.md`
records the console evidence; this note pins the vendor-side cause. The recovery loop
itself (CLOSED EventSource recreate with 500ms→64s backoff, reset on open, ext lines
~113-124) is CORRECT and is not part of any complaint.
