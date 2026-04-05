# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] — Code review fixes

### Fixed
- **Unsigned backward-inclusive to zero no longer underflows.** `NewNumberRange[uint](0, StartAt[uint](5), Inclusive[uint]())` previously computed `end - 1` on unsigned zero, wrapped to `MaxUint`, flipped direction, and produced a 10M-element forward garbage slice. It now returns `[5 4 3 2 1 0]`. Regression locked in by `Test_UnsignedInclusiveBackwardToZero`.
- **Signed inclusive ranges at the type's extrema no longer wrap.** `NewNumberRange[int8](127, StartAt[int8](125), Inclusive[int8]())` previously computed `end + 1` on `int8(127)`, wrapped to `-128`, and produced backward garbage. The mirror case at `int8(-128)` had the same bug. Both now produce the expected terminal-inclusive result. The `uint8(255)` forward case is also covered. Regressions locked in by `Test_SignedInclusiveForwardAtMaxT`, `Test_SignedInclusiveBackwardAtMinT`, and `Test_UnsignedInclusiveForwardAtMaxT`.
- **Negative step is now coerced to `abs(step)`** instead of silently returning `[0]`. Direction is derived from `start` vs `end`, never from the sign of `step`. Consistent with the existing zero-step coercion. Regression locked in by `Test_NegativeStepCoercedToAbs`.

### Changed
- `calculateActualEnd` is gone. The inclusive flag now flows directly into `generateForwardRange` / `generateBackwardRange`, which decide continuation via `!(i <= end)` / `!(i >= end)`. The `!(...)` form is deliberate — it preserves NaN-safe termination because every comparison with NaN evaluates to false, whereas the De Morgan rewrite `i > end` would not terminate. A `staticcheck -QF1001` exclusion in `.golangci.yml` documents this.
- NaN step is now rejected at the top of `NewNumberRange`, before any arithmetic. The package doc previously claimed this but the check lived inside the generators. Doc and behavior now agree.
- Every empty-result path returns a non-nil `[]T{}`. Previously some paths returned `nil` and others returned `[]T{}`, making `reflect.DeepEqual` comparisons flaky depending on which branch fired. Regression locked in by `Test_EmptyPathsAreNonNil` and an updated `Test_ProgressGuard_NaN`.
- Result slices are now preallocated to a capacity estimate computed in `float64` space and clamped to `MaxElements`. Benchmarks show allocs/op dropped to 2 across all range sizes (one config pointer, one slice) with no intermediate grow-reallocs.
- `isDegenerateStep` is renamed to `isNaN` — the old name implied defenses against a broader class of degenerate values than it actually provided.

### Added
- **`MaxElements` (exported constant, value `10_000_000`).** Callers who need to distinguish natural termination from cap truncation can now pre-check their bounds against this constant. The constant's contract is verified directly by `Test_MaxElementsCapFires`, which fires the cap on a well-formed finite input rather than relying on a `+Inf` side effect.
- Regression / coverage tests for the unsigned path, `float32`, `~`-aliased integer types (`type Port uint16`), and repeated-option last-wins semantics. The unsigned and `float32` paths had no direct coverage before.
- Float benchmark, progress-guard benchmark, and `b.ReportAllocs` on every benchmark. A refactor that reintroduces per-iteration allocation now shows up in the diff as an allocs/op change.

### Fixed (tooling)
- `.golangci.yml` depguard rule now allows the package under test to be imported from `_test.go` files. Previously the black-box `num_test` package tripped depguard; this was pre-existing but never blocked anything because the failure was on a path CI was not yet running strictly.

## [0.1.0] — Initial release

### Added
- `NewNumberRange[T Numeric](end T, opts ...RangeOption[T]) []T` — generate a slice of numbers with ergonomic options.
- `StartAt[T](start T) RangeOption[T]` — set the starting value of the range.
- `StepBy[T](step T) RangeOption[T]` — set the step size. A zero step is silently replaced with 1 to prevent infinite loops.
- `Inclusive[T]() RangeOption[T]` — include the end value in the range.
- `Numeric` type constraint — local union of all integer and floating-point types (and their `~`-underlying type aliases), replacing a previous dependency on `golang.org/x/exp/constraints`.
- Forward, backward, inclusive, exclusive, float-step, and negative-range coverage.
- DoS guards on `NewNumberRange`: 10-million element iteration cap, NaN step rejection, and per-iteration progress check. These bound worst-case memory and runtime even for hostile float inputs (NaN, Inf, precision-exhausted step). Regression tests lock each guard in place.
- Zero external dependencies; benchmarks for every path.

### Requires
- Go 1.21 or later
