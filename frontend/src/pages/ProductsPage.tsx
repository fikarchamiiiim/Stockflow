import { useState } from 'react'
import AdjustStockModal from '../components/AdjustStockModal'
import CreateOrderModal from '../components/CreateOrderModal'
import ProductTable from '../components/ProductTable'
import { EmptyState, ErrorState, LoadingState } from '../components/StateViews'
import SummaryCards from '../components/SummaryCards'
import { useDebounce } from '../hooks/useDebounce'
import { useProducts } from '../hooks/useProducts'
import type { Product } from '../types/api'

type ProductFilter = 'all' | 'active' | 'inactive' | 'low'

const FILTERS: ProductFilter[] = ['all', 'active', 'inactive', 'low']
const LOW_STOCK_THRESHOLD = 10

export default function ProductsPage() {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<ProductFilter>('all')
  const [selected, setSelected] = useState<Product | null>(null)
  const [creating, setCreating] = useState(false)

  const debouncedSearch = useDebounce(search, 300)
  const { products, total, error, loading, refreshing, reload } = useProducts({
    search: debouncedSearch,
    limit: 100,
  })

  const visible = products.filter((product) => matchesFilter(product, filter))

  return (
    <section>
      <div className="page-header">
        <h1>Products</h1>
        <button type="button" onClick={() => setCreating(true)}>
          Create order
        </button>
      </div>

      <div className="toolbar">
        <input
          type="search"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Search SKU or name"
          aria-label="Search products"
        />
        <select
          value={filter}
          onChange={(event) => {
            if (isFilter(event.target.value)) {
              setFilter(event.target.value)
            }
          }}
          aria-label="Product filter"
        >
          <option value="all">All</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="low">Low stock (≤ {LOW_STOCK_THRESHOLD})</option>
        </select>
      </div>

      {error !== null && <ErrorState message={error} onRetry={reload} />}

      {loading && <LoadingState />}

      {error === null && !loading && (
        <>
          <SummaryCards total={total} products={products} />

          {refreshing && <p className="state muted">loading...</p>}

          {visible.length === 0 ? (
            <EmptyState message="No matching products." />
          ) : (
            <ProductTable products={visible} onAdjustStock={(product) => setSelected(product)} />
          )}

          <p className="muted">
            Showing {visible.length} of {total} products
          </p>
        </>
      )}

      {selected !== null && (
        <AdjustStockModal
          product={selected}
          onClose={() => setSelected(null)}
          onSuccess={() => {
            setSelected(null)
            reload()
          }}
        />
      )}

      {creating && (
        <CreateOrderModal
          onClose={() => setCreating(false)}
          onCreated={() => reload()}
        />
      )}
    </section>
  )
}

function isFilter(value: string): value is ProductFilter {
  return (FILTERS as string[]).includes(value)
}

function matchesFilter(product: Product, filter: ProductFilter): boolean {
  switch (filter) {
    case 'active':
      return product.is_active
    case 'inactive':
      return !product.is_active
    case 'low':
      return product.stock <= LOW_STOCK_THRESHOLD
    default:
      return true
  }
}
