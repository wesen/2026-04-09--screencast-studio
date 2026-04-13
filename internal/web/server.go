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

	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
)

type Config struct {
	InitialDSL      string
	InitialDSLPath  string
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
	parentCtx  context.Context
	recordings *RecordingManager
	previews   *PreviewManager
	telemetry  *TelemetryManager
}

type serverOptions struct {
	previewRuntime media.PreviewRuntime
}

type ServerOption func(*serverOptions)

func WithPreviewRuntime(runtime media.PreviewRuntime) ServerOption {
	return func(opts *serverOptions) {
		opts.previewRuntime = runtime
	}
}

func NewServer(parentCtx context.Context, application ApplicationService, cfg Config) *Server {
	return NewServerWithOptions(parentCtx, application, cfg)
}

func NewServerWithOptions(parentCtx context.Context, application ApplicationService, cfg Config, opts ...ServerOption) *Server {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.PreviewLimit <= 0 {
		cfg.PreviewLimit = 4
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 5 * time.Second
	}
	if parentCtx == nil {
		parentCtx = context.Background()
	}

	resolvedOpts := serverOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&resolvedOpts)
		}
	}

	events := NewEventHub()
	server := &Server{
		app:       application,
		config:    cfg,
		mux:       http.NewServeMux(),
		events:    events,
		parentCtx: parentCtx,
		telemetry: NewTelemetryManager(events.Publish),
	}
	server.recordings = NewRecordingManager(parentCtx, application, events.Publish)
	server.previews = NewPreviewManager(parentCtx, application, events.Publish, cfg.PreviewLimit, resolvedOpts.previewRuntime)
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
	httpDone := make(chan struct{})
	telemetryDone := make(chan struct{})

	log.Info().
		Str("event", "runtime.context.bind").
		Msg("recording and preview managers already bound to serve runtime context")

	group.Go(func() error {
		defer close(httpDone)
		log.Info().
			Str("event", "runtime.http.start").
			Str("addr", s.config.Addr).
			Str("static_dir", s.config.StaticDir).
			Str("initial_dsl_path", s.config.InitialDSLPath).
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
		defer close(telemetryDone)
		// Telemetry remains context-driven for now: Run(ctx) exits when the
		// runtime context is canceled, and the server explicitly waits for this
		// goroutine during shutdown rather than introducing a second shutdown API.
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

		// Shutdown order is intentional:
		//  1. stop accepting new HTTP work via http.Server.Shutdown,
		//  2. explicitly drain serve-owned recording sessions,
		//  3. explicitly drain serve-owned previews,
		//  4. wait for the HTTP and telemetry goroutines to report exit,
		//  5. emit a final component summary before returning.
		//
		// This keeps new work from entering the system while still giving
		// already-started workers a bounded window to exit cleanly.
		log.Info().
			Str("event", "runtime.shutdown.begin").
			Str("reason", contextReason(groupCtx)).
			Dur("shutdown_timeout", s.config.ShutdownTimeout).
			Msg("runtime shutdown triggered")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		shutdownErrs := []string{}

		log.Info().
			Str("event", "runtime.http.shutdown.begin").
			Str("reason", contextReason(groupCtx)).
			Msg("http server shutdown starting")
		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().
				Str("event", "runtime.http.shutdown.error").
				Err(err).
				Msg("http server shutdown failed")
			shutdownErrs = append(shutdownErrs, errors.Wrap(err, "shutdown web server").Error())
		} else {
			log.Info().
				Str("event", "runtime.http.shutdown.done").
				Str("reason", contextReason(groupCtx)).
				Msg("http server shutdown finished")
		}

		log.Info().
			Str("event", "runtime.recordings.shutdown.begin").
			Msg("recording manager shutdown starting")
		if err := s.recordings.Shutdown(shutdownCtx); err != nil {
			log.Error().
				Str("event", "runtime.recordings.shutdown.error").
				Err(err).
				Msg("recording manager shutdown failed")
			shutdownErrs = append(shutdownErrs, errors.Wrap(err, "shutdown recordings").Error())
		} else {
			log.Info().
				Str("event", "runtime.recordings.shutdown.done").
				Msg("recording manager shutdown finished")
		}

		log.Info().
			Str("event", "runtime.previews.shutdown.begin").
			Msg("preview manager shutdown starting")
		if err := s.previews.Shutdown(shutdownCtx); err != nil {
			log.Error().
				Str("event", "runtime.previews.shutdown.error").
				Err(err).
				Msg("preview manager shutdown failed")
			shutdownErrs = append(shutdownErrs, errors.Wrap(err, "shutdown previews").Error())
		} else {
			log.Info().
				Str("event", "runtime.previews.shutdown.done").
				Msg("preview manager shutdown finished")
		}

		if err := waitForRuntimeParticipant(shutdownCtx, httpDone, "http server goroutine"); err != nil {
			shutdownErrs = append(shutdownErrs, err.Error())
		}
		log.Info().
			Str("event", "runtime.telemetry.wait.begin").
			Msg("waiting for telemetry goroutine to exit")
		if err := waitForRuntimeParticipant(shutdownCtx, telemetryDone, "telemetry goroutine"); err != nil {
			shutdownErrs = append(shutdownErrs, err.Error())
		} else {
			log.Info().
				Str("event", "runtime.telemetry.wait.done").
				Msg("telemetry goroutine exited")
		}

		finalSession := s.recordings.Current()
		remainingPreviews := s.previews.List()
		log.Info().
			Str("event", "runtime.shutdown.components").
			Bool("recording_active", finalSession.Active).
			Str("recording_session_id", finalSession.SessionID).
			Int("remaining_previews", len(remainingPreviews)).
			Msg("runtime shutdown component summary")

		if len(shutdownErrs) > 0 {
			return errors.New(strings.Join(shutdownErrs, "; "))
		}
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

func waitForRuntimeParticipant(ctx context.Context, done <-chan struct{}, name string) error {
	if done == nil {
		return nil
	}
	log.Info().
		Str("event", "runtime.participant.wait.begin").
		Str("participant", name).
		Msg("waiting for runtime participant to exit")
	select {
	case <-done:
		log.Info().
			Str("event", "runtime.participant.wait.done").
			Str("participant", name).
			Msg("runtime participant exited")
		return nil
	case <-ctx.Done():
		log.Error().
			Str("event", "runtime.participant.wait.timeout").
			Str("participant", name).
			Err(ctx.Err()).
			Msg("timed out waiting for runtime participant to exit")
		return errors.Wrapf(ctx.Err(), "wait for %s", name)
	}
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

	writeProtoJSON(w, http.StatusOK, mapHealthResponse(s.config.PreviewLimit, s.config.InitialDSL, s.config.InitialDSLPath))
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
