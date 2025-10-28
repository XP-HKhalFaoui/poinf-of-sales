package api

import (
	"database/sql"

	"pos-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupPaymentRoutes configures protected payment viewing endpoints
func SetupPaymentRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	paymentHandler := handlers.NewPaymentHandler(db)

	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/orders/:id/payments", paymentHandler.GetPayments)
		protected.GET("/orders/:id/payment-summary", paymentHandler.GetPaymentSummary)
	}
}