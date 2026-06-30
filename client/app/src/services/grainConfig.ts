import { http } from './request'
import { buildQuery } from './query'
import { normalizeFileUrl } from '@/utils/fileUrl'
import type { PageResponse } from '@/types/api'
import type {
  GrainPaymentMethod,
  GrainPurchasePlace,
  GrainPurchaseType,
  GrainStation,
  GrainStationDetail,
} from '@/types/grain'

export interface GrainStationQuery {
  page?: number
  pageIndex?: number
  pageSize?: number
  appUserId?: number
  search?: string
  status?: string
}

export function listGrainStations(query: GrainStationQuery = {}) {
  return http.get<PageResponse<GrainStation>>(`/grain-stations${buildQuery(query)}`)
}

export function getMyGrainStationDetail() {
  return http.get<GrainStationDetail>('/grain-stations/mine').then((detail) => ({
    ...detail,
    businessLicenseUrl: normalizeFileUrl(detail.businessLicenseUrl),
  }))
}

export function getCurrentSalesmanGrainStation() {
  return http.get<GrainStationDetail>('/grain-stations/mine')
}

export function downloadMyBusinessLicense() {
  return http.get<{ imageUrl: string }>('/grain-stations/mine/extra/business-license').then((result) => normalizeFileUrl(result.imageUrl))
}

export function createGrainStation(data: Partial<GrainStation>) {
  return http.post<GrainStation, Partial<GrainStation>>('/grain-stations', data)
}

export function listGrainPurchaseTypes() {
  return http.get<GrainPurchaseType[]>('/grain-purchase-types')
}

export function createGrainPurchaseType(data: Partial<GrainPurchaseType>) {
  return http.post<GrainPurchaseType, Partial<GrainPurchaseType>>('/grain-purchase-types', data)
}

export function listGrainPaymentMethods() {
  return http.get<GrainPaymentMethod[]>('/grain-payment-methods')
}

export function createGrainPaymentMethod(data: Partial<GrainPaymentMethod>) {
  return http.post<GrainPaymentMethod, Partial<GrainPaymentMethod>>('/grain-payment-methods', data)
}

export function listGrainPurchasePlaces() {
  return http.get<GrainPurchasePlace[]>('/grain-purchase-places')
}

export function createGrainPurchasePlace(data: Partial<GrainPurchasePlace>) {
  return http.post<GrainPurchasePlace, Partial<GrainPurchasePlace>>('/grain-purchase-places', data)
}
