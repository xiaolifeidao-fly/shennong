import { http } from './request'
import type { GrainCardOcrResult, GrainCardOcrType } from '@/types/grain'

export type IDCardSide = 'front' | 'back'

export async function recognizeGrainCard(
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
  return http.upload<GrainCardOcrResult>('/grain-card-ocr/recognize', filePath, formData)
}
