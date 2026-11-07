import { describe, expect, it } from 'vitest'
import { parseQuantity } from './validation'

describe('parseQuantity', () => {
  it('rejects empty input', () => {
    expect(parseQuantity('')).toEqual({ ok: false, message: 'Quantity is required.' })
    expect(parseQuantity('   ')).toEqual({ ok: false, message: 'Quantity is required.' })
  })

  it('rejects zero and negative numbers', () => {
    expect(parseQuantity('0').ok).toBe(false)
    expect(parseQuantity('-1').ok).toBe(false)
  })

  it('rejects fractions and non-numeric formats', () => {
    expect(parseQuantity('1.5').ok).toBe(false)
    expect(parseQuantity('abc').ok).toBe(false)
    expect(parseQuantity('1e3').ok).toBe(false)
  })

  it('accepts positive whole numbers', () => {
    expect(parseQuantity('12')).toEqual({ ok: true, value: 12 })
    expect(parseQuantity(' 007 ')).toEqual({ ok: true, value: 7 })
  })
})
