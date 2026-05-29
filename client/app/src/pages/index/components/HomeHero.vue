<template>
  <view class="hero">
    <text class="eyebrow">今日我的收粮</text>
    <text class="title">已录入 {{ entryCount }} 笔，合计 {{ totalTon }} 吨</text>
    <view class="hero-grid">
      <view class="hero-stat">
        <text class="stat-value">{{ amountText }}</text>
        <text class="stat-label">今日收购金额</text>
      </view>
      <view class="hero-stat">
        <text class="stat-value">{{ farmerCount }} 户</text>
        <text class="stat-label">今日农户数</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatAmount } from '@/utils/grain'

const props = defineProps<{
  entryCount: number
  totalQuantity: number
  totalAmount: number
  farmerCount: number
}>()

const totalTon = computed(() => (props.totalQuantity / 1000).toFixed(1))
const amountText = computed(() => formatAmount(props.totalAmount))
</script>

<style lang="scss" scoped>
.hero {
  position: relative;
  overflow: hidden;
  padding: 36rpx;
  border-radius: 8rpx;
  background: linear-gradient(135deg, rgba(21, 89, 54, 0.97), rgba(43, 119, 73, 0.91)), #237a4b;
  box-shadow: 0 14rpx 34rpx rgba(31, 47, 31, 0.08);
}

.hero::after {
  position: absolute;
  right: -84rpx;
  bottom: -104rpx;
  width: 336rpx;
  height: 336rpx;
  border: 48rpx solid rgba(255, 255, 255, 0.12);
  border-radius: 50%;
  content: '';
}

.eyebrow,
.title,
.hero-grid {
  position: relative;
  z-index: 1;
}

.eyebrow {
  display: block;
  color: rgba(255, 255, 255, 0.82);
  font-size: 24rpx;
}

.title {
  display: block;
  max-width: 620rpx;
  margin-top: 16rpx;
  color: #ffffff;
  font-size: 50rpx;
  font-weight: 800;
  line-height: 1.16;
}

.hero-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
  margin-top: 36rpx;
}

.hero-stat {
  padding: 24rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.18);
  border-radius: 8rpx;
  background: rgba(255, 255, 255, 0.14);
}

.stat-value,
.stat-label {
  display: block;
}

.stat-value {
  color: #ffffff;
  font-size: 40rpx;
  font-weight: 800;
}

.stat-label {
  margin-top: 10rpx;
  color: rgba(255, 255, 255, 0.82);
  font-size: 24rpx;
}
</style>
