package service

import (
	"context"
	"math"

	"github.com/remotely-works/frontend-challenge/server/domain"
)

const perPage = 25

type CandidateRepository interface {
	List(ctx context.Context, offset, limit int) ([]*domain.Candidate, int, error)
	FindByID(ctx context.Context, id int) (*domain.Candidate, error)
	FindByEmail(ctx context.Context, email string) (*domain.Candidate, error)
	Create(ctx context.Context, c *domain.Candidate) error
	Update(ctx context.Context, c *domain.Candidate) error
	Delete(ctx context.Context, id int) error
}

type CandidateService struct {
	repo CandidateRepository
}

func New(repo CandidateRepository) *CandidateService {
	return &CandidateService{repo: repo}
}

func (s *CandidateService) List(ctx context.Context, page int) (*domain.CandidateList, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage
	candidates, total, err := s.repo.List(ctx, offset, perPage)
	if err != nil {
		return nil, err
	}
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}
	return &domain.CandidateList{
		Total:      total,
		Candidates: candidates,
		Pagination: domain.Pagination{
			PerPage:    perPage,
			Page:       page,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *CandidateService) GetByID(ctx context.Context, id int) (*domain.Candidate, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CandidateService) Create(ctx context.Context, c *domain.Candidate) error {
	if _, err := s.repo.FindByEmail(ctx, c.Email); err == nil {
		return domain.ErrDuplicateEmail
	}
	return s.repo.Create(ctx, c)
}

func (s *CandidateService) Update(ctx context.Context, id int, c *domain.Candidate) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	if found, err := s.repo.FindByEmail(ctx, c.Email); err == nil && found.ID != id {
		return domain.ErrDuplicateEmail
	}
	c.ID = id
	return s.repo.Update(ctx, c)
}

func (s *CandidateService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
