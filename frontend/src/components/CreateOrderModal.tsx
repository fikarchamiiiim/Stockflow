import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ApiError } from '../api/client'
import { createOrder } from '../api/orders'
import { getProducts } from '../api/products'
import { formatRupiah } from '../lib/format'
import { newDraftItem } from '../lib/orderDraft'
import { parseQuantity } from '../lib/validation'
import OrderItemsEditor from './OrderItemsEditor'
import type { DraftItem } from '../lib/orderDraft'
import type { Order, Product } from '../types/api'

type CreateOrderModalProps = {
  onClose: () => void
  onCreated: () => void
}

type Insufficient = {
  sku: string
  available: number
}

export default function CreateOrderModal({ onClose, onCreated }: CreateOrderModalProps) {
  const [products, setProducts] = useState<Product[] | null>(null)
  const [productsError, setProductsError] = useState<string | null>(null)

  const [customer, setCustomer] = useState('')
  const [items, setItems] = useState<DraftItem[]>([newDraftItem()])
  const [saving, setSaving] = useState(false)
  const [messages, setMessages] = useState<string[]>([])
  const [insufficient, setInsufficient] = useState<Insufficient | null>(null)
  const [serverError, setServerError] = useState<string | null>(null)
  const [created, setCreated] = useState<Order | null>(null)

  useEffect(() => {
    const controller = new AbortController()

    getProducts({ limit: 100 }, controller.signal)
      .then((res) => setProducts(res.data))
      .catch((error: unknown) => {
        if (controller.signal.aborted) {
          return
        }
        setProductsError(
          error instanceof ApiError ? `${error.code} (${error.status})` : 'failed to load products',
        )
      })

    return () => controller.abort()
  }, [])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    setMessages([])
    setInsufficient(null)
    setServerError(null)

    const problems: string[] = []
    const customerName = customer.trim()
    if (customerName === '') {
      problems.push('Customer name is required.')
    }

    const seen = new Set<string>()
    const payloadItems: Array<{ sku: string; qty: number }> = []

    for (const item of items) {
      if (item.sku === '') {
        problems.push('Every row must have a SKU.')
        continue
      }
      if (seen.has(item.sku)) {
        problems.push(`Duplicate SKU: ${item.sku}.`)
        continue
      }
      seen.add(item.sku)

      const quantity = parseQuantity(item.qty)
      if (!quantity.ok) {
        problems.push(`${item.sku}: ${quantity.message}`)
        continue
      }
      payloadItems.push({ sku: item.sku, qty: quantity.value })
    }

    if (payloadItems.length === 0) {
      problems.push('At least 1 valid item is required.')
    }
    if (problems.length > 0) {
      setMessages(problems)
      return
    }

    setSaving(true)
    try {
      const order = await createOrder({ customer: customerName, items: payloadItems })
      setCreated(order)
      onCreated()
    } catch (error) {
      if (error instanceof ApiError && error.code === 'INSUFFICIENT_STOCK') {
        const sku = typeof error.details.sku === 'string' ? error.details.sku : ''
        setInsufficient({ sku, available: error.available ?? 0 })
      } else if (error instanceof ApiError && error.status === 400) {
        setMessages(error.message.split(', '))
      } else if (error instanceof ApiError) {
        setServerError(`Failed to create order: ${error.code} (${error.status}).`)
      } else {
        setServerError('Failed to create order.')
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-backdrop">
      <div className="modal" role="dialog" aria-modal="true" aria-labelledby="create-order-title">
        {created !== null ? (
          <>
            <h2 id="create-order-title">Order #{created.id} created</h2>
            <p className="muted">
              {created.customer} · status {created.status}
            </p>

            <table>
              <thead>
                <tr>
                  <th scope="col">SKU</th>
                  <th scope="col">Qty</th>
                  <th scope="col">Price</th>
                  <th scope="col">Subtotal</th>
                </tr>
              </thead>
              <tbody>
                {created.items.map((item) => (
                  <tr key={item.id}>
                    <td>{item.sku}</td>
                    <td>{item.qty}</td>
                    <td>{formatRupiah(item.unit_price)}</td>
                    <td>{formatRupiah(item.unit_price * item.qty)}</td>
                  </tr>
                ))}
              </tbody>
            </table>

            <p>
              <strong>Total: {formatRupiah(created.total)}</strong>{' '}
            </p>

            <div className="modal-actions">
              <Link to={`/orders/${created.id}`} onClick={onClose}>
                view order detail
              </Link>
              <button type="button" onClick={onClose}>
                close
              </button>
            </div>
          </>
        ) : (
          <>
            <h2 id="create-order-title">Create order</h2>

            {productsError !== null && (
              <p className="error-text">Failed to load products: {productsError}</p>
            )}
            {products === null && productsError === null && (
              <p className="muted">loading products...</p>
            )}

            {products !== null && (
              <form onSubmit={handleSubmit} className="modal-form">
                <label className="field">
                  Customer
                  <input
                    type="text"
                    value={customer}
                    onChange={(event) => setCustomer(event.target.value)}
                    disabled={saving}
                    autoFocus
                  />
                </label>

                <span className="field-label">Items</span>
                <OrderItemsEditor
                  products={products}
                  items={items}
                  failedSku={insufficient?.sku ?? null}
                  disabled={saving}
                  onChange={setItems}
                />

                {insufficient !== null && (
                  <p className="error-text" role="alert">
                    {insufficient.sku} does not have enough stock. Available:{' '}
                    {insufficient.available}.
                  </p>
                )}

                {messages.length > 0 && (
                  <ul className="message-list" role="alert">
                    {messages.map((message) => (
                      <li key={message}>{message}</li>
                    ))}
                  </ul>
                )}

                {serverError !== null && (
                  <p className="error-text" role="alert">
                    {serverError}
                  </p>
                )}

                <div className="modal-actions">
                  <button type="submit" disabled={saving}>
                    {saving ? 'saving...' : 'create order'}
                  </button>
                  <button type="button" onClick={onClose} disabled={saving}>
                    cancel
                  </button>
                </div>
              </form>
            )}

            {productsError !== null && (
              <div className="modal-actions">
                <button type="button" onClick={onClose}>
                  close
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
