# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
