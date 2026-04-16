package main

import (
	"fmt"

	"github.com/rin721/rei/pkg/cli"
	"github.com/rin721/rei/types/constants"
)

const cliSummary = "phased modular Go backend scaffold"

func buildCLI() (*cli.App, error) {
	app := cli.New(constants.ApplicationName, cliSummary)

	commands := []cli.Command{
		{
			Name:        "server",
			Description: "start server mode with config loading and app container wiring",
			Run:         runServer,
		},
		{
			Name:        "initdb",
			Description: "start initdb mode with config loading and minimal workflow wiring",
			Run:         runInitDB,
		},
	}

	for _, command := range commands {
		if err := app.Register(command); err != nil {
			return nil, fmt.Errorf("register command %q: %w", command.Name, err)
		}
	}

	return app, nil
}
