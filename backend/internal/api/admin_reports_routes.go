package api

import (
	"database/sql"

	"pos-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAdminReportsRoutes wires admin dashboard and reports endpoints
func SetupAdminReportsRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin")
	admin.Use(authMiddleware)
	admin.Use(middleware.RequireRoles([]string{"admin", "manager"}))
	{
		admin.GET("/dashboard/stats", getDashboardStats(db))
		admin.GET("/reports/sales", getSalesReport(db))
		admin.GET("/reports/orders", getOrdersReport(db))
		admin.GET("/reports/income", getIncomeReport(db))
	}
}