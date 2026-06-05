import { http, isCloudTransport, uploadToCloudStorage } from './request'
import { buildQuery } from './query'
import type { PageResponse } from '@/types/api'
import type {
  GrainEntryMaterialDTO,
  GrainEntrySnapshotDTO,
  GrainFarmerDailySummaryDTO,
  GrainFarmerPurchaseSummaryDTO,
  GrainPurchaseEntryDTO,
} from '@/types/grain'

const cloudStoragePrefix = import.meta.env.VITE_APP_CLOUD_STORAGE_PREFIX || 'grain-entry-materials'

export interface GrainPurchaseEntryQuery {
  page?: number
  pageIndex?: number
  pageSize?: number
  stationId?: number
  appUserId?: number
  farmerId?: number
  search?: string
  status?: string
  startTime?: string
  endTime?: string
}

export interface GrainEntrySnapshotQuery {
  page?: number
  pageIndex?: number
  pageSize?: number
  entryId?: number
  stationId?: number
  appUserId?: number
}

export interface GrainFarmerPurchaseSummaryQuery {
  page?: number
  pageIndex?: number
  pageSize?: number
  stationId?: number
  appUserId?: number
  farmerId?: number
  startDate?: string
  endDate?: string
  search?: string
}

export interface GrainFarmerDailySummaryQuery {
  page?: number
  pageIndex?: number
  pageSize?: number
  stationId?: number
  appUserId?: number
  farmerId?: number
  startDate?: string
  endDate?: string
  search?: string
}

export interface GrainEntryMaterialQuery {
  page?: number
  pageIndex?: number
  pageSize?: number
  stationId?: number
  entryId?: number
  farmerId?: number
  appUserId?: number
  materialBizType?: string
  materialType?: string
}

export function listGrainPurchaseEntries(query: GrainPurchaseEntryQuery = {}) {
  return http.get<PageResponse<GrainPurchaseEntryDTO>>(`/grain-purchase-entries${buildQuery(query)}`)
}

export function createGrainPurchaseEntry(data: Partial<GrainPurchaseEntryDTO>) {
  return http.post<GrainPurchaseEntryDTO, Partial<GrainPurchaseEntryDTO>>('/grain-purchase-entries', data)
}

export function updateGrainPurchaseEntry(id: string | number, data: Partial<GrainPurchaseEntryDTO>) {
  return http.put<GrainPurchaseEntryDTO, Partial<GrainPurchaseEntryDTO>>(`/grain-purchase-entries/${id}`, data)
}

export function voidGrainPurchaseEntry(id: string | number) {
  return http.put<{ voided: boolean }>(`/grain-purchase-entries/${id}/void`)
}

export function listGrainEntrySnapshots(query: GrainEntrySnapshotQuery = {}) {
  return http.get<PageResponse<GrainEntrySnapshotDTO>>(`/grain-entry-snapshots${buildQuery(query)}`)
}

export function listGrainFarmerPurchaseSummaries(query: GrainFarmerPurchaseSummaryQuery = {}) {
  return http.get<PageResponse<GrainFarmerPurchaseSummaryDTO>>(`/grain-farmer-purchase-summaries${buildQuery(query)}`)
}

export function listGrainFarmerDailySummaries(query: GrainFarmerDailySummaryQuery = {}) {
  return http.get<PageResponse<GrainFarmerDailySummaryDTO>>(`/grain-farmer-daily-summaries${buildQuery(query)}`)
}

export function listGrainEntryMaterials(query: GrainEntryMaterialQuery = {}) {
  return http.get<PageResponse<GrainEntryMaterialDTO>>(`/grain-entry-materials${buildQuery(query)}`)
}

export function createGrainEntryMaterial(data: Partial<GrainEntryMaterialDTO>) {
  return http.post<GrainEntryMaterialDTO, Partial<GrainEntryMaterialDTO>>('/grain-entry-materials', data)
}

export async function uploadGrainEntryMaterial(filePath: string, data: Partial<GrainEntryMaterialDTO>) {
  const formData: Record<string, string | number> = {}
  Object.entries(data).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      formData[key] = value as string | number
    }
  })
  if (isCloudTransport()) {
    try {
      const fileName = filePath.split('/').pop() || `material-${Date.now()}.jpg`
      const cloudFile = await uploadToCloudStorage(filePath, buildCloudPath(fileName))
      return await http.post<GrainEntryMaterialDTO>('/grain-entry-materials/upload', {
        ...formData,
        fileName,
        cloudFileId: cloudFile.fileID,
        imageUrl: cloudFile.tempFileURL,
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : ''
      if (!message.includes('不支持') && !message.includes('缺失') && !message.includes('云存储')) {
        throw error
      }
    }
  }
  return http.upload<GrainEntryMaterialDTO>('/grain-entry-materials/upload', filePath, formData)
}

export function deleteGrainEntryMaterial(id: string | number) {
  return http.delete<unknown>(`/grain-entry-materials/${id}`)
}

function buildCloudPath(fileName: string) {
  const ext = normalizeImageExt(fileName)
  const now = new Date()
  const date = [
    now.getFullYear(),
    String(now.getMonth() + 1).padStart(2, '0'),
    String(now.getDate()).padStart(2, '0'),
  ].join('')
  const stamp = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  return `${cloudStoragePrefix}/${date}/${stamp}${ext}`
}

function normalizeImageExt(fileName: string) {
  const match = fileName.toLowerCase().match(/\.(jpe?g|png|webp|bmp)$/)
  return match ? match[0] : '.jpg'
}
