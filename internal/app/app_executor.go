package app

import (
	"fmt"

	pkgexecutor "github.com/rin721/rei/pkg/executor"
)

func (a *App) initExecutor() error {
	if a.executor != nil || !a.cfg.Executor.Enabled {
		return nil
	}

	manager, err := pkgexecutor.New(toExecutorConfig(a.cfg.Executor))
	if err != nil {
		return fmt.Errorf("init executor: %w", err)
	}

	a.executor = manager
	a.syncExecutorBindings()
	return nil
}
