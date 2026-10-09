package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alex-zaleskih/interview/internal/auth"
	"github.com/alex-zaleskih/interview/internal/model"
	"github.com/alex-zaleskih/interview/internal/service"
)

type UserService interface {
	Login(ctx context.Context, email string) (model.User, error)
	GetUser(ctx context.Context, id string) (model.User, error)
}

type AuthHandler struct {
	users    UserService
	sessions *auth.Sessions
}

func NewAuthHandler(users UserService, sessions *auth.Sessions) *AuthHandler {
	return &AuthHandler{users: users, sessions: sessions}
}

type loginRequest struct {
	Email string `json:"email" binding:"required"`
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "email is required")
		return
	}

	user, err := h.users.Login(c.Request.Context(), req.Email)
	if errors.Is(err, service.ErrInvalidCredentials) {
		respondError(c, http.StatusUnauthorized, "user with this email not found")
		return
	}
	if err != nil {
		respondInternalError(c, err)
		return
	}

	h.setSessionCookie(c, h.sessions.Issue(user.ID), int(h.sessions.TTL().Seconds()))
	c.JSON(http.StatusOK, newUserResponse(user))
}

// Logout handles POST /api/auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	h.setSessionCookie(c, "", -1)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, value, maxAge, "/", "", false, true)
}
