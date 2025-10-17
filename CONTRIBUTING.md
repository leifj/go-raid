# Contributing to go-RAiD

Thank you for your interest in contributing to go-RAiD! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- (Optional) Docker for running CockroachDB
- (Optional) FoundationDB for FDB backend

### Clone and Setup

```bash
git clone <repository-url>
cd go-raid
go mod download
```

## Git Workflow

### Branch Naming

Use descriptive branch names with prefixes:

- `feature/` - New features (e.g., `feature/add-mongodb-backend`)
- `fix/` - Bug fixes (e.g., `fix/raid-version-history`)
- `refactor/` - Code refactoring (e.g., `refactor/storage-interface`)
- `docs/` - Documentation updates (e.g., `docs/update-quickstart`)
- `test/` - Test additions/updates (e.g., `test/add-storage-tests`)
- `chore/` - Maintenance tasks (e.g., `chore/update-dependencies`)

### Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

Format: `<type>: <subject>`

Types:
- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Code style changes (formatting, missing semi colons, etc)
- **refactor**: Code refactoring without changing functionality
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Build process or auxiliary tool changes

Examples:
```
feat: add MongoDB storage backend
fix: correct RAiD version retrieval in CockroachDB
docs: update storage backend comparison table
refactor: simplify file storage ID generation
test: add unit tests for FoundationDB backend
```

### Pull Request Process

1. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the coding standards

3. **Write tests** for your changes

4. **Run tests and linters**:
   ```bash
   go test ./...
   go vet ./...
   go fmt ./...
   ```

5. **Commit your changes** with clear, descriptive messages

6. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Create a Pull Request** with:
   - Clear title and description
   - Reference to any related issues
   - List of changes made
   - Testing performed

## Coding Standards

### Go Style Guide

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions small and focused
- Use meaningful variable and function names
- Add comments for exported functions and types

### Project-Specific Guidelines

1. **Storage Implementations**:
   - Implement the full `storage.Repository` interface
   - Register factory in package `init()`
   - Add comprehensive error handling
   - Include version history support
   - Provide health check implementation

2. **Error Handling**:
   - Use standard errors from `internal/storage`
   - Wrap errors with context: `fmt.Errorf("failed to create RAiD: %w", err)`
   - Log errors appropriately

3. **Testing**:
   - Write unit tests for new functionality
   - Use table-driven tests where appropriate
   - Mock external dependencies
   - Aim for >80% code coverage

4. **Documentation**:
   - Update relevant `.md` files in `docs/`
   - Add inline code comments for complex logic
   - Update `README.md` if adding major features
   - Include examples in documentation

## Adding a New Storage Backend

To add a new storage backend (e.g., MongoDB, Redis):

1. Create a new package under `internal/storage/<backend>/`
2. Implement the `storage.Repository` interface
3. Register the factory in package `init()`:
   ```go
   func init() {
       storage.RegisterFactory(storage.StorageTypeYourBackend, factory)
   }
   ```
4. Add configuration options to `internal/storage/factory.go`
5. Update `internal/config/config.go` with new env variables
6. Add documentation to `docs/storage-backends.md`
7. Update `.env.example` with configuration examples
8. Add to the comparison table in documentation
9. Write tests for the new backend

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/storage/file/...
```

### Writing Tests

Example test structure:
```go
func TestFileStorageCreate(t *testing.T) {
    // Setup
    cfg := &file.Config{DataDir: t.TempDir()}
    repo, err := file.New(cfg)
    require.NoError(t, err)
    defer repo.Close()

    // Test
    raid := &models.RAiD{...}
    result, err := repo.CreateRAiD(context.Background(), raid)

    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
    assert.NotEmpty(t, result.Identifier.ID)
}
```

## Documentation

When adding features:

1. Update relevant documentation in `docs/`
2. Add examples to `docs/QUICKSTART.md` if applicable
3. Update API documentation if endpoints change
4. Add inline code comments for complex logic

## Code Review

All submissions require review. When reviewing:

- Check for adherence to coding standards
- Verify tests are present and pass
- Ensure documentation is updated
- Look for potential bugs or edge cases
- Provide constructive feedback

## Release Process

Maintainers follow semantic versioning (SemVer):

- **MAJOR**: Incompatible API changes
- **MINOR**: Backwards-compatible functionality additions
- **PATCH**: Backwards-compatible bug fixes

## Questions?

- Open an issue for discussion
- Check existing documentation in `docs/`
- Review example code in the repository

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

## Thank You!

Your contributions make go-RAiD better for everyone. We appreciate your time and effort!
