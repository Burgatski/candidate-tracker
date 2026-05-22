package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/remotely-works/frontend-challenge/server/domain"
)

type SQLiteRepo struct {
	db *sql.DB
}

func New(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{db: db}
}

func (r *SQLiteRepo) List(ctx context.Context, offset, limit int) ([]*domain.Candidate, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM candidates`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list count: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, first_name, last_name, email, phone, picture FROM candidates ORDER BY id LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list query: %w", err)
	}
	defer rows.Close()

	var candidates []*domain.Candidate
	var ids []int
	for rows.Next() {
		c := &domain.Candidate{Skills: []string{}}
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Picture); err != nil {
			return nil, 0, err
		}
		candidates = append(candidates, c)
		ids = append(ids, c.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(ids) > 0 {
		skillMap, err := r.loadSkills(ctx, ids)
		if err != nil {
			return nil, 0, err
		}
		for _, c := range candidates {
			if s, ok := skillMap[c.ID]; ok {
				c.Skills = s
			}
		}
	}

	return candidates, total, nil
}

func (r *SQLiteRepo) FindByID(ctx context.Context, id int) (*domain.Candidate, error) {
	c := &domain.Candidate{Skills: []string{}}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, first_name, last_name, email, phone, picture FROM candidates WHERE id = ?`, id,
	).Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Picture)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("findByID: %w", err)
	}

	skillMap, err := r.loadSkills(ctx, []int{id})
	if err != nil {
		return nil, err
	}
	if s, ok := skillMap[id]; ok {
		c.Skills = s
	}
	return c, nil
}

func (r *SQLiteRepo) FindByEmail(ctx context.Context, email string) (*domain.Candidate, error) {
	c := &domain.Candidate{Skills: []string{}}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, first_name, last_name, email, phone, picture FROM candidates WHERE email = ?`, email,
	).Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Picture)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("findByEmail: %w", err)
	}
	return c, nil
}

func (r *SQLiteRepo) Create(ctx context.Context, c *domain.Candidate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("create begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	res, err := tx.ExecContext(ctx,
		`INSERT INTO candidates (first_name, last_name, email, phone, picture) VALUES (?, ?, ?, ?, ?)`,
		c.FirstName, c.LastName, c.Email, c.Phone, c.Picture,
	)
	if err != nil {
		return fmt.Errorf("create insert: %w", err)
	}
	id, _ := res.LastInsertId()
	c.ID = int(id)

	for _, skill := range c.Skills {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO candidate_skills (candidate_id, skill) VALUES (?, ?)`, c.ID, skill,
		); err != nil {
			return fmt.Errorf("create skill: %w", err)
		}
	}
	return tx.Commit()
}

func (r *SQLiteRepo) Update(ctx context.Context, c *domain.Candidate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("update begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx,
		`UPDATE candidates SET first_name=?, last_name=?, email=?, phone=?, picture=? WHERE id=?`,
		c.FirstName, c.LastName, c.Email, c.Phone, c.Picture, c.ID,
	); err != nil {
		return fmt.Errorf("update candidate: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM candidate_skills WHERE candidate_id=?`, c.ID); err != nil {
		return fmt.Errorf("update delete skills: %w", err)
	}

	for _, skill := range c.Skills {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO candidate_skills (candidate_id, skill) VALUES (?, ?)`, c.ID, skill,
		); err != nil {
			return fmt.Errorf("update skill: %w", err)
		}
	}
	return tx.Commit()
}

func (r *SQLiteRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM candidates WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SQLiteRepo) loadSkills(ctx context.Context, ids []int) (map[int][]string, error) {
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT candidate_id, skill FROM candidate_skills WHERE candidate_id IN (`+placeholders+`) ORDER BY candidate_id, skill`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("loadSkills: %w", err)
	}
	defer rows.Close()

	m := make(map[int][]string)
	for rows.Next() {
		var cid int
		var skill string
		if err := rows.Scan(&cid, &skill); err != nil {
			return nil, err
		}
		m[cid] = append(m[cid], skill)
	}
	return m, rows.Err()
}
