package domain

import "errors"

var ErrNotFound = errors.New("not found")
var ErrDuplicateEmail = errors.New("email already exists")

type Candidate struct {
	ID        int      `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Picture   string   `json:"picture"`
	Skills    []string `json:"skills"`
}

type CandidateList struct {
	Total      int          `json:"total"`
	Candidates []*Candidate `json:"candidates"`
	Pagination Pagination   `json:"pagination"`
}

type Pagination struct {
	PerPage    int `json:"per_page"`
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
}
