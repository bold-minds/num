package num_test

import (
	"testing"

	"github.com/bold-minds/num"
)

func Benchmark_NewNumberRange(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(100)
	}
}

func Benchmark_NewNumberRange_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(10000)
	}
}

func Benchmark_NewNumberRange_WithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(1000, num.StartAt[int](100), num.StepBy[int](2))
	}
}

func Benchmark_NewNumberRange_Reverse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(0, num.StartAt[int](999))
	}
}

func Benchmark_NewNumberRange_Inclusive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = num.NewNumberRange(1000, num.Inclusive[int]())
	}
}
