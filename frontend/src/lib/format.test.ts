import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatRelativeTime, formatRupiah } from './format'

const NOW = new Date('2026-01-01T12:00:00Z')

function normalizeSpaces(value: string): string {
  return value.replace(/\u00A0/g, ' ')
}

describe('formatRupiah', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(NOW)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('formats numbers with thousand separators and no decimals', () => {
    expect(normalizeSpaces(formatRupiah(1234567))).toBe('Rp 1.234.567')
  })

  it('formats zero', () => {
    expect(normalizeSpaces(formatRupiah(0))).toBe('Rp 0')
  })
})

describe('formatRelativeTime', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(NOW)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns a dash for null', () => {
    expect(formatRelativeTime(null)).toBe('—')
  })

  it('returns a dash for an invalid date', () => {
    expect(formatRelativeTime('not-a-date')).toBe('—')
  })

  it('shows "just now" for less than a minute', () => {
    expect(formatRelativeTime(new Date(NOW.getTime() - 30_000).toISOString())).toBe('just now')
  })

  it('shows minutes, hours, and days', () => {
    expect(formatRelativeTime(new Date(NOW.getTime() - 5 * 60_000).toISOString())).toBe('5m ago')
    expect(formatRelativeTime(new Date(NOW.getTime() - 2 * 3_600_000).toISOString())).toBe('2h ago')
    expect(formatRelativeTime(new Date(NOW.getTime() - 3 * 86_400_000).toISOString())).toBe('3d ago')
  })
})
