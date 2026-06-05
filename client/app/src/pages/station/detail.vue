<template>
  <view class="page station-detail-page">
    <view v-if="loading" class="empty-state">加载中...</view>
    <view v-else-if="errorMessage" class="empty-state">{{ errorMessage }}</view>
    <template v-else-if="station">
      <view class="section-card">
        <view class="card-header">
          <text class="station-title">{{ station.stationName || '所属粮站' }}</text>
          <text v-if="statusText" class="status-tag">{{ statusText }}</text>
        </view>
        <view class="info-grid">
          <view class="kv">
            <text class="kv-label">联系人</text>
            <text class="kv-value">{{ station.contactName || '-' }}</text>
          </view>
          <view class="kv">
            <text class="kv-label">联系电话</text>
            <text class="kv-value">{{ station.contactPhone || '-' }}</text>
          </view>
        </view>
        <view class="kv kv-full">
          <text class="kv-label">所在地区</text>
          <text class="kv-value">{{ regionText || '-' }}</text>
        </view>
        <view class="kv kv-full">
          <text class="kv-label">详细地址</text>
          <text class="kv-value">{{ station.address || '-' }}</text>
        </view>
        <view v-if="station.remark" class="kv kv-full">
          <text class="kv-label">备注</text>
          <text class="kv-value">{{ station.remark }}</text>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">营业执照</text>
        <view v-if="station.businessLicenseUrl" class="license-wrap">
          <image
            :src="businessLicenseImage"
            class="license-img"
            mode="widthFix"
            @click="previewLicense"
          />
        </view>
        <view v-else class="img-placeholder">
          <text>未上传营业执照</text>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getBusinessLicenseImagePath, getMyGrainStationDetail } from '@/services/grainConfig'
import type { GrainStationDetail } from '@/types/grain'

const station = ref<GrainStationDetail | null>(null)
const businessLicenseImage = ref('')
const loading = ref(false)
const errorMessage = ref('')

const statusText = computed(() => {
  switch (station.value?.status) {
    case 'active':
      return '正常'
    case 'inactive':
      return '停用'
    default:
      return ''
  }
})

const regionText = computed(() => {
  if (!station.value) return ''
  return [station.value.province, station.value.city, station.value.district]
    .filter((part) => typeof part === 'string' && part.trim())
    .join('')
})

onLoad(() => {
  void loadDetail()
})

async function loadDetail() {
  loading.value = true
  errorMessage.value = ''
  try {
    const detail = await getMyGrainStationDetail()
    station.value = detail
    businessLicenseImage.value = ''
    if (detail.businessLicenseUrl) {
      businessLicenseImage.value = await getBusinessLicenseImagePath(detail.businessLicenseUrl)
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : '获取粮站信息失败'
    errorMessage.value = message
    uni.showToast({ title: message, icon: 'none' })
  } finally {
    loading.value = false
  }
}

function previewLicense() {
  const url = businessLicenseImage.value
  if (!url) return
  uni.previewImage({ urls: [url], current: url })
}
</script>

<style lang="scss" scoped>
.station-detail-page {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
  padding: 24rpx 24rpx 48rpx;
}

.empty-state {
  padding: 80rpx 0;
  color: #6b7280;
  font-size: 28rpx;
  text-align: center;
}

.section-card {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
  padding: 28rpx;
  border-radius: 24rpx;
  background: #ffffff;
  box-shadow: 0 10rpx 24rpx rgba(31, 47, 31, 0.05);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.station-title {
  color: #172018;
  font-size: 34rpx;
  font-weight: 800;
}

.status-tag {
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
  background: #e7f4ec;
  color: #237a4b;
  font-size: 22rpx;
  font-weight: 600;
}

.section-title {
  color: #172018;
  font-size: 30rpx;
  font-weight: 800;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16rpx 24rpx;
}

.kv {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  min-width: 0;
}

.kv-full {
  grid-column: span 2;
}

.kv-label {
  color: #6b7280;
  font-size: 24rpx;
}

.kv-value {
  color: #172018;
  font-size: 28rpx;
  font-weight: 600;
  word-break: break-all;
}

.license-wrap {
  width: 100%;
  border-radius: 16rpx;
  overflow: hidden;
  background: #f5f7f3;
}

.license-img {
  width: 100%;
}

.img-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 240rpx;
  border-radius: 16rpx;
  background: #f5f7f3;
  color: #9ca3af;
  font-size: 26rpx;
}
</style>
