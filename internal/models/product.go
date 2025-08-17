package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null" validate:"required"`
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"not null" validate:"required,gt=0"`
	Stock       int            `json:"stock" gorm:"not null" validate:"required,gte=0"`
	CategoryID  uint           `json:"category_id"`
	Category    Category       `json:"category,omitempty"`
	Barcode     string         `json:"barcode" gorm:"unique"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type Category struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null" validate:"required"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
