package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

//go:generate go run ./tools/generate-data

//go:embed data.json
var rawData []byte

type Candidate struct {
	ID        int      `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Picture   string   `json:"picture"`
	Skills    []string `json:"skills"`
}

type Database struct {
	mut     sync.RWMutex
	records []*Candidate
}

func (db *Database) All() []*Candidate {
	db.mut.RLock()
	defer db.mut.RUnlock()
	return db.records
}

func (db *Database) Create(c *Candidate) {
	db.mut.Lock()
	defer db.mut.Unlock()
	c.ID = len(db.records) + 1
	db.records = append(db.records, c)
}

var ErrNotFound = errors.New("not found")

func (db *Database) Update(updated *Candidate) error {
	db.mut.Lock()
	defer db.mut.Unlock()
	for i, c := range db.records {
		if c.ID == updated.ID {
			db.records[i] = updated
			return nil
		}
	}
	return ErrNotFound
}

func (db *Database) FindByID(id int) (*Candidate, error) {
	db.mut.RLock()
	defer db.mut.RUnlock()
	for _, c := range db.records {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (db *Database) FindByEmail(email string) (*Candidate, error) {
	db.mut.RLock()
	defer db.mut.RUnlock()
	for _, c := range db.records {
		if c.Email == email {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func ReadDatabase() (*Database, error) {
	var candidates []*Candidate
	if err := json.Unmarshal(rawData, &candidates); err != nil {
		return nil, fmt.Errorf("could not read database: %w", err)
	}
	return &Database{records: candidates}, nil
}
