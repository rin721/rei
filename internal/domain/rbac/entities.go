package rbac

// Role is the pure domain entity for an RBAC role.
type Role struct {
	ID          string
	Name        string
	Description string
}

// RoleBinding is the pure domain entity for a user-to-role assignment.
type RoleBinding struct {
	ID       string
	UserID   string
	RoleName string
}

// Policy is the pure domain entity for an RBAC policy rule.
type Policy struct {
	ID      string
	Subject string
	Object  string
	Action  string
}
