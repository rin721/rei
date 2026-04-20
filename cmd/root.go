package main

import (
	"github.com/rin721/rei/cmd/db"
	"github.com/rin721/rei/cmd/server"
	"github.com/rin721/rei/pkg/cli"
	"github.com/rin721/rei/types/constants"
)

const applicationSummary = "Modular layered Go backend scaffold"

func buildApp() *cli.App {
	root := cli.BuildRootCmd(
		constants.ApplicationName,
		applicationSummary,
		&server.Cmd{},
		&db.Cmd{},
	)

	return cli.NewApp(root)
}
