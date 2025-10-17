# API Compatibility Analysis

This document analyzes the current go-RAiD implementation against the RAiD OpenAPI 3.0 specification (`raido-openapi-3.0.yaml`) to identify gaps, improvements, and missing pieces.

**Analysis Date**: October 17, 2025  
**OpenAPI Spec**: raido-openapi-3.0.yaml (v2.0.0)

## Executive Summary

### Current Status
- ✅ **Routing**: All API endpoints are defined and routed correctly
- ✅ **Data Models**: Core models match the OpenAPI schema structure
- ⚠️ **Authentication**: Not implemented (security requirement)
- ⚠️ **Validation**: Minimal validation exists
- ❌ **Request/Response Types**: Using generic RAiD model instead of specific request/response types
- ❌ **Error Handling**: Not following OpenAPI error schema
- ❌ **PATCH Support**: JSON Patch (RFC 6902) not implemented
- ⚠️ **Field Filtering**: `includeFields` parameter not implemented

### Priority Issues
1. **CRITICAL**: Authentication/Authorization (JWT Bearer tokens)
2. **CRITICAL**: Request validation against OpenAPI schema
3. **HIGH**: Proper request/response type separation
4. **HIGH**: Error response standardization
5. **MEDIUM**: JSON Patch implementation for PATCH endpoint
6. **MEDIUM**: Field filtering for GET endpoints

---

## 1. Authentication & Authorization

### OpenAPI Specification
```yaml
security:
  - bearerAuth: []

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

### Current Implementation
- ❌ No authentication middleware
- ❌ No JWT validation
- ❌ No bearer token extraction
- ❌ No authorization checks

### Required Implementation

1. **Create middleware package** (`internal/middleware/auth.go`)
   ```go
   func JWTAuthMiddleware(next http.Handler) http.Handler
   func extractToken(r *http.Request) (string, error)
   func validateJWT(token string) (*Claims, error)
   ```

2. **Add JWT library dependency**
   ```bash
   go get github.com/golang-jwt/jwt/v5
   ```

3. **Environment configuration**
   - `JWT_SECRET` or `JWT_PUBLIC_KEY`
   - `JWT_ISSUER`
   - `JWT_AUDIENCE`
   - `AUTH_ENABLED` (for development)

4. **Apply to all endpoints except** `/health`

### Estimated Effort
- **Time**: 2-3 days
- **Complexity**: Medium
- **Dependencies**: JWT library, key management

---

## 2. Request/Response Type Separation

### OpenAPI Specification
Defines distinct types:
- `RaidCreateRequest` - For POST /raid/
- `RaidUpdateRequest` - For PUT /raid/{prefix}/{suffix}
- `RaidPatchRequest` - For PATCH /raid/{prefix}/{suffix}
- `RaidDto` - For responses
- `ServicePointCreateRequest` - For POST /service-point/
- `ServicePointUpdateRequest` - For PUT /service-point/{id}
- `ServicePoint` - For responses

### Current Implementation
- ❌ Uses single `RAiD` model for all operations
- ❌ Uses single `ServicePoint` model for all operations
- ❌ No validation of required fields per operation

### Required Changes

1. **Create separate request models** (`internal/models/requests.go`)
   ```go
   type RAiDCreateRequest struct {
       // No Identifier field (server generates)
       Title        []Title       `json:"title" validate:"required,min=1"`
       Date         *Date         `json:"date" validate:"required"`
       Contributors []Contributor `json:"contributor" validate:"required,min=1"`
       Access       *Access       `json:"access" validate:"required"`
       // ... other fields
   }

   type RAiDUpdateRequest struct {
       Identifier   *Identifier   `json:"identifier" validate:"required"`
       Title        []Title       `json:"title" validate:"required,min=1"`
       // ... all fields including identifier
   }

   type RAiDPatchRequest struct {
       Contributor []Contributor `json:"contributor,omitempty"`
       // Only fields that can be patched
   }
   ```

2. **Create response models** (`internal/models/responses.go`)
   ```go
   type RAiDResponse struct {
       // Extends RAiDUpdateRequest
       RAiDUpdateRequest
       TraditionalKnowledgeLabel []TraditionalKnowledge `json:"traditionalKnowledgeLabel,omitempty"`
   }
   ```

3. **Update handlers** to use correct types

### Estimated Effort
- **Time**: 2 days
- **Complexity**: Low-Medium
- **Breaking**: Yes (API contract changes)

---

## 3. Input Validation

### OpenAPI Specification
All schemas have:
- Required fields
- Type constraints
- Format specifications
- Array min/max items
- Pattern matching

### Current Implementation
- ❌ No validation framework
- ❌ Only basic JSON decode error checking
- ❌ No required field validation
- ❌ No type/format validation

### Required Implementation

1. **Add validation library**
   ```bash
   go get github.com/go-playground/validator/v10
   ```

2. **Add validation tags to models**
   ```go
   type Title struct {
       Text      string    `json:"text" validate:"required,min=1"`
       Type      *IDSchema `json:"type" validate:"required"`
       StartDate string    `json:"startDate" validate:"required,datetime=2006-01-02"`
   }
   ```

3. **Create validation middleware/helper**
   ```go
   func ValidateStruct(v interface{}) error
   func ValidationErrorResponse(err error) ErrorResponse
   ```

4. **Validate in handlers before processing**

### Estimated Effort
- **Time**: 3-4 days (including all model updates)
- **Complexity**: Medium
- **Impact**: High (prevents invalid data)

---

## 4. Error Response Standardization

### OpenAPI Specification
```yaml
ValidationFailureResponse:
  properties:
    type: { type: string }
    title: { type: string }
    status: { type: integer }
    detail: { type: string }
    instance: { type: string }
    failures:
      type: array
      items:
        $ref: '#/components/schemas/ValidationFailure'
```

### Current Implementation
- ✅ Models defined correctly
- ❌ Handlers return plain text errors
- ❌ No structured error responses
- ❌ No validation failure details

### Required Changes

1. **Update error handling in handlers**
   ```go
   func sendValidationError(w http.ResponseWriter, failures []ValidationFailure) {
       resp := ValidationFailureResponse{
           Type:     "validation-error",
           Title:    "Validation Failed",
           Status:   400,
           Detail:   "Request validation failed",
           Instance: r.URL.Path,
           Failures: failures,
       }
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(http.StatusBadRequest)
       json.NewEncoder(w).Encode(resp)
   }
   ```

2. **Create error helper package** (`internal/errors/`)
   - `NotFoundError(resource string)`
   - `ValidationError(failures []ValidationFailure)`
   - `UnauthorizedError(message string)`
   - `InternalError(err error)`

3. **Update all handlers** to use standard errors

### Estimated Effort
- **Time**: 1-2 days
- **Complexity**: Low
- **Impact**: Medium (better API usability)

---

## 5. JSON Patch Implementation (PATCH Endpoint)

### OpenAPI Specification
```yaml
/raid/{prefix}/{suffix}:
  patch:
    operationId: patchRaid
    requestBody:
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/RaidPatchRequest'
```

### Current Implementation
```go
func (h *RAiDHandler) PatchRAiD(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement JSON Patch (RFC 6902) support
    w.WriteHeader(http.StatusNotImplemented)
}
```

### Required Implementation

1. **Add JSON Patch library**
   ```bash
   go get github.com/evanphx/json-patch/v5
   ```

2. **Implement PATCH handler**
   ```go
   func (h *RAiDHandler) PatchRAiD(w http.ResponseWriter, r *http.Request) {
       // 1. Get current RAiD
       current := h.storage.GetRAiD(ctx, prefix, suffix)
       
       // 2. Parse patch request
       var patchReq RaidPatchRequest
       json.NewDecoder(r.Body).Decode(&patchReq)
       
       // 3. Apply patch
       patch := createJSONPatch(current, patchReq)
       
       // 4. Validate patched result
       // 5. Update RAiD
   }
   ```

3. **Alternative: Use RFC 6902 directly**
   ```go
   // Accept JSON Patch documents
   var patchDoc jsonpatch.Patch
   patchDoc, err := jsonpatch.DecodePatch(r.Body)
   
   // Apply to current RAiD
   modified, err := patchDoc.Apply(currentJSON)
   ```

### Estimated Effort
- **Time**: 2-3 days
- **Complexity**: Medium
- **Priority**: Medium (endpoint exists but returns 501)

---

## 6. Field Filtering (includeFields Parameter)

### OpenAPI Specification
```yaml
/raid/:
  get:
    parameters:
      - name: includeFields
        description: The top level fields to include in each RAiD
        example: identifier,title,date
        in: query
        schema:
          type: array
          items:
            type: string
```

### Current Implementation
- ❌ Parameter not parsed
- ❌ No field filtering logic
- ✅ Full RAiD returned always

### Required Implementation

1. **Parse query parameter**
   ```go
   includeFields := r.URL.Query().Get("includeFields")
   if includeFields != "" {
       fields := strings.Split(includeFields, ",")
       raids = filterFields(raids, fields)
   }
   ```

2. **Implement field filtering**
   ```go
   func filterFields(raids []RAiD, fields []string) []map[string]interface{} {
       result := make([]map[string]interface{}, len(raids))
       for i, raid := range raids {
           filtered := make(map[string]interface{})
           for _, field := range fields {
               if val := getField(raid, field); val != nil {
                   filtered[field] = val
               }
           }
           result[i] = filtered
       }
       return result
   }
   ```

3. **Use reflection or struct tags** for field extraction

### Estimated Effort
- **Time**: 1-2 days
- **Complexity**: Low-Medium
- **Impact**: Medium (performance optimization)

---

## 7. Query Parameter Filtering

### OpenAPI Specification
```yaml
parameters:
  - name: contributor.id
    description: Only show RAiDs that include a contributor with the given id
  - name: organisation.id
    description: Only show RAiDs that include an organisation
  - name: organisation.role
  - name: contributor.role
```

### Current Implementation
- ✅ `contributor.id` and `organisation.id` parsed
- ❌ Not actually used in filtering
- ❌ Role-based filtering not implemented

### Required Changes

1. **Extend RAiDFilter struct**
   ```go
   type RAiDFilter struct {
       ContributorID     string
       ContributorRole   string
       OrganisationID    string
       OrganisationRole  string
       Limit            int
       Offset           int
   }
   ```

2. **Implement filtering in storage layer**
   - File storage: Filter in memory
   - SQL storage: Add WHERE clauses
   - NoSQL: Query filters

3. **Update handlers** to pass all filters

### Estimated Effort
- **Time**: 2-3 days (across all storage backends)
- **Complexity**: Medium
- **Impact**: High (essential for API usability)

---

## 8. Access Control & Closed RAiDs

### OpenAPI Specification
```yaml
/raid/{prefix}/{suffix}:
  get:
    responses:
      403:
        description: Closed or Embargoed raids return a 403 response
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ClosedRaid'
```

### Current Implementation
- ✅ Access model exists
- ❌ No access control enforcement
- ❌ No 403 responses for closed RAiDs
- ❌ No embargo date checking

### Required Implementation

1. **Add access control logic**
   ```go
   func (h *RAiDHandler) checkAccess(raid *RAiD, userID string) error {
       if raid.Access.Type.ID == "closed" {
           // Return ClosedRaid response
           return ErrAccessDenied
       }
       
       if raid.Access.EmbargoExpiry != "" {
           expiryDate := parseDate(raid.Access.EmbargoExpiry)
           if time.Now().Before(expiryDate) {
               // Check if user is authorized
               if !isAuthorized(raid, userID) {
                   return ErrAccessDenied
               }
           }
       }
       return nil
   }
   ```

2. **Return ClosedRaid for forbidden access**
   ```go
   type ClosedRaid struct {
       Identifier *Identifier `json:"identifier"`
       Access     *Access     `json:"access"`
   }
   ```

### Estimated Effort
- **Time**: 2 days
- **Complexity**: Low-Medium
- **Dependencies**: Authentication system

---

## 9. Model Field Naming Discrepancies

### Issues Found

1. **Language field naming**
   - OpenAPI: `code` (in required fields)
   - Current: `id`
   
   **Fix**: Rename to match spec
   ```go
   type Language struct {
       Code      string `json:"code"`      // Was: ID
       SchemaURI string `json:"schemaUri"`
   }
   ```

2. **ServicePoint password field**
   - OpenAPI: Only in `ServicePointCreateRequest`
   - Current: Not defined
   
   **Fix**: Add to create request only (never return)

3. **Metadata timestamps**
   - OpenAPI: `number` format `datetime`
   - Current: `time.Time`
   
   **Fix**: Use Unix timestamp or ensure JSON serialization matches

### Estimated Effort
- **Time**: 0.5 days
- **Complexity**: Low
- **Breaking**: Yes (field names change)

---

## 10. Missing Features

### 10.1 Version-specific RAiD Retrieval

**Endpoint**: `GET /raid/{prefix}/{suffix}/{version}`

- ✅ Route exists
- ✅ Handler exists
- ⚠️ Storage implementation may be incomplete
- ❌ No tests

**Action**: Verify storage backends return versioned data correctly

### 10.2 RAiD History (Change Log)

**Endpoint**: `GET /raid/{prefix}/{suffix}/history`

- ✅ Route exists
- ✅ Handler exists
- ✅ Storage interface defined
- ⚠️ Returns `RAiDChange` array
- ❌ `diff` field should be base64 encoded JSON Patch

**Required**:
```go
type RAiDChange struct {
    Handle    string `json:"handle"`
    Version   int    `json:"version"`
    Diff      string `json:"diff"` // Base64(JSON Patch RFC 6902)
    Timestamp string `json:"timestamp"` // ISO 8601
}
```

### 10.3 Service Point Password Handling

**OpenAPI**: Password only in create request, never returned

**Current**: No password field defined

**Required**:
1. Add password to create/update requests
2. Never include in responses
3. Hash before storing
4. Validate strength

### 10.4 Service Point Search

**OpenAPI**: `searchContent` field exists

**Current**: Field exists but not populated/used

**Required**: Define what populates this field (auto-generated search index?)

---

## 11. Configuration & Environment Variables

### Missing Configuration

1. **Authentication**
   - `JWT_SECRET` / `JWT_PUBLIC_KEY`
   - `JWT_ISSUER`
   - `JWT_AUDIENCE`
   - `AUTH_ENABLED`

2. **Handle/Identifier Generation**
   - `HANDLE_PREFIX` (e.g., "10.82481")
   - `HANDLE_BASE_URL` (e.g., "https://raid.org.au")
   - `REGISTRATION_AGENCY_ID`
   - `REGISTRATION_AGENCY_SCHEMA_URI`

3. **Service Configuration**
   - `SERVICE_BASE_URL`
   - `API_VERSION`

### Update Required
Add to `internal/config/config.go`:
```go
type Config struct {
    Server  ServerConfig
    Storage StorageConfig
    Auth    AuthConfig      // NEW
    Handle  HandleConfig    // NEW
}

type AuthConfig struct {
    Enabled       bool   `env:"AUTH_ENABLED" default:"true"`
    JWTSecret     string `env:"JWT_SECRET"`
    JWTPublicKey  string `env:"JWT_PUBLIC_KEY"`
    JWTIssuer     string `env:"JWT_ISSUER"`
    JWTAudience   string `env:"JWT_AUDIENCE"`
}

type HandleConfig struct {
    Prefix              string `env:"HANDLE_PREFIX" required:"true"`
    BaseURL             string `env:"HANDLE_BASE_URL" default:"https://raid.org"`
    RegistrationAgency  string `env:"REGISTRATION_AGENCY_ID" required:"true"`
    RegistrationSchemaURI string `env:"REGISTRATION_AGENCY_SCHEMA_URI"`
}
```

---

## 12. Data Validation Requirements from OpenAPI

### Required Field Validation

All these need `validate:"required"` tags:

**RaidCreateRequest**:
- `title` (min 1 item)
- `date`
- `contributor` (min 1 item)
- `access`

**RaidUpdateRequest**: All above PLUS
- `identifier`

**Title**:
- `text`
- `type`
- `startDate`

**Date**:
- `startDate`

**Access**:
- `type`

**Contributor**:
- `id`
- `schemaUri`
- `position` (array)
- `role` (array)

**Organisation**:
- `id`
- `schemaUri`
- `role` (min 1 item)

**OrganisationRole**:
- `id`
- `schemaUri`
- `startDate`

**ContributorPosition**:
- `schemaUri`
- `id`
- `startDate`

**Identifier**:
- `id`
- `schemaUri`
- `registrationAgency`
- `owner`
- `license`
- `version`

---

## 13. Testing Gaps

### OpenAPI Specification Coverage

Current test coverage: **36.3%** (handlers only)

**Missing**:
1. ❌ OpenAPI schema validation tests
2. ❌ Request/response contract tests
3. ❌ Authentication tests
4. ❌ Error response format tests
5. ❌ Field filtering tests
6. ❌ Query parameter filtering tests
7. ❌ Access control tests
8. ❌ Version retrieval tests
9. ❌ History/change log tests

### Recommended Testing Strategy

1. **Contract Testing** (OpenAPI validation)
   ```bash
   go get github.com/getkin/kin-openapi/openapi3
   ```
   
   Validate all requests/responses against schema

2. **Integration Tests** with real storage

3. **E2E Tests** with authentication

4. **Target Coverage**: 80%+

---

## 14. Implementation Priority & Roadmap

### Phase 1: Critical (Production Blockers)
**Duration**: 2-3 weeks

1. ✅ **Authentication & Authorization** (3 days)
   - JWT middleware
   - Token validation
   - Authorization checks

2. ✅ **Input Validation** (4 days)
   - Validation framework
   - Model tags
   - Error responses

3. ✅ **Request/Response Types** (2 days)
   - Separate create/update/response models
   - Update handlers

4. ✅ **Error Standardization** (2 days)
   - Error helper package
   - Update all handlers

5. ✅ **Model Field Fixes** (0.5 days)
   - Language.code
   - Timestamp formats

### Phase 2: High Priority (API Completeness)
**Duration**: 2 weeks

1. ✅ **Query Filtering** (3 days)
   - Contributor/org role filtering
   - Implementation across storage backends

2. ✅ **Field Filtering** (2 days)
   - includeFields parameter
   - Dynamic field selection

3. ✅ **Access Control** (2 days)
   - Closed RAiD handling
   - Embargo date logic
   - 403 responses

4. ✅ **Configuration** (1 day)
   - Auth config
   - Handle generation config

5. ✅ **Service Point Password** (1 day)
   - Add to requests
   - Hashing
   - Never return

### Phase 3: Medium Priority (Enhanced Features)
**Duration**: 1-2 weeks

1. ✅ **JSON Patch** (3 days)
   - RFC 6902 implementation
   - PATCH handler

2. ✅ **RAiD History Enhancement** (2 days)
   - Base64 encoding
   - JSON Patch diffs

3. ✅ **Version Retrieval** (1 day)
   - Verify implementations
   - Add tests

### Phase 4: Testing & Documentation
**Duration**: 2 weeks

1. ✅ **OpenAPI Contract Tests** (3 days)
2. ✅ **Integration Tests** (3 days)
3. ✅ **E2E Tests with Auth** (3 days)
4. ✅ **API Documentation** (2 days)
5. ✅ **Migration Guide** (1 day)

---

## 15. Breaking Changes Summary

The following changes will break backward compatibility:

1. **Authentication Required**
   - All endpoints (except /health) require JWT

2. **Request/Response Types**
   - POST /raid/ accepts `RaidCreateRequest` (no identifier)
   - PUT /raid/ accepts `RaidUpdateRequest` (with identifier)

3. **Field Name Changes**
   - `Language.id` → `Language.code`

4. **Error Response Format**
   - Plain text → Structured `ValidationFailureResponse`

5. **Required Field Enforcement**
   - Requests without required fields will be rejected

6. **Access Control**
   - Closed/embargoed RAiDs return 403 instead of 200

### Migration Strategy

1. **Version the API** (v1 vs v2)
2. **Feature flag** for authentication
3. **Deprecation period** for old endpoints
4. **Clear migration documentation**

---

## 16. Recommendations

### Immediate Actions (This Week)

1. ✅ Create feature branch for OpenAPI compliance
2. ✅ Set up validation framework
3. ✅ Implement authentication middleware (with feature flag)
4. ✅ Fix model field naming issues
5. ✅ Add configuration for handle generation

### Short Term (Next 2 Weeks)

1. ✅ Implement request/response type separation
2. ✅ Add input validation to all endpoints
3. ✅ Standardize error responses
4. ✅ Implement query filtering
5. ✅ Add OpenAPI contract tests

### Medium Term (Next Month)

1. ✅ Complete JSON Patch implementation
2. ✅ Implement field filtering
3. ✅ Add access control enforcement
4. ✅ Achieve 80%+ test coverage
5. ✅ Complete API documentation

### Long Term (Next Quarter)

1. ✅ OpenAPI code generation (consider using openapi-generator)
2. ✅ API versioning strategy
3. ✅ Performance optimization
4. ✅ Rate limiting
5. ✅ API metrics and monitoring

---

## 17. Tools & Libraries Needed

### Required Dependencies

```bash
# Authentication
go get github.com/golang-jwt/jwt/v5

# Validation
go get github.com/go-playground/validator/v10

# JSON Patch
go get github.com/evanphx/json-patch/v5

# OpenAPI Validation (testing)
go get github.com/getkin/kin-openapi/openapi3

# Password Hashing
go get golang.org/x/crypto/bcrypt
```

### Development Tools

```bash
# OpenAPI validation
npm install -g @stoplight/spectral-cli

# API testing
go get github.com/stretchr/testify

# Mock generation
go install github.com/golang/mock/mockgen@latest
```

---

## 18. Success Criteria

The implementation will be considered complete when:

1. ✅ All endpoints require authentication (with bypass option)
2. ✅ All requests are validated against OpenAPI schema
3. ✅ All responses match OpenAPI schema
4. ✅ Error responses follow standard format
5. ✅ Query filtering works for all specified parameters
6. ✅ Field filtering works correctly
7. ✅ Access control is enforced
8. ✅ JSON Patch is implemented
9. ✅ Test coverage ≥ 80%
10. ✅ OpenAPI contract tests pass 100%
11. ✅ Documentation is complete
12. ✅ No OpenAPI linter warnings

---

## Conclusion

The current go-RAiD implementation has excellent foundational architecture with proper routing and storage abstraction. However, to be fully compatible with the RAiD OpenAPI 3.0 specification, significant work is needed in:

1. **Authentication** (critical blocker)
2. **Validation** (critical for data integrity)
3. **Type Safety** (request/response separation)
4. **Error Handling** (API usability)
5. **Access Control** (security requirement)

**Estimated Total Effort**: 8-10 weeks with one developer

**Recommended Approach**: Phased implementation with feature flags to allow gradual rollout and testing.

**Next Steps**: 
1. Review and prioritize this analysis
2. Create GitHub issues for each major item
3. Begin Phase 1 implementation
4. Set up CI/CD pipeline for API contract testing
