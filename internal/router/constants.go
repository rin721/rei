package router

const (
	// RouteHealth 定义健康检查路由。
	RouteHealth = "/health"
	// RouteAuthRegister 定义注册路由。
	RouteAuthRegister = "/api/v1/auth/register"
	// RouteAuthLogin 定义登录路由。
	RouteAuthLogin = "/api/v1/auth/login"
	// RouteAuthRefresh 定义刷新令牌路由。
	RouteAuthRefresh = "/api/v1/auth/refresh"
	// RouteAuthLogout 定义登出路由。
	RouteAuthLogout = "/api/v1/auth/logout"
	// RouteAuthChangePassword 定义改密路由。
	RouteAuthChangePassword = "/api/v1/auth/change-password"
	// RouteUserMe 定义当前用户资料路由。
	RouteUserMe = "/api/v1/users/me"
	// RouteRBACCheck 定义权限检查路由。
	RouteRBACCheck = "/api/v1/rbac/check"
	// RouteRBACAssignRole 定义角色分配路由。
	RouteRBACAssignRole = "/api/v1/rbac/roles/assign"
	// RouteRBACRevokeRole 定义角色撤销路由。
	RouteRBACRevokeRole = "/api/v1/rbac/roles/revoke"
	// RouteRBACUserRoles 定义按用户查询角色路由。
	RouteRBACUserRoles = "/api/v1/rbac/users/:user_id/roles"
	// RouteRBACRoleUsers 定义按角色查询用户路由。
	RouteRBACRoleUsers = "/api/v1/rbac/roles/:role/users"
	// RouteRBACPolicies 定义策略管理路由。
	RouteRBACPolicies = "/api/v1/rbac/policies"
	// RouteSampleList 定义示例模块列表路由。
	RouteSampleList = "/api/v1/samples"
)
