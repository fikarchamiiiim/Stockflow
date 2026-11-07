import { formatRelativeTime, formatRupiah } from '../lib/format'
import type { Product } from '../types/api'

type ProductTableProps = {
  products: Product[]
  onAdjustStock: (product: Product) => void
}

export default function ProductTable({ products, onAdjustStock }: ProductTableProps) {
  return (
    <table>
      <thead>
        <tr>
          <th scope="col">SKU</th>
          <th scope="col">Name</th>
          <th scope="col">Price</th>
          <th scope="col">Stock</th>
          <th scope="col">Status</th>
          <th scope="col">Last movement</th>
          <th scope="col">Actions</th>
        </tr>
      </thead>
      <tbody>
        {products.map((product) => (
          <tr key={product.id}>
            <td>{product.sku}</td>
            <td>{product.name}</td>
            <td>{formatRupiah(product.price)}</td>
            <td>{product.stock}</td>
            <td>
              <span className={product.is_active ? 'badge badge-active' : 'badge badge-inactive'}>
                {product.is_active ? 'Active' : 'Inactive'}
              </span>
            </td>
            <td>{formatRelativeTime(product.last_movement_at)}</td>
            <td>
              <button type="button" onClick={() => onAdjustStock(product)}>
                Adjust stock
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
