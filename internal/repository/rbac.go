package repository

import (
	"context"

	"github.com/rei0721/go-scaffold2/internal/models"
)

// RoleRepository 定义角色仓储契约。
type RoleRepository interface {
	Repository[models.Role]
	Ensure(context.Context, *models.Role) error
	FindByName(context.Context, string) (*models.Role, error)
	List(context.Context) ([]models.Role, error)
}

// UserRoleRepository 定义用户角色关系仓储契约。
type UserRoleRepository interface {
	Repository[models.UserRole]
	Assign(context.Context, *models.UserRole) error
	Revoke(context.Context, string, string) error
	ListRolesByUser(context.Context, string) ([]string, error)
	ListUsersByRole(context.Context, string) ([]string, error)
	List(context.Context) ([]models.UserRole, error)
}

// PolicyRepository 定义策略仓储契约。
type PolicyRepository interface {
	Repository[models.Policy]
	Add(context.Context, *models.Policy) error
	Remove(context.Context, string, string, string) error
	List(context.Context) ([]models.Policy, error)
}
