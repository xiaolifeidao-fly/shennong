<template>
  <view class="card entry-card">
    <view class="card-head">
      <view>
        <text class="card-title">{{ title }}</text>
        <text class="card-meta">{{ entry.place }}</text>
      </view>
      <text class="badge">{{ getEntryPrice(entry) }}</text>
    </view>
    <view class="data-grid">
      <view class="kv">
        <text>数量</text>
        <text class="value">{{ formatQuantity(entry.quantity, entry.unit) }}</text>
      </view>
      <view class="kv">
        <text>金额</text>
        <text class="value">{{ formatAmount(entry.amount) }}</text>
      </view>
      <view class="kv">
        <text>付款</text>
        <text class="value">{{ entry.payType }}</text>
      </view>
    </view>
    <view v-if="entry.materialImages.length" class="materials">
      <text class="sub-title">其他材料</text>
      <text class="sub-value">{{ entry.materialImages.length }} 张图片</text>
    </view>
    <view v-if="entry.editLogs.length" class="logs">
      <text class="sub-title">修改记录</text>
      <view v-for="log in entry.editLogs" :key="log.id" class="log-row">
        <text>{{ log.time }} · {{ log.operator }}</text>
        <text class="log-summary">{{ log.summary }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatAmount, formatQuantity, getEntryPrice, pickTime } from '@/utils/grain'
import type { FarmerProfile, GrainEntry } from '@/types/grain'

const props = defineProps<{
  entry: GrainEntry
  farmer?: FarmerProfile
}>()

const title = computed(() => `${pickTime(props.entry.buyTime)} · ${props.entry.crop}`)
</script>

<style lang="scss" scoped>
.entry-card {
  padding: 28rpx;
}

.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}

.card-title,
.card-meta {
  display: block;
}

.card-title {
  color: #172018;
  font-size: 30rpx;
  font-weight: 760;
}

.card-meta {
  margin-top: 10rpx;
  color: #6d776c;
  font-size: 24rpx;
}

.badge {
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #e8f5ec;
  color: #145535;
  font-size: 22rpx;
  font-weight: 760;
  white-space: nowrap;
}

.data-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16rpx;
  margin-top: 24rpx;
}

.kv {
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
}

.materials,
.logs {
  padding-top: 20rpx;
  margin-top: 20rpx;
  border-top: 1rpx solid #edf1e8;
}

.sub-title,
.sub-value,
.log-row text {
  display: block;
}

.sub-title {
  color: #445044;
  font-size: 24rpx;
  font-weight: 760;
}

.sub-value,
.log-row {
  margin-top: 8rpx;
  color: #6d776c;
  font-size: 23rpx;
  line-height: 1.45;
}

.log-summary {
  margin-top: 4rpx;
  color: #172018;
}
</style>
