package main

import (
	"POS-Golang/internal/config"
	"POS-Golang/internal/database"
	"POS-Golang/internal/handlers"
	"POS-Golang/internal/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Setup router
	r := gin.Default()

	// Enable CORS for frontend
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Auth routes (no middleware)
		api.POST("/login", handlers.Login)
		api.POST("/register", handlers.Register)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Product routes
			protected.GET("/products", handlers.GetProducts)
			protected.GET("/products/:id", handlers.GetProduct)
			protected.POST("/products", handlers.CreateProduct)
			protected.PUT("/products/:id", handlers.UpdateProduct)
			protected.DELETE("/products/:id", handlers.DeleteProduct)

			// Transaction routes
			protected.GET("/transactions", handlers.GetTransactions)
			protected.POST("/transactions", handlers.CreateTransaction)
			protected.GET("/transactions/:id", handlers.GetTransaction)

			// Dashboard
			protected.GET("/dashboard", handlers.GetDashboard)
		}

		// Admin only routes
		admin := protected.Group("")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/users", handlers.GetUsers)
			admin.POST("/users", handlers.CreateUser)
			admin.PUT("/users/:id", handlers.UpdateUser)
			admin.DELETE("/users/:id", handlers.DeleteUser)
		}
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "POS API is running",
		})
	})

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
