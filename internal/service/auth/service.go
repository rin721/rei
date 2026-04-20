package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	domainuser "github.com/rin721/rei/internal/domain/user"
	apperrors "github.com/rin721/rei/types/errors"
)

const (
	defaultRoleAdmin      = "admin"
	defaultRoleUser       = "user"
	defaultBearerToken    = "Bearer"
	refreshTokenKeyPrefix = "auth:refresh:"
)

// Dependencies describes the ports required by the auth usecase.
type Dependencies struct {
	Users           UserStore
	Roles           RoleStore
	RoleBindings    RoleBindingStore
	IDProvider      IDProvider
	Password        PasswordManager
	Tokens          TokenManager
	RefreshTokens   RefreshTokenStore
	Tx              TransactionManager
	RoleManager     RoleManager
	RefreshTokenTTL time.Duration
}

// Service implements authentication application logic.
type Service struct {
	deps Dependencies
}

// New creates the auth usecase.
func New(deps Dependencies) (*Service, error) {
	switch {
	case deps.Users == nil:
		return nil, fmt.Errorf("user store is required")
	case deps.Roles == nil:
		return nil, fmt.Errorf("role store is required")
	case deps.RoleBindings == nil:
		return nil, fmt.Errorf("role binding store is required")
	case deps.IDProvider == nil:
		return nil, fmt.Errorf("id provider is required")
	case deps.Password == nil:
		return nil, fmt.Errorf("password manager is required")
	case deps.Tokens == nil:
		return nil, fmt.Errorf("token manager is required")
	case deps.RefreshTokens == nil:
		return nil, fmt.Errorf("refresh token store is required")
	case deps.Tx == nil:
		return nil, fmt.Errorf("transaction manager is required")
	case deps.RoleManager == nil:
		return nil, fmt.Errorf("role manager is required")
	}

	if deps.RefreshTokenTTL <= 0 {
		deps.RefreshTokenTTL = 72 * time.Hour
	}

	return &Service{deps: deps}, nil
}

// Register creates a user and returns authentication data.
func (s *Service) Register(ctx context.Context, cmd RegisterCommand) (Authentication, error) {
	username := normalizeUsername(cmd.Username)
	if username == "" {
		return Authentication{}, apperrors.BadRequest("username is required")
	}
	if len(strings.TrimSpace(cmd.Password)) < 8 {
		return Authentication{}, apperrors.BadRequest("password must be at least 8 characters")
	}

	var (
		user          *domainuser.User
		assignedRoles []string
	)
	err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context) error {
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
		passwordHash, err := s.deps.Password.HashPassword(strings.TrimSpace(cmd.Password))
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		displayName := strings.TrimSpace(cmd.DisplayName)
		if displayName == "" {
			displayName = username
		}

		user = &domainuser.User{
			ID:           strconv.FormatInt(id, 10),
			Username:     username,
			Email:        normalizeEmail(cmd.Email),
			DisplayName:  displayName,
			PasswordHash: passwordHash,
			Status:       "active",
		}
		if err := s.deps.Users.Create(txCtx, user); err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		assignedRoles = []string{defaultRoleUser}
		if count == 0 {
			assignedRoles = append(assignedRoles, defaultRoleAdmin)
		}
		for _, roleName := range assignedRoles {
			if err := s.ensureRole(txCtx, roleName); err != nil {
				return err
			}
			roleBindingID, err := s.deps.IDProvider.NextID()
			if err != nil {
				return fmt.Errorf("generate user role id: %w", err)
			}
			if err := s.deps.RoleBindings.Assign(txCtx, RoleBinding{
				ID:       strconv.FormatInt(roleBindingID, 10),
				UserID:   user.ID,
				RoleName: roleName,
			}); err != nil {
				return fmt.Errorf("assign role in store: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return Authentication{}, err
	}

	for _, roleName := range assignedRoles {
		if err := s.deps.RoleManager.AssignRole(user.ID, roleName); err != nil {
			return Authentication{}, fmt.Errorf("assign role in rbac: %w", err)
		}
	}

	return s.issueAuthentication(ctx, user)
}

// Login verifies credentials and issues a fresh token pair.
func (s *Service) Login(ctx context.Context, cmd LoginCommand) (Authentication, error) {
	user, err := s.deps.Users.FindByUsername(ctx, cmd.Username)
	if err != nil {
		return Authentication{}, fmt.Errorf("find user by username: %w", err)
	}
	if user == nil {
		return Authentication{}, apperrors.Unauthorized("invalid username or password")
	}
	if err := s.deps.Password.ComparePassword(user.PasswordHash, strings.TrimSpace(cmd.Password)); err != nil {
		return Authentication{}, apperrors.Unauthorized("invalid username or password")
	}

	return s.issueAuthentication(ctx, user)
}

// Logout clears the refresh-token session state for the current user.
func (s *Service) Logout(ctx context.Context, cmd LogoutCommand) error {
	userID := strings.TrimSpace(cmd.UserID)
	if userID == "" {
		return apperrors.Unauthorized("missing user identity")
	}
	if err := s.deps.RefreshTokens.DeleteRefreshToken(ctx, userID); err != nil {
		return fmt.Errorf("delete refresh token from store: %w", err)
	}
	return nil
}

// ChangePassword updates the current user's password.
func (s *Service) ChangePassword(ctx context.Context, cmd ChangePasswordCommand) error {
	userID := strings.TrimSpace(cmd.UserID)
	if userID == "" {
		return apperrors.Unauthorized("missing user identity")
	}
	if len(strings.TrimSpace(cmd.NewPassword)) < 8 {
		return apperrors.BadRequest("new password must be at least 8 characters")
	}

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context) error {
		user, err := s.deps.Users.FindByID(txCtx, userID)
		if err != nil {
			return fmt.Errorf("find user by id: %w", err)
		}
		if user == nil {
			return apperrors.NotFound("user not found")
		}
		if err := s.deps.Password.ComparePassword(user.PasswordHash, strings.TrimSpace(cmd.OldPassword)); err != nil {
			return apperrors.Unauthorized("old password is incorrect")
		}

		passwordHash, err := s.deps.Password.HashPassword(strings.TrimSpace(cmd.NewPassword))
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

	if err := s.deps.RefreshTokens.DeleteRefreshToken(ctx, userID); err != nil {
		return fmt.Errorf("delete refresh token after password change: %w", err)
	}

	return nil
}

// RefreshToken validates the refresh token and issues a new token pair.
func (s *Service) RefreshToken(ctx context.Context, cmd RefreshTokenCommand) (Authentication, error) {
	refreshToken := strings.TrimSpace(cmd.RefreshToken)
	if refreshToken == "" {
		return Authentication{}, apperrors.BadRequest("refresh token is required")
	}

	userID, err := s.deps.Tokens.ValidateRefreshToken(refreshToken)
	if err != nil {
		return Authentication{}, apperrors.Unauthorized("invalid refresh token")
	}

	cachedToken, ok, err := s.deps.RefreshTokens.GetRefreshToken(ctx, userID)
	if err != nil {
		return Authentication{}, fmt.Errorf("get refresh token from store: %w", err)
	}
	if !ok || cachedToken != refreshToken {
		return Authentication{}, apperrors.Unauthorized("refresh token has expired")
	}

	user, err := s.deps.Users.FindByID(ctx, userID)
	if err != nil {
		return Authentication{}, fmt.Errorf("find user by id: %w", err)
	}
	if user == nil {
		return Authentication{}, apperrors.NotFound("user not found")
	}

	return s.issueAuthentication(ctx, user)
}

func (s *Service) ensureRole(ctx context.Context, roleName string) error {
	id, err := s.deps.IDProvider.NextID()
	if err != nil {
		return fmt.Errorf("generate role id: %w", err)
	}
	return s.deps.Roles.Ensure(ctx, Role{
		ID:          strconv.FormatInt(id, 10),
		Name:        roleName,
		Description: roleDescription(roleName),
	})
}

func (s *Service) issueAuthentication(ctx context.Context, user *domainuser.User) (Authentication, error) {
	roles, err := s.deps.RoleBindings.ListRolesByUser(ctx, user.ID)
	if err != nil {
		return Authentication{}, fmt.Errorf("list user roles: %w", err)
	}

	extra := map[string]any{
		"username": user.Username,
		"roles":    append([]string(nil), roles...),
	}
	accessToken, err := s.deps.Tokens.GenerateAccessToken(user.ID, extra)
	if err != nil {
		return Authentication{}, fmt.Errorf("generate access token: %w", err)
	}
	refreshToken, err := s.deps.Tokens.GenerateRefreshToken(user.ID, extra)
	if err != nil {
		return Authentication{}, fmt.Errorf("generate refresh token: %w", err)
	}
	if err := s.deps.RefreshTokens.SetRefreshToken(ctx, user.ID, refreshToken, s.deps.RefreshTokenTTL); err != nil {
		return Authentication{}, fmt.Errorf("store refresh token: %w", err)
	}

	return Authentication{
		User: AuthenticatedUser{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			Roles:       append([]string(nil), roles...),
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
		Tokens: TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    defaultBearerToken,
		},
	}, nil
}

func normalizeUsername(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func normalizeEmail(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func roleDescription(roleName string) string {
	switch roleName {
	case defaultRoleAdmin:
		return "system administrator"
	case defaultRoleUser:
		return "registered user"
	default:
		return "custom role"
	}
}
