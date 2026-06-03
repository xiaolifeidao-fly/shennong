import { http } from './request'
import { buildQuery } from './query'
import type { GrainDashboard } from '@/types/grain'

export interface GrainDashboardQuery {
  startDate?: string
  endDate?: string
  stationId?: number
}

export function getGrainDashboard(query: GrainDashboardQuery = {}) {
  return http.get<GrainDashboard>(`/grain-purchase-dashboard${buildQuery(query)}`)
}
