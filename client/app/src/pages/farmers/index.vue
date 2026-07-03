<template>
  <view class="page farmers-page">
    <FarmerFilterBar
      v-model:keyword="keyword"
      v-model:custom-start="customStart"
      v-model:custom-end="customEnd"
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
        @delete="confirmDeleteFarmer"
      />
    </view>

  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onReachBottom, onShow } from '@dcloudio/uni-app'
import FarmerFilterBar from './components/FarmerFilterBar.vue'
import SectionHeader from '@/components/business/SectionHeader.vue'
import FarmerSummaryCard from '@/components/business/FarmerSummaryCard.vue'
import { useGrainStore, type FarmerSummaryDateRange } from '@/stores/grain'

const grainStore = useGrainStore()
const keyword = ref('')
const activeFilter = ref('今天')
const customStart = ref('')
const customEnd = ref('')

const filteredFarmers = computed(() => {
  return grainStore.farmerSummaries.filter((farmer) => {
    const matchFilter = activeFilter.value !== '资料待补' || farmer.status === 'missing-bank'
    return matchFilter
  })
})

onShow(() => {
  reloadFarmerSummaries(true)
})

onReachBottom(() => {
  reloadFarmerSummaries(false)
})

watch(activeFilter, (filter) => {
  if (filter === '自定义') {
    initCustomRange()
  }
  if (grainStore.dailySummaryLoading) {
    // 正在加载中，等加载完成后补发
    const stop = watch(
      () => grainStore.dailySummaryLoading,
      (loading) => {
        if (!loading) {
          stop()
          reloadFarmerSummaries(true)
        }
      },
    )
  } else {
    reloadFarmerSummaries(true)
  }
})

function handleSearch() {
  reloadFarmerSummaries(true)
}

function todayStr() {
  const date = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function initCustomRange() {
  const today = todayStr()
  customStart.value ||= today
  customEnd.value ||= customStart.value
}

function getCustomRange(): FarmerSummaryDateRange | undefined {
  if (activeFilter.value !== '自定义') {
    return undefined
  }
  initCustomRange()
  return {
    startDate: customStart.value,
    endDate: customEnd.value,
  }
}

function reloadFarmerSummaries(force: boolean) {
  void grainStore.loadTodayFarmerSummaries(force, keyword.value, activeFilter.value, getCustomRange())
}

function showFarmer(farmerId: string) {
  grainStore.selectFarmer(farmerId)
  uni.navigateTo({ url: '/pages/farmers/detail' })
}

function goNewEntry() {
  grainStore.selectFarmer('new')
  uni.switchTab({ url: '/pages/entry/index' })
}

function confirmDeleteFarmer(farmerId: string) {
  const farmer = grainStore.farmerSummaries.find((item) => item.id === farmerId)
  uni.showModal({
    title: '删除农户',
    content: `确认删除${farmer?.name ? `「${farmer.name}」` : '该农户'}的信息？删除后农户汇总中将不再显示。`,
    confirmText: '删除',
    confirmColor: '#c2412d',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await grainStore.deleteFarmer(farmerId)
        uni.showToast({ title: '已删除农户', icon: 'success' })
      } catch (error) {
        uni.showToast({ title: error instanceof Error ? error.message : '删除失败', icon: 'none' })
      }
    },
  })
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
