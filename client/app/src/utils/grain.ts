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

// 明细展示：按 displayUnit 换算（服务端存公斤，显示时按录入单位还原）
export function formatQuantityByDisplayUnit(quantityKg: number, displayUnit: string) {
  if (displayUnit === '吨') {
    const val = quantityKg / 1000
    return `${val % 1 === 0 ? val : val.toFixed(3).replace(/\.?0+$/, '')} 吨`
  }
  return `${quantityKg.toLocaleString('zh-CN')} 公斤`
}

// 汇总展示：>=1000公斤 显示吨，否则显示公斤
export function formatQuantitySummary(quantityKg: number) {
  if (quantityKg >= 1000) {
    const val = quantityKg / 1000
    const str = val % 1 === 0 ? String(val) : val.toFixed(3).replace(/\.?0+$/, '')
    return `${str} 吨`
  }
  return `${quantityKg.toLocaleString('zh-CN')} 公斤`
}
