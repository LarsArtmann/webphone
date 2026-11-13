# Verification verdict — the htmx `[object Event]` filing is a NO (2026-11-08)

**Gate:** verify-before-filing (mandatory before drafting to a repo we do not
own) · **Target considered:** `bigskysoftware/htmx` (the SSE extension) ·
**Verdict: DO NOT FILE — the premise dissolved under source inspection.**

## What the 2026-10-27 triage observed

The island's console showed `[object Event]` spam during SSE drops; the
triage (row 5 of its §f table) replaced it locally with a readable
`htmx:sseError` handler line and queued "upstream the ext patch to cqrs-htmx
(verify-before-filing applies)" — note the ext was ALWAYS cqrs-htmx's, not
bigskysoftware's: the served bundle is `cqrshtmx.HTMXExtensionsHandler`
(server.go:539) from the library, and the triage's row 45 says "check
cqrs-htmx issues first, verify the ext path".

## What this session verified at source

1. **The served extension is the cqrs-htmx fork**, not upstream htmx:
   `internal/server/server.go:539` serves
   `cqrshtmx.HTMXExtensionsHandler(cqrshtmx.HTMXExtSSE, cqrshtmx.HTMXExtIdiomorph)`.
2. **The current fork has NO console logging at all.** The forked
   `sse.min.js` at the pin (cqrs-htmx v4.16.0, module cache) contains no
   `console.*` call and passes `{ error: e, source: s }` as event detail —
   3,886 bytes, grep-checked. There is nothing left to stringify.
3. **Upstream htmx's sse.js has no console logging either.**
   `bigskysoftware/htmx-extensions` `src/sse/sse.js:199-217` (fetched today):
   the onerror handler only `triggerErrorEvent(elt, 'htmx:sseError',
   { error: err, source })` and manages the reconnect ladder. No
   `console.*`, no `logError` anywhere in the file. (The extension now lives
   in `bigskysoftware/htmx-extensions`, not `bigskysoftware/htmx` — both
   checked.)

## Conclusion

The `[object Event]` stringification existed in an EARLIER cqrs-htmx fork
generation (the version running during the 2026-10-27 incident) and is gone
at the current pin — either fixed in the fork or never in the paths we
checked. Filing against `bigskysoftware` would have been the classic
premise-error issue: professional-looking, evidence of OUR incident, and
wrong about THEIR source. The residual question lives in the owner's own
repo: whether the older fork's console line was already reported/fixed
upstream there — `gh issue list -R LarsArtmann/cqrs-htmx` before anything
else.

**Disposition of triage row 45:** resolved as a NO (with this note as the
record). The local readable handler (session.js) is the permanent fix; no
external filing is owed.
