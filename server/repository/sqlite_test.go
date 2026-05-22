package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/remotely-works/frontend-challenge/server/domain"
	"github.com/remotely-works/frontend-challenge/server/repository"
)

// setupRepo opens an in-memory SQLite database, runs migrations and returns a
// ready-to-use repository. The database is closed automatically when the test
// ends.
func setupRepo(t *testing.T) (*repository.SQLiteRepo, *sql.DB) {
	t.Helper()
	db, err := repository.OpenDB(":memory:")
	require.NoError(t, err, "open in-memory db")
	t.Cleanup(func() { db.Close() })
	require.NoError(t, repository.Migrate(context.Background(), db), "migrate")
	return repository.New(db), db
}

func candidate(email string) *domain.Candidate {
	return &domain.Candidate{
		FirstName: "John",
		LastName:  "Doe",
		Email:     email,
		Phone:     "+1234567890",
		Picture:   "data:image/png;base64,abc123",
		Skills:    []string{"Go", "SQL"},
	}
}

// Create

func TestCreate_SetsID(t *testing.T) {
	repo, _ := setupRepo(t)
	c := candidate("john@example.com")

	require.NoError(t, repo.Create(context.Background(), c))
	assert.NotZero(t, c.ID, "ID should be populated after Create")
}

func TestCreate_DuplicateEmail_ReturnsError(t *testing.T) {
	repo, _ := setupRepo(t)
	require.NoError(t, repo.Create(context.Background(), candidate("dupe@example.com")))

	err := repo.Create(context.Background(), candidate("dupe@example.com"))
	assert.Error(t, err, "second insert with same email must fail")
}

func TestCreate_StoresSkills(t *testing.T) {
	repo, _ := setupRepo(t)
	c := candidate("skills@example.com")
	c.Skills = []string{"Go", "React", "PostgreSQL"}
	require.NoError(t, repo.Create(context.Background(), c))

	found, err := repo.FindByID(context.Background(), c.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Go", "React", "PostgreSQL"}, found.Skills)
}

func TestCreate_NilSkills_ReturnsEmptySlice(t *testing.T) {
	repo, _ := setupRepo(t)
	c := candidate("noskills@example.com")
	c.Skills = nil
	require.NoError(t, repo.Create(context.Background(), c))

	found, err := repo.FindByID(context.Background(), c.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.Skills, "Skills must never be nil in JSON responses")
	assert.Empty(t, found.Skills)
}

// FindByID

func TestFindByID_ReturnsCandidate(t *testing.T) {
	repo, _ := setupRepo(t)
	c := candidate("find@example.com")
	require.NoError(t, repo.Create(context.Background(), c))

	found, err := repo.FindByID(context.Background(), c.ID)
	require.NoError(t, err)
	assert.Equal(t, c.ID, found.ID)
	assert.Equal(t, c.Email, found.Email)
	assert.Equal(t, c.FirstName, found.FirstName)
}

func TestFindByID_NotFound(t *testing.T) {
	repo, _ := setupRepo(t)

	_, err := repo.FindByID(context.Background(), 99999)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// FindByEmail

func TestFindByEmail_ReturnsCandidate(t *testing.T) {
	repo, _ := setupRepo(t)
	c := candidate("email@example.com")
	require.NoError(t, repo.Create(context.Background(), c))

	found, err := repo.FindByEmail(context.Background(), "email@example.com")
	require.NoError(t, err)
	assert.Equal(t, c.ID, found.ID)
}

func TestFindByEmail_NotFound(t *testing.T) {
	repo, _ := setupRepo(t)

	_, err := repo.FindByEmail(context.Background(), "ghost@example.com")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// List

func TestList_ReturnsAll(t *testing.T) {
	repo, _ := setupRepo(t)
	emails := []string{"a@test.com", "b@test.com", "c@test.com"}
	for _, e := range emails {
		require.NoError(t, repo.Create(context.Background(), candidate(e)))
	}

	results, total, err := repo.List(context.Background(), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, results, 3)
}

func TestList_Pagination(t *testing.T) {
	repo, _ := setupRepo(t)
	for _, e := range []string{"p1@test.com", "p2@test.com", "p3@test.com"} {
		require.NoError(t, repo.Create(context.Background(), candidate(e)))
	}

	tests := []struct {
		offset, limit int
		wantLen       int
		wantTotal     int
	}{
		{offset: 0, limit: 2, wantLen: 2, wantTotal: 3},
		{offset: 2, limit: 2, wantLen: 1, wantTotal: 3},
		{offset: 3, limit: 2, wantLen: 0, wantTotal: 3},
	}
	for _, tc := range tests {
		got, total, err := repo.List(context.Background(), tc.offset, tc.limit)
		require.NoError(t, err)
		assert.Equal(t, tc.wantTotal, total, "offset=%d limit=%d", tc.offset, tc.limit)
		assert.Len(t, got, tc.wantLen, "offset=%d limit=%d", tc.offset, tc.limit)
	}
}

func TestList_SkillsLoadedForAllCandidates(t *testing.T) {
	repo, _ := setupRepo(t)
	c1 := candidate("s1@test.com")
	c1.Skills = []string{"Go"}
	c2 := candidate("s2@test.com")
	c2.Skills = []string{"Python", "React"}
	require.NoError(t, repo.Create(context.Background(), c1))
	require.NoError(t, repo.Create(context.Background(), c2))

	results, _, err := repo.List(context.Background(), 0, 10)
	require.NoError(t, err)

	skillsByID := make(map[int][]string)
	for _, c := range results {
		skillsByID[c.ID] = c.Skills
	}
	assert.ElementsMatch(t, []string{"Go"}, skillsByID[c1.ID])
	assert.ElementsMatch(t, []string{"Python", "React"}, skillsByID[c2.ID])
}

func TestList_EmptyDB_ReturnsZeroTotal(t *testing.T) {
	repo, _ := setupRepo(t)

	results, total, err := repo.List(context.Background(), 0, 25)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, results)
}

// Update

func TestUpdate_ChangesFields(t *testing.T) {
	repo, _ := setupRepo(t)
	c := candidate("upd@test.com")
	require.NoError(t, repo.Create(context.Background(), c))

	c.FirstName = "Jane"
	c.Email = "new@test.com"
	require.NoError(t, repo.Update(context.Background(), c))

	found, err := repo.FindByID(context.Background(), c.ID)
	require.NoError(t, err)
	assert.Equal(t, "Jane", found.FirstName)
	assert.Equal(t, "new@test.com", found.Email)
}

func TestUpdate_ReplacesSkillsCompletely(t *testing.T) {
	repo, _ := setupRepo(t)
	c := candidate("updskills@test.com")
	c.Skills = []string{"Go", "SQL"}
	require.NoError(t, repo.Create(context.Background(), c))

	c.Skills = []string{"Python"}
	require.NoError(t, repo.Update(context.Background(), c))

	found, err := repo.FindByID(context.Background(), c.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"Python"}, found.Skills, "old skills must be replaced, not merged")
}

// Delete

func TestDelete_RemovesCandidate(t *testing.T) {
	repo, _ := setupRepo(t)
	c := candidate("del@test.com")
	require.NoError(t, repo.Create(context.Background(), c))

	require.NoError(t, repo.Delete(context.Background(), c.ID))

	_, err := repo.FindByID(context.Background(), c.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDelete_CascadesSkills(t *testing.T) {
	repo, db := setupRepo(t)
	c := candidate("cascade@test.com")
	c.Skills = []string{"Go", "Python"}
	require.NoError(t, repo.Create(context.Background(), c))

	require.NoError(t, repo.Delete(context.Background(), c.ID))

	var count int
	err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM candidate_skills WHERE candidate_id = ?`, c.ID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count, "skills must be deleted with the candidate")
}

func TestDelete_NotFound(t *testing.T) {
	repo, _ := setupRepo(t)

	err := repo.Delete(context.Background(), 99999)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// Seed

func TestSeed_PopulatesOnFirstRun(t *testing.T) {
	_, db := setupRepo(t)
	data := []byte(`[
		{"id":1,"first_name":"A","last_name":"B","email":"a@b.com","phone":"1","picture":"x","skills":["Go"]},
		{"id":2,"first_name":"C","last_name":"D","email":"c@d.com","phone":"2","picture":"y","skills":[]}
	]`)
	require.NoError(t, repository.Seed(context.Background(), db, data))

	repo := repository.New(db)
	_, total, err := repo.List(context.Background(), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
}

func TestSeed_IsIdempotent(t *testing.T) {
	_, db := setupRepo(t)
	data := []byte(`[{"id":1,"first_name":"A","last_name":"B","email":"a@b.com","phone":"1","picture":"x","skills":[]}]`)

	require.NoError(t, repository.Seed(context.Background(), db, data))
	require.NoError(t, repository.Seed(context.Background(), db, data), "second seed must not error")

	repo := repository.New(db)
	_, total, err := repo.List(context.Background(), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total, "second seed must not duplicate rows")
}

