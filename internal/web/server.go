package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	Addr            string
	StaticDir       string
	PreviewLimit    int
	ShutdownTimeout time.Duration
}

type Server struct {
	app        ApplicationService
	config     Config
	mux        *http.ServeMux
	events     *EventHub
	recordings *RecordingManager
	previews   *PreviewManager
}

func NewServer(application ApplicationService, cfg Config) *Server {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.PreviewLimit <= 0 {
		cfg.PreviewLimit = 4
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 5 * time.Second
	}

	events := NewEventHub()
	server := &Server{
		app:        application,
		config:     cfg,
		mux:        http.NewServeMux(),
		events:     events,
		recordings: NewRecordingManager(application, events.Publish),
		previews:   NewPreviewManager(application, events.Publish, cfg.PreviewLimit, nil),
	}
	server.registerRoutes()
	return server
}

func (s *Server) Handler() http.Handler {
	return withLogging(s.mux)
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:              s.config.Addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		log.Info().
			Str("addr", s.config.Addr).
			Str("static_dir", s.config.StaticDir).
			Int("preview_limit", s.config.PreviewLimit).
			Msg("web server starting")

		err := httpServer.ListenAndServe()
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errors.Wrap(err, "listen and serve")
	})
	group.Go(func() error {
		<-groupCtx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return errors.Wrap(err, "shutdown web server")
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		return err
	}
	return nil
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration", time.Since(startedAt)).
			Msg("http request")
	})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, apiHealthResponse{
		OK:           true,
		Service:      "screencast-studio",
		PreviewLimit: s.config.PreviewLimit,
	})
}

func (s *Server) handleWebsocketPlaceholder(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "websocket transport not implemented yet", http.StatusNotImplemented)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if s.config.StaticDir != "" {
		info, err := os.Stat(s.config.StaticDir)
		if err == nil && info.IsDir() {
			http.FileServer(http.Dir(s.config.StaticDir)).ServeHTTP(w, r)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, "<!doctype html><html><body><h1>screencast-studio</h1><p>web frontend not built yet</p></body></html>")
}
