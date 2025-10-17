# Architecture Decision Records

## ADR-001: Use Go as Implementation Language

**Status**: Accepted

**Context**: Need to implement RAiD service as a cleanroom implementation independent of the Java/Spring Boot reference implementation.

**Decision**: Use Go (Golang) for the implementation.

**Rationale**:
- Excellent performance and low resource usage
- Strong standard library and ecosystem
- Native concurrency support
- Easy deployment (single binary)
- Cloud-agnostic by design
- Strong typing and good tooling

**Consequences**:
- Need to manually implement business logic based on OpenAPI spec
- Different ecosystem than reference implementation
- Easier containerization and cloud deployment

---

## ADR-002: Use Chi for HTTP Routing

**Status**: Accepted

**Context**: Need a fast, idiomatic router for HTTP endpoints.

**Decision**: Use go-chi/chi for HTTP routing.

**Rationale**:
- Lightweight and idiomatic Go HTTP router
- Built on net/http standard library
- Excellent middleware support
- Good context handling
- Well-maintained and widely used

**Consequences**:
- Clean route definitions
- Easy middleware composition
- Compatible with standard http.Handler interface

---

## ADR-003: Use PostgreSQL for Data Persistence

**Status**: Accepted

**Context**: Need a reliable, scalable database for storing RAiD metadata.

**Decision**: Use PostgreSQL as the primary database.

**Rationale**:
- Mature, reliable, and well-understood
- Excellent JSON support for flexible metadata
- Strong consistency guarantees
- Good support for versioning patterns
- Cloud-agnostic (available on all major cloud providers)
- No AWS lock-in

**Consequences**:
- Need to design schema for RAiD storage
- Can leverage JSONB for flexible metadata
- Standard SQL tooling and migration tools available

---

## ADR-004: Environment-Based Configuration

**Status**: Accepted

**Context**: Need flexible configuration for different deployment environments.

**Decision**: Use environment variables for all configuration.

**Rationale**:
- 12-factor app best practice
- Works well with containers and orchestrators
- No config file management needed
- Easy to override in different environments
- Secure secret management via environment

**Consequences**:
- All configuration via environment variables
- Provide .env.example for development
- Document all configuration options

---

## ADR-005: Cleanroom Implementation Approach

**Status**: Accepted

**Context**: Implementing RAiD service independently of reference implementation.

**Decision**: Build entirely from OpenAPI specification without referencing reference implementation code.

**Rationale**:
- Ensures independent implementation
- Avoids license complications
- Forces understanding of specification
- Enables different architectural choices
- No dependency on Spring Boot/Java ecosystem

**Consequences**:
- Need to infer business logic from specification
- May discover specification gaps
- Can make different architectural decisions
- Fully independent codebase

---

## ADR-006: JWT-Based Authentication (Planned)

**Status**: Proposed

**Context**: Need authentication and authorization for API operations.

**Decision**: Use JWT tokens for authentication, with optional OAuth2/OIDC integration.

**Rationale**:
- Stateless authentication
- Standard approach for REST APIs
- Can integrate with existing identity providers
- Flexible for different deployment scenarios

**Consequences**:
- Need to implement JWT validation
- Service point permissions management
- Optional integration with Keycloak or other IdP
