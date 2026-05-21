import { clearToken, getToken } from '@/utils/token'
import type { ApiResponse } from '@/types/api'

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'
type RequestData = string | object | ArrayBuffer

interface RequestOptions<TBody = RequestData> {
  method?: HttpMethod
  data?: TBody
  showLoading?: boolean
}

const baseURL = import.meta.env.VITE_APP_API_BASE_URL

export async function request<TData, TBody extends RequestData = RequestData>(
  url: string,
  options: RequestOptions<TBody> = {},
) {
  const { method = 'GET', data, showLoading = false } = options

  if (showLoading) {
    uni.showLoading({ title: '加载中', mask: true })
  }

  try {
    const token = getToken()
    const response = await uni.request({
      url: `${baseURL}${url}`,
      method,
      data,
      header: {
        'content-type': 'application/json',
        ...(token ? { token, Authorization: `Bearer ${token}` } : {}),
      },
    })

    const payload = response.data as ApiResponse<TData>

    if (!payload?.success) {
      const message = payload?.message || payload?.error || '请求失败'

      if (message.includes('未登录') || message.toLowerCase().includes('login')) {
        clearToken()
      }

      throw new Error(message)
    }

    return payload.data
  } finally {
    if (showLoading) {
      uni.hideLoading()
    }
  }
}

export const http = {
  get: <TData>(url: string, options?: Omit<RequestOptions, 'method' | 'data'>) =>
    request<TData>(url, { ...options, method: 'GET' }),
  post: <TData, TBody extends RequestData = RequestData>(
    url: string,
    data?: TBody,
    options?: Omit<RequestOptions<TBody>, 'method' | 'data'>,
  ) => request<TData, TBody>(url, { ...options, method: 'POST', data }),
  put: <TData, TBody extends RequestData = RequestData>(
    url: string,
    data?: TBody,
    options?: Omit<RequestOptions<TBody>, 'method' | 'data'>,
  ) => request<TData, TBody>(url, { ...options, method: 'PUT', data }),
  delete: <TData>(url: string, options?: Omit<RequestOptions, 'method' | 'data'>) =>
    request<TData>(url, { ...options, method: 'DELETE' }),
}
