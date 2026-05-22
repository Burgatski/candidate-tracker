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

type responder struct {
	w   http.ResponseWriter
	log *slog.Logger
}

func respond(w http.ResponseWriter, log *slog.Logger) responder {
	return responder{w: w, log: log}
}

func (r responder) ok(status int, data any) {
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(status)
	if err := json.NewEncoder(r.w).Encode(okResponse{Status: status, Data: data}); err != nil {
		r.log.Error("encode ok response", slog.String("err", err.Error()))
	}
}

func (r responder) err(status int, errs ...string) {
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(status)
	if err := json.NewEncoder(r.w).Encode(errResponse{Status: status, Errors: errs}); err != nil {
		r.log.Error("encode error response", slog.String("err", err.Error()))
	}
}
