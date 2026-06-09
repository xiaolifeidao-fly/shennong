export function maskPhone(value?: string) {
  const text = String(value || '').trim()
  if (!text) return ''
  if (text.length <= 7) return maskMiddle(text)
  return `${text.slice(0, 3)}****${text.slice(-4)}`
}

export function maskIdNumber(value?: string) {
  const text = String(value || '').trim()
  if (!text) return ''
  if (text.length <= 8) return maskMiddle(text)
  return `${text.slice(0, 4)}**********${text.slice(-4)}`
}

function maskMiddle(value: string) {
  if (value.length <= 2) return '*'.repeat(value.length)
  return `${value.slice(0, 1)}${'*'.repeat(Math.max(2, value.length - 2))}${value.slice(-1)}`
}
