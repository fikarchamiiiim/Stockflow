const rupiah = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
})

const MINUTE = 60_000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

export function formatRupiah(value: number): string {
  return rupiah.format(value)
}

export function formatRelativeTime(iso: string | null): string {
  if (iso === null) {
    return '—'
  }

  const timestamp = new Date(iso).getTime()
  if (Number.isNaN(timestamp)) {
    return '—'
  }

  const diff = Date.now() - timestamp
  if (diff < MINUTE) {
    return 'just now'
  }
  if (diff < HOUR) {
    return `${Math.floor(diff / MINUTE)}m ago`
  }
  if (diff < DAY) {
    return `${Math.floor(diff / HOUR)}h ago`
  }
  return `${Math.floor(diff / DAY)}d ago`
}
