package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Deymos01/tsv-processing/internal/transport/http/middleware"
	"go.uber.org/zap"

	"github.com/Deymos01/tsv-processing/internal/config"
	"github.com/Deymos01/tsv-processing/internal/transport/http/handler"
)

// Server wraps the standard http.Server with application-level configuration.
type Server struct {
	httpServer *http.Server
	log        *zap.Logger
}

// NewServer builds the HTTP server, registers all routes, and returns a Server.
func NewServer(
	cfg config.ServerConfig,
	messageHandler *handler.MessageHandler,
	log *zap.Logger,
) *Server {
	mux := http.NewServeMux()
	registerRoutes(mux, messageHandler, log)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      mux,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		log: log,
	}
}

// Run starts listening and serving HTTP requests.
// Returns when the server is stopped or an error occurs.
func (s *Server) Run() error {
	s.log.Info("HTTP server starting", zap.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

// Shutdown gracefully drains active connections with the given context deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("HTTP server shutting down")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}
	return nil
}

// registerRoutes attaches all handlers to the mux, wrapped with middleware.
func registerRoutes(mux *http.ServeMux, messageHandler *handler.MessageHandler, log *zap.Logger) {
	loggingMW := middleware.Logging(log)

	mux.Handle("GET /api/v1/messages", loggingMW(http.HandlerFunc(messageHandler.GetByUnitGUID)))
}
