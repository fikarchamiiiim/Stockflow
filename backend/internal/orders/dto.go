package orders

import "time"

type CreateRequest struct {
	Customer string      `json:"customer" binding:"required"`
	Items    []ItemInput `json:"items" binding:"required,min=1,dive"`
}

type ItemInput struct {
	SKU string `json:"sku" binding:"required"`
	Qty int    `json:"qty" binding:"required,gt=0"`
}

type OrderResponse struct {
	ID          uint                `json:"id"`
	Customer    string              `json:"customer"`
	Status      string              `json:"status"`
	Total       int                 `json:"total"`
	CreatedAt   time.Time           `json:"created_at"`
	CancelledAt *time.Time          `json:"cancelled_at"`
	Items       []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ID        uint   `json:"id"`
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	UnitPrice int    `json:"unit_price"`
}
