<template>
  <view class="page mine-page">
    <view class="profile-row">
      <view class="avatar">
        <image v-if="avatarUrl" class="avatar-img" :src="avatarUrl" mode="aspectFill" />
        <text v-else>{{ avatarText }}</text>
      </view>
      <view class="profile-main">
        <text class="nickname">{{ displayNickname }}</text>
      </view>
      <button v-if="userStore.isLoggedIn" class="icon-btn logout-btn" aria-label="退出登录" @click="handleLogout">
        <text class="logout-icon">
          <text class="logout-door"></text>
          <text class="logout-arrow"></text>
        </text>
      </button>
      <button v-else class="login-btn" @click="goLogin">登录</button>
    </view>

    <view class="card info-card">
      <view class="info-head">
        <text class="info-title">个人信息</text>
        <button v-if="userStore.isLoggedIn" class="icon-btn edit-btn" aria-label="编辑资料" @click="goProfile">
          <text class="edit-icon">
            <text class="edit-line"></text>
          </text>
        </button>
      </view>
      <view v-for="item in profileItems" :key="item.label" class="info-row">
        <text class="info-label">{{ item.label }}</text>
        <text class="info-value">{{ item.value }}</text>
      </view>
    </view>

    <view v-if="userStore.isLoggedIn" class="card station-card">
      <view class="info-head">
        <text class="info-title">所属粮站</text>
      </view>
      <button class="station-name-btn" :class="{ disabled: !hasStation }" @click="goStationDetail">
        <text class="station-icon">
          <text class="station-roof"></text>
          <text class="station-body"></text>
        </text>
        <view class="station-text">
          <text class="station-label">粮站名称</text>
          <text class="station-name">{{ stationName || '暂未绑定粮站' }}</text>
        </view>
        <text v-if="hasStation" class="chevron-icon"></text>
      </button>
      <view v-if="stationInfoLoading" class="station-loading">加载粮站信息中...</view>
      <view v-else-if="stationInfoMessage" class="station-loading">{{ stationInfoMessage }}</view>
      <view v-else class="station-info">
        <view v-for="item in stationItems" :key="item.label" class="info-row">
          <text class="info-label">{{ item.label }}</text>
          <text class="info-value">{{ item.value }}</text>
        </view>
      </view>
    </view>

    <button v-if="userStore.isLoggedIn" class="password-btn" @click="goChangePassword">
      <text class="lock-icon">
        <text class="lock-shackle"></text>
        <text class="lock-body"></text>
      </text>
      <text>修改密码</text>
      <text class="chevron-icon"></text>
    </button>

    <button class="password-btn agreement-btn" @click="goAgreement">
      <text class="doc-icon">
        <text class="doc-fold"></text>
      </text>
      <text>用户协议与隐私政策</text>
      <text class="chevron-icon"></text>
    </button>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getCurrentSalesmanGrainStation } from '@/services/grainConfig'
import { useUserStore } from '@/stores/user'
import type { GrainStationDetail } from '@/types/grain'
import { redirectToLogin } from '@/utils/authGuard'

const userStore = useUserStore()
const stationInfo = ref<GrainStationDetail | null>(null)
const stationInfoLoading = ref(false)
const stationInfoMessage = ref('')

const displayNickname = computed(() => {
  const profile = userStore.profile
  return profile?.wxNickname || profile?.name || userStore.displayName || '未登录'
})
const avatarText = computed(() => displayNickname.value.slice(0, 1))
const avatarUrl = computed(() => userStore.profile?.wxAvatar || userStore.profile?.avatar || '')
const stationName = computed(() => {
  const value = stationInfo.value?.stationName
  return typeof value === 'string' ? value.trim() : ''
})
const hasStation = computed(() => Boolean(stationInfo.value?.id && stationName.value))
const stationRegionText = computed(() => {
  if (!stationInfo.value) return ''
  return [stationInfo.value.province, stationInfo.value.city, stationInfo.value.district]
    .filter((part) => typeof part === 'string' && part.trim())
    .join('')
})
const stationItems = computed(() => {
  const station = stationInfo.value
  if (!station) {
    return [
      { label: '粮站编码', value: '-' },
      { label: '联系人', value: '-' },
      { label: '联系电话', value: '-' },
      { label: '粮站地址', value: '-' },
    ]
  }

  return [
    { label: '粮站编码', value: station.stationCode || '-' },
    { label: '联系人', value: station.contactName || '-' },
    { label: '联系电话', value: station.contactPhone || '-' },
    { label: '粮站地址', value: [stationRegionText.value, station.address].filter(Boolean).join('') || '-' },
  ]
})
const profileItems = computed(() => {
  const profile = userStore.profile
  if (!profile) {
    return [
      { label: '手机号', value: '-' },
      { label: '昵称', value: '-' },
    ]
  }

  return [
    { label: '手机号', value: profile.phone ?? '-' },
    { label: '昵称', value: profile.wxNickname ?? '-' },
  ]
})

onShow(() => {
  if (userStore.isLoggedIn && !userStore.profile) {
    void loadProfile()
  }
  if (userStore.isLoggedIn) {
    void loadStationInfo()
  } else {
    stationInfo.value = null
  }
})

function goLogin() {
  uni.navigateTo({ url: '/pages/login/index' })
}

function goProfile() {
  uni.navigateTo({ url: '/pages/profile/index' })
}

function goChangePassword() {
  uni.navigateTo({ url: '/pages/change-password/index' })
}

function goAgreement() {
  uni.navigateTo({
    url: '/pages/agreement/index',
    fail: (err) => {
      console.error('navigate agreement failed:', err)
      uni.showToast({ title: err?.errMsg || '页面打开失败', icon: 'none' })
    },
  })
}

function goStationDetail() {
  if (!hasStation.value) {
    uni.showToast({ title: '当前账号暂未绑定粮站', icon: 'none' })
    return
  }
  uni.navigateTo({ url: '/pages/station/detail' })
}

async function loadProfile() {
  try {
    await userStore.refreshProfile()
    await loadStationInfo()
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '获取资料失败',
      icon: 'none',
    })
  }
}

async function loadStationInfo() {
  if (stationInfoLoading.value || !userStore.isLoggedIn) return
  stationInfoLoading.value = true
  stationInfoMessage.value = ''
  try {
    stationInfo.value = await getCurrentSalesmanGrainStation()
    if (!stationInfo.value?.id || !stationInfo.value?.stationName) {
      stationInfoMessage.value = '当前账号暂未绑定粮站'
    }
  } catch (error) {
    stationInfo.value = null
    const message = error instanceof Error ? error.message : ''
    stationInfoMessage.value = message.includes('粮站ID不能为空') ? '当前账号暂未绑定粮站' : message || '未获取到粮站信息'
  } finally {
    stationInfoLoading.value = false
  }
}

async function handleLogout() {
  await userStore.logout()
  stationInfo.value = null
  stationInfoMessage.value = ''
  uni.showToast({ title: '已退出', icon: 'success' })
  redirectToLogin()
}
</script>

<style lang="scss" scoped>
.mine-page {
  display: flex;
  flex-direction: column;
  gap: 28rpx;
}

.profile-row {
  display: flex;
  align-items: center;
  gap: 24rpx;
  padding: 40rpx 0 10rpx;
}

.avatar {
  display: flex;
  width: 112rpx;
  height: 112rpx;
  align-items: center;
  justify-content: center;
  border-radius: 34rpx;
  background: linear-gradient(135deg, #237a4b, #ffb84d);
  color: #ffffff;
  font-size: 42rpx;
  font-weight: 700;
  overflow: hidden;
}

.avatar-img {
  width: 112rpx;
  height: 112rpx;
}

.profile-main {
  min-width: 0;
  flex: 1;
}

.nickname {
  display: block;
  color: #172018;
  font-size: 40rpx;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.icon-btn,
.login-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 72rpx;
  padding: 0;
  border: 1rpx solid #e2e8dd;
  border-radius: 20rpx;
  background: #ffffff;
  color: #445044;
  font-size: 26rpx;
  font-weight: 800;
  line-height: 72rpx;
}

.icon-btn {
  width: 72rpx;
  position: relative;
  box-shadow: 0 8rpx 18rpx rgba(31, 47, 31, 0.06);
}

.icon-btn::after {
  border: 0;
}

.login-btn {
  width: 120rpx;
}

.logout-btn {
  border-color: #f1d8d2;
  background: linear-gradient(135deg, #fff8f6, #ffffff);
}

.logout-icon,
.edit-icon {
  position: relative;
  display: block;
  width: 34rpx;
  height: 34rpx;
}

.logout-door {
  position: absolute;
  left: 2rpx;
  top: 5rpx;
  width: 15rpx;
  height: 24rpx;
  border: 4rpx solid #b75b48;
  border-right: 0;
  border-radius: 4rpx 0 0 4rpx;
}

.logout-arrow {
  position: absolute;
  left: 12rpx;
  top: 15rpx;
  width: 20rpx;
  height: 4rpx;
  border-radius: 99rpx;
  background: #b75b48;
}

.logout-arrow::before,
.logout-arrow::after {
  position: absolute;
  right: 0;
  width: 13rpx;
  height: 4rpx;
  border-radius: 99rpx;
  background: #b75b48;
  content: '';
}

.logout-arrow::before {
  top: -4rpx;
  transform: rotate(38deg);
}

.logout-arrow::after {
  top: 4rpx;
  transform: rotate(-38deg);
}

.info-card {
  padding: 30rpx;
}

.password-btn,
.station-name-btn {
  display: grid;
  grid-template-columns: 40rpx 1fr 28rpx;
  gap: 18rpx;
  align-items: center;
  width: 100%;
  min-height: 96rpx;
  padding: 0 28rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 20rpx;
  background: #ffffff;
  color: #172018;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 96rpx;
  text-align: left;
  box-shadow: 0 10rpx 24rpx rgba(31, 47, 31, 0.05);
}

.password-btn::after,
.station-name-btn::after {
  border: 0;
}

.station-card {
  padding: 30rpx;
}

.station-name-btn {
  min-height: 108rpx;
  margin-top: 18rpx;
  padding: 16rpx 0 18rpx;
  border: 0;
  border-bottom: 1rpx solid #eef2eb;
  border-radius: 0;
  box-shadow: none;
  line-height: 1.3;
}

.station-name-btn.disabled {
  opacity: 0.72;
}

.station-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: 6rpx;
}

.station-label {
  color: #6b7280;
  font-size: 24rpx;
  font-weight: 500;
}

.station-name {
  color: #172018;
  font-size: 30rpx;
  font-weight: 800;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.station-loading {
  padding: 22rpx 0 6rpx;
  color: #8a9588;
  font-size: 26rpx;
}

.station-info {
  padding-top: 6rpx;
}

.station-icon {
  position: relative;
  display: block;
  width: 40rpx;
  height: 40rpx;
}

.station-roof {
  position: absolute;
  left: 0;
  top: 6rpx;
  width: 0;
  height: 0;
  border-left: 20rpx solid transparent;
  border-right: 20rpx solid transparent;
  border-bottom: 14rpx solid #237a4b;
}

.station-body {
  position: absolute;
  left: 5rpx;
  bottom: 2rpx;
  width: 30rpx;
  height: 22rpx;
  border-radius: 4rpx;
  background: #237a4b;
}

.station-body::after {
  position: absolute;
  left: 11rpx;
  top: 6rpx;
  width: 8rpx;
  height: 14rpx;
  border-radius: 2rpx;
  background: #ffffff;
  content: '';
}

.agreement-btn {
  margin-top: 20rpx;
}

.doc-icon {
  position: relative;
  display: block;
  width: 40rpx;
  height: 40rpx;
}

.doc-icon::before {
  position: absolute;
  left: 6rpx;
  top: 3rpx;
  width: 24rpx;
  height: 32rpx;
  border: 4rpx solid #237a4b;
  border-radius: 5rpx;
  content: '';
}

.doc-icon::after {
  position: absolute;
  left: 13rpx;
  top: 13rpx;
  width: 14rpx;
  height: 4rpx;
  border-radius: 99rpx;
  background: #237a4b;
  box-shadow: 0 7rpx 0 #237a4b;
  content: '';
}

.doc-fold {
  position: absolute;
  right: 6rpx;
  top: 3rpx;
  width: 0;
  height: 0;
  border-top: 10rpx solid #237a4b;
  border-left: 10rpx solid transparent;
}

.lock-icon,
.chevron-icon {
  position: relative;
  display: block;
}

.lock-icon {
  width: 40rpx;
  height: 40rpx;
}

.lock-shackle {
  position: absolute;
  left: 9rpx;
  top: 2rpx;
  width: 22rpx;
  height: 20rpx;
  border: 4rpx solid #237a4b;
  border-bottom: 0;
  border-radius: 14rpx 14rpx 0 0;
}

.lock-body {
  position: absolute;
  left: 5rpx;
  bottom: 3rpx;
  width: 30rpx;
  height: 24rpx;
  border-radius: 8rpx;
  background: #237a4b;
}

.lock-body::after {
  position: absolute;
  left: 13rpx;
  top: 8rpx;
  width: 6rpx;
  height: 10rpx;
  border-radius: 99rpx;
  background: #ffffff;
  content: '';
}

.chevron-icon {
  width: 18rpx;
  height: 18rpx;
  border-top: 4rpx solid #8a9488;
  border-right: 4rpx solid #8a9488;
  transform: rotate(45deg);
}

.info-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8rpx;
}

.info-title {
  color: #172018;
  font-size: 30rpx;
  font-weight: 800;
}

.edit-btn {
  border-color: #d6eadc;
  background: linear-gradient(135deg, #f6fbf7, #ffffff);
}

.edit-icon {
  transform: rotate(-45deg);
}

.edit-icon::before {
  position: absolute;
  left: 14rpx;
  top: 1rpx;
  width: 9rpx;
  height: 25rpx;
  border-radius: 5rpx;
  background: #237a4b;
  content: '';
}

.edit-icon::after {
  position: absolute;
  left: 15rpx;
  top: 27rpx;
  width: 0;
  height: 0;
  border-right: 4rpx solid transparent;
  border-left: 4rpx solid transparent;
  border-top: 8rpx solid #237a4b;
  content: '';
}

.edit-line {
  position: absolute;
  left: 12rpx;
  top: 7rpx;
  width: 13rpx;
  height: 4rpx;
  border-radius: 99rpx;
  background: #d6eadc;
  z-index: 1;
}

.info-row {
  display: flex;
  justify-content: space-between;
  gap: 24rpx;
  padding: 16rpx 0;
}

.info-label {
  color: #6b7280;
  font-size: 26rpx;
}

.info-value {
  color: #111827;
  font-size: 26rpx;
}
</style>
