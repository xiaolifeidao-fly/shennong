<template>
  <view class="page farmers-page">
    <FarmerFilterBar
      v-model:keyword="keyword"
      :active-filter="activeFilter"
      @filter-change="activeFilter = $event"
      @search="handleSearch"
    />

    <SectionHeader title="我的农户汇总" action-text="新增" @action="goNewEntry" />
    <view v-if="grainStore.farmersLoading || grainStore.dailySummaryLoading" class="loading-strip">正在加载农户汇总...</view>
    <view class="list">
      <FarmerSummaryCard
        v-for="farmer in filteredFarmers"
        :key="farmer.id"
        :farmer="farmer"
        compact
        @select="showFarmer"
      />
    </view>

    <FarmerDetailSheet
      :visible="detailVisible"
      :farmer="currentFarmer"
      :entries="currentEntries"
      :summaries="currentSummaries"
      :entries-loading="grainStore.entriesLoading"
      :summaries-loading="grainStore.summariesLoading"
      :editing="editing"
      @close="closeDetail"
      @edit="editing = $event"
      @save="saveFarmer"
      @entry="entryForFarmer"
    />
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onReachBottom, onShow } from '@dcloudio/uni-app'
import FarmerFilterBar from './components/FarmerFilterBar.vue'
import FarmerDetailSheet from './components/FarmerDetailSheet.vue'
import SectionHeader from '@/components/business/SectionHeader.vue'
import FarmerSummaryCard from '@/components/business/FarmerSummaryCard.vue'
import { useGrainStore } from '@/stores/grain'
import type { FarmerProfile } from '@/types/grain'

const grainStore = useGrainStore()
const keyword = ref('')
const activeFilter = ref('今天')
const detailVisible = ref(false)
const editing = ref(false)

const currentFarmer = computed(() => grainStore.selectedFarmer)
const currentEntries = computed(() => grainStore.entries.filter((entry) => entry.farmerId === grainStore.selectedFarmerId))
const currentSummaries = computed(() => grainStore.summaries.filter((summary) => summary.farmerId === grainStore.selectedFarmerId))
const filteredFarmers = computed(() => {
  const key = keyword.value.trim()
  return grainStore.farmerSummaries.filter((farmer) => {
    const matchFilter = activeFilter.value !== '资料待补' || farmer.status === 'missing-bank'
    const matchKeyword = !key || [farmer.name, farmer.idNumber, farmer.phone].some((value) => value.includes(key))
    return matchFilter && matchKeyword
  })
})

onShow(() => {
  void grainStore.loadTodayFarmerSummaries(true, keyword.value)
  if (grainStore.selectedFarmerId !== 'new' && currentFarmer.value) {
    detailVisible.value = false
  }
})

onReachBottom(() => {
  void grainStore.loadTodayFarmerSummaries(false, keyword.value)
})

function handleSearch() {
  void grainStore.loadTodayFarmerSummaries(true, keyword.value)
}

function showFarmer(farmerId: string) {
  grainStore.selectFarmer(farmerId)
  editing.value = false
  detailVisible.value = true
  void grainStore.loadEntries(true, farmerId)
  void grainStore.loadSummaries(true, farmerId)
}

function closeDetail() {
  detailVisible.value = false
  editing.value = false
}

async function saveFarmer(farmer: FarmerProfile) {
  try {
    await grainStore.updateFarmer(farmer.id, farmer)
    editing.value = false
    uni.showToast({ title: '已保存农户资料', icon: 'success' })
  } catch (error) {
    uni.showToast({ title: error instanceof Error ? error.message : '保存失败', icon: 'none' })
  }
}

function entryForFarmer(farmerId: string) {
  grainStore.selectFarmer(farmerId)
  closeDetail()
  uni.switchTab({ url: '/pages/entry/index' })
}

function goNewEntry() {
  grainStore.selectFarmer('new')
  uni.switchTab({ url: '/pages/entry/index' })
}
</script>

<style lang="scss" scoped>
.farmers-page {
  display: flex;
  flex-direction: column;
}

.list {
  display: grid;
  gap: 20rpx;
}

.loading-strip {
  padding: 18rpx 22rpx;
  border-radius: 8rpx;
  background: #f4f8f3;
  color: #52605a;
  font-size: 24rpx;
}
</style>
