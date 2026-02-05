package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/config"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	config     *config.Config
}

// New creates a new HTTP server
func New(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.GetAddress(),
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		config: cfg,
	}
}

// Start begins listening for requests and handles graceful shutdown
func (s *Server) Start() error {
	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Info().Str("address", s.config.GetAddress()).Msg("server starting")
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		log.Info().Str("signal", sig.String()).Msg("shutdown signal received")

		// Give outstanding requests time to complete
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			s.httpServer.Close()
			return fmt.Errorf("could not gracefully shutdown: %w", err)
		}

		log.Info().Msg("server stopped gracefully")
	}

	return nil
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
