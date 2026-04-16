package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/rin721/go-scaffold2/internal/models"
	"github.com/rin721/go-scaffold2/internal/repository"
	apperrors "github.com/rin721/go-scaffold2/types/errors"
	typesuser "github.com/rin721/go-scaffold2/types/user"
)

// Dependencies 描述用户服务依赖。
type Dependencies struct {
	Users     repository.UserRepository
	UserRoles repository.UserRoleRepository
}

// Service 实现用户资料查询与更新。
type Service struct {
	deps Dependencies
}

// New 创建用户服务。
func New(deps Dependencies) (*Service, error) {
	if deps.Users == nil {
		return nil, fmt.Errorf("users repository is required")
	}
	if deps.UserRoles == nil {
		return nil, fmt.Errorf("user roles repository is required")
	}

	return &Service{deps: deps}, nil
}

// GetProfile 返回当前用户资料。
func (s *Service) GetProfile(ctx context.Context, userID string) (typesuser.Profile, error) {
	user, err := s.deps.Users.FindByID(ctx, userID)
	if err != nil {
		return typesuser.Profile{}, fmt.Errorf("find user by id: %w", err)
	}
	if user == nil {
		return typesuser.Profile{}, apperrors.NotFound("user not found")
	}

	return s.buildProfile(ctx, user)
}

// UpdateProfile 更新当前用户资料。
func (s *Service) UpdateProfile(ctx context.Context, userID string, req typesuser.UpdateProfileRequest) (typesuser.Profile, error) {
	user, err := s.deps.Users.FindByID(ctx, userID)
	if err != nil {
		return typesuser.Profile{}, fmt.Errorf("find user by id: %w", err)
	}
	if user == nil {
		return typesuser.Profile{}, apperrors.NotFound("user not found")
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName != "" {
		user.DisplayName = displayName
	}
	user.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if err := s.deps.Users.Save(ctx, user); err != nil {
		return typesuser.Profile{}, fmt.Errorf("save user profile: %w", err)
	}

	return s.buildProfile(ctx, user)
}

func (s *Service) buildProfile(ctx context.Context, user *models.User) (typesuser.Profile, error) {
	roles, err := s.deps.UserRoles.ListRolesByUser(ctx, user.ID)
	if err != nil {
		return typesuser.Profile{}, fmt.Errorf("list user roles: %w", err)
	}

	return typesuser.Profile{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Roles:       append([]string(nil), roles...),
		CreatedAt:   user.CreatedAt.UTC().Unix(),
		UpdatedAt:   user.UpdatedAt.UTC().Unix(),
	}, nil
}
