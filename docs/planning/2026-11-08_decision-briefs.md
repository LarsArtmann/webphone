# Decision briefs — four open calls, one recommendation each (2026-11-08)

**Status:** BRIEFS (recommendation + the other branch's cost) · Feeds the
[owner sitting pack](2026-11-07_21-30_owner-sitting-pack.md): each brief
expands one pack row with the full consequence set. The pack holds the
one-line version; these are the reasoning.

## Brief 1 — Paperless seam shape: Settings row vs per-fax marker (pack B-side, T31)

**The state:** `paperless.url`/`token` build the fax archiver with NO Settings
row and NO per-fax archive marker. A deployment that believes it archives
learns otherwise only from the journal (`paperless` debug lines). This is the
capability-declaration class: the fax tab's own compose form was removed for
exactly this shape of dishonesty (2026-10-30), and the seam's VISIBILITY got
skipped.

**Options:**

1. **Settings row alone** (recommended): one conditional row in Settings
   ("Archive: Paperless-ngx, on/off") reading the same boot-time flag the
   view props already carry (`FaxPanelProps`-style). Cost: ~20 LOC view +
   i18n keys + one pin. The operator learns the state in the place they
   already check ASR/CRM/push states — ONE home per fact, no new per-object
   state.
2. **Per-fax archive marker**: every fax row in the tab carries an
   "archived/archiving/not-archived" state, fed by the fire-and-forget
   archiver's outcome. Cost: the archiver is fire-and-forget AFTER persist —
   correlating outcome back to the fax row needs a new readmodels field or an
   event round-trip, plus SSE morph churn on every archive attempt. ~2 session
   tranches, and it introduces a SECOND place where paperless state lives.
3. **Both**: the marker's correlation cost for a fact nobody asked for yet.

**Recommendation: option 1**, with option 2's trigger written down ("the
owner asks 'was THAT fax archived?' more than once"). The Settings row closes
the belief-mismatch harm (the audit's actual finding); per-fax markers are a
workflow feature, not a capability declaration.

**Risk of the other branch (option 2 alone):** the invisible-seam class
persists at the SETTINGS level (the operator still cannot see the seam is on
without inferring it from rows), and the correlation machinery is built
before anyone demonstrates the need.

## Brief 2 — F1 ink direction: ratify what shipped **[shipped 2026-11-07]**

The UX train flipped dark-theme `--on-accent` to dark ink (`#0b0d11`) in
BOTH mirrored token blocks and rebuilt login.css. Measured on the day:
3.21:1 → 6.06:1 at rest, 1.87:1 → 10.42:1 on hover — the product's only
WCAG-AA failure, gone on its five most-read surfaces (primary Call/Send
buttons, every outbound bubble, failed bubbles).

**Recommendation: ratify.** The review's assumption became the shipped
reality; the only open question was whether the owner prefers the light-ink
aesthetic enough to accept a documented AA failure. The consequence of
reversing: re-introducing the only accessibility failure in the product, on
the surfaces every user touches first.

**Cost of reversal:** 2-line diff + contrast re-measure + re-shot, trivially
reversible — but it must then be RECORDED as a deliberate AA exception with
the owner's name on it, not a regression.

## Brief 3 — F19 signed-out shape: ratify what shipped **[shipped 2026-11-07]**

The review's IA finding: an anonymous visitor "sees the full desk with dead
tabs — the one IA miss"; the correct shape is a door (sign-in surfaces only).
The train shipped it: the signed-out page drops the tab row and centers the
login.

**Recommendation: ratify.** The dead-tabs desk was the review's one structural
IA miss; the door is also the passkey-mode's natural shape (the ceremony
page is already a standalone surface).

**Cost of reversal:** anonymous visitors regain navigation affordances that
all answer 401 — the anti-enumeration gate makes them uniform, so the desk
would be uniformly useless: worse than the door at zero benefit.

## Brief 4 — The attended-transfer question: keep blind, do not build the button **[M19 NO-GO stands]**

**The state:** the island ships BLIND transfer only (REFER, NOTIFY sipfrag
verdict parsed). Attended transfer (consult → hold → "Bridge to held" via
REFER-with-Replaces) is NO-GO since the 2026-10-27 spike: the stack's
dialplan has neither a Replaces concept nor an attended-bridge extension,
and no FS lab exists on this host to verify the bridge. The failure mode is
SILENT — a REFER the dialplan cannot bridge gets 200-OK'd and dropped (the
DTMF class), so shipping the button against an unverified bridge puts a lie
in the dock. The honesty contract forbids exactly that.

**Options:**

1. **Keep blind, no attended button** (recommended): the spike's GO-day
   trigger stands (a stack VM test with two registered loopback legs, an
   attended REFER with Replaces, asserting bridged audio + both BYEs). Until
   that exists, the button cannot be honest.
2. **Ship the button with a warning**: the warn would be a permanent lie
   ("this may silently do nothing") on a surface the operator uses mid-call —
   the worst moment for an unreliable affordance.
3. **Remove the blind button too**: blind transfer WORKS and is verified
   (the stack E2E exercises it); removing a working capability because its
   sibling is deferred is scopeless.

**Recommendation: option 1.** The consult-flow sketch (F113) is carried
verbatim in the spike doc — when the stack grows the lab scenario, the island
delta is hold + the existing transfer pending path, no new state machine.
