<template>
  <view class="page home-page">
    <HomeHero
      :entry-count="grainStore.todayEntryCount"
      :total-quantity="grainStore.todayTotalQuantity"
      :total-amount="grainStore.todayTotalAmount"
      :farmer-count="grainStore.todayFarmerCount"
    />

    <SectionHeader title="今日农户汇总" action-text="全部农户" @action="goFarmers" />
    <view v-if="grainStore.farmersLoading || grainStore.dailySummaryLoading" class="loading-strip">正在加载今日汇总...</view>
    <view class="list">
      <FarmerSummaryCard
        v-for="farmer in topFarmers"
        :key="farmer.id"
        :farmer="farmer"
        @select="openFarmer"
      />
    </view>

    <SectionHeader title="最近录入" action-text="继续录入" @action="goEntry" />
    <view v-if="grainStore.entriesLoading" class="loading-strip">正在加载最近录入...</view>
    <RecentEntriesTable :entries="grainStore.recentEntries" :farmers="grainStore.farmers" />
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import HomeHero from './components/HomeHero.vue'
import RecentEntriesTable from './components/RecentEntriesTable.vue'
import SectionHeader from '@/components/business/SectionHeader.vue'
import FarmerSummaryCard from '@/components/business/FarmerSummaryCard.vue'
import { useGrainStore } from '@/stores/grain'

const grainStore = useGrainStore()
const topFarmers = computed(() => grainStore.farmerSummaries.slice(0, 2))

onShow(() => {
  void grainStore.loadTodayFarmerSummaries(true)
  void grainStore.loadFarmers()
  void grainStore.loadEntries(true)
})

function goEntry() {
  uni.switchTab({ url: '/pages/entry/index' })
}

function goFarmers() {
  uni.switchTab({ url: '/pages/farmers/index' })
}

function openFarmer(farmerId: string) {
  grainStore.selectFarmer(farmerId)
  uni.switchTab({ url: '/pages/farmers/index' })
}
</script>

<style lang="scss" scoped>
.home-page {
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
  background: #f6f8f4;
  color: #667266;
  font-size: 24rpx;
}
</style>
