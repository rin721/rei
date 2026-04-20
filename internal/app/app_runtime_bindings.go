package app

func (a *App) syncExecutorBindings() {
	if a.executor == nil {
		return
	}
	if a.logger != nil {
		a.logger.SetExecutor(a.executor)
	}
	if a.httpServer != nil {
		a.httpServer.SetExecutor(a.executor)
	}
}
