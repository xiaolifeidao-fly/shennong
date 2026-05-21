<template>
  <view class="page entry-page">
    <EntryNotice />

    <SectionHeader title="新增收粮录入" action-text="查看今日汇总" @action="goFarmers" />
    <FarmerIdentityForm
      v-model="draft"
      :farmers="grainStore.farmers"
      @farmer-change="handleFarmerChange"
      @scan-id="applyIdScan"
      @scan-bank="applyBankScan"
    />

    <SectionHeader title="本笔粮食信息" />
    <GrainPurchaseForm v-model="draft" :preset="grainStore.preset" @submit="saveEntry" />
  </view>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import EntryNotice from './components/EntryNotice.vue'
import FarmerIdentityForm from './components/FarmerIdentityForm.vue'
import GrainPurchaseForm from './components/GrainPurchaseForm.vue'
import SectionHeader from '@/components/business/SectionHeader.vue'
import { useGrainStore } from '@/stores/grain'
import type { GrainEntryDraft } from '@/types/grain'

const grainStore = useGrainStore()
const draft = ref<GrainEntryDraft>(grainStore.createEntryDraft())

onShow(() => {
  draft.value = grainStore.createEntryDraft()
})

watch(
  () => grainStore.selectedFarmerId,
  (farmerId) => {
    draft.value = grainStore.createEntryDraft(farmerId)
  },
)

function handleFarmerChange(farmerId: string) {
  grainStore.selectFarmer(farmerId)
  draft.value = grainStore.createEntryDraft(farmerId)
}

function applyIdScan() {
  draft.value = { ...draft.value, ...grainStore.simulateIdCardScan() }
  uni.showToast({ title: '已模拟识别身份证', icon: 'none' })
}

function applyBankScan() {
  draft.value = { ...draft.value, ...grainStore.simulateBankCardScan() }
  uni.showToast({ title: '已模拟识别银行卡', icon: 'none' })
}

function saveEntry() {
  if (!draft.value.farmerName || !draft.value.quantity || !draft.value.amount) {
    uni.showToast({ title: '请补全农户、数量和金额', icon: 'none' })
    return
  }

  grainStore.saveEntry(draft.value)
  uni.showToast({ title: '保存成功', icon: 'success' })
  draft.value = grainStore.createEntryDraft(grainStore.selectedFarmerId)
}

function goFarmers() {
  uni.switchTab({ url: '/pages/farmers/index' })
}
</script>

<style lang="scss" scoped>
.entry-page {
  display: flex;
  flex-direction: column;
}
</style>
