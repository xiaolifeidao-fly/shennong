<template>
  <view class="quick-actions">
    <button class="quick" @click="$emit('new-entry')">
      <text class="icon write-icon"></text>
      <text>新增录入</text>
    </button>
    <button class="quick" @click="$emit('scan-id')">
      <text class="icon card-icon"></text>
      <text>扫身份证</text>
    </button>
    <button class="quick" @click="$emit('scan-bank')">
      <text class="icon pay-icon"></text>
      <text>付款识别</text>
    </button>
  </view>
</template>

<script setup lang="ts">
defineEmits<{
  'new-entry': []
  'scan-id': []
  'scan-bank': []
}>()
</script>

<style lang="scss" scoped>
.quick-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20rpx;
  margin-top: 28rpx;
}

.quick {
  display: flex;
  min-height: 156rpx;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  padding: 20rpx 12rpx;
  border: 0;
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 16rpx 34rpx rgba(31, 47, 31, 0.08);
  color: #394439;
  font-size: 24rpx;
  font-weight: 700;
  line-height: 1.2;
  animation: action-pop 0.42s ease both;
}

.quick:nth-child(2) {
  animation-delay: 0.05s;
}

.quick:nth-child(3) {
  animation-delay: 0.1s;
}

.icon {
  position: relative;
  display: flex;
  width: 62rpx;
  height: 62rpx;
  align-items: center;
  justify-content: center;
  border-radius: 20rpx;
  background: linear-gradient(135deg, #e5f7dc, #dff3ff);
  color: #145535;
}

.write-icon::before {
  width: 28rpx;
  height: 8rpx;
  border-radius: 999rpx;
  background: #145535;
  content: '';
  transform: rotate(-36deg);
}

.write-icon::after {
  position: absolute;
  right: 16rpx;
  bottom: 16rpx;
  width: 16rpx;
  height: 4rpx;
  border-radius: 999rpx;
  background: #ffb84d;
  content: '';
}

.card-icon::before {
  width: 34rpx;
  height: 24rpx;
  border: 4rpx solid #145535;
  border-radius: 8rpx;
  content: '';
}

.card-icon::after {
  position: absolute;
  left: 20rpx;
  top: 31rpx;
  width: 22rpx;
  height: 4rpx;
  border-radius: 999rpx;
  background: #ffb84d;
  content: '';
}

.pay-icon::before {
  width: 32rpx;
  height: 24rpx;
  border: 4rpx solid #145535;
  border-top-width: 10rpx;
  border-radius: 8rpx;
  content: '';
}

.pay-icon::after {
  position: absolute;
  right: 14rpx;
  bottom: 14rpx;
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  background: #ffb84d;
  content: '';
}

@keyframes action-pop {
  from {
    opacity: 0;
    transform: translateY(18rpx);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
