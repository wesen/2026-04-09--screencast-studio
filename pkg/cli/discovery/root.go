package discovery

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/app"
)

type CobraBuilder func(cmds.Command) (*cobra.Command, error)
type SectionFactory func() ([]schema.Section, error)

func Register(parent *cobra.Command, application *app.Application, build CobraBuilder, sections SectionFactory) error {
	listCmd, err := newListCommand(application, sections)
	if err != nil {
		return errors.Wrap(err, "create discovery list command")
	}
	cobraCmd, err := build(listCmd)
	if err != nil {
		return errors.Wrap(err, "build discovery list cobra command")
	}
	parent.AddCommand(cobraCmd)
	return nil
}
