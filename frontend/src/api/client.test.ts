import { describe, expect, it, vi } from 'vitest'
import { ApiError, request } from './client'

function jsonResponse(status: number, body: unknown): Response {
  const text = body === null ? '' : JSON.stringify(body)
  return {
    ok: status >= 200 && status < 300,
    status,
    text: async () => text,
  } as Response
}

describe('request', () => {
  it('returns the JSON body for a successful response', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, { data: [1, 2] }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(request<{ data: number[] }>('/products')).resolves.toEqual({ data: [1, 2] })
    expect(fetchMock).toHaveBeenCalledWith('/api/products', expect.anything())
  })

  it('returns undefined for 204', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(204, null)))

    await expect(request<void>('/products/1')).resolves.toBeUndefined()
  })

  it('throws an ApiError carrying the code from a 400 envelope', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(400, { error: { code: 'VALIDATION_ERROR', message: 'SKU is required' } }),
      ),
    )

    const error = await request('/products').catch((caught: unknown) => caught)

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 400,
      code: 'VALIDATION_ERROR',
      message: 'SKU is required',
    })
  })

  it('exposes available from a 409 envelope', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(409, { error: { code: 'INSUFFICIENT_STOCK', available: 3 } }),
      ),
    )

    const error = await request('/products/1/stock', { method: 'POST' }).catch(
      (caught: unknown) => caught,
    )

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).code).toBe('INSUFFICIENT_STOCK')
    expect((error as ApiError).available).toBe(3)
  })

  it('maps a connection failure to NETWORK_ERROR', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    const error = await request('/products').catch((caught: unknown) => caught)

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 0, code: 'NETWORK_ERROR' })
  })

  it('flags a non-JSON body as INVALID_RESPONSE', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 502,
        text: async () => '<html>bad gateway</html>',
      } as Response),
    )

    const error = await request('/products').catch((caught: unknown) => caught)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).code).toBe('INVALID_RESPONSE')
  })
})
