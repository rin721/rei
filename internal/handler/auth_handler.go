package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rei0721/go-scaffold2/internal/service"
	"github.com/rei0721/go-scaffold2/types/constants"
	typesuser "github.com/rei0721/go-scaffold2/types/user"
)

// AuthHandler 负责认证接口绑定与响应。
type AuthHandler struct {
	service service.AuthService
}

// NewAuthHandler 创建认证处理器。
func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
}

// Register 处理注册请求。
func (h *AuthHandler) Register(c *gin.Context) {
	var req typesuser.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	response, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusCreated, response)
}

// Login 处理登录请求。
func (h *AuthHandler) Login(c *gin.Context) {
	var req typesuser.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	response, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, response)
}

// Logout 处理登出请求。
func (h *AuthHandler) Logout(c *gin.Context) {
	if err := h.service.Logout(c.Request.Context(), c.GetString(constants.ContextKeyUserID)); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{
		"message": "logout succeeded",
	})
}

// ChangePassword 处理改密请求。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req typesuser.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), c.GetString(constants.ContextKeyUserID), req); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{
		"message": "password changed",
	})
}

// Refresh 处理刷新令牌请求。
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req typesuser.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	response, err := h.service.RefreshToken(c.Request.Context(), req)
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, response)
}
