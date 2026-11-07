package testutil

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"stockflow/backend/internal/database"
	"stockflow/backend/internal/models"
)

const defaultBaseDSN = "postgres://stockflow:stockflow@localhost:5432"

func DB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := testDSN()

	db, err := database.Connect(dsn)
	if err != nil {
		if createErr := createDatabase(dsn); createErr != nil {
			t.Fatalf("cannot connect to test database: %v (create database: %v)", err, createErr)
		}
		db, err = database.Connect(dsn)
		if err != nil {
			t.Fatalf("cannot connect to test database: %v", err)
		}
	}

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	Reset(t, db)
	t.Cleanup(func() { Reset(t, db) })

	return db
}

func Reset(t *testing.T, db *gorm.DB) {
	t.Helper()

	err := db.Exec("TRUNCATE sales_order_items, sales_orders, stock_movements, products RESTART IDENTITY CASCADE").Error
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func CreateProduct(t *testing.T, db *gorm.DB, sku string, price, stock int) models.Product {
	t.Helper()

	p := models.Product{
		SKU:      sku,
		Name:     "Product " + sku,
		Price:    price,
		Stock:    stock,
		IsActive: true,
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	return p
}

func AssertStock(t *testing.T, db *gorm.DB, productID uint, want int) {
	t.Helper()

	var p models.Product
	if err := db.First(&p, productID).Error; err != nil {
		t.Fatalf("read product %d: %v", productID, err)
	}
	if p.Stock != want {
		t.Fatalf("stock of product %d = %d, want %d", productID, p.Stock, want)
	}
}

func AssertCount(t *testing.T, db *gorm.DB, model any, want int64) {
	t.Helper()

	var got int64
	if err := db.Model(model).Count(&got).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}

func testDSN() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}

	name := os.Getenv("TEST_DATABASE_NAME")
	if name == "" {
		name = "stockflow_test"
	}

	base := defaultBaseDSN
	if port := os.Getenv("TEST_DATABASE_PORT"); port != "" {
		if u, err := url.Parse(defaultBaseDSN); err == nil {
			u.Host = net.JoinHostPort(u.Hostname(), port)
			base = u.String()
		}
	}

	return base + "/" + name + "?sslmode=disable"
}

func createDatabase(dsn string) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return err
	}
	name := strings.TrimPrefix(u.Path, "/")
	if name == "" {
		return errors.New("database name missing from DSN")
	}

	admin := *u
	admin.Path = "/postgres"

	db, err := gorm.Open(postgres.Open(admin.String()), &gorm.Config{})
	if err != nil {
		return err
	}
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}

	return db.Exec(fmt.Sprintf(`CREATE DATABASE %q`, name)).Error
}
