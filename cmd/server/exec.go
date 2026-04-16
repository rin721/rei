package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/rin721/rei/internal/app"
)

func runServer(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	configPath := fs.String("config", "", "配置文件路径")
	dryRun := fs.Bool("dry-run", false, "仅初始化并立即退出")

	if err := fs.Parse(args); err != nil {
		return err
	}

	application, err := app.New(app.Options{
		Mode:       app.ModeServer,
		ConfigPath: *configPath,
		DryRun:     *dryRun,
	})
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	return application.Run(ctx)
}
