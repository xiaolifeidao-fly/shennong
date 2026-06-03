import { http } from './request'
import { buildQuery } from './query'
import type { PageResponse } from '@/types/api'
import type {
  GrainEntryMaterialDTO,
  GrainEntrySnapshotDTO,
  GrainFarmerDailySummaryDTO,
  GrainFarmerPurchaseSummaryDTO,
  GrainPurchaseEntryDTO,
} from '@/types/grain'

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

export function uploadGrainEntryMaterial(filePath: string, data: Partial<GrainEntryMaterialDTO>) {
  const formData: Record<string, string | number> = {}
  Object.entries(data).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      formData[key] = value as string | number
    }
  })
  return http.upload<GrainEntryMaterialDTO>('/grain-entry-materials/upload', filePath, formData)
}
