<template>
  <view class="period-tab">
    <view
      v-for="item in tabs"
      :key="item.value"
      class="period-tab__item"
      :class="{ 'period-tab__item--active': modelValue === item.value }"
      @tap="$emit('update:modelValue', item.value)"
    >
      <text>{{ item.label }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
export type PeriodValue = 'today' | 'week' | 'month' | 'custom'

defineProps<{ modelValue: PeriodValue }>()
defineEmits<{ (e: 'update:modelValue', value: PeriodValue): void }>()

const tabs: { label: string; value: PeriodValue }[] = [
  { label: '今日', value: 'today' },
  { label: '本周', value: 'week' },
  { label: '本月', value: 'month' },
  { label: '自定义', value: 'custom' },
]
</script>

<style lang="scss" scoped>
.period-tab {
  display: flex;
  gap: 12rpx;
  padding: 8rpx;
  border-radius: 20rpx;
  background: #f0f2f0;
}

.period-tab__item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 64rpx;
  border-radius: 14rpx;
  font-size: 26rpx;
  color: #6b7280;
  font-weight: 500;
  transition: all 0.2s;

  text {
    font-size: 26rpx;
  }
}

.period-tab__item--active {
  background: #ffffff;
  color: #145535;
  font-weight: 700;
  box-shadow: 0 2rpx 12rpx rgba(20, 85, 53, 0.12);

  text {
    font-size: 26rpx;
  }
}
</style>
