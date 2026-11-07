package models

import "time"

type StockMovement struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"not null;index:idx_movement_product_created,priority:1" json:"product_id"`
	Type      string    `gorm:"type:text;not null;index;check:movement_type_valid,type = 'in' OR type = 'out'" json:"type"`
	Qty       int       `gorm:"type:integer;not null;check:movement_qty_positive,qty > 0" json:"qty"`
	Note      *string   `json:"note"`
	CreatedAt time.Time `gorm:"not null;default:now();index:idx_movement_product_created,priority:2,sort:desc" json:"created_at"`
}
