<template>
  <view class="page login-page">
    <view class="header">
      <text class="title">手机号一键登录</text>
      <text class="subtitle">授权微信绑定手机号后自动登录，并由 app-api 返回后续请求使用的 token。</text>
    </view>

    <view class="wechat-card">
      <view class="wechat-icon">微</view>
      <view class="wechat-copy">
        <text class="wechat-title">使用微信手机号登录</text>
        <text class="wechat-desc">点击后会弹出微信手机号授权，允许后自动完成登录。</text>
      </view>
      <button class="wechat-login-btn" open-type="getPhoneNumber" :loading="wechatSubmitting" @getphonenumber="handleWechatPhoneLogin">
        一键登录
      </button>
      <button class="plain-login-btn" :disabled="wechatSubmitting" @click="handleWechatLogin">仅微信身份登录</button>
    </view>

    <view class="form">
      <text class="form-title">账号密码备用登录</text>
      <wd-input v-model="form.username" label="账号" clearable placeholder="请输入账号" />
      <wd-input
        v-model="form.password"
        label="密码"
        clearable
        show-password
        type="password"
        placeholder="请输入密码"
      />
      <wd-button type="primary" block :loading="submitting" @click="handleLogin">登录</wd-button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const submitting = ref(false)
const wechatSubmitting = ref(false)
const form = reactive({
  username: '',
  password: '',
})

interface GetPhoneNumberEvent {
  detail?: {
    code?: string
    errMsg?: string
  }
}

async function handleWechatPhoneLogin(event: GetPhoneNumberEvent) {
  const code = event.detail?.code
  if (!code) {
    const message = event.detail?.errMsg || '未授权手机号'
    console.warn('getPhoneNumber failed:', event.detail)
    uni.showModal({
      title: '获取手机号失败',
      content: message,
      showCancel: false,
    })
    return
  }

  wechatSubmitting.value = true
  try {
    await userStore.loginWithWechatPhone(code)
    uni.showToast({ title: '登录成功', icon: 'success' })
    const profile = userStore.profile
    if (!profile?.wxNickname || profile.wxNickname === '微信用户' || !profile?.wxAvatar) {
      uni.navigateTo({ url: '/pages/profile/index?from=login' })
      return
    }
    uni.switchTab({ url: '/pages/index/index' })
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '微信登录失败',
      icon: 'none',
    })
  } finally {
    wechatSubmitting.value = false
  }
}

async function handleWechatLogin() {
  wechatSubmitting.value = true
  try {
    await userStore.loginWithWechat()
    uni.showToast({ title: '登录成功', icon: 'success' })
    const profile = userStore.profile
    if (!profile?.phone || !profile?.wxNickname || profile.wxNickname === '微信用户') {
      uni.navigateTo({ url: '/pages/profile/index?from=login' })
      return
    }
    uni.switchTab({ url: '/pages/index/index' })
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '微信登录失败',
      icon: 'none',
    })
  } finally {
    wechatSubmitting.value = false
  }
}

async function handleLogin() {
  if (!form.username || !form.password) {
    uni.showToast({ title: '请输入账号和密码', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    await userStore.loginWithPassword({
      username: form.username,
      password: form.password,
    })
    uni.showToast({ title: '登录成功', icon: 'success' })
    uni.switchTab({ url: '/pages/mine/index' })
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '登录失败',
      icon: 'none',
    })
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
.login-page {
  display: flex;
  flex-direction: column;
  gap: 36rpx;
  background: linear-gradient(180deg, rgba(35, 122, 75, 0.12), rgba(245, 247, 243, 0) 320rpx), #f5f7f3;
}

.header {
  padding-top: 48rpx;
}

.title {
  display: block;
  color: #111827;
  font-size: 44rpx;
  font-weight: 700;
}

.subtitle {
  display: block;
  margin-top: 12rpx;
  color: #6d776c;
  font-size: 28rpx;
  line-height: 1.5;
}

.wechat-card {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
  padding: 32rpx;
  border: 1rpx solid rgba(226, 232, 221, 0.9);
  border-radius: 8rpx;
  background: #ffffff;
  box-shadow: 0 8rpx 22rpx rgba(31, 47, 31, 0.045);
}

.wechat-icon {
  display: flex;
  width: 96rpx;
  height: 96rpx;
  align-items: center;
  justify-content: center;
  border-radius: 8rpx;
  background: #e8f5ec;
  color: #145535;
  font-size: 38rpx;
  font-weight: 800;
}

.wechat-title,
.wechat-desc,
.form-title {
  display: block;
}

.wechat-title {
  color: #172018;
  font-size: 34rpx;
  font-weight: 760;
}

.wechat-desc {
  margin-top: 8rpx;
  color: #6d776c;
  font-size: 26rpx;
  line-height: 1.5;
}

.wechat-login-btn {
  width: 100%;
  min-height: 88rpx;
  border: 0;
  border-radius: 8rpx;
  background: #237a4b;
  color: #ffffff;
  font-size: 30rpx;
  font-weight: 800;
  line-height: 88rpx;
}

.plain-login-btn {
  width: 100%;
  min-height: 76rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 8rpx;
  background: #ffffff;
  color: #445044;
  font-size: 26rpx;
  font-weight: 700;
  line-height: 76rpx;
}

.form-title {
  color: #445044;
  font-size: 28rpx;
  font-weight: 700;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
  padding: 28rpx;
  border: 1rpx solid rgba(226, 232, 221, 0.9);
  border-radius: 8rpx;
  background: #ffffff;
}
</style>
