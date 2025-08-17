package handlers

import (
	"POS-Golang/internal/database"
	"POS-Golang/internal/models"
	"POS-Golang/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardStats struct {
	TotalProducts     int64   `json:"total_products"`
	TotalTransactions int64   `json:"total_transactions"`
	TodayTransactions int64   `json:"today_transactions"`
	TodayRevenue      float64 `json:"today_revenue"`
	MonthlyRevenue    float64 `json:"monthly_revenue"`
	LowStockProducts  int64   `json:"low_stock_products"`
}

type TopProduct struct {
	ProductName string  `json:"product_name"`
	TotalSold   int     `json:"total_sold"`
	Revenue     float64 `json:"revenue"`
}

type RevenueData struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
}

// Get dashboard data
func GetDashboard(c *gin.Context) {
	db := database.GetDB()
	now := time.Now()
	today := now.Format("2006-01-02")
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	var stats DashboardStats

	// Total products
	db.Model(&models.Product{}).Count(&stats.TotalProducts)

	// Total transactions
	db.Model(&models.Transaction{}).Where("status = ?", "completed").Count(&stats.TotalTransactions)

	// Today's transactions
	db.Model(&models.Transaction{}).Where("DATE(created_at) = ? AND status = ?", today, "completed").Count(&stats.TodayTransactions)

	// Today's revenue
	db.Model(&models.Transaction{}).Where("DATE(created_at) = ? AND status = ?", today, "completed").Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.TodayRevenue)

	// Monthly revenue
	db.Model(&models.Transaction{}).Where("DATE(created_at) >= ? AND status = ?", monthStart, "completed").Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.MonthlyRevenue)

	// Low stock products (stock <= 10)
	db.Model(&models.Product{}).Where("stock <= ? AND is_active = ?", 10, true).Count(&stats.LowStockProducts)

	// Top products this month
	var topProducts []TopProduct
	db.Table("transaction_items").
		Select("products.name as product_name, SUM(transaction_items.quantity) as total_sold, SUM(transaction_items.subtotal) as revenue").
		Joins("JOIN products ON transaction_items.product_id = products.id").
		Joins("JOIN transactions ON transaction_items.transaction_id = transactions.id").
		Where("DATE(transactions.created_at) >= ? AND transactions.status = ?", monthStart, "completed").
		Group("products.id, products.name").
		Order("total_sold DESC").
		Limit(5).
		Scan(&topProducts)

	// Revenue data for the last 7 days
	var revenueData []RevenueData
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		var revenue float64
		db.Model(&models.Transaction{}).
			Where("DATE(created_at) = ? AND status = ?", date, "completed").
			Select("COALESCE(SUM(total_amount), 0)").
			Scan(&revenue)

		revenueData = append(revenueData, RevenueData{
			Date:    date,
			Revenue: revenue,
		})
	}

	// Recent transactions
	var recentTransactions []models.Transaction
	db.Preload("User").Preload("Items.Product").
		Where("status = ?", "completed").
		Order("created_at DESC").
		Limit(5).
		Find(&recentTransactions)

	utils.SuccessResponse(c, "Dashboard data fetched successfully", gin.H{
		"stats":               stats,
		"top_products":        topProducts,
		"revenue_data":        revenueData,
		"recent_transactions": recentTransactions,
	})
}
