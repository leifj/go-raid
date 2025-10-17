package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/leifj/go-raid/internal/models"
	"github.com/leifj/go-raid/internal/storage"
	"github.com/leifj/go-raid/internal/storage/testutil"
)

func TestNewRAiDHandler(t *testing.T) {
	repo := testutil.NewMockRepository()
	handler := NewRAiDHandler(repo)

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
	if handler.storage == nil {
		t.Fatal("Expected storage to be set")
	}
}

func TestMintRAiD_Success(t *testing.T) {
	// Setup mock repository
	repo := testutil.NewMockRepository()
	prefix, suffix := "10.12345", "67890"
	testRAiD := testutil.NewTestRAiD(prefix, suffix)

	repo.CreateRAiDFunc = func(ctx context.Context, raid *models.RAiD) (*models.RAiD, error) {
		// Simulate identifier generation by storage backend
		if raid.Identifier == nil {
			raid.Identifier = &models.Identifier{}
		}
		raid.Identifier.ID = fmt.Sprintf("https://raid.org/%s/%s", prefix, suffix)
		raid.Identifier.Version = 1
		return raid, nil
	}

	// Setup HTTP request
	requestBody := testRAiD
	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/raid", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	handler := NewRAiDHandler(repo)
	handler.MintRAiD(rr, req)

	// Assert
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	var response models.RAiD
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Identifier == nil || response.Identifier.ID == "" {
		t.Error("Expected identifier to be set in response")
	}

	// Verify calls
	if repo.CreateRAiDCalls != 1 {
		t.Errorf("Expected 1 CreateRAiD call, got %d", repo.CreateRAiDCalls)
	}
}

func TestMintRAiD_InvalidJSON(t *testing.T) {
	repo := testutil.NewMockRepository()

	// Invalid JSON in request body
	req := httptest.NewRequest(http.MethodPost, "/raid", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := NewRAiDHandler(repo)
	handler.MintRAiD(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	// Should not have called any repository methods
	if repo.CreateRAiDCalls != 0 {
		t.Errorf("Expected 0 CreateRAiD calls, got %d", repo.CreateRAiDCalls)
	}
}

func TestMintRAiD_RepositoryError(t *testing.T) {
	repo := testutil.NewMockRepository()
	testRAiD := testutil.NewTestRAiD("10.12345", "67890")

	repo.CreateRAiDFunc = func(ctx context.Context, raid *models.RAiD) (*models.RAiD, error) {
		return nil, fmt.Errorf("database error")
	}

	bodyBytes, _ := json.Marshal(testRAiD)
	req := httptest.NewRequest(http.MethodPost, "/raid", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := NewRAiDHandler(repo)
	handler.MintRAiD(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestFindAllRAiDs_Success(t *testing.T) {
	repo := testutil.NewMockRepository()

	// Mock data
	raids := []*models.RAiD{
		testutil.NewTestRAiD("10.12345", "00001"),
		testutil.NewTestRAiD("10.12345", "00002"),
		testutil.NewTestRAiD("10.12345", "00003"),
	}

	repo.ListRAiDsFunc = func(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
		return raids, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/raid?limit=10&offset=0", nil)
	rr := httptest.NewRecorder()

	handler := NewRAiDHandler(repo)
	handler.FindAllRAiDs(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response []*models.RAiD
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 3 {
		t.Errorf("Expected 3 RAiDs, got %d", len(response))
	}

	if repo.ListRAiDsCalls != 1 {
		t.Errorf("Expected 1 ListRAiDs call, got %d", repo.ListRAiDsCalls)
	}
}

func TestFindAllRAiDs_WithFilters(t *testing.T) {
	repo := testutil.NewMockRepository()

	repo.ListRAiDsFunc = func(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
		// Verify filter parameters
		if filter.Limit != 20 {
			t.Errorf("Expected limit 20, got %d", filter.Limit)
		}
		if filter.Offset != 10 {
			t.Errorf("Expected offset 10, got %d", filter.Offset)
		}
		return []*models.RAiD{}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/raid?limit=20&offset=10", nil)
	rr := httptest.NewRecorder()

	handler := NewRAiDHandler(repo)
	handler.FindAllRAiDs(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestFindAllRAiDs_RepositoryError(t *testing.T) {
	repo := testutil.NewMockRepository()

	repo.ListRAiDsFunc = func(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
		return nil, fmt.Errorf("database connection error")
	}

	req := httptest.NewRequest(http.MethodGet, "/raid", nil)
	rr := httptest.NewRecorder()

	handler := NewRAiDHandler(repo)
	handler.FindAllRAiDs(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestFindRAiDByName_Success(t *testing.T) {
	repo := testutil.NewMockRepository()
	prefix, suffix := "10.12345", "67890"
	testRAiD := testutil.NewTestRAiD(prefix, suffix)

	repo.GetRAiDFunc = func(ctx context.Context, p, s string) (*models.RAiD, error) {
		if p != prefix || s != suffix {
			t.Errorf("Expected prefix=%s suffix=%s, got prefix=%s suffix=%s", prefix, suffix, p, s)
		}
		return testRAiD, nil
	}

	// Setup chi router context
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/raid/%s/%s", prefix, suffix), nil)
	rr := httptest.NewRecorder()

	// Add URL parameters via chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("prefix", prefix)
	rctx.URLParams.Add("suffix", suffix)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := NewRAiDHandler(repo)
	handler.FindRAiDByName(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response models.RAiD
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if repo.GetRAiDCalls != 1 {
		t.Errorf("Expected 1 GetRAiD call, got %d", repo.GetRAiDCalls)
	}
}

func TestFindRAiDByName_NotFound(t *testing.T) {
	repo := testutil.NewMockRepository()

	repo.GetRAiDFunc = func(ctx context.Context, prefix, suffix string) (*models.RAiD, error) {
		return nil, storage.ErrNotFound
	}

	req := httptest.NewRequest(http.MethodGet, "/raid/10.12345/99999", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("prefix", "10.12345")
	rctx.URLParams.Add("suffix", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := NewRAiDHandler(repo)
	handler.FindRAiDByName(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestUpdateRAiD_Success(t *testing.T) {
	repo := testutil.NewMockRepository()
	prefix, suffix := "10.12345", "67890"
	testRAiD := testutil.NewTestRAiD(prefix, suffix)

	repo.UpdateRAiDFunc = func(ctx context.Context, p, s string, raid *models.RAiD) (*models.RAiD, error) {
		// Increment version
		if raid.Identifier != nil {
			raid.Identifier.Version++
		}
		return raid, nil
	}

	// Modify the test RAiD for update
	updatedRAiD := testRAiD
	updatedRAiD.Title[0].Text = "Updated Title"

	bodyBytes, _ := json.Marshal(updatedRAiD)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/raid/%s/%s", prefix, suffix), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("prefix", prefix)
	rctx.URLParams.Add("suffix", suffix)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := NewRAiDHandler(repo)
	handler.UpdateRAiD(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response models.RAiD
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if repo.UpdateRAiDCalls != 1 {
		t.Errorf("Expected 1 UpdateRAiD call, got %d", repo.UpdateRAiDCalls)
	}
}

func TestUpdateRAiD_NotFound(t *testing.T) {
	repo := testutil.NewMockRepository()

	repo.UpdateRAiDFunc = func(ctx context.Context, prefix, suffix string, raid *models.RAiD) (*models.RAiD, error) {
		return nil, storage.ErrNotFound
	}

	testRAiD := testutil.NewTestRAiD("10.12345", "99999")
	bodyBytes, _ := json.Marshal(testRAiD)

	req := httptest.NewRequest(http.MethodPut, "/raid/10.12345/99999", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("prefix", "10.12345")
	rctx.URLParams.Add("suffix", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := NewRAiDHandler(repo)
	handler.UpdateRAiD(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}

	// Should have called UpdateRAiD which returned error
	if repo.UpdateRAiDCalls != 1 {
		t.Errorf("Expected 1 UpdateRAiD call, got %d", repo.UpdateRAiDCalls)
	}
}

func TestRAiDHistory_Success(t *testing.T) {
	repo := testutil.NewMockRepository()
	prefix, suffix := "10.12345", "67890"

	// Create history versions
	history := []*models.RAiD{
		testutil.NewTestRAiD(prefix, suffix),
		testutil.NewTestRAiD(prefix, suffix),
		testutil.NewTestRAiD(prefix, suffix),
	}
	// Set different versions
	history[0].Identifier.Version = 1
	history[1].Identifier.Version = 2
	history[2].Identifier.Version = 3

	repo.GetRAiDHistoryFunc = func(ctx context.Context, p, s string) ([]*models.RAiD, error) {
		if p != prefix || s != suffix {
			t.Errorf("Expected prefix=%s suffix=%s, got prefix=%s suffix=%s", prefix, suffix, p, s)
		}
		return history, nil
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/raid/%s/%s/history", prefix, suffix), nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("prefix", prefix)
	rctx.URLParams.Add("suffix", suffix)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := NewRAiDHandler(repo)
	handler.RAiDHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response []*models.RAiD
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(response))
	}

	// Verify versions are in sequence
	for i, raid := range response {
		expectedVersion := i + 1
		if raid.Identifier.Version != expectedVersion {
			t.Errorf("Expected version %d, got %d", expectedVersion, raid.Identifier.Version)
		}
	}

	if repo.GetRAiDHistoryCalls != 1 {
		t.Errorf("Expected 1 GetRAiDHistory call, got %d", repo.GetRAiDHistoryCalls)
	}
}

func TestRAiDHistory_NotFound(t *testing.T) {
	repo := testutil.NewMockRepository()

	repo.GetRAiDHistoryFunc = func(ctx context.Context, prefix, suffix string) ([]*models.RAiD, error) {
		return nil, storage.ErrNotFound
	}

	req := httptest.NewRequest(http.MethodGet, "/raid/10.12345/99999/history", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("prefix", "10.12345")
	rctx.URLParams.Add("suffix", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := NewRAiDHandler(repo)
	handler.RAiDHistory(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}
