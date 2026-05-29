import { http } from './request'
import type { GrainCardOcrResult, GrainCardOcrType } from '@/types/grain'

export function recognizeGrainCard(filePath: string, cardType: GrainCardOcrType) {
  return http.upload<GrainCardOcrResult>('/grain-card-ocr/recognize', filePath, { cardType })
}
