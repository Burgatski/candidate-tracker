package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/remotely-works/frontend-challenge/server/domain"
	"github.com/remotely-works/frontend-challenge/server/service"
)

type Handler struct {
	log *slog.Logger
	svc *service.CandidateService
	mux *http.ServeMux
}

func NewHandler(log *slog.Logger, svc *service.CandidateService) *Handler {
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

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	result, err := h.svc.List(r.Context(), page)
	if err != nil {
		h.log.Error("list", slog.String("err", err.Error()))
		writeError(w, h.log, http.StatusInternalServerError, "internal error")
		return
	}
	writeOK(w, h.log, http.StatusOK, result)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, h.log, http.StatusBadRequest, "invalid id")
		return
	}
	candidate, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, h.log, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		h.log.Error("get", slog.String("err", err.Error()))
		writeError(w, h.log, http.StatusInternalServerError, "internal error")
		return
	}
	writeOK(w, h.log, http.StatusOK, candidate)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req candidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, h.log, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := req.validate(); len(errs) > 0 {
		writeError(w, h.log, http.StatusBadRequest, errs...)
		return
	}
	c := req.toCandidate()
	if err := h.svc.Create(r.Context(), c); err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			writeError(w, h.log, http.StatusBadRequest, "email already exists")
			return
		}
		h.log.Error("create", slog.String("err", err.Error()))
		writeError(w, h.log, http.StatusInternalServerError, "internal error")
		return
	}
	writeOK(w, h.log, http.StatusCreated, c)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, h.log, http.StatusBadRequest, "invalid id")
		return
	}
	var req candidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, h.log, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := req.validate(); len(errs) > 0 {
		writeError(w, h.log, http.StatusBadRequest, errs...)
		return
	}
	c := req.toCandidate()
	if err := h.svc.Update(r.Context(), id, c); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, h.log, http.StatusNotFound, "not found")
		case errors.Is(err, domain.ErrDuplicateEmail):
			writeError(w, h.log, http.StatusBadRequest, "email already exists")
		default:
			h.log.Error("update", slog.String("err", err.Error()))
			writeError(w, h.log, http.StatusInternalServerError, "internal error")
		}
		return
	}
	writeOK(w, h.log, http.StatusOK, c)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, h.log, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, h.log, http.StatusNotFound, "not found")
			return
		}
		h.log.Error("delete", slog.String("err", err.Error()))
		writeError(w, h.log, http.StatusInternalServerError, "internal error")
		return
	}
	writeOK(w, h.log, http.StatusOK, nil)
}
