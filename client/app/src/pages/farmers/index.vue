<template>
  <view class="page farmers-page">
    <FarmerFilterBar
      v-model:keyword="keyword"
      :active-filter="activeFilter"
      @filter-change="activeFilter = $event"
      @search="handleSearch"
    />

    <SectionHeader title="我的农户汇总" action-text="新增" @action="goNewEntry" />
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
import { onShow } from '@dcloudio/uni-app'
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

const currentFarmer = computed(() => grainStore.farmers.find((farmer) => farmer.id === grainStore.selectedFarmerId))
const currentEntries = computed(() => grainStore.entries.filter((entry) => entry.farmerId === grainStore.selectedFarmerId))
const filteredFarmers = computed(() => {
  const key = keyword.value.trim()
  return grainStore.farmerSummaries.filter((farmer) => {
    const matchFilter = activeFilter.value !== '资料待补' || farmer.status === 'missing-bank'
    const matchKeyword = !key || [farmer.name, farmer.idNumber, farmer.phone].some((value) => value.includes(key))
    return matchFilter && matchKeyword
  })
})

onShow(() => {
  if (grainStore.selectedFarmerId !== 'new' && currentFarmer.value) {
    detailVisible.value = false
  }
})

function handleSearch() {
  uni.showToast({ title: keyword.value ? '已筛选' : '请输入关键词', icon: 'none' })
}

function showFarmer(farmerId: string) {
  grainStore.selectFarmer(farmerId)
  editing.value = false
  detailVisible.value = true
}

function closeDetail() {
  detailVisible.value = false
  editing.value = false
}

function saveFarmer(farmer: FarmerProfile) {
  grainStore.updateFarmer(farmer.id, farmer)
  editing.value = false
  uni.showToast({ title: '已保存农户资料', icon: 'success' })
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
</style>
