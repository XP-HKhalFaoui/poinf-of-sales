package api

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"pos-backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// SetupRoutes configures all API routes (orchestrator)
func SetupRoutes(router *gin.RouterGroup, db *sql.DB, authMiddleware gin.HandlerFunc) {
	// Delegate to modular route setups
	SetupAuthRoutes(router, db, authMiddleware)
	SetupProductRoutes(router, db, authMiddleware)
	SetupTableRoutes(router, db, authMiddleware)
	SetupOrderRoutes(router, db, authMiddleware)
	SetupPaymentRoutes(router, db, authMiddleware)
	SetupServerRoutes(router, db, authMiddleware)
	SetupCounterRoutes(router, db, authMiddleware)
	SetupAdminReportsRoutes(router, db, authMiddleware)
	SetupAdminRoutes(router, db, authMiddleware)
	SetupKitchenRoutes(router, db, authMiddleware)
}

// moved: getDashboardStats is now defined in admin_reports.go

// moved: getSalesReport is now defined in admin_reports.go

// moved: getOrdersReport is now defined in admin_reports.go

// moved: getKitchenOrders is now defined in kitchen_routes.go

// moved: updateOrderItemStatus is now defined in kitchen_routes.go

// moved: createDineInOrder is now defined in server_routes.go

// Admin handler - Income report
// moved: getIncomeReport is now defined in admin_reports.go

// Admin handler - Create category
func createCategory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string  `json:"name" binding:"required"`
			Description *string `json:"description"`
			Color       *string `json:"color"`
			SortOrder   int     `json:"sort_order"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		var categoryID string
		err := db.QueryRow(`
			INSERT INTO categories (name, description, color, sort_order)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, req.Name, req.Description, req.Color, req.SortOrder).Scan(&categoryID)

		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to create category",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"success": true,
			"message": "Category created successfully",
			"data":    map[string]interface{}{"id": categoryID},
		})
	}
}

// Admin handler - Update category
func updateCategory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Param("id")

		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
			Color       *string `json:"color"`
			SortOrder   *int    `json:"sort_order"`
			IsActive    *bool   `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		// Build dynamic update query
		updates := []string{}
		args := []interface{}{}
		argCount := 1

		if req.Name != nil {
			updates = append(updates, fmt.Sprintf("name = $%d", argCount))
			args = append(args, *req.Name)
			argCount++
		}
		if req.Description != nil {
			updates = append(updates, fmt.Sprintf("description = $%d", argCount))
			args = append(args, req.Description)
			argCount++
		}
		if req.Color != nil {
			updates = append(updates, fmt.Sprintf("color = $%d", argCount))
			args = append(args, req.Color)
			argCount++
		}
		if req.SortOrder != nil {
			updates = append(updates, fmt.Sprintf("sort_order = $%d", argCount))
			args = append(args, *req.SortOrder)
			argCount++
		}
		if req.IsActive != nil {
			updates = append(updates, fmt.Sprintf("is_active = $%d", argCount))
			args = append(args, *req.IsActive)
			argCount++
		}

		if len(updates) == 0 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "No fields to update",
			})
			return
		}

		updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
		args = append(args, categoryID)

		query := fmt.Sprintf(`
			UPDATE categories 
			SET %s 
			WHERE id = $%d
		`, strings.Join(updates, ", "), argCount)

		result, err := db.Exec(query, args...)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to update category",
				"error":   err.Error(),
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(404, gin.H{
				"success": false,
				"message": "Category not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Category updated successfully",
		})
	}
}

// Admin handler - Delete category
func deleteCategory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Param("id")

		// Check if category has products
		var productCount int
		db.QueryRow("SELECT COUNT(*) FROM products WHERE category_id = $1", categoryID).Scan(&productCount)

		if productCount > 0 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Cannot delete category with existing products",
				"error":   "category_has_products",
			})
			return
		}

		result, err := db.Exec("DELETE FROM categories WHERE id = $1", categoryID)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to delete category",
				"error":   err.Error(),
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(404, gin.H{
				"success": false,
				"message": "Category not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Category deleted successfully",
		})
	}
}

// Admin handler - Create product
func createProduct(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CategoryID      *string `json:"category_id"`
			Name            string  `json:"name" binding:"required"`
			Description     *string `json:"description"`
			Price           float64 `json:"price" binding:"required"`
			ImageURL        *string `json:"image_url"`
			Barcode         *string `json:"barcode"`
			SKU             *string `json:"sku"`
			PreparationTime int     `json:"preparation_time"`
			SortOrder       int     `json:"sort_order"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		var productID string
		err := db.QueryRow(`
			INSERT INTO products (category_id, name, description, price, image_url, barcode, sku, preparation_time, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`, req.CategoryID, req.Name, req.Description, req.Price, req.ImageURL, req.Barcode, req.SKU, req.PreparationTime, req.SortOrder).Scan(&productID)

		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to create product",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"success": true,
			"message": "Product created successfully",
			"data":    map[string]interface{}{"id": productID},
		})
	}
}

// Admin handler - Update product
func updateProduct(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		var req struct {
			CategoryID      *string  `json:"category_id"`
			Name            *string  `json:"name"`
			Description     *string  `json:"description"`
			Price           *float64 `json:"price"`
			ImageURL        *string  `json:"image_url"`
			Barcode         *string  `json:"barcode"`
			SKU             *string  `json:"sku"`
			IsAvailable     *bool    `json:"is_available"`
			PreparationTime *int     `json:"preparation_time"`
			SortOrder       *int     `json:"sort_order"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		// Build dynamic update query
		updates := []string{}
		args := []interface{}{}
		argCount := 1

		if req.CategoryID != nil {
			updates = append(updates, fmt.Sprintf("category_id = $%d", argCount))
			args = append(args, req.CategoryID)
			argCount++
		}
		if req.Name != nil {
			updates = append(updates, fmt.Sprintf("name = $%d", argCount))
			args = append(args, *req.Name)
			argCount++
		}
		if req.Description != nil {
			updates = append(updates, fmt.Sprintf("description = $%d", argCount))
			args = append(args, req.Description)
			argCount++
		}
		if req.Price != nil {
			updates = append(updates, fmt.Sprintf("price = $%d", argCount))
			args = append(args, *req.Price)
			argCount++
		}
		if req.ImageURL != nil {
			updates = append(updates, fmt.Sprintf("image_url = $%d", argCount))
			args = append(args, req.ImageURL)
			argCount++
		}
		if req.Barcode != nil {
			updates = append(updates, fmt.Sprintf("barcode = $%d", argCount))
			args = append(args, req.Barcode)
			argCount++
		}
		if req.SKU != nil {
			updates = append(updates, fmt.Sprintf("sku = $%d", argCount))
			args = append(args, req.SKU)
			argCount++
		}
		if req.IsAvailable != nil {
			updates = append(updates, fmt.Sprintf("is_available = $%d", argCount))
			args = append(args, *req.IsAvailable)
			argCount++
		}
		if req.PreparationTime != nil {
			updates = append(updates, fmt.Sprintf("preparation_time = $%d", argCount))
			args = append(args, *req.PreparationTime)
			argCount++
		}
		if req.SortOrder != nil {
			updates = append(updates, fmt.Sprintf("sort_order = $%d", argCount))
			args = append(args, *req.SortOrder)
			argCount++
		}

		if len(updates) == 0 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "No fields to update",
			})
			return
		}

		updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
		args = append(args, productID)

		query := fmt.Sprintf(`
			UPDATE products 
			SET %s 
			WHERE id = $%d
		`, strings.Join(updates, ", "), argCount)

		result, err := db.Exec(query, args...)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to update product",
				"error":   err.Error(),
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(404, gin.H{
				"success": false,
				"message": "Product not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Product updated successfully",
		})
	}
}

// Admin handler - Delete product
func deleteProduct(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		// Check if product is used in any active orders
		var orderCount int
		db.QueryRow(`
			SELECT COUNT(*) 
			FROM order_items oi 
			JOIN orders o ON oi.order_id = o.id 
			WHERE oi.product_id = $1 AND o.status NOT IN ('completed', 'cancelled')
		`, productID).Scan(&orderCount)

		if orderCount > 0 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Cannot delete product with active orders",
				"error":   "product_has_active_orders",
			})
			return
		}

		result, err := db.Exec("DELETE FROM products WHERE id = $1", productID)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to delete product",
				"error":   err.Error(),
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(404, gin.H{
				"success": false,
				"message": "Product not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Product deleted successfully",
		})
	}
}

// Admin handler - Create table
func createTable(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TableNumber     string  `json:"table_number" binding:"required"`
			SeatingCapacity int     `json:"seating_capacity"`
			Location        *string `json:"location"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		var tableID string
		err := db.QueryRow(`
			INSERT INTO dining_tables (table_number, seating_capacity, location)
			VALUES ($1, $2, $3)
			RETURNING id
		`, req.TableNumber, req.SeatingCapacity, req.Location).Scan(&tableID)

		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to create table",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"success": true,
			"message": "Table created successfully",
			"data":    map[string]interface{}{"id": tableID},
		})
	}
}

// Admin handler - Update table
func updateTable(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tableID := c.Param("id")

		var req struct {
			TableNumber     *string `json:"table_number"`
			SeatingCapacity *int    `json:"seating_capacity"`
			Location        *string `json:"location"`
			IsOccupied      *bool   `json:"is_occupied"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		// Build dynamic update query
		updates := []string{}
		args := []interface{}{}
		argCount := 1

		if req.TableNumber != nil {
			updates = append(updates, fmt.Sprintf("table_number = $%d", argCount))
			args = append(args, *req.TableNumber)
			argCount++
		}
		if req.SeatingCapacity != nil {
			updates = append(updates, fmt.Sprintf("seating_capacity = $%d", argCount))
			args = append(args, *req.SeatingCapacity)
			argCount++
		}
		if req.Location != nil {
			updates = append(updates, fmt.Sprintf("location = $%d", argCount))
			args = append(args, req.Location)
			argCount++
		}
		if req.IsOccupied != nil {
			updates = append(updates, fmt.Sprintf("is_occupied = $%d", argCount))
			args = append(args, *req.IsOccupied)
			argCount++
		}

		if len(updates) == 0 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "No fields to update",
			})
			return
		}

		updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
		args = append(args, tableID)

		query := fmt.Sprintf(`
			UPDATE dining_tables 
			SET %s 
			WHERE id = $%d
		`, strings.Join(updates, ", "), argCount)

		result, err := db.Exec(query, args...)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to update table",
				"error":   err.Error(),
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(404, gin.H{
				"success": false,
				"message": "Table not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Table updated successfully",
		})
	}
}

// Admin handler - Delete table
func deleteTable(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tableID := c.Param("id")

		// Check if table has active orders
		var orderCount int
		db.QueryRow(`
			SELECT COUNT(*) 
			FROM orders 
			WHERE table_id = $1 AND status NOT IN ('completed', 'cancelled')
		`, tableID).Scan(&orderCount)

		if orderCount > 0 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Cannot delete table with active orders",
				"error":   "table_has_active_orders",
			})
			return
		}

		result, err := db.Exec("DELETE FROM dining_tables WHERE id = $1", tableID)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to delete table",
				"error":   err.Error(),
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(404, gin.H{
				"success": false,
				"message": "Table not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Table deleted successfully",
		})
	}
}

// Admin handler - Create user
func createUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username  string `json:"username" binding:"required"`
			Email     string `json:"email" binding:"required"`
			Password  string `json:"password" binding:"required"`
			FirstName string `json:"first_name" binding:"required"`
			LastName  string `json:"last_name" binding:"required"`
			Role      string `json:"role" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to hash password",
				"error":   err.Error(),
			})
			return
		}

		var userID string
		err = db.QueryRow(`
			INSERT INTO users (username, email, password_hash, first_name, last_name, role)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, req.Username, req.Email, string(hashedPassword), req.FirstName, req.LastName, req.Role).Scan(&userID)

		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to create user",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"success": true,
			"message": "User created successfully",
			"data":    map[string]interface{}{"id": userID},
		})
	}
}

// Admin handler - Update user
func updateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var req struct {
			Username  *string `json:"username"`
			Email     *string `json:"email"`
			Password  *string `json:"password"`
			FirstName *string `json:"first_name"`
			LastName  *string `json:"last_name"`
			Role      *string `json:"role"`
			IsActive  *bool   `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		// Build dynamic update query
		updates := []string{}
		args := []interface{}{}
		argCount := 1

		if req.Username != nil {
			updates = append(updates, fmt.Sprintf("username = $%d", argCount))
			args = append(args, *req.Username)
			argCount++
		}
		if req.Email != nil {
			updates = append(updates, fmt.Sprintf("email = $%d", argCount))
			args = append(args, *req.Email)
			argCount++
		}
		if req.Password != nil {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(500, gin.H{
					"success": false,
					"message": "Failed to hash password",
					"error":   err.Error(),
				})
				return
			}
			updates = append(updates, fmt.Sprintf("password_hash = $%d", argCount))
			args = append(args, string(hashedPassword))
			argCount++
		}
		if req.FirstName != nil {
			updates = append(updates, fmt.Sprintf("first_name = $%d", argCount))
			args = append(args, *req.FirstName)
			argCount++
		}
		if req.LastName != nil {
			updates = append(updates, fmt.Sprintf("last_name = $%d", argCount))
			args = append(args, *req.LastName)
			argCount++
		}
		if req.Role != nil {
			updates = append(updates, fmt.Sprintf("role = $%d", argCount))
			args = append(args, *req.Role)
			argCount++
		}
		if req.IsActive != nil {
			updates = append(updates, fmt.Sprintf("is_active = $%d", argCount))
			args = append(args, *req.IsActive)
			argCount++
		}

		if len(updates) == 0 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "No fields to update",
			})
			return
		}

		updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
		args = append(args, userID)

		query := fmt.Sprintf(`
			UPDATE users 
			SET %s 
			WHERE id = $%d
		`, strings.Join(updates, ", "), argCount)

		result, err := db.Exec(query, args...)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to update user",
				"error":   err.Error(),
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(404, gin.H{
				"success": false,
				"message": "User not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "User updated successfully",
		})
	}
}

// Admin handler - Delete user
func deleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		// Prevent deletion if user has associated orders
		var orderCount int
		db.QueryRow("SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&orderCount)

		if orderCount > 0 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Cannot delete user with existing orders",
				"error":   "user_has_orders",
			})
			return
		}

		result, err := db.Exec("DELETE FROM users WHERE id = $1", userID)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to delete user",
				"error":   err.Error(),
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(404, gin.H{
				"success": false,
				"message": "User not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "User deleted successfully",
		})
	}
}

// Admin handler - Get users with pagination
func getAdminUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination parameters
		page := 1
		perPage := 20
		role := c.Query("role")
		isActive := c.Query("active")
		search := c.Query("search")

		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if perPageStr := c.Query("per_page"); perPageStr != "" {
			if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
				perPage = pp
			}
		}

		offset := (page - 1) * perPage

		// Build query with filters
		queryBuilder := "SELECT id, username, email, first_name, last_name, role, is_active, created_at FROM users WHERE 1=1"
		args := []interface{}{}
		argCount := 0

		if role != "" {
			argCount++
			queryBuilder += fmt.Sprintf(" AND role = $%d", argCount)
			args = append(args, role)
		}

		if isActive != "" {
			argCount++
			queryBuilder += fmt.Sprintf(" AND is_active = $%d", argCount)
			args = append(args, isActive == "true")
		}

		if search != "" {
			argCount++
			queryBuilder += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR username ILIKE $%d OR email ILIKE $%d)", argCount, argCount, argCount, argCount)
			args = append(args, "%"+search+"%")
		}

		// Count total records
		countQuery := "SELECT COUNT(*) FROM (" + queryBuilder + ") as count_query"
		var total int
		if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to count users",
				"error":   err.Error(),
			})
			return
		}

		// Add ordering and pagination
		queryBuilder += " ORDER BY created_at DESC"
		argCount++
		queryBuilder += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, perPage)

		argCount++
		queryBuilder += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)

		rows, err := db.Query(queryBuilder, args...)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to fetch users",
				"error":   err.Error(),
			})
			return
		}
		defer rows.Close()

		var users []map[string]interface{}
		for rows.Next() {
			var user map[string]interface{} = make(map[string]interface{})
			var id, username, email, firstName, lastName, userRole string
			var isActive bool
			var createdAt time.Time

			err := rows.Scan(&id, &username, &email, &firstName, &lastName, &userRole, &isActive, &createdAt)
			if err != nil {
				c.JSON(500, gin.H{
					"success": false,
					"message": "Failed to scan user data",
					"error":   err.Error(),
				})
				return
			}

			user["id"] = id
			user["username"] = username
			user["email"] = email
			user["first_name"] = firstName
			user["last_name"] = lastName
			user["role"] = userRole
			user["is_active"] = isActive
			user["created_at"] = createdAt

			users = append(users, user)
		}

		totalPages := (total + perPage - 1) / perPage

		c.JSON(200, gin.H{
			"success": true,
			"message": "Users retrieved successfully",
			"data":    users,
			"meta": gin.H{
				"current_page": page,
				"per_page":     perPage,
				"total":        total,
				"total_pages":  totalPages,
			},
		})
	}
}

// Admin handler - Get categories with pagination
func getAdminCategories(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination parameters
		page := 1
		perPage := 20
		activeOnly := c.Query("active_only") == "true"
		search := c.Query("search")

		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if perPageStr := c.Query("per_page"); perPageStr != "" {
			if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
				perPage = pp
			}
		}

		offset := (page - 1) * perPage

		// Build query with filters
		queryBuilder := "SELECT id, name, description, color, sort_order, is_active, created_at, updated_at FROM categories WHERE 1=1"
		args := []interface{}{}
		argCount := 0

		if activeOnly {
			queryBuilder += " AND is_active = true"
		}

		if search != "" {
			argCount++
			queryBuilder += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
			args = append(args, "%"+search+"%")
		}

		// Count total records
		countQuery := "SELECT COUNT(*) FROM (" + queryBuilder + ") as count_query"
		var total int
		if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to count categories",
				"error":   err.Error(),
			})
			return
		}

		// Add ordering and pagination
		queryBuilder += " ORDER BY sort_order ASC, name ASC"
		argCount++
		queryBuilder += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, perPage)

		argCount++
		queryBuilder += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)

		rows, err := db.Query(queryBuilder, args...)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to fetch categories",
				"error":   err.Error(),
			})
			return
		}
		defer rows.Close()

		var categories []models.Category
		for rows.Next() {
			var category models.Category

			err := rows.Scan(
				&category.ID, &category.Name, &category.Description, &category.Color,
				&category.SortOrder, &category.IsActive, &category.CreatedAt, &category.UpdatedAt,
			)
			if err != nil {
				c.JSON(500, gin.H{
					"success": false,
					"message": "Failed to scan category",
					"error":   err.Error(),
				})
				return
			}

			categories = append(categories, category)
		}

		totalPages := (total + perPage - 1) / perPage

		c.JSON(200, gin.H{
			"success": true,
			"message": "Categories retrieved successfully",
			"data":    categories,
			"meta": gin.H{
				"current_page": page,
				"per_page":     perPage,
				"total":        total,
				"total_pages":  totalPages,
			},
		})
	}
}

// Admin handler - Get tables with pagination
func getAdminTables(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination parameters
		page := 1
		perPage := 20
		location := c.Query("location")
		status := c.Query("status") // "occupied", "available", or empty for all
		search := c.Query("search")

		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if perPageStr := c.Query("per_page"); perPageStr != "" {
			if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
				perPage = pp
			}
		}

		offset := (page - 1) * perPage

		// Build query with filters
		queryBuilder := `
			SELECT t.id, t.table_number, t.seating_capacity, t.location, t.is_occupied, 
			       t.created_at, t.updated_at,
			       o.id as order_id, o.order_number, o.customer_name, o.status as order_status,
			       o.created_at as order_created_at, o.total_amount
			FROM dining_tables t
			LEFT JOIN orders o ON t.id = o.table_id AND o.status NOT IN ('completed', 'cancelled')
			WHERE 1=1
		`

		args := []interface{}{}
		argCount := 0

		if location != "" {
			argCount++
			queryBuilder += fmt.Sprintf(" AND t.location ILIKE $%d", argCount)
			args = append(args, "%"+location+"%")
		}

		if status == "occupied" {
			queryBuilder += " AND t.is_occupied = true"
		} else if status == "available" {
			queryBuilder += " AND t.is_occupied = false"
		}

		if search != "" {
			argCount++
			queryBuilder += fmt.Sprintf(" AND (t.table_number ILIKE $%d OR t.location ILIKE $%d)", argCount, argCount)
			args = append(args, "%"+search+"%")
		}

		// Count total records
		countQuery := "SELECT COUNT(*) FROM (" + queryBuilder + ") as count_query"
		var total int
		if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to count tables",
				"error":   err.Error(),
			})
			return
		}

		// Add ordering and pagination
		queryBuilder += " ORDER BY t.table_number ASC"
		argCount++
		queryBuilder += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, perPage)

		argCount++
		queryBuilder += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)

		rows, err := db.Query(queryBuilder, args...)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to fetch tables",
				"error":   err.Error(),
			})
			return
		}
		defer rows.Close()

		var tables []map[string]interface{}
		for rows.Next() {
			var table models.DiningTable
			var orderID, orderNumber, customerName, orderStatus sql.NullString
			var orderCreatedAt sql.NullTime
			var totalAmount sql.NullFloat64

			err := rows.Scan(
				&table.ID, &table.TableNumber, &table.SeatingCapacity, &table.Location, &table.IsOccupied,
				&table.CreatedAt, &table.UpdatedAt,
				&orderID, &orderNumber, &customerName, &orderStatus, &orderCreatedAt, &totalAmount,
			)
			if err != nil {
				c.JSON(500, gin.H{
					"success": false,
					"message": "Failed to scan table",
					"error":   err.Error(),
				})
				return
			}

			// Create table data with current order info
			tableData := map[string]interface{}{
				"id":               table.ID,
				"table_number":     table.TableNumber,
				"seating_capacity": table.SeatingCapacity,
				"location":         table.Location,
				"is_occupied":      table.IsOccupied,
				"created_at":       table.CreatedAt,
				"updated_at":       table.UpdatedAt,
				"current_order":    nil,
			}

			// Add current order info if available
			if orderID.Valid {
				tableData["current_order"] = map[string]interface{}{
					"id":            orderID.String,
					"order_number":  orderNumber.String,
					"customer_name": customerName.String,
					"status":        orderStatus.String,
					"created_at":    orderCreatedAt.Time,
					"total_amount":  totalAmount.Float64,
				}
			}

			tables = append(tables, tableData)
		}

		totalPages := (total + perPage - 1) / perPage

		c.JSON(200, gin.H{
			"success": true,
			"message": "Tables retrieved successfully",
			"data":    tables,
			"meta": gin.H{
				"current_page": page,
				"per_page":     perPage,
				"total":        total,
				"total_pages":  totalPages,
			},
		})
	}
}

// Helper function to convert string to pointer
func stringPtr(s string) *string {
	return &s
}
