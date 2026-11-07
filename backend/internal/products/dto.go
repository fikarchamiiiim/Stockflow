package products

import "time"

type ListQuery struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type CreateRequest struct {
	SKU   string `json:"sku" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Price *int   `json:"price" binding:"required,gte=0"`
}

type UpdateRequest struct {
	SKU      string `json:"sku" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Price    *int   `json:"price" binding:"required,gte=0"`
	IsActive *bool  `json:"is_active" binding:"required"`
}

type StockRequest struct {
	Type string `json:"type" binding:"required,oneof=in out"`
	Qty  int    `json:"qty" binding:"required,gt=0"`
	Note string `json:"note"`
}

type ProductResponse struct {
	ID             uint       `json:"id"`
	SKU            string     `json:"sku"`
	Name           string     `json:"name"`
	Price          int        `json:"price"`
	Stock          int        `json:"stock"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	LastMovementAt *time.Time `json:"last_movement_at"`
}

type ListResponse struct {
	Data  []ProductResponse `json:"data"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
	Total int64             `json:"total"`
}

type MovementResponse struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	Qty       int       `json:"qty"`
	Note      *string   `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type MovementListResponse struct {
	Data []MovementResponse `json:"data"`
}
