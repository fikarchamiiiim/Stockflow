package models

import "time"

type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SKU       string    `gorm:"type:text;not null;uniqueIndex" json:"sku"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	Price     int       `gorm:"type:integer;not null;check:price_non_negative,price >= 0" json:"price"`
	Stock     int       `gorm:"type:integer;not null;check:stock_non_negative,stock >= 0" json:"stock"`
	IsActive  bool      `gorm:"not null" json:"is_active"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}
