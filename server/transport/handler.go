package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/remotely-works/frontend-challenge/server/domain"
)

type candidateUseCase interface {
	List(ctx context.Context, page int) (*domain.CandidateList, error)
	GetByID(ctx context.Context, id int) (*domain.Candidate, error)
	Create(ctx context.Context, c *domain.Candidate) error
	Update(ctx context.Context, id int, c *domain.Candidate) error
	Delete(ctx context.Context, id int) error
}

type Handler struct {
	log *slog.Logger
	svc candidateUseCase
	mux *http.ServeMux
}

func NewHandler(log *slog.Logger, svc candidateUseCase) *Handler {
	h := &Handler{log: log, svc: svc, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /candidates", h.list)
	h.mux.HandleFunc("POST /candidates", h.create)
	h.mux.HandleFunc("GET /candidates/{id}", h.get)
	h.mux.HandleFunc("PATCH /candidates/{id}", h.update)
	h.mux.HandleFunc("DELETE /candidates/{id}", h.delete)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// Request DTO

type candidateRequest struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Picture   string   `json:"picture"`
	Skills    []string `json:"skills"`
}

var emailRe = regexp.MustCompile(`^.+@.+\..+$`)

func (r candidateRequest) validate() []string {
	var errs []string
	if strings.TrimSpace(r.FirstName) == "" {
		errs = append(errs, "first_name is empty")
	}
	if strings.TrimSpace(r.LastName) == "" {
		errs = append(errs, "last_name is empty")
	}
	if strings.TrimSpace(r.Phone) == "" {
		errs = append(errs, "phone is empty")
	}
	if strings.TrimSpace(r.Picture) == "" {
		errs = append(errs, "picture is empty")
	}
	if !emailRe.MatchString(r.Email) {
		errs = append(errs, "email is invalid")
	}
	for i, s := range r.Skills {
		if strings.TrimSpace(s) == "" {
			errs = append(errs, fmt.Sprintf("skills[%d] is empty", i))
		}
	}
	return errs
}

func (r candidateRequest) toCandidate() *domain.Candidate {
	skills := r.Skills
	if skills == nil {
		skills = []string{}
	}
	return &domain.Candidate{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Email:     r.Email,
		Phone:     r.Phone,
		Picture:   r.Picture,
		Skills:    skills,
	}
}

// Handlers

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	resp := respond(w, h.log)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	result, err := h.svc.List(r.Context(), page)
	if err != nil {
		h.log.Error("list", slog.String("err", err.Error()))
		resp.err(http.StatusInternalServerError, "internal error")
		return
	}
	resp.ok(http.StatusOK, result)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	resp := respond(w, h.log)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		resp.err(http.StatusBadRequest, "invalid id")
		return
	}
	candidate, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		resp.err(http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		h.log.Error("get", slog.String("err", err.Error()))
		resp.err(http.StatusInternalServerError, "internal error")
		return
	}
	resp.ok(http.StatusOK, candidate)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	resp := respond(w, h.log)
	var req candidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.err(http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := req.validate(); len(errs) > 0 {
		resp.err(http.StatusBadRequest, errs...)
		return
	}
	c := req.toCandidate()
	if err := h.svc.Create(r.Context(), c); err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			resp.err(http.StatusBadRequest, "email already exists")
			return
		}
		h.log.Error("create", slog.String("err", err.Error()))
		resp.err(http.StatusInternalServerError, "internal error")
		return
	}
	resp.ok(http.StatusCreated, c)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	resp := respond(w, h.log)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		resp.err(http.StatusBadRequest, "invalid id")
		return
	}
	var req candidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.err(http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := req.validate(); len(errs) > 0 {
		resp.err(http.StatusBadRequest, errs...)
		return
	}
	c := req.toCandidate()
	if err := h.svc.Update(r.Context(), id, c); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			resp.err(http.StatusNotFound, "not found")
		case errors.Is(err, domain.ErrDuplicateEmail):
			resp.err(http.StatusBadRequest, "email already exists")
		default:
			h.log.Error("update", slog.String("err", err.Error()))
			resp.err(http.StatusInternalServerError, "internal error")
		}
		return
	}
	resp.ok(http.StatusOK, c)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	resp := respond(w, h.log)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		resp.err(http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			resp.err(http.StatusNotFound, "not found")
			return
		}
		h.log.Error("delete", slog.String("err", err.Error()))
		resp.err(http.StatusInternalServerError, "internal error")
		return
	}
	resp.ok(http.StatusOK, nil)
}
