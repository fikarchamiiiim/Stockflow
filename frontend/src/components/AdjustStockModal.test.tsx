import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../api/client'
import type { Product } from '../types/api'
import AdjustStockModal from './AdjustStockModal'

const { adjustStock } = vi.hoisted(() => ({ adjustStock: vi.fn() }))

vi.mock('../api/products', () => ({ adjustStock }))

const product: Product = {
  id: 1,
  sku: 'AQUA-600',
  name: 'Aqua Air Mineral 600ml',
  price: 4000,
  stock: 5,
  is_active: true,
  created_at: '2026-01-01T00:00:00Z',
  last_movement_at: null,
}

function submitForm() {
  const button = screen.getByRole('button', { name: 'save' })
  const form = button.closest('form')
  if (form === null) {
    throw new Error('form not found')
  }
  fireEvent.submit(form)
}

describe('AdjustStockModal', () => {
  beforeEach(() => {
    adjustStock.mockReset()
  })

  afterEach(() => {
    cleanup()
  })

  it('rejects a quantity of 0 without calling the API', async () => {
    render(<AdjustStockModal product={product} onClose={vi.fn()} onSuccess={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Quantity'), { target: { value: '0' } })
    submitForm()

    expect(await screen.findByText(/Quantity must be a whole number/i)).toBeTruthy()
    expect(adjustStock).not.toHaveBeenCalled()
  })

  it('shows the available quantity from a 409 inside the modal', async () => {
    adjustStock.mockRejectedValue(
      new ApiError(409, 'INSUFFICIENT_STOCK', { available: 5 }, 'insufficient stock'),
    )
    const onSuccess = vi.fn()

    render(<AdjustStockModal product={product} onClose={vi.fn()} onSuccess={onSuccess} />)

    fireEvent.change(screen.getByLabelText('Quantity'), { target: { value: '99' } })
    fireEvent.click(screen.getByRole('radio', { name: 'Out' }))
    submitForm()

    expect(await screen.findByText(/Available: 5/)).toBeTruthy()
    expect(onSuccess).not.toHaveBeenCalled()
  })

  it('calls onSuccess after a successful save', async () => {
    adjustStock.mockResolvedValue({ ...product, stock: 8 })
    const onSuccess = vi.fn()

    render(<AdjustStockModal product={product} onClose={vi.fn()} onSuccess={onSuccess} />)

    fireEvent.change(screen.getByLabelText('Quantity'), { target: { value: '3' } })
    submitForm()

    await waitFor(() => expect(onSuccess).toHaveBeenCalled())
    expect(adjustStock).toHaveBeenCalledWith(1, { type: 'in', qty: 3, note: undefined })
  })
})
