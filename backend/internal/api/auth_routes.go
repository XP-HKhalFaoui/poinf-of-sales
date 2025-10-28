package api

import (
	"database/sql"

	"pos-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures public and protected auth endpoints
func SetupAuthRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	authHandler := handlers.NewAuthHandler(db)

	// Public routes (no authentication required)
	public := router.Group("/")
	{
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/logout", authHandler.Logout)
	}

	// Protected routes (authentication required)
	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/auth/me", authHandler.GetCurrentUser)
	}
}