package discovery

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

type listSettings struct {
	Kind string `glazed:"kind"`
}

type listCommand struct {
	*cmds.CommandDescription
	app *app.Application
}

func newListCommand(application *app.Application, sectionFactory SectionFactory) (*listCommand, error) {
	sections, err := sectionFactory()
	if err != nil {
		return nil, err
	}

	return &listCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List available capture sources"),
			cmds.WithFlags(
				fields.New("kind", fields.TypeString, fields.WithDefault("all"), fields.WithHelp("Discovery kind: display, window, camera, audio, or all")),
			),
			cmds.WithSections(sections...),
		),
		app: application,
	}, nil
}

func (c *listCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	settings := &listSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode list settings")
	}

	rows, err := c.app.DiscoveryList(ctx, settings.Kind)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		return gp.AddRow(ctx, types.NewRow(
			types.MRP("kind", settings.Kind),
			types.MRP("status", "no-results"),
		))
	}

	for _, row := range rows {
		if err := gp.AddRow(ctx, types.NewRowFromMap(row)); err != nil {
			return err
		}
	}
	return nil
}
