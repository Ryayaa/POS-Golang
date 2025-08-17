package handlers

import (
	"POS-Golang/internal/database"
	"POS-Golang/internal/models"
	"POS-Golang/internal/utils"
	"net/http"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

var validate = validator.New()

type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" validate:"required"`
	Password        string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin cashier"`
}

// Login handler
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, "Invalid request", map[string]string{"error": err.Error()})
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, "Validation failed", map[string]string{"error": err.Error()})
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "Validation failed"})
		return
	}

	var user models.User
	db := database.GetDB()

	// Find user by username or email
	if err := db.Where("username = ? OR email = ?", req.UsernameOrEmail, req.UsernameOrEmail).First(&user).Error; err != nil {
		utils.ErrorResponse(c, "Invalid credentials", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		utils.ErrorResponse(c, "Invalid credentials", nil)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 hours
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		utils.ErrorResponse(c, "Failed to generate token", err)
		return
	}

	utils.SuccessResponse(c, "Login successful", gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Register handler
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, "Invalid request", map[string]string{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, "Validation failed", map[string]string{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, "User already exists", nil)
		return
	}

	// Create new user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		utils.ErrorResponse(c, "Failed to hash password", err)
		return
	}

	// Save to database
	if err := db.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, "Failed to create user", err)
		return
	}

	utils.SuccessResponse(c, "User created successfully", gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}
