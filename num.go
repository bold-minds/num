// Package num generates numeric ranges with ergonomic options —
// StartAt, StepBy, Inclusive — for integer and floating-point types.
//
// Go 1.22 added the `for i := range N` language form, which covers
// the trivial `0..N-1` step-1 integer case. This package fills the
// gaps that language form leaves: non-zero starts, custom steps,
// inclusive bounds, reverse ranges, and floating-point ranges. The
// package itself targets Go 1.21 so it can be adopted by codebases
// that have not yet upgraded; the language-form comparison is
// framing, not a requirement. Zero external dependencies.
//
// Safety: NewNumberRange materializes at most [MaxElements] elements
// in a single call, rejects NaN step values up front, coerces a zero
// step to 1, and coerces a negative step to its absolute value.
// These guards bound worst-case memory and runtime even when a
// caller passes degenerate float inputs (NaN, Inf, or a step below
// float precision). Callers needing larger ranges should iterate
// manually or in chunks.
//
// Boundary safety: inclusive ranges whose terminal equals the
// maximum or minimum of T (for example uint8(255) or int8(-128))
// return the expected terminal-inclusive slice without integer
// wrap-around. Every internal empty-result path returns a non-nil
// empty slice so reflect.DeepEqual comparisons are stable.
package num

import "math"

// MaxElements is the hard cap on how many elements a single call to
// [NewNumberRange] will materialize. A call whose logical range
// exceeds this cap returns a truncated slice of length MaxElements
// rather than running to float64 precision exhaustion or OOM.
//
// Callers who need to distinguish natural termination from cap
// truncation can pre-check their bounds. The snippet below assumes
// signed integers with end > start; the exact pre-check depends on T:
//
//	want := uint64(end-start) / uint64(step)
//	if want > num.MaxElements {
//	    // chunk the work, or fail loudly
//	}
const MaxElements = 10_000_000

// Numeric is the type constraint for values that can form a range:
// any integer or floating-point type (including type-defined
// equivalents via `~`). It mirrors the union of
// `golang.org/x/exp/constraints.Integer | constraints.Float` but
// lives entirely in stdlib territory.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// NewNumberRange creates a slice of numbers with flexible options.
// The range is inclusive of start, exclusive of end, unless
// Inclusive is passed.
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
// Step coercion: a zero step is silently replaced with 1, and a
// negative step is replaced with its absolute value. Direction is
// derived from start vs end, not from the sign of step, so a
// negative step is almost certainly a caller mistake — coercion
// keeps behavior predictable without an infinite loop. A NaN step
// returns an empty slice.
//
// When start == end, an empty slice is returned regardless of the
// Inclusive option.
func NewNumberRange[T Numeric](end T, opts ...RangeOption[T]) []T {
	cfg := &rangeConfig[T]{step: 1}
	for _, opt := range opts {
		opt(cfg)
	}

	// Reject NaN step up front, before any arithmetic or branching
	// that would otherwise propagate the NaN through comparisons.
	if isNaN(cfg.step) {
		return []T{}
	}
	// Coerce zero and negative steps. Zero would infinite-loop; a
	// negative step contradicts the direction that start vs end
	// already encodes.
	if cfg.step == 0 {
		cfg.step = 1
	} else if cfg.step < 0 {
		cfg.step = -cfg.step
		// If negation overflowed (step was the minimum value of a
		// signed type, e.g. math.MinInt8), -step wraps back to the
		// same negative number. The per-iteration progress guard in
		// the generators will stop the loop in that case.
	}

	if cfg.start == end {
		return []T{}
	}
	if cfg.start < end {
		return generateForwardRange(cfg.start, end, cfg.step, cfg.inclusive)
	}
	return generateBackwardRange(cfg.start, end, cfg.step, cfg.inclusive)
}

// isNaN reports whether v is IEEE-754 NaN. Relies on the rule that
// NaN compares unequal to itself. For integer types, v == v is
// always true, so this is a compile-time no-op for ints after
// generic instantiation.
func isNaN[T Numeric](v T) bool {
	return v != v
}

// estimateCapacity returns a non-negative capacity hint for the
// result slice. It uses float64 arithmetic so it works uniformly
// across all Numeric types, and it caps at MaxElements so a hostile
// caller cannot force a gigantic allocation via the hint path. A
// return of 0 means "grow dynamically" and is used for any input
// where float64 conversion cannot produce a usable estimate
// (non-finite bounds, zero/NaN step after coercion).
func estimateCapacity[T Numeric](start, end, step T) int {
	diff := float64(end) - float64(start)
	if diff < 0 {
		diff = -diff
	}
	s := float64(step)
	if s < 0 {
		s = -s
	}
	if s == 0 || math.IsNaN(diff) || math.IsInf(diff, 0) || math.IsNaN(s) || math.IsInf(s, 0) {
		return 0
	}
	est := diff/s + 1
	if est > float64(MaxElements) {
		return MaxElements
	}
	if est < 1 {
		return 0
	}
	// est has been bounded to [1, MaxElements] above, so int(est) is
	// safe on both 32-bit and 64-bit int. gosec G115 flags float→int
	// conversions unconditionally; the annotation is preemptive for
	// CI that adopts gosec later.
	return int(est) //nolint:gosec // bounded by MaxElements, fits int32
}

// generateForwardRange walks start → end with +step. The loop is
// written so that:
//
//   - NaN bounds terminate immediately (every NaN comparison is
//     false, so the continuation predicate fails on the first pass).
//   - Inclusive terminal at the maximum value of T (e.g. uint8(255)
//     or int8(127)) stops before i + step wraps around.
//   - Floating-point precision exhaustion (i + step == i) is caught
//     by the per-iteration progress guard.
//   - Every exit path returns a non-nil slice so reflect.DeepEqual
//     comparisons against []T{} are stable.
func generateForwardRange[T Numeric](start, end, step T, inclusive bool) []T {
	result := make([]T, 0, estimateCapacity(start, end, step))
	i := start
	for iter := 0; iter < MaxElements; iter++ {
		// Continuation predicate. Written as `!(i <= end)` rather
		// than `i > end` so a NaN end terminates cleanly: any
		// comparison with NaN is false, so the inner expression is
		// false, its negation is true, and we break.
		if inclusive {
			if !(i <= end) {
				break
			}
		} else {
			if !(i < end) {
				break
			}
		}
		result = append(result, i)
		if i == end {
			// Inclusive terminal reached. Stop before computing
			// i + step — on T's maximum value that step would wrap.
			// Structurally only reachable in the inclusive path
			// (non-inclusive already broke above on !(i < end)); the
			// check stays unconditional for symmetry with backward
			// and because an integer compare per iteration is
			// negligible against the append.
			break
		}
		next := i + step
		// Per-iteration progress guard. Catches infinite step
		// (+Inf → next == i), precision exhaustion (float64 past
		// 2^53 with step == 1), and the pathological MinInt step
		// that survived negation in NewNumberRange.
		if !(next > i) {
			break
		}
		i = next
	}
	return result
}

// generateBackwardRange walks start → end with -step. Mirror of
// generateForwardRange; see there for the full rationale behind the
// continuation predicate and progress guard.
func generateBackwardRange[T Numeric](start, end, step T, inclusive bool) []T {
	result := make([]T, 0, estimateCapacity(start, end, step))
	i := start
	for iter := 0; iter < MaxElements; iter++ {
		if inclusive {
			if !(i >= end) {
				break
			}
		} else {
			if !(i > end) {
				break
			}
		}
		result = append(result, i)
		if i == end {
			// Inclusive terminal reached. Stop before computing
			// i - step — on T's minimum value (e.g. int8(-128) or
			// uint(0)) that step would wrap.
			break
		}
		next := i - step
		if !(next < i) {
			break
		}
		i = next
	}
	return result
}

// RangeOption configures range generation. The configuration
// receiver is intentionally unexported, which means RangeOption
// values can only be constructed by the helpers in this package
// (StartAt, StepBy, Inclusive). External packages cannot implement
// their own options; the option set is closed. Request new options
// via issues on the repository.
type RangeOption[T Numeric] func(*rangeConfig[T])

type rangeConfig[T Numeric] struct {
	start     T
	step      T
	inclusive bool
}

// Inclusive returns a RangeOption that includes the end value.
//
//	NewNumberRange(5, Inclusive[int]()) // [0,1,2,3,4,5]
//
// Caveat: when start == end, NewNumberRange returns an empty slice
// even with Inclusive — the option has no effect on that boundary
// case. Callers who want [start] for that case must special-case it.
func Inclusive[T Numeric]() RangeOption[T] {
	return func(cfg *rangeConfig[T]) {
		cfg.inclusive = true
	}
}

// StartAt sets the starting value for the range. When multiple
// StartAt options are supplied, the last one wins (standard Go
// functional-options idiom).
//
//	NewNumberRange(10, StartAt(5)) // [5,6,7,8,9]
func StartAt[T Numeric](start T) RangeOption[T] {
	return func(cfg *rangeConfig[T]) {
		cfg.start = start
	}
}

// StepBy sets the step size for the range. A zero step is coerced
// to 1, and a negative step is coerced to its absolute value, both
// inside NewNumberRange; direction is derived from start vs end.
// A NaN step produces an empty slice.
//
//	NewNumberRange(10, StepBy(2)) // [0,2,4,6,8]
func StepBy[T Numeric](step T) RangeOption[T] {
	return func(cfg *rangeConfig[T]) {
		cfg.step = step
	}
}
