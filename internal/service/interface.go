package service

import (
	"context"

	"github.com/rei0721/go-scaffold2/types"
	typesuser "github.com/rei0721/go-scaffold2/types/user"
)

// AuthService 定义认证服务契约。
type AuthService interface {
	Register(context.Context, typesuser.RegisterRequest) (typesuser.AuthResponse, error)
	Login(context.Context, typesuser.LoginRequest) (typesuser.AuthResponse, error)
	Logout(context.Context, string) error
	ChangePassword(context.Context, string, typesuser.ChangePasswordRequest) error
	RefreshToken(context.Context, typesuser.RefreshTokenRequest) (typesuser.AuthResponse, error)
}

// UserService 定义用户服务契约。
type UserService interface {
	GetProfile(context.Context, string) (typesuser.Profile, error)
	UpdateProfile(context.Context, string, typesuser.UpdateProfileRequest) (typesuser.Profile, error)
}

// RBACService 定义 RBAC 服务契约。
type RBACService interface {
	CheckPermission(context.Context, types.CheckPermissionRequest) (types.CheckPermissionResponse, error)
	AssignRole(context.Context, types.AssignRoleRequest) error
	RevokeRole(context.Context, types.RevokeRoleRequest) error
	GetUserRoles(context.Context, string) (types.UserRolesResponse, error)
	GetUsersForRole(context.Context, string) (types.RoleUsersResponse, error)
	AddPolicy(context.Context, types.PolicyRequest) error
	RemovePolicy(context.Context, types.PolicyRequest) error
	ListPolicies(context.Context) (types.PoliciesResponse, error)
	LoadFromStore(context.Context) error
}

// SampleService 定义示例模块服务契约。
type SampleService interface {
	List(context.Context) ([]types.SampleItemResponse, error)
}
