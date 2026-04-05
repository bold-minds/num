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

// runWithTimeout runs fn and fails the test if it does not complete
// within d. Used to lock in DoS guards — if a guard regresses, the
// test times out rather than hanging the whole suite.
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
// stays true forever. The guard detects that next == i and breaks.
func Test_ProgressGuard_Infinity(t *testing.T) {
	runWithTimeout(t, 2*time.Second, func() {
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
	runWithTimeout(t, 2*time.Second, func() {
		// Forward range with +inf end, -inf start.
		got := num.NewNumberRange(math.Inf(1), num.StartAt(math.Inf(-1)))
		_ = got // Just verify it returns within timeout.
	})
}

// Test_ProgressGuard_NaN verifies that NaN inputs return empty rather
// than hanging. NaN comparisons always evaluate to false, which makes
// all three branch conditions in NewNumberRange fall through to the
// backward path — the progress guard then catches the degenerate
// step and bails immediately.
func Test_ProgressGuard_NaN(t *testing.T) {
	runWithTimeout(t, 2*time.Second, func() {
		got := num.NewNumberRange(math.NaN())
		if len(got) != 0 {
			t.Errorf("Expected empty result for NaN end, got %v", got)
		}
	})
	runWithTimeout(t, 2*time.Second, func() {
		got := num.NewNumberRange(10.0, num.StepBy(math.NaN()))
		if len(got) != 0 {
			t.Errorf("Expected empty result for NaN step, got %v", got)
		}
	})
}

// Test_ProgressGuard_PrecisionExhaustion covers the third DoS case:
// a finite range where step is too small relative to start for
// float64 precision to represent the increment. The guard detects
// next == i and bails.
func Test_ProgressGuard_PrecisionExhaustion(t *testing.T) {
	runWithTimeout(t, 2*time.Second, func() {
		// 1e20 is well above 2^53, so 1e20 + 1 == 1e20 in float64.
		got := num.NewNumberRange(1e20+10, num.StartAt(1e20), num.StepBy(1.0))
		_ = got // Just verify it returns within timeout.
	})
}
