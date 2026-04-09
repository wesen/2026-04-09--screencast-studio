package cli

import (
	"context"

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
			types.MRP("print_plan", settings.PrintPlan),
		))
	}

	if err := c.app.RecordFile(ctx, settings.File); err != nil {
		return err
	}

	return gp.AddRow(ctx, types.NewRow(
		types.MRP("operation", "record"),
		types.MRP("file", settings.File),
		types.MRP("status", "started"),
		types.MRP("print_plan", settings.PrintPlan),
	))
}
