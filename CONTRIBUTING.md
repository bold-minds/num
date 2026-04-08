# Contributing

Thank you for your interest in contributing! We welcome contributions that improve the library while maintaining its focus on simplicity, performance, and Go idioms.

## Getting Started

### Prerequisites

- **Go 1.22+**
- **Git**
- **golangci-lint** (optional, for comprehensive linting)

### Development Setup

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/num.git
   cd num
   ```

2. **Run tests**:
   ```bash
   go test -race ./...
   ```

## What We're Looking For

### Encouraged

- **Bug fixes** — fix issues or edge cases
- **Performance improvements** — optimize without breaking compatibility
- **Test enhancements** — add test cases, improve coverage
- **Documentation improvements** — clarify usage, add examples

### Requires Discussion First

- **API changes** — modifications to public interfaces
- **New dependencies** — adding external packages
- **Breaking changes** — changes that affect backward compatibility

### Not Accepted

- **Feature creep** — complex features that don't align with Go idioms
- **Non-idiomatic Go** — code that doesn't follow Go conventions
- **Performance regressions** — changes that significantly slow down the library

## Contribution Process

### 1. Create an Issue First

For significant changes, please create an issue to discuss:
- What problem you're solving
- Your proposed approach
- Any potential breaking changes

### 2. Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** — follow the code style guidelines below, add tests, update documentation as needed.

3. **Validate your changes**:
   ```bash
   go fmt ./...
   go vet ./...
   go test -race ./...
   ```

4. **Commit your changes**:
   ```bash
   git commit -m "feat: add your feature description"
   ```

5. **Push and create a pull request**:
   ```bash
   git push origin feature/your-feature-name
   ```

### 3. Pull Request Guidelines

Your PR should:
- Have a clear title describing the change
- Reference any related issues using `Fixes #123` or `Closes #123`
- Include tests for new functionality
- Pass all CI checks
- Maintain backward compatibility unless discussed otherwise

## Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Write clear, concise comments for public APIs
- Follow Go's error handling patterns
- Write table-driven tests where appropriate
- Test both success and error cases
- Include edge cases (nil values, empty strings, etc.)
- Run tests with `-race` to ensure thread safety

## Commit Messages

We follow conventional commits:

```
type(scope): description
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `chore`

## Code Review

We look for: correctness, performance, style, tests, documentation, and backward compatibility. Initial review within 2-3 business days.

## License

By contributing, you agree that your contributions will be licensed under the same license that covers the project.
