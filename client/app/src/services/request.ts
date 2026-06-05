import { expireLogin } from '@/utils/authGuard'
import { getToken } from '@/utils/token'
import type { ApiResponse } from '@/types/api'

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'
type RequestData = string | object | ArrayBuffer
type ApiTransport = 'cloud' | 'https'

interface RequestOptions<TBody = RequestData> {
  method?: HttpMethod
  data?: TBody
  showLoading?: boolean
}

const baseURL = import.meta.env.VITE_APP_API_BASE_URL
const apiTransport = (import.meta.env.VITE_APP_API_TRANSPORT || 'https') as ApiTransport
const cloudEnv = import.meta.env.VITE_APP_CLOUD_ENV
const cloudService = import.meta.env.VITE_APP_CLOUD_SERVICE
const cloudFileBaseURL = import.meta.env.VITE_APP_CLOUD_FILE_BASE_URL || baseURL

interface ApiRawResponse<TData> {
  statusCode: number
  data: ApiResponse<TData>
}

interface WechatCloudContainer {
  callContainer<TData = unknown>(options: {
    config: {
      env: string
    }
    path: string
    header: Record<string, string>
    method: HttpMethod
    data?: RequestData
  }): Promise<ApiRawResponse<TData>>
  uploadFile(options: {
    cloudPath: string
    filePath: string
  }): Promise<{ fileID: string }>
  getTempFileURL(options: {
    fileList: string[]
  }): Promise<{
    fileList: Array<{
      fileID: string
      tempFileURL: string
      status: number
      errMsg?: string
    }>
  }>
}

declare const wx: {
  env?: {
    USER_DATA_PATH: string
  }
  cloud?: WechatCloudContainer
  uploadFile?(options: {
    url: string
    filePath: string
    name: string
    formData?: Record<string, string | number>
    header?: Record<string, string>
    success: (response: { statusCode: number; data: string }) => void
    fail: (error: { errMsg?: string }) => void
  }): unknown
  downloadFile?(options: {
    url: string
    header?: Record<string, string>
    success: (response: { statusCode: number; tempFilePath: string }) => void
    fail: (error: { errMsg?: string }) => void
  }): unknown
  getFileSystemManager?(): {
    writeFile(options: {
      filePath: string
      data: string
      encoding: 'base64'
      success: () => void
      fail: (error: { errMsg?: string }) => void
    }): void
  }
} | undefined

interface ImageBase64Payload {
  mode?: string
  mimeType?: string
  fileName?: string
  base64?: string
  dataUrl?: string
}

function getWechatCloud() {
  return (typeof wx === 'undefined' ? undefined : wx.cloud)
}

export function isCloudTransport() {
  return apiTransport === 'cloud'
}

function authHeaders() {
  const token = getToken()
  return {
    ...(token ? { token, Authorization: `Bearer ${token}` } : {}),
  }
}

function normalizeCloudPath(url: string) {
  return url.replace(/^\/+/, '')
}

function joinURL(base: string, url: string) {
  if (/^(https?:|wxfile:|blob:)/.test(url)) {
    return url
  }
  return `${base.replace(/\/+$/, '')}/${normalizeCloudPath(url)}`
}

function fileRequestURL(url: string) {
  return apiTransport === 'cloud' ? joinURL(cloudFileBaseURL, url) : joinURL(baseURL, url)
}

function fileHeaders(contentType?: string) {
  return {
    ...(contentType ? { 'content-type': contentType } : {}),
    ...(apiTransport === 'cloud' && cloudService ? { 'X-WX-SERVICE': cloudService } : {}),
    ...authHeaders(),
  }
}

async function sendRequest<TData, TBody extends RequestData>(
  url: string,
  method: HttpMethod,
  data?: TBody,
): Promise<ApiRawResponse<TData>> {
  if (apiTransport === 'cloud') {
    if (!cloudEnv || !cloudService) {
      throw new Error('云托管配置缺失')
    }
    const wechatCloud = getWechatCloud()
    if (!wechatCloud?.callContainer) {
      throw new Error('当前环境不支持云托管调用')
    }

    return wechatCloud.callContainer<TData>({
      config: {
        env: cloudEnv,
      },
      path: normalizeCloudPath(url),
      header: {
        'content-type': 'application/json',
        'X-WX-SERVICE': cloudService,
        ...authHeaders(),
      },
      method,
      data,
    })
  }

  const response = await uni.request({
    url: `${baseURL}${url}`,
    method,
    data,
    header: {
      'content-type': 'application/json',
      ...authHeaders(),
    },
  })

  return {
    statusCode: response.statusCode,
    data: response.data as ApiResponse<TData>,
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
    const response = await sendRequest<TData, TBody>(url, method, data)
    const payload = response.data
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
    const response = apiTransport === 'cloud'
      ? await uploadByWechat(url, filePath, formData, name)
      : await uni.uploadFile({
          url: fileRequestURL(url),
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

export async function uploadToCloudStorage(filePath: string, cloudPath: string) {
  if (!cloudEnv) {
    throw new Error('云存储配置缺失')
  }
  const wechatCloud = getWechatCloud()
  if (!wechatCloud?.uploadFile || !wechatCloud.getTempFileURL) {
    throw new Error('当前环境不支持云存储上传')
  }

  const uploadResult = await wechatCloud.uploadFile({
    cloudPath,
    filePath,
  })
  if (!uploadResult.fileID) {
    throw new Error('云存储上传失败')
  }

  const tempURLResult = await wechatCloud.getTempFileURL({
    fileList: [uploadResult.fileID],
  })
  const file = tempURLResult.fileList?.[0]
  if (!file?.tempFileURL || file.status !== 0) {
    throw new Error(file?.errMsg || '获取图片访问地址失败')
  }

  return {
    fileID: uploadResult.fileID,
    tempFileURL: file.tempFileURL,
  }
}

function uploadByWechat(
  url: string,
  filePath: string,
  formData: Record<string, string | number>,
  name: string,
) {
  if (typeof wx === 'undefined' || !wx.uploadFile) {
    throw new Error('当前环境不支持微信文件上传')
  }
  return new Promise<{ statusCode: number; data: string }>((resolve, reject) => {
    wx.uploadFile?.({
      url: fileRequestURL(url),
      filePath,
      name,
      formData,
      header: fileHeaders(),
      success: resolve,
      fail: (error) => reject(new Error(error.errMsg || '上传失败')),
    })
  })
}

export async function download(url: string) {
  if (!url || /^(wxfile:|blob:|data:)/.test(url)) {
    return url
  }
  if (apiTransport !== 'cloud' && /^https?:/.test(url)) {
    return url
  }
  const base64Path = await downloadBase64Image(url)
  if (base64Path) {
    return base64Path
  }
  const targetURL = fileRequestURL(url)
  const response = apiTransport === 'cloud'
    ? await downloadByWechat(targetURL)
    : await uni.downloadFile({
        url: targetURL,
        header: authHeaders(),
      })

  if (response.statusCode < 200 || response.statusCode >= 300) {
    throw new Error('下载失败')
  }
  return response.tempFilePath
}

function downloadByWechat(url: string) {
  if (typeof wx === 'undefined' || !wx.downloadFile) {
    throw new Error('当前环境不支持微信文件下载')
  }
  return new Promise<{ statusCode: number; tempFilePath: string }>((resolve, reject) => {
    wx.downloadFile?.({
      url,
      header: fileHeaders(),
      success: resolve,
      fail: (error) => reject(new Error(error.errMsg || '下载失败')),
    })
  })
}

async function downloadBase64Image(url: string) {
  try {
    const payload = await request<ImageBase64Payload>(url)
    if (payload?.mode !== 'base64' || !payload.base64) {
      return ''
    }
    return writeBase64Image(payload.base64, payload.mimeType, payload.fileName)
  } catch {
    return ''
  }
}

function writeBase64Image(base64: string, mimeType = 'image/jpeg', fileName = '') {
  if (typeof wx === 'undefined' || !wx.getFileSystemManager || !wx.env?.USER_DATA_PATH) {
    return Promise.resolve(`data:${mimeType};base64,${base64}`)
  }
  const ext = fileExtensionFromMime(mimeType, fileName)
  const filePath = `${wx.env.USER_DATA_PATH}/img-${Date.now()}-${Math.random().toString(16).slice(2)}${ext}`
  return new Promise<string>((resolve, reject) => {
    wx.getFileSystemManager?.().writeFile({
      filePath,
      data: base64,
      encoding: 'base64',
      success: () => resolve(filePath),
      fail: (error) => reject(new Error(error.errMsg || '图片写入失败')),
    })
  })
}

function fileExtensionFromMime(mimeType: string, fileName: string) {
  const ext = fileName.match(/\.[a-z0-9]+$/i)?.[0]
  if (ext) return ext
  if (mimeType.includes('png')) return '.png'
  if (mimeType.includes('webp')) return '.webp'
  if (mimeType.includes('gif')) return '.gif'
  return '.jpg'
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
  download,
}
