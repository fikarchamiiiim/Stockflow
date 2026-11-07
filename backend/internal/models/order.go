package models

import "time"

type SalesOrder struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	Customer    string           `gorm:"type:text;not null" json:"customer"`
	Status      string           `gorm:"type:text;not null;check:order_status_valid,status = 'CREATED' OR status = 'CANCELLED'" json:"status"`
	Total       int              `gorm:"type:integer;not null" json:"total"`
	CreatedAt   time.Time        `gorm:"not null;default:now()" json:"created_at"`
	CancelledAt *time.Time       `json:"cancelled_at"`
	Items       []SalesOrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type SalesOrderItem struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	OrderID   uint   `gorm:"not null;index" json:"order_id"`
	SKU       string `gorm:"type:text;not null;index" json:"sku"`
	Qty       int    `gorm:"type:integer;not null;check:item_qty_positive,qty > 0" json:"qty"`
	UnitPrice int    `gorm:"type:integer;not null" json:"unit_price"`
}
