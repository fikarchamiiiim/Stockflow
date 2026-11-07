import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError } from '../api/client'
import { cancelOrder, getOrder } from '../api/orders'
import { ErrorState, LoadingState } from '../components/StateViews'
import { formatRelativeTime, formatRupiah } from '../lib/format'
import type { Order } from '../types/api'

type DetailState =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'success'; order: Order }

export default function OrderDetailPage() {
  const params = useParams<{ id: string }>()
  const orderId = Number(params.id)
  const invalidId = !Number.isInteger(orderId) || orderId <= 0

  const [state, setState] = useState<DetailState>({ status: 'loading' })
  const [reloadKey, setReloadKey] = useState(0)
  const [cancelling, setCancelling] = useState(false)
  const [cancelError, setCancelError] = useState<string | null>(null)

  useEffect(() => {
    if (invalidId) {
      return
    }

    const controller = new AbortController()

    getOrder(orderId, controller.signal)
      .then((order) => setState({ status: 'success', order }))
      .catch((error: unknown) => {
        if (controller.signal.aborted) {
          return
        }
        const message =
          error instanceof ApiError ? `${error.code} (${error.status})` : 'failed to load order'
        setState({ status: 'error', message })
      })

    return () => controller.abort()
  }, [orderId, invalidId, reloadKey])

  async function handleCancel(order: Order) {
    const confirmed = window.confirm(
      `Cancel order #${order.id}? Stock for all items will be restored.`,
    )
    if (!confirmed) {
      return
    }

    setCancelling(true)
    setCancelError(null)

    try {
      const updated = await cancelOrder(order.id)
      setState({ status: 'success', order: updated })
    } catch (error) {
      if (error instanceof ApiError && error.code === 'ALREADY_CANCELLED') {
        setCancelError('This order has already been cancelled.')
        try {
          const fresh = await getOrder(order.id)
          setState({ status: 'success', order: fresh })
        } catch {
          setCancelError('This order has already been cancelled, but reloading it failed.')
        }
      } else if (error instanceof ApiError) {
        setCancelError(`Failed to cancel order: ${error.code} (${error.status}).`)
      } else {
        setCancelError('Failed to cancel order.')
      }
    } finally {
      setCancelling(false)
    }
  }

  return (
    <section>
      <p>
        <Link to="/">← back to products</Link>
      </p>

      {invalidId && <p className="error-text">Invalid order id.</p>}

      {!invalidId && state.status === 'loading' && <LoadingState />}

      {!invalidId && state.status === 'error' && (
        <ErrorState
          message={state.message}
          onRetry={() => {
            setState({ status: 'loading' })
            setReloadKey((key) => key + 1)
          }}
        />
      )}

      {!invalidId && state.status === 'success' && (
        <>
          <h1>Order #{state.order.id}</h1>

          <p className="detail-meta">
            <span>{state.order.customer}</span>
            <span
              className={state.order.status === 'CREATED' ? 'badge badge-active' : 'badge badge-inactive'}
            >
              {state.order.status}
            </span>
            <span className="muted">created {formatRelativeTime(state.order.created_at)}</span>
            {state.order.cancelled_at !== null && (
              <span className="muted">
                cancelled {formatRelativeTime(state.order.cancelled_at)}
              </span>
            )}
          </p>

          <table>
            <thead>
              <tr>
                <th scope="col">SKU</th>
                <th scope="col">Qty</th>
                <th scope="col">Unit price</th>
                <th scope="col">Subtotal</th>
              </tr>
            </thead>
            <tbody>
              {state.order.items.map((item) => (
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
            <strong>Total: {formatRupiah(state.order.total)}</strong>{' '}
            <span className="muted">(computed by server)</span>
          </p>

          {cancelError !== null && (
            <p className="error-text" role="alert">
              {cancelError}
            </p>
          )}

          <button
            type="button"
            onClick={() => handleCancel(state.order)}
            disabled={cancelling || state.order.status !== 'CREATED'}
          >
            {cancelling ? 'cancelling...' : 'cancel order'}
          </button>
        </>
      )}
    </section>
  )
}
