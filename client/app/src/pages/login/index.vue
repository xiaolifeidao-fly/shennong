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
            placeholder="请输入登录密码"
          />
        </view>

        <view class="agree-row">
          <view class="checkbox" :class="{ checked: agreed }" @click="toggleAgree">
            <text v-if="agreed" class="check-mark"></text>
          </view>
          <view class="agree-text">
            <text class="agree-plain" @click="toggleAgree">我已阅读并同意</text>
            <text class="agree-link" @click="openAgreement('user_agreement')">《用户协议》</text>
            <text class="agree-link" @click="openAgreement('privacy_policy')">《隐私政策》</text>
            <text class="agree-plain">及</text>
            <text class="agree-link" @click="openAgreement('privacy_guide')">《隐私保护指引》</text>
          </view>
        </view>

        <view class="login-actions">
          <wd-button
            custom-class="primary-login-btn"
            type="primary"
            block
            :loading="submitting"
            :disabled="!agreed"
            @click="handleLogin"
          >
            登录
          </wd-button>
          <button
            class="quick-login-btn"
            :class="{ disabled: !agreed }"
            open-type="getPhoneNumber"
            :loading="wechatSubmitting"
            :disabled="!agreed"
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
import { agreeAgreement, checkAgreementStatus } from '@/services/agreement'

const AGREEMENT_STORAGE_KEY = 'agreement_consent_agreed'

const userStore = useUserStore()
const submitting = ref(false)
const wechatSubmitting = ref(false)
const agreed = ref(false)
let statusChecked = false
let consentRecorded = false
const form = reactive({
  username: '',
  password: '',
})

function readLocalAgreed(): boolean {
  try {
    return Boolean(uni.getStorageSync(AGREEMENT_STORAGE_KEY))
  } catch {
    return false
  }
}

function writeLocalAgreed() {
  try {
    uni.setStorageSync(AGREEMENT_STORAGE_KEY, '1')
  } catch (error) {
    console.warn('persist agreement flag failed:', error)
  }
}

function delay(ms: number) {
  return new Promise<void>((resolve) => setTimeout(resolve, ms))
}

onShow(() => {
  redirectLoggedInUser()
  ensureAgreementStatus()
})

// 进入登录页时判断该微信是否已同意过协议，已同意则默认勾选。
// 本地已记录时直接跳过，避免额外的 wx.login 与登录时记录共用/抢占 code。
async function ensureAgreementStatus() {
  if (statusChecked) {
    return
  }
  if (readLocalAgreed()) {
    agreed.value = true
    consentRecorded = true
    statusChecked = true
    return
  }
  statusChecked = true
  try {
    const status = await checkAgreementStatus()
    if (status.agreed) {
      agreed.value = true
      consentRecorded = true
      writeLocalAgreed()
    }
  } catch (error) {
    console.warn('checkAgreementStatus failed:', error)
    statusChecked = false
  }
}

// 勾选框仅切换本地状态，真正的同意记录在登录成功后保存。
function toggleAgree() {
  agreed.value = !agreed.value
}

// 登录成功后保存该微信用户的同意记录。服务端按 openid 幂等。
// 失败会重试一次（重试可拿到新的 wx.login code），最终失败给出可见提示，避免静默丢失。
async function recordConsentIfNeeded() {
  if (consentRecorded) {
    return
  }
  for (let attempt = 0; attempt < 2; attempt += 1) {
    try {
      await agreeAgreement()
      consentRecorded = true
      writeLocalAgreed()
      return
    } catch (error) {
      console.error(`record consent failed (attempt ${attempt + 1}):`, error)
      if (attempt === 0) {
        await delay(400)
      }
    }
  }
  uni.showToast({ title: '协议同意记录保存失败，请重新登录一次', icon: 'none' })
}

function openAgreement(key: string) {
  uni.navigateTo({
    url: `/pages/agreement/index?key=${key}`,
    fail: (err) => {
      console.error('navigate agreement failed:', err)
      uni.showToast({ title: err?.errMsg || '页面打开失败', icon: 'none' })
    },
  })
}

interface GetPhoneNumberEvent {
  detail?: {
    code?: string
    errMsg?: string
  }
}

async function handleWechatPhoneLogin(event: GetPhoneNumberEvent) {
  if (!agreed.value) {
    uni.showToast({ title: '请先阅读并勾选同意协议', icon: 'none' })
    return
  }
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
    await recordConsentIfNeeded()
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
  if (!agreed.value) {
    uni.showToast({ title: '请先阅读并勾选同意协议', icon: 'none' })
    return
  }
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
    await recordConsentIfNeeded()
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
  /* 关闭全局 .page 的入场动画：其残留 transform 会让内嵌协议查看器(position:fixed)以本元素为定位基准而错位/被遮挡 */
  animation: none;
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

.quick-login-btn.disabled {
  opacity: 0.5;
}

.agree-row {
  display: flex;
  align-items: flex-start;
  gap: 14rpx;
  padding: 4rpx 4rpx 4rpx;
}

.checkbox {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 36rpx;
  height: 36rpx;
  margin-top: 4rpx;
  border: 2rpx solid #c3cdc4;
  border-radius: 50%;
  background: #ffffff;
}

.checkbox.checked {
  border-color: #237a4b;
  background: #237a4b;
}

.check-mark {
  width: 16rpx;
  height: 9rpx;
  margin-top: -4rpx;
  border-bottom: 4rpx solid #ffffff;
  border-left: 4rpx solid #ffffff;
  transform: rotate(-45deg);
}

.agree-text {
  flex: 1;
  color: #8a958c;
  font-size: 24rpx;
  line-height: 1.6;
}

.agree-plain {
  color: #8a958c;
}

.agree-link {
  color: #237a4b;
  font-weight: 700;
}
</style>
