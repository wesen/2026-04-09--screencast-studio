package setup

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
	commands := []cmds.Command{}

	compileCmd, err := newCompileCommand(application, sections)
	if err != nil {
		return errors.Wrap(err, "create setup compile command")
	}
	commands = append(commands, compileCmd)

	validateCmd, err := newValidateCommand(sections)
	if err != nil {
		return errors.Wrap(err, "create setup validate command")
	}
	commands = append(commands, validateCmd)

	for _, command := range commands {
		cobraCmd, err := build(command)
		if err != nil {
			return errors.Wrap(err, "build setup cobra command")
		}
		parent.AddCommand(cobraCmd)
	}

	return nil
}
