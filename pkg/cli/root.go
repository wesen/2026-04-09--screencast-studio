package cli

import (
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/app"
	discoverycmd "github.com/wesen/2026-04-09--screencast-studio/pkg/cli/discovery"
	setupcmd "github.com/wesen/2026-04-09--screencast-studio/pkg/cli/setup"
)

func NewRootCommand() (*cobra.Command, error) {
	application := app.New()

	rootCmd := &cobra.Command{
		Use:   "screencast-studio",
		Short: "CLI-first screencast studio for discovery, setup compile, and recording",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}

	if err := logging.AddLoggingSectionToRootCommand(rootCmd, "screencast-studio"); err != nil {
		return nil, errors.Wrap(err, "add logging section")
	}

	helpSystem := help.NewHelpSystem()
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	discoveryRoot := &cobra.Command{Use: "discovery", Short: "Inspect available capture sources"}
	setupRoot := &cobra.Command{Use: "setup", Short: "Work with setup files"}

	rootCmd.AddCommand(discoveryRoot, setupRoot)

	if err := discoverycmd.Register(discoveryRoot, application, buildCobra, defaultSections); err != nil {
		return nil, errors.Wrap(err, "register discovery commands")
	}
	if err := setupcmd.Register(setupRoot, application, buildCobra, defaultSections); err != nil {
		return nil, errors.Wrap(err, "register setup commands")
	}

	recordCmd, err := newRecordCommand(application)
	if err != nil {
		return nil, errors.Wrap(err, "create record command")
	}
	recordCobra, err := buildCobra(recordCmd)
	if err != nil {
		return nil, errors.Wrap(err, "build record cobra command")
	}
	rootCmd.AddCommand(recordCobra)

	return rootCmd, nil
}
