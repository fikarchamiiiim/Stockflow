export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly details: Record<string, unknown>

  constructor(
    status: number,
    code: string,
    details: Record<string, unknown>,
    message?: string,
  ) {
    super(message ?? code)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }

  get available(): number | undefined {
    const value = this.details.available
    return typeof value === 'number' ? value : undefined
  }
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers)
  if (init?.body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  let response: Response
  try {
    response = await fetch(`/api${path}`, { ...init, headers })
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') {
      throw cause
    }
    throw new ApiError(0, 'NETWORK_ERROR', {}, 'cannot reach the server')
  }

  if (response.status === 204) {
    return undefined as T
  }

  const text = await response.text()
  const body = parseBody(text, response.status)

  if (!response.ok) {
    throw toApiError(response.status, body)
  }

  return body as T
}

function parseBody(text: string, status: number): unknown {
  if (text === '') {
    return null
  }
  try {
    return JSON.parse(text) as unknown
  } catch {
    throw new ApiError(status, 'INVALID_RESPONSE', {}, 'invalid server response')
  }
}

function toApiError(status: number, body: unknown): ApiError {
  if (typeof body === 'object' && body !== null && 'error' in body) {
    const error = (body as { error: unknown }).error
    if (typeof error === 'object' && error !== null) {
      const details = error as Record<string, unknown>
      const code = typeof details.code === 'string' ? details.code : 'UNKNOWN'
      const message = typeof details.message === 'string' ? details.message : undefined
      return new ApiError(status, code, details, message)
    }
  }

  return new ApiError(status, 'UNKNOWN', {}, `HTTP ${status}`)
}
