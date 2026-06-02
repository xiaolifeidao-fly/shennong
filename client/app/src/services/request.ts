import { expireLogin } from '@/utils/authGuard'
import { getToken } from '@/utils/token'
import type { ApiResponse } from '@/types/api'

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'
type RequestData = string | object | ArrayBuffer

interface RequestOptions<TBody = RequestData> {
  method?: HttpMethod
  data?: TBody
  showLoading?: boolean
}

const baseURL = import.meta.env.VITE_APP_API_BASE_URL

function authHeaders() {
  const token = getToken()
  return {
    ...(token ? { token, Authorization: `Bearer ${token}` } : {}),
  }
}

export async function request<TData, TBody extends RequestData = RequestData>(
  url: string,
  options: RequestOptions<TBody> = {},
) {
  const { method = 'GET', data, showLoading = false } = options

  if (showLoading) {
    uni.showLoading({ title: '加载中', mask: true })
  }

  try {
    const response = await uni.request({
      url: `${baseURL}${url}`,
      method,
      data,
      header: {
        'content-type': 'application/json',
        ...authHeaders(),
      },
    })

    const payload = response.data as ApiResponse<TData>
    const isUnauthorized = response.statusCode === 401 || payload?.code === 401

    if (isUnauthorized || !payload?.success) {
      const message = payload?.message || payload?.error || '请求失败'

      if (isUnauthorized || message.includes('未登录') || message.includes('过期') || message.toLowerCase().includes('login')) {
        expireLogin()
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

export async function upload<TData>(
  url: string,
  filePath: string,
  formData: Record<string, string | number> = {},
  name = 'file',
) {
  uni.showLoading({ title: '识别中', mask: true })
  try {
    const response = await uni.uploadFile({
      url: `${baseURL}${url}`,
      filePath,
      name,
      formData,
      header: authHeaders(),
    })
    const payload = JSON.parse(String(response.data || '{}')) as ApiResponse<TData>
    const isUnauthorized = response.statusCode === 401 || payload?.code === 401

    if (isUnauthorized || !payload?.success) {
      const message = payload?.message || payload?.error || '上传失败'
      if (isUnauthorized || message.includes('未登录') || message.includes('过期') || message.toLowerCase().includes('login')) {
        expireLogin()
      }
      throw new Error(message)
    }

    return payload.data
  } finally {
    uni.hideLoading()
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
  upload,
}
