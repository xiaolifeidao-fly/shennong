<template>
  <view>
    <view class="search-card">
      <input :value="keyword" placeholder="搜索我的农户、身份证号、手机号" @input="handleInput" />
      <button @click="$emit('search')">搜索</button>
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
  grid-template-columns: 1fr auto;
  gap: 16rpx;
  align-items: center;
  padding: 20rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 8rpx;
  background: #ffffff;
}

.search-card input {
  min-width: 0;
  height: 76rpx;
  color: #172018;
  font-size: 28rpx;
}

.search-card button {
  width: 128rpx;
  height: 76rpx;
  border: 0;
  border-radius: 8rpx;
  background: #237a4b;
  color: #ffffff;
  font-size: 26rpx;
  font-weight: 760;
  line-height: 76rpx;
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
  background: #e8f5ec;
  color: #145535;
}
</style>
