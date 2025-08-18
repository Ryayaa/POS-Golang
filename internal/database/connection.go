package database

import (
	"POS-Golang/internal/models"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {
	var err error

	// Gunakan environment variable untuk database connection
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Default untuk development - sesuaikan dengan kredensial MySQL Anda
		// Format: username:password@tcp(host:port)/database_name?charset=utf8mb4&parseTime=True&loc=Local
		dsn = "root:@tcp(localhost:3306)/pos_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// Auto migrate models
	err = DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Transaction{},
		&models.TransactionItem{},
	)

	return err
}

func GetDB() *gorm.DB {
	return DB
}
