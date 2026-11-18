# Status — line-ownership follow-up train (§f execution, mid-flight) + honest self-audit

> STATUS SNAPSHOT — claims verified as of **2026-11-18 20:03 CET**. CI/push
> claims rot fastest: at snapshot time all session work except the final
> daemon sweep (`6a654fb2`, the ui-capture fixture edit) is PUSHED
> (origin/main = `30ec09dc`), the CI verdicts on the newest shas were
> still PENDING, and one background binary build (`nix build .#webphone`)
> was still RUNNING uncollected. Author: the session executing the
> 19-21 report's §f list under the owner's "GET SHIT DONE — the WHOLE
> TODO LIST" directive.

## The one-paragraph verdict

Of the 19-21 report's 22 §f items, the owner-directed sweep is roughly
TWO-THIRDS executed and everything executed is verified: the SSO-only
identity fix is pinned (f.1), `Line.Display()` is consumed at real
surfaces with a real bug fixed on the way (f.2 — DID-less lines
rendered "1001 · 1001"), the account list leads active-first (f.11),
the Admin numbers inventory chips like the users inventory (f.13), a
new arch gate rejects dead exported model API (f.9 — first catch
unexported, not deleted), ci-verdict now BLAMES failing tests per file
(f.7 — live-verified against two real CI failures), the TODO rows
exist (f.5/f.6), language-across-switch is proven by test (f.17), the
domain diagram landed (f.18), and the contracts README was re-verified
(f.20). The heavy verification legs are MID-FLIGHT, not done: the
stack browser E2E (f.3) is scoped and one command away, the ui-capture
run (f.4, which also discharges f.12) has its fixture upgraded but
never ran, and the hot-reload sketch (f.16) is referenced by a TODO
row that does not exist yet — a dangling reference this session owns.

## a) FULLY DONE (verified this session)

1. **f.1 — `TestSSOOnlySessionResolvesIdentity`** (sso_api_test.go):
   the SSO-only whoami fix is pinned end-to-end — begin → callback →
   303 → resume resolves `display_name` "Lars" + one labeled line
   `1000/+17165550100` with `deps.UserAuth` nil. The latent-gap rule
   (fix ships WITH its regression test) is now honored retroactively.
2. **f.2 — `Line.Display()` consumed, dead-code status gone**: the
   Settings account rows and BOTH Admin inventories render
   `Name` precomputed server-side via `authmap.Line.Display()` — the
   hand-rolled templ `if did != ""` fallbacks are deleted (they
   existed in BOTH surfaces; the "one fallback rule" was actually
   three implementations plus a client mirror).
3. **A real UI defect fixed on the way**: a DID-less line rendered
   "1001 · 1001" (name + unconditional secondary) on Settings AND
   Admin since the 2026-11-18 train — the secondary extension now
   renders only when the name is not already the extension. Pinned by
   negative assertions (" · 1000", "1003 · 1003") plus a DID-less
   fixture line (intern@example.com, extension 1003).
4. **f.11 — active-first ordering**: `accountSection` stable-partitions
   the owned lines so the session's own line leads, mapping order
   preserved after it. Pinned by `TestSettingsAccountLeadsWithTheActiveLine`
   (switch to 1001, then assert the +…312 row precedes +…311).
5. **f.13 — Admin numbers inventory chips**: `views.AdminLine{Name,
   Extension}` + one shared `adminLineChip` templ component; the
   numbers rows render the same labeled chip the users inventory
   does, keeping the "no DID configured" operator marker. The
   `wp-admin-ext` badge shape is gone from that section only (class
   stays, other sections use it).
6. **f.9 — dead-export arch gate** (`internal/arch/model_api_test.go`):
   exported methods on the MODEL packages (authmap, domain) must have
   a production consumer OUTSIDE their own package — test files do
   not count, generated templ code does, implicit interface methods
   (MarshalJSON & co) are exempt by construction. Mutation-proven
   (a planted `ProbeDeadExport` reds it; clean tree green). First
   real catch: `domain.Disposition.AnsweredClass` → UNEXPORTED
   (`answeredClass` — the class law stays domain-internal law, its
   test updated; consumers classify through the vocabulary table).
7. **f.7 — ci-verdict blame lane**: on a RED verdict the script maps
   every `--- FAIL:` name to its declaring file and the newest commit
   touching it; pending runs are refused (nothing concluded yet),
   non-test reds (format/lint/build) are named as such, and a test
   with no declaring file in the checkout is named honestly
   (renamed-since signal). Self-test 8/8 (three new blame cases).
   LIVE-VERIFIED twice: it correctly blamed the concurrent train's
   `voicemail_chain_test.go` → `1dacced6`, and correctly identified
   the `84b2ab1e` failure as a no-FAIL-lines format-class red.
8. **f.5/f.6 — TODO rows**: TODO_LIST gained the
   "Line-ownership model residue" section (Android `lines[]`
   adoption; SSO-only stack fixture lane; hot-reload owner call).
9. **f.17 — language across switches**: verified request-scoped by
   construction (wp-lang cookie, never the session row), then PINNED
   anyway — `TestSessionSwitchKeepsTheChosenLanguage` (switch, then
   the shell still renders `lang="de"`).
10. **f.18 — ownership model diagram** in DOMAIN_LANGUAGE: the three
    feeds → one join → User/Lines → session-on-one-line → surfaces,
    ASCII, with the fallback rule and the switch invariant named.
11. **f.20 — contracts README re-verified**: accurate after the wire
    change; one clause added (the brokered fixture pins the typed
    `lines[{extension,did}]` list).
12. **Depth-budget pin raised honestly**: the f.2 unit grew
    internal/server to 5212 > 5200; pin moved to 5250 WITH the reason
    in the same commit (`30ec09dc`) per the ratchet rule — direction
    stays down, the html carve row owns the split.
13. **Type gate green**: `tsc -p . --noEmit` — TSC_GREEN (collected
    from the background job; the vtsls ui.js editor errors all
    session were stale-editor phantoms, re-confirmed).
14. **Gates that ran green on the touched surfaces**: targeted go
    suites (server, web, app, authmap, domain, arch — the latter with
    an explicit unpiped exit-code check), templ regenerated with the
    pinned compiler, `gofmt` clean, `nix build
    .#checks.x86_64-linux.format` GREEN mid-train, CHANGELOG entry.

## b) PARTIALLY DONE (honest gaps in my own run)

1. **f.4 — ui-capture run NOT executed.** The fixture IS upgraded
   (admin: true + display_name "Capture Operator" + identities
   {1001: +15551234567} — committed via the daemon sweep `6a654fb2`),
   so the matrix would now gain the Admin surface and DID-labeled
   shots, and f.12 (whoami shows the DID for unmapped extension
   logins) rides it. But the harness never ran: no Settings/Admin
   shots, no compare.html, no owner's-eyes proof. The binary build
   it needs was still RUNNING uncollected at snapshot.
2. **f.3 — stack browser E2E scoped, not run.** The stack pins
   webphone `b63bb819` (43 commits behind main); the lane is
   `nix flake lock --update-input webphone` then `nix build -L
   .#telephony-browser` (on-demand VM lane, ~6 min, budget 445s,
   one rerun allowance for the known transfer-step flake). Neither
   step executed. The stack's tree was also mid-flight (its own
   daemon commits landing).
3. **f.16 — hot-reload sketch NOT written** — and the TODO_LIST row
   I wrote REFERENCES it
   (docs/planning/2026-11-18_ownership-hot-reload-sketch.md). A
   dangling reference in the interactive ledger, created by me, in
   the same session that wrote the row. Fix is one doc away.
4. **f.8 — daemon heuristic decided but the row never written.**
   Investigation concluded (the push lives inside go-commit's
   WithAutoPush in `~/projects/projects-management-automation` — a
   cross-repo, fleet-wide change to a LIVE daemon, not a hot-edit
   from a webphone train), but the TODO_LIST row recording that
   verdict was never added. The decision exists only in this report.
5. **Final battery never ran as a battery**: no full
   `go test ./...`, no island node specs, no smoke, no docs-gates
   `--check` (three markdown files landed: TODO_LIST, DOMAIN_LANGUAGE,
   contracts README — dprint owns those bytes in CI), no fresh
   format-check after the last edits. Per-surface gates only.
6. **Push/CI verification incomplete**: origin/main = `30ec09dc`
   (all session work pushed by the daemon except `6a654fb2`), but
   the CI verdicts on my shas were PENDING at snapshot; the known
   failures behind (84b2ab1e format-class, 72f67bc0 race-class)
   belong to the concurrent train (#253 still open, one issue, split
   trigger not fired — their fix still in flight).

## c) NOT STARTED (deliberate or deferred — none silently dropped)

1. f.10 (switcher option styling) — Train 2 visual pass owns control
   chrome; deferred BY THE LIST ITSELF.
2. f.14 (`admin.Number` internal rename) — declined twice now; stays
   declined unless the split-brain bites.
3. f.15 (cohesive config) — reopen trigger is an operator complaint;
   unlit.
4. f.19 (password-fixture lines variant) — correctly absent while
   extension logins carry no lines.
5. f.21 (release-ritual confirmation) — owned by the NEXT release
   train, not this one.
6. The 19-21 report's §g owner questions (switcher home, config
   ritual, SSO VM lane) — the "do the whole list" directive executed
   §f; §g was never answered and is re-presented below (trimmed to
   what still blocks).

## d) TOTALLY FUCKED UP (mine, no excuses)

1. **A piped go-test masked a FAIL and a commit slipped through
   red.** The f.9 commit (`aae159c0`) landed while
   `TestPackageDepthBudget` was RED — `go test | tail` eats the exit
   code, exactly the PIPESTATUS class AGENTS warns about. Caught ~2
   minutes later, healed by the pin-raise commit (`30ec09dc`). The
   fix pattern (capture to a file, then check `$?` explicitly) is
   what the very next command used — the lesson was IN AGENTS before
   I broke it.
2. **`git checkout --` once, reflexively**, to revert the
   mutation-probe in authmap (the NEVER list). `git restore` was the
   tool; no damage (the file was otherwise untouched), but the rule
   broke before the reflex fired.
3. **A multiedit reported "2 of 3 applied" and I almost moved on** —
   the missing piece was the entire active-first test. Caught by
   re-reading the file, re-applied. Half-applied edits are worse
   than failed ones.
4. **Two background jobs went uncollected at the wait boundary**
   (tsc — later collected green; the webphone binary build — still
   running at snapshot). Background work needs a collection step
   before any stopping point; the binary's output path was never
   captured, so f.4 cannot reuse the build without re-asking nix.
5. **Wrote a TODO row that references a document I had not written
   yet** (the hot-reload sketch) and then DID NOT write the sketch
   before the wait. The row is load-bearing in TODO_LIST; the
   reference dangles. Sequence failure, small but sloppy.
6. **Edit-tool mtime races twice** (CHANGELOG, ui-capture.py) — the
   daemon touched files between my read and my edit; both recovered
   by re-reading. Not damage, but each was a wasted round trip the
   read-immediately-before-edit discipline would have avoided.

## e) WHAT WE SHOULD IMPROVE (systemic, from this run)

1. **Never pipe a gate command.** Capture to a temp file, then check
   the exit code explicitly — the arch-commit pattern (`>
   /tmp/out; code=$?`) is the template. The bash-tool shell has no
   PIPESTATUS; piping a gate is structurally incapable of failing.
2. **Concurrent-session nix breakage has a documented fallback — use
   it without ceremony but RECORD it**: the templ FOD placeholder
   (daemon-committed mid-edit of the concurrent session) broke
   `nix develop` for a stretch; store toolchains + the pinned
   `go run templ@v0.3.1100` + `nix build .#checks...format` (which
   does NOT depend on the devshell) kept every gate runnable. The
   outage cost ~10 minutes, all of it discovery.
3. **The daemon's commit race keeps winning** (9 files swept into a
   `chore: heuristic` commit seconds before mine, twice this
   session). Committing immediately after each edit batch — before
   running the slow verification that gives the daemon its window —
   trades narrative purity for actually owning the commit. The
   release-runbook's one-call rule already says this; follow it
   per-unit, not per-train.
4. **Dead-export gate could catch views helpers too** — the scan is
   deliberately scoped to authmap+domain (noise); the views helper
   layer (one-home helpers with micro-tests) is the next natural
   scope if speculative API shows up there.
5. **The blame lane covers Go `--- FAIL:` lines only**; format-class
   reds are named but not file-blamed (dprint/treefmt output shapes
   differ). Cheap follow-up if format reds keep recurring.

## f) NEXT — up to 50, honest (not padded)

1. Collect or kill the running `nix build .#webphone` job; capture
   the output path for f.4.
2. Write the hot-reload sketch (f.16) — kill the dangling TODO_LIST
   reference. SIGHUP vs file-watch, the re-join seam, and the owner
   gate named explicitly.
3. Add the f.8 TODO row (pma push heuristic — decision recorded in
   §b.4 above; the row is the ledger's copy).
4. Run f.4: ui-capture with the upgraded fixture (Settings account
   group, Admin chips, whoami DID — f.12 rides it); eyeball
   compare.html.
5. Run f.3: stack input bump + `nix build -L .#telephony-browser`
   (445s budget; one rerun on the transfer-step flake).
6. Full battery: `go test -count=1 ./...` (env -u GOEXPERIMENT),
   island node specs, fresh `tsc -p . --noEmit`, smoke, format-check.
7. `docs-gates --check` over the landed markdown (TODO_LIST,
   DOMAIN_LANGUAGE, contracts README, this report).
8. Verify CI verdicts per pushed sha (the new blame lane makes
   attribution cheap); re-check #253 — split legs if the -race red
   persists past the concurrent train's fix.
9. Confirm the devshell healed (templ FOD hashes filled by the
   concurrent session) — the store-toolchain fallback retires.
10. Mirror the Android TODO row into webphone-android's TODO_LIST
    (their repo, their ledger).
11. CHANGELOG posture check: the ui-capture fixture change is
    test-tooling (changelog-silent per the bar) — confirm on re-read.
12. Consider one AGENTS.md line for the two new gates (model-api arch
    test + ci-verdict blame) under Commands — discoverability, not
    doctrine (doctrine lives in the test files).
13. The 19-21 §g questions that still block (see §g below) — collect
    owner answers before the next UI train.

## g) QUESTIONS ONLY THE OWNER CAN ANSWER

1. **Ride red main into the stack pin?** Webphone main currently
   carries the concurrent train's `-race` test reds (#253; binary
   builds and passes locally non-race; the browser E2E does not run
   -race). Ritual says bump the stack to main — do you want the E2E
   (f.3) on current main now, or after their race fix lands?
2. **Does the pma push-heuristic (f.8) deserve its own train in
   projects-management-automation now**, or does the TODO row wait
   for the next mid-edit-push CI incident? (Fleet-wide policy call —
   the daemon is live on every repo.)
3. **The two still-blocking §g questions from the 19-21 report**:
   long-term home for the switcher + "My account" (island whoami vs
   the workspace rail/header), and is config-edit+restart the right
   ownership ritual (or reopen cohesive config / hot-reload — the
   sketch above would be the input to that call)?
