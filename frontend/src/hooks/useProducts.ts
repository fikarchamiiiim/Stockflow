import { useEffect, useState } from 'react'
import { ApiError } from '../api/client'
import { getProducts } from '../api/products'
import type { Product, ProductListQuery } from '../types/api'

type Loaded = {
  key: string
  products: Product[]
  total: number
}

type Failure = {
  key: string
  message: string
}

export type ProductsState = {
  products: Product[]
  total: number
  error: string | null
  loading: boolean
  refreshing: boolean
  reload: () => void
}

export function useProducts(query: ProductListQuery): ProductsState {
  const search = query.search ?? ''
  const limit = query.limit ?? 100
  const key = `${search}|${limit}`

  const [result, setResult] = useState<Loaded | null>(null)
  const [failure, setFailure] = useState<Failure | null>(null)
  const [reloadKey, setReloadKey] = useState(0)

  useEffect(() => {
    const controller = new AbortController()

    getProducts({ search, limit }, controller.signal)
      .then((res) => {
        setResult({ key: `${search}|${limit}`, products: res.data, total: res.total })
        setFailure(null)
      })
      .catch((error: unknown) => {
        if (controller.signal.aborted) {
          return
        }
        const message =
          error instanceof ApiError ? `${error.code} (${error.status})` : 'failed to load data'
        setFailure({ key: `${search}|${limit}`, message })
      })

    return () => controller.abort()
  }, [search, limit, reloadKey])

  const error = failure !== null && failure.key === key ? failure.message : null
  const fresh = result !== null && result.key === key

  return {
    products: result?.products ?? [],
    total: result?.total ?? 0,
    error,
    loading: result === null && error === null,
    refreshing: result !== null && !fresh && error === null,
    reload: () => {
      setFailure(null)
      setReloadKey((value) => value + 1)
    },
  }
}
