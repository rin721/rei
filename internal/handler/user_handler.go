package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rin721/rei/internal/service"
	"github.com/rin721/rei/types/constants"
	typesuser "github.com/rin721/rei/types/user"
)

// UserHandler 负责用户资料接口绑定与响应。
type UserHandler struct {
	service service.UserService
}

// NewUserHandler 创建用户处理器。
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// GetMe 返回当前登录用户资料。
func (h *UserHandler) GetMe(c *gin.Context) {
	response, err := h.service.GetProfile(c.Request.Context(), c.GetString(constants.ContextKeyUserID))
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, response)
}

// UpdateMe 更新当前登录用户资料。
func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req typesuser.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	response, err := h.service.UpdateProfile(c.Request.Context(), c.GetString(constants.ContextKeyUserID), req)
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, response)
}
