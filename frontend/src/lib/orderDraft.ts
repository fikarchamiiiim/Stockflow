export type DraftItem = {
  id: string
  sku: string
  qty: string
}

export function newDraftItem(): DraftItem {
  return { id: crypto.randomUUID(), sku: '', qty: '1' }
}
