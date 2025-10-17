# Comprehensive Testing Plan for go-RAiD

## Testing Strategy Overview

This document outlines the comprehensive testing strategy for the go-RAiD project, covering all storage backends, HTTP handlers, and integration scenarios.

## Testing Pyramid

```
                    ┌─────────────────┐
                    │   E2E Tests     │  Small
                    │   (Realistic)   │
                    └─────────────────┘
                 ┌──────────────────────┐
                 │  Integration Tests   │  Medium
                 │  (API + Storage)     │
                 └──────────────────────┘
            ┌────────────────────────────────┐
            │       Unit Tests               │  Large
            │  (Individual Components)       │
            └────────────────────────────────┘
```

### Unit Tests (70% of tests)
- Individual storage backend implementations
- Handler logic with mocked dependencies
- Configuration parsing
- Model validation
- Utility functions

### Integration Tests (25% of tests)
- HTTP API with mock storage
- Storage backend with real dependencies (where feasible)
- Configuration loading and dependency injection
- Error propagation across layers

### End-to-End Tests (5% of tests)
- Complete workflows with real storage backends
- API endpoints with actual database/filesystem
- Migration scenarios between backends
- Performance and stress testing

## Test Coverage Goals

- **Overall Coverage**: ≥ 80%
- **Critical Paths**: ≥ 95% (storage CRUD operations, handler logic)
- **Error Handling**: 100% (all error paths tested)
- **Edge Cases**: Comprehensive (nil checks, boundary conditions, concurrency)

## Unit Test Structure

### 1. Storage Backends

#### File Storage (`internal/storage/file/file_test.go`)

**Test Categories:**
```go
// Basic CRUD Operations
TestFileStorage_CreateRAiD
TestFileStorage_GetRAiD
TestFileStorage_UpdateRAiD
TestFileStorage_DeleteRAiD
TestFileStorage_ListRAiDs

// ServicePoint Operations
TestFileStorage_CreateServicePoint
TestFileStorage_GetServicePoint
TestFileStorage_UpdateServicePoint
TestFileStorage_ListServicePoints

// History and Versioning
TestFileStorage_GetRAiDHistory
TestFileStorage_VersionTracking
TestFileStorage_HistoryRetrieval

// Concurrency
TestFileStorage_ConcurrentWrites
TestFileStorage_ConcurrentReads
TestFileStorage_RaceConditions

// Error Handling
TestFileStorage_NotFound
TestFileStorage_AlreadyExists
TestFileStorage_InvalidInput
TestFileStorage_FileSystemErrors
TestFileStorage_CorruptedData

// Edge Cases
TestFileStorage_EmptyDirectory
TestFileStorage_LargeDatasets
TestFileStorage_SpecialCharactersInIDs
TestFileStorage_UnicodeHandling

// Identifier Generation
TestFileStorage_GenerateIdentifier
TestFileStorage_UniqueIdentifiers
TestFileStorage_HandleCounter
```

#### Git Storage (`internal/storage/file/git_test.go`)

**Test Categories:**
```go
// Git Operations
TestGitStorage_AutoCommit
TestGitStorage_CommitMessages
TestGitStorage_GetGitLog
TestGitStorage_GitNotAvailable

// History Integration
TestGitStorage_HistoryWithGit
TestGitStorage_GitLogParsing
TestGitStorage_MultipleCommits

// Error Handling
TestGitStorage_GitCommandFailure
TestGitStorage_GracefulDegradation
TestGitStorage_NoGitRepository

// Wrapping Behavior
TestGitStorage_DelegatesAllMethods
TestGitStorage_TransparentWrapper
```

#### FoundationDB Storage (`internal/storage/fdb/fdb_test.go`)

**Test Categories:**
```go
// Database Operations (with mock FDB)
TestFDBStorage_CreateRAiD
TestFDBStorage_GetRAiD
TestFDBStorage_UpdateRAiD
TestFDBStorage_ListRAiDs
TestFDBStorage_GetRAiDHistory

// Transactions
TestFDBStorage_TransactionIsolation
TestFDBStorage_AtomicOperations
TestFDBStorage_ConflictHandling

// Directory Management
TestFDBStorage_DirectoryStructure
TestFDBStorage_KeyEncoding
TestFDBStorage_VersionKeys

// Atomic Counter
TestFDBStorage_GenerateIdentifier
TestFDBStorage_CounterConcurrency
TestFDBStorage_CounterPersistence

// Error Handling
TestFDBStorage_ConnectionFailure
TestFDBStorage_TransactionTimeout
TestFDBStorage_InvalidDirectory

// Pagination
TestFDBStorage_PaginationLimit
TestFDBStorage_PaginationContinuation
TestFDBStorage_EmptyResults
```

**Note:** FDB tests will use interfaces/mocks since FoundationDB requires external service.

#### CockroachDB Storage (`internal/storage/cockroach/cockroach_test.go`)

**Test Categories:**
```go
// Database Operations (with testcontainers or mock)
TestCockroachStorage_SchemaInit
TestCockroachStorage_CreateRAiD
TestCockroachStorage_GetRAiD
TestCockroachStorage_UpdateRAiD
TestCockroachStorage_ListRAiDs
TestCockroachStorage_GetRAiDHistory

// SQL Operations
TestCockroachStorage_JSONBQueries
TestCockroachStorage_InvertedIndexes
TestCockroachStorage_VersionTracking
TestCockroachStorage_Transactions

// ServicePoint Queries
TestCockroachStorage_ServicePointByIdentifier
TestCockroachStorage_ServicePointFiltering

// Pagination
TestCockroachStorage_OffsetLimit
TestCockroachStorage_EmptyResults

// Error Handling
TestCockroachStorage_ConnectionFailure
TestCockroachStorage_DuplicateKey
TestCockroachStorage_ConstraintViolation
TestCockroachStorage_InvalidJSON

// Concurrency
TestCockroachStorage_ConcurrentWrites
TestCockroachStorage_TransactionIsolation
```

**Note:** Can use `testcontainers-go` to spin up real PostgreSQL/CockroachDB for integration tests.

### 2. HTTP Handlers

#### RAiD Handlers (`internal/handlers/raid_test.go`)

**Test Categories:**
```go
// HTTP Endpoints
TestMintRAiD_Success
TestMintRAiD_InvalidInput
TestMintRAiD_StorageError

TestFindAllRAiDs_Success
TestFindAllRAiDs_EmptyList
TestFindAllRAiDs_Pagination
TestFindAllRAiDs_StorageError

TestUpdateRAiD_Success
TestUpdateRAiD_NotFound
TestUpdateRAiD_InvalidInput
TestUpdateRAiD_StorageError

TestRAiDHistory_Success
TestRAiDHistory_NotFound
TestRAiDHistory_EmptyHistory

// Request Validation
TestHandlers_ContentTypeValidation
TestHandlers_JSONParsing
TestHandlers_URLParameterExtraction

// Response Formatting
TestHandlers_SuccessResponse
TestHandlers_ErrorResponse
TestHandlers_JSONEncoding
```

#### ServicePoint Handlers (`internal/handlers/servicepoint_test.go`)

**Test Categories:**
```go
// HTTP Endpoints
TestCreateServicePoint_Success
TestCreateServicePoint_InvalidInput
TestCreateServicePoint_StorageError

TestFindAllServicePoints_Success
TestFindAllServicePoints_Filtering

TestUpdateServicePoint_Success
TestUpdateServicePoint_NotFound
```

### 3. Configuration

#### Config Tests (`internal/config/config_test.go`)

**Test Categories:**
```go
// Configuration Loading
TestConfig_LoadFromEnv
TestConfig_DefaultValues
TestConfig_StorageBackendSelection

// Validation
TestConfig_RequiredFields
TestConfig_InvalidStorageType
TestConfig_FileStorageConfig
TestConfig_FDBStorageConfig
TestConfig_CockroachStorageConfig

// Environment Variables
TestConfig_EnvParsing
TestConfig_BooleanConversion
TestConfig_IntegerConversion
```

## Test Infrastructure

### Test Utilities Package (`internal/storage/testutil/`)

```go
package testutil

// Mock storage implementation for testing handlers
type MockRepository struct {
    // Methods that can be configured to return specific values/errors
    CreateRAiDFunc func(...) error
    GetRAiDFunc    func(...) (*models.RAiD, error)
    // ... all other methods
}

// Test data fixtures
func NewTestRAiD(id string) *models.RAiD
func NewTestServicePoint(id string) *models.ServicePoint

// Assertions
func AssertRAiDEqual(t *testing.T, expected, actual *models.RAiD)
func AssertError(t *testing.T, err error, expectedMsg string)

// Test helpers
func CreateTempDirectory(t *testing.T) string
func CleanupTestData(t *testing.T, dir string)
```

### Table-Driven Tests

Use Go's table-driven test pattern extensively:

```go
func TestFileStorage_CreateRAiD(t *testing.T) {
    tests := []struct {
        name    string
        raid    *models.RAiD
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid raid",
            raid:    testutil.NewTestRAiD("raid-001"),
            wantErr: false,
        },
        {
            name:    "nil raid",
            raid:    nil,
            wantErr: true,
            errMsg:  "raid cannot be nil",
        },
        {
            name:    "empty identifier",
            raid:    &models.RAiD{Identifier: ""},
            wantErr: true,
            errMsg:  "identifier cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Subtests and Parallel Execution

```go
func TestFileStorage(t *testing.T) {
    t.Run("CRUD Operations", func(t *testing.T) {
        t.Parallel()
        
        t.Run("Create", func(t *testing.T) {
            // Create test
        })
        
        t.Run("Read", func(t *testing.T) {
            // Read test
        })
        
        t.Run("Update", func(t *testing.T) {
            // Update test
        })
        
        t.Run("Delete", func(t *testing.T) {
            // Delete test
        })
    })
}
```

## Integration Tests

### Handler Integration Tests (`tests/integration/handlers_test.go`)

```go
// Test complete HTTP request/response cycle
func TestAPI_MintRAiD_EndToEnd(t *testing.T) {
    // Setup: Create test server with file storage
    storage := setupTestFileStorage(t)
    handler := setupTestHandler(storage)
    server := httptest.NewServer(handler)
    defer server.Close()

    // Execute: Make HTTP request
    resp := makeRequest(t, server.URL+"/raid", "POST", raidJSON)

    // Assert: Check response
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    // ... more assertions
}
```

### Storage Backend Integration Tests (`tests/integration/storage_test.go`)

```go
// Test with real databases (using testcontainers)
func TestCockroachDB_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Start CockroachDB container
    ctx := context.Background()
    container := startCockroachDBContainer(t, ctx)
    defer container.Terminate(ctx)

    // Run tests against real database
    storage := setupCockroachStorage(t, container)
    runStorageTests(t, storage)
}
```

## End-to-End Tests

### Complete Workflow Tests (`tests/e2e/workflows_test.go`)

```go
// Test complete RAiD lifecycle
func TestE2E_RAiDLifecycle(t *testing.T) {
    // 1. Start server with chosen storage backend
    // 2. Mint new RAiD
    // 3. Retrieve RAiD
    // 4. Update RAiD
    // 5. Get history
    // 6. Verify all operations
}

// Test migration between backends
func TestE2E_StorageMigration(t *testing.T) {
    // 1. Create data in file storage
    // 2. Export data
    // 3. Import to CockroachDB
    // 4. Verify data integrity
}
```

## Performance and Stress Tests

### Benchmark Tests (`*_bench_test.go`)

```go
func BenchmarkFileStorage_CreateRAiD(b *testing.B) {
    storage := setupTestStorage(b)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        raid := testutil.NewTestRAiD(fmt.Sprintf("raid-%d", i))
        storage.CreateRAiD(context.Background(), raid)
    }
}

func BenchmarkCockroachStorage_ListRAiDs(b *testing.B) {
    // Benchmark pagination performance
}
```

### Concurrency Tests

```go
func TestConcurrency_MultipleWrites(t *testing.T) {
    storage := setupTestStorage(t)
    
    var wg sync.WaitGroup
    numGoroutines := 100
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            raid := testutil.NewTestRAiD(fmt.Sprintf("raid-%d", id))
            err := storage.CreateRAiD(context.Background(), raid)
            assert.NoError(t, err)
        }(i)
    }
    
    wg.Wait()
    
    // Verify all records created
    raids, err := storage.ListRAiDs(context.Background(), 1000, 0)
    assert.NoError(t, err)
    assert.Len(t, raids, numGoroutines)
}
```

## Test Execution

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests with detailed coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run only unit tests (fast)
go test ./... -short

# Run integration tests (slow)
go test ./tests/integration/... -v

# Run specific package tests
go test ./internal/storage/file/...

# Run tests with race detection
go test ./... -race

# Run benchmarks
go test ./... -bench=. -benchmem

# Run tests in parallel
go test ./... -parallel=4

# Run with verbose output
go test ./... -v

# Run specific test
go test ./internal/storage/file -run TestFileStorage_CreateRAiD
```

### Test Tags

Use build tags to separate test types:

```go
// +build integration

package integration

// Integration tests here...
```

```bash
# Run integration tests only
go test ./... -tags=integration

# Run unit tests only (exclude integration)
go test ./... -short
```

## Continuous Integration

### GitHub Actions Workflow (`.github/workflows/test.yml`)

```yaml
name: Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.23.x, 1.24.x, 1.25.x]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: go test ./... -short -race -coverprofile=coverage.out
    
    - name: Upload coverage
      uses: codecov/codecov-action@v4
      with:
        files: ./coverage.out
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v6

  integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: cockroachdb/cockroach:latest
        env:
          COCKROACH_DATABASE: raid_test
        ports:
          - 26257:26257
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: 1.25.x
    
    - name: Run integration tests
      run: go test ./tests/integration/... -v
      env:
        DATABASE_URL: postgresql://root@localhost:26257/raid_test?sslmode=disable
```

## Mock Patterns

### Interface-Based Mocking

```go
// Storage interface already defined - easy to mock
type MockRAiDRepository struct {
    mock.Mock
}

func (m *MockRAiDRepository) CreateRAiD(ctx context.Context, raid *models.RAiD) error {
    args := m.Called(ctx, raid)
    return args.Error(0)
}

// Use in tests
func TestHandler_WithMock(t *testing.T) {
    mockRepo := new(MockRAiDRepository)
    mockRepo.On("CreateRAiD", mock.Anything, mock.Anything).Return(nil)
    
    handler := NewRAiDHandler(mockRepo)
    // Test handler...
    
    mockRepo.AssertExpectations(t)
}
```

### Dependency Injection for Testability

Already implemented in handlers - easy to inject mocks:

```go
handler := &RAiDHandler{
    repository: mockRepository, // Inject mock for testing
}
```

## Test Data Management

### Fixtures Directory (`tests/fixtures/`)

```
tests/fixtures/
├── raids/
│   ├── valid_raid.json
│   ├── invalid_raid.json
│   └── raid_with_all_fields.json
├── servicepoints/
│   ├── valid_servicepoint.json
│   └── multiple_servicepoints.json
└── responses/
    ├── success_response.json
    └── error_response.json
```

### Loading Fixtures

```go
func loadFixture(t *testing.T, filename string) []byte {
    data, err := os.ReadFile(filepath.Join("tests/fixtures", filename))
    require.NoError(t, err)
    return data
}
```

## Coverage Reporting

### Coverage Goals by Package

| Package | Target Coverage | Critical Paths |
|---------|----------------|----------------|
| `internal/storage/file` | 85% | 95% |
| `internal/storage/fdb` | 80% | 95% |
| `internal/storage/cockroach` | 85% | 95% |
| `internal/handlers` | 90% | 100% |
| `internal/config` | 95% | 100% |
| `internal/models` | 80% | N/A |

### Exclusions

Some code may be excluded from coverage:
- Error messages and logging
- Trivial getters/setters
- Generated code
- Main function (tested via E2E)

## Testing Checklist

Before merging any PR, ensure:

- [ ] All unit tests pass
- [ ] Integration tests pass (if applicable)
- [ ] Code coverage meets target (≥80%)
- [ ] No race conditions detected (`go test -race`)
- [ ] Benchmarks show acceptable performance
- [ ] New features have corresponding tests
- [ ] Error paths are tested
- [ ] Edge cases are covered
- [ ] Documentation is updated
- [ ] CI pipeline passes

## Test Maintenance

### Regular Review

- **Weekly**: Review failing tests and flaky tests
- **Monthly**: Review test coverage reports
- **Quarterly**: Review and update test strategy
- **Per Release**: Run full E2E test suite

### Test Debt

Track test debt in issues:
- Missing test coverage for edge cases
- Flaky tests that need fixing
- Slow tests that need optimization
- Integration tests that need real backends

## Tools and Libraries

### Testing Libraries

- **Standard Library**: `testing` package
- **Assertions**: `github.com/stretchr/testify/assert`
- **Mocking**: `github.com/stretchr/testify/mock`
- **HTTP Testing**: `net/http/httptest`
- **Test Containers**: `github.com/testcontainers/testcontainers-go`
- **Coverage**: Built-in Go coverage tools

### Optional Enhancements

- **`go-cmp`**: Deep equality comparison
- **`gomega`**: BDD-style matchers
- **`ginkgo`**: BDD testing framework (if team prefers)
- **`mockgen`**: Generate mocks from interfaces

## Success Metrics

Testing success measured by:

1. **Coverage**: ≥80% overall, ≥95% critical paths
2. **Reliability**: No flaky tests, consistent CI passes
3. **Speed**: Unit tests < 10s, integration tests < 2min
4. **Maintainability**: Clear, readable test code
5. **Bug Detection**: Catch issues before production
6. **Confidence**: Team confident in making changes

## Next Steps

1. ✅ Document test plan (this document)
2. ⬜ Set up test infrastructure and utilities
3. ⬜ Implement unit tests for file storage
4. ⬜ Implement unit tests for git storage
5. ⬜ Implement unit tests for FDB storage
6. ⬜ Implement unit tests for CockroachDB storage
7. ⬜ Implement handler tests with mocks
8. ⬜ Implement integration tests
9. ⬜ Set up CI/CD pipeline
10. ⬜ Achieve 80%+ coverage

---

**Ready to build a robust, well-tested RAiD service!** 🧪
