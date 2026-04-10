package cli

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"

	"github.com/wesen/2026-04-09--screencast-studio/internal/web"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/app"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

type serveSettings struct {
	Addr            string `glazed:"addr"`
	File            string `glazed:"file"`
	StaticDir       string `glazed:"static-dir"`
	PreviewLimit    int    `glazed:"preview-limit"`
	ShutdownTimeout int    `glazed:"shutdown-timeout"`
}

type serveCommand struct {
	*cmds.CommandDescription
	app *app.Application
}

func newServeCommand(application *app.Application) (*serveCommand, error) {
	sections, err := defaultSections()
	if err != nil {
		return nil, err
	}

	return &serveCommand{
		CommandDescription: cmds.NewCommandDescription(
			"serve",
			cmds.WithShort("Run the local web API and control server"),
			cmds.WithFlags(
				fields.New("addr", fields.TypeString, fields.WithDefault(":7777"), fields.WithHelp("Bind address for the HTTP server")),
				fields.New("file", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional path to a setup DSL file to preload into the web UI")),
				fields.New("static-dir", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional directory to serve at / during development")),
				fields.New("preview-limit", fields.TypeInteger, fields.WithDefault(4), fields.WithHelp("Maximum number of preview workers allowed at once")),
				fields.New("shutdown-timeout", fields.TypeInteger, fields.WithDefault(5), fields.WithHelp("Graceful shutdown timeout in seconds")),
			),
			cmds.WithSections(sections...),
		),
		app: application,
	}, nil
}

func (c *serveCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, _ middlewares.Processor) error {
	settings := &serveSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode serve settings")
	}

	serverCtx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	initialDSL := ""
	initialDSLPath := strings.TrimSpace(settings.File)
	if initialDSLPath != "" {
		body, err := dsl.LoadFile(initialDSLPath)
		if err != nil {
			return err
		}
		if _, err := c.app.CompileDSL(serverCtx, body); err != nil {
			return errors.Wrap(err, "validate serve preload DSL")
		}
		initialDSL = string(body)
	}

	server := web.NewServer(serverCtx, c.app, web.Config{
		InitialDSL:      initialDSL,
		InitialDSLPath:  initialDSLPath,
		Addr:            settings.Addr,
		StaticDir:       settings.StaticDir,
		PreviewLimit:    settings.PreviewLimit,
		ShutdownTimeout: durationSeconds(settings.ShutdownTimeout),
	})

	return server.ListenAndServe(serverCtx)
}
