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
)

type validateSettings struct {
	File string `glazed:"file"`
}

type validateCommand struct {
	*cmds.CommandDescription
}

func newValidateCommand(sectionFactory SectionFactory) (*validateCommand, error) {
	sections, err := sectionFactory()
	if err != nil {
		return nil, err
	}

	return &validateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"validate",
			cmds.WithShort("Validate a setup file"),
			cmds.WithFlags(
				fields.New("file", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Path to the setup DSL file")),
			),
			cmds.WithSections(sections...),
		),
	}, nil
}

func (c *validateCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	settings := &validateSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode validate settings")
	}

	return gp.AddRow(ctx, types.NewRow(
		types.MRP("operation", "setup.validate"),
		types.MRP("file", settings.File),
		types.MRP("status", "not-implemented"),
	))
}
