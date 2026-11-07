import { request } from './client'
import type {
  AdjustStockInput,
  CreateProductInput,
  Paginated,
  Product,
  ProductListQuery,
  StockMovement,
  UpdateProductInput,
} from '../types/api'

export function getProducts(
  query: ProductListQuery = {},
  signal?: AbortSignal,
): Promise<Paginated<Product>> {
  const params = new URLSearchParams()
  if (query.search !== undefined && query.search !== '') {
    params.set('search', query.search)
  }
  if (query.page !== undefined) {
    params.set('page', String(query.page))
  }
  if (query.limit !== undefined) {
    params.set('limit', String(query.limit))
  }

  const suffix = params.toString()
  return request<Paginated<Product>>(`/products${suffix ? `?${suffix}` : ''}`, { signal })
}

export function createProduct(input: CreateProductInput): Promise<Product> {
  return request<Product>('/products', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function updateProduct(id: number, input: UpdateProductInput): Promise<Product> {
  return request<Product>(`/products/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}

export function deleteProduct(id: number): Promise<void> {
  return request<void>(`/products/${id}`, { method: 'DELETE' })
}

export function adjustStock(id: number, input: AdjustStockInput): Promise<Product> {
  return request<Product>(`/products/${id}/stock`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function getMovements(
  id: number,
  signal?: AbortSignal,
): Promise<{ data: StockMovement[] }> {
  return request<{ data: StockMovement[] }>(`/products/${id}/movements`, { signal })
}
