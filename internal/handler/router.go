package handler

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/alex-zaleskih/interview/internal/auth"
)

type Deps struct {
	Orders   OrderService
	Users    UserService
	Sessions *auth.Sessions
	WebDir   string // directory with index.html, login.html and static/
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Frontend
	web := r.Group("/", NoCache())
	web.StaticFile("/", filepath.Join(d.WebDir, "index.html"))
	web.StaticFile("/login", filepath.Join(d.WebDir, "login.html"))
	web.Static("/static", filepath.Join(d.WebDir, "static"))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authHandler := NewAuthHandler(d.Users, d.Sessions)
	userHandler := NewUserHandler(d.Users)
	orderHandler := NewOrderHandler(d.Orders)

	api := r.Group("/api")
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/logout", authHandler.Logout)

	protected := api.Group("", RequireAuth(d.Sessions))
	protected.GET("/me", userHandler.Me)
	protected.GET("/orders", orderHandler.List)

	return r
}
