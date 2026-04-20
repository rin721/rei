package app

import (
	"fmt"

	"github.com/rin721/rei/internal/handler"
	"github.com/rin721/rei/internal/repository"
	"github.com/rin721/rei/internal/service"
	authservice "github.com/rin721/rei/internal/service/auth"
	rbacservice "github.com/rin721/rei/internal/service/rbac"
	userservice "github.com/rin721/rei/internal/service/user"
)

type businessModules struct {
	Auth   authservice.UseCase
	User   userservice.UseCase
	RBAC   rbacservice.UseCase
	Sample service.SampleService
}

func (m *businessModules) Handlers() *handler.Bundle {
	return handler.NewBundle(m.Auth, m.User, m.RBAC, m.Sample)
}

func (p businessProvisioning) provideModules(repos *repository.Set) (*businessModules, error) {
	modules := &businessModules{}

	authModule, err := (authModuleProvider{}).Provide(p, repos)
	if err != nil {
		return nil, fmt.Errorf("init auth service: %w", err)
	}
	modules.Auth = authModule

	userModule, err := (userModuleProvider{}).Provide(p, repos)
	if err != nil {
		return nil, fmt.Errorf("init user service: %w", err)
	}
	modules.User = userModule

	rbacModule, err := (rbacModuleProvider{}).Provide(p, repos)
	if err != nil {
		return nil, fmt.Errorf("init rbac service: %w", err)
	}
	modules.RBAC = rbacModule

	sampleModule, err := (sampleModuleProvider{}).Provide(p, repos)
	if err != nil {
		return nil, fmt.Errorf("init sample service: %w", err)
	}
	modules.Sample = sampleModule

	return modules, nil
}
