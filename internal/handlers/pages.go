package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Index page
func IndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/index.html", gin.H{
		"title": "POS System",
	})
}

// Login page
func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/login.html", gin.H{
		"title": "Login - POS System",
	})
}

// Dashboard page
func DashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/dashboard.html", gin.H{
		"title": "Dashboard - POS System",
	})
}

// Products page
func ProductsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/products.html", gin.H{
		"title": "Products - POS System",
	})
}

// Transactions page
func TransactionsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/transactions.html", gin.H{
		"title": "Transactions - POS System",
	})
}
