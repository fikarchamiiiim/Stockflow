export interface Product {
  id: number
  sku: string
  name: string
  price: number
  stock: number
  is_active: boolean
  created_at: string
  last_movement_at: string | null
}

export interface Paginated<T> {
  data: T[]
  page: number
  limit: number
  total: number
}

export type MovementType = 'in' | 'out'

export interface StockMovement {
  id: number
  type: MovementType
  qty: number
  note: string | null
  created_at: string
}

export type OrderStatus = 'CREATED' | 'CANCELLED'

export interface OrderItem {
  id: number
  sku: string
  qty: number
  unit_price: number
}

export interface Order {
  id: number
  customer: string
  status: OrderStatus
  total: number
  created_at: string
  cancelled_at: string | null
  items: OrderItem[]
}

export interface ProductListQuery {
  search?: string
  page?: number
  limit?: number
}

export interface CreateProductInput {
  sku: string
  name: string
  price: number
}

export interface UpdateProductInput {
  sku: string
  name: string
  price: number
  is_active: boolean
}

export interface AdjustStockInput {
  type: MovementType
  qty: number
  note?: string
}

export interface CreateOrderInput {
  customer: string
  items: Array<{ sku: string; qty: number }>
}
