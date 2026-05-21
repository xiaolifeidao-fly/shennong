<template>
  <view class="card table-card">
    <view class="row head">
      <text>农户</text>
      <text>品类</text>
      <text>数量</text>
      <text>金额</text>
    </view>
    <view v-for="entry in entries" :key="entry.id" class="row">
      <text>{{ farmerName(entry.farmerId) }}</text>
      <text>{{ entry.crop }}</text>
      <text>{{ entry.quantity.toLocaleString('zh-CN') }}{{ entry.unit }}</text>
      <text class="amount">{{ formatAmount(entry.amount) }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { formatAmount } from '@/utils/grain'
import type { FarmerProfile, GrainEntry } from '@/types/grain'

const props = defineProps<{
  entries: GrainEntry[]
  farmers: FarmerProfile[]
}>()

function farmerName(farmerId: string) {
  return props.farmers.find((farmer) => farmer.id === farmerId)?.name || '-'
}
</script>

<style lang="scss" scoped>
.table-card {
  overflow: hidden;
}

.row {
  display: grid;
  grid-template-columns: 1fr 0.8fr 0.9fr 1fr;
  align-items: center;
  min-height: 88rpx;
  border-bottom: 1rpx solid #e2e8dd;
}

.row:last-child {
  border-bottom: 0;
}

.row text {
  padding: 0 16rpx;
  color: #172018;
  font-size: 25rpx;
}

.head {
  background: #fbfcfa;
}

.head text {
  color: #6d776c;
  font-size: 24rpx;
  font-weight: 700;
}

.amount {
  color: #145535 !important;
  font-weight: 800;
}
</style>
