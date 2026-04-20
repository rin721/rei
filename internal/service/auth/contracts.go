package auth

import (
	"context"
	"time"

	domainuser "github.com/rin721/rei/internal/domain/user"
)

// UseCase defines the authentication application contract.
type UseCase interface {
	Register(context.Context, RegisterCommand) (Authentication, error)
	Login(context.Context, LoginCommand) (Authentication, error)
	Logout(context.Context, LogoutCommand) error
	ChangePassword(context.Context, ChangePasswordCommand) error
	RefreshToken(context.Context, RefreshTokenCommand) (Authentication, error)
}

// RegisterCommand describes the input needed to register a new user.
type RegisterCommand struct {
	Username    string
	Password    string
	DisplayName string
	Email       string
}

// LoginCommand describes the input needed to authenticate a user.
type LoginCommand struct {
	Username string
	Password string
}

// LogoutCommand describes the input needed to end the current session.
type LogoutCommand struct {
	UserID string
}

// ChangePasswordCommand describes the input needed to change a password.
type ChangePasswordCommand struct {
	UserID      string
	OldPassword string
	NewPassword string
}

// RefreshTokenCommand describes the input needed to refresh tokens.
type RefreshTokenCommand struct {
	RefreshToken string
}

// Authentication is the application-layer result for auth flows.
type Authentication struct {
	User   AuthenticatedUser
	Tokens TokenPair
}

// AuthenticatedUser is the user projection returned by auth flows.
type AuthenticatedUser struct {
	ID          string
	Username    string
	DisplayName string
	Email       string
	Roles       []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TokenPair contains the issued access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
}

// Role describes a role that must exist for auth flows.
type Role struct {
	ID          string
	Name        string
	Description string
}

// RoleBinding describes a user-to-role relationship.
type RoleBinding struct {
	ID       string
	UserID   string
	RoleName string
}

// UserStore defines the persistence port for auth-facing user operations.
type UserStore interface {
	FindByID(context.Context, string) (*domainuser.User, error)
	FindByUsername(context.Context, string) (*domainuser.User, error)
	Create(context.Context, *domainuser.User) error
	Save(context.Context, *domainuser.User) error
	Count(context.Context) (int64, error)
}

// RoleStore defines the persistence port for role definitions.
type RoleStore interface {
	Ensure(context.Context, Role) error
}

// RoleBindingStore defines the persistence port for user-role bindings.
type RoleBindingStore interface {
	Assign(context.Context, RoleBinding) error
	ListRolesByUser(context.Context, string) ([]string, error)
}

// IDProvider defines ID generation capability.
type IDProvider interface {
	NextID() (int64, error)
}

// PasswordManager defines password hashing and verification.
type PasswordManager interface {
	HashPassword(string) (string, error)
	ComparePassword(string, string) error
}

// TokenManager defines access/refresh token generation and refresh token validation.
type TokenManager interface {
	GenerateAccessToken(string, map[string]any) (string, error)
	GenerateRefreshToken(string, map[string]any) (string, error)
	ValidateRefreshToken(string) (string, error)
}

// RefreshTokenStore defines refresh-token session persistence.
type RefreshTokenStore interface {
	GetRefreshToken(context.Context, string) (string, bool, error)
	SetRefreshToken(context.Context, string, string, time.Duration) error
	DeleteRefreshToken(context.Context, string) error
}

// TransactionManager defines the transaction boundary used by auth flows.
type TransactionManager interface {
	WithTx(context.Context, func(context.Context) error) error
}

// RoleManager defines the role synchronization behavior needed by auth flows.
type RoleManager interface {
	AssignRole(string, string) error
}
