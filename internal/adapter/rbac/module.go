package rbacadapter

import (
	"context"

	domainrbac "github.com/rin721/rei/internal/domain/rbac"
	domainuser "github.com/rin721/rei/internal/domain/user"
	"github.com/rin721/rei/internal/models"
	"github.com/rin721/rei/internal/repository"
	rbacservice "github.com/rin721/rei/internal/service/rbac"
	"gorm.io/gorm"
)

type userLookup struct {
	users repository.UserRepository
}

type roleStore struct {
	roles repository.RoleRepository
}

type roleBindingStore struct {
	roleBindings repository.UserRoleRepository
}

type policyStore struct {
	policies repository.PolicyRepository
}

type transactionManager struct {
	tx legacyTransactionManager
}

type legacyTransactionManager interface {
	WithTx(context.Context, func(context.Context, *gorm.DB) error) error
}

type legacyEnforcer interface {
	CheckPermission(string, string, string) (bool, error)
	AssignRole(string, string) error
	RevokeRole(string, string) error
	AddPolicy(string, string, string) error
	RemovePolicy(string, string, string) error
}

// NewUserLookup adapts the existing user repository to the RBAC usecase port.
func NewUserLookup(users repository.UserRepository) rbacservice.UserLookup {
	if users == nil {
		return nil
	}
	return &userLookup{users: users}
}

// NewRoleStore adapts the existing role repository to the RBAC usecase port.
func NewRoleStore(roles repository.RoleRepository) rbacservice.RoleStore {
	if roles == nil {
		return nil
	}
	return &roleStore{roles: roles}
}

// NewRoleBindingStore adapts the existing user-role repository to the RBAC usecase port.
func NewRoleBindingStore(roleBindings repository.UserRoleRepository) rbacservice.RoleBindingStore {
	if roleBindings == nil {
		return nil
	}
	return &roleBindingStore{roleBindings: roleBindings}
}

// NewPolicyStore adapts the existing policy repository to the RBAC usecase port.
func NewPolicyStore(policies repository.PolicyRepository) rbacservice.PolicyStore {
	if policies == nil {
		return nil
	}
	return &policyStore{policies: policies}
}

// NewTransactionManager adapts the legacy transaction manager to the RBAC usecase port.
func NewTransactionManager(tx legacyTransactionManager) rbacservice.TransactionManager {
	if tx == nil {
		return nil
	}
	return &transactionManager{tx: tx}
}

// NewEnforcer reuses the existing RBAC manager behind the RBAC usecase port.
func NewEnforcer(enforcer legacyEnforcer) rbacservice.Enforcer {
	if enforcer == nil {
		return nil
	}
	return enforcer
}

func (s *userLookup) FindByID(ctx context.Context, id string) (*domainuser.User, error) {
	model, err := s.users.FindByID(ctx, id)
	if err != nil || model == nil {
		return nil, err
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
	}, nil
}

func (s *roleStore) Ensure(ctx context.Context, role domainrbac.Role) error {
	return s.roles.Ensure(ctx, &models.Role{
		BaseModel: models.BaseModel{
			ID: role.ID,
		},
		Name:        role.Name,
		Description: role.Description,
	})
}

func (s *roleBindingStore) Assign(ctx context.Context, binding domainrbac.RoleBinding) error {
	return s.roleBindings.Assign(ctx, &models.UserRole{
		BaseModel: models.BaseModel{
			ID: binding.ID,
		},
		UserID:   binding.UserID,
		RoleName: binding.RoleName,
	})
}

func (s *roleBindingStore) Revoke(ctx context.Context, userID, roleName string) error {
	return s.roleBindings.Revoke(ctx, userID, roleName)
}

func (s *roleBindingStore) ListRolesByUser(ctx context.Context, userID string) ([]string, error) {
	return s.roleBindings.ListRolesByUser(ctx, userID)
}

func (s *roleBindingStore) ListUsersByRole(ctx context.Context, role string) ([]string, error) {
	return s.roleBindings.ListUsersByRole(ctx, role)
}

func (s *roleBindingStore) List(ctx context.Context) ([]domainrbac.RoleBinding, error) {
	items, err := s.roleBindings.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domainrbac.RoleBinding, 0, len(items))
	for _, item := range items {
		result = append(result, domainrbac.RoleBinding{
			ID:       item.ID,
			UserID:   item.UserID,
			RoleName: item.RoleName,
		})
	}
	return result, nil
}

func (s *policyStore) Add(ctx context.Context, policy domainrbac.Policy) error {
	return s.policies.Add(ctx, &models.Policy{
		BaseModel: models.BaseModel{
			ID: policy.ID,
		},
		Subject: policy.Subject,
		Object:  policy.Object,
		Action:  policy.Action,
	})
}

func (s *policyStore) Remove(ctx context.Context, subject, object, action string) error {
	return s.policies.Remove(ctx, subject, object, action)
}

func (s *policyStore) List(ctx context.Context) ([]domainrbac.Policy, error) {
	items, err := s.policies.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domainrbac.Policy, 0, len(items))
	for _, item := range items {
		result = append(result, domainrbac.Policy{
			ID:      item.ID,
			Subject: item.Subject,
			Object:  item.Object,
			Action:  item.Action,
		})
	}
	return result, nil
}

func (m *transactionManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return m.tx.WithTx(ctx, func(txCtx context.Context, _ *gorm.DB) error {
		return fn(txCtx)
	})
}
