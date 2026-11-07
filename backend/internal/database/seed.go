package database

import (
	"fmt"

	"stockflow/backend/internal/models"

	"gorm.io/gorm"
)

type seedProduct struct {
	SKU    string
	Name   string
	Price  int
	Stock  int
	Active bool
}

type seedItem struct {
	SKU string
	Qty int
}

type seedOrder struct {
	Customer string
	Items    []seedItem
}

var seedProducts = []seedProduct{
	{SKU: "AQUA-600", Name: "Aqua Air Mineral 600ml", Price: 4000, Stock: 120, Active: true},
	{SKU: "INDOMIE-GRG", Name: "Indomie Mi Goreng 85g", Price: 3500, Stock: 200, Active: true},
	{SKU: "SG-DNL-1L", Name: "Sunlight Jeruk Nipis 1L", Price: 22000, Stock: 45, Active: true},
	{SKU: "TELUR-1KG", Name: "Telur Ayam Negeri 1kg", Price: 32000, Stock: 30, Active: true},
	{SKU: "BRS-BMP-5KG", Name: "Beras Pandan Wangi 5kg", Price: 78000, Stock: 18, Active: true},
	{SKU: "MNYK-SAN-2L", Name: "Minyak Goreng Sania 2L", Price: 38000, Stock: 8, Active: true},
	{SKU: "GLK-GLK-1KG", Name: "Gula Pasir Gulaku 1kg", Price: 17000, Stock: 25, Active: true},
	{SKU: "KPI-KPA-250G", Name: "Kopi Kapal Api Special 250g", Price: 26500, Stock: 6, Active: true},
	{SKU: "TEH-SWG-25", Name: "Teh Celup Sariwangi 25 kantong", Price: 12500, Stock: 40, Active: true},
	{SKU: "SBN-LFB-75G", Name: "Sabun Lifebuoy 75g", Price: 4500, Stock: 60, Active: true},
	{SKU: "PST-PPD-190G", Name: "Pasta Gigi Pepsodent 190g", Price: 15500, Stock: 9, Active: true},
	{SKU: "DTG-RNS-800ML", Name: "Deterjen Rinso 800ml", Price: 24000, Stock: 35, Active: true},
	{SKU: "SKM-IND-545G", Name: "Susu Kental Manis Indomilk 545g", Price: 14000, Stock: 22, Active: true},
	{SKU: "SAB-MLM-900ML", Name: "Sabun Cuci Piring Mama Lemon 900ml", Price: 19500, Stock: 12, Active: true},
	{SKU: "MIE-SDP-SOTO", Name: "Mie Sedaap Soto 70g", Price: 3000, Stock: 0, Active: false},
}

var seedOrders = []seedOrder{
	{
		Customer: "Toko Berkah Jaya",
		Items: []seedItem{
			{SKU: "AQUA-600", Qty: 10},
			{SKU: "INDOMIE-GRG", Qty: 24},
		},
	},
	{
		Customer: "Warung Bu Siti",
		Items: []seedItem{
			{SKU: "SG-DNL-1L", Qty: 5},
			{SKU: "BRS-BMP-5KG", Qty: 2},
		},
	},
}

func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Product{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		products := make(map[string]*models.Product, len(seedProducts))

		for _, sp := range seedProducts {
			p := &models.Product{
				SKU:      sp.SKU,
				Name:     sp.Name,
				Price:    sp.Price,
				Stock:    sp.Stock,
				IsActive: sp.Active,
			}
			if err := tx.Create(p).Error; err != nil {
				return err
			}
			products[p.SKU] = p

			if p.Stock == 0 {
				continue
			}
			if err := tx.Create(&models.StockMovement{
				ProductID: p.ID,
				Type:      "in",
				Qty:       p.Stock,
				Note:      strPtr("initial stock"),
			}).Error; err != nil {
				return err
			}
		}

		for _, so := range seedOrders {
			order := models.SalesOrder{Customer: so.Customer, Status: "CREATED"}
			for _, li := range so.Items {
				p, ok := products[li.SKU]
				if !ok {
					return fmt.Errorf("seed: unknown sku %s", li.SKU)
				}
				order.Total += p.Price * li.Qty
			}

			if err := tx.Create(&order).Error; err != nil {
				return err
			}

			for _, li := range so.Items {
				p := products[li.SKU]

				if err := tx.Create(&models.SalesOrderItem{
					OrderID:   order.ID,
					SKU:       li.SKU,
					Qty:       li.Qty,
					UnitPrice: p.Price,
				}).Error; err != nil {
					return err
				}

				res := tx.Model(&models.Product{}).
					Where("id = ? AND stock >= ?", p.ID, li.Qty).
					Update("stock", gorm.Expr("stock - ?", li.Qty))
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					return fmt.Errorf("seed: insufficient stock for %s", li.SKU)
				}

				if err := tx.Create(&models.StockMovement{
					ProductID: p.ID,
					Type:      "out",
					Qty:       li.Qty,
					Note:      strPtr(fmt.Sprintf("order #%d", order.ID)),
				}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func strPtr(s string) *string {
	return &s
}
