package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"strings"

	"pos-backend/internal/handlers"
	"pos-backend/internal/middleware"
	"pos-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// SetupServerRoutes configures server role endpoints (dine-in orders only)
func SetupServerRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	server := router.Group("/server")
	server.Use(authMiddleware)
	server.Use(middleware.RequireRole("server"))
	{
		server.POST("/orders", createDineInOrder(db))
	}
}

// createDineInOrder enforces dine_in order type for server role
func createDineInOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TableID      *string `json:"table_id"`
			CustomerName *string `json:"customer_name"`
			Items        []struct {
				ProductID           string  `json:"product_id"`
				Quantity            int     `json:"quantity"`
				SpecialInstructions *string `json:"special_instructions"`
			} `json:"items"`
			Notes *string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		// Force order type to dine_in for servers
		orderRepo := repository.NewPostgresOrderRepository(db)
		orderHandler := handlers.NewOrderHandler(orderRepo)

		// Create order request with forced dine_in type
		createOrderReq := map[string]interface{}{
			"table_id":      req.TableID,
			"customer_name": req.CustomerName,
			"order_type":    "dine_in", // Force dine-in for servers
			"items":         req.Items,
			"notes":         req.Notes,
		}

		// Convert to JSON and back to simulate the request
		reqBytes, _ := json.Marshal(createOrderReq)
		c.Request.Body = io.NopCloser(strings.NewReader(string(reqBytes)))

		orderHandler.CreateOrder(c)
	}
}
