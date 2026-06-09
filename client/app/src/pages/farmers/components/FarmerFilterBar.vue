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
  </view>
</template>

<script setup lang="ts">
defineProps<{
  keyword: string
  activeFilter: string
}>()

const emit = defineEmits<{
  'update:keyword': [value: string]
  'filter-change': [value: string]
  search: []
}>()

const filters = ['今天', '本周', '本月', '资料待补']

function handleInput(event: Event) {
  const detail = (event as unknown as { detail?: { value?: string } }).detail
  emit('update:keyword', detail?.value || '')
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
</style>
