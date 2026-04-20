package repository

import (
	"context"
	"fmt"

	domainuser "github.com/rin721/rei/internal/domain/user"
	"github.com/rin721/rei/internal/models"
	userservice "github.com/rin721/rei/internal/service/user"
)

type userDomainStore struct {
	users UserRepository
}

type userRoleBindingReader struct {
	userRoles UserRoleRepository
}

// NewUserDomainStore adapts the existing GORM-backed user repository to the user usecase port.
func NewUserDomainStore(users UserRepository) userservice.UserStore {
	if users == nil {
		return nil
	}
	return &userDomainStore{users: users}
}

// NewUserRoleBindingReader adapts the existing role binding repository to the user usecase port.
func NewUserRoleBindingReader(userRoles UserRoleRepository) userservice.RoleBindingReader {
	if userRoles == nil {
		return nil
	}
	return &userRoleBindingReader{userRoles: userRoles}
}

func (s *userDomainStore) FindByID(ctx context.Context, id string) (*domainuser.User, error) {
	model, err := s.users.FindByID(ctx, id)
	if err != nil || model == nil {
		return nil, err
	}
	return toDomainUser(model), nil
}

func (s *userDomainStore) Save(ctx context.Context, user *domainuser.User) error {
	model := toUserModel(user)
	if model == nil {
		return fmt.Errorf("user entity is required")
	}
	return s.users.Save(ctx, model)
}

func (r *userRoleBindingReader) ListRolesByUser(ctx context.Context, userID string) ([]string, error) {
	return r.userRoles.ListRolesByUser(ctx, userID)
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
