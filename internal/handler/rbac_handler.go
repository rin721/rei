package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rin721/go-scaffold2/internal/service"
	"github.com/rin721/go-scaffold2/types"
	"github.com/rin721/go-scaffold2/types/constants"
)

// RBACHandler 负责 RBAC 管理接口绑定与响应。
type RBACHandler struct {
	service service.RBACService
}

// NewRBACHandler 创建 RBAC 处理器。
func NewRBACHandler(svc service.RBACService) *RBACHandler {
	return &RBACHandler{service: svc}
}

// Check 处理权限检查请求。
func (h *RBACHandler) Check(c *gin.Context) {
	response, err := h.service.CheckPermission(c.Request.Context(), types.CheckPermissionRequest{
		UserID: valueOrFallback(c.Query("user_id"), c.GetString(constants.ContextKeyUserID)),
		Object: valueOrFallback(c.Query("object"), c.FullPath()),
		Action: valueOrFallback(c.Query("action"), c.Request.Method),
	})
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, response)
}

// AssignRole 处理角色分配请求。
func (h *RBACHandler) AssignRole(c *gin.Context) {
	var req types.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.AssignRole(c.Request.Context(), req); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{"message": "role assigned"})
}

// RevokeRole 处理角色撤销请求。
func (h *RBACHandler) RevokeRole(c *gin.Context) {
	var req types.RevokeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.RevokeRole(c.Request.Context(), req); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{"message": "role revoked"})
}

// GetUserRoles 返回指定用户角色列表。
func (h *RBACHandler) GetUserRoles(c *gin.Context) {
	response, err := h.service.GetUserRoles(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, response)
}

// GetRoleUsers 返回指定角色用户列表。
func (h *RBACHandler) GetRoleUsers(c *gin.Context) {
	response, err := h.service.GetUsersForRole(c.Request.Context(), c.Param("role"))
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, response)
}

// AddPolicy 处理策略新增请求。
func (h *RBACHandler) AddPolicy(c *gin.Context) {
	var req types.PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.AddPolicy(c.Request.Context(), req); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusCreated, gin.H{"message": "policy added"})
}

// RemovePolicy 处理策略删除请求。
func (h *RBACHandler) RemovePolicy(c *gin.Context) {
	var req types.PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.RemovePolicy(c.Request.Context(), req); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{"message": "policy removed"})
}

// ListPolicies 返回全部策略。
func (h *RBACHandler) ListPolicies(c *gin.Context) {
	response, err := h.service.ListPolicies(c.Request.Context())
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, response)
}

func valueOrFallback(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
