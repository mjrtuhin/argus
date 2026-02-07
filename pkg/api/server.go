package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mjrtuhin/argus/pkg/storage"
)

type Server struct {
	router *mux.Router
	db     *storage.DB
	hub    *Hub
	port   string
}
func NewServer(db *storage.DB, port string) *Server {
	s := &Server{
		router: mux.NewRouter(),
		db:     db,
		hub:    NewHub(),
		port:   port,
	}

	s.setupRoutes()
	return s
}
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// WebSocket endpoint
	s.router.HandleFunc("/ws/anomalies", s.hub.ServeWS)

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/metrics", s.handleGetMetrics).Methods("GET")
	api.HandleFunc("/anomalies", s.handleGetAnomalies).Methods("GET")
	api.HandleFunc("/anomalies/{id}", s.handleGetAnomalyByID).Methods("GET")

	// CORS middleware
	s.router.Use(corsMiddleware)
	s.router.Use(loggingMiddleware)
}

func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:         ":" + s.port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("🛑 API server shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("🌐 API server started on http://localhost:%s", s.port)
	log.Printf("   Health: http://localhost:%s/health", s.port)
	log.Printf("   Metrics: http://localhost:%s/api/metrics", s.port)
	log.Printf("   Anomalies: http://localhost:%s/api/anomalies", s.port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("📡 %s %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}
func (s *Server) GetHub() *Hub {
	return s.hub
}
