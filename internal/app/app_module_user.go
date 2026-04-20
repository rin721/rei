package app

import (
	"github.com/rin721/rei/internal/repository"
	userservice "github.com/rin721/rei/internal/service/user"
)

type userModuleProvider struct{}

func (userModuleProvider) Provide(_ *App, repos *repository.Set) (userservice.UseCase, error) {
	return userservice.New(userservice.Dependencies{
		Users:     repository.NewUserDomainStore(repos.Users),
		UserRoles: repository.NewUserRoleBindingReader(repos.UserRoles),
	})
}
