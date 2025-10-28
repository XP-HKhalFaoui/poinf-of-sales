package api

import (
	"database/sql"

	"pos-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupTableRoutes configures table endpoints for protected users
func SetupTableRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	tableHandler := handlers.NewTableHandler(db)

	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/tables", tableHandler.GetTables)
		protected.GET("/tables/:id", tableHandler.GetTable)
		protected.GET("/tables/by-location", tableHandler.GetTablesByLocation)
		protected.GET("/tables/status", tableHandler.GetTableStatus)
	}
}