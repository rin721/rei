package handler

import (
	"github.com/rin721/rei/internal/service"
	authservice "github.com/rin721/rei/internal/service/auth"
	rbacservice "github.com/rin721/rei/internal/service/rbac"
	userservice "github.com/rin721/rei/internal/service/user"
)

// Bundle groups all HTTP handlers.
type Bundle struct {
	Auth   *AuthHandler
	User   *UserHandler
	RBAC   *RBACHandler
	Sample *SampleHandler
}

// NewBundle creates the HTTP handler bundle.
func NewBundle(authService authservice.UseCase, userService userservice.UseCase, rbacService rbacservice.UseCase, sampleService service.SampleService) *Bundle {
	return &Bundle{
		Auth:   NewAuthHandler(authService),
		User:   NewUserHandler(userService),
		RBAC:   NewRBACHandler(rbacService),
		Sample: NewSampleHandler(sampleService),
	}
}
