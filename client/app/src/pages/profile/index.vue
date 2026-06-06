<template>
  <view class="page profile-page">
    <view class="header">
      <text class="title">个人信息</text>
      <text class="subtitle">可修改昵称、头像、姓名和手机号。</text>
    </view>

    <view class="card profile-card">
      <button class="avatar-button" open-type="chooseAvatar" @chooseavatar="handleChooseAvatar">
        <image v-if="form.wxAvatar" class="avatar-img" :src="form.wxAvatar" mode="aspectFill" />
        <text v-else class="avatar-placeholder">
          <text class="camera-icon"></text>
        </text>
        <text class="avatar-badge">
          <text class="camera-icon small"></text>
        </text>
      </button>

      <view class="field">
        <text class="label">昵称</text>
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
          <button class="phone-btn icon-text-btn" open-type="getPhoneNumber" @getphonenumber="handleGetPhoneNumber">
            <text class="phone-icon"></text>
            <text>授权</text>
          </button>
        </view>
      </view>

      <button class="submit icon-text-btn" :loading="submitting" @click="saveProfile">
        <text class="check-icon"></text>
        <text>保存</text>
      </button>
      <button class="skip icon-text-btn" @click="goMine">
        <text class="back-icon"></text>
        <text>返回我的</text>
      </button>
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
    goMine()
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '保存失败',
      icon: 'none',
    })
  } finally {
    submitting.value = false
  }
}

function goMine() {
  uni.switchTab({ url: '/pages/mine/index' })
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
  position: relative;
  display: flex;
  width: 132rpx;
  height: 132rpx;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: 0;
  border-radius: 30rpx;
  background: linear-gradient(135deg, #e8f5ec, #fff5df);
  overflow: hidden;
}

.avatar-button::after,
.phone-btn::after,
.submit::after,
.skip::after {
  border: 0;
}

.avatar-img {
  width: 132rpx;
  height: 132rpx;
}

.avatar-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #145535;
  font-size: 28rpx;
  font-weight: 800;
}

.avatar-badge {
  position: absolute;
  right: 10rpx;
  bottom: 10rpx;
  display: flex;
  width: 42rpx;
  height: 42rpx;
  align-items: center;
  justify-content: center;
  border: 3rpx solid #ffffff;
  border-radius: 50%;
  background: #237a4b;
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
  border-radius: 18rpx;
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
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  min-height: 88rpx;
  border-radius: 18rpx;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 88rpx;
}

.icon-text-btn {
  padding: 0 24rpx;
}

.phone-btn {
  border: 1rpx solid #d6e8fb;
  background: #f3f8ff;
  color: #2563a8;
}

.submit {
  width: 100%;
  margin-top: 36rpx;
  border: 1rpx solid #237a4b;
  background: linear-gradient(135deg, #237a4b, #145535);
  color: #ffffff;
  box-shadow: 0 16rpx 30rpx rgba(35, 122, 75, 0.18);
}

.skip {
  width: 100%;
  margin-top: 16rpx;
  border: 1rpx solid #e2e8dd;
  background: #ffffff;
  color: #445044;
}

.camera-icon,
.phone-icon,
.check-icon,
.back-icon {
  position: relative;
  display: inline-block;
  flex: 0 0 auto;
}

.camera-icon {
  width: 44rpx;
  height: 32rpx;
  border: 4rpx solid #145535;
  border-radius: 8rpx;
}

.camera-icon::before {
  position: absolute;
  left: 8rpx;
  top: -11rpx;
  width: 18rpx;
  height: 8rpx;
  border-radius: 6rpx 6rpx 0 0;
  background: #145535;
  content: '';
}

.camera-icon::after {
  position: absolute;
  left: 13rpx;
  top: 7rpx;
  width: 10rpx;
  height: 10rpx;
  border: 4rpx solid #145535;
  border-radius: 50%;
  content: '';
}

.camera-icon.small {
  width: 22rpx;
  height: 16rpx;
  border-width: 3rpx;
  border-color: #ffffff;
}

.camera-icon.small::before {
  left: 4rpx;
  top: -7rpx;
  width: 10rpx;
  height: 5rpx;
  background: #ffffff;
}

.camera-icon.small::after {
  left: 6rpx;
  top: 3rpx;
  width: 5rpx;
  height: 5rpx;
  border-width: 3rpx;
  border-color: #ffffff;
}

.phone-icon {
  width: 22rpx;
  height: 34rpx;
  border: 4rpx solid #2563a8;
  border-radius: 8rpx;
}

.phone-icon::after {
  position: absolute;
  left: 7rpx;
  bottom: 3rpx;
  width: 6rpx;
  height: 6rpx;
  border-radius: 50%;
  background: #2563a8;
  content: '';
}

.check-icon {
  width: 32rpx;
  height: 20rpx;
  border-left: 5rpx solid #ffffff;
  border-bottom: 5rpx solid #ffffff;
  transform: rotate(-45deg) translateY(-4rpx);
}

.back-icon {
  width: 28rpx;
  height: 4rpx;
  border-radius: 99rpx;
  background: #445044;
}

.back-icon::before,
.back-icon::after {
  position: absolute;
  left: 0;
  width: 16rpx;
  height: 4rpx;
  border-radius: 99rpx;
  background: #445044;
  content: '';
}

.back-icon::before {
  top: -5rpx;
  transform: rotate(-42deg);
}

.back-icon::after {
  top: 5rpx;
  transform: rotate(42deg);
}
</style>
