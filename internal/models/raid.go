package models

import "time"

// RAiD represents a Research Activity Identifier
type RAiD struct {
	Metadata             *Metadata              `json:"metadata,omitempty"`
	Identifier           *Identifier            `json:"identifier"`
	Title                []Title                `json:"title"`
	Date                 *Date                  `json:"date"`
	Description          []Description          `json:"description,omitempty"`
	Access               *Access                `json:"access"`
	AlternateURL         []AlternateURL         `json:"alternateUrl,omitempty"`
	Contributor          []Contributor          `json:"contributor,omitempty"`
	Organisation         []Organisation         `json:"organisation,omitempty"`
	Subject              []Subject              `json:"subject,omitempty"`
	RelatedRAiD          []RelatedRAiD          `json:"relatedRaid,omitempty"`
	RelatedObject        []RelatedObject        `json:"relatedObject,omitempty"`
	AlternateIdentifier  []AlternateIdentifier  `json:"alternateIdentifier,omitempty"`
	SpatialCoverage      []SpatialCoverage      `json:"spatialCoverage,omitempty"`
	TraditionalKnowledge []TraditionalKnowledge `json:"traditionalKnowledgeLabel,omitempty"`
}

// Metadata contains timestamps for RAiD creation and updates
type Metadata struct {
	Created time.Time `json:"created,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

// Identifier represents the RAiD identifier with all its components
type Identifier struct {
	ID                 string              `json:"id"`
	SchemaURI          string              `json:"schemaUri"`
	RegistrationAgency *RegistrationAgency `json:"registrationAgency"`
	Owner              *Owner              `json:"owner"`
	RAIDAgencyURL      string              `json:"raidAgencyUrl,omitempty"`
	License            string              `json:"license"`
	Version            int                 `json:"version"`
}

// RegistrationAgency identifies the organisation operating the RAiD registration agency
type RegistrationAgency struct {
	ID        string `json:"id"`
	SchemaURI string `json:"schemaUri"`
}

// Owner represents the legal entity responsible for the RAiD
type Owner struct {
	ID           string `json:"id"`
	SchemaURI    string `json:"schemaUri"`
	ServicePoint int64  `json:"servicePoint,omitempty"`
}

// Title represents a title with type, dates, and optional language
type Title struct {
	Text      string    `json:"text"`
	Type      *IDSchema `json:"type"`
	StartDate string    `json:"startDate"`
	EndDate   string    `json:"endDate,omitempty"`
	Language  *Language `json:"language,omitempty"`
}

// Date contains start and end dates for the research activity
type Date struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate,omitempty"`
}

// Description represents a description with type and optional language
type Description struct {
	Text     string    `json:"text"`
	Type     *IDSchema `json:"type"`
	Language *Language `json:"language,omitempty"`
}

// Access defines the access type and optional embargo information
type Access struct {
	Type          *IDSchema        `json:"type"`
	Statement     *AccessStatement `json:"statement,omitempty"`
	EmbargoExpiry string           `json:"embargoExpiry,omitempty"`
}

// AccessStatement provides textual access statement with optional language
type AccessStatement struct {
	Text     string    `json:"text"`
	Language *Language `json:"language,omitempty"`
}

// Contributor represents a person contributing to the research activity
type Contributor struct {
	ID            string                `json:"id"`
	SchemaURI     string                `json:"schemaUri"`
	Status        string                `json:"status,omitempty"`
	StatusMessage string                `json:"statusMessage,omitempty"`
	Email         string                `json:"email,omitempty"`
	UUID          string                `json:"uuid,omitempty"`
	Position      []ContributorPosition `json:"position"`
	Role          []IDSchema            `json:"role"`
	Leader        bool                  `json:"leader,omitempty"`
	Contact       bool                  `json:"contact,omitempty"`
}

// ContributorPosition represents a contributor's position with dates
type ContributorPosition struct {
	SchemaURI string `json:"schemaUri"`
	ID        string `json:"id"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate,omitempty"`
}

// Organisation represents an organisation involved in the research
type Organisation struct {
	ID        string             `json:"id"`
	SchemaURI string             `json:"schemaUri"`
	Role      []OrganisationRole `json:"role"`
}

// OrganisationRole represents an organisation's role with dates
type OrganisationRole struct {
	SchemaURI string `json:"schemaUri"`
	ID        string `json:"id"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate,omitempty"`
}

// AlternateURL represents an alternate URL for the RAiD
type AlternateURL struct {
	URL string `json:"url"`
}

// Subject represents a subject/topic with optional keywords
type Subject struct {
	ID        string           `json:"id"`
	SchemaURI string           `json:"schemaUri"`
	Keyword   []SubjectKeyword `json:"keyword,omitempty"`
}

// SubjectKeyword represents a keyword with optional language
type SubjectKeyword struct {
	Text     string    `json:"text"`
	Language *Language `json:"language,omitempty"`
}

// RelatedRAiD represents a related RAiD with relationship type
type RelatedRAiD struct {
	ID   string    `json:"id"`
	Type *IDSchema `json:"type,omitempty"`
}

// RelatedObject represents a related object with type and categories
type RelatedObject struct {
	ID        string     `json:"id"`
	SchemaURI string     `json:"schemaUri,omitempty"`
	Type      *IDSchema  `json:"type,omitempty"`
	Category  []IDSchema `json:"category,omitempty"`
}

// AlternateIdentifier represents an alternate identifier
type AlternateIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type,omitempty"`
}

// SpatialCoverage represents spatial coverage with places
type SpatialCoverage struct {
	ID        string                 `json:"id"`
	SchemaURI string                 `json:"schemaUri,omitempty"`
	Place     []SpatialCoveragePlace `json:"place,omitempty"`
}

// SpatialCoveragePlace represents a place with optional language
type SpatialCoveragePlace struct {
	Text     string    `json:"text"`
	Language *Language `json:"language,omitempty"`
}

// TraditionalKnowledge represents traditional knowledge labels
type TraditionalKnowledge struct {
	ID        string `json:"id"`
	SchemaURI string `json:"schemaUri,omitempty"`
}

// Language represents an ISO 639-3 language code
type Language struct {
	ID        string `json:"id"`
	SchemaURI string `json:"schemaUri"`
}

// IDSchema is a generic type for identifier/schema pairs
type IDSchema struct {
	ID        string `json:"id"`
	SchemaURI string `json:"schemaUri"`
}

// ServicePoint represents a service point for minting RAiDs
type ServicePoint struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	IdentifierOwner  string `json:"identifierOwner"`
	RepositoryID     string `json:"repositoryId,omitempty"`
	Prefix           string `json:"prefix,omitempty"`
	GroupID          string `json:"groupId,omitempty"`
	SearchContent    string `json:"searchContent,omitempty"`
	TechEmail        string `json:"techEmail"`
	AdminEmail       string `json:"adminEmail"`
	Enabled          bool   `json:"enabled"`
	AppWritesEnabled bool   `json:"appWritesEnabled,omitempty"`
}

// RAiDChange represents a change to a RAiD
type RAiDChange struct {
	Handle    string    `json:"handle"`
	Version   int       `json:"version"`
	Diff      string    `json:"diff"` // Base64 encoded JSON Patch (RFC 6902)
	Timestamp time.Time `json:"timestamp"`
}

// ValidationFailure represents a validation error
type ValidationFailure struct {
	FieldID   string `json:"fieldId"`
	ErrorType string `json:"errorType"`
	Message   string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Type     string              `json:"type"`
	Title    string              `json:"title"`
	Status   int                 `json:"status"`
	Detail   string              `json:"detail"`
	Instance string              `json:"instance"`
	Failures []ValidationFailure `json:"failures,omitempty"`
}
