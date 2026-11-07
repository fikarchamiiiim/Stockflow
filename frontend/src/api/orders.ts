import { request } from './client'
import type { CreateOrderInput, Order } from '../types/api'

export function createOrder(input: CreateOrderInput): Promise<Order> {
  return request<Order>('/orders', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function getOrder(id: number, signal?: AbortSignal): Promise<Order> {
  return request<Order>(`/orders/${id}`, { signal })
}

export function cancelOrder(id: number): Promise<Order> {
  return request<Order>(`/orders/${id}/cancel`, { method: 'POST' })
}
