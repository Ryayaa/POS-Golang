package handlers

import (
	"POS-Golang/internal/database"
	"POS-Golang/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get all categories
func GetCategories(c *gin.Context) {
	db := database.GetDB()
	var categories []models.Category

	if err := db.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Categories fetched successfully",
		"categories": categories,
	})
}

// Create category
type CategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

func CreateCategory(c *gin.Context) {
	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	db := database.GetDB()
	category := models.Category{
		Name: req.Name,
	}

	if err := db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Category created successfully",
		"category": category,
	})
}
