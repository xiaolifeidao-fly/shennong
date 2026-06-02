<template>
  <view class="page login-page">
    <view class="header">
      <text class="brand-icon"></text>
      <text class="title">登录</text>
      <text class="subtitle">使用账号密码登录，或一键登录。</text>
    </view>

    <view class="form">
      <text class="form-title">账号密码登录</text>
      <wd-input v-model="form.username" label="账号" clearable placeholder="请输入账号" />
      <wd-input
        v-model="form.password"
        label="密码"
        clearable
        show-password
        type="password"
        placeholder="请输入密码"
      />
      <view class="login-actions">
        <wd-button type="primary" :loading="submitting" @click="handleLogin">登录</wd-button>
        <button class="quick-login-btn" open-type="getPhoneNumber" :loading="wechatSubmitting" @getphonenumber="handleWechatPhoneLogin">
          一键登录
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { redirectLoggedInUser } from '@/utils/authGuard'

const userStore = useUserStore()
const submitting = ref(false)
const wechatSubmitting = ref(false)
const form = reactive({
  username: '',
  password: '',
})

onShow(() => {
  redirectLoggedInUser()
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
  const username = form.username.trim()
  const password = form.password.trim()
  if (!username || !password) {
    uni.showToast({ title: '请输入账号和密码', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    await userStore.loginWithPassword({
      username,
      password,
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
  background:
    radial-gradient(circle at 18% 0%, rgba(255, 184, 77, 0.2), transparent 260rpx),
    linear-gradient(180deg, rgba(35, 122, 75, 0.12), rgba(245, 247, 243, 0) 320rpx),
    #eef4ed;
}

.header {
  padding-top: 48rpx;
}

.brand-icon {
  position: relative;
  display: block;
  width: 72rpx;
  height: 72rpx;
  margin-bottom: 18rpx;
  border-radius: 24rpx;
  background: linear-gradient(135deg, #237a4b, #ffb84d);
  box-shadow: 0 16rpx 30rpx rgba(35, 122, 75, 0.16);
}

.brand-icon::before {
  position: absolute;
  left: 27rpx;
  top: 16rpx;
  width: 18rpx;
  height: 40rpx;
  border-radius: 18rpx 18rpx 4rpx 4rpx;
  background: #ffffff;
  content: '';
  transform: rotate(34deg);
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

.form-title {
  display: block;
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
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 18rpx 44rpx rgba(31, 47, 31, 0.08);
}

.login-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
}

.quick-login-btn {
  width: 100%;
  min-height: 88rpx;
  padding: 0;
  border: 1rpx solid #237a4b;
  border-radius: 18rpx;
  background: #ffffff;
  color: #237a4b;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 88rpx;
}
</style>
