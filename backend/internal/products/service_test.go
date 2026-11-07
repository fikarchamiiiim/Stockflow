package products_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"stockflow/backend/internal/httpx"
	"stockflow/backend/internal/models"
	"stockflow/backend/internal/products"
	"stockflow/backend/internal/testutil"
)

func TestMain(m *testing.M) {
	os.Setenv("TEST_DATABASE_NAME", "stockflow_test_products")
	os.Exit(m.Run())
}

func TestAdjustStockInsufficientWritesNothing(t *testing.T) {
	db := testutil.DB(t)
	svc := products.NewService(db)
	p := testutil.CreateProduct(t, db, "SKU-OUT", 1000, 5)

	_, err := svc.AdjustStock(context.Background(), p.ID, products.StockRequest{Type: "out", Qty: 6})

	var apiErr *httpx.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got: %v", err)
	}
	if apiErr.Status != 409 || apiErr.Code != "INSUFFICIENT_STOCK" {
		t.Fatalf("unexpected error: status=%d code=%s", apiErr.Status, apiErr.Code)
	}
	if apiErr.Extra["available"] != 5 {
		t.Fatalf("available = %v, want 5", apiErr.Extra["available"])
	}

	testutil.AssertStock(t, db, p.ID, 5)
	testutil.AssertCount(t, db, &models.StockMovement{}, 0)
}

func TestAdjustStockOutHappyPath(t *testing.T) {
	db := testutil.DB(t)
	svc := products.NewService(db)
	p := testutil.CreateProduct(t, db, "SKU-OK", 1000, 10)

	res, err := svc.AdjustStock(context.Background(), p.ID, products.StockRequest{Type: "out", Qty: 3, Note: "test"})
	if err != nil {
		t.Fatalf("adjust stock: %v", err)
	}
	if res.Stock != 7 {
		t.Fatalf("response stock = %d, want 7", res.Stock)
	}

	testutil.AssertStock(t, db, p.ID, 7)
	testutil.AssertCount(t, db, &models.StockMovement{}, 1)

	var m models.StockMovement
	if err := db.Where("product_id = ?", p.ID).First(&m).Error; err != nil {
		t.Fatalf("read movement: %v", err)
	}
	if m.Type != "out" || m.Qty != 3 {
		t.Fatalf("movement = %s/%d, want out/3", m.Type, m.Qty)
	}
}

func TestAdjustStockConcurrentOut(t *testing.T) {
	db := testutil.DB(t)
	svc := products.NewService(db)
	p := testutil.CreateProduct(t, db, "SKU-CONC", 1000, 20)

	const attempts = 40
	var success, conflict int32
	var wg sync.WaitGroup

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := svc.AdjustStock(context.Background(), p.ID, products.StockRequest{Type: "out", Qty: 1})

			var apiErr *httpx.APIError
			switch {
			case err == nil:
				atomic.AddInt32(&success, 1)
			case errors.As(err, &apiErr) && apiErr.Code == "INSUFFICIENT_STOCK":
				atomic.AddInt32(&conflict, 1)
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if success != 20 || conflict != 20 {
		t.Fatalf("success=%d conflict=%d, want 20/20", success, conflict)
	}

	testutil.AssertStock(t, db, p.ID, 0)
	testutil.AssertCount(t, db, &models.StockMovement{}, 20)
}
