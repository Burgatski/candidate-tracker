package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/remotely-works/frontend-challenge/server/domain"
)

const schema = `
CREATE TABLE IF NOT EXISTS candidates (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT    NOT NULL,
    last_name  TEXT    NOT NULL,
    email      TEXT    NOT NULL UNIQUE,
    phone      TEXT    NOT NULL,
    picture    TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS candidate_skills (
    candidate_id INTEGER NOT NULL,
    skill        TEXT    NOT NULL,
    UNIQUE(candidate_id, skill),
    FOREIGN KEY (candidate_id) REFERENCES candidates(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_skills_candidate ON candidate_skills(candidate_id);
`

func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

func Seed(ctx context.Context, db *sql.DB, rawData []byte) error {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM candidates`).Scan(&count); err != nil {
		return fmt.Errorf("seed count: %w", err)
	}
	if count > 0 {
		return nil
	}

	var candidates []domain.Candidate
	if err := json.Unmarshal(rawData, &candidates); err != nil {
		return fmt.Errorf("seed unmarshal: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("seed begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	for _, c := range candidates {
		res, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO candidates (first_name, last_name, email, phone, picture) VALUES (?, ?, ?, ?, ?)`,
			c.FirstName, c.LastName, c.Email, c.Phone, c.Picture,
		)
		if err != nil {
			continue
		}
		id, _ := res.LastInsertId()
		for _, skill := range c.Skills {
			if skill == "" {
				continue
			}
			_, _ = tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO candidate_skills (candidate_id, skill) VALUES (?, ?)`,
				id, skill,
			)
		}
	}

	return tx.Commit()
}
