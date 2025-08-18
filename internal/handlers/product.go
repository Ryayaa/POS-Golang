package handlers

import (
	"POS-Golang/internal/database"
	"POS-Golang/internal/models"
	"POS-Golang/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"required,gte=0"`
	CategoryID  uint    `json:"category_id"`
	Barcode     string  `json:"barcode"`
	IsActive    *bool   `json:"is_active"`
}

// Get all products
func GetProducts(c *gin.Context) {
	db := database.GetDB()
	var products []models.Product

	// Query parameters for filtering
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	category := c.Query("category")

	query := db.Preload("Category")

	// Apply filters
	if search != "" {
		query = query.Where("name LIKE ? OR barcode LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if category != "" {
		categoryID, _ := strconv.Atoi(category)
		if categoryID > 0 {
			query = query.Where("category_id = ?", categoryID)
		}
	}

	// Count total records
	var total int64
	query.Model(&models.Product{}).Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		utils.ErrorResponse(c, "Failed to fetch products", err)
		return
	}

	utils.SuccessResponse(c, "Products fetched successfully", gin.H{
		"products": products,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// Get single product
func GetProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	db := database.GetDB()
	var product models.Product

	if err := db.Preload("Category").First(&product, id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product fetched successfully",
		"data":    product,
	})
}

// Create product
func CreateProduct(c *gin.Context) {
	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, "Validation error", map[string]string{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, "Validation error", map[string]string{"error": err.Error()})
		return
	}

	if req.CategoryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID tidak boleh 0"})
		return
	}

	db := database.GetDB()

	// Cek apakah kategori ada
	var category models.Category
	if err := db.First(&category, req.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category tidak ditemukan"})
		return
	}

	// Check if barcode already exists (if provided)
	if req.Barcode != "" {
		var existingProduct models.Product
		if err := db.Where("barcode = ?", req.Barcode).First(&existingProduct).Error; err == nil {
			utils.ErrorResponse(c, "Barcode already exists", nil)
			return
		}
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		Barcode:     req.Barcode,
		IsActive:    true,
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := db.Create(&product).Error; err != nil {
		utils.ErrorResponse(c, "Failed to create product", err)
		return
	}

	// Load the category relation
	db.Preload("Category").First(&product, product.ID)

	utils.SuccessResponse(c, "Product created successfully", product)
}

// Update product
func UpdateProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, "Invalid product ID", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, "Validation error", map[string]string{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, "Validation error", map[string]string{"error": err.Error()})
		return
	}

	if req.CategoryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID tidak boleh 0"})
		return
	}

	db := database.GetDB()
	var product models.Product

	if err := db.First(&product, id).Error; err != nil {
		utils.ErrorResponse(c, "Product not found", err)
		return
	}

	// Cek apakah kategori ada
	var category models.Category
	if err := db.First(&category, req.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category tidak ditemukan"})
		return
	}

	// Check if barcode already exists (if changed and provided)
	if req.Barcode != "" && req.Barcode != product.Barcode {
		var existingProduct models.Product
		if err := db.Where("barcode = ? AND id != ?", req.Barcode, id).First(&existingProduct).Error; err == nil {
			utils.ErrorResponse(c, "Barcode already exists", nil)
			return
		}
	}

	// Update fields
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.CategoryID = req.CategoryID
	product.Barcode = req.Barcode

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := db.Save(&product).Error; err != nil {
		utils.ErrorResponse(c, "Failed to update product", err)
		return
	}

	// Load the category relation
	db.Preload("Category").First(&product, product.ID)

	utils.SuccessResponse(c, "Product updated successfully", product)
}

// Delete product
func DeleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, "Invalid product ID", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	db := database.GetDB()
	var product models.Product

	if err := db.First(&product, id).Error; err != nil {
		utils.ErrorResponse(c, "Product not found", err)
		return
	}

	if err := db.Delete(&product).Error; err != nil {
		utils.ErrorResponse(c, "Failed to delete product", err)
		return
	}

	utils.SuccessResponse(c, "Product deleted successfully", nil)
}
