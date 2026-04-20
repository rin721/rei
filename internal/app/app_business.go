package app

import (
	"context"
	"fmt"

	"github.com/rin721/rei/internal/repository"
)

func (a *App) initBusiness() error {
	if a.handlers != nil {
		return nil
	}
	if err := a.requireBusinessInfrastructure(); err != nil {
		return err
	}

	repos := repository.NewSet(a.database.DB(), a.dbtx)
	if err := a.seedBusiness(context.Background(), repos); err != nil {
		return fmt.Errorf("seed business data: %w", err)
	}

	modules, err := a.provideBusinessModules(repos)
	if err != nil {
		return err
	}

	a.handlers = modules.Handlers()
	return nil
}

func (a *App) requireBusinessInfrastructure() error {
	if a.database == nil {
		return fmt.Errorf("init business: database is required")
	}
	if a.dbtx == nil {
		return fmt.Errorf("init business: dbtx is required")
	}
	if a.cache == nil {
		return fmt.Errorf("init business: cache is required")
	}
	if a.crypto == nil {
		return fmt.Errorf("init business: crypto service is required")
	}
	if a.jwt == nil {
		return fmt.Errorf("init business: jwt manager is required")
	}
	if a.rbac == nil {
		return fmt.Errorf("init business: rbac manager is required")
	}
	if a.idGen == nil {
		return fmt.Errorf("init business: id generator is required")
	}
	return nil
}
