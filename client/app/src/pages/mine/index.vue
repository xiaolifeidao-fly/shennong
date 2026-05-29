<template>
  <view class="page mine-page">
    <view class="profile">
      <view class="avatar">
        <text>{{ avatarText }}</text>
      </view>
      <view class="profile-main">
        <text class="name">{{ userStore.displayName }}</text>
        <text class="state">{{ userStore.isLoggedIn ? '已连接 app-api' : '登录后查看个人信息' }}</text>
      </view>
    </view>

    <view class="actions">
      <wd-button v-if="!userStore.isLoggedIn" type="primary" block @click="goLogin">去登录</wd-button>
      <wd-button v-else block plain :loading="loadingProfile" @click="loadProfile">刷新资料</wd-button>
      <wd-button v-if="userStore.isLoggedIn" block plain @click="handleLogout">退出登录</wd-button>
    </view>

    <SectionHeader title="我的预设" action-text="保存" @action="savePreset" />
    <view class="card form-card">
      <view class="field">
        <text class="label">业务员</text>
        <input v-model="presetForm.salesmanName" class="input" />
      </view>
      <view class="field">
        <text class="label">常用收购类型</text>
        <input v-model="presetForm.crops" class="input" />
      </view>
      <view class="field">
        <text class="label">付款登记默认方式</text>
        <input v-model="presetForm.payTypes" class="input" />
      </view>
      <view class="field last">
        <text class="label">常用收购地点</text>
        <input v-model="presetForm.places" class="input" />
      </view>
    </view>

    <view v-if="userStore.profile" class="section info-card">
      <view v-for="item in profileItems" :key="item.label" class="info-row">
        <text class="info-label">{{ item.label }}</text>
        <text class="info-value">{{ item.value }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SectionHeader from '@/components/business/SectionHeader.vue'
import { useGrainStore } from '@/stores/grain'
import { useUserStore } from '@/stores/user'
import { redirectToLogin } from '@/utils/authGuard'

const userStore = useUserStore()
const grainStore = useGrainStore()
const loadingProfile = ref(false)
const presetForm = reactive({
  salesmanName: grainStore.preset.salesmanName,
  crops: grainStore.preset.crops.join('、'),
  payTypes: grainStore.preset.payTypes.join('、'),
  places: grainStore.preset.places.join('、'),
})

const avatarText = computed(() => userStore.displayName.slice(0, 1))
const profileItems = computed(() => {
  const profile = userStore.profile
  if (!profile) {
    return []
  }

  return [
    { label: '用户 ID', value: profile.id ?? '-' },
    { label: '账号', value: profile.username ?? '-' },
    { label: '姓名', value: profile.name ?? '-' },
    { label: '手机号', value: profile.phone ?? '-' },
    { label: '微信 OpenID', value: profile.openUid ?? '-' },
    { label: '微信昵称', value: profile.wxNickname ?? '-' },
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

async function loadProfile() {
  loadingProfile.value = true
  try {
    await userStore.refreshProfile()
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '获取资料失败',
      icon: 'none',
    })
  } finally {
    loadingProfile.value = false
  }
}

async function handleLogout() {
  await userStore.logout()
  uni.showToast({ title: '已退出', icon: 'success' })
  redirectToLogin()
}

function savePreset() {
  grainStore.updatePreset({
    ...grainStore.preset,
    salesmanName: presetForm.salesmanName || '王强',
    crops: presetForm.crops.split(/[、,，]/).map((item) => item.trim()).filter(Boolean),
    payTypes: presetForm.payTypes.split(/[、,，]/).map((item) => item.trim()).filter(Boolean),
    places: presetForm.places.split(/[、,，]/).map((item) => item.trim()).filter(Boolean),
  })
  uni.showToast({ title: '已保存预设', icon: 'success' })
}
</script>

<style lang="scss" scoped>
.mine-page {
  display: flex;
  flex-direction: column;
}

.profile {
  display: flex;
  align-items: center;
  gap: 24rpx;
  padding: 36rpx 0 16rpx;
}

.avatar {
  display: flex;
  width: 112rpx;
  height: 112rpx;
  align-items: center;
  justify-content: center;
  border-radius: 56rpx;
  background: #1677ff;
  color: #ffffff;
  font-size: 42rpx;
  font-weight: 700;
}

.profile-main {
  min-width: 0;
  flex: 1;
}

.name {
  display: block;
  color: #111827;
  font-size: 40rpx;
  font-weight: 700;
}

.state {
  display: block;
  margin-top: 8rpx;
  color: #6b7280;
  font-size: 26rpx;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.form-card {
  padding: 28rpx;
}

.field {
  margin-bottom: 24rpx;
}

.field.last {
  margin-bottom: 0;
}

.label {
  display: block;
  margin-bottom: 14rpx;
  color: #445044;
  font-size: 26rpx;
  font-weight: 700;
}

.input {
  width: 100%;
  height: 88rpx;
  padding: 0 24rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 8rpx;
  background: #fbfcfa;
  color: #172018;
  font-size: 28rpx;
}

.info-card {
  padding: 28rpx;
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
