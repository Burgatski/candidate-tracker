package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/remotely-works/frontend-challenge/server/domain"
	"github.com/remotely-works/frontend-challenge/server/transport"
)

// Mock service

// mockSvc satisfies the unexported candidateUseCase interface via structural
// typing — no explicit interface declaration needed in this package.
type mockSvc struct {
	listFn    func(ctx context.Context, page int) (*domain.CandidateList, error)
	getByIDFn func(ctx context.Context, id int) (*domain.Candidate, error)
	createFn  func(ctx context.Context, c *domain.Candidate) error
	updateFn  func(ctx context.Context, id int, c *domain.Candidate) error
	deleteFn  func(ctx context.Context, id int) error
}

func (m *mockSvc) List(ctx context.Context, page int) (*domain.CandidateList, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page)
	}
	return &domain.CandidateList{Candidates: []*domain.Candidate{}}, nil
}
func (m *mockSvc) GetByID(ctx context.Context, id int) (*domain.Candidate, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}
func (m *mockSvc) Create(ctx context.Context, c *domain.Candidate) error {
	if m.createFn != nil {
		return m.createFn(ctx, c)
	}
	c.ID = 1
	return nil
}
func (m *mockSvc) Update(ctx context.Context, id int, c *domain.Candidate) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, c)
	}
	return nil
}
func (m *mockSvc) Delete(ctx context.Context, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

// Helpers

func newHandler(svc *mockSvc) http.Handler {
	return transport.NewHandler(slog.New(slog.NewTextHandler(nil, nil)), svc)
}

func do(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&m))
	return m
}

func jsonStatus(t *testing.T, rec *httptest.ResponseRecorder) int {
	t.Helper()
	body := decodeBody(t, rec)
	return int(body["status"].(float64))
}

var validPayload = map[string]any{
	"first_name": "Ada",
	"last_name":  "Lovelace",
	"email":      "ada@example.com",
	"phone":      "+1234567890",
	"picture":    "data:image/png;base64,abc",
	"skills":     []string{"Go"},
}

// GET /candidates

func TestList_ReturnsOK(t *testing.T) {
	svc := &mockSvc{
		listFn: func(_ context.Context, page int) (*domain.CandidateList, error) {
			assert.Equal(t, 2, page)
			return &domain.CandidateList{
				Total:      1,
				Candidates: []*domain.Candidate{{ID: 1, FirstName: "Ada"}},
				Pagination: domain.Pagination{Page: 2, PerPage: 25, TotalPages: 1},
			}, nil
		},
	}
	rec := do(t, newHandler(svc), http.MethodGet, "/candidates?page=2", nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 200, jsonStatus(t, rec))
}

func TestList_MissingPage_DefaultsTo1(t *testing.T) {
	svc := &mockSvc{
		listFn: func(_ context.Context, page int) (*domain.CandidateList, error) {
			assert.Equal(t, 0, page, "handler passes raw param; service normalises it")
			return &domain.CandidateList{Candidates: []*domain.Candidate{}}, nil
		},
	}
	rec := do(t, newHandler(svc), http.MethodGet, "/candidates", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// GET /candidates/{id}

func TestGet_ReturnsCandidate(t *testing.T) {
	svc := &mockSvc{
		getByIDFn: func(_ context.Context, id int) (*domain.Candidate, error) {
			assert.Equal(t, 42, id)
			return &domain.Candidate{ID: 42, FirstName: "Ada", Skills: []string{}}, nil
		},
	}
	rec := do(t, newHandler(svc), http.MethodGet, "/candidates/42", nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := decodeBody(t, rec)
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(42), data["id"])
}

func TestGet_NotFound_Returns404(t *testing.T) {
	// mockSvc.GetByID default returns ErrNotFound
	rec := do(t, newHandler(&mockSvc{}), http.MethodGet, "/candidates/999", nil)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, 404, jsonStatus(t, rec))
}

func TestGet_InvalidID_Returns400(t *testing.T) {
	rec := do(t, newHandler(&mockSvc{}), http.MethodGet, "/candidates/abc", nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 400, jsonStatus(t, rec))
}

// POST /candidates

func TestCreate_ValidPayload_Returns201(t *testing.T) {
	svc := &mockSvc{
		createFn: func(_ context.Context, c *domain.Candidate) error {
			c.ID = 7
			return nil
		},
	}
	rec := do(t, newHandler(svc), http.MethodPost, "/candidates", validPayload)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 201, jsonStatus(t, rec))
}

func TestCreate_MissingRequiredFields_Returns400(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]any
		wantErr string
	}{
		{"empty first_name", map[string]any{"first_name": "", "last_name": "X", "email": "a@b.com", "phone": "1", "picture": "x"}, "first_name"},
		{"invalid email", map[string]any{"first_name": "A", "last_name": "B", "email": "not-an-email", "phone": "1", "picture": "x"}, "email"},
		{"missing phone", map[string]any{"first_name": "A", "last_name": "B", "email": "a@b.com", "phone": "", "picture": "x"}, "phone"},
		{"missing picture", map[string]any{"first_name": "A", "last_name": "B", "email": "a@b.com", "phone": "1", "picture": ""}, "picture"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(t, newHandler(&mockSvc{}), http.MethodPost, "/candidates", tc.payload)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			body := decodeBody(t, rec)
			errs := fmt.Sprint(body["errors"])
			assert.Contains(t, errs, tc.wantErr)
		})
	}
}

func TestCreate_DuplicateEmail_Returns400(t *testing.T) {
	svc := &mockSvc{
		createFn: func(_ context.Context, _ *domain.Candidate) error {
			return domain.ErrDuplicateEmail
		},
	}
	rec := do(t, newHandler(svc), http.MethodPost, "/candidates", validPayload)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "email already exists")
}

func TestCreate_InvalidJSON_Returns400(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/candidates", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	newHandler(&mockSvc{}).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// PATCH /candidates/{id}

func TestUpdate_ValidPayload_Returns200(t *testing.T) {
	svc := &mockSvc{
		updateFn: func(_ context.Context, id int, c *domain.Candidate) error {
			assert.Equal(t, 3, id)
			c.ID = id
			return nil
		},
	}
	rec := do(t, newHandler(svc), http.MethodPatch, "/candidates/3", validPayload)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 200, jsonStatus(t, rec))
}

func TestUpdate_NotFound_Returns404(t *testing.T) {
	svc := &mockSvc{
		updateFn: func(_ context.Context, _ int, _ *domain.Candidate) error {
			return domain.ErrNotFound
		},
	}
	rec := do(t, newHandler(svc), http.MethodPatch, "/candidates/99", validPayload)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, 404, jsonStatus(t, rec))
}

func TestUpdate_DuplicateEmail_Returns400(t *testing.T) {
	svc := &mockSvc{
		updateFn: func(_ context.Context, _ int, _ *domain.Candidate) error {
			return domain.ErrDuplicateEmail
		},
	}
	rec := do(t, newHandler(svc), http.MethodPatch, "/candidates/1", validPayload)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdate_InvalidID_Returns400(t *testing.T) {
	rec := do(t, newHandler(&mockSvc{}), http.MethodPatch, "/candidates/xyz", validPayload)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// DELETE /candidates/{id}

func TestDelete_Returns200(t *testing.T) {
	svc := &mockSvc{
		deleteFn: func(_ context.Context, id int) error {
			assert.Equal(t, 5, id)
			return nil
		},
	}
	rec := do(t, newHandler(svc), http.MethodDelete, "/candidates/5", nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 200, jsonStatus(t, rec))
}

func TestDelete_NotFound_Returns404(t *testing.T) {
	svc := &mockSvc{
		deleteFn: func(_ context.Context, _ int) error { return domain.ErrNotFound },
	}
	rec := do(t, newHandler(svc), http.MethodDelete, "/candidates/1", nil)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDelete_InvalidID_Returns400(t *testing.T) {
	rec := do(t, newHandler(&mockSvc{}), http.MethodDelete, "/candidates/nope", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Response envelope

func TestResponseEnvelope_OKShape(t *testing.T) {
	svc := &mockSvc{
		getByIDFn: func(_ context.Context, _ int) (*domain.Candidate, error) {
			return &domain.Candidate{ID: 1, Skills: []string{}}, nil
		},
	}
	rec := do(t, newHandler(svc), http.MethodGet, "/candidates/1", nil)

	var envelope struct {
		Status int             `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&envelope))
	assert.Equal(t, 200, envelope.Status)
	assert.NotEmpty(t, envelope.Data)
}

func TestResponseEnvelope_ErrorShape(t *testing.T) {
	rec := do(t, newHandler(&mockSvc{}), http.MethodGet, "/candidates/999", nil)

	var envelope struct {
		Status int      `json:"status"`
		Errors []string `json:"errors"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&envelope))
	assert.Equal(t, 404, envelope.Status)
	assert.NotEmpty(t, envelope.Errors)
}
