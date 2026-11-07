import { formatRupiah } from '../lib/format'
import type { Product } from '../types/api'

type SummaryCardsProps = {
  total: number
  products: Product[]
}

const LOW_STOCK_THRESHOLD = 10

export default function SummaryCards({ total, products }: SummaryCardsProps) {
  const active = products.filter((product) => product.is_active).length
  const stockValue = products.reduce((sum, product) => sum + product.price * product.stock, 0)
  const lowStock = products.filter((product) => product.stock <= LOW_STOCK_THRESHOLD).length

  return (
    <div className="cards">
      <div className="card">
        <span className="card-label">Total products</span>
        <strong>{total}</strong>
      </div>
      <div className="card">
        <span className="card-label">Active</span>
        <strong>{active}</strong>
      </div>
      <div className="card">
        <span className="card-label">Stock value</span>
        <strong>{formatRupiah(stockValue)}</strong>
      </div>
      <div className="card">
        <span className="card-label">Low stock (≤ {LOW_STOCK_THRESHOLD})</span>
        <strong>{lowStock}</strong>
      </div>
    </div>
  )
}
