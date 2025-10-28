package api

import (
	"database/sql"

	"pos-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupProductRoutes configures product and category endpoints for protected users
func SetupProductRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	productHandler := handlers.NewProductHandler(db)

	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/products", productHandler.GetProducts)
		protected.GET("/products/:id", productHandler.GetProduct)
		protected.GET("/categories", productHandler.GetCategories)
		protected.GET("/categories/:id/products", productHandler.GetProductsByCategory)
	}
}