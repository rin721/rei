package user

import (
	"context"
	"fmt"
	"strings"

	domainuser "github.com/rin721/rei/internal/domain/user"
	apperrors "github.com/rin721/rei/types/errors"
)

// Dependencies describes the ports required by the user usecase.
type Dependencies struct {
	Users     UserStore
	UserRoles RoleBindingReader
}

// Service implements user application logic.
type Service struct {
	deps Dependencies
}

// New creates the user usecase.
func New(deps Dependencies) (*Service, error) {
	if deps.Users == nil {
		return nil, fmt.Errorf("user store is required")
	}
	if deps.UserRoles == nil {
		return nil, fmt.Errorf("role binding reader is required")
	}

	return &Service{deps: deps}, nil
}

// GetProfile returns the current user's profile.
func (s *Service) GetProfile(ctx context.Context, query GetProfileQuery) (Profile, error) {
	user, err := s.deps.Users.FindByID(ctx, strings.TrimSpace(query.UserID))
	if err != nil {
		return Profile{}, fmt.Errorf("find user by id: %w", err)
	}
	if user == nil {
		return Profile{}, apperrors.NotFound("user not found")
	}

	return s.buildProfile(ctx, user)
}

// UpdateProfile updates the current user's profile.
func (s *Service) UpdateProfile(ctx context.Context, cmd UpdateProfileCommand) (Profile, error) {
	user, err := s.deps.Users.FindByID(ctx, strings.TrimSpace(cmd.UserID))
	if err != nil {
		return Profile{}, fmt.Errorf("find user by id: %w", err)
	}
	if user == nil {
		return Profile{}, apperrors.NotFound("user not found")
	}

	displayName := strings.TrimSpace(cmd.DisplayName)
	if displayName != "" {
		user.DisplayName = displayName
	}
	user.Email = normalizeEmail(cmd.Email)

	if err := s.deps.Users.Save(ctx, user); err != nil {
		return Profile{}, fmt.Errorf("save user profile: %w", err)
	}

	return s.buildProfile(ctx, user)
}

func (s *Service) buildProfile(ctx context.Context, user *domainuser.User) (Profile, error) {
	roles, err := s.deps.UserRoles.ListRolesByUser(ctx, user.ID)
	if err != nil {
		return Profile{}, fmt.Errorf("list user roles: %w", err)
	}

	return Profile{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Roles:       append([]string(nil), roles...),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

func normalizeEmail(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}
