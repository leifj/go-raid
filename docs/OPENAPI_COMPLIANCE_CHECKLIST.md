# OpenAPI Compliance Checklist

Track progress towards full RAiD OpenAPI 3.0 specification compliance.

## 🔴 Critical Priority (Production Blockers)

### Authentication & Authorization
- [ ] Create `internal/middleware/auth.go`
- [ ] Implement JWT validation middleware
- [ ] Add bearer token extraction
- [ ] Add authorization checks
- [ ] Add `AUTH_ENABLED` feature flag
- [ ] Configure JWT secret/public key
- [ ] Apply middleware to all endpoints (except /health)
- [ ] Add authentication tests

**Dependencies**: `github.com/golang-jwt/jwt/v5`

### Input Validation  
- [ ] Add validation framework dependency
- [ ] Add validation tags to all models
- [ ] Create `ValidateStruct()` helper
- [ ] Validate RAiDCreateRequest required fields
- [ ] Validate RAiDUpdateRequest required fields
- [ ] Validate ServicePoint required fields
- [ ] Validate nested structures (Title, Date, etc.)
- [ ] Add validation tests

**Dependencies**: `github.com/go-playground/validator/v10`

### Request/Response Type Separation
- [ ] Create `internal/models/requests.go`
  - [ ] `RAiDCreateRequest` (no identifier)
  - [ ] `RAiDUpdateRequest` (with identifier)
  - [ ] `RAiDPatchRequest` (contributor only)
  - [ ] `ServicePointCreateRequest`
  - [ ] `ServicePointUpdateRequest`
- [ ] Create `internal/models/responses.go`
  - [ ] `RAiDResponse` (extends update + TK labels)
  - [ ] `ServicePointResponse`
- [ ] Update handlers to use new types
  - [ ] MintRAiD (use CreateRequest)
  - [ ] UpdateRAiD (use UpdateRequest)
  - [ ] PatchRAiD (use PatchRequest)
  - [ ] CreateServicePoint (use CreateRequest)
  - [ ] UpdateServicePoint (use UpdateRequest)
- [ ] Update tests for new types

### Error Response Standardization
- [ ] Create `internal/errors/errors.go`
  - [ ] `ValidationError()`
  - [ ] `NotFoundError()`
  - [ ] `UnauthorizedError()`
  - [ ] `ForbiddenError()`
  - [ ] `InternalError()`
- [ ] Update all handlers to use standard errors
  - [ ] MintRAiD
  - [ ] FindAllRAiDs
  - [ ] FindRAiDByName
  - [ ] UpdateRAiD
  - [ ] PatchRAiD
  - [ ] RAiDHistory
  - [ ] ServicePoint handlers
- [ ] Return proper HTTP status codes
- [ ] Include `ValidationFailureResponse` for 400 errors
- [ ] Add error response tests

### Model Field Corrections
- [ ] Fix `Language` struct
  - [ ] Rename `ID` → `Code`
  - [ ] Update all usages
  - [ ] Update JSON tags
- [ ] Fix `Metadata` timestamps
  - [ ] Verify JSON serialization format
  - [ ] Test Unix timestamp vs ISO 8601
- [ ] Add `Password` field to ServicePointCreateRequest
  - [ ] Never include in ServicePointResponse
  - [ ] Add password hashing
- [ ] Update tests for field changes

---

## 🟡 High Priority (API Completeness)

### Query Parameter Filtering
- [ ] Extend `RAiDFilter` struct
  - [ ] `ContributorRole string`
  - [ ] `OrganisationRole string`
- [ ] Parse additional query parameters
  - [ ] `contributor.role`
  - [ ] `organisation.role`
- [ ] Implement filtering in storage backends
  - [ ] File storage backend
  - [ ] Git storage backend  
  - [ ] CockroachDB backend (SQL WHERE)
  - [ ] FoundationDB backend
- [ ] Add query filter tests

### Field Filtering (includeFields)
- [ ] Parse `includeFields` query parameter
- [ ] Implement `filterFields()` helper
- [ ] Use reflection or struct tags for field extraction
- [ ] Support nested field filtering
- [ ] Handle invalid field names gracefully
- [ ] Add field filtering tests
- [ ] Performance test with large result sets

### Access Control
- [ ] Implement `checkAccess()` logic
  - [ ] Check for closed RAiDs
  - [ ] Check embargo dates
  - [ ] Validate user authorization
- [ ] Return 403 with `ClosedRaid` schema
  - [ ] Include identifier only
  - [ ] Include access information
- [ ] Implement embargo expiry checking
  - [ ] Parse embargo date
  - [ ] Compare with current time
- [ ] Add owner/authorized user checks
- [ ] Add access control tests

### Configuration Enhancements
- [ ] Add `AuthConfig` struct
  - [ ] `JWTSecret`
  - [ ] `JWTPublicKey`
  - [ ] `JWTIssuer`
  - [ ] `JWTAudience`
  - [ ] `Enabled`
- [ ] Add `HandleConfig` struct
  - [ ] `Prefix`
  - [ ] `BaseURL`
  - [ ] `RegistrationAgencyID`
  - [ ] `RegistrationAgencySchemaURI`
- [ ] Update config loading
- [ ] Add config validation
- [ ] Update documentation

### Service Point Enhancements
- [ ] Add password handling
  - [ ] Accept in create/update requests
  - [ ] Hash with bcrypt
  - [ ] Never return in responses
  - [ ] Add password strength validation
- [ ] Implement `searchContent` population
  - [ ] Define search indexing strategy
  - [ ] Auto-generate from name + other fields
- [ ] Add service point tests

---

## 🟢 Medium Priority (Enhanced Features)

### JSON Patch Implementation
- [ ] Add JSON Patch library dependency
- [ ] Implement PatchRAiD handler
  - [ ] Fetch current RAiD
  - [ ] Parse patch request
  - [ ] Apply patch operations
  - [ ] Validate result
  - [ ] Update RAiD
- [ ] Support RFC 6902 operations
  - [ ] `add`
  - [ ] `remove`
  - [ ] `replace`
  - [ ] `move`
  - [ ] `copy`
  - [ ] `test`
- [ ] Add JSON Patch tests
- [ ] Remove "not implemented" response

**Dependencies**: `github.com/evanphx/json-patch/v5`

### RAiD History Enhancement
- [ ] Update `RAiDChange.Diff` format
  - [ ] Generate JSON Patch documents
  - [ ] Base64 encode patches
- [ ] Update `RAiDChange.Timestamp` format
  - [ ] Use ISO 8601 string format
- [ ] Verify storage implementations
  - [ ] File storage tracks changes
  - [ ] Git storage uses git log
  - [ ] SQL storage stores diffs
- [ ] Add history generation tests

### Version-Specific Retrieval
- [ ] Verify `GetRAiDVersion()` implementation
  - [ ] File storage version lookup
  - [ ] Git storage checkout specific version
  - [ ] SQL storage version query
- [ ] Test version retrieval
  - [ ] Valid version
  - [ ] Invalid version (404)
  - [ ] Version out of range
- [ ] Add integration tests

---

## 🔵 Testing & Quality

### OpenAPI Contract Testing
- [ ] Add OpenAPI validation dependency
- [ ] Create contract test suite
  - [ ] Validate request schemas
  - [ ] Validate response schemas
  - [ ] Validate error schemas
- [ ] Test all endpoints
  - [ ] POST /raid/
  - [ ] GET /raid/
  - [ ] GET /raid/{prefix}/{suffix}
  - [ ] PUT /raid/{prefix}/{suffix}
  - [ ] PATCH /raid/{prefix}/{suffix}
  - [ ] GET /raid/{prefix}/{suffix}/{version}
  - [ ] GET /raid/{prefix}/{suffix}/history
  - [ ] GET /raid/all-public
  - [ ] Service point endpoints
- [ ] Add to CI/CD pipeline
- [ ] Achieve 100% contract test pass rate

**Dependencies**: `github.com/getkin/kin-openapi/openapi3`

### Integration Testing
- [ ] Test with real file storage
- [ ] Test with real CockroachDB
- [ ] Test with real FoundationDB
- [ ] Test authentication flow
- [ ] Test access control scenarios
- [ ] Test query filtering
- [ ] Test field filtering
- [ ] Test version history

### End-to-End Testing
- [ ] Set up E2E test environment
- [ ] Test complete RAiD lifecycle
  - [ ] Create
  - [ ] Read
  - [ ] Update
  - [ ] Patch
  - [ ] History
  - [ ] Version retrieval
- [ ] Test service point lifecycle
- [ ] Test authentication scenarios
- [ ] Test authorization scenarios
- [ ] Test error cases

### Test Coverage Goals
- [ ] Overall coverage ≥ 80%
- [ ] Handler coverage ≥ 90%
- [ ] Model validation ≥ 95%
- [ ] Storage backends ≥ 85%
- [ ] Middleware ≥ 90%

---

## 📚 Documentation

### API Documentation
- [ ] Update README with authentication info
- [ ] Document environment variables
  - [ ] Authentication config
  - [ ] Handle generation config
- [ ] Create API usage examples
  - [ ] cURL examples with JWT
  - [ ] Postman collection
  - [ ] OpenAPI UI setup
- [ ] Document error responses
- [ ] Document query parameters
- [ ] Document field filtering

### Developer Documentation
- [ ] Authentication setup guide
- [ ] Validation rules documentation
- [ ] Access control documentation
- [ ] Testing guide
- [ ] Migration guide (v1 → v2)
- [ ] Deployment guide with auth

### Code Documentation
- [ ] Add godoc comments to all exported functions
- [ ] Document validation tags
- [ ] Document middleware
- [ ] Add inline comments for complex logic

---

## 🚀 Deployment & Operations

### CI/CD Pipeline
- [ ] Add OpenAPI linting
- [ ] Add contract tests to CI
- [ ] Add integration tests to CI
- [ ] Add E2E tests to CI
- [ ] Set up code coverage reporting
- [ ] Add security scanning

### Environment Setup
- [ ] Development environment
  - [ ] Mock JWT for testing
  - [ ] Auth disabled flag
- [ ] Staging environment
  - [ ] Real JWT validation
  - [ ] Test service point
- [ ] Production environment
  - [ ] Production JWT keys
  - [ ] Proper handle prefix
  - [ ] Monitoring & alerting

### Monitoring
- [ ] Add authentication metrics
- [ ] Add validation error metrics
- [ ] Add access control metrics
- [ ] Add API performance metrics
- [ ] Set up alerts for errors

---

## 📋 Quality Gates

Before considering OpenAPI compliance complete:

### Functional Requirements
- [ ] All endpoints require authentication (or flag)
- [ ] All requests validated against schema
- [ ] All responses match schema
- [ ] All required fields enforced
- [ ] All query parameters work
- [ ] Field filtering works
- [ ] Access control enforced
- [ ] JSON Patch implemented
- [ ] History returns correct format

### Quality Requirements
- [ ] Test coverage ≥ 80%
- [ ] OpenAPI contract tests 100% pass
- [ ] No critical security issues
- [ ] No OpenAPI linter warnings
- [ ] Performance benchmarks met
- [ ] Documentation complete

### Operational Requirements
- [ ] CI/CD pipeline includes all tests
- [ ] Deployment documentation complete
- [ ] Migration guide available
- [ ] Monitoring configured
- [ ] Security scanning in place

---

## 📅 Milestone Tracking

### Milestone 1: Authentication & Validation (Week 1-2)
- [ ] Authentication middleware
- [ ] Request validation
- [ ] Error standardization
- [ ] Model fixes

### Milestone 2: Type Safety & Filtering (Week 3-4)
- [ ] Request/response types
- [ ] Query filtering
- [ ] Field filtering
- [ ] Configuration updates

### Milestone 3: Advanced Features (Week 5-6)
- [ ] Access control
- [ ] JSON Patch
- [ ] History enhancement
- [ ] Service point completion

### Milestone 4: Testing & Documentation (Week 7-8)
- [ ] Contract tests
- [ ] Integration tests
- [ ] E2E tests
- [ ] Documentation
- [ ] CI/CD updates

### Milestone 5: Production Ready (Week 9-10)
- [ ] Security review
- [ ] Performance testing
- [ ] Deployment guides
- [ ] Monitoring setup
- [ ] Final validation

---

## 🎯 Success Metrics

Track these metrics to measure progress:

- **API Compliance**: % of endpoints matching OpenAPI spec
- **Test Coverage**: % of code covered by tests
- **Contract Tests**: % passing OpenAPI validation
- **Documentation**: % of features documented
- **Performance**: Response time targets met
- **Security**: Zero critical vulnerabilities

Current Status:
- API Compliance: ~40%
- Test Coverage: 36.3%
- Contract Tests: 0%
- Documentation: 60%

Target Status:
- API Compliance: 100%
- Test Coverage: 80%+
- Contract Tests: 100%
- Documentation: 100%

---

## 📝 Notes

- Use feature flags for gradual rollout
- Maintain backward compatibility where possible
- Version breaking changes appropriately
- Keep OpenAPI spec as source of truth
- Regular sync meetings to track progress
- Update this checklist as work progresses

Last Updated: October 17, 2025
