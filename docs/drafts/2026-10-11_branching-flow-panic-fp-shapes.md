# Panic analysis: 3 guard shapes produce false positives (6 of 8 rows in one repo)

TL;DR: the panic-condition analyzer flags dereferences/indexing it cannot prove guarded — closure-carried nil checks, map-to-slice invariants, and `i<j` two-pointer swaps. In webphone's 2026-10-11 run 6 of 8 rows were false positives across exactly three shapes. Ask: flow-sensitivity for these shapes, or a documented limitations list triagers can pattern-match.

branching-flow 0.6.4 (192031d), `panic all .` against
LarsArtmann/webphone @ `85de30c`. All citations re-verified at that
commit today.

## The three shapes

**1. Nil guard carried in a closure (flagged: Nil Dereference, high).**
`internal/vcard/vcard.go:34` — `cards = append(cards, *current)` is
flagged, but the same closure opens with the guard:

```go
flush := func() {
    if current == nil {
        return
    }
    ...
    cards = append(cards, *current)
```

The analyzer does not connect the closure's entry guard to the later
deref.

**2. Map-to-slice index invariant (flagged: Index Out of Range,
medium).** `internal/store/transcripts.go:154` — `group := &groups[idx]`
where `idx` only ever comes from `byCall[seg.CallID]`, written as
`idx = len(groups)` immediately before `groups = append(...)` in the
same iteration (`:144-145`). The invariant `idx < len(groups)` lives in
the map-write discipline, not in an adjacent check.

**3. Two-pointer swap under `i < j`, and an index guarded in its own
condition.**

- `internal/store/transcripts.go:168` (flagged twice, for `i` and
  `j`): `lines[i], lines[j] = lines[j], lines[i]` inside
  `for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1` — `i < j <
  len` by construction.
- `cmd/webphone/bootreport.go:149-150` (both lines flagged):
  `return copyRows[c].name` is the body of
  `if c >= 0 && int(c) < len(copyRows) && copyRows[c].name != ""` —
  the same statement's condition guards both indexed reads.

## Counts

`panic all .` at `85de30c`: 8 rows, 6 false positives across the three
shapes above. The 2 remaining rows are deliberate boot-contract panics
(go:embed guarantees the subtree; already rationale-commented and
accepted in-repo).

## Ask

Either (cheapest): document these three shapes as known limitations in
`--help`/README so triagers stop re-deriving them; or (better): add
flow-sensitivity for (1) same-function nil guards through closures and
(3) index reads whose only use sits inside a bound-checking condition.
(2) may be genuinely hard — a docs note suffices there.

## Repro

```
git clone https://github.com/LarsArtmann/webphone && cd webphone
git checkout 85de30c
branching-flow panic . --format markdown
```

Full adjudication: `docs/status/2026-10-11_02-19_branching-flow-report-triage-status.md` § Panic conditions.

💘 Generated with Crush
