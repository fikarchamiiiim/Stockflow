export type QuantityResult = { ok: true; value: number } | { ok: false; message: string }

export function parseQuantity(input: string): QuantityResult {
  const trimmed = input.trim()

  if (trimmed === '') {
    return { ok: false, message: 'Quantity is required.' }
  }
  if (!/^\d+$/.test(trimmed)) {
    return { ok: false, message: 'Quantity must be a whole number greater than 0.' }
  }

  const value = Number(trimmed)
  if (value <= 0) {
    return { ok: false, message: 'Quantity must be a whole number greater than 0.' }
  }

  return { ok: true, value }
}
