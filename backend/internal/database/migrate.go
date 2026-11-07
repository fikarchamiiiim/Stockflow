package database

import (
	"stockflow/backend/internal/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Product{},
		&models.StockMovement{},
		&models.SalesOrder{},
		&models.SalesOrderItem{},
	); err != nil {
		return err
	}

	return db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'fk_sales_order_items_sku'
				  AND conrelid = 'sales_order_items'::regclass
			) THEN
				ALTER TABLE sales_order_items
					ADD CONSTRAINT fk_sales_order_items_sku
					FOREIGN KEY (sku) REFERENCES products (sku)
					ON UPDATE CASCADE ON DELETE RESTRICT;
			END IF;
		END
		$$;
	`).Error
}
