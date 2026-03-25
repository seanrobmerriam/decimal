# Contributing to Decimal Money Library

Thank you for your interest in contributing to the Decimal Money Library! This library is used in production financial systems, so we hold contributions to high standards for correctness, documentation, and backward compatibility.

## Development Setup

### Prerequisites

- **Go 1.21+** - For the core library and CLI tools
- **Rust (latest stable)** - For the WebAssembly implementation
- **Node.js 18+** - For JavaScript/TypeScript development
- **git** - For version control

### Initial Setup

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/decimal.git
   cd decimal
   ```

3. Install Go dependencies:
   ```bash
   cd go && go mod download
   ```

4. Install Rust dependencies:
   ```bash
   cd rust && cargo build
   ```

5. Install Node.js dependencies:
   ```bash
   cd js && npm install
   ```

### Running Tests

```bash
# Run all Go tests
make test

# Run Go tests with coverage
cd go && go test -v -race -coverprofile=coverage.out ./...

# Run Rust tests
cd rust && cargo test

# Run JavaScript tests
cd js && npm test
```

### Running Benchmarks

```bash
# Run Go benchmarks
make bench

# Results are in go/bench/ directory
```

## Code Style

### Go

- Run `go fmt` before committing
- Follow [Effective Go](https://go.dev/doc/effective_go) conventions
- Document all exported symbols with GoDoc comments
- Keep lines under 100 characters when practical
- Use meaningful variable and function names

```go
// Good
func NewMoney(amount int64, currency Currency) (*Money, error) {

// Avoid
func NewMoney(a int64, c Currency) (*Money, error) {
```

### Rust

- Follow [Rust Style Guide](https://doc.rust-lang.org/beta/style-guide/)
- Use `cargo fmt` before committing
- Run `cargo clippy` to catch common mistakes
- Document public API with doc comments

### JavaScript/TypeScript

- Use ESLint and Prettier configurations provided
- Follow the existing TypeScript conventions
- Write type definitions for all public APIs

## Pull Request Process

1. **Fork the repository** and create your branch from `main`:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes**. Ensure your code:
   - Compiles without errors
   - Passes all tests
   - Follows the code style guidelines
   - Includes tests for new functionality

3. **Add tests** for any new features or bug fixes:
   - Unit tests for core logic
   - Edge case tests for boundary conditions
   - Integration tests for API changes

4. **Update documentation** as needed:
   - Add GoDoc comments for new public APIs
   - Update README.md if adding new features
   - Add examples for new functionality
   - Update docs/ folder for design decisions

5. **Verify all tests pass**:
   ```bash
   make test
   make lint
   ```

6. **Submit a Pull Request** with:
   - A clear title describing the change
   - A detailed description of what changed and why
   - Reference any related issues (e.g., "Fixes #123")
   - Note any breaking changes

## Types of Contributions

### Bug Fixes

- Include a test case that reproduces the bug
- Explain why the bug occurred
- Describe the fix

### New Features

- Discuss major features in an issue first
- Provide motivation and use cases
- Include comprehensive tests
- Update documentation

### Performance Improvements

- Include benchmark results before/after
- Explain the optimization strategy
- Ensure correctness is not compromised

### Documentation

- Fix typos and improve clarity
- Add examples where helpful
- Translate to other languages if fluent

## Reporting Issues

### Bug Reports

Use the [Bug Report](.github/ISSUE_TEMPLATE/bug_report.md) template and include:

- Go version (`go version`)
- Minimal reproducible example
- Expected vs actual behavior
- Full error output if applicable

### Feature Requests

Use the [Feature Request](.github/ISSUE_TEMPLATE/feature_request.md) template and include:

- Clear problem statement
- Proposed solution
- Alternative solutions considered
- Use case examples

### Questions

Use the [Question](.github/ISSUE_TEMPLATE/question.md) template for:

- How-to questions
- Clarification requests
- Usage examples

## Financial Library Conventions

When contributing to a financial library, keep these principles in mind:

1. **Precision is paramount** - A penny off here can become millions lost
2. **Explicit over implicit** - Don't hide rounding, always make it explicit
3. **Fail loudly** - Return errors rather than silent incorrect results
4. **Test edge cases** - Zero, negative, very large numbers
5. **Consistent behavior** - Same inputs must always produce same outputs

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Accept criticism gracefully

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
