<template>
  <view>
    <view class="search-card">
      <text class="search-icon"></text>
      <input
        :value="keyword"
        confirm-type="search"
        placeholder="姓名前缀、身份证或后4位、手机号"
        @input="handleInput"
        @confirm="$emit('search')"
      />
      <button @click="$emit('search')">
        <text class="arrow-icon"></text>
      </button>
    </view>
    <scroll-view class="chips" scroll-x>
      <button
        v-for="item in filters"
        :key="item"
        class="chip"
        :class="{ active: item === activeFilter }"
        @click="$emit('filter-change', item)"
      >
        {{ item }}
      </button>
    </scroll-view>
    <view v-if="activeFilter === customFilter" class="custom-range">
      <picker mode="date" :value="customStart || today" :end="today" @change="handleDateChange('start', $event)">
        <view class="date-field">{{ customStart || '开始日期' }}</view>
      </picker>
      <text class="range-separator">至</text>
      <picker mode="date" :value="customEnd || customStart || today" :start="customStart || undefined" :end="today" @change="handleDateChange('end', $event)">
        <view class="date-field">{{ customEnd || '结束日期' }}</view>
      </picker>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  keyword: string
  activeFilter: string
  customStart: string
  customEnd: string
}>()

const emit = defineEmits<{
  'update:keyword': [value: string]
  'update:customStart': [value: string]
  'update:customEnd': [value: string]
  'filter-change': [value: string]
  search: []
}>()

const customFilter = '自定义'
const filters = ['今天', '本周', '本月', customFilter, '资料待补']
const today = computed(() => {
  const date = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
})

function handleInput(event: Event) {
  const detail = (event as unknown as { detail?: { value?: string } }).detail
  emit('update:keyword', detail?.value || '')
}

function handleDateChange(target: 'start' | 'end', event: { detail: { value: string } }) {
  const value = event.detail.value
  if (target === 'start') {
    emit('update:customStart', value)
    if (!props.customEnd || props.customEnd < value) {
      emit('update:customEnd', value)
    }
  } else {
    emit('update:customEnd', value)
  }
  emit('search')
}
</script>

<style lang="scss" scoped>
.search-card {
  display: grid;
  grid-template-columns: 36rpx 1fr auto;
  gap: 16rpx;
  align-items: center;
  padding: 20rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.72);
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 14rpx 34rpx rgba(31, 47, 31, 0.06);
}

.search-card input {
  min-width: 0;
  height: 76rpx;
  color: #172018;
  font-size: 28rpx;
}

.search-card button {
  display: flex;
  width: 76rpx;
  height: 76rpx;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 20rpx;
  background: linear-gradient(135deg, #237a4b, #145535);
  color: #ffffff;
  box-shadow: 0 10rpx 20rpx rgba(35, 122, 75, 0.18);
}

.search-icon {
  position: relative;
  width: 28rpx;
  height: 28rpx;
  border: 4rpx solid #237a4b;
  border-radius: 50%;
}

.search-icon::after {
  position: absolute;
  right: -8rpx;
  bottom: -5rpx;
  width: 14rpx;
  height: 4rpx;
  border-radius: 999rpx;
  background: #237a4b;
  content: '';
  transform: rotate(45deg);
}

.arrow-icon {
  position: relative;
  width: 28rpx;
  height: 4rpx;
  border-radius: 999rpx;
  background: #ffffff;
}

.arrow-icon::after {
  position: absolute;
  right: 0;
  top: -6rpx;
  width: 14rpx;
  height: 14rpx;
  border-top: 4rpx solid #ffffff;
  border-right: 4rpx solid #ffffff;
  content: '';
  transform: rotate(45deg);
}

.chips {
  width: 100%;
  white-space: nowrap;
  padding: 24rpx 0 4rpx;
}

.chip {
  display: inline-flex;
  margin-right: 16rpx;
  padding: 16rpx 24rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 999rpx;
  background: #ffffff;
  color: #485346;
  font-size: 26rpx;
  font-weight: 680;
  line-height: 1.2;
}

.chip.active {
  border-color: rgba(35, 122, 75, 0.35);
  background: linear-gradient(135deg, #e8f5ec, #fff8e8);
  color: #145535;
  box-shadow: 0 8rpx 18rpx rgba(35, 122, 75, 0.08);
}

.custom-range {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 44rpx minmax(0, 1fr);
  gap: 12rpx;
  align-items: center;
  padding: 12rpx 0 8rpx;
}

.date-field {
  height: 74rpx;
  padding: 0 18rpx;
  border: 1rpx solid #d8e5d6;
  border-radius: 18rpx;
  background: rgba(255, 255, 255, 0.96);
  color: #26362a;
  font-size: 25rpx;
  font-weight: 680;
  line-height: 74rpx;
  text-align: center;
}

.range-separator {
  color: #6a766a;
  font-size: 24rpx;
  text-align: center;
}
</style>
