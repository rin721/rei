package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rin721/go-scaffold2/pkg/cli"
)

func run(args []string) int {
	app, err := buildCLI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build cli: %v\n", err)
		return cli.ExitCodeExecution
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return app.Run(ctx, args)
}
