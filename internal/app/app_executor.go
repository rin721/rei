package app

import (
	"fmt"

	pkgexecutor "github.com/rin721/rei/pkg/executor"
)

func (p infrastructureProvisioning) initExecutor() error {
	if p.infra.executor != nil || !p.cfg.Executor.Enabled {
		return nil
	}

	manager, err := pkgexecutor.NewManager(toExecutorConfig(p.cfg.Executor))
	if err != nil {
		return fmt.Errorf("init executor: %w", err)
	}

	p.infra.executor = manager
	p.syncExecutorBindings()
	return nil
}
