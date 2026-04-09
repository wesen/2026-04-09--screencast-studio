package setup

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

type compileSettings struct {
	File string `glazed:"file"`
}

type compileCommand struct {
	*cmds.CommandDescription
	app *app.Application
}

func newCompileCommand(application *app.Application, sectionFactory SectionFactory) (*compileCommand, error) {
	sections, err := sectionFactory()
	if err != nil {
		return nil, err
	}

	return &compileCommand{
		CommandDescription: cmds.NewCommandDescription(
			"compile",
			cmds.WithShort("Compile a setup file into an execution plan"),
			cmds.WithFlags(
				fields.New("file", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Path to the setup DSL file")),
			),
			cmds.WithSections(sections...),
		),
		app: application,
	}, nil
}

func (c *compileCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	settings := &compileSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode compile settings")
	}

	summary, err := c.app.CompileFile(ctx, settings.File)
	if err != nil {
		return err
	}

	return gp.AddRow(ctx, types.NewRow(
		types.MRP("operation", "setup.compile"),
		types.MRP("file", settings.File),
		types.MRP("session_id", summary.SessionID),
	))
}
