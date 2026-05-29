type QueryValue = string | number | boolean | null | undefined

export function buildQuery<T extends object>(params: T) {
  const pairs = Object.entries(params as Record<string, QueryValue>)
    .filter(([, value]) => value !== undefined && value !== null && value !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)

  return pairs.length ? `?${pairs.join('&')}` : ''
}
