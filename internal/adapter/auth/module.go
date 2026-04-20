package authadapter

import (
	"context"
	"fmt"
	"time"

	domainuser "github.com/rin721/rei/internal/domain/user"
	"github.com/rin721/rei/internal/models"
	"github.com/rin721/rei/internal/repository"
	authservice "github.com/rin721/rei/internal/service/auth"
	pkgjwt "github.com/rin721/rei/pkg/jwt"
	"gorm.io/gorm"
)

const refreshTokenCachePrefix = "auth:refresh:"

type userStore struct {
	users repository.UserRepository
}

type roleStore struct {
	roles repository.RoleRepository
}

type roleBindingStore struct {
	roleBindings repository.UserRoleRepository
}

type tokenManager struct {
	tokens legacyTokenManager
}

type refreshTokenStore struct {
	cache legacyCacheStore
}

type transactionManager struct {
	tx legacyTransactionManager
}

type legacyTokenManager interface {
	GenerateToken(string, pkgjwt.TokenType, map[string]any) (string, error)
	ValidateToken(string) (*pkgjwt.Claims, error)
}

type legacyCacheStore interface {
	Get(context.Context, string) (any, bool)
	Set(context.Context, string, any, time.Duration) error
	Delete(context.Context, string) error
}

type legacyTransactionManager interface {
	WithTx(context.Context, func(context.Context, *gorm.DB) error) error
}

// NewUserStore adapts the existing user repository to the auth usecase port.
func NewUserStore(users repository.UserRepository) authservice.UserStore {
	if users == nil {
		return nil
	}
	return &userStore{users: users}
}

// NewRoleStore adapts the existing role repository to the auth usecase port.
func NewRoleStore(roles repository.RoleRepository) authservice.RoleStore {
	if roles == nil {
		return nil
	}
	return &roleStore{roles: roles}
}

// NewRoleBindingStore adapts the existing user-role repository to the auth usecase port.
func NewRoleBindingStore(roleBindings repository.UserRoleRepository) authservice.RoleBindingStore {
	if roleBindings == nil {
		return nil
	}
	return &roleBindingStore{roleBindings: roleBindings}
}

// NewTokenManager adapts the legacy JWT manager to the auth usecase port.
func NewTokenManager(tokens legacyTokenManager) authservice.TokenManager {
	if tokens == nil {
		return nil
	}
	return &tokenManager{tokens: tokens}
}

// NewRefreshTokenStore adapts the cache store to the auth usecase port.
func NewRefreshTokenStore(cache legacyCacheStore) authservice.RefreshTokenStore {
	if cache == nil {
		return nil
	}
	return &refreshTokenStore{cache: cache}
}

// NewTransactionManager adapts the legacy transaction manager to the auth usecase port.
func NewTransactionManager(tx legacyTransactionManager) authservice.TransactionManager {
	if tx == nil {
		return nil
	}
	return &transactionManager{tx: tx}
}

func (s *userStore) FindByID(ctx context.Context, id string) (*domainuser.User, error) {
	model, err := s.users.FindByID(ctx, id)
	if err != nil || model == nil {
		return nil, err
	}
	return toDomainUser(model), nil
}

func (s *userStore) FindByUsername(ctx context.Context, username string) (*domainuser.User, error) {
	model, err := s.users.FindByUsername(ctx, username)
	if err != nil || model == nil {
		return nil, err
	}
	return toDomainUser(model), nil
}

func (s *userStore) Create(ctx context.Context, user *domainuser.User) error {
	model := toUserModel(user)
	if model == nil {
		return fmt.Errorf("user entity is required")
	}
	return s.users.Create(ctx, model)
}

func (s *userStore) Save(ctx context.Context, user *domainuser.User) error {
	model := toUserModel(user)
	if model == nil {
		return fmt.Errorf("user entity is required")
	}
	return s.users.Save(ctx, model)
}

func (s *userStore) Count(ctx context.Context) (int64, error) {
	return s.users.Count(ctx)
}

func (s *roleStore) Ensure(ctx context.Context, role authservice.Role) error {
	return s.roles.Ensure(ctx, &models.Role{
		BaseModel: models.BaseModel{
			ID: role.ID,
		},
		Name:        role.Name,
		Description: role.Description,
	})
}

func (s *roleBindingStore) Assign(ctx context.Context, binding authservice.RoleBinding) error {
	return s.roleBindings.Assign(ctx, &models.UserRole{
		BaseModel: models.BaseModel{
			ID: binding.ID,
		},
		UserID:   binding.UserID,
		RoleName: binding.RoleName,
	})
}

func (s *roleBindingStore) ListRolesByUser(ctx context.Context, userID string) ([]string, error) {
	return s.roleBindings.ListRolesByUser(ctx, userID)
}

func (m *tokenManager) GenerateAccessToken(subject string, extra map[string]any) (string, error) {
	return m.tokens.GenerateToken(subject, pkgjwt.TokenTypeAccess, extra)
}

func (m *tokenManager) GenerateRefreshToken(subject string, extra map[string]any) (string, error) {
	return m.tokens.GenerateToken(subject, pkgjwt.TokenTypeRefresh, extra)
}

func (m *tokenManager) ValidateRefreshToken(token string) (string, error) {
	claims, err := m.tokens.ValidateToken(token)
	if err != nil {
		return "", err
	}
	if claims.TokenType != string(pkgjwt.TokenTypeRefresh) {
		return "", fmt.Errorf("token is not a refresh token")
	}
	return claims.Subject, nil
}

func (s *refreshTokenStore) GetRefreshToken(ctx context.Context, userID string) (string, bool, error) {
	value, ok := s.cache.Get(ctx, refreshTokenKey(userID))
	if !ok {
		return "", false, nil
	}
	return fmt.Sprint(value), true, nil
}

func (s *refreshTokenStore) SetRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	return s.cache.Set(ctx, refreshTokenKey(userID), token, ttl)
}

func (s *refreshTokenStore) DeleteRefreshToken(ctx context.Context, userID string) error {
	return s.cache.Delete(ctx, refreshTokenKey(userID))
}

func (m *transactionManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return m.tx.WithTx(ctx, func(txCtx context.Context, _ *gorm.DB) error {
		return fn(txCtx)
	})
}

func toDomainUser(model *models.User) *domainuser.User {
	if model == nil {
		return nil
	}
	return &domainuser.User{
		ID:           model.ID,
		Username:     model.Username,
		Email:        model.Email,
		DisplayName:  model.DisplayName,
		PasswordHash: model.PasswordHash,
		Status:       model.Status,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func toUserModel(user *domainuser.User) *models.User {
	if user == nil {
		return nil
	}
	return &models.User{
		BaseModel: models.BaseModel{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Username:     user.Username,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		PasswordHash: user.PasswordHash,
		Status:       user.Status,
	}
}

func refreshTokenKey(userID string) string {
	return refreshTokenCachePrefix + userID
}
