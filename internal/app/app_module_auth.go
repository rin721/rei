package app

import (
	authadapter "github.com/rin721/rei/internal/adapter/auth"
	"github.com/rin721/rei/internal/repository"
	authservice "github.com/rin721/rei/internal/service/auth"
)

type authModuleProvider struct{}

func (authModuleProvider) Provide(deps businessProvisioning, repos *repository.Set) (authservice.UseCase, error) {
	return authservice.New(authservice.Dependencies{
		Users:           authadapter.NewUserStore(repos.Users),
		Roles:           authadapter.NewRoleStore(repos.Roles),
		RoleBindings:    authadapter.NewRoleBindingStore(repos.UserRoles),
		IDProvider:      deps.idGen,
		Password:        deps.crypto,
		Tokens:          authadapter.NewTokenManager(deps.jwt),
		RefreshTokens:   authadapter.NewRefreshTokenStore(deps.cache),
		Tx:              authadapter.NewTransactionManager(deps.dbtx),
		RoleManager:     deps.rbac,
		RefreshTokenTTL: deps.refreshTokenTTL,
	})
}
