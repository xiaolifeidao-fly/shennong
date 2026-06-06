<template>
  <view class="page data-page">
    <!-- 顶部标题区 -->
    <view class="data-header">
      <view class="data-header__title-row">
        <text class="data-header__title">收粮数据</text>
        <view class="data-header__refresh" :class="{ 'data-header__refresh--loading': loading }" @tap="load">
          <text class="data-header__refresh-icon">↻</text>
        </view>
      </view>
      <text class="data-header__range">{{ rangeLabel }}</text>
    </view>

    <!-- 时间段选择 -->
    <view class="section-gap">
      <PeriodTab v-model="period" />
    </view>

    <!-- 自定义日期选择器 -->
    <view v-if="period === 'custom'" class="custom-date section-gap">
      <view class="custom-date__row">
        <view class="custom-date__field" @tap="pickDate('start')">
          <text class="custom-date__label">开始日期</text>
          <text class="custom-date__value">{{ customStart || '请选择' }}</text>
        </view>
        <text class="custom-date__sep">—</text>
        <view class="custom-date__field" @tap="pickDate('end')">
          <text class="custom-date__label">结束日期</text>
          <text class="custom-date__value">{{ customEnd || '请选择' }}</text>
        </view>
      </view>
    </view>

    <!-- 骨架屏 / 加载中 -->
    <view v-if="loading && !dashboard" class="skeleton section-gap">
      <view class="skeleton__grid">
        <view v-for="i in 4" :key="i" class="skeleton__card" />
      </view>
      <view class="skeleton__block" />
    </view>

    <!-- 数据内容 -->
    <template v-else>
      <!-- 核心指标 -->
      <view class="section-gap">
        <DataMetricGrid :overview="dashboard?.overview ?? null" />
      </view>

      <!-- 粮食类型分布 -->
      <view class="section-gap">
        <CropRankList :rows="dashboard?.byCrop ?? []" />
      </view>

      <!-- 数据更新时间 -->
      <view v-if="dashboard?.generated" class="data-footer">
        <text>数据更新于 {{ formatGenerated(dashboard.generated) }}</text>
      </view>
    </template>

    <!-- 错误提示 -->
    <view v-if="errorMsg" class="error-tip">
      <text>{{ errorMsg }}</text>
      <text class="error-tip__retry" @tap="load">重试</text>
    </view>

    <!-- 日期选择器（隐藏，通过 picker 触发） -->
    <picker
      v-if="showPicker"
      mode="date"
      :value="pickerValue"
      :start="pickerMin"
      :end="pickerMax"
      @change="onPickerChange"
      @cancel="showPicker = false"
    />
  </view>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PeriodTab from './components/PeriodTab.vue'
import DataMetricGrid from './components/DataMetricGrid.vue'
import CropRankList from './components/CropRankList.vue'
import { getGrainDashboard } from '@/services/grainDashboard'
import type { GrainDashboard } from '@/types/grain'
import type { PeriodValue } from './components/PeriodTab.vue'

const period = ref<PeriodValue>('today')
const customStart = ref('')
const customEnd = ref('')
const dashboard = ref<GrainDashboard | null>(null)
const loading = ref(false)
const errorMsg = ref('')

// 日期选择器控制
const showPicker = ref(false)
const pickerTarget = ref<'start' | 'end'>('start')
const pickerValue = ref('')
const pickerMin = ref('2020-01-01')
const pickerMax = ref(todayStr())

function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function weekStart() {
  const d = new Date()
  const day = d.getDay() || 7
  d.setDate(d.getDate() - day + 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function monthStart() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`
}

const dateRange = computed<{ startDate: string; endDate: string }>(() => {
  const today = todayStr()
  if (period.value === 'today') return { startDate: today, endDate: today }
  if (period.value === 'week') return { startDate: weekStart(), endDate: today }
  if (period.value === 'month') return { startDate: monthStart(), endDate: today }
  return { startDate: customStart.value || today, endDate: customEnd.value || today }
})

const rangeLabel = computed(() => {
  const { startDate, endDate } = dateRange.value
  if (startDate === endDate) return startDate
  return `${startDate} 至 ${endDate}`
})

function formatGenerated(str: string) {
  return str.replace('T', ' ').slice(0, 16)
}

async function load() {
  if (loading.value) return
  errorMsg.value = ''
  loading.value = true
  try {
    dashboard.value = await getGrainDashboard(dateRange.value)
  } catch (e) {
    errorMsg.value = e instanceof Error ? e.message : '加载失败，请重试'
  } finally {
    loading.value = false
  }
}

watch(period, (val) => {
  if (val !== 'custom') {
    void load()
  }
})

function pickDate(target: 'start' | 'end') {
  pickerTarget.value = target
  pickerValue.value = target === 'start' ? (customStart.value || todayStr()) : (customEnd.value || todayStr())
  showPicker.value = true
}

function onPickerChange(e: { detail: { value: string } }) {
  const val = e.detail.value
  showPicker.value = false
  if (pickerTarget.value === 'start') {
    customStart.value = val
    if (customEnd.value && customEnd.value < val) customEnd.value = val
  } else {
    customEnd.value = val
    if (customStart.value && customStart.value > val) customStart.value = val
  }
  if (customStart.value && customEnd.value) {
    void load()
  }
}

onShow(() => {
  void load()
})
</script>

<style lang="scss" scoped>
.data-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #f5f7f3;
}

.data-header {
  padding: 32rpx 32rpx 0;
}

.data-header__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.data-header__title {
  font-size: 48rpx;
  font-weight: 800;
  color: #1a2a1e;
}

.data-header__refresh {
  width: 64rpx;
  height: 64rpx;
  border-radius: 20rpx;
  background: #e8f5ee;
  display: flex;
  align-items: center;
  justify-content: center;

  &--loading {
    animation: spin 0.8s linear infinite;
  }
}

.data-header__refresh-icon {
  font-size: 32rpx;
  color: #145535;
  font-weight: 700;
}

.data-header__range {
  display: block;
  font-size: 24rpx;
  color: #9ca3af;
  margin-top: 8rpx;
}

.section-gap {
  padding: 20rpx 32rpx 0;
}

.custom-date {
  // already has section-gap padding
}

.custom-date__row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 24rpx 28rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.05);
}

.custom-date__field {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.custom-date__label {
  font-size: 22rpx;
  color: #9ca3af;
}

.custom-date__value {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a2a1e;
}

.custom-date__sep {
  font-size: 28rpx;
  color: #d1d5db;
  flex-shrink: 0;
}

/* 骨架屏 */
.skeleton {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.skeleton__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
}

.skeleton__card {
  height: 196rpx;
  border-radius: 24rpx;
  background: linear-gradient(90deg, #e8ede8 0%, #f0f5f0 50%, #e8ede8 100%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}

.skeleton__block {
  height: 320rpx;
  border-radius: 24rpx;
  background: linear-gradient(90deg, #e8ede8 0%, #f0f5f0 50%, #e8ede8 100%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}

.data-footer {
  padding: 24rpx 32rpx 48rpx;

  text {
    font-size: 22rpx;
    color: #c0c6bf;
  }
}

.error-tip {
  margin: 20rpx 32rpx;
  padding: 28rpx;
  border-radius: 20rpx;
  background: #fff5f5;
  border: 1rpx solid #fee2e2;
  display: flex;
  align-items: center;
  justify-content: space-between;

  text {
    font-size: 26rpx;
    color: #dc2626;
  }
}

.error-tip__retry {
  font-size: 26rpx;
  color: #145535 !important;
  font-weight: 600;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
