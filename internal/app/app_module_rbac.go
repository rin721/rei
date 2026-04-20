package app

import (
	"context"

	rbacadapter "github.com/rin721/rei/internal/adapter/rbac"
	"github.com/rin721/rei/internal/repository"
	rbacservice "github.com/rin721/rei/internal/service/rbac"
)

type rbacModuleProvider struct{}

func (rbacModuleProvider) Provide(a *App, repos *repository.Set) (rbacservice.UseCase, error) {
	svc, err := rbacservice.New(rbacservice.Dependencies{
		Users:      rbacadapter.NewUserLookup(repos.Users),
		Roles:      rbacadapter.NewRoleStore(repos.Roles),
		RoleBinds:  rbacadapter.NewRoleBindingStore(repos.UserRoles),
		Policies:   rbacadapter.NewPolicyStore(repos.Policies),
		IDProvider: a.idGen,
		Tx:         rbacadapter.NewTransactionManager(a.dbtx),
		Enforcer:   rbacadapter.NewEnforcer(a.rbac),
	})
	if err != nil {
		return nil, err
	}
	if err := svc.LoadFromStore(context.Background()); err != nil {
		return nil, err
	}
	return svc, nil
}
