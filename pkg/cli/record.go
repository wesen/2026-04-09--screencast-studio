package cli

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/app"
)

type recordSettings struct {
	File      string `glazed:"file"`
	DryRun    bool   `glazed:"dry-run"`
	PrintPlan bool   `glazed:"print-plan"`
	Duration  int    `glazed:"duration"`
}

type recordCommand struct {
	*cmds.CommandDescription
	app *app.Application
}

func newRecordCommand(application *app.Application) (*recordCommand, error) {
	sections, err := defaultSections()
	if err != nil {
		return nil, err
	}

	return &recordCommand{
		CommandDescription: cmds.NewCommandDescription(
			"record",
			cmds.WithShort("Compile a setup file and execute it"),
			cmds.WithFlags(
				fields.New("file", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Path to the setup DSL file")),
				fields.New("dry-run", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Compile only; do not execute capture")),
				fields.New("print-plan", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Print the compiled plan before execution")),
				fields.New("duration", fields.TypeInteger, fields.WithDefault(0), fields.WithHelp("Stop after N seconds; 0 means run until interrupted")),
			),
			cmds.WithSections(sections...),
		),
		app: application,
	}, nil
}

func (c *recordCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	settings := &recordSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode record settings")
	}

	if settings.DryRun {
		summary, err := c.app.CompileFile(ctx, settings.File)
		if err != nil {
			return err
		}
		return gp.AddRow(ctx, types.NewRow(
			types.MRP("operation", "record.dry-run"),
			types.MRP("file", settings.File),
			types.MRP("session_id", summary.SessionID),
			types.MRP("output_count", len(summary.Outputs)),
			types.MRP("warning_count", len(summary.Warnings)),
			types.MRP("warnings", strings.Join(summary.Warnings, "; ")),
			types.MRP("print_plan", settings.PrintPlan),
		))
	}

	recordCtx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	if settings.PrintPlan {
		summary, err := c.app.CompileFile(recordCtx, settings.File)
		if err != nil {
			return err
		}
		for i, output := range summary.Outputs {
			if err := gp.AddRow(ctx, types.NewRow(
				types.MRP("operation", "record.plan"),
				types.MRP("file", settings.File),
				types.MRP("session_id", summary.SessionID),
				types.MRP("output_index", i),
				types.MRP("output_count", len(summary.Outputs)),
				types.MRP("warning_count", len(summary.Warnings)),
				types.MRP("warnings", strings.Join(summary.Warnings, "; ")),
				types.MRP("kind", output.Kind),
				types.MRP("source_id", output.SourceID),
				types.MRP("name", output.Name),
				types.MRP("path", output.Path),
			)); err != nil {
				return err
			}
		}
	}

	summary, err := c.app.RecordFile(recordCtx, settings.File, app.RecordOptions{
		GracePeriod: 5 * time.Second,
		MaxDuration: time.Duration(settings.Duration) * time.Second,
	})
	if err != nil {
		return err
	}

	for i, output := range summary.Outputs {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("operation", "record"),
			types.MRP("file", settings.File),
			types.MRP("session_id", summary.SessionID),
			types.MRP("output_index", i),
			types.MRP("output_count", len(summary.Outputs)),
			types.MRP("warning_count", len(summary.Warnings)),
			types.MRP("warnings", strings.Join(summary.Warnings, "; ")),
			types.MRP("kind", output.Kind),
			types.MRP("source_id", output.SourceID),
			types.MRP("name", output.Name),
			types.MRP("path", output.Path),
			types.MRP("status", "completed"),
			types.MRP("final_state", summary.State),
			types.MRP("state_reason", summary.Reason),
			types.MRP("started_at", summary.StartedAt.Format(time.RFC3339)),
			types.MRP("finished_at", summary.FinishedAt.Format(time.RFC3339)),
			types.MRP("duration_seconds", int(summary.FinishedAt.Sub(summary.StartedAt).Seconds())),
		)); err != nil {
			return err
		}
	}

	return nil
}
