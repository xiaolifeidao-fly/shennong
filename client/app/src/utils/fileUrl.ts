const OSS_HTTP_HOSTS = [
  'sn-2026.oss-cn-hangzhou.aliyuncs.com',
]

export function normalizeFileUrl(url?: string) {
  const value = String(url || '').trim()
  if (!value) return ''

  const matchedHost = OSS_HTTP_HOSTS.find((host) => value.startsWith(`http://${host}/`))
  if (matchedHost) {
    return value.replace(`http://${matchedHost}/`, `https://${matchedHost}/`)
  }

  return value
}

export function normalizeFileUrls(urls: string[] = []) {
  return urls.map((url) => normalizeFileUrl(url)).filter(Boolean)
}
