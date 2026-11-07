package orders_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"stockflow/backend/internal/httpx"
	"stockflow/backend/internal/models"
	"stockflow/backend/internal/orders"
	"stockflow/backend/internal/testutil"
)

func TestMain(m *testing.M) {
	os.Setenv("TEST_DATABASE_NAME", "stockflow_test_orders")
	os.Exit(m.Run())
}

func TestCreateOrderInsufficientWritesNothing(t *testing.T) {
	db := testutil.DB(t)
	svc := orders.NewService(db)
	plenty := testutil.CreateProduct(t, db, "SKU-PLENTY", 1000, 10)
	scarce := testutil.CreateProduct(t, db, "SKU-SCARCE", 2000, 1)

	_, err := svc.Create(context.Background(), orders.CreateRequest{
		Customer: "Test Store",
		Items: []orders.ItemInput{
			{SKU: plenty.SKU, Qty: 2},
			{SKU: scarce.SKU, Qty: 5},
		},
	})

	var apiErr *httpx.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got: %v", err)
	}
	if apiErr.Status != 409 || apiErr.Code != "INSUFFICIENT_STOCK" {
		t.Fatalf("unexpected error: status=%d code=%s", apiErr.Status, apiErr.Code)
	}
	if apiErr.Extra["sku"] != scarce.SKU || apiErr.Extra["available"] != 1 {
		t.Fatalf("unexpected 409 details: %+v", apiErr.Extra)
	}

	testutil.AssertCount(t, db, &models.SalesOrder{}, 0)
	testutil.AssertCount(t, db, &models.SalesOrderItem{}, 0)
	testutil.AssertCount(t, db, &models.StockMovement{}, 0)
	testutil.AssertStock(t, db, plenty.ID, 10)
	testutil.AssertStock(t, db, scarce.ID, 1)
}

func TestCreateOrderHappyPath(t *testing.T) {
	db := testutil.DB(t)
	svc := orders.NewService(db)
	a := testutil.CreateProduct(t, db, "SKU-A", 1000, 10)
	b := testutil.CreateProduct(t, db, "SKU-B", 500, 10)

	res, err := svc.Create(context.Background(), orders.CreateRequest{
		Customer: "Test Store",
		Items: []orders.ItemInput{
			{SKU: a.SKU, Qty: 2},
			{SKU: b.SKU, Qty: 3},
		},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	if res.Total != 3500 {
		t.Fatalf("total = %d, want 3500", res.Total)
	}
	if res.Status != "CREATED" {
		t.Fatalf("status = %s, want CREATED", res.Status)
	}
	if len(res.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(res.Items))
	}
	if res.Items[0].UnitPrice != 1000 || res.Items[1].UnitPrice != 500 {
		t.Fatalf("unit price not from database: %+v", res.Items)
	}

	testutil.AssertStock(t, db, a.ID, 8)
	testutil.AssertStock(t, db, b.ID, 7)
	testutil.AssertCount(t, db, &models.SalesOrder{}, 1)
	testutil.AssertCount(t, db, &models.SalesOrderItem{}, 2)
	testutil.AssertCount(t, db, &models.StockMovement{}, 2)

	var movements []models.StockMovement
	if err := db.Order("id").Find(&movements).Error; err != nil {
		t.Fatalf("read movements: %v", err)
	}
	for _, m := range movements {
		if m.Type != "out" {
			t.Fatalf("movement type = %s, want out", m.Type)
		}
		if m.Note == nil || *m.Note == "" {
			t.Fatal("order movement must have a note")
		}
	}
}

func TestCreateOrderConcurrent(t *testing.T) {
	db := testutil.DB(t)
	svc := orders.NewService(db)
	p := testutil.CreateProduct(t, db, "SKU-LIMIT", 1000, 10)

	const (
		attempts = 5
		qty      = 3
	)
	var success, conflict int32
	var wg sync.WaitGroup

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := svc.Create(context.Background(), orders.CreateRequest{
				Customer: "Parallel Store",
				Items:    []orders.ItemInput{{SKU: p.SKU, Qty: qty}},
			})

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

	if success != 3 || conflict != 2 {
		t.Fatalf("success=%d conflict=%d, want 3/2", success, conflict)
	}

	testutil.AssertStock(t, db, p.ID, 10-int(success)*qty)
	testutil.AssertCount(t, db, &models.SalesOrder{}, int64(success))
	testutil.AssertCount(t, db, &models.StockMovement{}, int64(success))
}

func TestCancelRestoresStockOnce(t *testing.T) {
	db := testutil.DB(t)
	svc := orders.NewService(db)
	a := testutil.CreateProduct(t, db, "SKU-A", 4000, 10)
	b := testutil.CreateProduct(t, db, "SKU-B", 3000, 10)

	order, err := svc.Create(context.Background(), orders.CreateRequest{
		Customer: "Test Store",
		Items: []orders.ItemInput{
			{SKU: a.SKU, Qty: 3},
			{SKU: b.SKU, Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	testutil.AssertStock(t, db, a.ID, 7)
	testutil.AssertStock(t, db, b.ID, 8)

	cancelled, err := svc.Cancel(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancelled.Status != "CANCELLED" || cancelled.CancelledAt == nil {
		t.Fatalf("unexpected status after cancel: %+v", cancelled)
	}

	testutil.AssertStock(t, db, a.ID, 10)
	testutil.AssertStock(t, db, b.ID, 10)

	_, err = svc.Cancel(context.Background(), order.ID)
	var apiErr *httpx.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "ALREADY_CANCELLED" {
		t.Fatalf("second cancel should be ALREADY_CANCELLED, got: %v", err)
	}

	testutil.AssertStock(t, db, a.ID, 10)
	testutil.AssertStock(t, db, b.ID, 10)
	testutil.AssertCount(t, db, &models.StockMovement{}, 4)
}
