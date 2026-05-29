import type { GrainEntry } from '@/types/grain'

export function formatAmount(value: number) {
  return `¥${value.toLocaleString('zh-CN', { maximumFractionDigits: 2 })}`
}

export function formatQuantity(value: number, unit = '公斤') {
  return `${value.toLocaleString('zh-CN')} ${unit}`
}

export function calcUnitPrice(amount: number, quantity: number, unit = '公斤') {
  if (!amount || !quantity) {
    return ''
  }

  return `${(amount / quantity).toFixed(2)} 元/${unit}`
}

export function getEntryPrice(entry: GrainEntry) {
  return calcUnitPrice(entry.amount, entry.quantity, entry.unit)
}

export function pickTime(value: string) {
  return value.split(' ')[1] || value
}
