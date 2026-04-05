// Package num generates numeric ranges with ergonomic options —
// StartAt, StepBy, Inclusive — for integer and floating-point types.
//
// Go 1.22 added the `for i := range N` language form, which covers the
// trivial `0..N-1 step 1` case for ints. This package fills the gaps
// that language form leaves: non-zero starts, custom steps, inclusive
// bounds, reverse ranges, and floating-point ranges. Zero external
// dependencies.
//
// Safety: NewNumberRange materializes at most 10,000,000 elements in a
// single call, and rejects NaN step values up front. These guards
// bound worst-case memory and runtime even when a caller passes
// degenerate float inputs (NaN, Inf, or a step below float precision).
// Callers needing larger ranges should iterate manually or in chunks.
package num

// Numeric is the type constraint for values that can form a range:
// any integer or floating-point type (including type-defined
// equivalents via `~`). It mirrors the union of
// `golang.org/x/exp/constraints.Integer | constraints.Float` but lives
// entirely in stdlib territory.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// NewNumberRange creates a slice of numbers with flexible options.
// The range is inclusive of start, exclusive of end, unless Inclusive
// is passed.
//
// Usage:
//
//	NewNumberRange(5)                           // [0,1,2,3,4] — basic range
//	NewNumberRange(10, StartAt(5))              // [5,6,7,8,9]
//	NewNumberRange(10, StepBy(2))               // [0,2,4,6,8]
//	NewNumberRange(5, Inclusive[int]())         // [0,1,2,3,4,5]
//	NewNumberRange(0, StartAt(-5))              // [-5,-4,-3,-2,-1]
//	NewNumberRange(10, StartAt(2), StepBy(3))   // [2,5,8]
//
// A zero step is silently replaced with step=1 to prevent infinite
// loops. When start == end (or their inclusive equivalents), an empty
// slice is returned.
func NewNumberRange[T Numeric](end T, opts ...RangeOption[T]) []T {
	cfg := &rangeConfig[T]{step: 1}
	for _, opt := range opts {
		opt(cfg)
	}

	// Defend against infinite loops on zero step.
	if cfg.step == 0 {
		cfg.step = 1
	}

	actualEnd := calculateActualEnd(end, cfg)

	if cfg.start == actualEnd {
		return []T{}
	}
	if cfg.start < actualEnd {
		return generateForwardRange(cfg.start, actualEnd, cfg.step)
	}
	return generateBackwardRange(cfg.start, actualEnd, cfg.step)
}

// calculateActualEnd shifts the end boundary by one step when
// Inclusive is set, honoring direction.
func calculateActualEnd[T Numeric](end T, cfg *rangeConfig[T]) T {
	if !cfg.inclusive {
		return end
	}
	if cfg.start < end {
		return end + 1
	}
	if cfg.start > end {
		return end - 1
	}
	return end
}

// maxRangeIterations caps the number of elements NewNumberRange will
// materialize in a single call. This bounds worst-case memory and
// runtime when a caller passes degenerate float inputs (e.g.
// start=0, end=+Inf, step=1) that the per-iteration progress guard
// alone would not catch in practical time — float64 precision only
// pins `i + 1 == i` once i reaches ~2^53, which is ~9 quadrillion
// iterations. 10M is two orders of magnitude above realistic slice
// sizes for an in-memory numeric range and catches the DoS.
const maxRangeIterations = 10_000_000

// isDegenerateStep reports whether a step value would prevent the
// loop from making forward progress. NaN is detected via the IEEE
// 754 rule that NaN compares unequal to itself. For integer types,
// step == step is always true, so this is a no-op on ints.
func isDegenerateStep[T Numeric](step T) bool {
	return step != step
}

func generateForwardRange[T Numeric](start, actualEnd, step T) []T {
	if isDegenerateStep(step) {
		return nil
	}
	var result []T
	i := start
	for iter := 0; iter < maxRangeIterations && i < actualEnd; iter++ {
		result = append(result, i)
		next := i + step
		// Per-iteration progress guard. Catches:
		//   - Infinite step (float64 +Inf): next == i after the first
		//     push to +Inf, so next > i is false.
		//   - Float precision exhaustion: once i exceeds 2^53 with
		//     step == 1, the mantissa cannot represent the increment
		//     and next == i.
		// For integer types, step == 0 is already coerced to 1 in
		// NewNumberRange, so this branch is a no-op under well-formed
		// integer inputs.
		if !(next > i) {
			break
		}
		i = next
	}
	return result
}

func generateBackwardRange[T Numeric](start, actualEnd, step T) []T {
	if isDegenerateStep(step) {
		return nil
	}
	var result []T
	i := start
	for iter := 0; iter < maxRangeIterations && i > actualEnd; iter++ {
		result = append(result, i)
		next := i - step
		// Mirror of the progress guard in generateForwardRange. See
		// that function for the full rationale.
		if !(next < i) {
			break
		}
		i = next
	}
	return result
}

// RangeOption configures range generation.
type RangeOption[T Numeric] func(*rangeConfig[T])

type rangeConfig[T Numeric] struct {
	start     T
	step      T
	inclusive bool
}

// Inclusive returns a RangeOption that includes the end value.
//
//	NewNumberRange(5, Inclusive[int]()) // [0,1,2,3,4,5]
func Inclusive[T Numeric]() RangeOption[T] {
	return func(cfg *rangeConfig[T]) {
		cfg.inclusive = true
	}
}

// StartAt sets the starting value for the range.
//
//	NewNumberRange(10, StartAt(5)) // [5,6,7,8,9]
func StartAt[T Numeric](start T) RangeOption[T] {
	return func(cfg *rangeConfig[T]) {
		cfg.start = start
	}
}

// StepBy sets the step size for the range. A zero step is silently
// replaced with 1 inside NewNumberRange to prevent infinite loops.
//
//	NewNumberRange(10, StepBy(2)) // [0,2,4,6,8]
func StepBy[T Numeric](step T) RangeOption[T] {
	return func(cfg *rangeConfig[T]) {
		cfg.step = step
	}
}
