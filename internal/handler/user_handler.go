package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	userservice "github.com/rin721/rei/internal/service/user"
	"github.com/rin721/rei/types/constants"
	typesuser "github.com/rin721/rei/types/user"
)

// UserHandler handles user profile HTTP requests.
type UserHandler struct {
	service userservice.UseCase
}

// NewUserHandler creates a user handler.
func NewUserHandler(svc userservice.UseCase) *UserHandler {
	return &UserHandler{service: svc}
}

// GetMe returns the current user's profile.
func (h *UserHandler) GetMe(c *gin.Context) {
	response, err := h.service.GetProfile(c.Request.Context(), userservice.GetProfileQuery{
		UserID: c.GetString(constants.ContextKeyUserID),
	})
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, toUserProfileResponse(response))
}

// UpdateMe updates the current user's profile.
func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req typesuser.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindFailure(c, err)
		return
	}

	response, err := h.service.UpdateProfile(c.Request.Context(), userservice.UpdateProfileCommand{
		UserID:      c.GetString(constants.ContextKeyUserID),
		DisplayName: req.DisplayName,
		Email:       req.Email,
	})
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, toUserProfileResponse(response))
}

func toUserProfileResponse(profile userservice.Profile) typesuser.Profile {
	return typesuser.Profile{
		ID:          profile.ID,
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
		Email:       profile.Email,
		Roles:       append([]string(nil), profile.Roles...),
		CreatedAt:   profile.CreatedAt.UTC().Unix(),
		UpdatedAt:   profile.UpdatedAt.UTC().Unix(),
	}
}
