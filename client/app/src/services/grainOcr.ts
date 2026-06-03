import { http, isCloudTransport, uploadToCloudStorage } from './request'
import type { GrainCardOcrResult, GrainCardOcrType } from '@/types/grain'

export type IDCardSide = 'front' | 'back'
const cloudStoragePrefix = import.meta.env.VITE_APP_CLOUD_STORAGE_PREFIX || 'grain-card-ocr'

export function recognizeGrainCard(
  filePath: string,
  cardType: GrainCardOcrType,
  options: { farmerId?: string | number; imageSide?: IDCardSide } = {},
) {
  const formData: Record<string, string | number> = { cardType }
  if (options.farmerId) {
    formData.farmerId = String(options.farmerId)
  }
  if (options.imageSide) {
    formData.imageSide = options.imageSide
  }
  if (!isCloudTransport()) {
    return http.upload<GrainCardOcrResult>('/grain-card-ocr/recognize', filePath, formData)
  }
  return recognizeCloudStorageCard(filePath, formData)
}

async function recognizeCloudStorageCard(filePath: string, formData: Record<string, string | number>) {
  uni.showLoading({ title: '识别中', mask: true })
  try {
    const fileName = filePath.split('/').pop() || `card-${Date.now()}.jpg`
    const cloudPath = buildCloudPath(String(formData.cardType), fileName)
    const cloudFile = await uploadToCloudStorage(filePath, cloudPath)
    return await http.post<GrainCardOcrResult>('/grain-card-ocr/recognize', {
      ...formData,
      fileName,
      cloudFileId: cloudFile.fileID,
      imageUrl: cloudFile.tempFileURL,
    })
  } finally {
    uni.hideLoading()
  }
}

function buildCloudPath(cardType: string, fileName: string) {
  const ext = normalizeImageExt(fileName)
  const now = new Date()
  const date = [
    now.getFullYear(),
    String(now.getMonth() + 1).padStart(2, '0'),
    String(now.getDate()).padStart(2, '0'),
  ].join('')
  const stamp = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  return `${cloudStoragePrefix}/${cardType}/${date}/${stamp}${ext}`
}

function normalizeImageExt(fileName: string) {
  const match = fileName.toLowerCase().match(/\.(jpe?g|png|webp|bmp)$/)
  return match ? match[0] : '.jpg'
}
