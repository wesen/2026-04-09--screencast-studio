package main

import (
	"os"

	"github.com/pkg/errors"

	cliapp "github.com/wesen/2026-04-09--screencast-studio/pkg/cli"
)

func main() {
	rootCmd, err := cliapp.NewRootCommand()
	if err != nil {
		_, _ = os.Stderr.WriteString(errors.Wrap(err, "create root command").Error() + "\n")
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
