<template>
  <view class="section-header">
    <text class="section-title">{{ title }}</text>
    <button v-if="actionText" class="link-button" @click="$emit('action')">
      <text>{{ actionText }}</text>
      <text class="arrow-icon"></text>
    </button>
  </view>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  actionText?: string
}>()

defineEmits<{
  action: []
}>()
</script>

<style lang="scss" scoped>
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  margin: 36rpx 4rpx 18rpx;
}

.section-title {
  color: #172018;
  font-size: 32rpx;
  font-weight: 700;
  line-height: 1.3;
}

.link-button {
  display: flex;
  align-items: center;
  gap: 10rpx;
  min-height: 58rpx;
  padding: 0 18rpx;
  border: 0;
  border-radius: 999rpx;
  background: rgba(35, 122, 75, 0.1);
  color: #237a4b;
  font-size: 24rpx;
  font-weight: 700;
  line-height: 58rpx;
}

.arrow-icon {
  position: relative;
  width: 20rpx;
  height: 4rpx;
  border-radius: 999rpx;
  background: #237a4b;
}

.arrow-icon::after {
  position: absolute;
  right: 0;
  top: -5rpx;
  width: 12rpx;
  height: 12rpx;
  border-top: 4rpx solid #237a4b;
  border-right: 4rpx solid #237a4b;
  content: '';
  transform: rotate(45deg);
}
</style>
