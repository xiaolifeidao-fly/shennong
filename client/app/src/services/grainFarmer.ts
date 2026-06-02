import { http } from './request'
import { buildQuery } from './query'
import type { PageResponse } from '@/types/api'
import type { GrainDraftCardImage, GrainFarmerDTO } from '@/types/grain'

export interface GrainFarmerQuery {
  page?: number
  pageIndex?: number
  pageSize?: number
  stationId?: number
  appUserId?: number
  search?: string
  status?: string
}

export function listGrainFarmers(query: GrainFarmerQuery = {}) {
  return http.get<PageResponse<GrainFarmerDTO>>(`/grain-farmers${buildQuery(query)}`)
}

export function createGrainFarmer(data: Partial<GrainFarmerDTO>) {
  return http.post<GrainFarmerDTO, Partial<GrainFarmerDTO>>('/grain-farmers', data)
}

export function updateGrainFarmer(id: string | number, data: Partial<GrainFarmerDTO>) {
  return http.put<GrainFarmerDTO, Partial<GrainFarmerDTO>>(`/grain-farmers/${id}`, data)
}

export interface FarmerImagesResult {
  idCardFront: string
  idCardBack: string
  bankCard: string
}

export function getGrainFarmerImages(farmerId: string | number) {
  return http.get<FarmerImagesResult>(`/grain-farmer-images?farmerId=${farmerId}`)
}

export function saveGrainFarmerImage(farmerId: string | number, image: GrainDraftCardImage) {
  return http.post<unknown, GrainDraftCardImage & { farmerId: string | number; imageName?: string }>('/grain-farmer-images', {
    ...image,
    farmerId,
    imageName: image.fileName,
  })
}
