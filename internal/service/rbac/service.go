package rbac

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/rin721/rei/internal/models"
	"github.com/rin721/rei/internal/repository"
	"github.com/rin721/rei/internal/service"
	"github.com/rin721/rei/types"
	apperrors "github.com/rin721/rei/types/errors"
	"gorm.io/gorm"
)

// Dependencies 描述 RBAC 服务依赖。
type Dependencies struct {
	Users       repository.UserRepository
	Roles       repository.RoleRepository
	UserRoles   repository.UserRoleRepository
	Policies    repository.PolicyRepository
	IDProvider  service.IDProvider
	Tx          service.TxManager
	RoleManager service.RoleManager
}

// Service 实现 RBAC 业务逻辑。
type Service struct {
	deps Dependencies
}

// New 创建 RBAC 服务。
func New(deps Dependencies) (*Service, error) {
	switch {
	case deps.Users == nil:
		return nil, fmt.Errorf("users repository is required")
	case deps.Roles == nil:
		return nil, fmt.Errorf("roles repository is required")
	case deps.UserRoles == nil:
		return nil, fmt.Errorf("user roles repository is required")
	case deps.Policies == nil:
		return nil, fmt.Errorf("policies repository is required")
	case deps.IDProvider == nil:
		return nil, fmt.Errorf("id provider is required")
	case deps.Tx == nil:
		return nil, fmt.Errorf("tx manager is required")
	case deps.RoleManager == nil:
		return nil, fmt.Errorf("role manager is required")
	}

	return &Service{deps: deps}, nil
}

// LoadFromStore 将数据库中的角色关系和策略同步到内存 RBAC 管理器。
func (s *Service) LoadFromStore(ctx context.Context) error {
	policies, err := s.deps.Policies.List(ctx)
	if err != nil {
		return fmt.Errorf("list policies from store: %w", err)
	}
	for _, policy := range policies {
		if err := s.deps.RoleManager.AddPolicy(policy.Subject, policy.Object, policy.Action); err != nil {
			return fmt.Errorf("load policy into rbac manager: %w", err)
		}
	}

	bindings, err := s.deps.UserRoles.List(ctx)
	if err != nil {
		return fmt.Errorf("list user roles from store: %w", err)
	}
	for _, binding := range bindings {
		if err := s.deps.RoleManager.AssignRole(binding.UserID, binding.RoleName); err != nil {
			return fmt.Errorf("load role binding into rbac manager: %w", err)
		}
	}

	return nil
}

// CheckPermission 检查指定主体是否具备权限。
func (s *Service) CheckPermission(ctx context.Context, req types.CheckPermissionRequest) (types.CheckPermissionResponse, error) {
	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		return types.CheckPermissionResponse{}, apperrors.BadRequest("userId is required")
	}

	object := strings.TrimSpace(req.Object)
	if object == "" {
		return types.CheckPermissionResponse{}, apperrors.BadRequest("object is required")
	}

	action := strings.TrimSpace(strings.ToLower(req.Action))
	if action == "" {
		return types.CheckPermissionResponse{}, apperrors.BadRequest("action is required")
	}

	allowed, err := s.deps.RoleManager.CheckPermission(userID, object, action)
	if err != nil {
		return types.CheckPermissionResponse{}, fmt.Errorf("check permission via rbac manager: %w", err)
	}

	return types.CheckPermissionResponse{
		Allowed: allowed,
		UserID:  userID,
		Object:  object,
		Action:  action,
	}, nil
}

// AssignRole 为指定用户分配角色。
func (s *Service) AssignRole(ctx context.Context, req types.AssignRoleRequest) error {
	userID := strings.TrimSpace(req.UserID)
	roleName := normalizeRole(req.Role)
	if userID == "" || roleName == "" {
		return apperrors.BadRequest("userId and role are required")
	}

	user, err := s.deps.Users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user by id: %w", err)
	}
	if user == nil {
		return apperrors.NotFound("user not found")
	}

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context, _ *gorm.DB) error {
		if err := s.ensureRole(txCtx, roleName); err != nil {
			return err
		}
		id, err := s.deps.IDProvider.NextID()
		if err != nil {
			return fmt.Errorf("generate user role id: %w", err)
		}
		if err := s.deps.UserRoles.Assign(txCtx, &models.UserRole{
			BaseModel: models.BaseModel{
				ID: strconv.FormatInt(id, 10),
			},
			UserID:   userID,
			RoleName: roleName,
		}); err != nil {
			return fmt.Errorf("assign role in store: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := s.deps.RoleManager.AssignRole(userID, roleName); err != nil {
		return fmt.Errorf("assign role in rbac manager: %w", err)
	}
	return nil
}

// RevokeRole 撤销指定用户角色。
func (s *Service) RevokeRole(ctx context.Context, req types.RevokeRoleRequest) error {
	userID := strings.TrimSpace(req.UserID)
	roleName := normalizeRole(req.Role)
	if userID == "" || roleName == "" {
		return apperrors.BadRequest("userId and role are required")
	}

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context, _ *gorm.DB) error {
		if err := s.deps.UserRoles.Revoke(txCtx, userID, roleName); err != nil {
			return fmt.Errorf("revoke role in store: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := s.deps.RoleManager.RevokeRole(userID, roleName); err != nil {
		return fmt.Errorf("revoke role in rbac manager: %w", err)
	}
	return nil
}

// GetUserRoles 返回用户角色列表。
func (s *Service) GetUserRoles(ctx context.Context, userID string) (types.UserRolesResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return types.UserRolesResponse{}, apperrors.BadRequest("userId is required")
	}

	roles, err := s.deps.UserRoles.ListRolesByUser(ctx, userID)
	if err != nil {
		return types.UserRolesResponse{}, fmt.Errorf("list user roles from store: %w", err)
	}
	return types.UserRolesResponse{
		UserID: userID,
		Roles:  roles,
	}, nil
}

// GetUsersForRole 返回角色下的用户列表。
func (s *Service) GetUsersForRole(ctx context.Context, role string) (types.RoleUsersResponse, error) {
	role = normalizeRole(role)
	if role == "" {
		return types.RoleUsersResponse{}, apperrors.BadRequest("role is required")
	}

	userIDs, err := s.deps.UserRoles.ListUsersByRole(ctx, role)
	if err != nil {
		return types.RoleUsersResponse{}, fmt.Errorf("list users by role from store: %w", err)
	}
	return types.RoleUsersResponse{
		Role:    role,
		UserIDs: userIDs,
	}, nil
}

// AddPolicy 添加一条策略。
func (s *Service) AddPolicy(ctx context.Context, req types.PolicyRequest) error {
	subject, object, action, err := validatePolicyRequest(req)
	if err != nil {
		return err
	}

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context, _ *gorm.DB) error {
		id, genErr := s.deps.IDProvider.NextID()
		if genErr != nil {
			return fmt.Errorf("generate policy id: %w", genErr)
		}
		if err := s.deps.Policies.Add(txCtx, &models.Policy{
			BaseModel: models.BaseModel{
				ID: strconv.FormatInt(id, 10),
			},
			Subject: subject,
			Object:  object,
			Action:  action,
		}); err != nil {
			return fmt.Errorf("add policy in store: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := s.deps.RoleManager.AddPolicy(subject, object, action); err != nil {
		return fmt.Errorf("add policy in rbac manager: %w", err)
	}
	return nil
}

// RemovePolicy 删除一条策略。
func (s *Service) RemovePolicy(ctx context.Context, req types.PolicyRequest) error {
	subject, object, action, err := validatePolicyRequest(req)
	if err != nil {
		return err
	}

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context, _ *gorm.DB) error {
		if err := s.deps.Policies.Remove(txCtx, subject, object, action); err != nil {
			return fmt.Errorf("remove policy in store: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := s.deps.RoleManager.RemovePolicy(subject, object, action); err != nil {
		return fmt.Errorf("remove policy in rbac manager: %w", err)
	}
	return nil
}

// ListPolicies 返回当前策略列表。
func (s *Service) ListPolicies(ctx context.Context) (types.PoliciesResponse, error) {
	rawPolicies, err := s.deps.Policies.List(ctx)
	if err != nil {
		return types.PoliciesResponse{}, fmt.Errorf("list policies from store: %w", err)
	}

	items := make([]types.PolicyRequest, 0, len(rawPolicies))
	for _, policy := range rawPolicies {
		items = append(items, types.PolicyRequest{
			Subject: policy.Subject,
			Object:  policy.Object,
			Action:  policy.Action,
		})
	}

	return types.PoliciesResponse{Items: items}, nil
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

func validatePolicyRequest(req types.PolicyRequest) (string, string, string, error) {
	subject := normalizeRole(req.Subject)
	object := strings.TrimSpace(req.Object)
	action := strings.TrimSpace(strings.ToLower(req.Action))
	if subject == "" || object == "" || action == "" {
		return "", "", "", apperrors.BadRequest("subject, object and action are required")
	}
	return subject, object, action, nil
}

func normalizeRole(role string) string {
	return strings.TrimSpace(strings.ToLower(role))
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
