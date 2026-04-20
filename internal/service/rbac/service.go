package rbac

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	domainrbac "github.com/rin721/rei/internal/domain/rbac"
	apperrors "github.com/rin721/rei/types/errors"
)

const (
	defaultRoleAdmin = "admin"
	defaultRoleUser  = "user"
)

// Dependencies describes the ports required by the RBAC usecase.
type Dependencies struct {
	Users      UserLookup
	Roles      RoleStore
	RoleBinds  RoleBindingStore
	Policies   PolicyStore
	IDProvider IDProvider
	Tx         TransactionManager
	Enforcer   Enforcer
}

// Service implements RBAC application logic.
type Service struct {
	deps Dependencies
}

// New creates the RBAC usecase.
func New(deps Dependencies) (*Service, error) {
	switch {
	case deps.Users == nil:
		return nil, fmt.Errorf("user lookup is required")
	case deps.Roles == nil:
		return nil, fmt.Errorf("role store is required")
	case deps.RoleBinds == nil:
		return nil, fmt.Errorf("role binding store is required")
	case deps.Policies == nil:
		return nil, fmt.Errorf("policy store is required")
	case deps.IDProvider == nil:
		return nil, fmt.Errorf("id provider is required")
	case deps.Tx == nil:
		return nil, fmt.Errorf("transaction manager is required")
	case deps.Enforcer == nil:
		return nil, fmt.Errorf("enforcer is required")
	}

	return &Service{deps: deps}, nil
}

// LoadFromStore synchronizes persisted RBAC state into the runtime enforcer.
func (s *Service) LoadFromStore(ctx context.Context) error {
	policies, err := s.deps.Policies.List(ctx)
	if err != nil {
		return fmt.Errorf("list policies from store: %w", err)
	}
	for _, policy := range policies {
		if err := s.deps.Enforcer.AddPolicy(policy.Subject, policy.Object, policy.Action); err != nil {
			return fmt.Errorf("load policy into enforcer: %w", err)
		}
	}

	bindings, err := s.deps.RoleBinds.List(ctx)
	if err != nil {
		return fmt.Errorf("list role bindings from store: %w", err)
	}
	for _, binding := range bindings {
		if err := s.deps.Enforcer.AssignRole(binding.UserID, binding.RoleName); err != nil {
			return fmt.Errorf("load role binding into enforcer: %w", err)
		}
	}

	return nil
}

// CheckPermission checks whether a subject is allowed to perform an action on an object.
func (s *Service) CheckPermission(ctx context.Context, query CheckPermissionQuery) (CheckPermissionResult, error) {
	userID := strings.TrimSpace(query.UserID)
	if userID == "" {
		return CheckPermissionResult{}, apperrors.BadRequest("userId is required")
	}

	object := strings.TrimSpace(query.Object)
	if object == "" {
		return CheckPermissionResult{}, apperrors.BadRequest("object is required")
	}

	action := strings.TrimSpace(strings.ToLower(query.Action))
	if action == "" {
		return CheckPermissionResult{}, apperrors.BadRequest("action is required")
	}

	allowed, err := s.deps.Enforcer.CheckPermission(userID, object, action)
	if err != nil {
		return CheckPermissionResult{}, fmt.Errorf("check permission via enforcer: %w", err)
	}

	return CheckPermissionResult{
		Allowed: allowed,
		UserID:  userID,
		Object:  object,
		Action:  action,
	}, nil
}

// AssignRole assigns a role to a user.
func (s *Service) AssignRole(ctx context.Context, cmd AssignRoleCommand) error {
	userID := strings.TrimSpace(cmd.UserID)
	roleName := normalizeRole(cmd.Role)
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

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.ensureRole(txCtx, roleName); err != nil {
			return err
		}
		id, err := s.deps.IDProvider.NextID()
		if err != nil {
			return fmt.Errorf("generate user role id: %w", err)
		}
		if err := s.deps.RoleBinds.Assign(txCtx, domainrbac.RoleBinding{
			ID:       strconv.FormatInt(id, 10),
			UserID:   userID,
			RoleName: roleName,
		}); err != nil {
			return fmt.Errorf("assign role in store: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := s.deps.Enforcer.AssignRole(userID, roleName); err != nil {
		return fmt.Errorf("assign role in enforcer: %w", err)
	}
	return nil
}

// RevokeRole revokes a role from a user.
func (s *Service) RevokeRole(ctx context.Context, cmd RevokeRoleCommand) error {
	userID := strings.TrimSpace(cmd.UserID)
	roleName := normalizeRole(cmd.Role)
	if userID == "" || roleName == "" {
		return apperrors.BadRequest("userId and role are required")
	}

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.deps.RoleBinds.Revoke(txCtx, userID, roleName); err != nil {
			return fmt.Errorf("revoke role in store: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := s.deps.Enforcer.RevokeRole(userID, roleName); err != nil {
		return fmt.Errorf("revoke role in enforcer: %w", err)
	}
	return nil
}

// GetUserRoles returns the roles assigned to a user.
func (s *Service) GetUserRoles(ctx context.Context, query GetUserRolesQuery) (UserRoles, error) {
	userID := strings.TrimSpace(query.UserID)
	if userID == "" {
		return UserRoles{}, apperrors.BadRequest("userId is required")
	}

	roles, err := s.deps.RoleBinds.ListRolesByUser(ctx, userID)
	if err != nil {
		return UserRoles{}, fmt.Errorf("list user roles from store: %w", err)
	}

	return UserRoles{
		UserID: userID,
		Roles:  roles,
	}, nil
}

// GetUsersForRole returns the users assigned to a role.
func (s *Service) GetUsersForRole(ctx context.Context, query GetUsersForRoleQuery) (RoleUsers, error) {
	role := normalizeRole(query.Role)
	if role == "" {
		return RoleUsers{}, apperrors.BadRequest("role is required")
	}

	userIDs, err := s.deps.RoleBinds.ListUsersByRole(ctx, role)
	if err != nil {
		return RoleUsers{}, fmt.Errorf("list users by role from store: %w", err)
	}

	return RoleUsers{
		Role:    role,
		UserIDs: userIDs,
	}, nil
}

// AddPolicy adds a policy rule.
func (s *Service) AddPolicy(ctx context.Context, cmd PolicyCommand) error {
	subject, object, action, err := validatePolicyCommand(cmd)
	if err != nil {
		return err
	}

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context) error {
		id, genErr := s.deps.IDProvider.NextID()
		if genErr != nil {
			return fmt.Errorf("generate policy id: %w", genErr)
		}
		if err := s.deps.Policies.Add(txCtx, domainrbac.Policy{
			ID:      strconv.FormatInt(id, 10),
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

	if err := s.deps.Enforcer.AddPolicy(subject, object, action); err != nil {
		return fmt.Errorf("add policy in enforcer: %w", err)
	}
	return nil
}

// RemovePolicy removes a policy rule.
func (s *Service) RemovePolicy(ctx context.Context, cmd PolicyCommand) error {
	subject, object, action, err := validatePolicyCommand(cmd)
	if err != nil {
		return err
	}

	if err := s.deps.Tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.deps.Policies.Remove(txCtx, subject, object, action); err != nil {
			return fmt.Errorf("remove policy in store: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := s.deps.Enforcer.RemovePolicy(subject, object, action); err != nil {
		return fmt.Errorf("remove policy in enforcer: %w", err)
	}
	return nil
}

// ListPolicies returns all current policies.
func (s *Service) ListPolicies(ctx context.Context) (PolicyList, error) {
	rawPolicies, err := s.deps.Policies.List(ctx)
	if err != nil {
		return PolicyList{}, fmt.Errorf("list policies from store: %w", err)
	}

	items := make([]PolicyItem, 0, len(rawPolicies))
	for _, policy := range rawPolicies {
		items = append(items, PolicyItem{
			Subject: policy.Subject,
			Object:  policy.Object,
			Action:  policy.Action,
		})
	}

	return PolicyList{Items: items}, nil
}

func (s *Service) ensureRole(ctx context.Context, roleName string) error {
	id, err := s.deps.IDProvider.NextID()
	if err != nil {
		return fmt.Errorf("generate role id: %w", err)
	}
	return s.deps.Roles.Ensure(ctx, domainrbac.Role{
		ID:          strconv.FormatInt(id, 10),
		Name:        roleName,
		Description: roleDescription(roleName),
	})
}

func validatePolicyCommand(cmd PolicyCommand) (string, string, string, error) {
	subject := normalizeRole(cmd.Subject)
	object := strings.TrimSpace(cmd.Object)
	action := strings.TrimSpace(strings.ToLower(cmd.Action))
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
	case defaultRoleAdmin:
		return "system administrator"
	case defaultRoleUser:
		return "registered user"
	default:
		return "custom role"
	}
}
