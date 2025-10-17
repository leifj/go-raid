package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/leifj/go-raid/internal/config"
)

// TestJWTAuth_Disabled tests that requests pass through when auth is disabled
func TestJWTAuth_Disabled(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled: false,
	}

	handler := JWTAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestJWTAuth_MissingToken tests that missing token returns 401
func TestJWTAuth_MissingToken(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled: true,
	}

	handler := JWTAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// TestJWTAuth_ValidToken tests that valid token is accepted
func TestJWTAuth_ValidToken(t *testing.T) {
	secret := "test-secret"
	cfg := &config.AuthConfig{
		Enabled:     true,
		JWTSecret:   secret,
		JWTIssuer:   "https://raid.org",
		JWTAudience: "raid-api",
	}

	// Create a valid token
	token := createTestToken(t, secret, "user123", "test@example.com", nil, []string{"admin"}, cfg.JWTIssuer, cfg.JWTAudience)

	var capturedUserID string
	handler := JWTAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, _ = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if capturedUserID != "user123" {
		t.Errorf("expected user ID 'user123', got '%s'", capturedUserID)
	}
}

// TestJWTAuth_InvalidToken tests that invalid token returns 401
func TestJWTAuth_InvalidToken(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:   true,
		JWTSecret: "test-secret",
	}

	handler := JWTAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// TestJWTAuth_WrongSecret tests that token with wrong secret is rejected
func TestJWTAuth_WrongSecret(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:   true,
		JWTSecret: "correct-secret",
	}

	// Create token with wrong secret
	token := createTestToken(t, "wrong-secret", "user123", "test@example.com", nil, []string{"admin"}, "", "")

	handler := JWTAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// TestJWTAuth_InvalidIssuer tests that token with wrong issuer is rejected
func TestJWTAuth_InvalidIssuer(t *testing.T) {
	secret := "test-secret"
	cfg := &config.AuthConfig{
		Enabled:   true,
		JWTSecret: secret,
		JWTIssuer: "https://raid.org",
	}

	// Create token with wrong issuer
	token := createTestToken(t, secret, "user123", "test@example.com", nil, []string{"admin"}, "https://wrong-issuer.org", "")

	handler := JWTAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// TestJWTAuth_InvalidAudience tests that token with wrong audience is rejected
func TestJWTAuth_InvalidAudience(t *testing.T) {
	secret := "test-secret"
	cfg := &config.AuthConfig{
		Enabled:     true,
		JWTSecret:   secret,
		JWTAudience: "raid-api",
	}

	// Create token with wrong audience
	token := createTestToken(t, secret, "user123", "test@example.com", nil, []string{"admin"}, "", "wrong-audience")

	handler := JWTAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// TestJWTAuth_WithServicePointID tests token with service point ID
func TestJWTAuth_WithServicePointID(t *testing.T) {
	secret := "test-secret"
	spID := int64(42)
	cfg := &config.AuthConfig{
		Enabled:   true,
		JWTSecret: secret,
	}

	token := createTestToken(t, secret, "user123", "test@example.com", &spID, []string{"admin"}, "", "")

	var capturedSPID int64
	var spIDFound bool
	handler := JWTAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedSPID, spIDFound = GetServicePointID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !spIDFound {
		t.Error("expected service point ID to be found in context")
	}

	if capturedSPID != spID {
		t.Errorf("expected service point ID %d, got %d", spID, capturedSPID)
	}
}

// TestExtractToken tests token extraction from Authorization header
func TestExtractToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantError bool
	}{
		{
			name:      "valid bearer token",
			header:    "Bearer abc123",
			wantToken: "abc123",
			wantError: false,
		},
		{
			name:      "lowercase bearer",
			header:    "bearer xyz789",
			wantToken: "xyz789",
			wantError: false,
		},
		{
			name:      "mixed case bearer",
			header:    "BeArEr test123",
			wantToken: "test123",
			wantError: false,
		},
		{
			name:      "missing header",
			header:    "",
			wantToken: "",
			wantError: true,
		},
		{
			name:      "invalid format - basic auth",
			header:    "Basic abc123",
			wantToken: "",
			wantError: true,
		},
		{
			name:      "missing token",
			header:    "Bearer",
			wantToken: "",
			wantError: true,
		},
		{
			name:      "only bearer word",
			header:    "Bearer ",
			wantToken: "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			token, err := extractToken(req)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if token != tt.wantToken {
				t.Errorf("expected token '%s', got '%s'", tt.wantToken, token)
			}
		})
	}
}

// TestContextHelpers tests the context helper functions
func TestContextHelpers(t *testing.T) {
	ctx := context.Background()
	spID := int64(123)

	// Test setting and getting values
	ctx = context.WithValue(ctx, UserIDKey, "user123")
	ctx = context.WithValue(ctx, UserEmailKey, "test@example.com")
	ctx = context.WithValue(ctx, ServicePointIDKey, spID)
	ctx = context.WithValue(ctx, RolesKey, []string{"admin", "user"})

	// Test GetUserID
	userID, ok := GetUserID(ctx)
	if !ok || userID != "user123" {
		t.Errorf("expected user ID 'user123', got '%s', ok=%v", userID, ok)
	}

	// Test GetUserEmail
	email, ok := GetUserEmail(ctx)
	if !ok || email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s', ok=%v", email, ok)
	}

	// Test GetServicePointID
	gotSPID, ok := GetServicePointID(ctx)
	if !ok || gotSPID != spID {
		t.Errorf("expected service point ID %d, got %d, ok=%v", spID, gotSPID, ok)
	}

	// Test GetRoles
	roles, ok := GetRoles(ctx)
	if !ok || len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d, ok=%v", len(roles), ok)
	}

	// Test missing values
	emptyCtx := context.Background()
	_, ok = GetUserID(emptyCtx)
	if ok {
		t.Error("expected ok=false for missing user ID")
	}

	_, ok = GetUserEmail(emptyCtx)
	if ok {
		t.Error("expected ok=false for missing email")
	}

	_, ok = GetServicePointID(emptyCtx)
	if ok {
		t.Error("expected ok=false for missing service point ID")
	}

	_, ok = GetRoles(emptyCtx)
	if ok {
		t.Error("expected ok=false for missing roles")
	}
}

// TestRequireRole tests the role-based middleware
func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		userRoles      []string
		requiredRole   string
		expectedStatus int
	}{
		{
			name:           "user has required role",
			userRoles:      []string{"admin", "user"},
			requiredRole:   "admin",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user missing required role",
			userRoles:      []string{"user"},
			requiredRole:   "admin",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "user has no roles",
			userRoles:      []string{},
			requiredRole:   "admin",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "user has multiple roles including required",
			userRoles:      []string{"user", "editor", "admin"},
			requiredRole:   "editor",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireRole(tt.requiredRole)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), RolesKey, tt.userRoles)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestRequireRole_NoRolesInContext tests RequireRole when no roles in context
func TestRequireRole_NoRolesInContext(t *testing.T) {
	handler := RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// TestValidateJWT tests JWT validation logic
func TestValidateJWT(t *testing.T) {
	secret := "test-secret"
	
	tests := []struct {
		name      string
		setupFunc func() (string, *config.AuthConfig)
		wantError bool
	}{
		{
			name: "valid token with all claims",
			setupFunc: func() (string, *config.AuthConfig) {
				cfg := &config.AuthConfig{
					JWTSecret:   secret,
					JWTIssuer:   "https://raid.org",
					JWTAudience: "raid-api",
				}
				token := createTestToken(t, secret, "user123", "test@example.com", nil, []string{"admin"}, cfg.JWTIssuer, cfg.JWTAudience)
				return token, cfg
			},
			wantError: false,
		},
		{
			name: "valid token without issuer validation",
			setupFunc: func() (string, *config.AuthConfig) {
				cfg := &config.AuthConfig{
					JWTSecret: secret,
				}
				token := createTestToken(t, secret, "user123", "test@example.com", nil, []string{"admin"}, "", "")
				return token, cfg
			},
			wantError: false,
		},
		{
			name: "expired token",
			setupFunc: func() (string, *config.AuthConfig) {
				cfg := &config.AuthConfig{
					JWTSecret: secret,
				}
				// Create an expired token
				claims := Claims{
					UserID: "user123",
					Email:  "test@example.com",
					Roles:  []string{"admin"},
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secret))
				return tokenString, cfg
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, cfg := tt.setupFunc()
			claims, err := validateJWT(tokenString, cfg)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantError && claims == nil {
				t.Error("expected claims, got nil")
			}
		})
	}
}

// createTestToken creates a JWT token for testing
func createTestToken(t *testing.T, secret, userID, email string, servicePointID *int64, roles []string, issuer, audience string) string {
	claims := Claims{
		UserID:         userID,
		Email:          email,
		ServicePointID: servicePointID,
		Roles:          roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	if issuer != "" {
		claims.Issuer = issuer
	}

	if audience != "" {
		claims.Audience = []string{audience}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	return tokenString
}
