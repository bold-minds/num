package num_test

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/bold-minds/num"
)

func Test_NewNumberRange(t *testing.T) {
	tests := []struct {
		name     string
		end      int
		opts     []num.RangeOption[int]
		expected []int
	}{
		{
			name:     "basic_range",
			end:      5,
			opts:     nil,
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			name:     "range_with_start",
			end:      10,
			opts:     []num.RangeOption[int]{num.StartAt[int](5)},
			expected: []int{5, 6, 7, 8, 9},
		},
		{
			name:     "range_with_step",
			end:      10,
			opts:     []num.RangeOption[int]{num.StepBy[int](2)},
			expected: []int{0, 2, 4, 6, 8},
		},
		{
			name:     "range_reverse",
			end:      0,
			opts:     []num.RangeOption[int]{num.StartAt[int](4)},
			expected: []int{4, 3, 2, 1},
		},
		{
			name:     "range_inclusive",
			end:      5,
			opts:     []num.RangeOption[int]{num.Inclusive[int]()},
			expected: []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:     "range_negative",
			end:      0,
			opts:     []num.RangeOption[int]{num.StartAt[int](-5)},
			expected: []int{-5, -4, -3, -2, -1},
		},
		{
			name:     "range_negative_inclusive",
			end:      -2,
			opts:     []num.RangeOption[int]{num.StartAt[int](-5), num.Inclusive[int]()},
			expected: []int{-5, -4, -3, -2},
		},
		{
			name:     "range_combined_options",
			end:      10,
			opts:     []num.RangeOption[int]{num.StartAt[int](2), num.StepBy[int](3)},
			expected: []int{2, 5, 8},
		},
		{
			name:     "range_zero_step_defaults_to_one",
			end:      3,
			opts:     []num.RangeOption[int]{num.StepBy[int](0)},
			expected: []int{0, 1, 2},
		},
		{
			name:     "range_empty_when_start_equals_end",
			end:      5,
			opts:     []num.RangeOption[int]{num.StartAt[int](5)},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := num.NewNumberRange(tt.end, tt.opts...)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func Test_CalculateActualEnd_Coverage(t *testing.T) {
	// non-inclusive
	if got, want := num.NewNumberRange(5, num.StartAt(1)), []int{1, 2, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("Non-inclusive range: expected %v, got %v", want, got)
	}
	// inclusive forward
	if got, want := num.NewNumberRange(5, num.StartAt(1), num.Inclusive[int]()), []int{1, 2, 3, 4, 5}; !reflect.DeepEqual(got, want) {
		t.Errorf("Inclusive forward range: expected %v, got %v", want, got)
	}
	// inclusive backward
	if got, want := num.NewNumberRange(1, num.StartAt(5), num.Inclusive[int]()), []int{5, 4, 3, 2, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("Inclusive backward range: expected %v, got %v", want, got)
	}
	// start equals end
	if got, want := num.NewNumberRange(5, num.StartAt(5), num.Inclusive[int]()), []int{}; !reflect.DeepEqual(got, want) {
		t.Errorf("Start equals end inclusive: expected %v, got %v", want, got)
	}
}

func Test_FloatRange(t *testing.T) {
	// Floats — a case that `for i := range N` cannot express.
	got := num.NewNumberRange(1.0, num.StartAt(0.0), num.StepBy(0.25))
	want := []float64{0.0, 0.25, 0.5, 0.75}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Float step range: expected %v, got %v", want, got)
	}
}

// progressGuardTimeout is the deadline for the ProgressGuard regression
// tests. It has to be loose enough to absorb the 10M-iteration cap
// path on the slowest supported runner (GitHub Actions shared runners
// are ~4× slower than modern dev hardware for append-heavy loops, and
// Test_ProgressGuard_Infinity materializes ~80 MB of float64), while
// still being short enough that a real regression (a truly infinite
// loop) surfaces within one CI run before the leaked goroutine (see
// runWithTimeout) saturates a CPU core and destabilises the rest of
// the suite. 10s is ~10× headroom over the observed local runtime of
// <1s and matches the fail-fast preference.
const progressGuardTimeout = 10 * time.Second

// runWithTimeout runs fn and fails the test if it does not complete
// within d. Used to lock in DoS guards — if a guard regresses, the
// test times out rather than hanging the whole suite.
//
// Known limitation: if a guard regresses into an actually infinite
// loop, the spawned goroutine survives until the process exits and
// keeps running alongside the rest of the suite. There is no clean
// way to cancel a pure-CPU Go function that takes no context, so
// this tradeoff is accepted: the test fails loudly and the CI run
// will be torn down shortly afterward. If NewNumberRange ever grows
// a context parameter, convert this helper to use context.WithTimeout
// and cancel the worker.
func runWithTimeout(t *testing.T, d time.Duration, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("function did not complete within %s — progress guard likely regressed", d)
	}
}

// Test_ProgressGuard_Infinity is a regression test for a float-infinity
// DoS that existed before the progress guard was added. Without the
// guard, NewNumberRange(math.Inf(1)) enters a forward loop where
// i += 1 makes steady progress until i reaches 2^53, at which point
// float64 precision exhaustion pins i but the i < +Inf bound check
// stays true forever. The iteration cap bails after 10M loops.
func Test_ProgressGuard_Infinity(t *testing.T) {
	runWithTimeout(t, progressGuardTimeout, func() {
		got := num.NewNumberRange(math.Inf(1))
		// The exact count is implementation-defined; we only care
		// that the function terminates and returns a bounded result.
		if len(got) == 0 {
			t.Error("Expected non-empty result for range [0, +Inf), got empty")
		}
	})
}

// Test_ProgressGuard_NegativeInfinity mirrors the above for the
// backward direction — start = -Inf should terminate cleanly.
func Test_ProgressGuard_NegativeInfinity(t *testing.T) {
	runWithTimeout(t, progressGuardTimeout, func() {
		// Forward range with +inf end, -inf start.
		got := num.NewNumberRange(math.Inf(1), num.StartAt(math.Inf(-1)))
		// A subtly broken implementation could return nil (bypassing
		// the non-nil contract) or a slice longer than the documented
		// cap. Lock both properties in, not just termination.
		if got == nil {
			t.Error("Expected non-nil slice for range [-Inf, +Inf), got nil")
		}
		if len(got) == 0 || len(got) > num.MaxElements {
			t.Errorf("Expected 1..MaxElements entries for [-Inf, +Inf), got len=%d", len(got))
		}
	})
}

// Test_ProgressGuard_NaN verifies that NaN inputs return an empty,
// non-nil slice rather than hanging or returning nil. NaN comparisons
// always evaluate to false, which makes every direction and
// continuation predicate in NewNumberRange fail on the first pass.
// The DeepEqual check against []float64{} locks in the empty-but-
// non-nil contract so future refactors cannot silently switch one
// path to returning nil.
func Test_ProgressGuard_NaN(t *testing.T) {
	runWithTimeout(t, progressGuardTimeout, func() {
		got := num.NewNumberRange(math.NaN())
		if !reflect.DeepEqual(got, []float64{}) {
			t.Errorf("Expected empty non-nil slice for NaN end, got %#v", got)
		}
	})
	runWithTimeout(t, progressGuardTimeout, func() {
		got := num.NewNumberRange(10.0, num.StepBy(math.NaN()))
		if !reflect.DeepEqual(got, []float64{}) {
			t.Errorf("Expected empty non-nil slice for NaN step, got %#v", got)
		}
	})
}

// Test_ProgressGuard_PrecisionExhaustion covers the third DoS case:
// a finite range where step is too small relative to start for
// float64 precision to represent the increment. The guard detects
// next == i and bails.
func Test_ProgressGuard_PrecisionExhaustion(t *testing.T) {
	runWithTimeout(t, progressGuardTimeout, func() {
		// 1e20 is well above 2^53, so 1e20 + 1 == 1e20 in float64.
		// The correct behavior is well-defined: the guard fires after
		// the first append, producing a single-element slice (the
		// starting value). A subtly broken implementation might fall
		// through to the iteration cap and return 10M copies of 1e20,
		// so assert the bounded-length contract explicitly.
		got := num.NewNumberRange(1e20+10, num.StartAt(1e20), num.StepBy(1.0))
		if got == nil {
			t.Error("Expected non-nil slice for precision-exhausted range, got nil")
		}
		if len(got) > 1 {
			t.Errorf("Expected at most 1 element for precision-exhausted range, got len=%d", len(got))
		}
	})
}

// -----------------------------------------------------------------
// Regression tests for bugs reported in the v0.1.1 code review.
// Each test references the specific review item it locks in so a
// future refactor that reintroduces the bug fails with an obvious
// name.
// -----------------------------------------------------------------

// Test_UnsignedInclusiveBackwardToZero is the regression for
// review item #1 (HIGH). Before the fix, NewNumberRange[uint](0,
// StartAt(5), Inclusive()) computed end-1 on unsigned zero, wrapped
// to MaxUint, then flipped direction and produced a 10M-element
// forward garbage slice. The expected result is a simple descending
// range that terminates cleanly at zero without underflow.
func Test_UnsignedInclusiveBackwardToZero(t *testing.T) {
	got := num.NewNumberRange[uint](0, num.StartAt[uint](5), num.Inclusive[uint]())
	want := []uint{5, 4, 3, 2, 1, 0}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uint backward-inclusive to zero: expected %v, got %v", want, got)
	}
}

// Test_SignedInclusiveForwardAtMaxT is the regression for review
// item #2 (MEDIUM). Before the fix, NewNumberRange[int8](127,
// StartAt(125), Inclusive()) computed end+1 on int8(127), wrapped
// to int8(-128), flipped direction, and produced backward garbage.
func Test_SignedInclusiveForwardAtMaxT(t *testing.T) {
	got := num.NewNumberRange[int8](127, num.StartAt[int8](125), num.Inclusive[int8]())
	want := []int8{125, 126, 127}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("int8 forward-inclusive to MaxInt8: expected %v, got %v", want, got)
	}
}

// Test_SignedInclusiveBackwardAtMinT is the mirror of the above for
// the signed minimum. int8(-128) on the backward-inclusive path
// previously fell into the same end±1 wrap trap.
func Test_SignedInclusiveBackwardAtMinT(t *testing.T) {
	got := num.NewNumberRange[int8](-128, num.StartAt[int8](-126), num.Inclusive[int8]())
	want := []int8{-126, -127, -128}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("int8 backward-inclusive to MinInt8: expected %v, got %v", want, got)
	}
}

// Test_UnsignedInclusiveForwardAtMaxT covers the symmetric uint8
// case: a forward inclusive range whose terminal is the unsigned
// maximum. The fix-side code must avoid computing i + step at the
// terminal, because that computation wraps to zero.
func Test_UnsignedInclusiveForwardAtMaxT(t *testing.T) {
	got := num.NewNumberRange[uint8](255, num.StartAt[uint8](253), num.Inclusive[uint8]())
	want := []uint8{253, 254, 255}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uint8 forward-inclusive to MaxUint8: expected %v, got %v", want, got)
	}
}

// Test_NegativeStepCoercedToAbs is the regression for review item
// #3 (MEDIUM). Before the fix, a negative step was neither rejected
// nor coerced, and the per-iteration progress guard broke out of
// the loop after a single append, silently returning [0]. The fix
// coerces the step to its absolute value so direction is derived
// purely from start vs end.
func Test_NegativeStepCoercedToAbs(t *testing.T) {
	// Forward range with negative step — should behave identically
	// to step = +2 because direction comes from start < end.
	got := num.NewNumberRange(10, num.StepBy(-2))
	want := []int{0, 2, 4, 6, 8}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("forward range with negative step: expected %v, got %v", want, got)
	}
	// Backward range with negative step — also coerced; direction
	// comes from start > end.
	gotBack := num.NewNumberRange(0, num.StartAt(10), num.StepBy(-2))
	wantBack := []int{10, 8, 6, 4, 2}
	if !reflect.DeepEqual(gotBack, wantBack) {
		t.Errorf("backward range with negative step: expected %v, got %v", wantBack, gotBack)
	}
}

// Test_MinIntStepNegationSurvives locks in the documented behavior
// for the one case where -step cannot be represented in T: the
// signed minimum. For int8(-128), -(-128) == -128 (two's complement
// wrap), so after the negative-step coercion in NewNumberRange the
// step is still negative. The per-iteration progress guard must
// catch this and terminate the loop rather than running to the
// iteration cap.
func Test_MinIntStepNegationSurvives(t *testing.T) {
	runWithTimeout(t, progressGuardTimeout, func() {
		got := num.NewNumberRange[int8](10, num.StepBy[int8](math.MinInt8))
		// The exact result depends on guard ordering; what matters is
		// that the call returns promptly with a bounded, non-nil slice.
		if got == nil {
			t.Error("Expected non-nil slice for MinInt8 step, got nil")
		}
		if len(got) > num.MaxElements {
			t.Errorf("Expected bounded slice for MinInt8 step, got len=%d", len(got))
		}
	})
}

// Test_EmptyPathsAreNonNil is the regression for review item #5.
// Every path that returns an empty slice must return a non-nil
// []T{} so that reflect.DeepEqual against []T{} succeeds
// regardless of which branch produced the result. The NaN branches
// are exercised separately in Test_ProgressGuard_NaN.
func Test_EmptyPathsAreNonNil(t *testing.T) {
	// start == end.
	if got := num.NewNumberRange(5, num.StartAt(5)); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("start==end: expected []int{}, got %#v", got)
	}
	// start == end with Inclusive — existing documented behavior
	// returns empty, not [start]. The test at line 105 already
	// covers this; this variant locks the non-nil property in.
	if got := num.NewNumberRange(5, num.StartAt(5), num.Inclusive[int]()); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("start==end inclusive: expected []int{}, got %#v", got)
	}
}

// Test_UnsignedForwardRange covers the basic unsigned path, which
// had no direct test before the v0.1.1 review. Integer underflow
// and overflow concerns are unique to unsigned types (for the
// zero-terminal case) so an explicit smoke test here prevents
// regressions from a future generic refactor that accidentally
// specializes only on signed types.
func Test_UnsignedForwardRange(t *testing.T) {
	got := num.NewNumberRange[uint](5, num.StartAt[uint](1))
	want := []uint{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uint forward range: expected %v, got %v", want, got)
	}
}

// Test_Float32Range exercises the float32 path. float32's smaller
// mantissa (24 bits) makes precision exhaustion easier to hit than
// float64, and the generic instantiation is distinct from float64.
func Test_Float32Range(t *testing.T) {
	got := num.NewNumberRange[float32](1.0, num.StartAt[float32](0.0), num.StepBy[float32](0.25))
	want := []float32{0.0, 0.25, 0.5, 0.75}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("float32 range: expected %v, got %v", want, got)
	}
}

// Port is a ~-aliased integer type used to verify that the Numeric
// constraint's tilde binding produces a working generic
// instantiation. One-liner test; the compilation itself is most of
// the proof.
type Port uint16

func Test_AliasedTypeRange(t *testing.T) {
	got := num.NewNumberRange[Port](8083, num.StartAt[Port](8080))
	want := []Port{8080, 8081, 8082}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("aliased-type range: expected %v, got %v", want, got)
	}
}

// Test_MaxElementsCapFires asserts that a deliberately oversized
// range truncates at exactly MaxElements. Test_ProgressGuard_Infinity
// exercises the cap incidentally via +Inf; this test fires the cap
// on a well-formed finite range so the constant's public contract
// ("returns at most MaxElements") is directly verified.
func Test_MaxElementsCapFires(t *testing.T) {
	runWithTimeout(t, progressGuardTimeout, func() {
		got := num.NewNumberRange(num.MaxElements + 5)
		if len(got) != num.MaxElements {
			t.Errorf("expected cap to fire at %d, got len=%d", num.MaxElements, len(got))
		}
		// Spot-check first and last elements so we know the cap
		// truncated at the tail, not somewhere mid-stream.
		if got[0] != 0 || got[num.MaxElements-1] != num.MaxElements-1 {
			t.Errorf("capped slice has wrong endpoints: first=%d last=%d",
				got[0], got[num.MaxElements-1])
		}
	})
}

// Test_RepeatedOptionLastWins locks in the standard functional-
// options idiom: if the same option is supplied twice, the last
// supplied value takes effect. This was undocumented before v0.1.1.
func Test_RepeatedOptionLastWins(t *testing.T) {
	if got, want := num.NewNumberRange(10, num.StartAt(1), num.StartAt(5)), []int{5, 6, 7, 8, 9}; !reflect.DeepEqual(got, want) {
		t.Errorf("StartAt last-wins: expected %v, got %v", want, got)
	}
	if got, want := num.NewNumberRange(10, num.StepBy(2), num.StepBy(3)), []int{0, 3, 6, 9}; !reflect.DeepEqual(got, want) {
		t.Errorf("StepBy last-wins: expected %v, got %v", want, got)
	}
}
