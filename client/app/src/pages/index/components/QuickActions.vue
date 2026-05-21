<template>
  <view class="quick-actions">
    <button class="quick" @click="$emit('new-entry')">
      <text class="icon">录</text>
      <text>新增录入</text>
    </button>
    <button class="quick" @click="$emit('scan-id')">
      <text class="icon">证</text>
      <text>扫身份证</text>
    </button>
    <button class="quick" @click="$emit('scan-bank')">
      <text class="icon">卡</text>
      <text>扫银行卡</text>
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
  border-radius: 8rpx;
  background: #ffffff;
  box-shadow: 0 8rpx 22rpx rgba(31, 47, 31, 0.06);
  color: #394439;
  font-size: 24rpx;
  font-weight: 700;
  line-height: 1.2;
}

.icon {
  display: flex;
  width: 54rpx;
  height: 54rpx;
  align-items: center;
  justify-content: center;
  border-radius: 8rpx;
  background: #e8f5ec;
  color: #145535;
  font-size: 26rpx;
  font-weight: 800;
}
</style>
