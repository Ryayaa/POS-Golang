package models

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	UserID        uint              `json:"user_id" gorm:"not null"`
	TransactionNo string            `json:"transaction_no" gorm:"unique;not null" validate:"required"`
	User          User              `json:"user,omitempty"`
	Items         []TransactionItem `json:"items,omitempty" gorm:"foreignKey:TransactionID"`
	TotalAmount   float64           `json:"total_amount" gorm:"not null" validate:"required,gt=0"`
	PaymentMethod string            `json:"payment_method" gorm:"not null" validate:"required,oneof=cash card transfer"`
	PaymentAmount float64           `json:"payment_amount" gorm:"not null" validate:"required,gt=0"`
	ChangeAmount  float64           `json:"change_amount" gorm:"not null" validate:"required,gte=0"`
	Status        string            `json:"status" gorm:"not null;default:'completed'" validate:"required,oneof=pending completed cancelled"`
	CreatedAt     time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt    `json:"deleted_at,omitempty" gorm:"index"`
}

type TransactionItem struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	TransactionID uint      `json:"transaction_id"`
	ProductID     uint      `json:"product_id"`
	Product       Product   `json:"product,omitempty"`
	Quantity      int       `json:"quantity" validate:"required,gt=0"`
	Price         float64   `json:"price"`
	Subtotal      float64   `json:"subtotal"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
