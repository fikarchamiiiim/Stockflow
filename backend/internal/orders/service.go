package orders

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"stockflow/backend/internal/httpx"
	"stockflow/backend/internal/models"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, in CreateRequest) (*OrderResponse, error) {
	skus := make([]string, 0, len(in.Items))
	seen := make(map[string]struct{}, len(in.Items))
	for _, it := range in.Items {
		if _, dup := seen[it.SKU]; dup {
			return nil, badRequest("DUPLICATE_SKU", "duplicate sku in one order: "+it.SKU)
		}
		seen[it.SKU] = struct{}{}
		skus = append(skus, it.SKU)
	}

	var order models.SalesOrder

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var products []models.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("sku IN ?", skus).
			Order("sku").
			Find(&products).Error; err != nil {
			return err
		}
		if len(products) != len(skus) {
			return badRequest("UNKNOWN_SKU", "unknown sku: "+missingSKUs(skus, products))
		}

		bySKU := make(map[string]models.Product, len(products))
		for _, p := range products {
			bySKU[p.SKU] = p
		}

		total := 0
		for _, it := range in.Items {
			p := bySKU[it.SKU]
			if p.Stock < it.Qty {
				return insufficientStock(it.SKU, p.Stock)
			}
			total += p.Price * it.Qty
		}

		order = models.SalesOrder{
			Customer: in.Customer,
			Status:   "CREATED",
			Total:    total,
		}
		for _, it := range in.Items {
			order.Items = append(order.Items, models.SalesOrderItem{
				SKU:       it.SKU,
				Qty:       it.Qty,
				UnitPrice: bySKU[it.SKU].Price,
			})
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		for _, it := range order.Items {
			productID := bySKU[it.SKU].ID

			res := tx.Model(&models.Product{}).
				Where("id = ? AND stock >= ?", productID, it.Qty).
				Update("stock", gorm.Expr("stock - ?", it.Qty))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				var current models.Product
				if err := tx.First(&current, productID).Error; err != nil {
					return err
				}
				return insufficientStock(it.SKU, current.Stock)
			}

			if err := tx.Create(&models.StockMovement{
				ProductID: productID,
				Type:      "out",
				Qty:       it.Qty,
				Note:      strPtr(fmt.Sprintf("order #%d", order.ID)),
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	res := toResponse(order)
	return &res, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*OrderResponse, error) {
	var order models.SalesOrder
	if err := s.db.WithContext(ctx).Preload("Items").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, orderNotFound()
		}
		return nil, err
	}

	res := toResponse(order)
	return &res, nil
}

func (s *Service) Cancel(ctx context.Context, id uint) (*OrderResponse, error) {
	var order models.SalesOrder

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.SalesOrder{}).
			Where("id = ? AND status = ?", id, "CREATED").
			Updates(map[string]any{"status": "CANCELLED", "cancelled_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			var existing models.SalesOrder
			if err := tx.First(&existing, id).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return orderNotFound()
				}
				return err
			}
			return alreadyCancelled()
		}

		var items []models.SalesOrderItem
		if err := tx.Where("order_id = ?", id).Order("sku").Find(&items).Error; err != nil {
			return err
		}

		for _, it := range items {
			var p models.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("sku = ?", it.SKU).
				First(&p).Error; err != nil {
				return err
			}

			if err := tx.Model(&models.Product{}).
				Where("id = ?", p.ID).
				Update("stock", gorm.Expr("stock + ?", it.Qty)).Error; err != nil {
				return err
			}

			if err := tx.Create(&models.StockMovement{
				ProductID: p.ID,
				Type:      "in",
				Qty:       it.Qty,
				Note:      strPtr(fmt.Sprintf("cancel order #%d", id)),
			}).Error; err != nil {
				return err
			}
		}

		return tx.Preload("Items").First(&order, id).Error
	})
	if err != nil {
		return nil, err
	}

	res := toResponse(order)
	return &res, nil
}

func toResponse(o models.SalesOrder) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, OrderItemResponse{
			ID:        it.ID,
			SKU:       it.SKU,
			Qty:       it.Qty,
			UnitPrice: it.UnitPrice,
		})
	}

	return OrderResponse{
		ID:          o.ID,
		Customer:    o.Customer,
		Status:      o.Status,
		Total:       o.Total,
		CreatedAt:   o.CreatedAt,
		CancelledAt: o.CancelledAt,
		Items:       items,
	}
}

func missingSKUs(skus []string, products []models.Product) string {
	found := make(map[string]struct{}, len(products))
	for _, p := range products {
		found[p.SKU] = struct{}{}
	}

	missing := make([]string, 0, len(skus))
	for _, sku := range skus {
		if _, ok := found[sku]; !ok {
			missing = append(missing, sku)
		}
	}
	return strings.Join(missing, ", ")
}

func badRequest(code, message string) error {
	return &httpx.APIError{
		Status:  400,
		Code:    code,
		Message: message,
	}
}

func insufficientStock(sku string, available int) error {
	return &httpx.APIError{
		Status:  409,
		Code:    "INSUFFICIENT_STOCK",
		Message: "insufficient stock",
		Extra:   map[string]any{"sku": sku, "available": available},
	}
}

func orderNotFound() error {
	return &httpx.APIError{
		Status:  404,
		Code:    "ORDER_NOT_FOUND",
		Message: "order not found",
	}
}

func alreadyCancelled() error {
	return &httpx.APIError{
		Status:  409,
		Code:    "ALREADY_CANCELLED",
		Message: "order already cancelled",
	}
}

func strPtr(s string) *string {
	return &s
}
