<template>
  <view class="metric-grid">
    <view
      v-for="item in metrics"
      :key="item.label"
      class="metric-card"
      :class="`metric-card--${item.tone}`"
    >
      <view class="metric-card__icon">
        <text class="metric-card__emoji">{{ item.emoji }}</text>
      </view>
      <text class="metric-card__value">{{ item.value }}</text>
      <text class="metric-card__label">{{ item.label }}</text>
      <text v-if="item.hint" class="metric-card__hint">{{ item.hint }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatAmount, formatQuantitySummary } from '@/utils/grain'
import type { GrainDashboardOverview } from '@/types/grain'

const props = defineProps<{ overview: GrainDashboardOverview | null }>()

const metrics = computed(() => {
  const o = props.overview
  return [
    {
      label: '收粮总量',
      value: o ? formatQuantitySummary(o.totalQuantity) : '--',
      hint: o ? `均价 ${formatAmount(o.averageUnitPrice)}/公斤` : '',
      emoji: '🌾',
      tone: 'green',
    },
    {
      label: '收粮金额',
      value: o ? formatAmount(o.totalAmount) : '--',
      hint: o ? `户均 ${formatAmount(o.averageFarmerDeal)}` : '',
      emoji: '💰',
      tone: 'gold',
    },
    {
      label: '收粮农户',
      value: o ? `${o.newFarmerCount} 户` : '--',
      hint: o ? `${o.entryCount} 笔流水` : '',
      emoji: '👨‍🌾',
      tone: 'blue',
    },
    {
      label: '参与业务员',
      value: o ? `${o.activeUserCount} 人` : '--',
      hint: '',
      emoji: '👤',
      tone: 'purple',
    },
  ]
})
</script>

<style lang="scss" scoped>
.metric-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
}

.metric-card {
  padding: 32rpx 28rpx 28rpx;
  border-radius: 24rpx;
  background: #ffffff;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.metric-card__icon {
  width: 60rpx;
  height: 60rpx;
  border-radius: 18rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8rpx;
}

.metric-card--green .metric-card__icon { background: #e8f5ee; }
.metric-card--gold .metric-card__icon { background: #fef7e6; }
.metric-card--blue .metric-card__icon { background: #e8f0fe; }
.metric-card--purple .metric-card__icon { background: #f3eefe; }

.metric-card__emoji {
  font-size: 32rpx;
}

.metric-card__value {
  font-size: 40rpx;
  font-weight: 800;
  color: #1a2a1e;
  line-height: 1.1;
}

.metric-card__label {
  font-size: 24rpx;
  color: #6b7280;
  font-weight: 500;
  margin-top: 4rpx;
}

.metric-card__hint {
  font-size: 22rpx;
  color: #9ca3af;
  margin-top: 2rpx;
}
</style>
