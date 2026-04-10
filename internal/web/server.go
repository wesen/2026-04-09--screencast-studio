package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
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
	telemetry  *TelemetryManager
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
		telemetry:  NewTelemetryManager(events.Publish),
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
			Str("event", "runtime.http.start").
			Str("addr", s.config.Addr).
			Str("static_dir", s.config.StaticDir).
			Int("preview_limit", s.config.PreviewLimit).
			Dur("shutdown_timeout", s.config.ShutdownTimeout).
			Msg("web server starting")

		go s.openBrowser()

		err := httpServer.ListenAndServe()
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			log.Info().
				Str("event", "runtime.http.exit").
				Str("addr", s.config.Addr).
				Msg("http server exited cleanly")
			return nil
		}
		log.Error().
			Str("event", "runtime.http.exit").
			Str("addr", s.config.Addr).
			Err(err).
			Msg("http server exited with error")
		return errors.Wrap(err, "listen and serve")
	})
	group.Go(func() error {
		log.Info().
			Str("event", "runtime.telemetry.start").
			Msg("telemetry manager starting")
		err := s.telemetry.Run(groupCtx)
		if err != nil {
			log.Error().
				Str("event", "runtime.telemetry.exit").
				Err(err).
				Msg("telemetry manager exited with error")
			return err
		}
		log.Info().
			Str("event", "runtime.telemetry.exit").
			Str("reason", contextReason(groupCtx)).
			Msg("telemetry manager exited")
		return nil
	})
	group.Go(func() error {
		<-groupCtx.Done()
		log.Info().
			Str("event", "runtime.shutdown.begin").
			Str("reason", contextReason(groupCtx)).
			Dur("shutdown_timeout", s.config.ShutdownTimeout).
			Msg("runtime shutdown triggered")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		log.Info().
			Str("event", "runtime.http.shutdown.begin").
			Str("reason", contextReason(groupCtx)).
			Msg("http server shutdown starting")
		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().
				Str("event", "runtime.http.shutdown.error").
				Err(err).
				Msg("http server shutdown failed")
			return errors.Wrap(err, "shutdown web server")
		}
		log.Info().
			Str("event", "runtime.http.shutdown.done").
			Str("reason", contextReason(groupCtx)).
			Msg("http server shutdown finished")
		return nil
	})

	if err := group.Wait(); err != nil {
		log.Error().
			Str("event", "runtime.shutdown.summary").
			Err(err).
			Msg("server runtime exited with error")
		return err
	}
	log.Info().
		Str("event", "runtime.shutdown.summary").
		Msg("server runtime exited cleanly")
	return nil
}

func (s *Server) openBrowser() {
	log.Info().
		Str("event", "runtime.browser.open.begin").
		Dur("delay", 500*time.Millisecond).
		Msg("browser open scheduled")

	// Small delay to ensure server is ready
	time.Sleep(500 * time.Millisecond)

	addr := s.config.Addr
	if !strings.HasPrefix(addr, "http") {
		if strings.HasPrefix(addr, ":") {
			addr = "http://localhost" + addr
		} else {
			addr = "http://" + addr
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", addr)
	case "linux":
		cmd = exec.Command("xdg-open", addr)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", addr)
	default:
		return
	}

	if err := cmd.Start(); err != nil {
		log.Warn().
			Str("event", "runtime.browser.open.error").
			Str("url", addr).
			Err(err).
			Msg("failed to open browser")
	} else {
		log.Info().
			Str("event", "runtime.browser.open.done").
			Str("url", addr).
			Int("pid", cmd.Process.Pid).
			Msg("opened browser")
	}
}

func contextReason(ctx context.Context) string {
	if ctx == nil || ctx.Err() == nil {
		return "running"
	}
	return ctx.Err().Error()
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

	writeProtoJSON(w, http.StatusOK, mapHealthResponse(s.config.PreviewLimit))
}

func (s *Server) handleWebsocketPlaceholder(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "websocket transport not implemented yet", http.StatusNotImplemented)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if s.config.StaticDir != "" {
		info, err := os.Stat(s.config.StaticDir)
		if err == nil && info.IsDir() {
			if serveSPAFromFS(w, r, os.DirFS(s.config.StaticDir)) {
				return
			}
		}
	}

	if serveSPAFromFS(w, r, publicFS) {
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, "<!doctype html><html><body><h1>screencast-studio</h1><p>web frontend not built yet</p></body></html>")
}
