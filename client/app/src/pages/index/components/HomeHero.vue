<template>
  <view class="hero">
    <view class="hero-top">
      <text class="eyebrow">今日我的收粮</text>
      <text class="spark-icon"></text>
    </view>
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
  padding: 40rpx;
  border-radius: 32rpx;
  background:
    radial-gradient(circle at 88% 20%, rgba(255, 207, 108, 0.44), transparent 160rpx),
    linear-gradient(135deg, rgba(19, 74, 52, 0.98), rgba(34, 132, 83, 0.94) 58%, rgba(110, 147, 65, 0.96)),
    #237a4b;
  box-shadow: 0 24rpx 54rpx rgba(31, 72, 45, 0.2);
  animation: hero-in 0.48s ease both;
}

.hero::after {
  position: absolute;
  right: -72rpx;
  bottom: -96rpx;
  width: 310rpx;
  height: 310rpx;
  border: 42rpx solid rgba(255, 255, 255, 0.14);
  border-radius: 40%;
  content: '';
  transform: rotate(18deg);
}

.eyebrow,
.title,
.hero-top,
.hero-grid {
  position: relative;
  z-index: 1;
}

.hero-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.eyebrow {
  display: block;
  color: rgba(255, 255, 255, 0.84);
  font-size: 24rpx;
  font-weight: 700;
  letter-spacing: 0;
}

.spark-icon {
  position: relative;
  width: 54rpx;
  height: 54rpx;
  border-radius: 18rpx;
  background: rgba(255, 255, 255, 0.16);
}

.spark-icon::before,
.spark-icon::after {
  position: absolute;
  background: #ffcf6c;
  content: '';
}

.spark-icon::before {
  left: 25rpx;
  top: 10rpx;
  width: 5rpx;
  height: 34rpx;
  border-radius: 999rpx;
}

.spark-icon::after {
  left: 10rpx;
  top: 25rpx;
  width: 34rpx;
  height: 5rpx;
  border-radius: 999rpx;
}

.title {
  display: block;
  max-width: 620rpx;
  margin-top: 16rpx;
  color: #ffffff;
  font-size: 52rpx;
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
  padding: 26rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.24);
  border-radius: 22rpx;
  background: rgba(255, 255, 255, 0.16);
  backdrop-filter: blur(14rpx);
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

@keyframes hero-in {
  from {
    opacity: 0;
    transform: translateY(22rpx) scale(0.98);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>
