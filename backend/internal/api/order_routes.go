package api

import (
	"database/sql"

	"pos-backend/internal/handlers"
	"pos-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// SetupOrderRoutes configures protected order viewing endpoints
func SetupOrderRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	orderRepo := repository.NewPostgresOrderRepository(db)
	orderHandler := handlers.NewOrderHandler(orderRepo)

	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/orders", orderHandler.GetOrders)
		protected.GET("/orders/:id", orderHandler.GetOrder)
		protected.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)
	}
}
