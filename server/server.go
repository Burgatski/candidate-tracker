package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Server struct {
	log *slog.Logger
	db  *Database
	mux *http.ServeMux
}

func NewServer(log *slog.Logger, db *Database) *Server {
	server := &Server{
		log: log,
		db:  db,
		mux: http.NewServeMux(),
	}

	server.mux.HandleFunc("GET /candidates", server.List)
	server.mux.HandleFunc("POST /candidates", server.Create)
	server.mux.HandleFunc("PATCH /candidates/{id}", server.Update)
	server.mux.HandleFunc("GET /candidates/{id}", server.Get)

	return server
}

type ListResponse struct {
	Total      int          `json:"total"`
	Candidates []*Candidate `json:"candidates"`
	Pagination struct {
		PerPage    int `json:"per_page"`
		Page       int `json:"page"`
		TotalPages int `json:"total_pages"`
	}
}

const perPage = 25

func (s *Server) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	var result = []*Candidate{}
	candidates := s.db.All()
	from, to := (page-1)*perPage, page*perPage
	if from < len(candidates)-1 {
		if to <= len(candidates) {
			result = candidates[from:to]
		} else {
			result = candidates[from:]
		}
	}

	var resp = ListResponse{
		Total:      len(candidates),
		Candidates: result,
	}
	resp.Pagination.Page = page
	resp.Pagination.PerPage = perPage
	resp.Pagination.TotalPages = int(math.Ceil(float64(len(candidates)) / float64(perPage)))

	s.ok(w, http.StatusOK, resp)
}

type CreateRequest struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Picture   string   `json:"picture"`
	Skills    []string `json:"skills"`
}

var emailRegex = regexp.MustCompile(`^.+@.+\..+$`)

func (r CreateRequest) Validate() []error {
	var errors []error

	if len(strings.TrimSpace(r.FirstName)) == 0 {
		errors = append(errors, fmt.Errorf("first_name is empty"))
	}

	if len(strings.TrimSpace(r.LastName)) == 0 {
		errors = append(errors, fmt.Errorf("last_name is empty"))
	}

	if len(strings.TrimSpace(r.Phone)) == 0 {
		errors = append(errors, fmt.Errorf("phone is empty"))
	}

	if len(strings.TrimSpace(r.Picture)) == 0 {
		errors = append(errors, fmt.Errorf("picture is empty"))
	}

	if !emailRegex.MatchString(r.Email) {
		errors = append(errors, fmt.Errorf("email is invalid"))
	}

	for i, s := range r.Skills {
		if len(strings.TrimSpace(s)) == 0 {
			errors = append(errors, fmt.Errorf("skills[%d] is empty", i))
		}
	}

	return errors
}

func (r CreateRequest) ToCandidate() *Candidate {
	return &Candidate{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Email:     r.Email,
		Phone:     r.Phone,
		Picture:   r.Picture,
		Skills:    r.Skills,
	}
}

func (s *Server) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.error(w, http.StatusBadRequest, err)
		return
	}

	if errors := req.Validate(); len(errors) > 0 {
		s.error(w, http.StatusBadRequest, errors...)
		return
	}

	if _, err := s.db.FindByEmail(req.Email); err == nil {
		s.error(w, http.StatusBadRequest, fmt.Errorf("email already exists"))
		return
	}

	var candidate = req.ToCandidate()
	s.db.Create(candidate)

	s.ok(w, http.StatusCreated, candidate)
}

type UpdateRequest = CreateRequest

func (s *Server) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if _, err := s.db.FindByID(id); err != nil {
		s.error(w, http.StatusNotFound, err)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.error(w, http.StatusBadRequest, err)
		return
	}

	if errors := req.Validate(); len(errors) > 0 {
		s.error(w, http.StatusBadRequest, errors...)
		return
	}

	found, err := s.db.FindByEmail(req.Email)
	if err == nil && found.ID != id {
		s.error(w, http.StatusBadRequest, fmt.Errorf("email already exists"))
		return
	}

	var updated = req.ToCandidate()
	updated.ID = id
	if err := s.db.Update(updated); err != nil {
		s.error(w, http.StatusInternalServerError, err)
		return
	}

	s.ok(w, http.StatusOK, updated)
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	candidate, err := s.db.FindByID(id)
	if err != nil {
		s.error(w, http.StatusNotFound, err)
		return
	}

	s.ok(w, http.StatusOK, candidate)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type OkResponse struct {
	Status int `json:"status"`
	Data   any `json:"data"`
}

func (s *Server) ok(w http.ResponseWriter, status int, data any) {
	resp, err := json.Marshal(OkResponse{
		Status: status,
		Data:   data,
	})
	if err != nil {
		s.log.Error("could not marshal ok response", slog.String("err", err.Error()))
		return
	}

	if _, err := w.Write(resp); err != nil {
		s.log.Error("could not write ok response", slog.String("err", err.Error()))
	}
}

type ErrorResponse struct {
	Status int      `json:"status"`
	Errors []string `json:"errors"`
}

func (s *Server) error(w http.ResponseWriter, status int, errors ...error) {
	resp, err := json.Marshal(ErrorResponse{
		Status: status,
		Errors: errorsToStrings(errors),
	})
	if err != nil {
		s.log.Error("could not marshal error response", slog.String("err", err.Error()))
		return
	}

	if _, err := w.Write(resp); err != nil {
		s.log.Error("could not write error response", slog.String("err", err.Error()))
	}
}

func errorsToStrings(errs []error) []string {
	var result []string
	for _, e := range errs {
		result = append(result, e.Error())
	}
	return result
}
