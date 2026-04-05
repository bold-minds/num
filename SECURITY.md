# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. **Do Not** Create a Public Issue

Please do not report security vulnerabilities through public GitHub issues, discussions, or pull requests.

### 2. Report Privately

Send an email to **security@boldminds.tech** with the following information:

- **Subject**: Security Vulnerability in bold-minds/num
- **Description**: Detailed description of the vulnerability
- **Steps to Reproduce**: Clear steps to reproduce the issue
- **Impact**: Potential impact and severity assessment
- **Suggested Fix**: If you have ideas for a fix (optional)

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Varies based on complexity, typically within 30 days

### 4. Disclosure Process

1. We will acknowledge receipt of your vulnerability report
2. We will investigate and validate the vulnerability
3. We will develop and test a fix
4. We will coordinate disclosure timing with you
5. We will release a security update
6. We will publicly acknowledge your responsible disclosure (if desired)

## Security Considerations

`num` is a pure-computation library with a very small attack surface:

- **No network I/O.** `num` does not make network calls.
- **No file I/O.** `num` does not read or write files.
- **No reflection.** `num` uses generic type parameters only.
- **No external dependencies.** `num` is pure Go stdlib.
- **No mutation.** `num` never modifies input values.

### DoS Guards

`NewNumberRange` is callable with untrusted numeric input without hanging the caller's goroutine. Three guards cover the degenerate cases:

- **Iteration cap (10,000,000).** Every call materializes at most ten million elements. This bounds both memory (~80 MB for `float64`) and wall-clock time even when the combination of start, end, and step would otherwise produce an unbounded loop — for example, `NewNumberRange(math.Inf(1))` returns a 10-million-element slice rather than looping to float64 precision exhaustion at 2⁵³.
- **NaN step rejection.** A NaN `StepBy` value returns an empty slice immediately. (NaN detection uses the IEEE 754 rule that `x != x` is true iff x is NaN, which is a no-op for integer types.)
- **Zero step coercion.** `StepBy(0)` is silently replaced with `StepBy(1)` inside `NewNumberRange` to prevent infinite loops.
- **Per-iteration progress guard.** Inside the loop, if `i + step` does not strictly move toward the end in the iteration's direction (because of infinite step, NaN propagation, or precision loss), the loop breaks.

### Known Limitations

- **Memory on bounded-but-large ranges.** Below the 10M cap, callers can still request sizable allocations (e.g. 9.9M × `int64` = ~80 MB). Validate the expected range size at your trust boundary if callers are untrusted.
- **NaN `end` or `start` produces empty results**, not an error. This is the deliberate consequence of NaN comparisons always returning false — the direction branches all fall through and the loop never begins. If explicit error reporting matters, validate finite bounds before calling.
- **Two-digit year interpretation for non-numeric types is not applicable** — `num` is numeric only.

## Security Updates

Security updates will be:

- Released as patch versions (e.g., 0.1.1)
- Documented in the CHANGELOG.md
- Announced through GitHub releases
- Tagged with security labels

## Acknowledgments

We appreciate responsible disclosure and will acknowledge security researchers who help improve the security of this project.

Thank you for helping keep our project and users safe!
