package handlers

import (
	"POS-Golang/internal/database"
	"POS-Golang/internal/models"
	"POS-Golang/internal/utils"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TransactionItemRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

type TransactionRequest struct {
	Items         []TransactionItemRequest `json:"items" validate:"required,min=1"`
	PaymentMethod string                   `json:"payment_method" validate:"required,oneof=cash card transfer"`
	PaymentAmount float64                  `json:"payment_amount" validate:"required,gt=0"`
}

// Generate transaction number
func generateTransactionNo() string {
	now := time.Now()
	return fmt.Sprintf("TRX-%s-%d", now.Format("20060102"), now.Unix())
}

// Get all transactions
func GetTransactions(c *gin.Context) {
	db := database.GetDB()
	var transactions []models.Transaction

	// Query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	status := c.Query("status")

	query := db.Preload("User").Preload("Items.Product")

	// Apply filters
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}

	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records
	var total int64
	query.Model(&models.Transaction{}).Count(&total)

	// Apply pagination and ordering
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		utils.ErrorResponse(c, "Failed to fetch transactions", err)
		return
	}

	utils.SuccessResponse(c, "Transactions fetched successfully", gin.H{
		"transactions": transactions,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// Get single transaction
func GetTransaction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	db := database.GetDB()
	var transaction models.Transaction

	if err := db.Preload("User").Preload("Items.Product.Category").First(&transaction, id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction fetched successfully",
		"data":    transaction,
	})
}

// Create transaction
func CreateTransaction(c *gin.Context) {
	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, "Validation error", map[string]string{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, "Validation error", map[string]string{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, "User not authenticated", nil)
		return
	}

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate products and calculate total
	var totalAmount float64
	var transactionItems []models.TransactionItem

	for _, item := range req.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, fmt.Sprintf("Product with ID %d not found", item.ProductID), err)
			return
		}

		// Check stock availability
		if product.Stock < item.Quantity {
			tx.Rollback()
			utils.ErrorResponse(c, fmt.Sprintf("Insufficient stock for product %s", product.Name), nil)
			return
		}

		// Check if product is active
		if !product.IsActive {
			tx.Rollback()
			utils.ErrorResponse(c, fmt.Sprintf("Product %s is not active", product.Name), nil)
			return
		}

		subtotal := product.Price * float64(item.Quantity)
		totalAmount += subtotal

		transactionItems = append(transactionItems, models.TransactionItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			Subtotal:  subtotal,
		})

		// Update product stock
		product.Stock -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, "Failed to update product stock", err)
			return
		}
	}

	// Validate payment amount
	if req.PaymentAmount < totalAmount {
		tx.Rollback()
		utils.ErrorResponse(c, "Insufficient payment amount", nil)
		return
	}

	// Calculate change
	changeAmount := req.PaymentAmount - totalAmount

	// Create transaction
	transaction := models.Transaction{
		TransactionNo: generateTransactionNo(),
		UserID:        uint(userID.(float64)),
		TotalAmount:   totalAmount,
		PaymentMethod: req.PaymentMethod,
		PaymentAmount: req.PaymentAmount,
		ChangeAmount:  changeAmount,
		Status:        "completed",
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, "Failed to create transaction", err)
		return
	}

	// Create transaction items
	for i := range transactionItems {
		transactionItems[i].TransactionID = transaction.ID
	}

	if err := tx.Create(&transactionItems).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, "Failed to create transaction items", err)
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, "Failed to commit transaction", err)
		return
	}

	// Load complete transaction data
	db.Preload("User").Preload("Items.Product").First(&transaction, transaction.ID)

	utils.SuccessResponse(c, "Transaction created successfully", transaction)
}
