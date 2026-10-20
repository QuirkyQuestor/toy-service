package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/QuirkyQuestor/toy-service/internal/store"
)

type Server struct {
	ShutdownStarted atomic.Bool
	db              *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) Routes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", s.PostNewItemHandler)
	mux.HandleFunc("GET /items/{id}", s.GetItemByIdHandler)
	mux.HandleFunc("GET /items", s.GetItemsHandler)
	mux.HandleFunc("DELETE /items/{id}", s.DeleteItemByIdHandler)
	mux.HandleFunc("GET /healthz", s.GetHealthzHandler) // the process is up
	mux.HandleFunc("GET /readyz", s.GetReadyzHandler)   // pings the DB
	return s.JsonMiddleware(mux)
}

func (s *Server) GetItemsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Server) GetItemByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	item, err := store.GetItemById(r.Context(), s.db, id)
	if err != nil {
		if errors.Is(err, store.ErrNoItemResponse) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(item)
	if err != nil {
		slog.Error("Could not marshal data object for response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func (s *Server) PostNewItemHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Server) DeleteItemByIdHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// checks only that the process is alive
func (s *Server) GetHealthzHandler(w http.ResponseWriter, r *http.Request) {
	if s.ShutdownStarted.Load() {
		slog.Error("shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "shutting down",
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// Does db.PingContext()
func (s *Server) GetReadyzHandler(w http.ResponseWriter, r *http.Request) {
	if s.ShutdownStarted.Load() {
		slog.Error("shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "shutting down",
		})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		slog.Warn("readiness: db ping failed", "err", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "db unavailable",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (s *Server) JsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Enforce JSON on incoming requests (for methods with a body)
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := r.Header.Get("Content-Type")
			if ct != "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnsupportedMediaType)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Content-Type must be application/json",
				})
				return
			}
		}

		// Force JSON on all responses
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
