<template>
  <view class="section-header">
    <text class="section-title">{{ title }}</text>
    <button v-if="actionText" class="link-button" @click="$emit('action')">{{ actionText }}</button>
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
  font-size: 34rpx;
  font-weight: 700;
  line-height: 1.3;
}

.link-button {
  padding: 8rpx;
  border: 0;
  background: transparent;
  color: #237a4b;
  font-size: 26rpx;
  font-weight: 700;
  line-height: 1.3;
}
</style>
