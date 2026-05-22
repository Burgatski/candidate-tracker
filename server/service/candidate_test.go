package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/remotely-works/frontend-challenge/server/domain"
	"github.com/remotely-works/frontend-challenge/server/service"
)

// Mock repository

// mockRepo implements service.CandidateRepository via optional function fields.
// Any nil field returns a sensible zero-value so tests only set what they need.
type mockRepo struct {
	listFn        func(ctx context.Context, offset, limit int) ([]*domain.Candidate, int, error)
	findByIDFn    func(ctx context.Context, id int) (*domain.Candidate, error)
	findByEmailFn func(ctx context.Context, email string) (*domain.Candidate, error)
	createFn      func(ctx context.Context, c *domain.Candidate) error
	updateFn      func(ctx context.Context, c *domain.Candidate) error
	deleteFn      func(ctx context.Context, id int) error
}

func (m *mockRepo) List(ctx context.Context, offset, limit int) ([]*domain.Candidate, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, offset, limit)
	}
	return []*domain.Candidate{}, 0, nil
}

func (m *mockRepo) FindByID(ctx context.Context, id int) (*domain.Candidate, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockRepo) FindByEmail(ctx context.Context, email string) (*domain.Candidate, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, domain.ErrNotFound
}

func (m *mockRepo) Create(ctx context.Context, c *domain.Candidate) error {
	if m.createFn != nil {
		return m.createFn(ctx, c)
	}
	c.ID = 1
	return nil
}

func (m *mockRepo) Update(ctx context.Context, c *domain.Candidate) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, c)
	}
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

// List

func TestList_PaginationMath(t *testing.T) {
	tests := []struct {
		name            string
		page, total     int
		wantPage        int
		wantTotalPages  int
		wantOffset      int
	}{
		{"page 1 of 4", 1, 80, 1, 4, 0},
		{"page 2 of 4", 2, 80, 2, 4, 25},
		{"page 4 of 4", 4, 80, 4, 4, 75},
		{"page 0 coerced to 1", 0, 80, 1, 4, 0},
		{"negative page coerced to 1", -5, 80, 1, 4, 0},
		{"partial last page rounds up", 1, 26, 1, 2, 0},
		{"empty DB gives 1 total page", 1, 0, 1, 1, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRepo{
				listFn: func(_ context.Context, offset, limit int) ([]*domain.Candidate, int, error) {
					assert.Equal(t, tc.wantOffset, offset, "wrong offset passed to repo")
					assert.Equal(t, 25, limit, "per-page must always be 25")
					return []*domain.Candidate{}, tc.total, nil
				},
			}
			svc := service.New(repo)
			result, err := svc.List(context.Background(), tc.page)
			require.NoError(t, err)
			assert.Equal(t, tc.wantPage, result.Pagination.Page)
			assert.Equal(t, tc.wantTotalPages, result.Pagination.TotalPages)
			assert.Equal(t, 25, result.Pagination.PerPage)
			assert.Equal(t, tc.total, result.Total)
		})
	}
}

func TestList_PropagatesRepoError(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, _, _ int) ([]*domain.Candidate, int, error) {
			return nil, 0, assert.AnError
		},
	}
	_, err := service.New(repo).List(context.Background(), 1)
	assert.Error(t, err)
}

// GetByID

func TestGetByID_DelegatesToRepo(t *testing.T) {
	want := &domain.Candidate{ID: 7, FirstName: "Ada"}
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, id int) (*domain.Candidate, error) {
			assert.Equal(t, 7, id)
			return want, nil
		},
	}
	got, err := service.New(repo).GetByID(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetByID_NotFound(t *testing.T) {
	// mockRepo.FindByID default returns ErrNotFound
	_, err := service.New(&mockRepo{}).GetByID(context.Background(), 99)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// Create

func TestCreate_RejectsDuplicateEmail(t *testing.T) {
	existing := &domain.Candidate{ID: 1, Email: "taken@example.com"}
	repo := &mockRepo{
		findByEmailFn: func(_ context.Context, email string) (*domain.Candidate, error) {
			return existing, nil // email already exists
		},
	}
	err := service.New(repo).Create(context.Background(), &domain.Candidate{Email: "taken@example.com"})
	assert.ErrorIs(t, err, domain.ErrDuplicateEmail)
}

func TestCreate_CallsRepoWhenEmailFree(t *testing.T) {
	created := false
	repo := &mockRepo{
		createFn: func(_ context.Context, c *domain.Candidate) error {
			created = true
			c.ID = 42
			return nil
		},
	}
	c := &domain.Candidate{Email: "new@example.com"}
	require.NoError(t, service.New(repo).Create(context.Background(), c))
	assert.True(t, created)
	assert.Equal(t, 42, c.ID, "ID set by repo must be visible to caller")
}

// Update

func TestUpdate_SetsIDOnCandidate(t *testing.T) {
	var updatedC *domain.Candidate
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, id int) (*domain.Candidate, error) {
			return &domain.Candidate{ID: id}, nil
		},
		updateFn: func(_ context.Context, c *domain.Candidate) error {
			updatedC = c
			return nil
		},
	}
	c := &domain.Candidate{Email: "same@example.com"}
	require.NoError(t, service.New(repo).Update(context.Background(), 5, c))
	require.NotNil(t, updatedC)
	assert.Equal(t, 5, updatedC.ID, "service must inject the route ID into the candidate")
}

func TestUpdate_RejectsEmailTakenByOtherCandidate(t *testing.T) {
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, id int) (*domain.Candidate, error) {
			return &domain.Candidate{ID: id}, nil
		},
		findByEmailFn: func(_ context.Context, _ string) (*domain.Candidate, error) {
			return &domain.Candidate{ID: 99}, nil
		},
	}
	err := service.New(repo).Update(context.Background(), 5, &domain.Candidate{Email: "taken@example.com"})
	assert.ErrorIs(t, err, domain.ErrDuplicateEmail)
}

func TestUpdate_AllowsSameEmailForSameCandidate(t *testing.T) {
	// Candidate 5 keeps their own email — must NOT return ErrDuplicateEmail
	repo := &mockRepo{
		findByIDFn: func(_ context.Context, id int) (*domain.Candidate, error) {
			return &domain.Candidate{ID: id}, nil
		},
		findByEmailFn: func(_ context.Context, _ string) (*domain.Candidate, error) {
			return &domain.Candidate{ID: 5}, nil // same candidate, OK
		},
	}
	err := service.New(repo).Update(context.Background(), 5, &domain.Candidate{Email: "mine@example.com"})
	assert.NoError(t, err)
}

func TestUpdate_NotFound(t *testing.T) {
	// mockRepo.FindByID default returns ErrNotFound
	err := service.New(&mockRepo{}).Update(context.Background(), 99, &domain.Candidate{})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// Delete

func TestDelete_DelegatesToRepo(t *testing.T) {
	called := false
	repo := &mockRepo{
		deleteFn: func(_ context.Context, id int) error {
			assert.Equal(t, 3, id)
			called = true
			return nil
		},
	}
	require.NoError(t, service.New(repo).Delete(context.Background(), 3))
	assert.True(t, called)
}

func TestDelete_PropagatesNotFound(t *testing.T) {
	repo := &mockRepo{
		deleteFn: func(_ context.Context, _ int) error { return domain.ErrNotFound },
	}
	assert.ErrorIs(t, service.New(repo).Delete(context.Background(), 1), domain.ErrNotFound)
}
