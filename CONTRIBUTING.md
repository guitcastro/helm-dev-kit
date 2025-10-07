# Contributing to Helm Dev Kit

Thank you for your interest in contributing to Helm Dev Kit! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Install Go 1.20 or later
4. Install development dependencies:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install gosec
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
```

## Development Workflow

### Before You Start

1. Check existing issues and PRs to avoid duplicating work
2. Create an issue to discuss significant changes
3. Fork and create a feature branch from `main`

### Making Changes

1. Write clear, concise commit messages
2. Add tests for new functionality
3. Update documentation as needed
4. Run the full test suite before committing

### Testing

Run all tests:

```bash
make test
```

Run integration tests:

```bash
make test-integration
```

Check test coverage:

```bash
make coverage
```

### Code Quality

Run linting and static analysis:

```bash
make lint
make staticcheck
make security
```

Or run all quality checks:

```bash
make check
```

### Building

Build the binary:

```bash
make build
```

Clean build artifacts:

```bash
make clean
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Follow the existing code style
- Write meaningful variable and function names
- Add comments for exported functions and complex logic

## Testing Guidelines

- Write unit tests for all new functions
- Include integration tests for CLI functionality
- Test both success and error scenarios
- Use table-driven tests when appropriate
- Maintain test coverage above 80%

## Documentation

- Update README.md for user-facing changes
- Update USAGE.md for CLI changes
- Add inline documentation for complex code
- Include examples in documentation

## Pull Request Process

1. Update CHANGELOG.md with your changes
2. Ensure all tests pass and coverage is maintained
3. Update documentation as needed
4. Create a pull request with:
   - Clear title and description
   - Reference to related issues
   - Screenshots/examples if applicable

### PR Checklist

- [ ] Tests pass locally
- [ ] Code follows project style guidelines
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] No breaking changes (or clearly documented)

## Commit Message Format

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions/modifications
- `chore`: Maintenance tasks

Examples:
```
feat(converter): add support for ConfigMaps
fix(parser): handle empty HCL files correctly
docs(readme): update installation instructions
```

## Release Process

Releases are automated via GitHub Actions when version tags are pushed:

1. Update version in `go.mod` if needed
2. Update CHANGELOG.md
3. Create and push a git tag: `git tag v1.0.0 && git push origin v1.0.0`
4. GitHub Actions will build and publish the release

## Getting Help

- Check existing issues and documentation
- Ask questions in GitHub Discussions
- Open an issue for bugs or feature requests

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Follow GitHub's Community Guidelines

Thank you for contributing! 🚀