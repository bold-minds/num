# num

[![Go Reference](https://pkg.go.dev/badge/github.com/bold-minds/num.svg)](https://pkg.go.dev/github.com/bold-minds/num)
[![Build](https://img.shields.io/github/actions/workflow/status/bold-minds/num/test.yaml?branch=main&label=tests)](https://github.com/bold-minds/num/actions/workflows/test.yaml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bold-minds/num)](go.mod)

**Numeric ranges with ergonomic options — StartAt, StepBy, Inclusive — for integer *and* floating-point types.**

Go 1.22 added `for i := range N` at the language level, which covers the `0..N-1` step-1 integer case. `num` fills the gaps that language form leaves: custom starts, non-unit steps, inclusive bounds, reverse ranges, and anything involving floats.

```go
// 0..9
nums := num.NewNumberRange(10)

// 5..9 with step 1
nums = num.NewNumberRange(10, num.StartAt(5))

// 0, 2, 4, 6, 8
nums = num.NewNumberRange(10, num.StepBy(2))

// 0, 0.25, 0.5, 0.75 — floats, not expressible with `for i := range N`
fs := num.NewNumberRange(1.0, num.StartAt(0.0), num.StepBy(0.25))

// 5, 4, 3, 2, 1 — reverse
nums = num.NewNumberRange(0, num.StartAt(5))

// 0..10 inclusive
nums = num.NewNumberRange(10, num.Inclusive[int]())
```

## ✨ Why num?

- 🎯 **One function, option-configured** — `NewNumberRange` with small composable options. No `Range`, `RangeInclusive`, `RangeFloat`, `RangeReverse` explosion.
- 🔢 **Integers and floats** — generic over a local `Numeric` constraint covering all integer and float types, including `~`-aliased types.
- 🛡️ **DoS-safe for hostile float input** — NaN step rejected, Inf step terminates, 10M element iteration cap prevents precision-exhaustion hangs. Safe to call with user-supplied numeric input.
- 🪶 **Zero dependencies** — pure Go stdlib. The `Numeric` constraint is local, not `golang.org/x/exp/constraints`.
- 🎓 **Retires gracefully** — if Go ever adds float/step/inclusive range syntax, `num` becomes obsolete. That's the point of small libraries.

## 📦 Installation

```bash
go get github.com/bold-minds/num
```

Requires Go 1.21 or later.

## 🎯 Quick Start

```go
package main

import (
    "fmt"

    "github.com/bold-minds/num"
)

func main() {
    // Basic: 0..4
    fmt.Println(num.NewNumberRange(5))
    // → [0 1 2 3 4]

    // Start + step: 2, 5, 8
    fmt.Println(num.NewNumberRange(10, num.StartAt(2), num.StepBy(3)))
    // → [2 5 8]

    // Inclusive end: 0..5
    fmt.Println(num.NewNumberRange(5, num.Inclusive[int]()))
    // → [0 1 2 3 4 5]

    // Negative range: -5..-1
    fmt.Println(num.NewNumberRange(0, num.StartAt(-5)))
    // → [-5 -4 -3 -2 -1]

    // Float step: 0, 0.25, 0.5, 0.75
    fmt.Println(num.NewNumberRange(1.0, num.StartAt(0.0), num.StepBy(0.25)))
    // → [0 0.25 0.5 0.75]
}
```

## 📚 API

### `NewNumberRange[T Numeric](end T, opts ...RangeOption[T]) []T`

Returns a slice of values from `start` (default `0`) to `end`, exclusive of `end` unless `Inclusive` is passed. Direction is inferred from `start` vs `end` — a higher start produces a backward range.

### Options

| Option | Effect |
|---|---|
| `StartAt[T](start T)` | Override the default start of `0` |
| `StepBy[T](step T)` | Override the default step of `1`. `0` is silently replaced with `1`. |
| `Inclusive[T]()` | Include `end` in the result |

### `Numeric` constraint

```go
type Numeric interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
        ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
        ~float32 | ~float64
}
```

`Numeric` mirrors `golang.org/x/exp/constraints.Integer | constraints.Float` but is defined locally so `num` has zero external dependencies. `~` means type-defined aliases (`type Port int`) work without explicit conversion.

## 🧭 When to use `num.NewNumberRange` vs `for i := range N`

Prefer the language form when it fits:

```go
// Preferred for the trivial case
for i := range 10 { /* ... */ }
```

Use `num.NewNumberRange` when you need any of:

- A non-zero start
- A step other than 1
- An inclusive end
- A reverse range
- A float range
- A **materialized slice** rather than an iteration (for sorting, slicing, passing to APIs that take `[]T`)

## 🔗 Related bold-minds libraries

- [`bold-minds/each`](https://github.com/bold-minds/each) — slice operations like filter/group. `num` produces slices; `each` operates on them.
- [`bold-minds/list`](https://github.com/bold-minds/list) — set operations on slices (union, intersect, difference). Pairs with `num` when you need numeric set algebra.
- [`bold-minds/to`](https://github.com/bold-minds/to) — safe type conversion. Useful when feeding user-input numerics into `NewNumberRange`.

## 🛡️ Safety guarantees

`NewNumberRange` is callable with untrusted numeric input:

- **10M element cap.** Every call materializes at most 10,000,000 elements. Beyond that, the loop bails with a truncated result rather than looping to float64 precision exhaustion.
- **NaN step is rejected** up front — returns an empty slice.
- **Inf step terminates** — the per-iteration progress guard catches it.
- **Zero step is coerced to 1** to prevent infinite loops from an explicit `StepBy(0)`.
- **Precision exhaustion** (e.g. `NewNumberRange(1e20+10, StartAt(1e20), StepBy(1.0))`) is caught by the same progress guard.

## 🚫 Non-goals

- **No iterator form.** `NewNumberRange` returns a materialized `[]T`. If you want lazy iteration, use `for i := range N` (integers) or write a short generator yourself.
- **No `Must*` variants.** The function cannot fail on well-typed input; there's nothing to panic about.
- **No automatic overflow protection.** If you pass an integer range that would overflow `T`, you get the hardware wrap-around. Validate bounds beforehand on untrusted input.
- **NaN `end` or `start` produces empty, not an error.** If explicit error reporting matters, validate finite bounds before calling.
- **No unbounded ranges.** The 10M cap is a hard ceiling. If you legitimately need a billion-element numeric sequence, write your own loop — you don't want it materialized in memory anyway.

## 📄 License

MIT — see [LICENSE](LICENSE).
