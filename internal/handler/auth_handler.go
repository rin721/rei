package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authservice "github.com/rin721/rei/internal/service/auth"
	"github.com/rin721/rei/types/constants"
	typesuser "github.com/rin721/rei/types/user"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	service authservice.UseCase
}

// NewAuthHandler creates an auth handler.
func NewAuthHandler(svc authservice.UseCase) *AuthHandler {
	return &AuthHandler{service: svc}
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req typesuser.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	response, err := h.service.Register(c.Request.Context(), authservice.RegisterCommand{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Email:       req.Email,
	})
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusCreated, toAuthResponse(response))
}

// Login handles authentication requests.
func (h *AuthHandler) Login(c *gin.Context) {
	var req typesuser.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	response, err := h.service.Login(c.Request.Context(), authservice.LoginCommand{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, toAuthResponse(response))
}

// Logout clears the current session.
func (h *AuthHandler) Logout(c *gin.Context) {
	if err := h.service.Logout(c.Request.Context(), authservice.LogoutCommand{
		UserID: c.GetString(constants.ContextKeyUserID),
	}); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{
		"message": "logout succeeded",
	})
}

// ChangePassword updates the current user's password.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req typesuser.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), authservice.ChangePasswordCommand{
		UserID:      c.GetString(constants.ContextKeyUserID),
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}); err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, gin.H{
		"message": "password changed",
	})
}

// Refresh issues a new token pair from a refresh token.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req typesuser.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	response, err := h.service.RefreshToken(c.Request.Context(), authservice.RefreshTokenCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, toAuthResponse(response))
}

func toAuthResponse(result authservice.Authentication) typesuser.AuthResponse {
	return typesuser.AuthResponse{
		User: typesuser.Profile{
			ID:          result.User.ID,
			Username:    result.User.Username,
			DisplayName: result.User.DisplayName,
			Email:       result.User.Email,
			Roles:       append([]string(nil), result.User.Roles...),
			CreatedAt:   result.User.CreatedAt.UTC().Unix(),
			UpdatedAt:   result.User.UpdatedAt.UTC().Unix(),
		},
		Tokens: typesuser.TokenPair{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			TokenType:    result.Tokens.TokenType,
		},
	}
}
