package api

import (
	"database/sql"

	"pos-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupKitchenRoutes configures endpoints for kitchen staff
func SetupKitchenRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	kitchen := router.Group("/kitchen")
	kitchen.Use(authMiddleware)
	kitchen.Use(middleware.RequireRoles([]string{"kitchen", "admin", "manager"}))
	{
		kitchen.GET("/orders", getKitchenOrders(db))
		kitchen.PATCH("/orders/:id/items/:item_id/status", updateOrderItemStatus(db))
	}
}

// getKitchenOrders retrieves active orders for kitchen processing
func getKitchenOrders(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.DefaultQuery("status", "all")

		query := `
			SELECT DISTINCT o.id::text, o.order_number, o.table_id::text, o.order_type, o.status, 
		       o.created_at, o.customer_name,
		       t.table_number
			FROM orders o
			LEFT JOIN dining_tables t ON o.table_id = t.id
			WHERE o.status IN ('pending', 'confirmed', 'preparing', 'ready') AND o.created_at::date = CURRENT_DATE
		`
		// WHERE o.status IN ('pending', 'confirmed', 'preparing', 'ready') AND o.created_at::date = CURRENT_DATE

		if status != "all" {
			query += ` AND o.status = '` + status + `'`
		}

		query += ` ORDER BY o.created_at ASC`

		rows, err := db.Query(query)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to fetch kitchen orders",
				"error":   err.Error(),
			})
			return
		}
		defer rows.Close()

		var orders []map[string]interface{}
		for rows.Next() {
			var orderID, tableID interface{}
			var orderNumber, orderType, orderStatus, customerName, tableNumber sql.NullString
			var createdAt interface{}

			err := rows.Scan(&orderID, &orderNumber, &tableID, &orderType, &orderStatus,
				&createdAt, &customerName, &tableNumber)
			if err != nil {
				c.JSON(500, gin.H{
					"success": false,
					"message": "Failed to scan kitchen order",
					"error":   err.Error(),
				})
				return
			}

			order := map[string]interface{}{
				"id":            orderID,
				"order_number":  orderNumber.String,
				"table_id":      tableID,
				"table_number":  tableNumber.String,
				"order_type":    orderType.String,
				"status":        orderStatus.String,
				"customer_name": customerName.String,
				"created_at":    createdAt,
			}

			orders = append(orders, order)
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Kitchen orders retrieved successfully",
			"data":    orders,
		})
	}
}

// updateOrderItemStatus updates status of an order item
func updateOrderItemStatus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		itemID := c.Param("item_id")

		var req struct {
			Status string `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		// Update order item status
		_, err := db.Exec(`
			UPDATE order_items 
			SET status = $1, updated_at = CURRENT_TIMESTAMP 
			WHERE id = $2 AND order_id = $3
		`, req.Status, itemID, orderID)

		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to update order item status",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Order item status updated successfully",
		})
	}
}
