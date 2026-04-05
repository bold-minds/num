package num_test

import (
	"math"
	"testing"

	"github.com/bold-minds/num"
)

// All benchmarks call b.ReportAllocs so that a refactor that
// re-introduces per-iteration allocation (e.g. removing the
// estimateCapacity preallocation) shows up in the diff as a change
// in allocs/op, not just as a slightly slower ns/op number.

func Benchmark_NewNumberRange(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(100)
	}
}

func Benchmark_NewNumberRange_Large(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(10000)
	}
}

func Benchmark_NewNumberRange_WithOptions(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(1000, num.StartAt[int](100), num.StepBy[int](2))
	}
}

func Benchmark_NewNumberRange_Reverse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(0, num.StartAt[int](999))
	}
}

func Benchmark_NewNumberRange_Inclusive(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(1000, num.Inclusive[int]())
	}
}

// Benchmark_NewNumberRange_Float exercises the float64 path, which
// has a different progress guard profile (next > i can fail on
// precision exhaustion) and was entirely unbenchmarked before.
func Benchmark_NewNumberRange_Float(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(1000.0, num.StartAt(0.0), num.StepBy(0.5))
	}
}

// Benchmark_NewNumberRange_GuardFastPath exercises a degenerate
// float input whose progress guard fires on the first iteration
// (step = +Inf, so next == i). If the guard ever regresses into a
// slow fallback, this benchmark's ns/op moves from a handful of
// nanoseconds to something much larger.
func Benchmark_NewNumberRange_GuardFastPath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(10.0, num.StepBy(math.Inf(1)))
	}
}
