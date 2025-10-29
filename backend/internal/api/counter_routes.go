package api

import (
	"database/sql"

	"pos-backend/internal/handlers"
	"pos-backend/internal/middleware"
	"pos-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// SetupCounterRoutes configures counter role endpoints (orders and payments)
func SetupCounterRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	counter := router.Group("/counter")
	counter.Use(authMiddleware)
	counter.Use(middleware.RequireRoles([]string{"admin", "counter"}))
	{
		orderRepo := repository.NewPostgresOrderRepository(db)
		orderHandler := handlers.NewOrderHandler(orderRepo)
		paymentHandler := handlers.NewPaymentHandler(db)

		counter.POST("/orders", orderHandler.CreateOrder)
		counter.POST("/orders/:id/payments", paymentHandler.ProcessPayment)
	}
}
