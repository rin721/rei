package service

import (
	"context"

	authservice "github.com/rin721/rei/internal/service/auth"
	rbacservice "github.com/rin721/rei/internal/service/rbac"
	userservice "github.com/rin721/rei/internal/service/user"
	"github.com/rin721/rei/types"
)

// AuthService aliases the auth module usecase contract.
type AuthService = authservice.UseCase

// UserService aliases the user module usecase contract.
type UserService = userservice.UseCase

// RBACService aliases the RBAC module usecase contract.
type RBACService = rbacservice.UseCase

// SampleService defines the sample module contract.
type SampleService interface {
	List(context.Context) ([]types.SampleItemResponse, error)
}
