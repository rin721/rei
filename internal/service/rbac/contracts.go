package rbac

import (
	"context"

	domainrbac "github.com/rin721/rei/internal/domain/rbac"
	domainuser "github.com/rin721/rei/internal/domain/user"
)

// UseCase defines the RBAC application contract.
type UseCase interface {
	LoadFromStore(context.Context) error
	CheckPermission(context.Context, CheckPermissionQuery) (CheckPermissionResult, error)
	AssignRole(context.Context, AssignRoleCommand) error
	RevokeRole(context.Context, RevokeRoleCommand) error
	GetUserRoles(context.Context, GetUserRolesQuery) (UserRoles, error)
	GetUsersForRole(context.Context, GetUsersForRoleQuery) (RoleUsers, error)
	AddPolicy(context.Context, PolicyCommand) error
	RemovePolicy(context.Context, PolicyCommand) error
	ListPolicies(context.Context) (PolicyList, error)
}

// CheckPermissionQuery describes the input needed for a permission check.
type CheckPermissionQuery struct {
	UserID string
	Object string
	Action string
}

// CheckPermissionResult is the application-layer result for permission checks.
type CheckPermissionResult struct {
	Allowed bool
	UserID  string
	Object  string
	Action  string
}

// AssignRoleCommand describes the input needed to assign a role.
type AssignRoleCommand struct {
	UserID string
	Role   string
}

// RevokeRoleCommand describes the input needed to revoke a role.
type RevokeRoleCommand struct {
	UserID string
	Role   string
}

// GetUserRolesQuery describes the input needed to fetch a user's roles.
type GetUserRolesQuery struct {
	UserID string
}

// UserRoles is the application-layer result for user role listings.
type UserRoles struct {
	UserID string
	Roles  []string
}

// GetUsersForRoleQuery describes the input needed to fetch users for a role.
type GetUsersForRoleQuery struct {
	Role string
}

// RoleUsers is the application-layer result for role membership listings.
type RoleUsers struct {
	Role    string
	UserIDs []string
}

// PolicyCommand describes the input needed to add or remove a policy.
type PolicyCommand struct {
	Subject string
	Object  string
	Action  string
}

// PolicyItem is the application-layer projection of a policy.
type PolicyItem struct {
	Subject string
	Object  string
	Action  string
}

// PolicyList is the application-layer result for listing policies.
type PolicyList struct {
	Items []PolicyItem
}

// UserLookup defines the user lookup port needed by RBAC flows.
type UserLookup interface {
	FindByID(context.Context, string) (*domainuser.User, error)
}

// RoleStore defines the role persistence port.
type RoleStore interface {
	Ensure(context.Context, domainrbac.Role) error
}

// RoleBindingStore defines the role-binding persistence port.
type RoleBindingStore interface {
	Assign(context.Context, domainrbac.RoleBinding) error
	Revoke(context.Context, string, string) error
	ListRolesByUser(context.Context, string) ([]string, error)
	ListUsersByRole(context.Context, string) ([]string, error)
	List(context.Context) ([]domainrbac.RoleBinding, error)
}

// PolicyStore defines the policy persistence port.
type PolicyStore interface {
	Add(context.Context, domainrbac.Policy) error
	Remove(context.Context, string, string, string) error
	List(context.Context) ([]domainrbac.Policy, error)
}

// IDProvider defines ID generation.
type IDProvider interface {
	NextID() (int64, error)
}

// TransactionManager defines the transaction boundary used by RBAC flows.
type TransactionManager interface {
	WithTx(context.Context, func(context.Context) error) error
}

// Enforcer defines the in-memory RBAC synchronization and evaluation behavior.
type Enforcer interface {
	CheckPermission(string, string, string) (bool, error)
	AssignRole(string, string) error
	RevokeRole(string, string) error
	AddPolicy(string, string, string) error
	RemovePolicy(string, string, string) error
}
