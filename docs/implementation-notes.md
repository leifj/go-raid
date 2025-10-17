# RAiD Implementation Notes

## Overview

This document contains notes about the RAiD specification and implementation decisions.

## RAiD Identifier Structure

RAiD identifiers follow the Handle System format:
- Format: `https://raid.org/{prefix}/{suffix}`
- Example: `https://raid.org/10.25.1.1/abcde`
- Prefix: Assigned to service points (e.g., 10.25.1.1)
- Suffix: Unique identifier within the prefix

## Key Concepts

### Service Point
- Analogous to DataCite "Repository"
- Belongs to an Owner (legal entity)
- Has a prefix for minting RAiDs
- Can have multiple per Owner
- Controls permissions for minting/updating RAiDs

### Owner
- Legal entity responsible for RAiDs
- Analogous to DataCite "Member"
- Has legal agreement with Registration Agency
- Identified by ROR (Research Organization Registry) ID

### Registration Agency
- Organization operating the RAiD registration agency software
- Identified by ROR ID
- Responsible for minting and managing RAiDs

## Metadata Schema

### Required Fields for RAiD Creation
- `title` (array) - At least one title with text, type, startDate
- `dates` - Start date required, end date optional
- `contributors` (array) - People involved in the research
- `access` - Access type (open, embargoed, closed)

### Optional Fields
- `description` - Textual descriptions
- `alternateUrl` - Additional URLs
- `organisation` - Organizations involved
- `subject` - Research subjects/topics
- `relatedRaid` - Links to other RAiDs
- `relatedObject` - Links to other research objects
- `alternateIdentifier` - Other identifiers for the same research
- `spatialCoverage` - Geographic coverage
- `traditionalKnowledgeLabel` - Indigenous knowledge labels

## Access Control

Three access levels:
1. **Open**: Publicly accessible
2. **Embargoed**: Accessible after embargo expiry date (max 18 months)
3. **Closed**: Restricted access with access statement

Closed/Embargoed RAiDs return 403 with access statement when accessed.

## Versioning

- RAiDs are versioned
- Version increments on each update
- Full history available via JSON Patch (RFC 6902) format
- Can retrieve specific versions

## API Operations

### Core Operations
- **Mint**: Create new RAiD (POST /raid/)
- **Update**: Full update (PUT /raid/{prefix}/{suffix})
- **Patch**: Partial update (PATCH /raid/{prefix}/{suffix})
- **Read**: Get current version (GET /raid/{prefix}/{suffix})
- **List**: Query RAiDs with filters (GET /raid/)
- **History**: Get change history (GET /raid/{prefix}/{suffix}/history)

### Filtering
- By contributor ID (ORCID)
- By organization ID (ROR)
- Field selection via includeFields parameter

## Implementation Priorities

### Phase 1: Core Infrastructure ✅
- [x] Project structure
- [x] Data models
- [x] HTTP server and routing
- [x] Configuration management

### Phase 2: Data Layer (Current)
- [ ] PostgreSQL integration
- [ ] Schema design
- [ ] Repository pattern implementation
- [ ] Migration scripts

### Phase 3: Business Logic
- [ ] RAiD identifier generation
- [ ] Validation logic
- [ ] Versioning implementation
- [ ] History tracking (JSON Patch)

### Phase 4: Security
- [ ] JWT authentication
- [ ] Service point authorization
- [ ] Permission checking

### Phase 5: Advanced Features
- [ ] Contributor verification
- [ ] OAuth2/OIDC integration
- [ ] Advanced querying
- [ ] Bulk operations

## Database Schema Design Notes

### Tables Needed

1. **service_points**
   - id, name, identifier_owner (ROR), prefix, group_id, etc.

2. **raids**
   - id, handle (prefix/suffix), current_version, created_at, updated_at
   - JSONB column for flexible metadata storage

3. **raid_versions**
   - raid_id, version, metadata (JSONB), created_at
   - Full snapshot of each version

4. **raid_changes**
   - raid_id, version, diff (JSON Patch), timestamp
   - Change tracking for history

5. **contributors**
   - id, raid_id, orcid, email, status, etc.
   - Separate for easier querying

6. **organisations**
   - id, raid_id, ror_id, role, etc.

### Indexing Strategy
- Index on handles for fast lookup
- Index on contributor ORCID for filtering
- Index on organization ROR for filtering
- GIN index on JSONB columns for metadata queries

## Questions/Gaps in Specification

1. How exactly is the suffix generated? Random? Sequential?
2. What is the exact contributor verification flow?
3. How are service point permissions managed?
4. What are the validation rules for dates (format flexibility)?
5. Handle minting - do we need Handle.net integration?

## Reference Implementation Observations

From analyzing the reference implementation (raid-au):
- Uses Spring Boot + Java
- Uses Keycloak for authentication
- Uses Flyway for database migrations
- Uses JOOQ for database access
- Integrates with DataCite for DOI-like functionality
- Uses AWS services (we're avoiding this)

## Differences from Reference Implementation

1. **No AWS dependencies**: Pure PostgreSQL, no AWS-specific services
2. **Simplified authentication**: JWT-based, optional external IdP
3. **Go ecosystem**: Different libraries and patterns
4. **Cleaner separation**: Clear internal package structure
5. **Container-first**: Designed for Kubernetes deployment
