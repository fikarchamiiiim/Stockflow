import { newDraftItem } from '../lib/orderDraft'
import type { DraftItem } from '../lib/orderDraft'
import type { Product } from '../types/api'

type OrderItemsEditorProps = {
  products: Product[]
  items: DraftItem[]
  failedSku: string | null
  disabled: boolean
  onChange: (items: DraftItem[]) => void
}

export default function OrderItemsEditor({
  products,
  items,
  failedSku,
  disabled,
  onChange,
}: OrderItemsEditorProps) {
  function update(id: string, patch: Partial<DraftItem>) {
    onChange(items.map((item) => (item.id === id ? { ...item, ...patch } : item)))
  }

  function remove(id: string) {
    onChange(items.filter((item) => item.id !== id))
  }

  return (
    <div className="items-editor">
      {items.map((item) => {
        const isFailed = failedSku !== null && item.sku === failedSku

        return (
          <div key={item.id} className={isFailed ? 'order-item row-error' : 'order-item'}>
            <select
              value={item.sku}
              onChange={(event) => update(item.id, { sku: event.target.value })}
              disabled={disabled}
              aria-label="SKU"
            >
              <option value="">select SKU</option>
              {products.map((product) => (
                <option key={product.id} value={product.sku}>
                  {product.sku} — {product.name} (stock {product.stock})
                </option>
              ))}
            </select>

            <input
              type="number"
              min="1"
              step="1"
              value={item.qty}
              onChange={(event) => update(item.id, { qty: event.target.value })}
              disabled={disabled}
              aria-label="Quantity"
            />

            <button
              type="button"
              onClick={() => remove(item.id)}
              disabled={disabled || items.length === 1}
            >
              remove
            </button>
          </div>
        )
      })}

      <button
        type="button"
        onClick={() => onChange([...items, newDraftItem()])}
        disabled={disabled}
      >
        add item
      </button>
    </div>
  )
}
