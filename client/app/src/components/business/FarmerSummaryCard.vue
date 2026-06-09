<template>
  <view class="card farmer-card" @click="$emit('select', farmer.id)">
    <view class="card-head">
      <view class="card-main">
        <text class="card-title">{{ farmer.name }} · {{ farmer.entryCount }} 笔录入</text>
        <text class="card-meta">{{ metaText }}</text>
      </view>
      <text class="badge" :class="{ warn: farmer.status === 'missing-bank' }">
        <text class="badge-dot"></text>
        <text>{{ badgeText }}</text>
      </text>
      <button class="delete-btn" @click.stop="$emit('delete', farmer.id)">删除</button>
    </view>

    <view class="data-grid">
      <view class="kv">
        <text>合计数量</text>
        <text class="value">{{ formatQuantitySummary(farmer.totalQuantity) }}</text>
      </view>
      <view class="kv">
        <text>合计金额</text>
        <text class="value">{{ formatAmount(farmer.totalAmount) }}</text>
      </view>
      <view class="kv">
        <text>均价</text>
        <text class="value">{{ farmer.avgPrice.toFixed(2) }} 元/公斤</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatAmount, formatQuantitySummary, pickTime } from '@/utils/grain'
import type { FarmerSummary } from '@/types/grain'
import { maskIdNumber, maskPhone } from '@/utils/privacy'

const props = defineProps<{
  farmer: FarmerSummary
  compact?: boolean
}>()

defineEmits<{
  select: [farmerId: string]
  delete: [farmerId: string]
}>()

const badgeText = computed(() => (props.farmer.status === 'missing-bank' ? '待补' : '完整'))
const metaText = computed(() => {
  if (props.compact) {
    return [props.farmer.statusText, maskPhone(props.farmer.phone), maskIdNumber(props.farmer.idNumber)].filter(Boolean).join(' · ')
  }
  return `身份证 ${maskIdNumber(props.farmer.idNumber) || '-'} · 手机 ${maskPhone(props.farmer.phone) || '-'} · 最近 ${pickTime(props.farmer.latestTime)}`
})
</script>

<style lang="scss" scoped>
.farmer-card {
  padding: 28rpx;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.farmer-card:active {
  transform: scale(0.985);
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
  display: flex;
  align-items: center;
  gap: 8rpx;
  flex: 0 0 auto;
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #e8f5ec;
  color: #145535;
  font-size: 22rpx;
  font-weight: 760;
}

.badge-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: currentColor;
}

.badge.warn {
  background: #fff4dd;
  color: #b7791f;
}

.delete-btn {
  flex: 0 0 auto;
  min-width: 88rpx;
  min-height: 52rpx;
  padding: 0 18rpx;
  border: 1rpx solid #f0c7bf;
  border-radius: 12rpx;
  background: #fff7f5;
  color: #c2412d;
  font-size: 22rpx;
  font-weight: 760;
  line-height: 52rpx;
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
  border-radius: 18rpx;
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
