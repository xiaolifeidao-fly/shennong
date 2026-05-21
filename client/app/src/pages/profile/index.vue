<template>
  <view class="page profile-page">
    <view class="header">
      <text class="title">完善业务员资料</text>
      <text class="subtitle">微信登录只绑定身份，昵称、头像和手机号需要你主动授权或填写。</text>
    </view>

    <view class="card profile-card">
      <button class="avatar-button" open-type="chooseAvatar" @chooseavatar="handleChooseAvatar">
        <image v-if="form.wxAvatar" class="avatar-img" :src="form.wxAvatar" mode="aspectFill" />
        <text v-else class="avatar-placeholder">头像</text>
      </button>

      <view class="field">
        <text class="label">微信昵称</text>
        <input v-model="form.wxNickname" class="input" type="nickname" placeholder="点击选择或输入昵称" />
      </view>

      <view class="field">
        <text class="label">姓名</text>
        <input v-model="form.name" class="input" placeholder="业务员姓名" />
      </view>

      <view class="field">
        <text class="label">手机号</text>
        <view class="phone-row">
          <input v-model="form.phone" class="input" placeholder="授权后自动填入，也可手动输入" />
          <button class="phone-btn" open-type="getPhoneNumber" @getphonenumber="handleGetPhoneNumber">授权</button>
        </view>
      </view>

      <button class="submit" :loading="submitting" @click="saveProfile">保存并进入</button>
      <button class="skip" @click="goHome">暂时跳过</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'

interface ChooseAvatarEvent {
  detail?: {
    avatarUrl?: string
  }
}

interface GetPhoneNumberEvent {
  detail?: {
    code?: string
    errMsg?: string
  }
}

const userStore = useUserStore()
const submitting = ref(false)
const form = reactive({
  name: '',
  wxNickname: '',
  wxAvatar: '',
  phone: '',
})

onShow(() => {
  const profile = userStore.profile
  form.name = profile?.name || ''
  form.wxNickname = profile?.wxNickname || ''
  form.wxAvatar = profile?.wxAvatar || ''
  form.phone = profile?.phone || ''
})

function handleChooseAvatar(event: ChooseAvatarEvent) {
  const avatarUrl = event.detail?.avatarUrl
  if (avatarUrl) {
    form.wxAvatar = avatarUrl
  }
}

async function handleGetPhoneNumber(event: GetPhoneNumberEvent) {
  const code = event.detail?.code
  if (!code) {
    uni.showToast({ title: '未授权手机号', icon: 'none' })
    return
  }

  try {
    const profile = await userStore.bindWechatPhone(code)
    form.phone = profile?.phone || ''
    uni.showToast({ title: '手机号已绑定', icon: 'success' })
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '手机号绑定失败',
      icon: 'none',
    })
  }
}

async function saveProfile() {
  if (!form.wxNickname && !form.name) {
    uni.showToast({ title: '请填写昵称或姓名', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    const name = form.name || form.wxNickname
    await userStore.updateProfile({
      name,
      phone: form.phone,
      wxNickname: form.wxNickname || name,
      wxAvatar: form.wxAvatar,
    })
    uni.showToast({ title: '资料已保存', icon: 'success' })
    goHome()
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '保存失败',
      icon: 'none',
    })
  } finally {
    submitting.value = false
  }
}

function goHome() {
  uni.switchTab({ url: '/pages/index/index' })
}
</script>

<style lang="scss" scoped>
.profile-page {
  display: flex;
  flex-direction: column;
  gap: 32rpx;
}

.header {
  padding-top: 48rpx;
}

.title,
.subtitle {
  display: block;
}

.title {
  color: #172018;
  font-size: 44rpx;
  font-weight: 800;
}

.subtitle {
  margin-top: 12rpx;
  color: #6d776c;
  font-size: 28rpx;
  line-height: 1.5;
}

.profile-card {
  padding: 32rpx;
}

.avatar-button {
  display: flex;
  width: 132rpx;
  height: 132rpx;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: 0;
  border-radius: 8rpx;
  background: #e8f5ec;
  overflow: hidden;
}

.avatar-img {
  width: 132rpx;
  height: 132rpx;
}

.avatar-placeholder {
  color: #145535;
  font-size: 28rpx;
  font-weight: 800;
}

.field {
  margin-top: 28rpx;
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

.phone-row {
  display: grid;
  grid-template-columns: 1fr 144rpx;
  gap: 16rpx;
}

.phone-btn,
.submit,
.skip {
  min-height: 88rpx;
  border-radius: 8rpx;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 88rpx;
}

.phone-btn {
  border: 0;
  background: #eaf2fb;
  color: #2563a8;
}

.submit {
  width: 100%;
  margin-top: 36rpx;
  border: 1rpx solid #237a4b;
  background: #237a4b;
  color: #ffffff;
}

.skip {
  width: 100%;
  margin-top: 16rpx;
  border: 1rpx solid #e2e8dd;
  background: #ffffff;
  color: #445044;
}
</style>
