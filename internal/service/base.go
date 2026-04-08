package service

import (
	"context"
	"time"

	pkgjwt "github.com/rei0721/go-scaffold2/pkg/jwt"
	pkgrbac "github.com/rei0721/go-scaffold2/pkg/rbac"
	"gorm.io/gorm"
)

// IDProvider 定义 ID 生成能力。
type IDProvider interface {
	NextID() (int64, error)
}

// PasswordManager 定义密码处理能力。
type PasswordManager interface {
	HashPassword(string) (string, error)
	ComparePassword(string, string) error
}

// TokenManager 定义 JWT 管理能力。
type TokenManager interface {
	GenerateToken(string, pkgjwt.TokenType, map[string]any) (string, error)
	ValidateToken(string) (*pkgjwt.Claims, error)
}

// RoleManager 定义 RBAC 管理能力。
type RoleManager interface {
	CheckPermission(string, string, string) (bool, error)
	AssignRole(string, string) error
	RevokeRole(string, string) error
	GetUserRoles(string) ([]string, error)
	GetUsersForRole(string) ([]string, error)
	AddPolicy(string, string, string) error
	RemovePolicy(string, string, string) error
	GetPolicies() ([]pkgrbac.Policy, error)
}

// CacheStore 定义缓存依赖能力。
type CacheStore interface {
	Get(context.Context, string) (any, bool)
	Set(context.Context, string, any, time.Duration) error
	Delete(context.Context, string) error
}

// TxManager 定义事务边界能力。
type TxManager interface {
	WithTx(context.Context, func(context.Context, *gorm.DB) error) error
}
