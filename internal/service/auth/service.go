package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rin721/rei/internal/models"
	"github.com/rin721/rei/internal/repository"
	"github.com/rin721/rei/internal/service"
	pkgjwt "github.com/rin721/rei/pkg/jwt"
	apperrors "github.com/rin721/rei/types/errors"
	typesuser "github.com/rin721/rei/types/user"
	"gorm.io/gorm"
)

// Dependencies 描述认证服务依赖。
type Dependencies struct {
	Users           repository.UserRepository
	Roles           repository.RoleRepository
	UserRoles       repository.UserRoleRepository
	IDProvider      service.IDProvider
	Password        service.PasswordManager
	Tokens          service.TokenManager
	Cache           service.CacheStore
	Tx              service.TxManager
	RoleManager     service.RoleManager
	RefreshTokenTTL time.Duration
}

// Service 实现认证业务流程。
type Service struct {
	deps Dependencies
}

// New 创建认证服务。
func New(deps Dependencies) (*Service, error) {
	switch {
	case deps.Users == nil:
		return nil, fmt.Errorf("users repository is required")
	case deps.Roles == nil:
		return nil, fmt.Errorf("roles repository is required")
	case deps.UserRoles == nil:
		return nil, fmt.Errorf("user roles repository is required")
	case deps.IDProvider == nil:
		return nil, fmt.Errorf("id provider is required")
	case deps.Password == nil:
		return nil, fmt.Errorf("password manager is required")
	case deps.Tokens == nil:
		return nil, fmt.Errorf("token manager is required")
	case deps.Cache == nil:
		return nil, fmt.Errorf("cache store is required")
	case deps.Tx == nil:
		return nil, fmt.Errorf("tx manager is required")
	case deps.RoleManager == nil:
		return nil, fmt.Errorf("role manager is required")
	}

	if deps.RefreshTokenTTL <= 0 {
		deps.RefreshTokenTTL = 72 * time.Hour
	}

	return &Service{deps: deps}, nil
}

// Register 创建用户并返回认证信息。
func (s *Service) Register(ctx context.Context, req typesuser.RegisterRequest) (typesuser.AuthResponse, error) {
	username := normalizeUsername(req.Username)
	if username == "" {
		return typesuser.AuthResponse{}, apperrors.BadRequest("username is required")
	}
	if len(strings.TrimSpace(req.Password)) < 8 {
		return typesuser.AuthResponse{}, apperrors.BadRequest("password must be at least 8 characters")
	}

	var (
		user          *models.User
		assignedRoles []string
	)
	err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context, _ *gorm.DB) error {
		existing, err := s.deps.Users.FindByUsername(txCtx, username)
		if err != nil {
			return fmt.Errorf("find existing user: %w", err)
		}
		if existing != nil {
			return apperrors.BadRequest("username already exists")
		}

		count, err := s.deps.Users.Count(txCtx)
		if err != nil {
			return fmt.Errorf("count users: %w", err)
		}

		id, err := s.deps.IDProvider.NextID()
		if err != nil {
			return fmt.Errorf("generate user id: %w", err)
		}
		passwordHash, err := s.deps.Password.HashPassword(strings.TrimSpace(req.Password))
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		displayName := strings.TrimSpace(req.DisplayName)
		if displayName == "" {
			displayName = username
		}

		user = &models.User{
			BaseModel: models.BaseModel{
				ID: strconv.FormatInt(id, 10),
			},
			Username:     username,
			Email:        normalizeEmail(req.Email),
			DisplayName:  displayName,
			PasswordHash: passwordHash,
			Status:       "active",
		}
		if err := s.deps.Users.Create(txCtx, user); err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		assignedRoles = []string{service.DefaultRoleUser}
		if count == 0 {
			assignedRoles = append(assignedRoles, service.DefaultRoleAdmin)
		}
		for _, roleName := range assignedRoles {
			if err := s.ensureRole(txCtx, roleName); err != nil {
				return err
			}
			roleBindingID, err := s.deps.IDProvider.NextID()
			if err != nil {
				return fmt.Errorf("generate user role id: %w", err)
			}
			if err := s.deps.UserRoles.Assign(txCtx, &models.UserRole{
				BaseModel: models.BaseModel{
					ID: strconv.FormatInt(roleBindingID, 10),
				},
				UserID:   user.ID,
				RoleName: roleName,
			}); err != nil {
				return fmt.Errorf("assign role in store: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return typesuser.AuthResponse{}, err
	}

	for _, roleName := range assignedRoles {
		if err := s.deps.RoleManager.AssignRole(user.ID, roleName); err != nil {
			return typesuser.AuthResponse{}, fmt.Errorf("assign role in rbac: %w", err)
		}
	}

	return s.issueAuthResponse(ctx, user)
}

// Login 校验用户名密码并签发新令牌。
func (s *Service) Login(ctx context.Context, req typesuser.LoginRequest) (typesuser.AuthResponse, error) {
	user, err := s.deps.Users.FindByUsername(ctx, req.Username)
	if err != nil {
		return typesuser.AuthResponse{}, fmt.Errorf("find user by username: %w", err)
	}
	if user == nil {
		return typesuser.AuthResponse{}, apperrors.Unauthorized("invalid username or password")
	}
	if err := s.deps.Password.ComparePassword(user.PasswordHash, strings.TrimSpace(req.Password)); err != nil {
		return typesuser.AuthResponse{}, apperrors.Unauthorized("invalid username or password")
	}

	return s.issueAuthResponse(ctx, user)
}

// Logout 清除用户可刷新的会话状态。
func (s *Service) Logout(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return apperrors.Unauthorized("missing user identity")
	}
	if err := s.deps.Cache.Delete(ctx, refreshTokenCacheKey(userID)); err != nil {
		return fmt.Errorf("delete refresh token from cache: %w", err)
	}
	return nil
}

// ChangePassword 修改当前用户密码。
func (s *Service) ChangePassword(ctx context.Context, userID string, req typesuser.ChangePasswordRequest) error {
	if strings.TrimSpace(userID) == "" {
		return apperrors.Unauthorized("missing user identity")
	}
	if len(strings.TrimSpace(req.NewPassword)) < 8 {
		return apperrors.BadRequest("new password must be at least 8 characters")
	}

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context, _ *gorm.DB) error {
		user, err := s.deps.Users.FindByID(txCtx, userID)
		if err != nil {
			return fmt.Errorf("find user by id: %w", err)
		}
		if user == nil {
			return apperrors.NotFound("user not found")
		}
		if err := s.deps.Password.ComparePassword(user.PasswordHash, strings.TrimSpace(req.OldPassword)); err != nil {
			return apperrors.Unauthorized("old password is incorrect")
		}

		passwordHash, err := s.deps.Password.HashPassword(strings.TrimSpace(req.NewPassword))
		if err != nil {
			return fmt.Errorf("hash new password: %w", err)
		}
		user.PasswordHash = passwordHash
		if err := s.deps.Users.Save(txCtx, user); err != nil {
			return fmt.Errorf("save updated user password: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := s.deps.Cache.Delete(ctx, refreshTokenCacheKey(userID)); err != nil {
		return fmt.Errorf("delete refresh token after password change: %w", err)
	}

	return nil
}

// RefreshToken 刷新访问令牌和刷新令牌。
func (s *Service) RefreshToken(ctx context.Context, req typesuser.RefreshTokenRequest) (typesuser.AuthResponse, error) {
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		return typesuser.AuthResponse{}, apperrors.BadRequest("refresh token is required")
	}

	claims, err := s.deps.Tokens.ValidateToken(refreshToken)
	if err != nil {
		return typesuser.AuthResponse{}, apperrors.Unauthorized("invalid refresh token")
	}
	if claims.TokenType != string(pkgjwt.TokenTypeRefresh) {
		return typesuser.AuthResponse{}, apperrors.Unauthorized("token is not a refresh token")
	}

	userID := claims.Subject
	cachedToken, ok := s.deps.Cache.Get(ctx, refreshTokenCacheKey(userID))
	if !ok || fmt.Sprint(cachedToken) != refreshToken {
		return typesuser.AuthResponse{}, apperrors.Unauthorized("refresh token has expired")
	}

	user, err := s.deps.Users.FindByID(ctx, userID)
	if err != nil {
		return typesuser.AuthResponse{}, fmt.Errorf("find user by id: %w", err)
	}
	if user == nil {
		return typesuser.AuthResponse{}, apperrors.NotFound("user not found")
	}

	return s.issueAuthResponse(ctx, user)
}

func (s *Service) ensureRole(ctx context.Context, roleName string) error {
	id, err := s.deps.IDProvider.NextID()
	if err != nil {
		return fmt.Errorf("generate role id: %w", err)
	}
	return s.deps.Roles.Ensure(ctx, &models.Role{
		BaseModel: models.BaseModel{
			ID: strconv.FormatInt(id, 10),
		},
		Name:        roleName,
		Description: roleDescription(roleName),
	})
}

func (s *Service) issueAuthResponse(ctx context.Context, user *models.User) (typesuser.AuthResponse, error) {
	roles, err := s.deps.UserRoles.ListRolesByUser(ctx, user.ID)
	if err != nil {
		return typesuser.AuthResponse{}, fmt.Errorf("list user roles: %w", err)
	}

	extra := map[string]any{
		"username": user.Username,
		"roles":    append([]string(nil), roles...),
	}
	accessToken, err := s.deps.Tokens.GenerateToken(user.ID, pkgjwt.TokenTypeAccess, extra)
	if err != nil {
		return typesuser.AuthResponse{}, fmt.Errorf("generate access token: %w", err)
	}
	refreshToken, err := s.deps.Tokens.GenerateToken(user.ID, pkgjwt.TokenTypeRefresh, extra)
	if err != nil {
		return typesuser.AuthResponse{}, fmt.Errorf("generate refresh token: %w", err)
	}
	if err := s.deps.Cache.Set(ctx, refreshTokenCacheKey(user.ID), refreshToken, s.deps.RefreshTokenTTL); err != nil {
		return typesuser.AuthResponse{}, fmt.Errorf("cache refresh token: %w", err)
	}

	return typesuser.AuthResponse{
		User: toProfile(user, roles),
		Tokens: typesuser.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
		},
	}, nil
}

func toProfile(user *models.User, roles []string) typesuser.Profile {
	return typesuser.Profile{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Roles:       append([]string(nil), roles...),
		CreatedAt:   user.CreatedAt.UTC().Unix(),
		UpdatedAt:   user.UpdatedAt.UTC().Unix(),
	}
}

func refreshTokenCacheKey(userID string) string {
	return service.RefreshTokenCacheKey(userID)
}

func normalizeUsername(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func normalizeEmail(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func roleDescription(roleName string) string {
	switch roleName {
	case service.DefaultRoleAdmin:
		return "system administrator"
	case service.DefaultRoleUser:
		return "registered user"
	default:
		return "custom role"
	}
}
