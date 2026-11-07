# StockFlow

Mini ERP module: manage products, adjust stock, and record sales orders that consume stock.

- **Backend:** Go + Gin + GORM + PostgreSQL
- **Frontend:** React + TypeScript (Vite), single API client, plain CSS

## Prerequisites

- Docker + Docker Compose
- Go 1.25+ (backend dev)
- Node.js 20+ (frontend dev)

## Quick start (Docker)

```bash
docker compose up -d --build
curl localhost:8080/healthz
```

Expected: `{"status":"ok"}`.

Reset to a clean database:

```bash
docker compose down -v
docker compose up -d --build
```

## Local development (without the backend container)

```bash
# 1) database
docker compose up -d db

# 2) backend
cd backend
go run ./cmd/api            # listens on :8080

# 3) frontend
cd frontend
npm install
npm run dev                 # http://localhost:5173
```

### Configuration

| Env | Default | Used by | Description |
|---|---|---|---|
| `DATABASE_URL` | `postgres://stockflow:stockflow@localhost:5432/stockflow?sslmode=disable` | backend | Postgres DSN |
| `PORT` | `8080` | backend | HTTP port inside the process/container |
| `BACKEND_PORT` | `8080` | compose | Host port mapped to the backend |
| `DB_PORT` | `5432` | compose | Host port mapped to Postgres |
| `TEST_DATABASE_URL` / `TEST_DATABASE_NAME` / `TEST_DATABASE_PORT` | `stockflow_test_*` / `5432` | backend tests | Separate databases for tests |

## API reference

All responses are JSON. Errors always use one envelope:

```json
{ "error": { "code": "SOMETHING", "message": "human readable", "...": "extra fields" } }
```

Set a base URL for the examples:

```bash
BASE=http://localhost:8080
```

### 1. List products

```bash
curl -s "$BASE/api/products?search=aqua&page=1&limit=5"
```

```json
{
  "data": [
    {
      "id": 1,
      "sku": "AQUA-600",
      "name": "Aqua Air Mineral 600ml",
      "price": 4000,
      "stock": 110,
      "is_active": true,
      "created_at": "2026-11-07T00:32:39.234151Z",
      "last_movement_at": "2026-11-07T00:32:39.234151Z"
    }
  ],
  "page": 1,
  "limit": 5,
  "total": 1
}
```

### 2. Create product

```bash
curl -s -X POST "$BASE/api/products" \
  -H 'Content-Type: application/json' \
  -d '{"sku":"TEST-001","name":"Test Product","price":15000}'
```

Errors:

```json
{ "error": { "code": "SKU_ALREADY_EXISTS", "message": "sku already exists" } }        // 409
{ "error": { "code": "VALIDATION_ERROR", "message": "Name is required, Price must be >= 0" } } // 400
```

### 3. Update product

```bash
curl -s -X PUT "$BASE/api/products/16" \
  -H 'Content-Type: application/json' \
  -d '{"sku":"TEST-001","name":"Test Product v2","price":16000,"is_active":false}'
```

### 4. Delete product

```bash
curl -i -X DELETE "$BASE/api/products/16"
# 204 No Content

curl -i -X DELETE "$BASE/api/products/1"
```

```json
{ "error": { "code": "PRODUCT_HAS_HISTORY", "message": "product has stock history" } } // 409
```

### 5. Adjust stock

```bash
curl -s -X POST "$BASE/api/products/1/stock" \
  -H 'Content-Type: application/json' \
  -d '{"type":"in","qty":12,"note":"restock"}'
```

Returns the updated product. Overdrawing returns `409` with the current stock:

```bash
curl -i -X POST "$BASE/api/products/1/stock" \
  -H 'Content-Type: application/json' \
  -d '{"type":"out","qty":999999}'
```

```json
{ "error": { "code": "INSUFFICIENT_STOCK", "message": "insufficient stock", "available": 122 } }
```

### 6. Movement history

```bash
curl -s "$BASE/api/products/1/movements"
```

```json
{
  "data": [
    { "id": 15, "type": "out", "qty": 10, "note": "order #1", "created_at": "..." },
    { "id": 1, "type": "in", "qty": 120, "note": "initial stock", "created_at": "..." }
  ]
}
```

### 7. Create order

```bash
curl -s -X POST "$BASE/api/orders" \
  -H 'Content-Type: application/json' \
  -d '{
    "customer": "Acme Retail",
    "items": [
      { "sku": "AQUA-600", "qty": 2 },
      { "sku": "INDOMIE-GRG", "qty": 3 }
    ]
  }'
```

```json
{
  "id": 3,
  "customer": "Acme Retail",
  "status": "CREATED",
  "total": 18500,
  "created_at": "...",
  "cancelled_at": null,
  "items": [
    { "id": 5, "sku": "AQUA-600", "qty": 2, "unit_price": 4000 },
    { "id": 6, "sku": "INDOMIE-GRG", "qty": 3, "unit_price": 3500 }
  ]
}
```

Errors:

```json
{ "error": { "code": "INSUFFICIENT_STOCK", "message": "insufficient stock", "sku": "KPI-KPA-250G", "available": 6 } } // 409
{ "error": { "code": "UNKNOWN_SKU", "message": "unknown sku: NOPE-1" } }        // 400
{ "error": { "code": "DUPLICATE_SKU", "message": "duplicate sku in one order: AQUA-600" } } // 400
```

### 8. Order detail

```bash
curl -s "$BASE/api/orders/1"
```

`404 ORDER_NOT_FOUND` for unknown ids.

### 9. Cancel order

```bash
curl -s -X POST "$BASE/api/orders/1/cancel"
```

Returns the order with `status: "CANCELLED"`, `cancelled_at` set, and stock restored for every item (one `in` movement per item, note `cancel order #<id>`).

```json
{ "error": { "code": "ALREADY_CANCELLED", "message": "order already cancelled" } } // 409
```
## Design decisions

1. **Gin + GORM, migrations via `AutoMigrate` + Go seeder instead of SQL files.**

2. **`last_movement_at` is derived, not stored.**

3. **Stock changes are guarded inside `UPDATE`.**
   `UPDATE products SET stock = stock - qty WHERE id = ? AND stock >= qty`

4. **Order creation locks rows in a deterministic order.**
   `SELECT ... FOR UPDATE` for all SKUs `ORDER BY sku`, then a guarded `UPDATE` per item, all inside one transaction. The lock makes the stock check consistent

5. **Cancel is idempotent by status guard.**
   `UPDATE ... WHERE id = ? AND status = 'CREATED'` decides whether the restoration runs. Two concurrent cancels produce one `200` and one `409 ALREADY_CANCELLED`, and stock is restored exactly once. Items are locked in the same SKU order as creation.

6. **Money is integer IDR.**
   `price`, `unit_price`, and `total` are `integer` — no floats anywhere. `total` is always computed server-side from current product prices.

7. **One error envelope with stable codes.**
   `{"error":{"code","message",...}}`. The frontend switches on `code` (`INSUFFICIENT_STOCK`, `PRODUCT_HAS_HISTORY`, ...), never on message text. Extra fields (`available`, `sku`) are carried in the same object.

8. **The `sales_order_items.sku` foreign key is created manually.**
   GORM infers the wrong direction when the foreign key and the referenced column share a name (`sku`), generating `products.sku REFERENCES sales_order_items.sku`. A small idempotent `DO $$ ... $$` block in `Migrate` adds the correct constraint. It is `ON UPDATE CASCADE`, so renaming a product SKU keeps history attached.

9. **One API client module, typed from the server contract.**
    `src/api/client.ts` owns URL building, JSON parsing, the error envelope, and `AbortSignal`. Domain types live in `src/types/api.ts`, components never build URLs. Reads accept an `AbortSignal` (results can go stale), mutations deliberately do not, because aborting a POST does not undo it on the server.

10. **Search is server-side, filters are client-side.**
    `search` goes to `?search=`, debounced 300 ms. `All/Active/Inactive/Low stock` filter the fetched page in the browser because the API contract has no parameter for them. Known limitation: with more than 100 products, the low-stock filter only sees the first page.

11. **Standard ports, overridable.**
    Defaults are `8080` (API) and `5432` (Postgres).

## What I would improve next

- **Pagination in the UI.** The API already supports `page`/`limit`, the table currently requests the first 100 rows and filters client-side.
- **Server-side low-stock filter** (for example `?low_stock=true`) so the filter does not depend on the page size.
- **Idempotency-Key on `POST /api/orders`**, so a retried request cannot create a duplicate order.
- **Order history per customer / daily sales report** (`GET /api/reports/daily-sales?date=`).
- **CI pipeline** running `go vet`, `go test -race`, `tsc -b`, `oxlint`, and `vitest`.
- **Use OpenSearch** Since the user will search any SKU in bunch of data, OpenSearch is right choice to execute this.
- **End-to-end test** (Playwright) for the create-order → cancel → stock-restored flow.
- **Auth/roles** — out of scope for this repo, every endpoint is currently open.

## Time spent

**Hours spent:** ~10 Hours
