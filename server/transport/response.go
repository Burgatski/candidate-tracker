package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type okResponse struct {
	Status int `json:"status"`
	Data   any `json:"data"`
}

type errResponse struct {
	Status int      `json:"status"`
	Errors []string `json:"errors"`
}

func writeOK(w http.ResponseWriter, log *slog.Logger, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	if err := json.NewEncoder(w).Encode(okResponse{Status: status, Data: data}); err != nil {
		log.Error("writeOK encode", slog.String("err", err.Error()))
	}
}

func writeError(w http.ResponseWriter, log *slog.Logger, status int, errs ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errResponse{Status: status, Errors: errs}); err != nil {
		log.Error("writeError encode", slog.String("err", err.Error()))
	}
}
