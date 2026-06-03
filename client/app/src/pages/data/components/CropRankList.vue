<template>
  <view class="crop-rank">
    <view class="crop-rank__header">
      <text class="crop-rank__title">粮食类型分布</text>
      <text class="crop-rank__sub">按金额排序</text>
    </view>
    <view v-if="sorted.length === 0" class="crop-rank__empty">
      <text>暂无数据</text>
    </view>
    <view v-else class="crop-rank__list">
      <view v-for="(item, index) in sorted" :key="item.key || item.name" class="crop-rank__row">
        <view class="crop-rank__index" :class="`crop-rank__index--${index < 3 ? index + 1 : 'rest'}`">
          <text>{{ index + 1 }}</text>
        </view>
        <view class="crop-rank__info">
          <text class="crop-rank__name">{{ item.name || '未命名' }}</text>
          <text class="crop-rank__meta">{{ formatQuantitySummary(item.totalQuantity) }} · {{ item.entryCount }} 笔 · {{ item.farmerCount }} 户</text>
          <view class="crop-rank__bar-track">
            <view
              class="crop-rank__bar-fill"
              :style="{ width: barWidth(item.totalAmount) + '%' }"
            />
          </view>
        </view>
        <view class="crop-rank__amount">
          <text class="crop-rank__amount-value">{{ formatAmount(item.totalAmount) }}</text>
          <text class="crop-rank__share">{{ shareText(item.amountShare) }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatAmount, formatQuantitySummary } from '@/utils/grain'
import type { GrainDashboardDimension } from '@/types/grain'

const props = defineProps<{ rows: GrainDashboardDimension[] }>()

const sorted = computed(() =>
  [...props.rows].sort((a, b) => b.totalAmount - a.totalAmount || b.totalQuantity - a.totalQuantity).slice(0, 8)
)

const maxAmount = computed(() => Math.max(...sorted.value.map((r) => r.totalAmount), 0))

function barWidth(amount: number) {
  if (!maxAmount.value) return 8
  return Math.max(4, (amount / maxAmount.value) * 100)
}

function shareText(share: number) {
  return `${Math.min(100, Math.max(0, Number(((share || 0) * 100).toFixed(1))))}%`
}
</script>

<style lang="scss" scoped>
.crop-rank {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.05);
}

.crop-rank__header {
  display: flex;
  align-items: baseline;
  gap: 16rpx;
  margin-bottom: 28rpx;
}

.crop-rank__title {
  font-size: 30rpx;
  font-weight: 700;
  color: #1a2a1e;
}

.crop-rank__sub {
  font-size: 24rpx;
  color: #9ca3af;
}

.crop-rank__empty {
  padding: 40rpx 0;
  text-align: center;

  text {
    font-size: 28rpx;
    color: #9ca3af;
  }
}

.crop-rank__list {
  display: flex;
  flex-direction: column;
  gap: 32rpx;
}

.crop-rank__row {
  display: flex;
  align-items: flex-start;
  gap: 20rpx;
}

.crop-rank__index {
  width: 44rpx;
  height: 44rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 4rpx;

  text {
    font-size: 24rpx;
    font-weight: 700;
    color: #9ca3af;
  }
}

.crop-rank__index--1 {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  text { color: #fff; }
}

.crop-rank__index--2 {
  background: linear-gradient(135deg, #94a3b8, #64748b);
  text { color: #fff; }
}

.crop-rank__index--3 {
  background: linear-gradient(135deg, #cd7c3a, #a85e23);
  text { color: #fff; }
}

.crop-rank__index--rest {
  background: #f3f4f6;
}

.crop-rank__info {
  flex: 1;
  min-width: 0;
}

.crop-rank__name {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #1a2a1e;
  margin-bottom: 6rpx;
}

.crop-rank__meta {
  display: block;
  font-size: 22rpx;
  color: #9ca3af;
  margin-bottom: 12rpx;
}

.crop-rank__bar-track {
  height: 8rpx;
  border-radius: 99rpx;
  background: #f0f2f0;
  overflow: hidden;
}

.crop-rank__bar-fill {
  height: 100%;
  border-radius: 99rpx;
  background: linear-gradient(90deg, #145535, #22843d);
  transition: width 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.crop-rank__amount {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6rpx;
  flex-shrink: 0;
}

.crop-rank__amount-value {
  font-size: 26rpx;
  font-weight: 700;
  color: #1a2a1e;
}

.crop-rank__share {
  font-size: 22rpx;
  color: #145535;
  font-weight: 600;
  background: #e8f5ee;
  padding: 2rpx 12rpx;
  border-radius: 99rpx;
}
</style>
