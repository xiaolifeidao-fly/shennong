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
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { redirectToLogin } from '@/utils/authGuard'

const userStore = useUserStore()

const displayNickname = computed(() => {
  const profile = userStore.profile
  return profile?.wxNickname || profile?.name || userStore.displayName || '未登录'
})
const avatarText = computed(() => displayNickname.value.slice(0, 1))
const avatarUrl = computed(() => userStore.profile?.wxAvatar || userStore.profile?.avatar || '')
const profileItems = computed(() => {
  const profile = userStore.profile
  if (!profile) {
    return [
      { label: '姓名', value: '-' },
      { label: '手机号', value: '-' },
      { label: '昵称', value: '-' },
    ]
  }

  return [
    { label: '姓名', value: profile.name ?? '-' },
    { label: '手机号', value: profile.phone ?? '-' },
    { label: '昵称', value: profile.wxNickname ?? '-' },
  ]
})

onShow(() => {
  if (userStore.isLoggedIn && !userStore.profile) {
    void loadProfile()
  }
})

function goLogin() {
  uni.navigateTo({ url: '/pages/login/index' })
}

function goProfile() {
  uni.navigateTo({ url: '/pages/profile/index' })
}

async function loadProfile() {
  try {
    await userStore.refreshProfile()
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '获取资料失败',
      icon: 'none',
    })
  }
}

async function handleLogout() {
  await userStore.logout()
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
