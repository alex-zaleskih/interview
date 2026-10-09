package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alex-zaleskih/interview/internal/auth"
)

const (
	sessionCookieName = "session"
	userIDContextKey  = "userID"
)

// RequireAuth rejects requests without a valid session cookie
// and stores the authenticated user ID in the gin context.
func RequireAuth(sessions *auth.Sessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(sessionCookieName)
		if err != nil {
			respondError(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID, ok := sessions.Verify(token)
		if !ok {
			respondError(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		c.Set(userIDContextKey, userID)
		c.Next()
	}
}

// NoCache makes the browser revalidate responses on every request,
// so changes in web/ are visible after a regular page reload.
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.Next()
	}
}

// currentUserID returns the ID of the authenticated user. Use only behind RequireAuth.
func currentUserID(c *gin.Context) string {
	return c.GetString(userIDContextKey)
}
