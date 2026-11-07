import { useState } from 'react'
import { ApiError } from '../api/client'
import { adjustStock } from '../api/products'
import { parseQuantity } from '../lib/validation'
import type { MovementType, Product } from '../types/api'

type AdjustStockModalProps = {
  product: Product
  onClose: () => void
  onSuccess: () => void
}

export default function AdjustStockModal({ product, onClose, onSuccess }: AdjustStockModalProps) {
  const [type, setType] = useState<MovementType>('in')
  const [qty, setQty] = useState('')
  const [note, setNote] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const quantity = parseQuantity(qty)
    if (!quantity.ok) {
      setError(quantity.message)
      return
    }

    setSaving(true)
    setError(null)

    try {
      await adjustStock(product.id, {
        type,
        qty: quantity.value,
        note: note.trim() === '' ? undefined : note.trim(),
      })
      onSuccess()
    } catch (err) {
      if (err instanceof ApiError && err.code === 'INSUFFICIENT_STOCK') {
        setError(`Insufficient stock. Available: ${err.available ?? 0}.`)
      } else if (err instanceof ApiError) {
        setError(`Failed to save: ${err.code} (${err.status}).`)
      } else {
        setError('Failed to save.')
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-backdrop">
      <div className="modal" role="dialog" aria-modal="true" aria-labelledby="adjust-stock-title">
        <h2 id="adjust-stock-title">Adjust stock — {product.sku}</h2>
        <p className="muted">
          {product.name} · current stock {product.stock}
        </p>

        <form onSubmit={handleSubmit} className="modal-form">
          <div className="field-row">
            <label>
              <input
                type="radio"
                name="type"
                value="in"
                checked={type === 'in'}
                onChange={() => setType('in')}
                disabled={saving}
              />{' '}
              In
            </label>
            <label>
              <input
                type="radio"
                name="type"
                value="out"
                checked={type === 'out'}
                onChange={() => setType('out')}
                disabled={saving}
              />{' '}
              Out
            </label>
          </div>

          <label className="field">
            Quantity
            <input
              type="number"
              min="1"
              step="1"
              value={qty}
              onChange={(event) => setQty(event.target.value)}
              disabled={saving}
              autoFocus
            />
          </label>

          <label className="field">
            Note (optional)
            <input
              type="text"
              value={note}
              onChange={(event) => setNote(event.target.value)}
              disabled={saving}
            />
          </label>

          {error !== null && (
            <p className="error-text" role="alert">
              {error}
            </p>
          )}

          <div className="modal-actions">
            <button type="submit" disabled={saving}>
              {saving ? 'saving...' : 'save'}
            </button>
            <button type="button" onClick={onClose} disabled={saving}>
              cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
