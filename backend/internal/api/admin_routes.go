package api

import (
	"database/sql"

	"pos-backend/internal/handlers"
	"pos-backend/internal/middleware"
	"pos-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes configures endpoints for admin/manager roles
func SetupAdminRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin")
	admin.Use(authMiddleware)
	admin.Use(middleware.RequireRoles([]string{"admin", "manager"}))
	{
		// Dashboard and monitoring (moved to admin_reports_routes.go)

		// Menu management with pagination
		productHandler := handlers.NewProductHandler(db)
		admin.GET("/products", productHandler.GetProducts) // Use existing paginated handler
		admin.GET("/categories", getAdminCategories(db))   // Add pagination
		admin.POST("/categories", createCategory(db))
		admin.PUT("/categories/:id", updateCategory(db))
		admin.DELETE("/categories/:id", deleteCategory(db))
		admin.POST("/products", createProduct(db))
		admin.PUT("/products/:id", updateProduct(db))
		admin.DELETE("/products/:id", deleteProduct(db))

		// Table management with pagination
		admin.GET("/tables", getAdminTables(db)) // Add pagination
		admin.POST("/tables", createTable(db))
		admin.PUT("/tables/:id", updateTable(db))
		admin.DELETE("/tables/:id", deleteTable(db))

		// User management with pagination
		admin.GET("/users", getAdminUsers(db)) // Update with pagination
		admin.POST("/users", createUser(db))
		admin.PUT("/users/:id", updateUser(db))
		admin.DELETE("/users/:id", deleteUser(db))

		// Advanced order management
		orderRepo := repository.NewPostgresOrderRepository(db)
		orderHandler := handlers.NewOrderHandler(orderRepo)
		paymentHandler := handlers.NewPaymentHandler(db)
		admin.POST("/orders", orderHandler.CreateOrder)                   // Admins can create any type of order
		admin.POST("/orders/:id/payments", paymentHandler.ProcessPayment) // Admins can process payments
	}
}
