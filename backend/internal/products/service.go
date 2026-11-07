package products

import (
	"context"
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"

	"stockflow/backend/internal/httpx"
	"stockflow/backend/internal/models"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type listRow struct {
	models.Product
	LastMovementAt *time.Time
}

func (s *Service) List(ctx context.Context, q ListQuery) (*ListResponse, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	filter := func(db *gorm.DB) *gorm.DB {
		if q.Search == "" {
			return db
		}
		like := "%" + q.Search + "%"
		return db.Where("products.sku ILIKE ? OR products.name ILIKE ?", like, like)
	}

	var total int64
	if err := filter(s.db.WithContext(ctx).Model(&models.Product{})).Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []listRow
	if err := filter(s.db.WithContext(ctx).Model(&models.Product{})).
		Select("products.*, (SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = products.id) AS last_movement_at").
		Order("products.id").
		Limit(limit).
		Offset((page - 1) * limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	data := make([]ProductResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toResponse(row.Product, row.LastMovementAt))
	}

	return &ListResponse{Data: data, Page: page, Limit: limit, Total: total}, nil
}

func (s *Service) Create(ctx context.Context, in CreateRequest) (*ProductResponse, error) {
	p := models.Product{
		SKU:      in.SKU,
		Name:     in.Name,
		Price:    *in.Price,
		IsActive: true,
	}

	if err := s.db.WithContext(ctx).Create(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, skuTaken()
		}
		return nil, err
	}

	res := toResponse(p, nil)
	return &res, nil
}

func (s *Service) Update(ctx context.Context, id uint, in UpdateRequest) (*ProductResponse, error) {
	res := s.db.WithContext(ctx).Model(&models.Product{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"sku":       in.SKU,
			"name":      in.Name,
			"price":     *in.Price,
			"is_active": *in.IsActive,
		})
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
			return nil, skuTaken()
		}
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, productNotFound()
	}

	return s.findByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).
		Where("id = ? AND NOT EXISTS (SELECT 1 FROM stock_movements m WHERE m.product_id = products.id)", id).
		Delete(&models.Product{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return productNotFound()
	}

	return &httpx.APIError{
		Status:  http.StatusConflict,
		Code:    "PRODUCT_HAS_HISTORY",
		Message: "product has stock history",
	}
}

func (s *Service) AdjustStock(ctx context.Context, id uint, in StockRequest) (*ProductResponse, error) {
	var (
		product  models.Product
		movement models.StockMovement
	)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if in.Type == "out" {
			res := tx.Model(&models.Product{}).
				Where("id = ? AND stock >= ?", id, in.Qty).
				Update("stock", gorm.Expr("stock - ?", in.Qty))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				var current models.Product
				if err := tx.First(&current, id).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return productNotFound()
					}
					return err
				}
				return insufficientStock(current.Stock)
			}
		} else {
			res := tx.Model(&models.Product{}).
				Where("id = ?", id).
				Update("stock", gorm.Expr("stock + ?", in.Qty))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return productNotFound()
			}
		}

		movement = models.StockMovement{
			ProductID: id,
			Type:      in.Type,
			Qty:       in.Qty,
			Note:      optionalString(in.Note),
		}
		if err := tx.Create(&movement).Error; err != nil {
			return err
		}

		return tx.First(&product, id).Error
	})
	if err != nil {
		return nil, err
	}

	res := toResponse(product, &movement.CreatedAt)
	return &res, nil
}

func (s *Service) ListMovements(ctx context.Context, id uint) (*MovementListResponse, error) {
	if _, err := s.findByID(ctx, id); err != nil {
		return nil, err
	}

	var rows []models.StockMovement
	if err := s.db.WithContext(ctx).
		Where("product_id = ?", id).
		Order("created_at DESC, id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	data := make([]MovementResponse, 0, len(rows))
	for _, m := range rows {
		data = append(data, MovementResponse{
			ID:        m.ID,
			Type:      m.Type,
			Qty:       m.Qty,
			Note:      m.Note,
			CreatedAt: m.CreatedAt,
		})
	}

	return &MovementListResponse{Data: data}, nil
}

func (s *Service) findByID(ctx context.Context, id uint) (*ProductResponse, error) {
	var row listRow
	if err := s.db.WithContext(ctx).Model(&models.Product{}).
		Select("products.*, (SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = products.id) AS last_movement_at").
		Where("products.id = ?", id).
		Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, productNotFound()
	}

	res := toResponse(row.Product, row.LastMovementAt)
	return &res, nil
}

func toResponse(p models.Product, lastMovementAt *time.Time) ProductResponse {
	return ProductResponse{
		ID:             p.ID,
		SKU:            p.SKU,
		Name:           p.Name,
		Price:          p.Price,
		Stock:          p.Stock,
		IsActive:       p.IsActive,
		CreatedAt:      p.CreatedAt,
		LastMovementAt: lastMovementAt,
	}
}

func skuTaken() error {
	return &httpx.APIError{
		Status:  http.StatusConflict,
		Code:    "SKU_ALREADY_EXISTS",
		Message: "sku already exists",
	}
}

func productNotFound() error {
	return &httpx.APIError{
		Status:  http.StatusNotFound,
		Code:    "PRODUCT_NOT_FOUND",
		Message: "product not found",
	}
}

func insufficientStock(available int) error {
	return &httpx.APIError{
		Status:  http.StatusConflict,
		Code:    "INSUFFICIENT_STOCK",
		Message: "insufficient stock",
		Extra:   map[string]any{"available": available},
	}
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
