<template>
  <view class="card farmer-card" @click="$emit('select', farmer.id)">
    <view class="card-head">
      <view class="card-main">
        <text class="card-title">{{ farmer.name }} · {{ farmer.entryCount }} 笔录入</text>
        <text class="card-meta">{{ metaText }}</text>
      </view>
      <text class="badge" :class="{ warn: farmer.status === 'missing-bank' }">{{ badgeText }}</text>
    </view>

    <view class="data-grid">
      <view class="kv">
        <text>合计数量</text>
        <text class="value">{{ formatQuantity(farmer.totalQuantity) }}</text>
      </view>
      <view class="kv">
        <text>合计金额</text>
        <text class="value">{{ formatAmount(farmer.totalAmount) }}</text>
      </view>
      <view class="kv">
        <text>均价</text>
        <text class="value">{{ farmer.avgPrice.toFixed(2) }} 元/斤</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatAmount, formatQuantity, pickTime } from '@/utils/grain'
import type { FarmerSummary } from '@/types/grain'

const props = defineProps<{
  farmer: FarmerSummary
  compact?: boolean
}>()

defineEmits<{
  select: [farmerId: string]
}>()

const badgeText = computed(() => (props.farmer.status === 'missing-bank' ? '待补' : '完整'))
const metaText = computed(() => {
  if (props.compact) {
    return props.farmer.statusText
  }
  return `身份证 ${props.farmer.idNumber} · 最近 ${pickTime(props.farmer.latestTime)}`
})
</script>

<style lang="scss" scoped>
.farmer-card {
  padding: 28rpx;
}

.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}

.card-main {
  min-width: 0;
  flex: 1;
}

.card-title {
  display: block;
  color: #172018;
  font-size: 30rpx;
  font-weight: 760;
  line-height: 1.35;
}

.card-meta {
  display: block;
  margin-top: 10rpx;
  color: #6d776c;
  font-size: 24rpx;
  line-height: 1.45;
}

.badge {
  flex: 0 0 auto;
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #e8f5ec;
  color: #145535;
  font-size: 22rpx;
  font-weight: 760;
}

.badge.warn {
  background: #fff4dd;
  color: #b7791f;
}

.data-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16rpx;
  margin-top: 24rpx;
}

.kv {
  min-width: 0;
  padding: 20rpx;
  border-radius: 8rpx;
  background: #f8faf6;
}

.kv text {
  display: block;
  color: #6d776c;
  font-size: 22rpx;
}

.kv .value {
  margin-top: 10rpx;
  color: #172018;
  font-size: 25rpx;
  font-weight: 700;
  overflow-wrap: anywhere;
}
</style>
