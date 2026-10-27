# erraudit `generic_return` / opaque-wrap rule probes

Empirical probes that decoded the erraudit `generic_return` rule (tool build
`61a744c`, 2026-10-27). Run `erraudit lint` against a copy of this directory to
re-verify a future tool build still behaves the same; if the verdicts change,
the train that relied on them (zero typed-error conversions, see
`docs/status/2026-10-27_18-23_erraudit-max-zero-typed-error-train.md`) needs a
re-read BEFORE trusting the rule.

## Decoded rule (empirical, not a spec)

`generic_return` flags a function ONLY when BOTH hold:

1. It returns bare `error` (not a specific error type).
2. It CONSTRUCTS a stdlib error somewhere in its body — `errors.New`,
   `fmt.Errorf`, `errors.Join` — including inside nested closures and
   composite literals (`&joinish{errs: []error{errors.New("a"), …}}` counts,
   see b.go `k`).

Exemptions (verified):

- Pure error propagation, even from a constructing same-package callee
  (a.go `d` calls `g` — only `g` is flagged).
- `New`-prefixed function names (g.go `NewThing` — constructs, not flagged).
- Functions returning a specific error type (b.go `NewRejection`/`WrapRejection`).
- Same-package callees are ANALYZED: a lowercase same-package constructor with
  a typed return is clean and does NOT flag its caller (c.go/g.go `newRej`,
  `wrapf`).
- Opaque cross-package constructors are NOT analyzed, but a cross-package
  constructor NAMED `Wrapf` still flags the caller (h.go `verify2` +
  other/o.go — the sole trigger that flagged `userauth.VerifyEnrollToken`,
  which had no stdlib construction of its own). `other.NewRejection` does not
  flag. Treat the name match as a heuristic, not a guarantee.

## Files

- `a.go` — bare construction flags (`b`, `c`, `g`); propagation exempt (`d`);
  interface-method shape.
- `b.go` — typed same-package constructors clean; composite-literal
  construction counts (`k`); `joinish` mirrors the pre-train readmodels shape.
- `c.go` — same-package lowercase typed constructor (`newRej`) clean.
- `g.go` — `New`-prefix exemption (`NewThing`); same-package `wrapf` clean.
- `h.go` + `other/` — the opaque-wrap heuristic: cross-package `Wrapf` flags
  the caller, `NewRejection` does not.

`go.mod` exists only so the probes form a module for standalone erraudit runs;
the directory is Go-toolchain-inert inside this repo (testdata is skipped by
`./...`).
