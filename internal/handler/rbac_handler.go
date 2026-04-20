package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	rbacservice "github.com/rin721/rei/internal/service/rbac"
	"github.com/rin721/rei/types"
	"github.com/rin721/rei/types/constants"
)

// RBACHandler handles RBAC HTTP requests.
type RBACHandler struct {
	service rbacservice.UseCase
}

// NewRBACHandler creates an RBAC handler.
func NewRBACHandler(svc rbacservice.UseCase) *RBACHandler {
	return &RBACHandler{service: svc}
}

// Check handles permission-check requests.
func (h *RBACHandler) Check(c *gin.Context) {
	response, err := h.service.CheckPermission(c.Request.Context(), rbacservice.CheckPermissionQuery{
		UserID: valueOrFallback(c.Query("user_id"), c.GetString(constants.ContextKeyUserID)),
		Object: valueOrFallback(c.Query("object"), c.FullPath()),
		Action: valueOrFallback(c.Query("action"), c.Request.Method),
	})
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, types.CheckPermissionResponse{
		Allowed: response.Allowed,
		UserID:  response.UserID,
		Object:  response.Object,
		Action:  response.Action,
	})
}

// AssignRole handles role-assignment requests.
func (h *RBACHandler) AssignRole(c *gin.Context) {
	var req types.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.AssignRole(c.Request.Context(), rbacservice.AssignRoleCommand{
		UserID: req.UserID,
		Role:   req.Role,
	}); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{"message": "role assigned"})
}

// RevokeRole handles role-revocation requests.
func (h *RBACHandler) RevokeRole(c *gin.Context) {
	var req types.RevokeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.RevokeRole(c.Request.Context(), rbacservice.RevokeRoleCommand{
		UserID: req.UserID,
		Role:   req.Role,
	}); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{"message": "role revoked"})
}

// GetUserRoles returns the roles assigned to a user.
func (h *RBACHandler) GetUserRoles(c *gin.Context) {
	response, err := h.service.GetUserRoles(c.Request.Context(), rbacservice.GetUserRolesQuery{
		UserID: c.Param("user_id"),
	})
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, types.UserRolesResponse{
		UserID: response.UserID,
		Roles:  response.Roles,
	})
}

// GetRoleUsers returns the users assigned to a role.
func (h *RBACHandler) GetRoleUsers(c *gin.Context) {
	response, err := h.service.GetUsersForRole(c.Request.Context(), rbacservice.GetUsersForRoleQuery{
		Role: c.Param("role"),
	})
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, types.RoleUsersResponse{
		Role:    response.Role,
		UserIDs: response.UserIDs,
	})
}

// AddPolicy handles policy-creation requests.
func (h *RBACHandler) AddPolicy(c *gin.Context) {
	var req types.PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.AddPolicy(c.Request.Context(), rbacservice.PolicyCommand{
		Subject: req.Subject,
		Object:  req.Object,
		Action:  req.Action,
	}); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusCreated, gin.H{"message": "policy added"})
}

// RemovePolicy handles policy-removal requests.
func (h *RBACHandler) RemovePolicy(c *gin.Context) {
	var req types.PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.RemovePolicy(c.Request.Context(), rbacservice.PolicyCommand{
		Subject: req.Subject,
		Object:  req.Object,
		Action:  req.Action,
	}); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{"message": "policy removed"})
}

// ListPolicies returns all current policies.
func (h *RBACHandler) ListPolicies(c *gin.Context) {
	response, err := h.service.ListPolicies(c.Request.Context())
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	items := make([]types.PolicyRequest, 0, len(response.Items))
	for _, item := range response.Items {
		items = append(items, types.PolicyRequest{
			Subject: item.Subject,
			Object:  item.Object,
			Action:  item.Action,
		})
	}

	writeSuccess(c, http.StatusOK, types.PoliciesResponse{Items: items})
}

func valueOrFallback(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
