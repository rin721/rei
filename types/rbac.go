package types

// CheckPermissionRequest 定义 RBAC 检查请求。
type CheckPermissionRequest struct {
	UserID string `json:"userId"`
	Object string `json:"object"`
	Action string `json:"action"`
}

// CheckPermissionResponse 定义 RBAC 检查响应。
type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	UserID  string `json:"userId"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

// AssignRoleRequest 定义角色分配请求。
type AssignRoleRequest struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// RevokeRoleRequest 定义角色撤销请求。
type RevokeRoleRequest struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// UserRolesResponse 定义用户角色列表响应。
type UserRolesResponse struct {
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"`
}

// RoleUsersResponse 定义角色成员列表响应。
type RoleUsersResponse struct {
	Role    string   `json:"role"`
	UserIDs []string `json:"userIds"`
}

// PolicyRequest 定义单条策略请求。
type PolicyRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

// PoliciesResponse 定义策略列表响应。
type PoliciesResponse struct {
	Items []PolicyRequest `json:"items"`
}
