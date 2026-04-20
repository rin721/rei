package app

import (
	"time"

	authadapter "github.com/rin721/rei/internal/adapter/auth"
	"github.com/rin721/rei/internal/repository"
	authservice "github.com/rin721/rei/internal/service/auth"
)

type authModuleProvider struct{}

func (authModuleProvider) Provide(a *App, repos *repository.Set) (authservice.UseCase, error) {
	return authservice.New(authservice.Dependencies{
		Users:           authadapter.NewUserStore(repos.Users),
		Roles:           authadapter.NewRoleStore(repos.Roles),
		RoleBindings:    authadapter.NewRoleBindingStore(repos.UserRoles),
		IDProvider:      a.idGen,
		Password:        a.crypto,
		Tokens:          authadapter.NewTokenManager(a.jwt),
		RefreshTokens:   authadapter.NewRefreshTokenStore(a.cache),
		Tx:              authadapter.NewTransactionManager(a.dbtx),
		RoleManager:     a.rbac,
		RefreshTokenTTL: time.Duration(a.cfg.JWT.RefreshTokenTTLHours) * time.Hour,
	})
}
