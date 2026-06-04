export type FarmerStatus = 'complete' | 'missing-bank' | 'inactive'

export interface FarmerProfile {
  id: string
  stationId: number
  appUserId: number
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
  stationId: number
  appUserId: number
  farmerId: string
  purchaseTypeId: number
  crop: string
  quantity: number
  unit: string
  displayUnit: string
  amount: number
  buyTime: string
  payTime: string
  placeId: number
  place: string
  locationName: string
  locationAddress: string
  longitude: string
  latitude: string
  province: string
  city: string
  district: string
  paymentMethodId: number
  payType: string
  materialImages: string[]
  editLogs: GrainEntryEditLog[]
}

export interface GrainEntryEditLog {
  id: string
  time: string
  operator: string
  summary: string
}

export interface GrainPreset {
  salesmanName: string
  crops: string[]
  payTypes: string[]
  places: string[]
  purchaseTypes: GrainPurchaseType[]
  paymentMethods: GrainPaymentMethod[]
  purchasePlaces: GrainPurchasePlace[]
}

export interface GrainEntryDraft {
  farmerId: string
  farmerName: string
  idNumber: string
  phone: string
  address: string
  bankNumber: string
  bankName: string
  purchaseTypeId: number
  crop: string
  quantity: number
  unit: string
  amount: number
  buyTime: string
  payTime: string
  placeId: number
  place: string
  locationName: string
  locationAddress: string
  longitude: string
  latitude: string
  province: string
  city: string
  district: string
  paymentMethodId: number
  paymentMethodCode: string
  payType: string
  materialImages: string[]
  cardImages?: GrainDraftCardImages
}

export interface GrainDraftCardImage {
  cardType: GrainCardOcrType
  imageSide?: 'front' | 'back'
  ossBucket?: string
  ossObjectKey?: string
  ossUrl?: string
  fileName?: string
  fileSize?: number
  mimeType?: string
}

export interface GrainDraftCardImages {
  idCardFront?: GrainDraftCardImage
  idCardBack?: GrainDraftCardImage
  bankCard?: GrainDraftCardImage
}

export interface FarmerSummary extends FarmerProfile {
  entryCount: number
  totalQuantity: number
  totalAmount: number
  avgPrice: number
  latestTime: string
  mainCrop: string
}

export interface GrainFarmerPurchaseSummary {
  id: number
  stationId: number
  appUserId: number
  purchaseTypeId: number
  crop: string
  summaryDate: string
  farmerId: string
  paymentMethodId: number
  payType: string
  entryCount: number
  totalAmount: number
  totalQuantity: number
}

export interface GrainStation {
  id: number
  stationName: string
  stationCode: string
  contactName: string
  contactPhone: string
  province: string
  city: string
  district: string
  address: string
  longitude: string
  latitude: string
  status: string
  remark: string
}

export interface GrainStationDetail {
  id: number
  stationName: string
  stationCode: string
  tenantId: number
  tenantName: string
  contactName: string
  contactPhone: string
  province: string
  city: string
  district: string
  address: string
  longitude: string
  latitude: string
  status: string
  remark: string
  businessLicenseUrl: string
}

export interface GrainPurchaseType {
  id: number
  stationId: number
  typeName: string
  unit: string
  sortOrder: number
  status: string
}

export interface GrainPaymentMethod {
  id: number
  methodCode: string
  methodName: string
  sortOrder: number
  status: string
}

export interface GrainPurchasePlace {
  id: number
  appUserId: number
  placeName: string
  placeType: string
  province: string
  city: string
  district: string
  address: string
  longitude: string
  latitude: string
  sortOrder: number
  status: string
}

export interface GrainFarmerDTO {
  id: number
  stationId: number
  appUserId: number
  name: string
  idNumber: string
  phone: string
  address: string
  bankNumber: string
  bankName: string
  status: FarmerStatus
  statusText: string
  remark: string
}

export interface GrainPurchaseEntryDTO {
  id: number
  stationId: number
  appUserId: number
  farmerId: number
  purchaseTypeId: number
  crop: string
  quantity: number
  unit: string
  displayUnit: string
  amount: number
  unitPrice: number
  buyTime: string
  payTime: string
  placeId: number
  place: string
  locationName: string
  locationAddress: string
  longitude: string
  latitude: string
  province: string
  city: string
  district: string
  paymentMethodId: number
  payType: string
  status: string
  version: number
  remark: string
}

export interface GrainEntrySnapshotDTO {
  id: number
  entryId: number
  snapshotVersion: number
  snapshotAction: string
  snapshotTime: string
  operatorAppUserId: number
  operatorName: string
  stationId: number
  appUserId: number
  farmerId: number
  farmerName: string
  farmerIdNumber: string
  farmerPhone: string
  farmerAddress: string
  farmerBankNumber: string
  farmerBankName: string
  purchaseTypeId: number
  crop: string
  quantity: number
  unit: string
  amount: number
  unitPrice: number
  buyTime: string
  payTime: string
  placeId: number
  place: string
  locationName: string
  locationAddress: string
  longitude: string
  latitude: string
  province: string
  city: string
  district: string
  paymentMethodId: number
  payType: string
  entryStatus: string
  entryRemark: string
}

export interface GrainFarmerPurchaseSummaryDTO {
  id: number
  stationId: number
  appUserId: number
  purchaseTypeId: number
  crop: string
  summaryDate: string
  farmerId: number
  paymentMethodId: number
  payType: string
  entryCount: number
  totalAmount: number
  totalQuantity: number
}

export interface GrainFarmerDailySummaryDTO {
  id: number
  stationId: number
  appUserId: number
  farmerId: number
  farmerName: string
  farmerIdNumber: string
  farmerPhone: string
  farmerAddress: string
  bankNumber: string
  bankName: string
  status: FarmerStatus
  statusText: string
  summaryDate: string
  mainCrop: string
  entryCount: number
  totalAmount: number
  totalQuantity: number
  latestTime: string
}

export interface GrainEntryMaterialDTO {
  id: number
  stationId: number
  entryId: number
  farmerId: number
  appUserId: number
  materialBizType: string
  materialType: string
  ossBucket: string
  ossObjectKey: string
  ossUrl: string
  imageUrl: string
  fileName: string
  fileSize: number
  mimeType: string
  etag: string
  sortOrder: number
}

export interface GrainDashboardOverview {
  newFarmerCount: number
  activeUserCount: number
  totalQuantity: number
  totalAmount: number
  entryCount: number
  averageUnitPrice: number
  averageFarmerDeal: number
}

export interface GrainDashboardDimension {
  key: string
  name: string
  stationId: number
  purchaseTypeId: number
  entryCount: number
  farmerCount: number
  activeUserCount: number
  totalQuantity: number
  totalAmount: number
  averageUnitPrice: number
  amountShare: number
  quantityShare: number
}

export interface GrainDashboard {
  startDate: string
  endDate: string
  stationId: number
  overview: GrainDashboardOverview
  byStation: GrainDashboardDimension[]
  byCrop: GrainDashboardDimension[]
  generated?: string
}

export type GrainCardOcrType = 'id-card' | 'bank-card'

export interface GrainCardOcrResult {
  cardType: GrainCardOcrType
  mock: boolean
  ossBucket: string
  ossObjectKey: string
  ossUrl: string
  fileName: string
  fileSize: number
  mimeType: string
  name?: string
  idNumber?: string
  address?: string
  bankNumber?: string
  bankName?: string
}
