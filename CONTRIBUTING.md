# Contributing to Pulumi Provider for Dex

Thank you for your interest in contributing to the Pulumi Provider for Dex! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- Go 1.24+ installed
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- Docker and Docker Compose (for integration tests)
- Access to a Dex instance with gRPC API enabled (or use the local Docker Compose setup)

### Development Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/kotaicode/pulumi-provource-dex.git
   cd pulumi-provider-dex
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the provider:**
   ```bash
   make build
   ```

4. **Install the provider locally:**
   ```bash
   make install
   ```

5. **Start local Dex for testing:**
   ```bash
   make dex-up
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes:**
   - Follow Go code style guidelines
   - Add tests for new functionality
   - Update documentation as needed

3. **Run tests:**
   ```bash
   make test
   ```

4. **Run linter:**
   ```bash
   golangci-lint run
   ```

5. **Test with local Dex:**
   ```bash
   # Start Dex
   make dex-up
   
   # Run integration tests (if available)
   go test ./... -tags=integration -v
   
   # Test with example
   cd examples/typescript
   pulumi preview
   ```

6. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

### Code Style

- Follow standard Go formatting (`go fmt`)
- Use `golangci-lint` for code quality checks
- Write clear, descriptive commit messages
- Add comments for exported functions and types
- Keep functions focused and small

### Testing

- **Unit tests:** Place test files next to source files (e.g., `client_test.go`)
- **Integration tests:** Use the `integration` build tag
- **Test coverage:** Aim for >80% coverage for new code

Example test structure:
```go
// client_test.go
package main

import (
    "testing"
    // ...
)

func TestClientCreate(t *testing.T) {
    // Test implementation
}
```

### Documentation

- Update `README.md` for user-facing changes
- Update `CHANGELOG.md` for notable changes
- Add code comments for complex logic
- Update examples if API changes

## Pull Request Process

1. **Ensure your code passes all checks:**
   - Tests pass
   - Linter passes
   - Code is formatted
   - Documentation is updated

2. **Create a pull request:**
   - Use a descriptive title
   - Provide a clear description of changes
   - Reference any related issues
   - Include test results if applicable

3. **Respond to feedback:**
   - Address review comments
   - Update code as needed
   - Keep the PR focused on a single feature/fix

### PR Title Format

Use conventional commit format:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test additions/changes
- `refactor:` for code refactoring
- `chore:` for maintenance tasks

Examples:
- `feat: add LDAP connector support`
- `fix: handle preview mode correctly in Update methods`
- `docs: update installation instructions`

## Project Structure

```
.
├── cmd/
│   └── pulumi-resource-dex/    # Provider implementation
├── sdk/                         # Generated SDKs (do not edit manually)
├── examples/                    # Example Pulumi programs
├── docs/                        # Design documents
├── .github/
│   └── workflows/               # CI/CD workflows
└── Makefile                     # Build automation
```

## Adding New Resources

1. **Define the resource struct:**
   - Create `Args` and `State` structs
   - Add `Annotate` method for documentation
   - Implement `Create`, `Read`, `Update`, `Delete` methods

2. **Register the resource:**
   - Add to `main.go` in `WithResources()`

3. **Add tests:**
   - Unit tests for CRUD operations
   - Integration tests with Dex

4. **Update documentation:**
   - Add to README.md
   - Add example usage
   - Update CHANGELOG.md

## Reporting Issues

When reporting issues, please include:

- **Description:** Clear description of the issue
- **Steps to reproduce:** Minimal steps to reproduce
- **Expected behavior:** What should happen
- **Actual behavior:** What actually happens
- **Environment:** Pulumi version, Go version, Dex version
- **Logs:** Relevant error messages or logs

## Security Issues

Please report security vulnerabilities privately via GitHub Security Advisories or email (see SECURITY.md). Do not open public issues for security vulnerabilities.

## Questions?

- Open a GitHub Discussion for questions
- Check existing issues and discussions
- Review the documentation in `docs/`

Thank you for contributing! 🎉

