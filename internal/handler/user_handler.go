package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alex-zaleskih/interview/internal/model"
)

type UserHandler struct {
	users UserService
}

func NewUserHandler(users UserService) *UserHandler {
	return &UserHandler{users: users}
}

// Me handles GET /api/me and returns the profile of the current user.
func (h *UserHandler) Me(c *gin.Context) {
	user, err := h.users.GetUser(c.Request.Context(), currentUserID(c))
	if errors.Is(err, model.ErrNotFound) {
		// The session is valid but the user no longer exists.
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, newUserResponse(user))
}
