package database

import (
	"POS-Golang/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {
	var err error
	// Menggunakan SQLite untuk kemudahan belajar
	DB, err = gorm.Open(mysql.Open("pos.db"), &gorm.Config{})
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
