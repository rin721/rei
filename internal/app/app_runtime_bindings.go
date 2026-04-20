package app

import (
	"context"

	pkgexecutor "github.com/rin721/rei/pkg/executor"
	pkghttpserver "github.com/rin721/rei/pkg/httpserver"
)

const defaultExecutorPoolName pkgexecutor.PoolName = "default"

type executorAsyncSubmitter struct {
	manager  pkgexecutor.Manager
	poolName pkgexecutor.PoolName
}

func newExecutorAsyncSubmitter(manager pkgexecutor.Manager) pkghttpserver.AsyncSubmitter {
	if manager == nil {
		return nil
	}

	return executorAsyncSubmitter{
		manager:  manager,
		poolName: defaultExecutorPoolName,
	}
}

func (s executorAsyncSubmitter) SubmitDefault(_ context.Context, task func()) error {
	return s.manager.Execute(s.poolName, task)
}

func (p infrastructureProvisioning) syncExecutorBindings() {
	if p.delivery.httpServer == nil {
		return
	}

	p.delivery.httpServer.SetExecutor(newExecutorAsyncSubmitter(p.infra.executor))
}
