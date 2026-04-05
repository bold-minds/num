# 🚀 Validation Script

This directory contains scripts for validating the codebase in both local development and CI/CD environments.

## 📋 validate.sh

The main validation script that runs a comprehensive pipeline to ensure code quality, correctness, and readiness for deployment.

### 🎯 Usage

```bash
# Local development mode (default)
./scripts/validate.sh

# CI/CD mode (stricter validation)
./scripts/validate.sh ci

# With custom coverage threshold
COVERAGE_THRESHOLD=85 ./scripts/validate.sh ci

# With custom test timeout
TEST_TIMEOUT=15m ./scripts/validate.sh
```

### 🔧 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `COVERAGE_THRESHOLD` | `80` | Minimum test coverage percentage required |
| `TEST_TIMEOUT` | `10m` | Maximum time allowed for test execution |
| `INTEGRATION_TAG` | `integration` | Build tag for integration tests |

### 🏃‍♂️ Validation Steps

The script runs the following validation steps in order:

1. **🔍 Environment Check** - Verifies Go version, git, and repository status
2. **🎨 Code Formatting** - Ensures code is properly formatted with `go fmt`
3. **🔍 Comprehensive Linting** - Runs `golangci-lint` with security scan, TODO detection, and style checks
4. **🔬 Static Analysis** - Performs static analysis with `go vet`
5. **🏗️ Build Validation** - Validates clean builds and dependency management
6. **🧪 Unit Tests** - Runs all unit tests with race detection
7. **🔗 Integration Tests** - Executes integration test suite
8. **📊 Coverage Check** - Validates test coverage meets threshold
9. **📚 Documentation** - Checks for missing README files
10. **🎯 TODO Check** - Validates outstanding TODO items in Claude_TODO.md
11. **🧹 Final Validation** - Ensures clean git status (CI mode)

### 🎨 Features

- **🌈 Colorful Output** - Beautiful, emoji-rich terminal output
- **⚡ Fast Feedback** - Fails fast on first error for quick iteration
- **🔄 Mode Awareness** - Different behavior for local vs CI environments
- **📊 Detailed Reporting** - Comprehensive summary with timing and statistics
- **🐧 Linux Compatible** - Fully compatible with Linux, macOS, and CI environments
- **🛠️ Tool Installation** - Auto-installs missing tools in CI mode

### 🚨 Exit Codes

- `0` - All validations passed ✅
- `1` - One or more validations failed ❌

### 📋 Prerequisites

**Required:**
- Go 1.19+ 
- Git
- Linux/macOS/WSL environment

**Optional (auto-installed in CI):**
- `golangci-lint` - For comprehensive linting
- `gosec` - For security scanning
- `bc` - For coverage calculations

### 🔧 CI/CD Integration

#### GitHub Actions Example
```yaml
- name: Run Validation Pipeline
  run: ./scripts/validate.sh ci
  env:
    COVERAGE_THRESHOLD: 85
```

#### GitLab CI Example
```yaml
validate:
  script:
    - ./scripts/validate.sh ci
  variables:
    COVERAGE_THRESHOLD: "85"
```

### 🎯 Local Development

For local development, the script is more forgiving:
- Missing tools show warnings instead of failures
- Documentation issues are non-blocking
- TODO items don't fail the pipeline

### 🚀 Quick Start

```bash
# Make sure you're in the project root
cd /path/to/tvzr

# Run the validation
./scripts/validate.sh

# If everything passes, you'll see:
# 🎉 ALL VALIDATIONS PASSED! 🎉
# ✨ Your code is ready to ship! ✨
```

## 🤝 Contributing

When adding new validation steps:
1. Create a new function following the naming pattern `validate_*` or `run_*`
2. Add proper error handling and informative output
3. Include emojis for visual consistency 🎨
4. Test in both local and CI modes
5. Update this README with the new step

## 🐛 Troubleshooting

**Common Issues:**

- **Go version too old**: Upgrade to Go 1.19+
- **golangci-lint not found**: Install from https://golangci-lint.run/usage/install/
- **Coverage below threshold**: Write more tests or lower `COVERAGE_THRESHOLD`
- **Integration tests failing**: Check test setup and database connections
- **Git status dirty**: Commit or stash your changes before running in CI mode
