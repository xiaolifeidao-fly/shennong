import { defineStore } from 'pinia'
import { getLocalGrainSeed } from '@/services/grain'
import type { FarmerProfile, FarmerSummary, GrainEntry, GrainEntryDraft, GrainPreset } from '@/types/grain'

interface GrainState {
  farmers: FarmerProfile[]
  entries: GrainEntry[]
  preset: GrainPreset
  selectedFarmerId: string
}

const seed = getLocalGrainSeed()

function createDraft(farmer: FarmerProfile | undefined, preset: GrainPreset): GrainEntryDraft {
  return {
    farmerId: farmer?.id || 'new',
    farmerName: farmer?.name || '',
    idNumber: farmer?.idNumber || '',
    phone: farmer?.phone || '',
    address: farmer?.address || '',
    bankNumber: farmer?.bankNumber || '',
    bankName: farmer?.bankName || '',
    crop: preset.crops[0] || '',
    quantity: 4200,
    unit: '斤',
    amount: 5460,
    buyTime: '2026-05-11 10:36',
    place: preset.places[0] || '',
    payType: preset.payTypes[0] || '',
  }
}

export const useGrainStore = defineStore('grain', {
  state: (): GrainState => ({
    farmers: seed.farmers,
    entries: seed.entries,
    preset: seed.preset,
    selectedFarmerId: seed.farmers[0]?.id || 'new',
  }),
  getters: {
    todayEntryCount: (state) => state.entries.length,
    todayTotalQuantity: (state) => state.entries.reduce((sum, item) => sum + item.quantity, 0),
    todayTotalAmount: (state) => state.entries.reduce((sum, item) => sum + item.amount, 0),
    todayFarmerCount: (state) => new Set(state.entries.map((item) => item.farmerId)).size,
    recentEntries: (state) => state.entries.slice(0, 3),
    farmerSummaries: (state): FarmerSummary[] =>
      state.farmers.map((farmer) => {
        const entries = state.entries.filter((item) => item.farmerId === farmer.id)
        const totalQuantity = entries.reduce((sum, item) => sum + item.quantity, 0)
        const totalAmount = entries.reduce((sum, item) => sum + item.amount, 0)
        const latestTime = entries[0]?.buyTime || ''

        return {
          ...farmer,
          entryCount: entries.length,
          totalQuantity,
          totalAmount,
          avgPrice: totalQuantity ? totalAmount / totalQuantity : 0,
          latestTime,
          mainCrop: entries[0]?.crop || '-',
        }
      }),
    selectedFarmer: (state) => state.farmers.find((farmer) => farmer.id === state.selectedFarmerId),
  },
  actions: {
    selectFarmer(farmerId: string) {
      this.selectedFarmerId = farmerId
    },
    createEntryDraft(farmerId?: string) {
      const targetFarmerId = farmerId || this.selectedFarmerId
      return createDraft(this.farmers.find((farmer) => farmer.id === targetFarmerId), this.preset)
    },
    simulateIdCardScan() {
      return {
        farmerName: '李建国',
        idNumber: '410***********3215',
        address: '河南省周口市示范区北城街道丰收村 3 组',
      }
    },
    simulateBankCardScan() {
      return {
        bankNumber: '6228 **** **** 6631',
        bankName: '农商银行北城支行',
      }
    },
    saveEntry(draft: GrainEntryDraft) {
      let farmerId = draft.farmerId

      if (farmerId === 'new') {
        farmerId = `farmer-${Date.now()}`
        this.farmers.unshift({
          id: farmerId,
          name: draft.farmerName || '新农户',
          idNumber: draft.idNumber,
          phone: draft.phone,
          address: draft.address,
          bankNumber: draft.bankNumber,
          bankName: draft.bankName,
          status: draft.bankNumber ? 'complete' : 'missing-bank',
          statusText: draft.bankNumber ? '资料完整' : '银行卡照片待补',
        })
      }

      const entry: GrainEntry = {
        id: `entry-${Date.now()}`,
        farmerId,
        crop: draft.crop,
        quantity: Number(draft.quantity) || 0,
        unit: draft.unit,
        amount: Number(draft.amount) || 0,
        buyTime: draft.buyTime,
        place: draft.place,
        payType: draft.payType,
      }

      this.entries.unshift(entry)
      this.selectedFarmerId = farmerId
      return entry
    },
    updateFarmer(farmerId: string, patch: Partial<FarmerProfile>) {
      const index = this.farmers.findIndex((farmer) => farmer.id === farmerId)
      if (index >= 0) {
        this.farmers[index] = { ...this.farmers[index], ...patch }
      }
    },
    updatePreset(preset: GrainPreset) {
      this.preset = preset
    },
  },
})
