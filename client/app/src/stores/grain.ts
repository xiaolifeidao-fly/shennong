import { defineStore } from 'pinia'
import { useUserStore } from './user'
import {
  listGrainPaymentMethods,
  listGrainPurchasePlaces,
  listGrainPurchaseTypes,
  listGrainStations,
} from '@/services/grainConfig'
import { createGrainFarmer, listGrainFarmers, updateGrainFarmer } from '@/services/grainFarmer'
import {
  createGrainEntryMaterial,
  listGrainFarmerDailySummaries,
  createGrainPurchaseEntry,
  listGrainFarmerPurchaseSummaries,
  listGrainPurchaseEntries,
  updateGrainPurchaseEntry,
} from '@/services/grainPurchase'
import { recognizeGrainCard, type IDCardSide } from '@/services/grainOcr'
import { DEFAULT_GRAIN_STATION_ID } from '@/config/app'
import type {
  FarmerProfile,
  FarmerStatus,
  FarmerSummary,
  GrainEntry,
  GrainEntryDraft,
  GrainFarmerDailySummaryDTO,
  GrainFarmerPurchaseSummary,
  GrainFarmerPurchaseSummaryDTO,
  GrainFarmerDTO,
  GrainPaymentMethod,
  GrainPreset,
  GrainPurchaseEntryDTO,
  GrainPurchasePlace,
  GrainPurchaseType,
} from '@/types/grain'

interface NullablePage<T> {
  data?: T[] | null
}

interface GrainState {
  farmers: FarmerProfile[]
  dailyFarmerSummaries: FarmerSummary[]
  entries: GrainEntry[]
  summaries: GrainFarmerPurchaseSummary[]
  preset: GrainPreset
  selectedFarmerId: string
  stationId: number
  loading: boolean
  presetLoading: boolean
  farmersLoading: boolean
  entriesLoading: boolean
  summariesLoading: boolean
  dailySummaryLoading: boolean
  initialized: boolean
  presetLoaded: boolean
  farmersLoaded: boolean
  entriesLoaded: boolean
  summariesLoaded: boolean
  dailySummaryLoaded: boolean
  dailySummaryPageIndex: number
  dailySummaryPageSize: number
  dailySummaryTotal: number
}

function safeArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : []
}

function pageItems<T>(page: NullablePage<T> | null | undefined): T[] {
  return safeArray(page?.data)
}

const defaultPreset: GrainPreset = {
  salesmanName: '当前业务员',
  crops: [],
  payTypes: [],
  places: [],
  purchaseTypes: [],
  paymentMethods: [],
  purchasePlaces: [],
}

function formatDateTime(value?: string) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  const pad = (num: number) => String(num).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function nowText() {
  return formatDateTime(new Date().toISOString())
}

function getPaymentInfoStatusText(hasPaymentAccount: boolean) {
  return hasPaymentAccount ? '资料完整' : '收款信息待补'
}

function toServerTime(value: string) {
  if (!value) {
    return new Date().toISOString()
  }
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const date = new Date(normalized)
  return Number.isNaN(date.getTime()) ? new Date().toISOString() : date.toISOString()
}

function toFarmerProfile(dto: GrainFarmerDTO): FarmerProfile {
  return {
    id: String(dto.id),
    stationId: Number(dto.stationId) || 0,
    appUserId: Number(dto.appUserId) || 0,
    name: dto.name || '',
    idNumber: dto.idNumber || '',
    phone: dto.phone || '',
    address: dto.address || '',
    bankNumber: dto.bankNumber || '',
    bankName: dto.bankName || '',
    status: (dto.status || 'missing-bank') as FarmerStatus,
    statusText: dto.statusText || getPaymentInfoStatusText(Boolean(dto.bankNumber)),
  }
}

function toEntry(dto: GrainPurchaseEntryDTO): GrainEntry {
  return {
    id: String(dto.id),
    stationId: Number(dto.stationId) || 0,
    appUserId: Number(dto.appUserId) || 0,
    farmerId: String(dto.farmerId),
    purchaseTypeId: Number(dto.purchaseTypeId) || 0,
    crop: dto.crop || '',
    quantity: Number(dto.quantity) || 0,
    unit: dto.unit || '公斤',
    amount: Number(dto.amount) || 0,
    buyTime: formatDateTime(dto.buyTime),
    placeId: Number(dto.placeId) || 0,
    place: dto.place || '',
    locationName: dto.locationName || '',
    locationAddress: dto.locationAddress || '',
    longitude: dto.longitude || '',
    latitude: dto.latitude || '',
    province: dto.province || '',
    city: dto.city || '',
    district: dto.district || '',
    paymentMethodId: Number(dto.paymentMethodId) || 0,
    payType: dto.payType || '',
    materialImages: [],
    editLogs: [],
  }
}

function toSummary(dto: GrainFarmerPurchaseSummaryDTO): GrainFarmerPurchaseSummary {
  return {
    id: Number(dto.id) || 0,
    stationId: Number(dto.stationId) || 0,
    appUserId: Number(dto.appUserId) || 0,
    purchaseTypeId: Number(dto.purchaseTypeId) || 0,
    crop: dto.crop || '',
    summaryDate: formatDateTime(dto.summaryDate).slice(0, 10),
    farmerId: String(dto.farmerId),
    paymentMethodId: Number(dto.paymentMethodId) || 0,
    payType: dto.payType || '',
    entryCount: Number(dto.entryCount) || 0,
    totalAmount: Number(dto.totalAmount) || 0,
    totalQuantity: Number(dto.totalQuantity) || 0,
  }
}

function toFarmerDailySummary(dto: GrainFarmerDailySummaryDTO): FarmerSummary {
  const totalQuantity = Number(dto.totalQuantity) || 0
  const totalAmount = Number(dto.totalAmount) || 0

  return {
    id: String(dto.farmerId),
    stationId: Number(dto.stationId) || 0,
    appUserId: Number(dto.appUserId) || 0,
    name: dto.farmerName || '',
    idNumber: dto.farmerIdNumber || '',
    phone: dto.farmerPhone || '',
    address: dto.farmerAddress || '',
    bankNumber: dto.bankNumber || '',
    bankName: dto.bankName || '',
    status: (dto.status || 'missing-bank') as FarmerStatus,
    statusText: dto.statusText || getPaymentInfoStatusText(Boolean(dto.bankNumber)),
    entryCount: Number(dto.entryCount) || 0,
    totalQuantity,
    totalAmount,
    avgPrice: totalQuantity ? totalAmount / totalQuantity : 0,
    latestTime: formatDateTime(dto.latestTime || dto.summaryDate),
    mainCrop: dto.mainCrop || '-',
  }
}

function createDraft(farmer: FarmerProfile | undefined, preset: GrainPreset): GrainEntryDraft {
  const firstPurchaseType = preset.purchaseTypes[0]
  const firstPlace = preset.purchasePlaces[0]
  const firstPaymentMethod = preset.paymentMethods[0]

  return {
    farmerId: farmer?.id || 'new',
    farmerName: farmer?.name || '',
    idNumber: farmer?.idNumber || '',
    phone: farmer?.phone || '',
    address: farmer?.address || '',
    bankNumber: farmer?.bankNumber || '',
    bankName: farmer?.bankName || '',
    purchaseTypeId: firstPurchaseType?.id || 0,
    crop: firstPurchaseType?.typeName || preset.crops[0] || '',
    quantity: 0,
    unit: firstPurchaseType?.unit || '公斤',
    amount: 0,
    buyTime: nowText(),
    placeId: firstPlace?.id || 0,
    place: firstPlace?.placeName || preset.places[0] || '',
    locationName: firstPlace?.placeName || '',
    locationAddress: firstPlace?.address || '',
    longitude: firstPlace?.longitude || '',
    latitude: firstPlace?.latitude || '',
    province: firstPlace?.province || '',
    city: firstPlace?.city || '',
    district: firstPlace?.district || '',
    paymentMethodId: firstPaymentMethod?.id || 0,
    paymentMethodCode: firstPaymentMethod?.methodCode || '',
    payType: firstPaymentMethod?.methodName || preset.payTypes[0] || '',
    materialImages: [],
  }
}

function createDraftFromEntry(entry: GrainEntry, farmer: FarmerProfile | undefined): GrainEntryDraft {
  return {
    farmerId: farmer?.id || entry.farmerId,
    farmerName: farmer?.name || '',
    idNumber: farmer?.idNumber || '',
    phone: farmer?.phone || '',
    address: farmer?.address || '',
    bankNumber: farmer?.bankNumber || '',
    bankName: farmer?.bankName || '',
    purchaseTypeId: entry.purchaseTypeId,
    crop: entry.crop,
    quantity: entry.quantity,
    unit: entry.unit || '公斤',
    amount: entry.amount,
    buyTime: entry.buyTime,
    placeId: entry.placeId,
    place: entry.place,
    locationName: entry.locationName,
    locationAddress: entry.locationAddress,
    longitude: entry.longitude,
    latitude: entry.latitude,
    province: entry.province,
    city: entry.city,
    district: entry.district,
    paymentMethodId: entry.paymentMethodId,
    paymentMethodCode: '',
    payType: entry.payType,
    materialImages: [...(entry.materialImages || [])],
  }
}

function applyMaterialImage(draft: GrainEntryDraft, ossUrl: string): string[] {
  if (!ossUrl || draft.materialImages.includes(ossUrl)) {
    return [...draft.materialImages]
  }
  return [...draft.materialImages, ossUrl]
}

function buildFarmerPayload(draft: GrainEntryDraft, stationId: number): Partial<GrainFarmerDTO> {
  const userStore = useUserStore()

  return {
    stationId,
    appUserId: userStore.profile?.id || 0,
    name: draft.farmerName,
    idNumber: draft.idNumber,
    phone: draft.phone,
    address: draft.address,
    bankNumber: draft.bankNumber,
    bankName: draft.bankName,
    status: draft.bankNumber ? 'complete' : 'missing-bank',
    statusText: getPaymentInfoStatusText(Boolean(draft.bankNumber)),
  }
}

function buildEntryPayload(draft: GrainEntryDraft, farmerId: number, stationId: number): Partial<GrainPurchaseEntryDTO> {
  const userStore = useUserStore()

  return {
    stationId,
    appUserId: userStore.profile?.id || 0,
    farmerId,
    purchaseTypeId: draft.purchaseTypeId || 0,
    crop: draft.crop,
    quantity: Number(draft.quantity) || 0,
    unit: draft.unit || '公斤',
    amount: Number(draft.amount) || 0,
    buyTime: toServerTime(draft.buyTime),
    placeId: draft.placeId || 0,
    place: draft.place,
    locationName: draft.locationName,
    locationAddress: draft.locationAddress,
    longitude: draft.longitude,
    latitude: draft.latitude,
    province: draft.province,
    city: draft.city,
    district: draft.district,
    paymentMethodId: draft.paymentMethodId || 0,
    payType: draft.payType,
    status: 'submitted',
  }
}

function buildPreset(
  purchaseTypes: GrainPurchaseType[],
  paymentMethods: GrainPaymentMethod[],
  purchasePlaces: GrainPurchasePlace[],
): GrainPreset {
  return {
    salesmanName: defaultPreset.salesmanName,
    crops: purchaseTypes.map((item) => item.typeName).filter(Boolean),
    payTypes: paymentMethods.map((item) => item.methodName).filter(Boolean),
    places: purchasePlaces.map((item) => item.placeName).filter(Boolean),
    purchaseTypes,
    paymentMethods,
    purchasePlaces,
  }
}

export const useGrainStore = defineStore('grain', {
  state: (): GrainState => ({
    farmers: [],
    dailyFarmerSummaries: [],
    entries: [],
    summaries: [],
    preset: defaultPreset,
    selectedFarmerId: 'new',
    stationId: DEFAULT_GRAIN_STATION_ID,
    loading: false,
    presetLoading: false,
    farmersLoading: false,
    entriesLoading: false,
    summariesLoading: false,
    dailySummaryLoading: false,
    initialized: false,
    presetLoaded: false,
    farmersLoaded: false,
    entriesLoaded: false,
    summariesLoaded: false,
    dailySummaryLoaded: false,
    dailySummaryPageIndex: 0,
    dailySummaryPageSize: 20,
    dailySummaryTotal: 0,
  }),
  getters: {
    todayEntryCount: (state) => state.dailyFarmerSummaries.reduce((sum, item) => sum + item.entryCount, 0),
    todayTotalQuantity: (state) => state.dailyFarmerSummaries.reduce((sum, item) => sum + item.totalQuantity, 0),
    todayTotalAmount: (state) => state.dailyFarmerSummaries.reduce((sum, item) => sum + item.totalAmount, 0),
    todayFarmerCount: (state) => state.dailySummaryTotal || state.dailyFarmerSummaries.length,
    recentEntries: (state) => state.entries.slice(0, 10),
    farmerSummaries: (state): FarmerSummary[] => state.dailyFarmerSummaries,
    selectedFarmer: (state) =>
      state.farmers.find((farmer) => farmer.id === state.selectedFarmerId) ||
      state.dailyFarmerSummaries.find((farmer) => farmer.id === state.selectedFarmerId),
  },
  actions: {
    async ensureUserAndStation() {
      const userStore = useUserStore()
      if (!userStore.profile) {
        await userStore.refreshProfile().catch(() => null)
      }
      const appUserId = userStore.profile?.id || 0
      if (!this.stationId) {
        const stationPage = await listGrainStations({ pageIndex: 1, pageSize: 1, status: 'active', appUserId: appUserId || undefined })
        this.stationId = pageItems(stationPage)[0]?.id || 0
      }
      return { appUserId, stationId: this.stationId }
    },
    async loadPreset(force = false) {
      if (this.presetLoading || (this.presetLoaded && !force)) {
        return
      }
      this.presetLoading = true
      try {
        const { appUserId, stationId } = await this.ensureUserAndStation()
        const [purchaseTypes, paymentMethods, purchasePlaces] = await Promise.all([
          listGrainPurchaseTypes(stationId),
          listGrainPaymentMethods(),
          listGrainPurchasePlaces(appUserId),
        ])
        this.preset = buildPreset(safeArray(purchaseTypes), safeArray(paymentMethods), safeArray(purchasePlaces))
        this.presetLoaded = true
      } catch (error) {
        uni.showToast({ title: error instanceof Error ? error.message : '录入配置加载失败', icon: 'none' })
      } finally {
        this.presetLoading = false
      }
    },
    async loadFarmers(force = false) {
      if (this.farmersLoading || (this.farmersLoaded && !force)) {
        return
      }
      this.farmersLoading = true
      try {
        const { appUserId, stationId } = await this.ensureUserAndStation()
        const farmerPage = await listGrainFarmers({ pageIndex: 1, pageSize: 200, stationId: stationId || undefined, appUserId: appUserId || undefined })
        this.farmers = pageItems(farmerPage).map(toFarmerProfile)
        if (this.selectedFarmerId !== 'new' && !this.farmers.some((item) => item.id === this.selectedFarmerId)) {
          this.selectedFarmerId = this.farmers[0]?.id || 'new'
        }
        this.farmersLoaded = true
      } catch (error) {
        uni.showToast({ title: error instanceof Error ? error.message : '农户加载失败', icon: 'none' })
      } finally {
        this.farmersLoading = false
      }
    },
    async loadSummaries(force = false, farmerId?: string) {
      if (this.summariesLoading || (this.summariesLoaded && !force && !farmerId)) {
        return
      }
      this.summariesLoading = true
      try {
        const { appUserId, stationId } = await this.ensureUserAndStation()
        const summaryPage = await listGrainFarmerPurchaseSummaries({
          pageIndex: 1,
          pageSize: 200,
          stationId: stationId || undefined,
          appUserId: appUserId || undefined,
          farmerId: farmerId ? Number(farmerId) : undefined,
        })
        const nextSummaries = pageItems(summaryPage).map(toSummary)
        if (farmerId) {
          this.summaries = [...this.summaries.filter((item) => item.farmerId !== farmerId), ...nextSummaries]
        } else {
          this.summaries = nextSummaries
          this.summariesLoaded = true
        }
      } catch (error) {
        uni.showToast({ title: error instanceof Error ? error.message : '农户汇总加载失败', icon: 'none' })
      } finally {
        this.summariesLoading = false
      }
    },
    async loadTodayFarmerSummaries(force = false, search = '') {
      if (this.dailySummaryLoading) {
        return
      }
      if (!force && this.dailySummaryLoaded && this.dailyFarmerSummaries.length >= this.dailySummaryTotal) {
        return
      }
      this.dailySummaryLoading = true
      try {
        const { appUserId, stationId } = await this.ensureUserAndStation()
        const pageIndex = force ? 1 : this.dailySummaryPageIndex + 1
        const today = formatDateTime(new Date().toISOString()).slice(0, 10)
        const summaryPage = await listGrainFarmerDailySummaries({
          pageIndex,
          pageSize: this.dailySummaryPageSize,
          stationId: stationId || undefined,
          appUserId: appUserId || undefined,
          startDate: today,
          endDate: today,
          search: search.trim() || undefined,
        })
        const nextSummaries = pageItems(summaryPage).map(toFarmerDailySummary)
        this.dailyFarmerSummaries = force ? nextSummaries : [...this.dailyFarmerSummaries, ...nextSummaries]
        this.dailySummaryTotal = Number(summaryPage?.total) || this.dailyFarmerSummaries.length
        this.dailySummaryPageIndex = pageIndex
        this.dailySummaryLoaded = true
      } catch (error) {
        uni.showToast({ title: error instanceof Error ? error.message : '今日农户汇总加载失败', icon: 'none' })
      } finally {
        this.dailySummaryLoading = false
      }
    },
    async loadEntries(force = false, farmerId?: string) {
      if (this.entriesLoading || (this.entriesLoaded && !force && !farmerId)) {
        return
      }
      this.entriesLoading = true
      try {
        const { appUserId, stationId } = await this.ensureUserAndStation()
        const entryPage = await listGrainPurchaseEntries({
          pageIndex: 1,
          pageSize: 200,
          stationId: stationId || undefined,
          appUserId: appUserId || undefined,
          farmerId: farmerId ? Number(farmerId) : undefined,
        })
        const nextEntries = pageItems(entryPage).map(toEntry)
        if (farmerId) {
          this.entries = [...this.entries.filter((item) => item.farmerId !== farmerId), ...nextEntries]
        } else {
          this.entries = nextEntries
          this.entriesLoaded = true
        }
      } catch (error) {
        uni.showToast({ title: error instanceof Error ? error.message : '录入明细加载失败', icon: 'none' })
      } finally {
        this.entriesLoading = false
      }
    },
    async loadGrainData(force = false) {
      if (this.loading || (this.initialized && !force)) {
        return
      }
      this.loading = true
      try {
        await Promise.all([
          this.loadPreset(force),
          this.loadFarmers(force),
          this.loadTodayFarmerSummaries(force),
          this.loadEntries(force),
        ])
        this.initialized = true
      } catch (error) {
        uni.showToast({ title: error instanceof Error ? error.message : '粮食数据加载失败', icon: 'none' })
      } finally {
        this.loading = false
      }
    },
    selectFarmer(farmerId: string) {
      this.selectedFarmerId = farmerId
    },
    createEntryDraft(farmerId?: string) {
      const targetFarmerId = farmerId || this.selectedFarmerId
      return createDraft(this.farmers.find((farmer) => farmer.id === targetFarmerId), this.preset)
    },
    createEntryDraftFromHistory(entryId: string) {
      const entry = this.entries.find((item) => item.id === entryId)
      if (!entry) {
        return this.createEntryDraft()
      }

      return createDraftFromEntry(entry, this.farmers.find((farmer) => farmer.id === entry.farmerId))
    },
    simulateIdCardScan() {
      return {
        farmerName: '',
        idNumber: '',
        address: '',
      }
    },
    simulateBankCardScan() {
      return {
        bankNumber: '',
        bankName: '',
      }
    },
    async recognizeIdCard(filePath: string, draft: GrainEntryDraft, side: IDCardSide = 'front') {
      const farmerId = draft.farmerId !== 'new' ? draft.farmerId : undefined
      const result = await recognizeGrainCard(filePath, 'id-card', { farmerId, imageSide: side })
      return {
        farmerName: result.name || draft.farmerName,
        idNumber: result.idNumber || draft.idNumber,
        address: result.address || draft.address,
        materialImages: applyMaterialImage(draft, result.ossUrl),
      }
    },
    async recognizeBankCard(filePath: string, draft: GrainEntryDraft) {
      const farmerId = draft.farmerId !== 'new' ? draft.farmerId : undefined
      const result = await recognizeGrainCard(filePath, 'bank-card', { farmerId })
      return {
        bankNumber: result.bankNumber || draft.bankNumber,
        bankName: result.bankName || draft.bankName,
        materialImages: applyMaterialImage(draft, result.ossUrl),
      }
    },
    async saveEntry(draft: GrainEntryDraft) {
      let farmerId = Number(draft.farmerId)

      if (draft.farmerId === 'new' || !farmerId) {
        const farmer = await createGrainFarmer(buildFarmerPayload(draft, this.stationId))
        farmerId = farmer.id
      } else {
        await updateGrainFarmer(farmerId, buildFarmerPayload(draft, this.stationId))
      }

      const entry = await createGrainPurchaseEntry(buildEntryPayload(draft, farmerId, this.stationId))
      await this.saveEntryMaterials(entry, draft.materialImages)
      this.selectedFarmerId = String(farmerId)
      await Promise.all([this.loadFarmers(true), this.loadTodayFarmerSummaries(true), this.loadSummaries(true, String(farmerId)), this.loadEntries(true, String(farmerId))])
      return toEntry(entry)
    },
    async updateEntry(entryId: string, draft: GrainEntryDraft) {
      const farmerId = Number(draft.farmerId)
      if (!farmerId) {
        return this.saveEntry(draft)
      }

      await updateGrainFarmer(farmerId, buildFarmerPayload(draft, this.stationId))
      const entry = await updateGrainPurchaseEntry(entryId, buildEntryPayload(draft, farmerId, this.stationId))
      await this.saveEntryMaterials(entry, draft.materialImages)
      this.selectedFarmerId = String(farmerId)
      await Promise.all([this.loadFarmers(true), this.loadTodayFarmerSummaries(true), this.loadSummaries(true, String(farmerId)), this.loadEntries(true, String(farmerId))])
      return toEntry(entry)
    },
    async saveEntryMaterials(entry: GrainPurchaseEntryDTO, images: string[]) {
      const tasks = images
        .filter(Boolean)
        .map((image, index) =>
          createGrainEntryMaterial({
            stationId: entry.stationId,
            entryId: entry.id,
            farmerId: entry.farmerId,
            appUserId: entry.appUserId,
            materialBizType: 'entry',
            materialType: 'image',
            ossUrl: image,
            fileName: image.split('/').pop() || `material-${index + 1}`,
            sortOrder: index,
          }),
        )
      await Promise.all(tasks)
    },
    async updateFarmer(farmerId: string, patch: Partial<FarmerProfile>) {
      const result = await updateGrainFarmer(farmerId, {
        stationId: this.stationId,
        appUserId: useUserStore().profile?.id || 0,
        name: patch.name,
        idNumber: patch.idNumber,
        phone: patch.phone,
        address: patch.address,
        bankNumber: patch.bankNumber,
        bankName: patch.bankName,
        status: patch.status,
        statusText: patch.statusText,
      })
      const index = this.farmers.findIndex((farmer) => farmer.id === farmerId)
      if (index >= 0) {
        this.farmers[index] = toFarmerProfile(result)
      }
    },
    updatePreset(preset: GrainPreset) {
      this.preset = preset
    },
  },
})
