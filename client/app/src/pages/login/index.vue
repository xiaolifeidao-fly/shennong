<template>
  <view class="page login-page">
    <view class="login-shell">
      <view class="brand">
        <view class="brand-mark">
          <text>神</text>
        </view>
        <view class="brand-copy">
          <text class="brand-title">神农收粮助手</text>
          <text class="brand-subtitle">业务员移动工作台</text>
        </view>
      </view>

      <view class="form-card">
        <view class="form-header">
          <text class="form-title">账号登录</text>
          <text class="form-desc">请输入账号信息进入系统</text>
        </view>

        <view class="fields">
          <wd-input v-model="form.username" label="账号" clearable placeholder="用户名或手机号" />
          <wd-input
            v-model="form.password"
            label="密码"
            clearable
            show-password
            type="password"
            placeholder="请输入登录密码"
          />
        </view>

        <view class="login-actions">
          <wd-button custom-class="primary-login-btn" type="primary" block :loading="submitting" @click="handleLogin">
            登录
          </wd-button>
          <button
            class="quick-login-btn"
            open-type="getPhoneNumber"
            :loading="wechatSubmitting"
            @getphonenumber="handleWechatPhoneLogin"
          >
            手机号快捷登录
          </button>
        </view>
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
    uni.showToast({ title: '请输入用户名/手机号和密码', icon: 'none' })
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
  justify-content: flex-start;
  min-height: 100vh;
  padding: 76rpx 32rpx calc(56rpx + env(safe-area-inset-bottom));
  background:
    radial-gradient(circle at 18% 0%, rgba(255, 184, 77, 0.18), transparent 300rpx),
    linear-gradient(180deg, rgba(35, 122, 75, 0.13), rgba(245, 247, 243, 0.2) 360rpx),
    #eef4ed;
}

.login-shell {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 44rpx;
  width: 100%;
}

.brand {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 8rpx 8rpx 0;
}

.brand-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 76rpx;
  height: 76rpx;
  border-radius: 22rpx;
  background: linear-gradient(145deg, #1d6a42, #2d8a5b);
  box-shadow: 0 14rpx 30rpx rgba(31, 97, 62, 0.22);
  color: #ffffff;
  font-size: 34rpx;
  font-weight: 800;
}

.brand-copy {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.brand-title,
.brand-subtitle,
.form-title,
.form-desc,
.login-tip text {
  display: block;
}

.brand-title {
  color: #172018;
  font-size: 38rpx;
  font-weight: 800;
  line-height: 1.2;
}

.brand-subtitle {
  color: #607066;
  font-size: 24rpx;
  line-height: 1.4;
}

.form-card {
  display: flex;
  flex-direction: column;
  gap: 32rpx;
  padding: 36rpx 30rpx 30rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.72);
  border-radius: 28rpx;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 20rpx 52rpx rgba(31, 47, 31, 0.1);
}

.form-header {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}

.form-title {
  color: #203327;
  font-size: 34rpx;
  font-weight: 800;
  line-height: 1.25;
}

.form-desc {
  color: #758277;
  font-size: 24rpx;
  line-height: 1.5;
}

.fields {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1rpx solid #edf1e9;
  border-radius: 20rpx;
  background: #fbfcfa;
}

.login-actions {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

:deep(.primary-login-btn) {
  height: 88rpx;
  border-radius: 18rpx;
  background: #237a4b;
  box-shadow: 0 14rpx 28rpx rgba(35, 122, 75, 0.18);
  font-size: 30rpx;
  font-weight: 800;
}

.quick-login-btn {
  width: 100%;
  min-height: 88rpx;
  padding: 0;
  border: 1rpx solid rgba(35, 122, 75, 0.34);
  border-radius: 18rpx;
  background: #f7fbf7;
  color: #237a4b;
  font-size: 28rpx;
  font-weight: 700;
  line-height: 88rpx;
}

.login-tip {
  padding: 20rpx 22rpx;
  border-radius: 18rpx;
  background: #f4f8f1;
}

.login-tip text {
  color: #6f7d70;
  font-size: 23rpx;
  line-height: 1.6;
}
</style>
