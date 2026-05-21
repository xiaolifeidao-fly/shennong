export type FarmerStatus = 'complete' | 'missing-bank'

export interface FarmerProfile {
  id: string
  name: string
  idNumber: string
  phone: string
  address: string
  bankNumber: string
  bankName: string
  status: FarmerStatus
  statusText: string
}

export interface GrainEntry {
  id: string
  farmerId: string
  crop: string
  quantity: number
  unit: string
  amount: number
  buyTime: string
  place: string
  payType: string
}

export interface GrainPreset {
  salesmanName: string
  crops: string[]
  payTypes: string[]
  places: string[]
}

export interface GrainEntryDraft {
  farmerId: string
  farmerName: string
  idNumber: string
  phone: string
  address: string
  bankNumber: string
  bankName: string
  crop: string
  quantity: number
  unit: string
  amount: number
  buyTime: string
  place: string
  payType: string
}

export interface FarmerSummary extends FarmerProfile {
  entryCount: number
  totalQuantity: number
  totalAmount: number
  avgPrice: number
  latestTime: string
  mainCrop: string
}
