import { http } from './request'
import { buildQuery } from './query'
import { DEFAULT_GRAIN_STATION_ID } from '@/config/app'
import type { PageResponse } from '@/types/api'
import type { GrainPaymentMethod, GrainPurchasePlace, GrainPurchaseType, GrainStation } from '@/types/grain'

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

export function createGrainStation(data: Partial<GrainStation>) {
  return http.post<GrainStation, Partial<GrainStation>>('/grain-stations', data)
}

export function listGrainPurchaseTypes(stationId = DEFAULT_GRAIN_STATION_ID) {
  return http.get<GrainPurchaseType[]>(`/grain-purchase-types${buildQuery({ stationId })}`)
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

export function listGrainPurchasePlaces(appUserId = 0) {
  return http.get<GrainPurchasePlace[]>(`/grain-purchase-places${buildQuery({ appUserId })}`)
}

export function createGrainPurchasePlace(data: Partial<GrainPurchasePlace>) {
  return http.post<GrainPurchasePlace, Partial<GrainPurchasePlace>>('/grain-purchase-places', data)
}
