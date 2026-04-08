package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/rei0721/go-scaffold2/internal/app"
)

func runInitDB(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("initdb", flag.ContinueOnError)
	configPath := fs.String("config", "", "配置文件路径")
	dryRun := fs.Bool("dry-run", false, "仅初始化并立即退出")

	if err := fs.Parse(args); err != nil {
		return err
	}

	application, err := app.New(app.Options{
		Mode:       app.ModeInitDB,
		ConfigPath: *configPath,
		DryRun:     *dryRun,
	})
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	return application.Run(ctx)
}
